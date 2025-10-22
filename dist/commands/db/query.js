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
class DbQuery extends core_1.Command {
    async run() {
        const { flags, args } = await this.parse(DbQuery);
        try {
            // Resolve ID from URL, direct ID, or name (future)
            const databaseId = await (0, notion_resolver_1.resolveNotionId)(args.database_id, 'database');
            let queryParams;
            // Build query parameters
            try {
                if (flags.rawFilter != undefined) {
                    const filter = JSON.parse(flags.rawFilter);
                    queryParams = {
                        data_source_id: databaseId,
                        filter: filter,
                        page_size: flags.pageSize,
                    };
                }
                else if (flags.fileFilter != undefined) {
                    const fp = path.join('./', flags.fileFilter);
                    const fj = fs.readFileSync(fp, { encoding: 'utf-8' });
                    const filter = JSON.parse(fj);
                    queryParams = {
                        data_source_id: databaseId,
                        filter: filter,
                        page_size: flags.pageSize,
                    };
                }
                else {
                    let sorts = [];
                    const direction = flags.sortDirection == 'desc' ? 'descending' : 'ascending';
                    if (flags.sortProperty != undefined) {
                        sorts.push({
                            property: flags.sortProperty,
                            direction: direction,
                        });
                    }
                    queryParams = {
                        data_source_id: databaseId,
                        sorts: sorts,
                        page_size: flags.pageSize,
                    };
                }
            }
            catch (e) {
                throw new errors_1.NotionCLIError('VALIDATION_ERROR', `Failed to parse filter: ${e.message}`, { error: e });
            }
            // Fetch pages from database
            let pages = [];
            if (flags.pageAll) {
                pages = await notion.fetchAllPagesInDS(databaseId, queryParams.filter);
            }
            else {
                const res = await notion_1.client.dataSources.query(queryParams);
                pages.push(...res.results);
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
            core_1.ux.table(pages, columns, options);
            // Show hint after table output to make -r flag discoverable
            // Use first page as sample to count fields
            if (pages.length > 0) {
                (0, helper_1.showRawFlagHint)(pages.length, pages[0]);
            }
            process.exit(0);
        }
        catch (error) {
            const cliError = (0, errors_1.wrapNotionError)(error);
            if (flags.json) {
                this.log(JSON.stringify(cliError.toJSON(), null, 2));
            }
            else {
                this.error(cliError.message);
            }
            process.exit(1);
        }
    }
}
exports.default = DbQuery;
DbQuery.description = 'Query a database';
DbQuery.aliases = ['db:q'];
DbQuery.examples = [
    {
        description: 'Query a database with full data (recommended for AI assistants)',
        command: `$ notion-cli db query DATABASE_ID -r`,
    },
    {
        description: 'Query a db with a specific database_id',
        command: `$ notion-cli db query DATABASE_ID`,
    },
    {
        description: 'Query a db using database URL',
        command: `$ notion-cli db query https://notion.so/DATABASE_ID`,
    },
    {
        description: 'Query a db with a specific database_id and raw filter string',
        command: `$ notion-cli db query -a '{"and": ...}' DATABASE_ID`,
    },
    {
        description: 'Query a db with a specific database_id and filter file',
        command: `$ notion-cli db query -f ./path/to/filter.json DATABASE_ID`,
    },
    {
        description: 'Query a db with a specific database_id and output CSV',
        command: `$ notion-cli db query --csv DATABASE_ID`,
    },
    {
        description: 'Query a db with a specific database_id and output raw json',
        command: `$ notion-cli db query --raw DATABASE_ID`,
    },
    {
        description: 'Query a db with a specific database_id and output markdown table',
        command: `$ notion-cli db query --markdown DATABASE_ID`,
    },
    {
        description: 'Query a db with a specific database_id and output compact json',
        command: `$ notion-cli db query --compact-json DATABASE_ID`,
    },
    {
        description: 'Query a db with a specific database_id and output pretty table',
        command: `$ notion-cli db query --pretty DATABASE_ID`,
    },
    {
        description: 'Query a db with a specific database_id and page size',
        command: `$ notion-cli db query -p 10 DATABASE_ID`,
    },
    {
        description: 'Query a db with a specific database_id and get all pages',
        command: `$ notion-cli db query -A DATABASE_ID`,
    },
    {
        description: 'Query a db with a specific database_id and sort property and sort direction',
        command: `$ notion-cli db query -s Name -d desc DATABASE_ID`,
    },
];
DbQuery.args = {
    database_id: core_1.Args.string({
        required: true,
        description: 'Database or data source ID or URL (required for automation)',
    }),
};
DbQuery.flags = {
    rawFilter: core_1.Flags.string({
        char: 'a',
        description: 'JSON stringified filter string',
    }),
    fileFilter: core_1.Flags.string({
        char: 'f',
        description: 'JSON filter file path',
    }),
    pageSize: core_1.Flags.integer({
        char: 'p',
        description: 'The number of results to return(1-100). ',
        min: 1,
        max: 100,
        default: 10,
    }),
    pageAll: core_1.Flags.boolean({
        char: 'A',
        description: 'get all pages',
        default: false,
    }),
    sortProperty: core_1.Flags.string({
        char: 's',
        description: 'The property to sort results by',
    }),
    sortDirection: core_1.Flags.string({
        char: 'd',
        options: ['asc', 'desc'],
        description: 'The direction to sort results',
        default: 'asc',
    }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json (recommended for AI assistants - returns all page data)',
        default: false,
    }),
    ...core_1.ux.table.flags(),
    ...base_flags_1.AutomationFlags,
    ...base_flags_1.OutputFormatFlags,
};
