import { Client, LogLevel, isFullBlock, isFullPage } from '@notionhq/client'
import {
  CreateDatabaseParameters,
  QueryDataSourceParameters,
  QueryDataSourceResponse,
  GetDatabaseResponse,
  GetDataSourceResponse,
  CreateDatabaseResponse,
  UpdateDatabaseParameters,
  UpdateDataSourceParameters,
  GetPageParameters,
  CreatePageParameters,
  BlockObjectRequest,
  UpdatePageParameters,
  AppendBlockChildrenParameters,
  UpdateBlockParameters,
  SearchParameters,
} from '@notionhq/client/build/src/api-endpoints'
import { cacheManager } from './cache'
import { fetchWithRetry as enhancedFetchWithRetry, RetryConfig } from './retry'
import { deduplicationManager } from './deduplication'

export const client = new Client({
  auth: process.env.NOTION_TOKEN,
  logLevel: process.env.DEBUG ? LogLevel.DEBUG : null,
})

/**
 * Legacy fetchWithRetry for backward compatibility
 * @deprecated Use the enhanced retry logic from retry.ts
 */
export const fetchWithRetry = async (
  fn: () => Promise<any>,
  retries = 3
): Promise<any> => {
  return enhancedFetchWithRetry(fn, {
    config: { maxRetries: retries },
  })
}

/**
 * Cached wrapper for API calls with retry logic and deduplication
 */
async function cachedFetch<T>(
  cacheType: string,
  cacheKey: any,
  fetchFn: () => Promise<T>,
  options: {
    cacheTtl?: number
    skipCache?: boolean
    skipDedup?: boolean
    retryConfig?: Partial<RetryConfig>
  } = {}
): Promise<T> {
  const { cacheTtl, skipCache = false, skipDedup = false, retryConfig } = options

  // Check cache first (unless skipped or cache disabled)
  if (!skipCache) {
    const cached = cacheManager.get<T>(cacheType, cacheKey)
    if (cached !== null) {
      if (process.env.DEBUG) {
        console.log(`Cache HIT: ${cacheType}:${cacheKey}`)
      }
      return cached
    }
    if (process.env.DEBUG) {
      console.log(`Cache MISS: ${cacheType}:${cacheKey}`)
    }
  }

  // Generate deduplication key
  const dedupKey = `${cacheType}:${JSON.stringify(cacheKey)}`

  // Wrap fetch function with deduplication (unless disabled)
  const dedupEnabled = process.env.NOTION_CLI_DEDUP_ENABLED !== 'false' && !skipDedup
  const fetchWithDedup = dedupEnabled
    ? () => deduplicationManager.execute(dedupKey, async () => {
        if (process.env.DEBUG) {
          console.log(`Dedup MISS: ${dedupKey}`)
        }
        return enhancedFetchWithRetry(fetchFn, {
          config: retryConfig,
          context: `${cacheType}:${cacheKey}`,
        })
      })
    : () => enhancedFetchWithRetry(fetchFn, {
        config: retryConfig,
        context: `${cacheType}:${cacheKey}`,
      })

  // Execute fetch (with or without deduplication)
  const data = await fetchWithDedup()

  // Store in cache
  if (!skipCache) {
    cacheManager.set(cacheType, data, cacheTtl, cacheKey)
  }

  return data
}

/**
 * Fetch all pages in a data source with pagination
 */
export const fetchAllPagesInDS = async (
  databaseId: string,
  filter?: object | undefined
): Promise<QueryDataSourceResponse['results']> => {
  const f = filter as QueryDataSourceParameters['filter']
  const pages = []
  let cursor: string | undefined = undefined

  while (true) {
    const { results, next_cursor } = await enhancedFetchWithRetry(
      () => client.dataSources.query({
        data_source_id: databaseId,
        filter: f,
        start_cursor: cursor,
      }),
      { context: `fetchAllPagesInDS:${databaseId}` }
    )
    pages.push(...results)
    if (!next_cursor) {
      break
    }
    cursor = next_cursor
  }

  return pages
}

/**
 * Create a database
 */
export const createDb = async (
  dbProps: CreateDatabaseParameters
): Promise<CreateDatabaseResponse> => {
  const result = await enhancedFetchWithRetry(
    () => client.databases.create(dbProps),
    { context: 'createDb' }
  )

  // Invalidate database list cache
  cacheManager.invalidate('search')

  return result
}

/**
 * Update a database
 */
export const updateDb = async (
  dbProps: UpdateDatabaseParameters
): Promise<GetDatabaseResponse> => {
  const result = await enhancedFetchWithRetry(
    () => client.databases.update(dbProps),
    { context: `updateDb:${dbProps.database_id}` }
  )

  // Invalidate this database's cache
  cacheManager.invalidate('database', dbProps.database_id)
  cacheManager.invalidate('dataSource', dbProps.database_id)

  return result
}

/**
 * Retrieve a database (cached)
 */
export const retrieveDb = async (databaseId: string): Promise<GetDatabaseResponse> => {
  return cachedFetch(
    'database',
    databaseId,
    () => client.databases.retrieve({ database_id: databaseId })
  )
}

/**
 * Retrieve a data source (cached)
 */
export const retrieveDataSource = async (dataSourceId: string): Promise<GetDataSourceResponse> => {
  return cachedFetch(
    'dataSource',
    dataSourceId,
    () => client.dataSources.retrieve({ data_source_id: dataSourceId })
  )
}

/**
 * Update a data source
 */
export const updateDataSource = async (
  dsProps: UpdateDataSourceParameters
): Promise<GetDataSourceResponse> => {
  const result = await enhancedFetchWithRetry(
    () => client.dataSources.update(dsProps),
    { context: `updateDataSource:${dsProps.data_source_id}` }
  )

  // Invalidate this data source's cache
  cacheManager.invalidate('dataSource', dsProps.data_source_id)

  return result
}

/**
 * Retrieve a page (cached with short TTL)
 */
export const retrievePage = async (pageProp: GetPageParameters) => {
  return cachedFetch(
    'page',
    pageProp.page_id,
    () => client.pages.retrieve(pageProp)
  )
}

/**
 * Retrieve page property
 */
export const retrievePageProperty = async (pageId: string, propId: string) => {
  return enhancedFetchWithRetry(
    () => client.pages.properties.retrieve({
      page_id: pageId,
      property_id: propId,
    }),
    { context: `retrievePageProperty:${pageId}:${propId}` }
  )
}

/**
 * Create a page
 */
export const createPage = async (pageProps: CreatePageParameters) => {
  const result = await enhancedFetchWithRetry(
    () => client.pages.create(pageProps),
    { context: 'createPage' }
  )

  // Invalidate parent database/page cache
  if ('parent' in pageProps && 'database_id' in pageProps.parent) {
    cacheManager.invalidate('dataSource', pageProps.parent.database_id)
  }

  return result
}

/**
 * Update page properties
 */
export const updatePageProps = async (pageParams: UpdatePageParameters) => {
  const result = await enhancedFetchWithRetry(
    () => client.pages.update(pageParams),
    { context: `updatePageProps:${pageParams.page_id}` }
  )

  // Invalidate this page's cache
  cacheManager.invalidate('page', pageParams.page_id)

  return result
}

/**
 * Update page content by replacing all blocks
 * To keep the same page URL, remove all blocks in the page and add new blocks
 */
export const updatePage = async (pageId: string, blocks: BlockObjectRequest[]) => {
  // Get all blocks
  const blks = await enhancedFetchWithRetry(
    () => client.blocks.children.list({ block_id: pageId }),
    { context: `updatePage:list:${pageId}` }
  )

  // Delete all blocks
  for (const blk of blks.results) {
    await enhancedFetchWithRetry(
      () => client.blocks.delete({ block_id: blk.id }),
      { context: `updatePage:delete:${blk.id}` }
    )
  }

  // Append new blocks
  const res = await enhancedFetchWithRetry(
    () => client.blocks.children.append({
      block_id: pageId,

      children: blocks,
    }),
    { context: `updatePage:append:${pageId}` }
  )

  // Invalidate caches
  cacheManager.invalidate('page', pageId)
  cacheManager.invalidate('block', pageId)

  return res
}

/**
 * Retrieve a block (cached with very short TTL)
 */
export const retrieveBlock = async (blockId: string) => {
  return cachedFetch(
    'block',
    blockId,
    () => client.blocks.retrieve({ block_id: blockId })
  )
}

/**
 * Update a block
 */
export const updateBlock = async (params: UpdateBlockParameters) => {
  const result = await enhancedFetchWithRetry(
    () => client.blocks.update(params),
    { context: `updateBlock:${params.block_id}` }
  )

  // Invalidate this block's cache
  cacheManager.invalidate('block', params.block_id)

  return result
}

/**
 * Retrieve block children (cached with very short TTL)
 */
export const retrieveBlockChildren = async (blockId: string) => {
  return cachedFetch(
    'block',
    `${blockId}:children`,
    () => client.blocks.children.list({ block_id: blockId })
  )
}

/**
 * Append block children
 */
export const appendBlockChildren = async (params: AppendBlockChildrenParameters) => {
  const result = await enhancedFetchWithRetry(
    () => client.blocks.children.append(params),
    { context: `appendBlockChildren:${params.block_id}` }
  )

  // Invalidate parent block's cache
  cacheManager.invalidate('block', params.block_id)
  cacheManager.invalidate('block', `${params.block_id}:children`)

  return result
}

/**
 * Delete a block
 */
export const deleteBlock = async (blockId: string) => {
  const result = await enhancedFetchWithRetry(
    () => client.blocks.delete({ block_id: blockId }),
    { context: `deleteBlock:${blockId}` }
  )

  // Invalidate this block's cache
  cacheManager.invalidate('block', blockId)

  return result
}

/**
 * Retrieve a user (cached with long TTL)
 */
export const retrieveUser = async (userId: string) => {
  return cachedFetch(
    'user',
    userId,
    () => client.users.retrieve({ user_id: userId })
  )
}

/**
 * List all users (cached with long TTL)
 */
export const listUser = async () => {
  return cachedFetch(
    'user',
    'list',
    () => client.users.list({})
  )
}

/**
 * Get bot user info (cached with long TTL)
 */
export const botUser = async () => {
  return cachedFetch(
    'user',
    'me',
    () => client.users.me({})
  )
}

/**
 * Search for databases (cached with medium TTL)
 */
export const searchDb = async () => {
  const { results } = await cachedFetch(
    'search',
    'databases',
    async () => {
      return await client.search({
        filter: {
          value: 'data_source',
          property: 'object',
        },
      })
    }
  )
  return results
}

/**
 * General search (not cached due to variable parameters)
 */
export const search = async (params: SearchParameters) => {
  return enhancedFetchWithRetry(
    () => client.search(params),
    { context: 'search' }
  )
}

/**
 * Export cache manager for external use
 */
export { cacheManager } from './cache'

/**
 * Export retry utilities for external use
 */
export { fetchWithRetry as enhancedFetchWithRetry, CircuitBreaker } from './retry'

/**
 * Recursively retrieve a page with all its blocks and nested content
 * @param pageId - The ID of the page to retrieve
 * @param depth - Current recursion depth (internal use)
 * @param maxDepth - Maximum depth to recurse (default: 3)
 * @returns Object containing page metadata, blocks, and optional warnings
 */
export const retrievePageRecursive = async (
  pageId: string,
  depth = 0,
  maxDepth = 3
): Promise<{
  page: any
  blocks: any[]
  warnings?: Array<{
    block_id: string
    type: string
    notion_type?: string
    message: string
    has_children: boolean
  }>
}> => {
  // Prevent infinite recursion
  if (depth >= maxDepth) {
    return {
      page: null,
      blocks: [],
      warnings: [
        {
          block_id: pageId,
          type: 'max_depth_reached',
          message: `Maximum recursion depth of ${maxDepth} reached`,
          has_children: false,
        },
      ],
    }
  }

  // Retrieve the page
  const page = await retrievePage({ page_id: pageId })

  // Retrieve all blocks (children)
  const blocksResponse = await retrieveBlockChildren(pageId)
  const blocks = blocksResponse.results || []

  const warnings: any[] = []

  // Recursively fetch nested blocks
  for (const block of blocks) {
    // Skip partial blocks
    if (!isFullBlock(block)) {
      continue
    }

    // Handle unsupported blocks
    if (block.type === 'unsupported') {
      warnings.push({
        block_id: block.id,
        type: 'unsupported',
        notion_type: (block as any).unsupported?.type || 'unknown',
        message: `Block type '${(block as any).unsupported?.type || 'unknown'}' not supported by Notion API`,
        has_children: block.has_children,
      })
      continue
    }

    // Recursively fetch children for blocks that have them
    if (block.has_children) {
      try {
        const childrenResponse = await retrieveBlockChildren(block.id)
        ;(block as any).children = childrenResponse.results || []

        // If this is a child_page block, recursively fetch that page too
        if (block.type === 'child_page' && depth + 1 < maxDepth) {
          const childPageData = await retrievePageRecursive(
            block.id,
            depth + 1,
            maxDepth
          )
          ;(block as any).child_page_details = childPageData

          // Merge warnings from recursive calls
          if (childPageData.warnings) {
            warnings.push(...childPageData.warnings)
          }
        }
      } catch (error) {
        // If we can't fetch children, add a warning
        warnings.push({
          block_id: block.id,
          type: 'fetch_error',
          message: `Failed to fetch children for block: ${error instanceof Error ? error.message : 'Unknown error'}`,
          has_children: true,
        })
      }
    }
  }

  return {
    page,
    blocks,
    ...(warnings.length > 0 && { warnings }),
  }
}

/**
 * Map page structure (fast page discovery with parallel fetching)
 * Returns minimal structure info (titles, types, IDs) instead of full content
 * @param pageId - The ID of the page to map
 * @returns Object containing page ID, title, icon, and structure overview
 */
export const mapPageStructure = async (pageId: string): Promise<{
  id: string
  title: string
  type: string
  icon?: string
  structure: Array<{
    type: string
    id: string
    title?: string
    text?: string
  }>
}> => {
  // Parallel fetch: get page and blocks simultaneously
  const [page, blocksResponse] = await Promise.all([
    retrievePage({ page_id: pageId }),
    retrieveBlockChildren(pageId),
  ])

  const blocks = blocksResponse.results || []

  // Extract page title
  let pageTitle = 'Untitled'
  if (page.object === 'page' && isFullPage(page)) {
    Object.entries(page.properties).find(([, prop]: [string, any]) => {
      if (prop.type === 'title' && prop.title.length > 0) {
        pageTitle = prop.title[0].plain_text
        return true
      }
      return false
    })
  }

  // Extract page icon
  let pageIcon: string | undefined
  if (isFullPage(page) && page.icon) {
    if (page.icon.type === 'emoji') {
      pageIcon = page.icon.emoji
    } else if (page.icon.type === 'external') {
      pageIcon = page.icon.external.url
    } else if (page.icon.type === 'file') {
      pageIcon = page.icon.file.url
    }
  }

  // Build minimal structure
  const structure = blocks.map((block: any) => {
    const structureItem: any = {
      type: block.type,
      id: block.id,
    }

    // Extract title/text based on block type
    try {
      switch (block.type) {
        case 'child_page':
          structureItem.title = block[block.type].title
          break
        case 'child_database':
          structureItem.title = block[block.type].title
          break
        case 'heading_1':
        case 'heading_2':
        case 'heading_3':
        case 'paragraph':
        case 'bulleted_list_item':
        case 'numbered_list_item':
        case 'to_do':
        case 'toggle':
        case 'quote':
        case 'callout':
        case 'code':
          if (block[block.type].rich_text && block[block.type].rich_text.length > 0) {
            structureItem.text = block[block.type].rich_text[0].plain_text
          }
          break
        case 'bookmark':
        case 'embed':
        case 'link_preview':
          structureItem.text = block[block.type].url
          break
        case 'equation':
          structureItem.text = block[block.type].expression
          break
        case 'image':
        case 'file':
        case 'video':
        case 'pdf':
          if (block[block.type].type === 'file') {
            structureItem.text = block[block.type].file.url
          } else if (block[block.type].type === 'external') {
            structureItem.text = block[block.type].external.url
          }
          break
        // For other types, just include type and id
        default:
          break
      }
    } catch {
      // If extraction fails, just include type and id
    }

    return structureItem
  })

  return {
    id: pageId,
    title: pageTitle,
    type: 'page',
    ...(pageIcon && { icon: pageIcon }),
    structure,
  }
}
