"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const table_formatter_1 = require("../../utils/table-formatter");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
const notion_resolver_1 = require("../../utils/notion-resolver");
class DbRetrieve extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(DbRetrieve);
        try {
            // Resolve ID from URL, direct ID, or name (future)
            const dataSourceId = await (0, notion_resolver_1.resolveNotionId)(args.database_id, 'database');
            let res = await notion.retrieveDataSource(dataSourceId);
            // Apply minimal flag to strip metadata
            if (flags.minimal) {
                res = (0, helper_1.stripMetadata)(res);
            }
            // Define columns for table output
            const columns = {
                title: {
                    get: (row) => {
                        return (0, helper_1.getDataSourceTitle)(row);
                    },
                },
                object: {},
                id: {},
                url: {},
            };
            // Handle compact JSON output
            if (flags['compact-json']) {
                (0, helper_1.outputCompactJson)(res);
                process.exit(0);
                return;
            }
            // Handle markdown table output
            if (flags.markdown) {
                (0, helper_1.outputMarkdownTable)([res], columns);
                process.exit(0);
                return;
            }
            // Handle pretty table output
            if (flags.pretty) {
                (0, helper_1.outputPrettyTable)([res], columns);
                // Show hint after table output
                (0, helper_1.showRawFlagHint)(1, res);
                process.exit(0);
                return;
            }
            // Handle JSON output for automation
            if (flags.json) {
                this.log(JSON.stringify({
                    success: true,
                    data: res,
                    timestamp: new Date().toISOString()
                }, null, 2));
                process.exit(0);
                return;
            }
            // Handle raw JSON output (legacy)
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
            (0, table_formatter_1.formatTable)([res], columns, options);
            // Show hint after table output to make -r flag discoverable
            (0, helper_1.showRawFlagHint)(1, res);
            process.exit(0);
        }
        catch (error) {
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    resourceType: 'database',
                    endpoint: 'dataSources.retrieve'
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
DbRetrieve.description = 'Retrieve a data source (table) schema and properties';
DbRetrieve.aliases = ['db:r', 'ds:retrieve', 'ds:r'];
DbRetrieve.examples = [
    {
        description: 'Retrieve a data source with full schema (recommended for AI assistants)',
        command: 'notion-cli db retrieve DATA_SOURCE_ID -r',
    },
    {
        description: 'Retrieve a data source schema via data_source_id',
        command: 'notion-cli db retrieve DATA_SOURCE_ID',
    },
    {
        description: 'Retrieve a data source via URL',
        command: 'notion-cli db retrieve https://notion.so/DATABASE_ID',
    },
    {
        description: 'Retrieve a data source and output as markdown table',
        command: 'notion-cli db retrieve DATA_SOURCE_ID --markdown',
    },
    {
        description: 'Retrieve a data source and output as compact JSON',
        command: 'notion-cli db retrieve DATA_SOURCE_ID --compact-json',
    },
];
DbRetrieve.args = {
    database_id: core_1.Args.string({
        required: true,
        description: 'Data source ID or URL (the ID of the table whose schema you want to retrieve)',
    }),
};
DbRetrieve.flags = {
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json (recommended for AI assistants - returns full schema)',
    }),
    ...table_formatter_1.tableFlags,
    ...base_flags_1.AutomationFlags,
    ...base_flags_1.OutputFormatFlags,
};
exports.default = DbRetrieve;
