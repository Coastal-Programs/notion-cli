"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.CircuitBreaker = exports.enhancedFetchWithRetry = exports.cacheManager = exports.search = exports.searchDb = exports.botUser = exports.listUser = exports.retrieveUser = exports.deleteBlock = exports.appendBlockChildren = exports.retrieveBlockChildren = exports.updateBlock = exports.retrieveBlock = exports.updatePage = exports.updatePageProps = exports.createPage = exports.retrievePageProperty = exports.retrievePage = exports.updateDataSource = exports.retrieveDataSource = exports.retrieveDb = exports.updateDb = exports.createDb = exports.fetchAllPagesInDS = exports.fetchWithRetry = exports.client = void 0;
const client_1 = require("@notionhq/client");
const cache_1 = require("./cache");
const retry_1 = require("./retry");
exports.client = new client_1.Client({
    auth: process.env.NOTION_TOKEN,
    logLevel: process.env.DEBUG ? client_1.LogLevel.DEBUG : null,
});
/**
 * Legacy fetchWithRetry for backward compatibility
 * @deprecated Use the enhanced retry logic from retry.ts
 */
const fetchWithRetry = async (fn, retries = 3) => {
    return (0, retry_1.fetchWithRetry)(fn, {
        config: { maxRetries: retries },
    });
};
exports.fetchWithRetry = fetchWithRetry;
/**
 * Cached wrapper for API calls with retry logic
 */
async function cachedFetch(cacheType, cacheKey, fetchFn, options = {}) {
    const { cacheTtl, skipCache = false, retryConfig } = options;
    // Check cache first (unless skipped or cache disabled)
    if (!skipCache) {
        const cached = cache_1.cacheManager.get(cacheType, cacheKey);
        if (cached !== null) {
            if (process.env.DEBUG) {
                console.log(`Cache HIT: ${cacheType}:${cacheKey}`);
            }
            return cached;
        }
        if (process.env.DEBUG) {
            console.log(`Cache MISS: ${cacheType}:${cacheKey}`);
        }
    }
    // Fetch with retry logic
    const data = await (0, retry_1.fetchWithRetry)(fetchFn, {
        config: retryConfig,
        context: `${cacheType}:${cacheKey}`,
    });
    // Store in cache
    if (!skipCache) {
        cache_1.cacheManager.set(cacheType, data, cacheTtl, cacheKey);
    }
    return data;
}
/**
 * Fetch all pages in a data source with pagination
 */
const fetchAllPagesInDS = async (databaseId, filter) => {
    const f = filter;
    const pages = [];
    let cursor = undefined;
    while (true) {
        const { results, next_cursor } = await (0, retry_1.fetchWithRetry)(() => exports.client.dataSources.query({
            data_source_id: databaseId,
            filter: f,
            start_cursor: cursor,
        }), { context: `fetchAllPagesInDS:${databaseId}` });
        pages.push(...results);
        if (!next_cursor) {
            break;
        }
        cursor = next_cursor;
    }
    return pages;
};
exports.fetchAllPagesInDS = fetchAllPagesInDS;
/**
 * Create a database
 */
const createDb = async (dbProps) => {
    const result = await (0, retry_1.fetchWithRetry)(() => exports.client.databases.create(dbProps), { context: 'createDb' });
    // Invalidate database list cache
    cache_1.cacheManager.invalidate('search');
    return result;
};
exports.createDb = createDb;
/**
 * Update a database
 */
const updateDb = async (dbProps) => {
    const result = await (0, retry_1.fetchWithRetry)(() => exports.client.databases.update(dbProps), { context: `updateDb:${dbProps.database_id}` });
    // Invalidate this database's cache
    cache_1.cacheManager.invalidate('database', dbProps.database_id);
    cache_1.cacheManager.invalidate('dataSource', dbProps.database_id);
    return result;
};
exports.updateDb = updateDb;
/**
 * Retrieve a database (cached)
 */
const retrieveDb = async (databaseId) => {
    return cachedFetch('database', databaseId, () => exports.client.databases.retrieve({ database_id: databaseId }));
};
exports.retrieveDb = retrieveDb;
/**
 * Retrieve a data source (cached)
 */
const retrieveDataSource = async (dataSourceId) => {
    return cachedFetch('dataSource', dataSourceId, () => exports.client.dataSources.retrieve({ data_source_id: dataSourceId }));
};
exports.retrieveDataSource = retrieveDataSource;
/**
 * Update a data source
 */
const updateDataSource = async (dsProps) => {
    const result = await (0, retry_1.fetchWithRetry)(() => exports.client.dataSources.update(dsProps), { context: `updateDataSource:${dsProps.data_source_id}` });
    // Invalidate this data source's cache
    cache_1.cacheManager.invalidate('dataSource', dsProps.data_source_id);
    return result;
};
exports.updateDataSource = updateDataSource;
/**
 * Retrieve a page (cached with short TTL)
 */
const retrievePage = async (pageProp) => {
    return cachedFetch('page', pageProp.page_id, () => exports.client.pages.retrieve(pageProp));
};
exports.retrievePage = retrievePage;
/**
 * Retrieve page property
 */
const retrievePageProperty = async (pageId, propId) => {
    return (0, retry_1.fetchWithRetry)(() => exports.client.pages.properties.retrieve({
        page_id: pageId,
        property_id: propId,
    }), { context: `retrievePageProperty:${pageId}:${propId}` });
};
exports.retrievePageProperty = retrievePageProperty;
/**
 * Create a page
 */
const createPage = async (pageProps) => {
    const result = await (0, retry_1.fetchWithRetry)(() => exports.client.pages.create(pageProps), { context: 'createPage' });
    // Invalidate parent database/page cache
    if ('parent' in pageProps && 'database_id' in pageProps.parent) {
        cache_1.cacheManager.invalidate('dataSource', pageProps.parent.database_id);
    }
    return result;
};
exports.createPage = createPage;
/**
 * Update page properties
 */
const updatePageProps = async (pageParams) => {
    const result = await (0, retry_1.fetchWithRetry)(() => exports.client.pages.update(pageParams), { context: `updatePageProps:${pageParams.page_id}` });
    // Invalidate this page's cache
    cache_1.cacheManager.invalidate('page', pageParams.page_id);
    return result;
};
exports.updatePageProps = updatePageProps;
/**
 * Update page content by replacing all blocks
 * To keep the same page URL, remove all blocks in the page and add new blocks
 */
const updatePage = async (pageId, blocks) => {
    // Get all blocks
    const blks = await (0, retry_1.fetchWithRetry)(() => exports.client.blocks.children.list({ block_id: pageId }), { context: `updatePage:list:${pageId}` });
    // Delete all blocks
    for (const blk of blks.results) {
        await (0, retry_1.fetchWithRetry)(() => exports.client.blocks.delete({ block_id: blk.id }), { context: `updatePage:delete:${blk.id}` });
    }
    // Append new blocks
    const res = await (0, retry_1.fetchWithRetry)(() => exports.client.blocks.children.append({
        block_id: pageId,
        // @ts-ignore
        children: blocks,
    }), { context: `updatePage:append:${pageId}` });
    // Invalidate caches
    cache_1.cacheManager.invalidate('page', pageId);
    cache_1.cacheManager.invalidate('block', pageId);
    return res;
};
exports.updatePage = updatePage;
/**
 * Retrieve a block (cached with very short TTL)
 */
const retrieveBlock = async (blockId) => {
    return cachedFetch('block', blockId, () => exports.client.blocks.retrieve({ block_id: blockId }));
};
exports.retrieveBlock = retrieveBlock;
/**
 * Update a block
 */
const updateBlock = async (params) => {
    const result = await (0, retry_1.fetchWithRetry)(() => exports.client.blocks.update(params), { context: `updateBlock:${params.block_id}` });
    // Invalidate this block's cache
    cache_1.cacheManager.invalidate('block', params.block_id);
    return result;
};
exports.updateBlock = updateBlock;
/**
 * Retrieve block children (cached with very short TTL)
 */
const retrieveBlockChildren = async (blockId) => {
    return cachedFetch('block', `${blockId}:children`, () => exports.client.blocks.children.list({ block_id: blockId }));
};
exports.retrieveBlockChildren = retrieveBlockChildren;
/**
 * Append block children
 */
const appendBlockChildren = async (params) => {
    const result = await (0, retry_1.fetchWithRetry)(() => exports.client.blocks.children.append(params), { context: `appendBlockChildren:${params.block_id}` });
    // Invalidate parent block's cache
    cache_1.cacheManager.invalidate('block', params.block_id);
    cache_1.cacheManager.invalidate('block', `${params.block_id}:children`);
    return result;
};
exports.appendBlockChildren = appendBlockChildren;
/**
 * Delete a block
 */
const deleteBlock = async (blockId) => {
    const result = await (0, retry_1.fetchWithRetry)(() => exports.client.blocks.delete({ block_id: blockId }), { context: `deleteBlock:${blockId}` });
    // Invalidate this block's cache
    cache_1.cacheManager.invalidate('block', blockId);
    return result;
};
exports.deleteBlock = deleteBlock;
/**
 * Retrieve a user (cached with long TTL)
 */
const retrieveUser = async (userId) => {
    return cachedFetch('user', userId, () => exports.client.users.retrieve({ user_id: userId }));
};
exports.retrieveUser = retrieveUser;
/**
 * List all users (cached with long TTL)
 */
const listUser = async () => {
    return cachedFetch('user', 'list', () => exports.client.users.list({}));
};
exports.listUser = listUser;
/**
 * Get bot user info (cached with long TTL)
 */
const botUser = async () => {
    return cachedFetch('user', 'me', () => exports.client.users.me({}));
};
exports.botUser = botUser;
/**
 * Search for databases (cached with medium TTL)
 */
const searchDb = async () => {
    const { results } = await cachedFetch('search', 'databases', async () => {
        return await exports.client.search({
            filter: {
                value: 'data_source',
                property: 'object',
            },
        });
    });
    return results;
};
exports.searchDb = searchDb;
/**
 * General search (not cached due to variable parameters)
 */
const search = async (params) => {
    return (0, retry_1.fetchWithRetry)(() => exports.client.search(params), { context: 'search' });
};
exports.search = search;
/**
 * Export cache manager for external use
 */
var cache_2 = require("./cache");
Object.defineProperty(exports, "cacheManager", { enumerable: true, get: function () { return cache_2.cacheManager; } });
/**
 * Export retry utilities for external use
 */
var retry_2 = require("./retry");
Object.defineProperty(exports, "enhancedFetchWithRetry", { enumerable: true, get: function () { return retry_2.fetchWithRetry; } });
Object.defineProperty(exports, "CircuitBreaker", { enumerable: true, get: function () { return retry_2.CircuitBreaker; } });
