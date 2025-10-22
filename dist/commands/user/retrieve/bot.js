"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../../notion");
const helper_1 = require("../../../helper");
const base_flags_1 = require("../../../base-flags");
const errors_1 = require("../../../errors");
class UserRetrieveBot extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(UserRetrieveBot);
        try {
            const res = await notion.botUser();
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
exports.default = UserRetrieveBot;
UserRetrieveBot.description = 'Retrieve a bot user';
UserRetrieveBot.aliases = ['user:r:b'];
UserRetrieveBot.examples = [
    {
        description: 'Retrieve a bot user',
        command: `$ notion-cli user retrieve:bot`,
    },
    {
        description: 'Retrieve a bot user and output raw json',
        command: `$ notion-cli user retrieve:bot -r`,
    },
    {
        description: 'Retrieve a bot user and output JSON for automation',
        command: `$ notion-cli user retrieve:bot --json`,
    },
];
UserRetrieveBot.args = {};
UserRetrieveBot.flags = {
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
    ...base_flags_1.AutomationFlags,
};
