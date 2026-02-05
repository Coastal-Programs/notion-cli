/**
 * Simple in-memory caching layer for Notion API responses
 * Supports TTL (time-to-live) and cache invalidation
 * Integrated with disk cache for persistence across CLI invocations
 */

import { diskCacheManager, DiskCacheEntry } from './utils/disk-cache'

export interface CacheEntry<T> {
  data: T
  timestamp: number
  ttl: number
}

export interface CacheStats {
  hits: number
  misses: number
  sets: number
  evictions: number
  size: number
}

export interface CacheConfig {
  enabled: boolean
  defaultTtl: number
  maxSize: number
  ttlByType: {
    dataSource: number
    database: number
    user: number
    page: number
    block: number
  }
}

/**
 * Structured cache event for logging to stderr
 */
interface CacheEvent {
  level: 'debug' | 'info'
  event: 'cache_hit' | 'cache_miss' | 'cache_set' | 'cache_invalidate' | 'cache_evict'
  namespace: string
  key?: string
  age_ms?: number
  ttl_ms?: number
  cache_size?: number
  timestamp: string
}

/**
 * Check if verbose logging is enabled
 */
function isVerboseEnabled(): boolean {
  return process.env.DEBUG === 'true' ||
         process.env.NOTION_CLI_DEBUG === 'true' ||
         process.env.NOTION_CLI_VERBOSE === 'true'
}

/**
 * Log structured cache event to stderr
 * Never pollutes stdout - safe for JSON output
 */
function logCacheEvent(event: CacheEvent): void {
  // Only log if verbose mode is enabled
  if (!isVerboseEnabled()) {
    return
  }

  // Always write to stderr, never stdout
  console.error(JSON.stringify(event))
}

export class CacheManager {
  private cache: Map<string, CacheEntry<unknown>>
  private stats: CacheStats
  private config: CacheConfig

  constructor(config?: Partial<CacheConfig>) {
    this.cache = new Map()
    this.stats = {
      hits: 0,
      misses: 0,
      sets: 0,
      evictions: 0,
      size: 0,
    }

    // Default configuration
    this.config = {
      enabled: process.env.NOTION_CLI_CACHE_ENABLED !== 'false',
      defaultTtl: parseInt(process.env.NOTION_CLI_CACHE_TTL || '300000', 10), // 5 minutes default
      maxSize: parseInt(process.env.NOTION_CLI_CACHE_MAX_SIZE || '1000', 10),
      ttlByType: {
        dataSource: parseInt(process.env.NOTION_CLI_CACHE_DS_TTL || '600000', 10), // 10 min
        database: parseInt(process.env.NOTION_CLI_CACHE_DB_TTL || '600000', 10), // 10 min
        user: parseInt(process.env.NOTION_CLI_CACHE_USER_TTL || '3600000', 10), // 1 hour
        page: parseInt(process.env.NOTION_CLI_CACHE_PAGE_TTL || '60000', 10), // 1 min
        block: parseInt(process.env.NOTION_CLI_CACHE_BLOCK_TTL || '30000', 10), // 30 sec
      },
      ...config,
    }
  }

  /**
   * Generate a cache key from resource type and identifiers
   */
  private generateKey(type: string, ...identifiers: Array<string | number | object>): string {
    return `${type}:${identifiers.map(id =>
      typeof id === 'object' ? JSON.stringify(id) : String(id)
    ).join(':')}`
  }

  /**
   * Check if a cache entry is still valid
   */
  private isValid(entry: CacheEntry<unknown>): boolean {
    const now = Date.now()
    return now - entry.timestamp < entry.ttl
  }

  /**
   * Evict expired entries
   */
  private evictExpired(): void {
    let evictedCount = 0

    for (const [key, entry] of this.cache.entries()) {
      if (!this.isValid(entry)) {
        this.cache.delete(key)
        this.stats.evictions++
        evictedCount++
      }
    }

    this.stats.size = this.cache.size

    // Log eviction event if any entries were evicted
    if (evictedCount > 0 && isVerboseEnabled()) {
      logCacheEvent({
        level: 'debug',
        event: 'cache_evict',
        namespace: 'expired',
        cache_size: this.cache.size,
        timestamp: new Date().toISOString(),
      })
    }
  }

  /**
   * Evict oldest entries if cache is full
   */
  private evictOldest(): void {
    if (this.cache.size >= this.config.maxSize) {
      // Find and remove oldest entry
      let oldestKey: string | null = null
      let oldestTime = Infinity

      for (const [key, entry] of this.cache.entries()) {
        if (entry.timestamp < oldestTime) {
          oldestTime = entry.timestamp
          oldestKey = key
        }
      }

      if (oldestKey) {
        this.cache.delete(oldestKey)
        this.stats.evictions++

        // Log LRU eviction
        logCacheEvent({
          level: 'debug',
          event: 'cache_evict',
          namespace: 'lru',
          key: oldestKey,
          cache_size: this.cache.size,
          timestamp: new Date().toISOString(),
        })
      }
    }
  }

  /**
   * Get a value from cache (checks memory, then disk)
   */
  get<T>(type: string, ...identifiers: Array<string | number | object>): T | null {
    if (!this.config.enabled) {
      return null
    }

    const key = this.generateKey(type, ...identifiers)
    const entry = this.cache.get(key)

    // Check memory cache first
    if (entry && this.isValid(entry)) {
      this.stats.hits++

      // Log cache hit
      logCacheEvent({
        level: 'debug',
        event: 'cache_hit',
        namespace: type,
        key: identifiers.join(':'),
        age_ms: Date.now() - entry.timestamp,
        ttl_ms: entry.ttl,
        timestamp: new Date().toISOString(),
      })

      return entry.data as T
    }

    // Remove invalid memory entry
    if (entry) {
      this.cache.delete(key)
      this.stats.evictions++
    }

    // Check disk cache (synchronously, but non-blocking for performance)
    // Note: This is a synchronous wrapper for the async disk cache
    // We use a sync approach here to maintain the current API signature
    this.checkDiskCache<T>(key, type, identifiers)

    // If we still have no entry after disk check, it's a miss
    const finalEntry = this.cache.get(key)
    if (!finalEntry || !this.isValid(finalEntry)) {
      this.stats.misses++

      // Log cache miss
      logCacheEvent({
        level: 'debug',
        event: 'cache_miss',
        namespace: type,
        key: identifiers.join(':'),
        timestamp: new Date().toISOString(),
      })

      return null
    }

    this.stats.hits++

    // Log cache hit (from disk)
    logCacheEvent({
      level: 'debug',
      event: 'cache_hit',
      namespace: type,
      key: identifiers.join(':'),
      age_ms: Date.now() - finalEntry.timestamp,
      ttl_ms: finalEntry.ttl,
      timestamp: new Date().toISOString(),
    })

    return finalEntry.data as T
  }

  /**
   * Check disk cache and promote to memory if found (synchronous wrapper)
   * @private
   */
  private checkDiskCache<T>(key: string, type: string, identifiers: Array<string | number | object>): void {
    // Only check disk cache if enabled
    const diskEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED !== 'false'
    if (!diskEnabled) {
      return
    }

    // Fire-and-forget disk cache check
    // We don't await here to keep the API synchronous
    diskCacheManager.get<CacheEntry<T>>(key).then(diskEntry => {
      if (diskEntry && diskEntry.data) {
        const entry = diskEntry.data as CacheEntry<T>

        // Validate disk entry
        if (this.isValid(entry)) {
          // Promote to memory cache
          this.cache.set(key, entry)

          if (process.env.DEBUG) {
            console.error(JSON.stringify({
              level: 'debug',
              event: 'disk_cache_hit',
              namespace: type,
              key: identifiers.join(':'),
              timestamp: new Date().toISOString(),
            }))
          }
        } else {
          // Remove expired disk entry
          diskCacheManager.invalidate(key).catch(() => {})
        }
      }
    }).catch(() => {
      // Silently ignore disk cache errors
    })
  }

  /**
   * Set a value in cache with optional custom TTL (writes to memory and disk)
   */
  set<T>(type: string, data: T, customTtl?: number, ...identifiers: Array<string | number | object>): void {
    if (!this.config.enabled) {
      return
    }

    // Evict expired entries periodically
    if (this.cache.size > 0 && Math.random() < 0.1) {
      this.evictExpired()
    }

    // Evict oldest if at capacity
    this.evictOldest()

    const key = this.generateKey(type, ...identifiers)
    const ttl = customTtl || this.config.ttlByType[type as keyof typeof this.config.ttlByType] || this.config.defaultTtl

    const entry: CacheEntry<T> = {
      data,
      timestamp: Date.now(),
      ttl,
    }

    this.cache.set(key, entry)

    this.stats.sets++
    this.stats.size = this.cache.size

    // Log cache set
    logCacheEvent({
      level: 'debug',
      event: 'cache_set',
      namespace: type,
      key: identifiers.join(':'),
      ttl_ms: ttl,
      cache_size: this.cache.size,
      timestamp: new Date().toISOString(),
    })

    // Async write to disk cache (fire-and-forget)
    const diskEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED !== 'false'
    if (diskEnabled) {
      diskCacheManager.set(key, entry, ttl).catch(() => {
        // Silently ignore disk cache errors
      })
    }
  }

  /**
   * Invalidate specific cache entries by type and optional identifiers
   */
  invalidate(type: string, ...identifiers: Array<string | number | object>): void {
    const diskEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED !== 'false'

    if (identifiers.length === 0) {
      // Invalidate all entries of this type
      const pattern = `${type}:`
      let invalidatedCount = 0

      for (const key of this.cache.keys()) {
        if (key.startsWith(pattern)) {
          this.cache.delete(key)
          this.stats.evictions++
          invalidatedCount++

          // Also invalidate from disk (fire-and-forget)
          if (diskEnabled) {
            diskCacheManager.invalidate(key).catch(() => {})
          }
        }
      }

      // Log bulk invalidation
      if (invalidatedCount > 0) {
        logCacheEvent({
          level: 'debug',
          event: 'cache_invalidate',
          namespace: type,
          cache_size: this.cache.size,
          timestamp: new Date().toISOString(),
        })
      }
    } else {
      // Invalidate specific entry
      const key = this.generateKey(type, ...identifiers)
      if (this.cache.delete(key)) {
        this.stats.evictions++

        // Also invalidate from disk (fire-and-forget)
        if (diskEnabled) {
          diskCacheManager.invalidate(key).catch(() => {})
        }

        // Log specific invalidation
        logCacheEvent({
          level: 'debug',
          event: 'cache_invalidate',
          namespace: type,
          key: identifiers.join(':'),
          cache_size: this.cache.size,
          timestamp: new Date().toISOString(),
        })
      }
    }
    this.stats.size = this.cache.size
  }

  /**
   * Clear all cache entries (memory and disk)
   */
  clear(): void {
    const previousSize = this.cache.size
    this.cache.clear()
    this.stats.evictions += this.stats.size
    this.stats.size = 0

    // Also clear disk cache (fire-and-forget)
    const diskEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED !== 'false'
    if (diskEnabled) {
      diskCacheManager.clear().catch(() => {})
    }

    // Log cache clear
    if (previousSize > 0) {
      logCacheEvent({
        level: 'info',
        event: 'cache_invalidate',
        namespace: 'all',
        cache_size: 0,
        timestamp: new Date().toISOString(),
      })
    }
  }

  /**
   * Get cache statistics
   */
  getStats(): CacheStats {
    return { ...this.stats }
  }

  /**
   * Get cache hit rate
   */
  getHitRate(): number {
    const total = this.stats.hits + this.stats.misses
    return total > 0 ? this.stats.hits / total : 0
  }

  /**
   * Check if cache is enabled
   */
  isEnabled(): boolean {
    return this.config.enabled
  }

  /**
   * Get current configuration
   */
  getConfig(): CacheConfig {
    return { ...this.config }
  }
}

// Singleton instance
export const cacheManager = new CacheManager()
