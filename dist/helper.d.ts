import { QueryDataSourceResponse, GetDataSourceResponse, DatabaseObjectResponse, DataSourceObjectResponse, PageObjectResponse, BlockObjectResponse } from '@notionhq/client/build/src/api-endpoints';
export declare const outputRawJson: (res: any) => Promise<void>;
/**
 * Output data as compact JSON (single-line, no formatting)
 * Useful for piping to other tools or scripts
 */
export declare const outputCompactJson: (res: any) => void;
/**
 * Strip unnecessary metadata from Notion API responses to reduce size
 * Removes created_by, last_edited_by, object fields, request_id, empty values, etc.
 * Keeps timestamps (created_time, last_edited_time) and essential data
 *
 * @param data The data to strip metadata from (single object or array)
 * @returns The stripped data
 */
export declare const stripMetadata: (data: any) => any;
/**
 * Output data as a markdown table
 * Converts column data into GitHub-flavored markdown table format
 */
export declare const outputMarkdownTable: (data: any[], columns: Record<string, any>) => void;
/**
 * Output data as a pretty table with borders
 * Enhanced table format with better visual separation
 */
export declare const outputPrettyTable: (data: any[], columns: Record<string, any>) => void;
/**
 * Show a hint to users (especially AI assistants) that more data is available with the -r flag
 * This makes the -r flag more discoverable for automation and AI use cases
 *
 * @param itemCount Number of items displayed in the table
 * @param item The item object to count total fields from
 * @param visibleFields Number of fields shown in the table (default: 4 for title, object, id, url)
 */
export declare function showRawFlagHint(itemCount: number, item: any, visibleFields?: number): void;
export declare const getFilterFields: (type: string) => Promise<{
    title: string;
}[]>;
export declare const buildDatabaseQueryFilter: (name: string, type: string, field: string, value: string | string[] | boolean) => Promise<object | null>;
export declare const buildPagePropUpdateData: (name: string, type: string, value: string) => Promise<object | null>;
export declare const buildOneDepthJson: (pages: QueryDataSourceResponse['results']) => Promise<{
    oneDepthJson: any[];
    relationJson: any[];
}>;
export declare const getDbTitle: (row: DatabaseObjectResponse) => string;
export declare const getDataSourceTitle: (row: GetDataSourceResponse | DataSourceObjectResponse) => string;
export declare const getPageTitle: (row: PageObjectResponse) => string;
export declare const getBlockPlainText: (row: BlockObjectResponse) => any;
/**
 * Build block JSON from simple text-based flags
 * Returns an array of block objects ready for Notion API
 */
export declare const buildBlocksFromTextFlags: (flags: {
    text?: string;
    heading1?: string;
    heading2?: string;
    heading3?: string;
    bullet?: string;
    numbered?: string;
    todo?: string;
    toggle?: string;
    code?: string;
    language?: string;
    quote?: string;
    callout?: string;
}) => any[];
/**
 * Attempt to enrich a child_database block with its queryable data_source_id
 *
 * The Notion API returns child_database blocks without the database/data_source ID,
 * making them unqueryable. This function attempts to resolve the block ID to a
 * queryable data_source_id by trying to retrieve it as a data source.
 *
 * @param block The child_database block to enrich
 * @returns The enriched block with data_source_id and database_id fields, or original block if resolution fails
 */
export declare const enrichChildDatabaseBlock: (block: BlockObjectResponse) => Promise<BlockObjectResponse>;
/**
 * Get all child_database blocks from a list of blocks and enrich them with queryable IDs
 *
 * @param blocks Array of blocks to filter and enrich
 * @returns Array of enriched child_database blocks with title, block_id, data_source_id, and database_id
 */
export declare const getChildDatabasesWithIds: (blocks: BlockObjectResponse[]) => Promise<any[]>;
/**
 * Build block update content from simple text flags
 * Returns an object with the block type properties for updating
 */
export declare const buildBlockUpdateFromTextFlags: (blockType: string, flags: {
    text?: string;
    heading1?: string;
    heading2?: string;
    heading3?: string;
    bullet?: string;
    numbered?: string;
    todo?: string;
    toggle?: string;
    code?: string;
    language?: string;
    quote?: string;
    callout?: string;
}) => any;
