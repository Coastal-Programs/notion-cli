/**
 * Disk Cache Manager
 *
 * Provides persistent caching to disk, maintaining cache across CLI invocations.
 * Cache entries are stored in ~/.notion-cli/cache/ directory.
 */
export interface DiskCacheEntry<T = any> {
    key: string;
    data: T;
    expiresAt: number;
    createdAt: number;
    size: number;
}
export interface DiskCacheStats {
    totalEntries: number;
    totalSize: number;
    oldestEntry: number | null;
    newestEntry: number | null;
}
export declare class DiskCacheManager {
    private cacheDir;
    private maxSize;
    private syncInterval;
    private dirtyKeys;
    private syncTimer;
    private initialized;
    constructor(options?: {
        cacheDir?: string;
        maxSize?: number;
        syncInterval?: number;
    });
    /**
     * Initialize disk cache (create directory, start sync timer)
     */
    initialize(): Promise<void>;
    /**
     * Get a cache entry from disk
     */
    get<T>(key: string): Promise<DiskCacheEntry<T> | null>;
    /**
     * Set a cache entry to disk
     */
    set<T>(key: string, data: T, ttl: number): Promise<void>;
    /**
     * Invalidate (delete) a cache entry
     */
    invalidate(key: string): Promise<void>;
    /**
     * Clear all cache entries
     */
    clear(): Promise<void>;
    /**
     * Sync dirty entries to disk
     */
    sync(): Promise<void>;
    /**
     * Shutdown (flush and cleanup)
     */
    shutdown(): Promise<void>;
    /**
     * Get cache statistics
     */
    getStats(): Promise<DiskCacheStats>;
    /**
     * Enforce maximum cache size by removing oldest entries
     */
    private enforceMaxSize;
    /**
     * Ensure cache directory exists
     */
    private ensureCacheDir;
    /**
     * Get file path for a cache key
     */
    private getFilePath;
}
/**
 * Global singleton instance
 */
export declare const diskCacheManager: DiskCacheManager;
