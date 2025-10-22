/**
 * Property Example Generator for Notion API
 *
 * Generates copy-pastable property payload examples based on database schema.
 * Helps AI agents understand the correct format for create/update operations.
 */
/**
 * Property example with simple value and Notion API payload
 */
export interface PropertyExample {
    property_name: string;
    property_type: string;
    simple_value: string | number | boolean | string[] | null;
    notion_payload: Record<string, any>;
    description: string;
}
/**
 * Generate property examples for all properties in a data source schema
 *
 * @param properties - Properties object from GetDataSourceResponse
 * @returns Array of property examples
 */
export declare function generatePropertyExamples(properties: Record<string, any>): PropertyExample[];
/**
 * Format examples for human-readable console output
 *
 * @param examples - Array of property examples
 * @returns Formatted string
 */
export declare function formatExamplesForConsole(examples: PropertyExample[]): string;
/**
 * Group examples by writability (writable vs read-only)
 *
 * @param examples - Array of property examples
 * @returns Grouped examples
 */
export declare function groupExamplesByWritability(examples: PropertyExample[]): {
    writable: PropertyExample[];
    readOnly: PropertyExample[];
};
