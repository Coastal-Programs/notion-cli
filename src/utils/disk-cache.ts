/**
 * Disk Cache Manager
 *
 * Provides persistent caching to disk, maintaining cache across CLI invocations.
 * Cache entries are stored in ~/.notion-cli/cache/ directory.
 */

import * as fs from 'fs/promises'
import * as path from 'path'
import * as os from 'os'
import * as crypto from 'crypto'

export interface DiskCacheEntry<T = any> {
  key: string
  data: T
  expiresAt: number
  createdAt: number
  size: number
}

export interface DiskCacheStats {
  totalEntries: number
  totalSize: number
  oldestEntry: number | null
  newestEntry: number | null
}

const CACHE_DIR_NAME = '.notion-cli'
const CACHE_SUBDIR = 'cache'
const DEFAULT_MAX_SIZE = 100 * 1024 * 1024 // 100MB
const DEFAULT_SYNC_INTERVAL = 5000 // 5 seconds

export class DiskCacheManager {
  private cacheDir: string
  private maxSize: number
  private syncInterval: number
  private dirtyKeys: Set<string> = new Set()
  private syncTimer: NodeJS.Timeout | null = null
  private initialized = false

  constructor(options: {
    cacheDir?: string
    maxSize?: number
    syncInterval?: number
  } = {}) {
    this.cacheDir = options.cacheDir || path.join(os.homedir(), CACHE_DIR_NAME, CACHE_SUBDIR)
    this.maxSize = options.maxSize || parseInt(process.env.NOTION_CLI_DISK_CACHE_MAX_SIZE || String(DEFAULT_MAX_SIZE), 10)
    this.syncInterval = options.syncInterval || parseInt(process.env.NOTION_CLI_DISK_CACHE_SYNC_INTERVAL || String(DEFAULT_SYNC_INTERVAL), 10)
  }

  /**
   * Initialize disk cache (create directory, start sync timer)
   */
  async initialize(): Promise<void> {
    if (this.initialized) {
      return
    }

    await this.ensureCacheDir()
    await this.enforceMaxSize()

    // Start periodic sync timer
    if (this.syncInterval > 0) {
      this.syncTimer = setInterval(() => {
        this.sync().catch(error => {
          if (process.env.DEBUG) {
            console.warn('Disk cache sync error:', error)
          }
        })
      }, this.syncInterval)

      // Don't keep the process alive
      if (this.syncTimer.unref) {
        this.syncTimer.unref()
      }
    }

    this.initialized = true
  }

  /**
   * Get a cache entry from disk
   */
  async get<T>(key: string): Promise<DiskCacheEntry<T> | null> {
    try {
      const filePath = this.getFilePath(key)
      const content = await fs.readFile(filePath, 'utf-8')
      const entry: DiskCacheEntry<T> = JSON.parse(content)

      // Check if expired
      if (Date.now() > entry.expiresAt) {
        // Delete expired entry
        await this.invalidate(key)
        return null
      }

      return entry
    } catch (error: any) {
      if (error.code === 'ENOENT') {
        return null
      }

      if (process.env.DEBUG) {
        console.warn(`Failed to read cache entry ${key}:`, error.message)
      }
      return null
    }
  }

  /**
   * Set a cache entry to disk
   */
  async set<T>(key: string, data: T, ttl: number): Promise<void> {
    const entry: DiskCacheEntry<T> = {
      key,
      data,
      expiresAt: Date.now() + ttl,
      createdAt: Date.now(),
      size: JSON.stringify(data).length,
    }

    const filePath = this.getFilePath(key)
    const tmpPath = `${filePath}.tmp`

    try {
      // Write to temporary file
      await fs.writeFile(tmpPath, JSON.stringify(entry), 'utf-8')

      // Atomic rename
      await fs.rename(tmpPath, filePath)

      this.dirtyKeys.delete(key)
    } catch (error: any) {
      // Clean up temp file if it exists
      try {
        await fs.unlink(tmpPath)
      } catch {
        // Ignore cleanup errors
      }

      if (process.env.DEBUG) {
        console.warn(`Failed to write cache entry ${key}:`, error.message)
      }
    }

    // Check if we need to enforce size limits
    const stats = await this.getStats()
    if (stats.totalSize > this.maxSize) {
      await this.enforceMaxSize()
    }
  }

  /**
   * Invalidate (delete) a cache entry
   */
  async invalidate(key: string): Promise<void> {
    try {
      const filePath = this.getFilePath(key)
      await fs.unlink(filePath)
      this.dirtyKeys.delete(key)
    } catch (error: any) {
      if (error.code !== 'ENOENT') {
        if (process.env.DEBUG) {
          console.warn(`Failed to delete cache entry ${key}:`, error.message)
        }
      }
    }
  }

  /**
   * Clear all cache entries
   */
  async clear(): Promise<void> {
    try {
      const files = await fs.readdir(this.cacheDir)
      await Promise.all(
        files
          .filter(file => !file.endsWith('.tmp'))
          .map(file => fs.unlink(path.join(this.cacheDir, file)).catch(() => {}))
      )
      this.dirtyKeys.clear()
    } catch (error: any) {
      if (error.code !== 'ENOENT') {
        if (process.env.DEBUG) {
          console.warn('Failed to clear cache:', error.message)
        }
      }
    }
  }

  /**
   * Sync dirty entries to disk
   */
  async sync(): Promise<void> {
    // In our implementation, writes are immediate (no write buffering)
    // This method is here for API compatibility
    this.dirtyKeys.clear()
  }

  /**
   * Shutdown (flush and cleanup)
   */
  async shutdown(): Promise<void> {
    if (this.syncTimer) {
      clearInterval(this.syncTimer)
      this.syncTimer = null
    }

    await this.sync()
    this.initialized = false
  }

  /**
   * Get cache statistics
   */
  async getStats(): Promise<DiskCacheStats> {
    try {
      const files = await fs.readdir(this.cacheDir)
      const entries: DiskCacheEntry[] = []

      for (const file of files) {
        if (file.endsWith('.tmp')) {
          continue
        }

        try {
          const content = await fs.readFile(path.join(this.cacheDir, file), 'utf-8')
          const entry: DiskCacheEntry = JSON.parse(content)
          entries.push(entry)
        } catch {
          // Skip corrupted entries
        }
      }

      const totalSize = entries.reduce((sum, entry) => sum + entry.size, 0)
      const timestamps = entries.map(e => e.createdAt)

      return {
        totalEntries: entries.length,
        totalSize,
        oldestEntry: timestamps.length > 0 ? Math.min(...timestamps) : null,
        newestEntry: timestamps.length > 0 ? Math.max(...timestamps) : null,
      }
    } catch (error: any) {
      return {
        totalEntries: 0,
        totalSize: 0,
        oldestEntry: null,
        newestEntry: null,
      }
    }
  }

  /**
   * Enforce maximum cache size by removing oldest entries
   */
  private async enforceMaxSize(): Promise<void> {
    try {
      const files = await fs.readdir(this.cacheDir)
      const entries: Array<{ file: string; entry: DiskCacheEntry }> = []

      // Load all entries
      for (const file of files) {
        if (file.endsWith('.tmp')) {
          continue
        }

        try {
          const filePath = path.join(this.cacheDir, file)
          const content = await fs.readFile(filePath, 'utf-8')
          const entry: DiskCacheEntry = JSON.parse(content)

          // Remove expired entries
          if (Date.now() > entry.expiresAt) {
            await fs.unlink(filePath)
            continue
          }

          entries.push({ file, entry })
        } catch {
          // Skip corrupted entries
        }
      }

      // Calculate total size
      const totalSize = entries.reduce((sum, { entry }) => sum + entry.size, 0)

      // If under limit, we're done
      if (totalSize <= this.maxSize) {
        return
      }

      // Sort by creation time (oldest first)
      entries.sort((a, b) => a.entry.createdAt - b.entry.createdAt)

      // Remove oldest entries until under limit
      let currentSize = totalSize
      for (const { file, entry } of entries) {
        if (currentSize <= this.maxSize) {
          break
        }

        try {
          await fs.unlink(path.join(this.cacheDir, file))
          currentSize -= entry.size
        } catch {
          // Skip deletion errors
        }
      }
    } catch (error: any) {
      if (process.env.DEBUG) {
        console.warn('Failed to enforce max size:', error.message)
      }
    }
  }

  /**
   * Ensure cache directory exists
   */
  private async ensureCacheDir(): Promise<void> {
    try {
      await fs.mkdir(this.cacheDir, { recursive: true })
    } catch (error: any) {
      if (error.code !== 'EEXIST') {
        throw new Error(`Failed to create cache directory: ${error.message}`)
      }
    }
  }

  /**
   * Get file path for a cache key
   */
  private getFilePath(key: string): string {
    // Hash the key to create a safe filename
    const hash = crypto.createHash('sha256').update(key).digest('hex')
    return path.join(this.cacheDir, `${hash}.json`)
  }
}

/**
 * Global singleton instance
 */
export const diskCacheManager = new DiskCacheManager()
