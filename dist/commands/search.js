"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../notion");
const client_1 = require("@notionhq/client");
const helper_1 = require("../helper");
const base_flags_1 = require("../base-flags");
class Search extends core_1.Command {
    async run() {
        const { flags } = await this.parse(Search);
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
        if (flags.page_size) {
            params.page_size = flags.page_size;
        }
        if (process.env.DEBUG) {
            console.log(params);
        }
        const res = await notion.search(params);
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
            this.exit(0);
        }
        // Handle markdown table output
        if (flags.markdown) {
            (0, helper_1.outputMarkdownTable)(res.results, columns);
            this.exit(0);
        }
        // Handle pretty table output
        if (flags.pretty) {
            (0, helper_1.outputPrettyTable)(res.results, columns);
            this.exit(0);
        }
        // Handle raw JSON output
        if (flags.raw) {
            (0, helper_1.outputRawJson)(res);
            this.exit(0);
        }
        // Handle table output (default)
        const options = {
            printLine: this.log.bind(this),
            ...flags,
        };
        core_1.ux.table(res.results, columns, options);
    }
}
exports.default = Search;
Search.description = 'Search by title';
Search.examples = [
    {
        description: 'Search by title',
        command: `$ notion-cli search -q 'My Page'`,
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
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
    ...base_flags_1.OutputFormatFlags,
};
