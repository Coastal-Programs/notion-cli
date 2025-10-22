"use strict";
/**
 * Notion URL Parser
 *
 * Extracts clean Notion IDs from various input formats:
 * - Full URLs: https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00?v=...
 * - Short URLs: notion.so/1fb79d4c71bb8032b722c82305b63a00
 * - Raw IDs with dashes: 1fb79d4c-71bb-8032-b722-c82305b63a00
 * - Raw IDs without dashes: 1fb79d4c71bb8032b722c82305b63a00
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.isValidNotionId = exports.isNotionUrl = exports.extractNotionId = void 0;
/**
 * Extract Notion ID from URL or raw ID
 *
 * @param input - Full Notion URL, partial URL, or raw ID
 * @returns Clean Notion ID (32 hex characters without dashes)
 * @throws Error if input is invalid
 *
 * @example
 * // Full URL
 * extractNotionId('https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00?v=...')
 * // Returns: '1fb79d4c71bb8032b722c82305b63a00'
 *
 * @example
 * // Raw ID with dashes
 * extractNotionId('1fb79d4c-71bb-8032-b722-c82305b63a00')
 * // Returns: '1fb79d4c71bb8032b722c82305b63a00'
 *
 * @example
 * // Already clean ID
 * extractNotionId('1fb79d4c71bb8032b722c82305b63a00')
 * // Returns: '1fb79d4c71bb8032b722c82305b63a00'
 */
function extractNotionId(input) {
    if (!input || typeof input !== 'string') {
        throw new Error('Input must be a non-empty string');
    }
    const trimmed = input.trim();
    // Check if it's a URL (contains notion.so or http)
    if (trimmed.includes('notion.so') || trimmed.includes('http')) {
        return extractIdFromUrl(trimmed);
    }
    // Not a URL, treat as raw ID
    return cleanRawId(trimmed);
}
exports.extractNotionId = extractNotionId;
/**
 * Extract ID from Notion URL
 */
function extractIdFromUrl(url) {
    // Notion URL patterns:
    // https://www.notion.so/{id}
    // https://www.notion.so/{id}?v={view_id}
    // https://notion.so/{id}
    // www.notion.so/{id}
    // Match notion.so/ followed by hex characters and optional dashes
    const match = url.match(/notion\.so\/([a-f0-9-]{32,36})/i);
    if (match) {
        return cleanRawId(match[1]);
    }
    throw new Error(`Could not extract Notion ID from URL: ${url}\n\n` +
        `Expected format: https://www.notion.so/{id}\n` +
        `Example: https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00`);
}
/**
 * Clean raw ID by removing dashes and validating format
 */
function cleanRawId(id) {
    // Remove all dashes
    const cleaned = id.replace(/-/g, '');
    // Validate: must be exactly 32 hex characters
    if (!/^[a-f0-9]{32}$/i.test(cleaned)) {
        throw new Error(`Invalid Notion ID format: ${id}\n\n` +
            `Expected: 32 hexadecimal characters (with or without dashes)\n` +
            `Example: 1fb79d4c71bb8032b722c82305b63a00\n` +
            `Example: 1fb79d4c-71bb-8032-b722-c82305b63a00`);
    }
    return cleaned.toLowerCase();
}
/**
 * Check if a string looks like a Notion URL
 *
 * @param input - String to check
 * @returns True if input appears to be a Notion URL
 */
function isNotionUrl(input) {
    if (!input || typeof input !== 'string') {
        return false;
    }
    return input.includes('notion.so');
}
exports.isNotionUrl = isNotionUrl;
/**
 * Check if a string looks like a valid Notion ID
 *
 * @param input - String to check
 * @returns True if input appears to be a valid Notion ID
 */
function isValidNotionId(input) {
    if (!input || typeof input !== 'string') {
        return false;
    }
    try {
        extractNotionId(input);
        return true;
    }
    catch {
        return false;
    }
}
exports.isValidNotionId = isValidNotionId;
