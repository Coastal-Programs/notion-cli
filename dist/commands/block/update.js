"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const notion_resolver_1 = require("../../utils/notion-resolver");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
class BlockUpdate extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(BlockUpdate);
        try {
            // Resolve block ID from URL or direct ID
            const blockId = await (0, notion_resolver_1.resolveNotionId)(args.block_id, 'page');
            const params = {
                block_id: blockId,
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
exports.default = BlockUpdate;
BlockUpdate.description = 'Update a block';
BlockUpdate.aliases = ['block:u'];
BlockUpdate.examples = [
    {
        description: 'Archive a block',
        command: `$ notion-cli block update BLOCK_ID -a`,
    },
    {
        description: 'Archive a block via URL',
        command: `$ notion-cli block update https://notion.so/BLOCK_ID -a`,
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
    {
        description: 'Update a block and output JSON for automation',
        command: `$ notion-cli block update BLOCK_ID -a --json`,
    },
];
BlockUpdate.args = {
    block_id: core_1.Args.string({ description: 'Block ID or URL', required: true }),
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
    ...base_flags_1.AutomationFlags,
};
