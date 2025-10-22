/**
 * Notion URL Parser
 *
 * Extracts clean Notion IDs from various input formats:
 * - Full URLs: https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00?v=...
 * - Short URLs: notion.so/1fb79d4c71bb8032b722c82305b63a00
 * - Raw IDs with dashes: 1fb79d4c-71bb-8032-b722-c82305b63a00
 * - Raw IDs without dashes: 1fb79d4c71bb8032b722c82305b63a00
 */
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
export declare function extractNotionId(input: string): string;
/**
 * Check if a string looks like a Notion URL
 *
 * @param input - String to check
 * @returns True if input appears to be a Notion URL
 */
export declare function isNotionUrl(input: string): boolean;
/**
 * Check if a string looks like a valid Notion ID
 *
 * @param input - String to check
 * @returns True if input appears to be a valid Notion ID
 */
export declare function isValidNotionId(input: string): boolean;
