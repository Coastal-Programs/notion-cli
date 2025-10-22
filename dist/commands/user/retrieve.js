"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
class UserRetrieve extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(UserRetrieve);
        const res = await notion.retrieveUser(args.user_id);
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
exports.default = UserRetrieve;
UserRetrieve.description = 'Retrieve a user';
UserRetrieve.aliases = ['user:r'];
UserRetrieve.examples = [
    {
        description: 'Retrieve a user',
        command: `$ notion-cli user retrieve USER_ID`,
    },
    {
        description: 'Retrieve a user and output raw json',
        command: `$ notion-cli user retrieve USER_ID -r`,
    },
];
UserRetrieve.args = {
    user_id: core_1.Args.string(),
};
UserRetrieve.flags = {
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
};
