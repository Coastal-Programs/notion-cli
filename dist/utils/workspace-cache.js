"use strict";
/**
 * Workspace Cache Utility
 *
 * Manages persistent caching of workspace databases for fast name-to-ID resolution.
 * Cache is stored at ~/.notion-cli/databases.json
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.getCacheDir = getCacheDir;
exports.getCachePath = getCachePath;
exports.ensureCacheDir = ensureCacheDir;
exports.loadCache = loadCache;
exports.saveCache = saveCache;
exports.generateAliases = generateAliases;
exports.buildCacheEntry = buildCacheEntry;
exports.createEmptyCache = createEmptyCache;
const fs = require("fs/promises");
const path = require("path");
const os = require("os");
const client_1 = require("@notionhq/client");
const CACHE_VERSION = '1.0.0';
const CACHE_DIR_NAME = '.notion-cli';
const CACHE_FILE_NAME = 'databases.json';
/**
 * Get the cache directory path
 */
function getCacheDir() {
    return path.join(os.homedir(), CACHE_DIR_NAME);
}
/**
 * Get the cache file path
 */
async function getCachePath() {
    return path.join(getCacheDir(), CACHE_FILE_NAME);
}
/**
 * Ensure cache directory exists
 */
async function ensureCacheDir() {
    const cacheDir = getCacheDir();
    try {
        await fs.mkdir(cacheDir, { recursive: true });
    }
    catch (error) {
        if (error.code !== 'EEXIST') {
            throw new Error(`Failed to create cache directory: ${error.message}`);
        }
    }
}
/**
 * Load cache from disk
 * Returns null if cache doesn't exist or is corrupted
 */
async function loadCache() {
    try {
        const cachePath = await getCachePath();
        const content = await fs.readFile(cachePath, 'utf-8');
        const cache = JSON.parse(content);
        // Validate cache structure
        if (!cache.version || !Array.isArray(cache.databases)) {
            console.warn('Cache file is corrupted, will rebuild on next sync');
            return null;
        }
        return cache;
    }
    catch (error) {
        if (error.code === 'ENOENT') {
            // Cache doesn't exist yet
            return null;
        }
        // Parse error or other error
        console.warn(`Failed to load cache: ${error.message}`);
        return null;
    }
}
/**
 * Save cache to disk (atomic write)
 */
async function saveCache(data) {
    await ensureCacheDir();
    const cachePath = await getCachePath();
    const tmpPath = `${cachePath}.tmp`;
    try {
        // Write to temporary file
        await fs.writeFile(tmpPath, JSON.stringify(data, null, 2), 'utf-8');
        // Atomic rename (replaces old file)
        await fs.rename(tmpPath, cachePath);
    }
    catch (error) {
        // Clean up temp file if it exists
        try {
            await fs.unlink(tmpPath);
        }
        catch {
            // Intentionally empty - cache directory may not exist
        }
        throw new Error(`Failed to save cache: ${error.message}`);
    }
}
/**
 * Generate search aliases from a database title
 *
 * @example
 * generateAliases('Tasks Database')
 * // Returns: ['tasks database', 'tasks', 'task', 'tasks db', 'task db', 'td']
 */
function generateAliases(title) {
    const aliases = new Set();
    const normalized = title.toLowerCase().trim();
    // Add full title (normalized)
    aliases.add(normalized);
    // Add title without common suffixes
    const withoutSuffixes = normalized.replace(/\s+(database|db|table|list|tracker|log)$/i, '');
    if (withoutSuffixes !== normalized) {
        aliases.add(withoutSuffixes);
    }
    // Add title with common suffixes
    if (withoutSuffixes !== normalized) {
        aliases.add(`${withoutSuffixes} db`);
        aliases.add(`${withoutSuffixes} database`);
    }
    // Add singular/plural variants
    if (withoutSuffixes.endsWith('s')) {
        aliases.add(withoutSuffixes.slice(0, -1)); // Remove 's'
    }
    else {
        aliases.add(`${withoutSuffixes}s`); // Add 's'
    }
    // Add acronym if multi-word (e.g., "Meeting Notes" → "mn")
    const words = withoutSuffixes.split(/\s+/);
    if (words.length > 1) {
        const acronym = words.map(w => w[0]).join('');
        if (acronym.length >= 2) {
            aliases.add(acronym);
        }
    }
    return Array.from(aliases);
}
/**
 * Build a cached database entry from a data source response
 */
function buildCacheEntry(dataSource) {
    let title = 'Untitled';
    let properties = {};
    let url;
    let lastEditedTime;
    if ((0, client_1.isFullDataSource)(dataSource)) {
        if (dataSource.title && dataSource.title.length > 0) {
            title = dataSource.title[0].plain_text;
        }
        // Extract property schema
        if (dataSource.properties) {
            properties = dataSource.properties;
        }
        // Extract URL if available
        if ('url' in dataSource) {
            url = dataSource.url;
        }
        // Extract last edited time
        if ('last_edited_time' in dataSource) {
            lastEditedTime = dataSource.last_edited_time;
        }
    }
    const titleNormalized = title.toLowerCase().trim();
    const aliases = generateAliases(title);
    return {
        id: dataSource.id,
        title,
        titleNormalized,
        aliases,
        url,
        lastEditedTime,
        properties,
    };
}
/**
 * Create an empty cache
 */
function createEmptyCache() {
    return {
        version: CACHE_VERSION,
        lastSync: new Date().toISOString(),
        databases: [],
    };
}
