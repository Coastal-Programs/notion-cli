import { GetDataSourceResponse } from '@notionhq/client/build/src/api-endpoints';
/**
 * Simple flat property format for AI agents
 * Instead of complex Notion nested structures, use simple key-value pairs:
 * { "Name": "Task", "Status": "Done", "Tags": ["urgent", "bug"] }
 */
export interface SimpleProperties {
    [key: string]: string | number | boolean | string[] | null;
}
/**
 * Notion API property format (deeply nested)
 */
export interface NotionProperties {
    [key: string]: any;
}
/**
 * Expand simple flat properties to Notion API format
 *
 * This function takes simplified property values and automatically expands them
 * to the correct Notion API structure based on the database schema.
 *
 * @param simple - Flat key-value property object
 * @param schema - Database properties schema from data source
 * @returns Properly formatted Notion properties object
 *
 * @example
 * // Input (simple):
 * { "Name": "My Task", "Status": "In Progress", "Priority": 5 }
 *
 * // Output (Notion format):
 * {
 *   "Name": { "title": [{ "text": { "content": "My Task" } }] },
 *   "Status": { "select": { "name": "In Progress" } },
 *   "Priority": { "number": 5 }
 * }
 */
export declare function expandSimpleProperties(simple: SimpleProperties, schema: GetDataSourceResponse['properties']): Promise<NotionProperties>;
/**
 * Validate simple properties against schema before expansion
 * This can be called optionally before expandSimpleProperties to get detailed errors
 */
export declare function validateSimpleProperties(simple: SimpleProperties, schema: GetDataSourceResponse['properties']): {
    valid: boolean;
    errors: string[];
};
