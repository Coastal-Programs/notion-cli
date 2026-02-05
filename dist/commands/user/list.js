"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const table_formatter_1 = require("../../utils/table-formatter");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
class UserList extends core_1.Command {
    async run() {
        const { flags } = await this.parse(UserList);
        try {
            let res = await notion.listUser();
            // Apply minimal flag to strip metadata
            if (flags.minimal) {
                res = (0, helper_1.stripMetadata)(res);
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
            // Handle table output
            const columns = {
                id: {},
                name: {},
                object: {},
                type: {},
                person_or_bot: {
                    header: 'person/bot',
                    get: (row) => {
                        if (row.type === 'person') {
                            return row.person;
                        }
                        return row.bot;
                    },
                },
                avatar_url: {},
            };
            const options = {
                printLine: this.log.bind(this),
                ...flags,
            };
            (0, table_formatter_1.formatTable)(res.results, columns, options);
            process.exit(0);
        }
        catch (error) {
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    resourceType: 'user',
                    endpoint: 'users.list'
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
UserList.description = 'List all users';
UserList.aliases = ['user:l'];
UserList.examples = [
    {
        description: 'List all users',
        command: `$ notion-cli user list`,
    },
    {
        description: 'List all users and output raw json',
        command: `$ notion-cli user list -r`,
    },
    {
        description: 'List all users and output JSON for automation',
        command: `$ notion-cli user list --json`,
    },
];
UserList.flags = {
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...table_formatter_1.tableFlags,
    ...base_flags_1.AutomationFlags,
};
exports.default = UserList;
