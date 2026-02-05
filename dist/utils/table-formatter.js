"use strict";
/**
 * Table formatting utility to replace oclif v2's ux.table
 * Provides backward-compatible table flags and formatting
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.tableFlags = void 0;
exports.formatTable = formatTable;
const core_1 = require("@oclif/core");
const Table = require("cli-table3");
/**
 * Table flags compatible with oclif v2's ux.table.flags()
 */
exports.tableFlags = {
    columns: core_1.Flags.string({
        description: 'Only show provided columns (comma-separated)',
        exclusive: ['extended'],
    }),
    sort: core_1.Flags.string({
        description: 'Property to sort by (prepend with - for descending)',
    }),
    filter: core_1.Flags.string({
        description: 'Filter property by substring match',
    }),
    csv: core_1.Flags.boolean({
        description: 'Output in CSV format',
        exclusive: ['no-truncate'],
    }),
    extended: core_1.Flags.boolean({
        char: 'x',
        description: 'Show extra columns',
    }),
    'no-truncate': core_1.Flags.boolean({
        description: 'Do not truncate output to fit screen',
        exclusive: ['csv'],
    }),
    'no-header': core_1.Flags.boolean({
        description: 'Hide table header from output',
    }),
};
/**
 * Format and display a table (compatible with oclif v2's ux.table)
 */
function formatTable(data, columns, options = {}) {
    if (data.length === 0) {
        return;
    }
    const printLine = options.printLine || console.log;
    // Filter columns based on options
    let selectedColumns = Object.keys(columns);
    if (options.columns) {
        const requestedCols = options.columns.split(',').map(c => c.trim());
        selectedColumns = selectedColumns.filter(col => requestedCols.includes(col));
    }
    if (!options.extended) {
        selectedColumns = selectedColumns.filter(col => !columns[col].extended);
    }
    // Filter rows
    let filteredData = data;
    if (options.filter) {
        const [filterCol, filterVal] = options.filter.split('=');
        if (filterVal) {
            filteredData = data.filter(row => {
                var _a;
                const val = ((_a = columns[filterCol]) === null || _a === void 0 ? void 0 : _a.get) ? columns[filterCol].get(row) : row[filterCol];
                return String(val).includes(filterVal);
            });
        }
    }
    // Sort data
    if (options.sort) {
        const descending = options.sort.startsWith('-');
        const sortCol = descending ? options.sort.slice(1) : options.sort;
        filteredData = [...filteredData].sort((a, b) => {
            var _a, _b;
            const aVal = ((_a = columns[sortCol]) === null || _a === void 0 ? void 0 : _a.get) ? columns[sortCol].get(a) : a[sortCol];
            const bVal = ((_b = columns[sortCol]) === null || _b === void 0 ? void 0 : _b.get) ? columns[sortCol].get(b) : b[sortCol];
            const comparison = String(aVal).localeCompare(String(bVal));
            return descending ? -comparison : comparison;
        });
    }
    // Output as CSV
    if (options.csv) {
        if (!options['no-header']) {
            const headers = selectedColumns.map(col => columns[col].header || col);
            printLine(headers.join(','));
        }
        filteredData.forEach(row => {
            const values = selectedColumns.map(col => {
                const val = columns[col].get ? columns[col].get(row) : row[col];
                const str = String(val || '');
                // Escape CSV values
                return str.includes(',') || str.includes('"') ? `"${str.replace(/"/g, '""')}"` : str;
            });
            printLine(values.join(','));
        });
        return;
    }
    // Output as table
    const headers = selectedColumns.map(col => columns[col].header || col);
    const table = new Table({
        head: options['no-header'] ? [] : headers,
        style: {
            head: ['cyan'],
            border: ['gray']
        },
        wordWrap: !options['no-truncate'],
        colWidths: selectedColumns.map(col => {
            if (options['no-truncate'])
                return undefined;
            return columns[col].minWidth || undefined;
        }),
    });
    filteredData.forEach(row => {
        const values = selectedColumns.map(col => {
            const val = columns[col].get ? columns[col].get(row) : row[col];
            return String(val !== undefined && val !== null ? val : '');
        });
        table.push(values);
    });
    printLine(table.toString());
}
