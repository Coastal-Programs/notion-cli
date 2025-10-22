/**
 * Workspace Cache Utility
 *
 * Manages persistent caching of workspace databases for fast name-to-ID resolution.
 * Cache is stored at ~/.notion-cli/databases.json
 */

import * as fs from 'fs/promises'
import * as path from 'path'
import * as os from 'os'
import { GetDataSourceResponse, DataSourceObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import { isFullDataSource } from '@notionhq/client'

export interface CachedDatabase {
  id: string
  title: string
  titleNormalized: string
  aliases: string[]
  url?: string
  lastEditedTime?: string
  properties?: Record<string, any>
}

export interface WorkspaceCache {
  version: string
  lastSync: string
  databases: CachedDatabase[]
}

const CACHE_VERSION = '1.0.0'
const CACHE_DIR_NAME = '.notion-cli'
const CACHE_FILE_NAME = 'databases.json'

/**
 * Get the cache directory path
 */
export function getCacheDir(): string {
  return path.join(os.homedir(), CACHE_DIR_NAME)
}

/**
 * Get the cache file path
 */
export async function getCachePath(): Promise<string> {
  return path.join(getCacheDir(), CACHE_FILE_NAME)
}

/**
 * Ensure cache directory exists
 */
export async function ensureCacheDir(): Promise<void> {
  const cacheDir = getCacheDir()
  try {
    await fs.mkdir(cacheDir, { recursive: true })
  } catch (error: any) {
    if (error.code !== 'EEXIST') {
      throw new Error(`Failed to create cache directory: ${error.message}`)
    }
  }
}

/**
 * Load cache from disk
 * Returns null if cache doesn't exist or is corrupted
 */
export async function loadCache(): Promise<WorkspaceCache | null> {
  try {
    const cachePath = await getCachePath()
    const content = await fs.readFile(cachePath, 'utf-8')
    const cache = JSON.parse(content)

    // Validate cache structure
    if (!cache.version || !Array.isArray(cache.databases)) {
      console.warn('Cache file is corrupted, will rebuild on next sync')
      return null
    }

    return cache as WorkspaceCache
  } catch (error: any) {
    if (error.code === 'ENOENT') {
      // Cache doesn't exist yet
      return null
    }

    // Parse error or other error
    console.warn(`Failed to load cache: ${error.message}`)
    return null
  }
}

/**
 * Save cache to disk (atomic write)
 */
export async function saveCache(data: WorkspaceCache): Promise<void> {
  await ensureCacheDir()

  const cachePath = await getCachePath()
  const tmpPath = `${cachePath}.tmp`

  try {
    // Write to temporary file
    await fs.writeFile(tmpPath, JSON.stringify(data, null, 2), 'utf-8')

    // Atomic rename (replaces old file)
    await fs.rename(tmpPath, cachePath)
  } catch (error: any) {
    // Clean up temp file if it exists
    try {
      await fs.unlink(tmpPath)
    } catch {}

    throw new Error(`Failed to save cache: ${error.message}`)
  }
}

/**
 * Generate search aliases from a database title
 *
 * @example
 * generateAliases('Tasks Database')
 * // Returns: ['tasks database', 'tasks', 'task', 'tasks db', 'task db', 'td']
 */
export function generateAliases(title: string): string[] {
  const aliases = new Set<string>()
  const normalized = title.toLowerCase().trim()

  // Add full title (normalized)
  aliases.add(normalized)

  // Add title without common suffixes
  const withoutSuffixes = normalized.replace(/\s+(database|db|table|list|tracker|log)$/i, '')
  if (withoutSuffixes !== normalized) {
    aliases.add(withoutSuffixes)
  }

  // Add title with common suffixes
  if (withoutSuffixes !== normalized) {
    aliases.add(`${withoutSuffixes} db`)
    aliases.add(`${withoutSuffixes} database`)
  }

  // Add singular/plural variants
  if (withoutSuffixes.endsWith('s')) {
    aliases.add(withoutSuffixes.slice(0, -1)) // Remove 's'
  } else {
    aliases.add(`${withoutSuffixes}s`) // Add 's'
  }

  // Add acronym if multi-word (e.g., "Meeting Notes" → "mn")
  const words = withoutSuffixes.split(/\s+/)
  if (words.length > 1) {
    const acronym = words.map(w => w[0]).join('')
    if (acronym.length >= 2) {
      aliases.add(acronym)
    }
  }

  return Array.from(aliases)
}

/**
 * Build a cached database entry from a data source response
 */
export function buildCacheEntry(dataSource: GetDataSourceResponse | DataSourceObjectResponse): CachedDatabase {
  let title = 'Untitled'
  let properties: Record<string, any> = {}
  let url: string | undefined
  let lastEditedTime: string | undefined

  if (isFullDataSource(dataSource)) {
    if (dataSource.title && dataSource.title.length > 0) {
      title = dataSource.title[0].plain_text
    }

    // Extract property schema
    if (dataSource.properties) {
      properties = dataSource.properties
    }

    // Extract URL if available
    if ('url' in dataSource) {
      url = dataSource.url as string
    }

    // Extract last edited time
    if ('last_edited_time' in dataSource) {
      lastEditedTime = (dataSource as any).last_edited_time
    }
  }

  const titleNormalized = title.toLowerCase().trim()
  const aliases = generateAliases(title)

  return {
    id: dataSource.id,
    title,
    titleNormalized,
    aliases,
    url,
    lastEditedTime,
    properties,
  }
}

/**
 * Create an empty cache
 */
export function createEmptyCache(): WorkspaceCache {
  return {
    version: CACHE_VERSION,
    lastSync: new Date().toISOString(),
    databases: [],
  }
}
