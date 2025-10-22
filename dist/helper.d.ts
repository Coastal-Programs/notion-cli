import { QueryDataSourceResponse, GetDataSourceResponse, DatabaseObjectResponse, DataSourceObjectResponse, PageObjectResponse, BlockObjectResponse } from '@notionhq/client/build/src/api-endpoints';
export declare const outputRawJson: (res: any) => Promise<void>;
/**
 * Output data as compact JSON (single-line, no formatting)
 * Useful for piping to other tools or scripts
 */
export declare const outputCompactJson: (res: any) => void;
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
