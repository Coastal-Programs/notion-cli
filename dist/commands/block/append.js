"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
class BlockAppend extends core_1.Command {
    // TODO: Add support children params building prompt
    async run() {
        const { flags } = await this.parse(BlockAppend);
        const params = {
            block_id: flags.block_id,
            children: JSON.parse(flags.children),
        };
        if (flags.after) {
            params.after = flags.after;
        }
        const res = await notion.appendBlockChildren(params);
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
        core_1.ux.table(res.results, columns, options);
    }
}
exports.default = BlockAppend;
BlockAppend.description = 'Append block children';
BlockAppend.aliases = ['block:a'];
BlockAppend.examples = [
    {
        description: 'Append block children',
        command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]'`,
    },
    {
        description: 'Append block children after a block',
        command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]' -a AFTER_BLOCK_ID`,
    },
    {
        description: 'Append block children and output raw json',
        command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]' -r`,
    },
];
BlockAppend.flags = {
    block_id: core_1.Flags.string({
        char: 'b',
        description: 'Parent block ID',
        required: true,
    }),
    children: core_1.Flags.string({
        char: 'c',
        description: 'Block children (JSON array)',
        required: true,
    }),
    after: core_1.Flags.string({
        char: 'a',
        description: 'Block ID to append after (optional)',
    }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
};
