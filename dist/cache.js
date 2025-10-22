"use strict";
/**
 * Simple in-memory caching layer for Notion API responses
 * Supports TTL (time-to-live) and cache invalidation
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.cacheManager = exports.CacheManager = void 0;
class CacheManager {
    constructor(config) {
        this.cache = new Map();
        this.stats = {
            hits: 0,
            misses: 0,
            sets: 0,
            evictions: 0,
            size: 0,
        };
        // Default configuration
        this.config = {
            enabled: process.env.NOTION_CLI_CACHE_ENABLED !== 'false',
            defaultTtl: parseInt(process.env.NOTION_CLI_CACHE_TTL || '300000', 10),
            maxSize: parseInt(process.env.NOTION_CLI_CACHE_MAX_SIZE || '1000', 10),
            ttlByType: {
                dataSource: parseInt(process.env.NOTION_CLI_CACHE_DS_TTL || '600000', 10),
                database: parseInt(process.env.NOTION_CLI_CACHE_DB_TTL || '600000', 10),
                user: parseInt(process.env.NOTION_CLI_CACHE_USER_TTL || '3600000', 10),
                page: parseInt(process.env.NOTION_CLI_CACHE_PAGE_TTL || '60000', 10),
                block: parseInt(process.env.NOTION_CLI_CACHE_BLOCK_TTL || '30000', 10), // 30 sec
            },
            ...config,
        };
    }
    /**
     * Generate a cache key from resource type and identifiers
     */
    generateKey(type, ...identifiers) {
        return `${type}:${identifiers.map(id => typeof id === 'object' ? JSON.stringify(id) : String(id)).join(':')}`;
    }
    /**
     * Check if a cache entry is still valid
     */
    isValid(entry) {
        const now = Date.now();
        return now - entry.timestamp < entry.ttl;
    }
    /**
     * Evict expired entries
     */
    evictExpired() {
        const now = Date.now();
        for (const [key, entry] of this.cache.entries()) {
            if (!this.isValid(entry)) {
                this.cache.delete(key);
                this.stats.evictions++;
            }
        }
        this.stats.size = this.cache.size;
    }
    /**
     * Evict oldest entries if cache is full
     */
    evictOldest() {
        if (this.cache.size >= this.config.maxSize) {
            // Find and remove oldest entry
            let oldestKey = null;
            let oldestTime = Infinity;
            for (const [key, entry] of this.cache.entries()) {
                if (entry.timestamp < oldestTime) {
                    oldestTime = entry.timestamp;
                    oldestKey = key;
                }
            }
            if (oldestKey) {
                this.cache.delete(oldestKey);
                this.stats.evictions++;
            }
        }
    }
    /**
     * Get a value from cache
     */
    get(type, ...identifiers) {
        if (!this.config.enabled) {
            return null;
        }
        const key = this.generateKey(type, ...identifiers);
        const entry = this.cache.get(key);
        if (!entry) {
            this.stats.misses++;
            return null;
        }
        if (!this.isValid(entry)) {
            this.cache.delete(key);
            this.stats.misses++;
            this.stats.evictions++;
            return null;
        }
        this.stats.hits++;
        return entry.data;
    }
    /**
     * Set a value in cache with optional custom TTL
     */
    set(type, data, customTtl, ...identifiers) {
        if (!this.config.enabled) {
            return;
        }
        // Evict expired entries periodically
        if (this.cache.size > 0 && Math.random() < 0.1) {
            this.evictExpired();
        }
        // Evict oldest if at capacity
        this.evictOldest();
        const key = this.generateKey(type, ...identifiers);
        const ttl = customTtl || this.config.ttlByType[type] || this.config.defaultTtl;
        this.cache.set(key, {
            data,
            timestamp: Date.now(),
            ttl,
        });
        this.stats.sets++;
        this.stats.size = this.cache.size;
    }
    /**
     * Invalidate specific cache entries by type and optional identifiers
     */
    invalidate(type, ...identifiers) {
        if (identifiers.length === 0) {
            // Invalidate all entries of this type
            const pattern = `${type}:`;
            for (const key of this.cache.keys()) {
                if (key.startsWith(pattern)) {
                    this.cache.delete(key);
                    this.stats.evictions++;
                }
            }
        }
        else {
            // Invalidate specific entry
            const key = this.generateKey(type, ...identifiers);
            if (this.cache.delete(key)) {
                this.stats.evictions++;
            }
        }
        this.stats.size = this.cache.size;
    }
    /**
     * Clear all cache entries
     */
    clear() {
        this.cache.clear();
        this.stats.evictions += this.stats.size;
        this.stats.size = 0;
    }
    /**
     * Get cache statistics
     */
    getStats() {
        return { ...this.stats };
    }
    /**
     * Get cache hit rate
     */
    getHitRate() {
        const total = this.stats.hits + this.stats.misses;
        return total > 0 ? this.stats.hits / total : 0;
    }
    /**
     * Check if cache is enabled
     */
    isEnabled() {
        return this.config.enabled;
    }
    /**
     * Get current configuration
     */
    getConfig() {
        return { ...this.config };
    }
}
exports.CacheManager = CacheManager;
// Singleton instance
exports.cacheManager = new CacheManager();
