import { Client, LogLevel } from '@notionhq/client'
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
 * Cached wrapper for API calls with retry logic
 */
async function cachedFetch<T>(
  cacheType: string,
  cacheKey: any,
  fetchFn: () => Promise<T>,
  options: {
    cacheTtl?: number
    skipCache?: boolean
    retryConfig?: Partial<RetryConfig>
  } = {}
): Promise<T> {
  const { cacheTtl, skipCache = false, retryConfig } = options

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

  // Fetch with retry logic
  const data = await enhancedFetchWithRetry(fetchFn, {
    config: retryConfig,
    context: `${cacheType}:${cacheKey}`,
  })

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
      // @ts-ignore
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
