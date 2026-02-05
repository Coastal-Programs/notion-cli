"use strict";
/**
 * Simple in-memory caching layer for Notion API responses
 * Supports TTL (time-to-live) and cache invalidation
 * Integrated with disk cache for persistence across CLI invocations
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.cacheManager = exports.CacheManager = void 0;
const disk_cache_1 = require("./utils/disk-cache");
/**
 * Check if verbose logging is enabled
 */
function isVerboseEnabled() {
    return process.env.DEBUG === 'true' ||
        process.env.NOTION_CLI_DEBUG === 'true' ||
        process.env.NOTION_CLI_VERBOSE === 'true';
}
/**
 * Log structured cache event to stderr
 * Never pollutes stdout - safe for JSON output
 */
function logCacheEvent(event) {
    // Only log if verbose mode is enabled
    if (!isVerboseEnabled()) {
        return;
    }
    // Always write to stderr, never stdout
    console.error(JSON.stringify(event));
}
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
        let evictedCount = 0;
        for (const [key, entry] of this.cache.entries()) {
            if (!this.isValid(entry)) {
                this.cache.delete(key);
                this.stats.evictions++;
                evictedCount++;
            }
        }
        this.stats.size = this.cache.size;
        // Log eviction event if any entries were evicted
        if (evictedCount > 0 && isVerboseEnabled()) {
            logCacheEvent({
                level: 'debug',
                event: 'cache_evict',
                namespace: 'expired',
                cache_size: this.cache.size,
                timestamp: new Date().toISOString(),
            });
        }
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
                // Log LRU eviction
                logCacheEvent({
                    level: 'debug',
                    event: 'cache_evict',
                    namespace: 'lru',
                    key: oldestKey,
                    cache_size: this.cache.size,
                    timestamp: new Date().toISOString(),
                });
            }
        }
    }
    /**
     * Get a value from cache (checks memory, then disk)
     */
    async get(type, ...identifiers) {
        if (!this.config.enabled) {
            return null;
        }
        const key = this.generateKey(type, ...identifiers);
        const entry = this.cache.get(key);
        // Check memory cache first
        if (entry && this.isValid(entry)) {
            this.stats.hits++;
            // Log cache hit
            logCacheEvent({
                level: 'debug',
                event: 'cache_hit',
                namespace: type,
                key: identifiers.join(':'),
                age_ms: Date.now() - entry.timestamp,
                ttl_ms: entry.ttl,
                timestamp: new Date().toISOString(),
            });
            return entry.data;
        }
        // Remove invalid memory entry
        if (entry) {
            this.cache.delete(key);
            this.stats.evictions++;
            // Log eviction event
            logCacheEvent({
                level: 'debug',
                event: 'cache_evict',
                namespace: type,
                key: identifiers.join(':'),
                cache_size: this.cache.size,
                timestamp: new Date().toISOString(),
            });
        }
        // Check disk cache (only if enabled)
        const diskEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED !== 'false';
        if (diskEnabled) {
            try {
                const diskEntry = await disk_cache_1.diskCacheManager.get(key);
                if (diskEntry && diskEntry.data) {
                    const entry = diskEntry.data;
                    // Validate disk entry
                    if (this.isValid(entry)) {
                        // Promote to memory cache
                        this.cache.set(key, entry);
                        this.stats.hits++;
                        // Log cache hit (from disk)
                        logCacheEvent({
                            level: 'debug',
                            event: 'cache_hit',
                            namespace: type,
                            key: identifiers.join(':'),
                            age_ms: Date.now() - entry.timestamp,
                            ttl_ms: entry.ttl,
                            timestamp: new Date().toISOString(),
                        });
                        return entry.data;
                    }
                    else {
                        // Remove expired disk entry
                        disk_cache_1.diskCacheManager.invalidate(key).catch(() => { });
                    }
                }
            }
            catch (error) {
                // Silently ignore disk cache errors
            }
        }
        // Cache miss
        this.stats.misses++;
        // Log cache miss
        logCacheEvent({
            level: 'debug',
            event: 'cache_miss',
            namespace: type,
            key: identifiers.join(':'),
            timestamp: new Date().toISOString(),
        });
        return null;
    }
    /**
     * Set a value in cache with optional custom TTL (writes to memory and disk)
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
        const entry = {
            data,
            timestamp: Date.now(),
            ttl,
        };
        this.cache.set(key, entry);
        this.stats.sets++;
        this.stats.size = this.cache.size;
        // Log cache set
        logCacheEvent({
            level: 'debug',
            event: 'cache_set',
            namespace: type,
            key: identifiers.join(':'),
            ttl_ms: ttl,
            cache_size: this.cache.size,
            timestamp: new Date().toISOString(),
        });
        // Async write to disk cache (fire-and-forget)
        const diskEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED !== 'false';
        if (diskEnabled) {
            disk_cache_1.diskCacheManager.set(key, entry, ttl).catch(() => {
                // Silently ignore disk cache errors
            });
        }
    }
    /**
     * Invalidate specific cache entries by type and optional identifiers
     */
    invalidate(type, ...identifiers) {
        const diskEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED !== 'false';
        if (identifiers.length === 0) {
            // Invalidate all entries of this type
            const pattern = `${type}:`;
            let invalidatedCount = 0;
            for (const key of this.cache.keys()) {
                if (key.startsWith(pattern)) {
                    this.cache.delete(key);
                    this.stats.evictions++;
                    invalidatedCount++;
                    // Also invalidate from disk (fire-and-forget)
                    if (diskEnabled) {
                        disk_cache_1.diskCacheManager.invalidate(key).catch(() => { });
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
                });
            }
        }
        else {
            // Invalidate specific entry
            const key = this.generateKey(type, ...identifiers);
            if (this.cache.delete(key)) {
                this.stats.evictions++;
                // Also invalidate from disk (fire-and-forget)
                if (diskEnabled) {
                    disk_cache_1.diskCacheManager.invalidate(key).catch(() => { });
                }
                // Log specific invalidation
                logCacheEvent({
                    level: 'debug',
                    event: 'cache_invalidate',
                    namespace: type,
                    key: identifiers.join(':'),
                    cache_size: this.cache.size,
                    timestamp: new Date().toISOString(),
                });
            }
        }
        this.stats.size = this.cache.size;
    }
    /**
     * Clear all cache entries (memory and disk)
     */
    clear() {
        const previousSize = this.cache.size;
        this.cache.clear();
        this.stats.evictions += this.stats.size;
        this.stats.size = 0;
        // Also clear disk cache (fire-and-forget)
        const diskEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED !== 'false';
        if (diskEnabled) {
            disk_cache_1.diskCacheManager.clear().catch(() => { });
        }
        // Log cache clear
        if (previousSize > 0) {
            logCacheEvent({
                level: 'info',
                event: 'cache_invalidate',
                namespace: 'all',
                cache_size: 0,
                timestamp: new Date().toISOString(),
            });
        }
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
