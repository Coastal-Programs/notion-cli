"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const client_1 = require("@notionhq/client");
const fs = require("fs");
const path = require("path");
const helper_1 = require("../../helper");
const notion_1 = require("../../notion");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
const notion_resolver_1 = require("../../utils/notion-resolver");
const table_formatter_1 = require("../../utils/table-formatter");
class DbQuery extends core_1.Command {
    async run() {
        const { flags, args } = await this.parse(DbQuery);
        try {
            // Handle deprecation warnings (output to stderr to not pollute stdout)
            if (flags.rawFilter) {
                console.error('⚠️  Warning: --rawFilter is deprecated and will be removed in v6.0.0');
                console.error('   Use --filter instead: notion-cli db query DS_ID --filter \'...\'');
                console.error('');
            }
            if (flags.fileFilter) {
                console.error('⚠️  Warning: --fileFilter is deprecated and will be removed in v6.0.0');
                console.error('   Use --file-filter instead: notion-cli db query DS_ID --file-filter ./filter.json');
                console.error('');
            }
            // Resolve ID from URL, direct ID, or name (future)
            const databaseId = await (0, notion_resolver_1.resolveNotionId)(args.database_id, 'database');
            let queryParams;
            // Build filter
            let filter = undefined;
            try {
                if (flags.filter || flags.rawFilter) {
                    // JSON filter object (new flag or deprecated rawFilter)
                    const filterStr = flags.filter || flags.rawFilter;
                    try {
                        filter = JSON.parse(filterStr);
                    }
                    catch (error) {
                        throw errors_1.NotionCLIErrorFactory.invalidJson(filterStr, error);
                    }
                }
                else if (flags['file-filter'] || flags.fileFilter) {
                    // Load from file (new flag or deprecated fileFilter)
                    const filterFile = flags['file-filter'] || flags.fileFilter;
                    const fp = path.join('./', filterFile);
                    let fj;
                    try {
                        fj = fs.readFileSync(fp, { encoding: 'utf-8' });
                        filter = JSON.parse(fj);
                    }
                    catch (error) {
                        if (error.code === 'ENOENT') {
                            throw errors_1.NotionCLIErrorFactory.invalidJson(filterFile, new Error(`File not found: ${filterFile}`));
                        }
                        throw errors_1.NotionCLIErrorFactory.invalidJson(fj, error);
                    }
                }
                else if (flags.search) {
                    // Simple text search - convert to Notion filter
                    // Search across common text properties using OR
                    // Note: This searches properties named "Name", "Title", and "Description"
                    // For more complex searches, use --filter with explicit property names
                    filter = {
                        or: [
                            { property: 'Name', title: { contains: flags.search } },
                            { property: 'Title', title: { contains: flags.search } },
                            { property: 'Description', rich_text: { contains: flags.search } },
                            { property: 'Name', rich_text: { contains: flags.search } },
                        ]
                    };
                }
                // Build sorts
                const sorts = [];
                const direction = flags['sort-direction'] == 'desc' ? 'descending' : 'ascending';
                if (flags['sort-property']) {
                    sorts.push({
                        property: flags['sort-property'],
                        direction: direction,
                    });
                }
                // Build query parameters
                queryParams = {
                    data_source_id: databaseId,
                    filter: filter,
                    sorts: sorts.length > 0 ? sorts : undefined,
                    page_size: flags['page-size'],
                };
            }
            catch (e) {
                // Re-throw NotionCLIError, wrap others
                if (e instanceof errors_1.NotionCLIError) {
                    throw e;
                }
                throw (0, errors_1.wrapNotionError)(e, {
                    resourceType: 'database',
                    userInput: args.database_id
                });
            }
            // Fetch pages from database
            let pages = [];
            if (flags['page-all']) {
                pages = await notion.fetchAllPagesInDS(databaseId, queryParams.filter);
            }
            else {
                const res = await notion_1.client.dataSources.query(queryParams);
                pages.push(...res.results);
            }
            // Apply minimal flag to strip metadata
            if (flags.minimal) {
                pages = (0, helper_1.stripMetadata)(pages);
            }
            // Apply property selection if --select flag is used
            if (flags.select) {
                const selectedProps = flags.select.split(',').map(p => p.trim());
                pages = pages.map((page) => {
                    if (page.object === 'page' && page.properties) {
                        // Keep core fields, filter properties
                        const filtered = {
                            ...page,
                            properties: {}
                        };
                        // Copy only selected properties
                        selectedProps.forEach(propName => {
                            if (page.properties[propName]) {
                                filtered.properties[propName] = page.properties[propName];
                            }
                        });
                        return filtered;
                    }
                    return page;
                });
            }
            // Define columns for table output
            const columns = {
                title: {
                    get: (row) => {
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
                (0, helper_1.outputCompactJson)(pages);
                process.exit(0);
                return;
            }
            // Handle markdown table output
            if (flags.markdown) {
                (0, helper_1.outputMarkdownTable)(pages, columns);
                process.exit(0);
                return;
            }
            // Handle pretty table output
            if (flags.pretty) {
                (0, helper_1.outputPrettyTable)(pages, columns);
                // Show hint after table output (use first page as sample)
                if (pages.length > 0) {
                    (0, helper_1.showRawFlagHint)(pages.length, pages[0]);
                }
                process.exit(0);
                return;
            }
            // Handle JSON output for automation
            if (flags.json) {
                this.log(JSON.stringify({
                    success: true,
                    data: pages,
                    count: pages.length,
                    timestamp: new Date().toISOString()
                }, null, 2));
                process.exit(0);
                return;
            }
            // Handle raw JSON output (legacy)
            if (flags.raw) {
                (0, helper_1.outputRawJson)(pages);
                process.exit(0);
                return;
            }
            // Handle table output (default)
            const options = {
                printLine: this.log.bind(this),
                ...flags,
            };
            (0, table_formatter_1.formatTable)(pages, columns, options);
            // Show hint after table output to make -r flag discoverable
            // Use first page as sample to count fields
            if (pages.length > 0) {
                (0, helper_1.showRawFlagHint)(pages.length, pages[0]);
            }
            process.exit(0);
        }
        catch (error) {
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    resourceType: 'database',
                    attemptedId: args.database_id,
                    endpoint: 'dataSources.query'
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
DbQuery.description = 'Query a database';
DbQuery.aliases = ['db:q'];
DbQuery.examples = [
    {
        description: 'Query a database with full data (recommended for AI assistants)',
        command: `$ notion-cli db query DATABASE_ID --raw`,
    },
    {
        description: 'Query all records as JSON',
        command: `$ notion-cli db query DATABASE_ID --json`,
    },
    {
        description: 'Filter with JSON object (recommended for AI agents)',
        command: `$ notion-cli db query DATABASE_ID --filter '{"property": "Status", "select": {"equals": "Done"}}' --json`,
    },
    {
        description: 'Simple text search across properties',
        command: `$ notion-cli db query DATABASE_ID --search "urgent" --json`,
    },
    {
        description: 'Load complex filter from file',
        command: `$ notion-cli db query DATABASE_ID --file-filter ./filter.json --json`,
    },
    {
        description: 'Query with AND filter',
        command: `$ notion-cli db query DATABASE_ID --filter '{"and": [{"property": "Status", "select": {"equals": "Done"}}, {"property": "Priority", "number": {"greater_than": 5}}]}' --json`,
    },
    {
        description: 'Query using database URL',
        command: `$ notion-cli db query https://notion.so/DATABASE_ID --json`,
    },
    {
        description: 'Query with sorting',
        command: `$ notion-cli db query DATABASE_ID --sort-property Name --sort-direction desc`,
    },
    {
        description: 'Query with pagination',
        command: `$ notion-cli db query DATABASE_ID --page-size 50`,
    },
    {
        description: 'Get all pages (bypass pagination)',
        command: `$ notion-cli db query DATABASE_ID --page-all`,
    },
    {
        description: 'Output as CSV',
        command: `$ notion-cli db query DATABASE_ID --csv`,
    },
    {
        description: 'Output as markdown table',
        command: `$ notion-cli db query DATABASE_ID --markdown`,
    },
    {
        description: 'Output as compact JSON',
        command: `$ notion-cli db query DATABASE_ID --compact-json`,
    },
    {
        description: 'Output as pretty table',
        command: `$ notion-cli db query DATABASE_ID --pretty`,
    },
    {
        description: 'Select specific properties (60-80% token reduction)',
        command: `$ notion-cli db query DATABASE_ID --select "title,status,priority" --json`,
    },
];
DbQuery.args = {
    database_id: core_1.Args.string({
        required: true,
        description: 'Database or data source ID or URL (required for automation)',
    }),
};
DbQuery.flags = {
    'page-size': core_1.Flags.integer({
        char: 'p',
        description: 'The number of results to return (1-100)',
        min: 1,
        max: 100,
        default: 10,
    }),
    'page-all': core_1.Flags.boolean({
        char: 'A',
        description: 'Get all pages (bypass pagination)',
        default: false,
    }),
    'sort-property': core_1.Flags.string({
        description: 'The property to sort results by',
    }),
    'sort-direction': core_1.Flags.string({
        options: ['asc', 'desc'],
        description: 'The direction to sort results',
        default: 'asc',
    }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'Output raw JSON (recommended for AI assistants - returns all page data)',
        default: false,
    }),
    ...table_formatter_1.tableFlags,
    ...base_flags_1.AutomationFlags,
    ...base_flags_1.OutputFormatFlags,
    // New simplified filter interface (placed AFTER table flags to override)
    filter: core_1.Flags.string({
        char: 'f',
        description: 'Filter as JSON object (Notion filter API format)',
        exclusive: ['search', 'file-filter', 'rawFilter', 'fileFilter'],
    }),
    'file-filter': core_1.Flags.string({
        char: 'F',
        description: 'Load filter from JSON file',
        exclusive: ['filter', 'search', 'rawFilter', 'fileFilter'],
    }),
    search: core_1.Flags.string({
        char: 's',
        description: 'Simple text search (searches across title and common text properties)',
        exclusive: ['filter', 'file-filter', 'rawFilter', 'fileFilter'],
    }),
    select: core_1.Flags.string({
        description: 'Select specific properties to return (comma-separated). Reduces token usage by 60-80%.',
        examples: ['title,status', 'title,status,priority,due_date'],
    }),
    // DEPRECATED: Keep for backward compatibility
    rawFilter: core_1.Flags.string({
        char: 'a',
        description: 'DEPRECATED: Use --filter instead. JSON stringified filter string',
        hidden: true,
        exclusive: ['filter', 'search', 'file-filter', 'fileFilter'],
    }),
    fileFilter: core_1.Flags.string({
        description: 'DEPRECATED: Use --file-filter instead. JSON filter file path',
        hidden: true,
        exclusive: ['filter', 'search', 'file-filter', 'rawFilter'],
    }),
};
exports.default = DbQuery;
