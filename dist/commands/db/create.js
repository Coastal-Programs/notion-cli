"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
class DbCreate extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(DbCreate);
        console.log(`Creating a database in page ${args.page_id}`);
        const dbTitle = flags.title;
        try {
            // TODO: support other properties
            const dbProps = {
                parent: {
                    type: 'page_id',
                    page_id: args.page_id,
                },
                title: [
                    {
                        type: 'text',
                        text: {
                            content: dbTitle,
                        },
                    },
                ],
                initial_data_source: {
                    properties: {
                        Name: {
                            title: {},
                        },
                    },
                },
            };
            const res = await notion.createDb(dbProps);
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
                        return (0, helper_1.getDbTitle)(row);
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
exports.default = DbCreate;
DbCreate.description = 'Create a database with an initial data source (table)';
DbCreate.aliases = ['db:c'];
DbCreate.examples = [
    {
        description: 'Create a database with an initial data source',
        command: `$ notion-cli db create PAGE_ID -t 'My Database'`,
    },
    {
        description: 'Create a database with an initial data source and output raw json',
        command: `$ notion-cli db create PAGE_ID -t 'My Database' -r`,
    },
];
DbCreate.args = {
    page_id: core_1.Args.string({ required: true, description: 'Parent page ID where the database will be created' }),
};
DbCreate.flags = {
    title: core_1.Flags.string({
        char: 't',
        description: 'Title for the database (and initial data source)',
        required: true,
    }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
    ...base_flags_1.AutomationFlags,
};
