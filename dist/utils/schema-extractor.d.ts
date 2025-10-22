import { GetDataSourceResponse } from '@notionhq/client/build/src/api-endpoints';
/**
 * Property schema for AI agents - simplified and easy to parse
 */
export interface PropertySchema {
    name: string;
    type: string;
    description?: string;
    required?: boolean;
    options?: string[];
    config?: Record<string, any>;
}
/**
 * Database schema in AI-friendly format
 */
export interface DataSourceSchema {
    id: string;
    title: string;
    description?: string;
    properties: PropertySchema[];
    url?: string;
}
/**
 * Extract clean, AI-parseable schema from Notion data source response
 *
 * This transforms the complex nested Notion API structure into a flat,
 * easy-to-understand format that AI agents can work with directly.
 *
 * @param dataSource - Raw Notion data source response
 * @returns Simplified schema object
 */
export declare function extractSchema(dataSource: GetDataSourceResponse): DataSourceSchema;
/**
 * Filter properties by names
 *
 * @param schema - Full schema
 * @param propertyNames - Array of property names to include
 * @returns Filtered schema
 */
export declare function filterProperties(schema: DataSourceSchema, propertyNames: string[]): DataSourceSchema;
/**
 * Format schema as human-readable table data
 *
 * @param schema - Schema to format
 * @returns Array of objects for table display
 */
export declare function formatSchemaForTable(schema: DataSourceSchema): Array<Record<string, string>>;
/**
 * Format schema as markdown documentation
 *
 * @param schema - Schema to format
 * @returns Markdown string
 */
export declare function formatSchemaAsMarkdown(schema: DataSourceSchema): string;
/**
 * Validate that a data object matches the schema
 *
 * @param schema - Schema to validate against
 * @param data - Data object to validate
 * @returns Validation result with errors
 */
export declare function validateAgainstSchema(schema: DataSourceSchema, data: Record<string, any>): {
    valid: boolean;
    errors: string[];
};
