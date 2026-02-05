"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.mapPageStructure = exports.retrievePageRecursive = exports.CircuitBreaker = exports.enhancedFetchWithRetry = exports.cacheManager = exports.search = exports.searchDb = exports.botUser = exports.listUser = exports.retrieveUser = exports.deleteBlock = exports.appendBlockChildren = exports.retrieveBlockChildren = exports.updateBlock = exports.retrieveBlock = exports.updatePage = exports.updatePageProps = exports.createPage = exports.retrievePageProperty = exports.retrievePage = exports.updateDataSource = exports.retrieveDataSource = exports.retrieveDb = exports.updateDb = exports.createDb = exports.fetchAllPagesInDS = exports.fetchWithRetry = exports.client = void 0;
const client_1 = require("@notionhq/client");
const cache_1 = require("./cache");
const retry_1 = require("./retry");
const deduplication_1 = require("./deduplication");
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
 * Cached wrapper for API calls with retry logic and deduplication
 */
async function cachedFetch(cacheType, cacheKey, fetchFn, options = {}) {
    const { cacheTtl, skipCache = false, skipDedup = false, retryConfig } = options;
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
    // Generate deduplication key
    const dedupKey = `${cacheType}:${JSON.stringify(cacheKey)}`;
    // Wrap fetch function with deduplication (unless disabled)
    const dedupEnabled = process.env.NOTION_CLI_DEDUP_ENABLED !== 'false' && !skipDedup;
    const fetchWithDedup = dedupEnabled
        ? () => deduplication_1.deduplicationManager.execute(dedupKey, async () => {
            if (process.env.DEBUG) {
                console.log(`Dedup MISS: ${dedupKey}`);
            }
            return (0, retry_1.fetchWithRetry)(fetchFn, {
                config: retryConfig,
                context: `${cacheType}:${cacheKey}`,
            });
        })
        : () => (0, retry_1.fetchWithRetry)(fetchFn, {
            config: retryConfig,
            context: `${cacheType}:${cacheKey}`,
        });
    // Execute fetch (with or without deduplication)
    const data = await fetchWithDedup();
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
/**
 * Recursively retrieve a page with all its blocks and nested content
 * @param pageId - The ID of the page to retrieve
 * @param depth - Current recursion depth (internal use)
 * @param maxDepth - Maximum depth to recurse (default: 3)
 * @returns Object containing page metadata, blocks, and optional warnings
 */
const retrievePageRecursive = async (pageId, depth = 0, maxDepth = 3) => {
    var _a, _b;
    // Prevent infinite recursion
    if (depth >= maxDepth) {
        return {
            page: null,
            blocks: [],
            warnings: [
                {
                    block_id: pageId,
                    type: 'max_depth_reached',
                    message: `Maximum recursion depth of ${maxDepth} reached`,
                    has_children: false,
                },
            ],
        };
    }
    // Retrieve the page
    const page = await (0, exports.retrievePage)({ page_id: pageId });
    // Retrieve all blocks (children)
    const blocksResponse = await (0, exports.retrieveBlockChildren)(pageId);
    const blocks = blocksResponse.results || [];
    const warnings = [];
    // Recursively fetch nested blocks
    for (const block of blocks) {
        // Skip partial blocks
        if (!(0, client_1.isFullBlock)(block)) {
            continue;
        }
        // Handle unsupported blocks
        if (block.type === 'unsupported') {
            warnings.push({
                block_id: block.id,
                type: 'unsupported',
                notion_type: ((_a = block.unsupported) === null || _a === void 0 ? void 0 : _a.type) || 'unknown',
                message: `Block type '${((_b = block.unsupported) === null || _b === void 0 ? void 0 : _b.type) || 'unknown'}' not supported by Notion API`,
                has_children: block.has_children,
            });
            continue;
        }
        // Recursively fetch children for blocks that have them
        if (block.has_children) {
            try {
                const childrenResponse = await (0, exports.retrieveBlockChildren)(block.id);
                block.children = childrenResponse.results || [];
                // If this is a child_page block, recursively fetch that page too
                if (block.type === 'child_page' && depth + 1 < maxDepth) {
                    const childPageData = await (0, exports.retrievePageRecursive)(block.id, depth + 1, maxDepth);
                    block.child_page_details = childPageData;
                    // Merge warnings from recursive calls
                    if (childPageData.warnings) {
                        warnings.push(...childPageData.warnings);
                    }
                }
            }
            catch (error) {
                // If we can't fetch children, add a warning
                warnings.push({
                    block_id: block.id,
                    type: 'fetch_error',
                    message: `Failed to fetch children for block: ${error instanceof Error ? error.message : 'Unknown error'}`,
                    has_children: true,
                });
            }
        }
    }
    return {
        page,
        blocks,
        ...(warnings.length > 0 && { warnings }),
    };
};
exports.retrievePageRecursive = retrievePageRecursive;
/**
 * Map page structure (fast page discovery with parallel fetching)
 * Returns minimal structure info (titles, types, IDs) instead of full content
 * @param pageId - The ID of the page to map
 * @returns Object containing page ID, title, icon, and structure overview
 */
const mapPageStructure = async (pageId) => {
    // Parallel fetch: get page and blocks simultaneously
    const [page, blocksResponse] = await Promise.all([
        (0, exports.retrievePage)({ page_id: pageId }),
        (0, exports.retrieveBlockChildren)(pageId),
    ]);
    const blocks = blocksResponse.results || [];
    // Extract page title
    let pageTitle = 'Untitled';
    if (page.object === 'page' && (0, client_1.isFullPage)(page)) {
        Object.entries(page.properties).find(([, prop]) => {
            if (prop.type === 'title' && prop.title.length > 0) {
                pageTitle = prop.title[0].plain_text;
                return true;
            }
            return false;
        });
    }
    // Extract page icon
    let pageIcon;
    if ((0, client_1.isFullPage)(page) && page.icon) {
        if (page.icon.type === 'emoji') {
            pageIcon = page.icon.emoji;
        }
        else if (page.icon.type === 'external') {
            pageIcon = page.icon.external.url;
        }
        else if (page.icon.type === 'file') {
            pageIcon = page.icon.file.url;
        }
    }
    // Build minimal structure
    const structure = blocks.map((block) => {
        const structureItem = {
            type: block.type,
            id: block.id,
        };
        // Extract title/text based on block type
        try {
            switch (block.type) {
                case 'child_page':
                    structureItem.title = block[block.type].title;
                    break;
                case 'child_database':
                    structureItem.title = block[block.type].title;
                    break;
                case 'heading_1':
                case 'heading_2':
                case 'heading_3':
                case 'paragraph':
                case 'bulleted_list_item':
                case 'numbered_list_item':
                case 'to_do':
                case 'toggle':
                case 'quote':
                case 'callout':
                case 'code':
                    if (block[block.type].rich_text && block[block.type].rich_text.length > 0) {
                        structureItem.text = block[block.type].rich_text[0].plain_text;
                    }
                    break;
                case 'bookmark':
                case 'embed':
                case 'link_preview':
                    structureItem.text = block[block.type].url;
                    break;
                case 'equation':
                    structureItem.text = block[block.type].expression;
                    break;
                case 'image':
                case 'file':
                case 'video':
                case 'pdf':
                    if (block[block.type].type === 'file') {
                        structureItem.text = block[block.type].file.url;
                    }
                    else if (block[block.type].type === 'external') {
                        structureItem.text = block[block.type].external.url;
                    }
                    break;
                // For other types, just include type and id
                default:
                    break;
            }
        }
        catch {
            // If extraction fails, just include type and id
        }
        return structureItem;
    });
    return {
        id: pageId,
        title: pageTitle,
        type: 'page',
        ...(pageIcon && { icon: pageIcon }),
        structure,
    };
};
exports.mapPageStructure = mapPageStructure;
