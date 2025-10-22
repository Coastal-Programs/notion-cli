"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../../notion");
const helper_1 = require("../../../helper");
class UserRetrieveBot extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(UserRetrieveBot);
        const res = await notion.botUser();
        if (flags.raw) {
            (0, helper_1.outputRawJson)(res);
            this.exit(0);
        }
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
];
UserRetrieveBot.args = {};
UserRetrieveBot.flags = {
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
};
