/**
 * Simple in-memory caching layer for Notion API responses
 * Supports TTL (time-to-live) and cache invalidation
 */
export interface CacheEntry<T> {
    data: T;
    timestamp: number;
    ttl: number;
}
export interface CacheStats {
    hits: number;
    misses: number;
    sets: number;
    evictions: number;
    size: number;
}
export interface CacheConfig {
    enabled: boolean;
    defaultTtl: number;
    maxSize: number;
    ttlByType: {
        dataSource: number;
        database: number;
        user: number;
        page: number;
        block: number;
    };
}
export declare class CacheManager {
    private cache;
    private stats;
    private config;
    constructor(config?: Partial<CacheConfig>);
    /**
     * Generate a cache key from resource type and identifiers
     */
    private generateKey;
    /**
     * Check if a cache entry is still valid
     */
    private isValid;
    /**
     * Evict expired entries
     */
    private evictExpired;
    /**
     * Evict oldest entries if cache is full
     */
    private evictOldest;
    /**
     * Get a value from cache
     */
    get<T>(type: string, ...identifiers: any[]): T | null;
    /**
     * Set a value in cache with optional custom TTL
     */
    set<T>(type: string, data: T, customTtl?: number, ...identifiers: any[]): void;
    /**
     * Invalidate specific cache entries by type and optional identifiers
     */
    invalidate(type: string, ...identifiers: any[]): void;
    /**
     * Clear all cache entries
     */
    clear(): void;
    /**
     * Get cache statistics
     */
    getStats(): CacheStats;
    /**
     * Get cache hit rate
     */
    getHitRate(): number;
    /**
     * Check if cache is enabled
     */
    isEnabled(): boolean;
    /**
     * Get current configuration
     */
    getConfig(): CacheConfig;
}
export declare const cacheManager: CacheManager;
