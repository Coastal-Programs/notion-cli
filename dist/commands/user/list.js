"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
class UserList extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(UserList);
        const res = await notion.listUser();
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
        core_1.ux.table(res.results, columns, options);
    }
}
exports.default = UserList;
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
];
UserList.flags = {
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
};
