"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../../notion");
const helper_1 = require("../../../helper");
class BlockRetrieveChildren extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(BlockRetrieveChildren);
        // TODO: Add support start_cursor, page_size
        const res = await notion.retrieveBlockChildren(args.block_id);
        if (flags.raw) {
            (0, helper_1.outputRawJson)(res);
            this.exit(0);
        }
        const columns = {
            object: {},
            id: {},
            type: {},
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
        core_1.ux.table(res.results, columns, options);
    }
}
exports.default = BlockRetrieveChildren;
BlockRetrieveChildren.description = 'Retrieve block children';
BlockRetrieveChildren.aliases = ['block:r:c'];
BlockRetrieveChildren.examples = [
    {
        description: 'Retrieve block children',
        command: `$ notion-cli block retrieve:children BLOCK_ID`,
    },
    {
        description: 'Retrieve block children and output raw json',
        command: `$ notion-cli block retrieve:children BLOCK_ID -r`,
    },
];
BlockRetrieveChildren.args = {
    block_id: core_1.Args.string({
        description: 'block_id or page_id',
        required: true,
    }),
};
BlockRetrieveChildren.flags = {
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
};
