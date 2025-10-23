/**
 * Notion ID Resolver
 *
 * Hybrid resolution system that supports:
 * - URLs: https://www.notion.so/database-id
 * - Direct IDs: database-id
 * - Names: "Tasks Database" (via cache lookup and API fallback)
 * - Smart database_id → data_source_id conversion
 *
 * Resolution stages:
 * 1. URL extraction
 * 2. Direct ID validation
 * 3. Cache lookup (exact + aliases)
 * 4. API search fallback
 * 5. Smart database_id → data_source_id resolution (for databases)
 */

import { extractNotionId, isNotionUrl } from './notion-url-parser'
import { NotionCLIError, NotionCLIErrorCode, wrapNotionError } from '../errors/enhanced-errors'
import { loadCache } from './workspace-cache'
import { search, retrieveDataSource } from '../notion'
import { isFullPage } from '@notionhq/client'

/**
 * Resolve Notion input (URL, ID, or name) to a clean Notion ID
 *
 * Supports URLs, IDs, and name-based lookups via cache and API search.
 * For databases, automatically detects and converts database_id to data_source_id.
 *
 * @param input - Database/page name, ID, or URL
 * @param type - Resource type (for better error messages)
 * @returns Clean Notion ID (32 hex characters without dashes)
 * @throws NotionCLIError if input cannot be resolved
 *
 * @example
 * // URL
 * await resolveNotionId('https://notion.so/1fb79d4c71bb8032b722c82305b63a00')
 * // Returns: '1fb79d4c71bb8032b722c82305b63a00'
 *
 * @example
 * // Direct ID
 * await resolveNotionId('1fb79d4c71bb8032b722c82305b63a00')
 * // Returns: '1fb79d4c71bb8032b722c82305b63a00'
 *
 * @example
 * // Name (via cache or API)
 * await resolveNotionId('Tasks Database', 'database')
 * // Returns: '1fb79d4c71bb8032b722c82305b63a00'
 *
 * @example
 * // database_id auto-conversion
 * await resolveNotionId('abc123...', 'database')
 * // If abc123 is a database_id, auto-resolves to data_source_id
 */
export async function resolveNotionId(
  input: string,
  type: 'database' | 'page' = 'database'
): Promise<string> {
  if (!input || typeof input !== 'string') {
    throw new NotionCLIError(
      NotionCLIErrorCode.VALIDATION_ERROR,
      `Invalid input: expected a ${type} name, ID, or URL`,
      []
    )
  }

  const trimmed = input.trim()

  // Stage 1: URL extraction
  if (isNotionUrl(trimmed)) {
    try {
      const extractedId = extractNotionId(trimmed)
      // For databases, try smart resolution in case URL contains database_id
      if (type === 'database') {
        return await trySmartDatabaseResolution(extractedId)
      }
      return extractedId
    } catch (error) {
      throw new NotionCLIError(
        NotionCLIErrorCode.INVALID_URL,
        `Invalid Notion URL: ${trimmed}\n\n` +
        `Expected format: https://www.notion.so/{id}\n` +
        `Example: https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00`,
        [],
        { originalError: error }
      )
    }
  }

  // Stage 2: Direct ID validation
  if (isValidNotionId(trimmed)) {
    const extractedId = extractNotionId(trimmed)
    // For databases, try smart resolution in case it's a database_id
    if (type === 'database') {
      return await trySmartDatabaseResolution(extractedId)
    }
    return extractedId
  }

  // Stage 3: Cache lookup (exact + aliases)
  const fromCache = await searchCache(trimmed, type)
  if (fromCache) return fromCache

  // Stage 4: API search as fallback
  const fromApi = await searchNotionApi(trimmed, type)
  if (fromApi) return fromApi

  // Nothing found - throw helpful error
  throw new NotionCLIError(
    NotionCLIErrorCode.NOT_FOUND,
    `${type === 'database' ? 'Database' : 'Page'} "${input}" not found.\n\n` +
    `Try:\n` +
    `  1. Run 'notion-cli sync' to refresh your workspace index\n` +
    `  2. Use the full Notion URL instead\n` +
    `  3. Check available databases with 'notion-cli list'`,
    []
  )
}

/**
 * Smart database resolution: handles database_id → data_source_id conversion
 *
 * When a user provides a database_id (from parent.database_id field),
 * this function detects the error and automatically resolves it to the
 * correct data_source_id.
 *
 * @param databaseId - Potential database_id or data_source_id
 * @returns data_source_id if valid, throws error otherwise
 */
async function trySmartDatabaseResolution(databaseId: string): Promise<string> {
  try {
    // Try direct lookup with data_source_id
    await retrieveDataSource(databaseId)
    // If successful, it's a valid data_source_id
    return databaseId
  } catch (error: any) {
    // Check if this is an object_not_found error (404)
    const isNotFound = error.status === 404 ||
                      error.code === 'object_not_found' ||
                      (error.notionError && error.notionError.code === 'object_not_found')

    if (isNotFound) {
      // Try to resolve database_id → data_source_id
      const dataSourceId = await resolveDatabaseIdToDataSourceId(databaseId)
      if (dataSourceId) {
        // Log helpful message about conversion
        console.log(`\nInfo: Resolved database_id to data_source_id`)
        console.log(`  database_id:    ${databaseId}`)
        console.log(`  data_source_id: ${dataSourceId}`)
        console.log(`\nNote: Use data_source_id for database operations.`)
        console.log(`      The database_id from parent.database_id won't work directly.\n`)
        return dataSourceId
      }
    }

    // If we can't resolve it, throw the original error
    throw wrapNotionError(error)
  }
}

/**
 * Resolve database_id to data_source_id by searching for pages
 *
 * When a user provides a database_id (from parent.database_id field),
 * we search for pages that have this database as their parent, and
 * extract the data_source_id from the parent field.
 *
 * @param databaseId - The database_id to resolve
 * @returns data_source_id if found, null otherwise
 */
async function resolveDatabaseIdToDataSourceId(databaseId: string): Promise<string | null> {
  try {
    // Search for pages with this database_id as parent
    const response = await search({
      filter: {
        property: 'object',
        value: 'page'
      },
      page_size: 100 // Search more pages to increase chance of finding one
    })

    if (!response || !response.results || response.results.length === 0) {
      return null
    }

    // Look through results for a page with matching parent.database_id
    for (const result of response.results) {
      if (result.object !== 'page') continue

      // Use type guard to ensure we have a full page with parent
      if (!isFullPage(result)) continue

      // Check if parent type is database_id and matches our search
      if (result.parent &&
          result.parent.type === 'database_id' &&
          result.parent.database_id === databaseId) {

        // Extract data_source_id from the same parent object
        // In the Notion API v5, pages have both database_id and data_source_id in parent
        if ('data_source_id' in result.parent) {
          return result.parent.data_source_id as string
        }
      }
    }

    return null
  } catch (error) {
    // If search fails, return null and let the main error handling deal with it
    if (process.env.DEBUG) {
      console.error('Debug: Failed to resolve database_id to data_source_id:', error)
    }
    return null
  }
}

/**
 * Check if a string is a valid Notion ID (32 hex chars with optional dashes)
 */
function isValidNotionId(input: string): boolean {
  const cleaned = input.replace(/-/g, '')
  return /^[a-f0-9]{32}$/i.test(cleaned)
}

/**
 * Search cache for database/page by name
 *
 * Searches in this order:
 * 1. Exact title match (case-insensitive)
 * 2. Alias match (case-insensitive)
 * 3. Partial title match (case-insensitive substring)
 *
 * @param query - Search query (database/page name)
 * @param type - Resource type ('database' or 'page')
 * @returns Database/page ID if found, null otherwise
 */
async function searchCache(query: string, type: 'database' | 'page'): Promise<string | null> {
  const cache = await loadCache()
  if (!cache) return null

  const normalized = query.toLowerCase().trim()

  // 1. Try exact title match
  for (const db of cache.databases) {
    if (db.titleNormalized === normalized) {
      return db.id
    }
  }

  // 2. Try alias match
  for (const db of cache.databases) {
    if (db.aliases.includes(normalized)) {
      return db.id
    }
  }

  // 3. Try partial match (substring in title)
  for (const db of cache.databases) {
    if (db.titleNormalized.includes(normalized)) {
      return db.id
    }
  }

  return null
}

/**
 * Search Notion API for database/page by name
 *
 * Uses Notion's search API as a fallback when cache lookup fails.
 *
 * @param query - Search query (database/page name)
 * @param type - Resource type ('database' or 'page')
 * @returns Database/page ID if found, null otherwise
 */
async function searchNotionApi(query: string, type: 'database' | 'page'): Promise<string | null> {
  try {
    // Search Notion API
    const response = await search({
      query,
      filter: {
        property: 'object',
        value: type === 'database' ? 'data_source' : 'page'
      },
      page_size: 10
    })

    // Return first match
    if (response && response.results && response.results.length > 0) {
      return response.results[0].id
    }

    return null
  } catch (error) {
    // API search failed, return null
    // The caller will throw a more helpful error message
    return null
  }
}
