"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
class BlockUpdate extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(BlockUpdate);
        const params = {
            block_id: args.block_id,
        };
        // Handle archived flag
        if (flags.archived !== undefined) {
            params.archived = flags.archived;
        }
        // Handle content updates
        if (flags.content) {
            try {
                const content = JSON.parse(flags.content);
                Object.assign(params, content);
            }
            catch (error) {
                this.error('Invalid JSON in --content flag. Please provide valid JSON.');
            }
        }
        // Handle color updates
        if (flags.color) {
            params.color = flags.color;
        }
        const res = await notion.updateBlock(params);
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
exports.default = BlockUpdate;
BlockUpdate.description = 'Update a block';
BlockUpdate.aliases = ['block:u'];
BlockUpdate.examples = [
    {
        description: 'Archive a block',
        command: `$ notion-cli block update BLOCK_ID -a`,
    },
    {
        description: 'Update block content',
        command: `$ notion-cli block update BLOCK_ID -c '{"paragraph":{"rich_text":[{"text":{"content":"Updated text"}}]}}'`,
    },
    {
        description: 'Update block color',
        command: `$ notion-cli block update BLOCK_ID --color blue`,
    },
    {
        description: 'Update a block and output raw json',
        command: `$ notion-cli block update BLOCK_ID -a -r`,
    },
];
BlockUpdate.args = {
    block_id: core_1.Args.string({ description: 'block_id', required: true }),
};
BlockUpdate.flags = {
    archived: core_1.Flags.boolean({
        char: 'a',
        description: 'Archive the block',
    }),
    content: core_1.Flags.string({
        char: 'c',
        description: 'Updated block content (JSON object with block type properties)',
    }),
    color: core_1.Flags.string({
        description: 'Block color (for supported block types)',
        options: ['default', 'gray', 'brown', 'orange', 'yellow', 'green', 'blue', 'purple', 'pink', 'red'],
    }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
};
