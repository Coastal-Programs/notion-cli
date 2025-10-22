/**
 * Simple in-memory caching layer for Notion API responses
 * Supports TTL (time-to-live) and cache invalidation
 */

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

export class CacheManager {
  private cache: Map<string, CacheEntry<any>>
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
  private generateKey(type: string, ...identifiers: any[]): string {
    return `${type}:${identifiers.map(id =>
      typeof id === 'object' ? JSON.stringify(id) : String(id)
    ).join(':')}`
  }

  /**
   * Check if a cache entry is still valid
   */
  private isValid(entry: CacheEntry<any>): boolean {
    const now = Date.now()
    return now - entry.timestamp < entry.ttl
  }

  /**
   * Evict expired entries
   */
  private evictExpired(): void {
    const now = Date.now()
    for (const [key, entry] of this.cache.entries()) {
      if (!this.isValid(entry)) {
        this.cache.delete(key)
        this.stats.evictions++
      }
    }
    this.stats.size = this.cache.size
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
      }
    }
  }

  /**
   * Get a value from cache
   */
  get<T>(type: string, ...identifiers: any[]): T | null {
    if (!this.config.enabled) {
      return null
    }

    const key = this.generateKey(type, ...identifiers)
    const entry = this.cache.get(key)

    if (!entry) {
      this.stats.misses++
      return null
    }

    if (!this.isValid(entry)) {
      this.cache.delete(key)
      this.stats.misses++
      this.stats.evictions++
      return null
    }

    this.stats.hits++
    return entry.data as T
  }

  /**
   * Set a value in cache with optional custom TTL
   */
  set<T>(type: string, data: T, customTtl?: number, ...identifiers: any[]): void {
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

    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl,
    })

    this.stats.sets++
    this.stats.size = this.cache.size
  }

  /**
   * Invalidate specific cache entries by type and optional identifiers
   */
  invalidate(type: string, ...identifiers: any[]): void {
    if (identifiers.length === 0) {
      // Invalidate all entries of this type
      const pattern = `${type}:`
      for (const key of this.cache.keys()) {
        if (key.startsWith(pattern)) {
          this.cache.delete(key)
          this.stats.evictions++
        }
      }
    } else {
      // Invalidate specific entry
      const key = this.generateKey(type, ...identifiers)
      if (this.cache.delete(key)) {
        this.stats.evictions++
      }
    }
    this.stats.size = this.cache.size
  }

  /**
   * Clear all cache entries
   */
  clear(): void {
    this.cache.clear()
    this.stats.evictions += this.stats.size
    this.stats.size = 0
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
