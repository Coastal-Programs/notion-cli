/**
 * Workspace Cache Utility
 *
 * Manages persistent caching of workspace databases for fast name-to-ID resolution.
 * Cache is stored at ~/.notion-cli/databases.json
 */
import { GetDataSourceResponse, DataSourceObjectResponse } from '@notionhq/client/build/src/api-endpoints';
export interface CachedDatabase {
    id: string;
    title: string;
    titleNormalized: string;
    aliases: string[];
    url?: string;
    lastEditedTime?: string;
    properties?: Record<string, any>;
}
export interface WorkspaceCache {
    version: string;
    lastSync: string;
    databases: CachedDatabase[];
}
/**
 * Get the cache directory path
 */
export declare function getCacheDir(): string;
/**
 * Get the cache file path
 */
export declare function getCachePath(): Promise<string>;
/**
 * Ensure cache directory exists
 */
export declare function ensureCacheDir(): Promise<void>;
/**
 * Load cache from disk
 * Returns null if cache doesn't exist or is corrupted
 */
export declare function loadCache(): Promise<WorkspaceCache | null>;
/**
 * Save cache to disk (atomic write)
 */
export declare function saveCache(data: WorkspaceCache): Promise<void>;
/**
 * Generate search aliases from a database title
 *
 * @example
 * generateAliases('Tasks Database')
 * // Returns: ['tasks database', 'tasks', 'task', 'tasks db', 'task db', 'td']
 */
export declare function generateAliases(title: string): string[];
/**
 * Build a cached database entry from a data source response
 */
export declare function buildCacheEntry(dataSource: GetDataSourceResponse | DataSourceObjectResponse): CachedDatabase;
/**
 * Create an empty cache
 */
export declare function createEmptyCache(): WorkspaceCache;
