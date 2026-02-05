/**
 * Table formatting utility to replace oclif v2's ux.table
 * Provides backward-compatible table flags and formatting
 */
/**
 * Table flags compatible with oclif v2's ux.table.flags()
 */
export declare const tableFlags: {
    columns: import("@oclif/core/lib/interfaces").OptionFlag<string, import("@oclif/core/lib/interfaces").CustomOptions>;
    sort: import("@oclif/core/lib/interfaces").OptionFlag<string, import("@oclif/core/lib/interfaces").CustomOptions>;
    filter: import("@oclif/core/lib/interfaces").OptionFlag<string, import("@oclif/core/lib/interfaces").CustomOptions>;
    csv: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
    extended: import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
    'no-truncate': import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
    'no-header': import("@oclif/core/lib/interfaces").BooleanFlag<boolean>;
};
export interface ColumnOptions<T> {
    header?: string;
    get?: (row: T) => any;
    extended?: boolean;
    minWidth?: number;
}
export interface TableOptions {
    columns?: string;
    sort?: string;
    filter?: string;
    csv?: boolean;
    extended?: boolean;
    'no-truncate'?: boolean;
    'no-header'?: boolean;
    printLine?: (s: string) => void;
}
/**
 * Format and display a table (compatible with oclif v2's ux.table)
 */
export declare function formatTable<T extends Record<string, any>>(data: T[], columns: Record<string, ColumnOptions<T>>, options?: TableOptions): void;
