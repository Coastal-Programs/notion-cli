"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
class BlockDelete extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(BlockDelete);
        const res = await notion.deleteBlock(args.block_id);
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
exports.default = BlockDelete;
BlockDelete.description = 'Delete a block';
BlockDelete.aliases = ['block:d'];
BlockDelete.examples = [
    {
        description: 'Delete a block',
        command: `$ notion-cli block delete BLOCK_ID`,
    },
    {
        description: 'Delete a block and output raw json',
        command: `$ notion-cli block delete BLOCK_ID -r`,
    },
];
BlockDelete.args = {
    block_id: core_1.Args.string({ required: true }),
};
BlockDelete.flags = {
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
};
