"use strict";
/**
 * Request deduplication manager
 * Ensures only one in-flight request per unique key
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.deduplicationManager = exports.DeduplicationManager = void 0;
class DeduplicationManager {
    constructor() {
        this.pending = new Map();
        this.stats = { hits: 0, misses: 0 };
    }
    /**
     * Execute a function with deduplication
     * If the same key is already in-flight, returns the existing promise
     * @param key Unique identifier for the request
     * @param fn Function to execute if no in-flight request exists
     * @returns Promise resolving to the function result
     */
    async execute(key, fn) {
        // Check for in-flight request
        const existing = this.pending.get(key);
        if (existing) {
            this.stats.hits++;
            return existing;
        }
        // Create new request
        this.stats.misses++;
        const promise = fn().finally(() => {
            this.pending.delete(key);
        });
        this.pending.set(key, promise);
        return promise;
    }
    /**
     * Get deduplication statistics
     * @returns Object containing hits, misses, and pending count
     */
    getStats() {
        return {
            ...this.stats,
            pending: this.pending.size,
        };
    }
    /**
     * Clear all pending requests and reset statistics
     */
    clear() {
        this.pending.clear();
        this.stats = { hits: 0, misses: 0 };
    }
    /**
     * Safety cleanup for stale entries
     * This should rarely be needed as promises clean themselves up
     * @param _maxAge Maximum age in milliseconds (default: 30000)
     */
    cleanup(_maxAge = 30000) {
        // Note: In practice, promises clean themselves up via finally()
        // This is a safety mechanism for edge cases
        const currentSize = this.pending.size;
        if (currentSize > 0) {
            // Log warning if cleanup is needed
            console.warn(`DeduplicationManager cleanup called with ${currentSize} pending requests`);
        }
    }
}
exports.DeduplicationManager = DeduplicationManager;
/**
 * Global singleton instance for use across the application
 */
exports.deduplicationManager = new DeduplicationManager();
