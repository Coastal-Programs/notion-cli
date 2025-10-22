"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
class BlockRetrieve extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(BlockRetrieve);
        const res = await notion.retrieveBlock(args.block_id);
        if (flags.raw) {
            (0, helper_1.outputRawJson)(res);
            this.exit(0);
        }
        const columns = {
            object: {},
            id: {},
            type: {},
            parent: {},
            content: {
                get: (row) => {
                    return (0, helper_1.getBlockPlainText)(row);
                },
            },
        };
        const options = {
            printLine: this.log.bind(this),
            ...flags,
        };
        core_1.ux.table([res], columns, options);
    }
}
exports.default = BlockRetrieve;
BlockRetrieve.description = 'Retrieve a block';
BlockRetrieve.aliases = ['block:r'];
BlockRetrieve.examples = [
    {
        description: 'Retrieve a block',
        command: `$ notion-cli block retrieve BLOCK_ID`,
    },
    {
        description: 'Retrieve a block and output raw json',
        command: `$ notion-cli block retrieve BLOCK_ID -r`,
    },
];
BlockRetrieve.args = {
    block_id: core_1.Args.string({ required: true }),
};
BlockRetrieve.flags = {
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
};
