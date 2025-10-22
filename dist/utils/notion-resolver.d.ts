/**
 * Notion ID Resolver
 *
 * Hybrid resolution system that supports:
 * - URLs: https://www.notion.so/database-id
 * - Direct IDs: database-id
 * - Names: "Tasks Database" (via cache lookup and API fallback)
 * - Smart database_id → data_source_id conversion
 *
 * Resolution stages:
 * 1. URL extraction
 * 2. Direct ID validation
 * 3. Cache lookup (exact + aliases)
 * 4. API search fallback
 * 5. Smart database_id → data_source_id resolution (for databases)
 */
/**
 * Resolve Notion input (URL, ID, or name) to a clean Notion ID
 *
 * Supports URLs, IDs, and name-based lookups via cache and API search.
 * For databases, automatically detects and converts database_id to data_source_id.
 *
 * @param input - Database/page name, ID, or URL
 * @param type - Resource type (for better error messages)
 * @returns Clean Notion ID (32 hex characters without dashes)
 * @throws NotionCLIError if input cannot be resolved
 *
 * @example
 * // URL
 * await resolveNotionId('https://notion.so/1fb79d4c71bb8032b722c82305b63a00')
 * // Returns: '1fb79d4c71bb8032b722c82305b63a00'
 *
 * @example
 * // Direct ID
 * await resolveNotionId('1fb79d4c71bb8032b722c82305b63a00')
 * // Returns: '1fb79d4c71bb8032b722c82305b63a00'
 *
 * @example
 * // Name (via cache or API)
 * await resolveNotionId('Tasks Database', 'database')
 * // Returns: '1fb79d4c71bb8032b722c82305b63a00'
 *
 * @example
 * // database_id auto-conversion
 * await resolveNotionId('abc123...', 'database')
 * // If abc123 is a database_id, auto-resolves to data_source_id
 */
export declare function resolveNotionId(input: string, type?: 'database' | 'page'): Promise<string>;
