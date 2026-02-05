/**
 * Request deduplication manager
 * Ensures only one in-flight request per unique key
 */
export interface DeduplicationStats {
    hits: number;
    misses: number;
    pending: number;
}
export declare class DeduplicationManager {
    private pending;
    private stats;
    constructor();
    /**
     * Execute a function with deduplication
     * If the same key is already in-flight, returns the existing promise
     * @param key Unique identifier for the request
     * @param fn Function to execute if no in-flight request exists
     * @returns Promise resolving to the function result
     */
    execute<T>(key: string, fn: () => Promise<T>): Promise<T>;
    /**
     * Get deduplication statistics
     * @returns Object containing hits, misses, and pending count
     */
    getStats(): DeduplicationStats;
    /**
     * Clear all pending requests and reset statistics
     */
    clear(): void;
    /**
     * Safety cleanup for stale entries
     * This should rarely be needed as promises clean themselves up
     * @param _maxAge Maximum age in milliseconds (default: 30000)
     */
    cleanup(_maxAge?: number): void;
}
/**
 * Global singleton instance for use across the application
 */
export declare const deduplicationManager: DeduplicationManager;
