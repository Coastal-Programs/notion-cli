"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
const notion_resolver_1 = require("../../utils/notion-resolver");
class DbUpdate extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(DbUpdate);
        try {
            // Resolve ID from URL, direct ID, or name (future)
            const dataSourceId = await (0, notion_resolver_1.resolveNotionId)(args.database_id, 'database');
            const dsTitle = flags.title;
            // TODO: support other properties (description, properties schema, etc.)
            const dsProps = {
                data_source_id: dataSourceId,
                title: [
                    {
                        type: 'text',
                        text: {
                            content: dsTitle,
                        },
                    },
                ],
            };
            const res = await notion.updateDataSource(dsProps);
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
            // Handle table output
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
            const options = {
                printLine: this.log.bind(this),
                ...flags,
            };
            core_1.ux.table([res], columns, options);
            process.exit(0);
        }
        catch (error) {
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    resourceType: 'database',
                    attemptedId: args.database_id,
                    endpoint: 'dataSources.update'
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
exports.default = DbUpdate;
DbUpdate.description = 'Update a data source (table) title and properties';
DbUpdate.aliases = ['db:u', 'ds:update', 'ds:u'];
DbUpdate.examples = [
    {
        description: 'Update a data source with a specific data_source_id and title',
        command: `$ notion-cli db update DATA_SOURCE_ID -t 'My Data Source'`,
    },
    {
        description: 'Update a data source via URL',
        command: `$ notion-cli db update https://notion.so/DATABASE_ID -t 'My Data Source'`,
    },
    {
        description: 'Update a data source with a specific data_source_id and output raw json',
        command: `$ notion-cli db update DATA_SOURCE_ID -t 'My Table' -r`,
    },
];
DbUpdate.args = {
    database_id: core_1.Args.string({
        required: true,
        description: 'Data source ID or URL (the ID of the table you want to update)',
    }),
};
DbUpdate.flags = {
    title: core_1.Flags.string({
        char: 't',
        description: 'New database title',
        required: true,
    }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
    ...base_flags_1.AutomationFlags,
};
