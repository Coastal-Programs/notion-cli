"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../notion");
const client_1 = require("@notionhq/client");
const helper_1 = require("../helper");
const base_flags_1 = require("../base-flags");
const errors_1 = require("../errors");
const dayjs = require("dayjs");
class Search extends core_1.Command {
    async run() {
        const { flags } = await this.parse(Search);
        try {
            // Validate date filters
            if (flags['created-after'] && !dayjs(flags['created-after']).isValid()) {
                throw new errors_1.NotionCLIError(errors_1.NotionCLIErrorCode.VALIDATION_ERROR, `Invalid date format for --created-after: ${flags['created-after']}. Use ISO 8601 format (YYYY-MM-DD).`, [], { userInput: flags['created-after'] });
            }
            if (flags['created-before'] && !dayjs(flags['created-before']).isValid()) {
                throw new errors_1.NotionCLIError(errors_1.NotionCLIErrorCode.VALIDATION_ERROR, `Invalid date format for --created-before: ${flags['created-before']}. Use ISO 8601 format (YYYY-MM-DD).`, [], { userInput: flags['created-before'] });
            }
            if (flags['edited-after'] && !dayjs(flags['edited-after']).isValid()) {
                throw new errors_1.NotionCLIError(errors_1.NotionCLIErrorCode.VALIDATION_ERROR, `Invalid date format for --edited-after: ${flags['edited-after']}. Use ISO 8601 format (YYYY-MM-DD).`, [], { userInput: flags['edited-after'] });
            }
            if (flags['edited-before'] && !dayjs(flags['edited-before']).isValid()) {
                throw new errors_1.NotionCLIError(errors_1.NotionCLIErrorCode.VALIDATION_ERROR, `Invalid date format for --edited-before: ${flags['edited-before']}. Use ISO 8601 format (YYYY-MM-DD).`, [], { userInput: flags['edited-before'] });
            }
            const params = {};
            if (flags.query) {
                params.query = flags.query;
            }
            if (flags.sort_direction) {
                let direction;
                if (flags.sort_direction == 'asc') {
                    direction = 'ascending';
                }
                else {
                    direction = 'descending';
                }
                params.sort = {
                    direction: direction,
                    timestamp: 'last_edited_time',
                };
            }
            if (flags.property == 'data_source' || flags.property == 'page') {
                params.filter = {
                    value: flags.property,
                    property: 'object',
                };
            }
            if (flags.start_cursor) {
                params.start_cursor = flags.start_cursor;
            }
            // Increase page_size if we need to apply client-side filters
            // This ensures we get enough results before filtering
            const hasClientSideFilters = flags.database || flags['created-after'] ||
                flags['created-before'] || flags['edited-after'] || flags['edited-before'];
            if (hasClientSideFilters) {
                // Use 100 (max) to get more results for filtering
                params.page_size = 100;
            }
            else if (flags.page_size) {
                params.page_size = flags.page_size;
            }
            if (process.env.DEBUG) {
                console.log(params);
            }
            let res = await notion.search(params);
            // Apply minimal flag to strip metadata
            if (flags.minimal) {
                res = (0, helper_1.stripMetadata)(res);
            }
            // Apply client-side filters (Notion API doesn't support these natively in search)
            let filteredResults = res.results;
            // Filter by database (parent)
            if (flags.database) {
                filteredResults = filteredResults.filter((result) => {
                    if ((0, client_1.isFullPage)(result) && result.parent) {
                        if ('database_id' in result.parent) {
                            return result.parent.database_id === flags.database;
                        }
                    }
                    return false;
                });
            }
            // Filter by created date
            if (flags['created-after']) {
                const afterDate = dayjs(flags['created-after']);
                filteredResults = filteredResults.filter((result) => {
                    if ('created_time' in result) {
                        return dayjs(result.created_time).isAfter(afterDate) ||
                            dayjs(result.created_time).isSame(afterDate, 'day');
                    }
                    return false;
                });
            }
            if (flags['created-before']) {
                const beforeDate = dayjs(flags['created-before']);
                filteredResults = filteredResults.filter((result) => {
                    if ('created_time' in result) {
                        return dayjs(result.created_time).isBefore(beforeDate) ||
                            dayjs(result.created_time).isSame(beforeDate, 'day');
                    }
                    return false;
                });
            }
            // Filter by edited date
            if (flags['edited-after']) {
                const afterDate = dayjs(flags['edited-after']);
                filteredResults = filteredResults.filter((result) => {
                    if ('last_edited_time' in result) {
                        return dayjs(result.last_edited_time).isAfter(afterDate) ||
                            dayjs(result.last_edited_time).isSame(afterDate, 'day');
                    }
                    return false;
                });
            }
            if (flags['edited-before']) {
                const beforeDate = dayjs(flags['edited-before']);
                filteredResults = filteredResults.filter((result) => {
                    if ('last_edited_time' in result) {
                        return dayjs(result.last_edited_time).isBefore(beforeDate) ||
                            dayjs(result.last_edited_time).isSame(beforeDate, 'day');
                    }
                    return false;
                });
            }
            // Apply limit after all filters
            if (flags.limit) {
                filteredResults = filteredResults.slice(0, flags.limit);
            }
            // Update res.results with filtered results
            res.results = filteredResults;
            // Handle JSON output for automation (takes precedence)
            if (flags.json) {
                this.log(JSON.stringify({
                    success: true,
                    data: res,
                    timestamp: new Date().toISOString()
                }, null, 2));
                process.exit(0);
                return;
            }
            // Define columns for table output
            const columns = {
                title: {
                    get: (row) => {
                        if (row.object == 'database' && (0, client_1.isFullDatabase)(row)) {
                            return (0, helper_1.getDbTitle)(row);
                        }
                        if (row.object == 'data_source' && (0, client_1.isFullDataSource)(row)) {
                            return (0, helper_1.getDataSourceTitle)(row);
                        }
                        if (row.object == 'page' && (0, client_1.isFullPage)(row)) {
                            return (0, helper_1.getPageTitle)(row);
                        }
                        return 'Untitled';
                    },
                },
                object: {},
                id: {},
                url: {},
            };
            // Handle compact JSON output
            if (flags['compact-json']) {
                (0, helper_1.outputCompactJson)(res.results);
                process.exit(0);
                return;
            }
            // Handle markdown table output
            if (flags.markdown) {
                (0, helper_1.outputMarkdownTable)(res.results, columns);
                process.exit(0);
                return;
            }
            // Handle pretty table output
            if (flags.pretty) {
                (0, helper_1.outputPrettyTable)(res.results, columns);
                // Show hint after table output (use first result as sample)
                if (res.results.length > 0) {
                    (0, helper_1.showRawFlagHint)(res.results.length, res.results[0]);
                }
                process.exit(0);
                return;
            }
            // Handle raw JSON output
            if (flags.raw) {
                (0, helper_1.outputRawJson)(res);
                process.exit(0);
                return;
            }
            // Handle table output (default)
            const options = {
                printLine: this.log.bind(this),
                ...flags,
            };
            core_1.ux.table(res.results, columns, options);
            // Show hint after table output to make -r flag discoverable
            // Use first result as sample to count fields
            if (res.results.length > 0) {
                (0, helper_1.showRawFlagHint)(res.results.length, res.results[0]);
            }
        }
        catch (error) {
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    endpoint: 'search',
                    userInput: flags.query || flags.filter
                });
            if (flags.json) {
                this.log(JSON.stringify(cliError.toJSON(), null, 2));
            }
            else {
                this.error(cliError.toHumanString());
            }
            process.exit(1);
        }
    }
}
exports.default = Search;
Search.description = 'Search by title';
Search.examples = [
    {
        description: 'Search with full data (recommended for AI assistants)',
        command: `$ notion-cli search -q 'My Page' -r`,
    },
    {
        description: 'Search by title',
        command: `$ notion-cli search -q 'My Page'`,
    },
    {
        description: 'Search only within a specific database',
        command: `$ notion-cli search -q 'meeting' --database DB_ID`,
    },
    {
        description: 'Search with created date filter',
        command: `$ notion-cli search -q 'report' --created-after 2025-10-01`,
    },
    {
        description: 'Search with edited date filter',
        command: `$ notion-cli search -q 'project' --edited-after 2025-10-20`,
    },
    {
        description: 'Limit number of results',
        command: `$ notion-cli search -q 'task' --limit 20`,
    },
    {
        description: 'Combined filters',
        command: `$ notion-cli search -q 'project' -d DB_ID --edited-after 2025-10-20 --limit 10`,
    },
    {
        description: 'Search by title and output csv',
        command: `$ notion-cli search -q 'My Page' --csv`,
    },
    {
        description: 'Search by title and output raw json',
        command: `$ notion-cli search -q 'My Page' -r`,
    },
    {
        description: 'Search by title and output markdown table',
        command: `$ notion-cli search -q 'My Page' --markdown`,
    },
    {
        description: 'Search by title and output compact JSON',
        command: `$ notion-cli search -q 'My Page' --compact-json`,
    },
    {
        description: 'Search by title and output pretty table',
        command: `$ notion-cli search -q 'My Page' --pretty`,
    },
    {
        description: 'Search by title and output table with specific columns',
        command: `$ notion-cli search -q 'My Page' --columns=title,object`,
    },
    {
        description: 'Search by title and output table with specific columns and sort direction',
        command: `$ notion-cli search -q 'My Page' --columns=title,object -d asc`,
    },
    {
        description: 'Search by title and output table with specific columns and sort direction and page size',
        command: `$ notion-cli search -q 'My Page' -columns=title,object -d asc -s 10`,
    },
    {
        description: 'Search by title and output table with specific columns and sort direction and page size and start cursor',
        command: `$ notion-cli search -q 'My Page' --columns=title,object -d asc -s 10 -c START_CURSOR_ID`,
    },
    {
        description: 'Search by title and output table with specific columns and sort direction and page size and start cursor and property',
        command: `$ notion-cli search -q 'My Page' --columns=title,object -d asc -s 10 -c START_CURSOR_ID -p page`,
    },
    {
        description: 'Search and output JSON for automation',
        command: `$ notion-cli search -q 'My Page' --json`,
    },
];
Search.flags = {
    query: core_1.Flags.string({
        char: 'q',
        description: 'The text that the API compares page and database titles against',
    }),
    sort_direction: core_1.Flags.string({
        char: 'd',
        options: ['asc', 'desc'],
        description: 'The direction to sort results. The only supported timestamp value is "last_edited_time"',
        default: 'desc',
    }),
    property: core_1.Flags.string({
        char: 'p',
        options: ['data_source', 'page'],
    }),
    start_cursor: core_1.Flags.string({
        char: 'c',
    }),
    page_size: core_1.Flags.integer({
        char: 's',
        description: 'The number of results to return. The default is 5, with a minimum of 1 and a maximum of 100.',
        min: 1,
        max: 100,
        default: 5,
    }),
    database: core_1.Flags.string({
        description: 'Limit search to pages within a specific database (data source ID)',
    }),
    'created-after': core_1.Flags.string({
        description: 'Filter results created after this date (ISO 8601 format: YYYY-MM-DD)',
    }),
    'created-before': core_1.Flags.string({
        description: 'Filter results created before this date (ISO 8601 format: YYYY-MM-DD)',
    }),
    'edited-after': core_1.Flags.string({
        description: 'Filter results edited after this date (ISO 8601 format: YYYY-MM-DD)',
    }),
    'edited-before': core_1.Flags.string({
        description: 'Filter results edited before this date (ISO 8601 format: YYYY-MM-DD)',
    }),
    limit: core_1.Flags.integer({
        description: 'Maximum number of results to return (applied after filters)',
        min: 1,
    }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json (recommended for AI assistants - returns all search results)',
    }),
    ...core_1.ux.table.flags(),
    ...base_flags_1.OutputFormatFlags,
    ...base_flags_1.AutomationFlags,
};
