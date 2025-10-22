"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const notion_resolver_1 = require("../../utils/notion-resolver");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
class BlockAppend extends core_1.Command {
    // TODO: Add support children params building prompt
    async run() {
        const { flags } = await this.parse(BlockAppend);
        try {
            // Resolve block ID from URL or direct ID
            const blockId = await (0, notion_resolver_1.resolveNotionId)(flags.block_id, 'page');
            const params = {
                block_id: blockId,
                children: JSON.parse(flags.children),
            };
            if (flags.after) {
                // Resolve after block ID from URL or direct ID
                const afterBlockId = await (0, notion_resolver_1.resolveNotionId)(flags.after, 'page');
                params.after = afterBlockId;
            }
            const res = await notion.appendBlockChildren(params);
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
            core_1.ux.table(res.results, columns, options);
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
exports.default = BlockAppend;
BlockAppend.description = 'Append block children';
BlockAppend.aliases = ['block:a'];
BlockAppend.examples = [
    {
        description: 'Append block children',
        command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]'`,
    },
    {
        description: 'Append block children via URL',
        command: `$ notion-cli block append -b https://notion.so/BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]'`,
    },
    {
        description: 'Append block children after a block',
        command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]' -a AFTER_BLOCK_ID`,
    },
    {
        description: 'Append block children and output raw json',
        command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]' -r`,
    },
    {
        description: 'Append block children and output JSON for automation',
        command: `$ notion-cli block append -b BLOCK_ID -c '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello world!"}}]}}]' --json`,
    },
];
BlockAppend.flags = {
    block_id: core_1.Flags.string({
        char: 'b',
        description: 'Parent block ID or URL',
        required: true,
    }),
    children: core_1.Flags.string({
        char: 'c',
        description: 'Block children (JSON array)',
        required: true,
    }),
    after: core_1.Flags.string({
        char: 'a',
        description: 'Block ID or URL to append after (optional)',
    }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
    ...base_flags_1.AutomationFlags,
};
