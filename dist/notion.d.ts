import { Client } from '@notionhq/client';
import { CreateDatabaseParameters, QueryDataSourceResponse, GetDatabaseResponse, GetDataSourceResponse, CreateDatabaseResponse, UpdateDatabaseParameters, UpdateDataSourceParameters, GetPageParameters, CreatePageParameters, BlockObjectRequest, UpdatePageParameters, AppendBlockChildrenParameters, UpdateBlockParameters, SearchParameters } from '@notionhq/client/build/src/api-endpoints';
export declare const client: Client;
/**
 * Legacy fetchWithRetry for backward compatibility
 * @deprecated Use the enhanced retry logic from retry.ts
 */
export declare const fetchWithRetry: (fn: () => Promise<any>, retries?: number) => Promise<any>;
/**
 * Fetch all pages in a data source with pagination
 */
export declare const fetchAllPagesInDS: (databaseId: string, filter?: object | undefined) => Promise<QueryDataSourceResponse["results"]>;
/**
 * Create a database
 */
export declare const createDb: (dbProps: CreateDatabaseParameters) => Promise<CreateDatabaseResponse>;
/**
 * Update a database
 */
export declare const updateDb: (dbProps: UpdateDatabaseParameters) => Promise<GetDatabaseResponse>;
/**
 * Retrieve a database (cached)
 */
export declare const retrieveDb: (databaseId: string) => Promise<GetDatabaseResponse>;
/**
 * Retrieve a data source (cached)
 */
export declare const retrieveDataSource: (dataSourceId: string) => Promise<GetDataSourceResponse>;
/**
 * Update a data source
 */
export declare const updateDataSource: (dsProps: UpdateDataSourceParameters) => Promise<GetDataSourceResponse>;
/**
 * Retrieve a page (cached with short TTL)
 */
export declare const retrievePage: (pageProp: GetPageParameters) => Promise<import("@notionhq/client").GetPageResponse>;
/**
 * Retrieve page property
 */
export declare const retrievePageProperty: (pageId: string, propId: string) => Promise<import("@notionhq/client").GetPagePropertyResponse>;
/**
 * Create a page
 */
export declare const createPage: (pageProps: CreatePageParameters) => Promise<import("@notionhq/client").CreatePageResponse>;
/**
 * Update page properties
 */
export declare const updatePageProps: (pageParams: UpdatePageParameters) => Promise<import("@notionhq/client").UpdatePageResponse>;
/**
 * Update page content by replacing all blocks
 * To keep the same page URL, remove all blocks in the page and add new blocks
 */
export declare const updatePage: (pageId: string, blocks: BlockObjectRequest[]) => Promise<import("@notionhq/client").AppendBlockChildrenResponse>;
/**
 * Retrieve a block (cached with very short TTL)
 */
export declare const retrieveBlock: (blockId: string) => Promise<import("@notionhq/client").GetBlockResponse>;
/**
 * Update a block
 */
export declare const updateBlock: (params: UpdateBlockParameters) => Promise<import("@notionhq/client").UpdateBlockResponse>;
/**
 * Retrieve block children (cached with very short TTL)
 */
export declare const retrieveBlockChildren: (blockId: string) => Promise<import("@notionhq/client").ListBlockChildrenResponse>;
/**
 * Append block children
 */
export declare const appendBlockChildren: (params: AppendBlockChildrenParameters) => Promise<import("@notionhq/client").AppendBlockChildrenResponse>;
/**
 * Delete a block
 */
export declare const deleteBlock: (blockId: string) => Promise<import("@notionhq/client").DeleteBlockResponse>;
/**
 * Retrieve a user (cached with long TTL)
 */
export declare const retrieveUser: (userId: string) => Promise<import("@notionhq/client").UserObjectResponse>;
/**
 * List all users (cached with long TTL)
 */
export declare const listUser: () => Promise<import("@notionhq/client").ListUsersResponse>;
/**
 * Get bot user info (cached with long TTL)
 */
export declare const botUser: () => Promise<import("@notionhq/client").UserObjectResponse>;
/**
 * Search for databases (cached with medium TTL)
 */
export declare const searchDb: () => Promise<(import("@notionhq/client").DataSourceObjectResponse | import("@notionhq/client").PageObjectResponse | import("@notionhq/client").PartialDataSourceObjectResponse | import("@notionhq/client").PartialPageObjectResponse)[]>;
/**
 * General search (not cached due to variable parameters)
 */
export declare const search: (params: SearchParameters) => Promise<import("@notionhq/client").SearchResponse>;
/**
 * Export cache manager for external use
 */
export { cacheManager } from './cache';
/**
 * Export retry utilities for external use
 */
export { fetchWithRetry as enhancedFetchWithRetry, CircuitBreaker } from './retry';
/**
 * Recursively retrieve a page with all its blocks and nested content
 * @param pageId - The ID of the page to retrieve
 * @param depth - Current recursion depth (internal use)
 * @param maxDepth - Maximum depth to recurse (default: 3)
 * @returns Object containing page metadata, blocks, and optional warnings
 */
export declare const retrievePageRecursive: (pageId: string, depth?: number, maxDepth?: number) => Promise<{
    page: any;
    blocks: any[];
    warnings?: Array<{
        block_id: string;
        type: string;
        notion_type?: string;
        message: string;
        has_children: boolean;
    }>;
}>;
/**
 * Map page structure (fast page discovery with parallel fetching)
 * Returns minimal structure info (titles, types, IDs) instead of full content
 * @param pageId - The ID of the page to map
 * @returns Object containing page ID, title, icon, and structure overview
 */
export declare const mapPageStructure: (pageId: string) => Promise<{
    id: string;
    title: string;
    type: string;
    icon?: string;
    structure: Array<{
        type: string;
        id: string;
        title?: string;
        text?: string;
    }>;
}>;
