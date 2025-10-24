"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
class BlockDelete extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(BlockDelete);
        try {
            const res = await notion.deleteBlock(args.block_id);
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
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    resourceType: 'block',
                    attemptedId: args.block_id,
                    endpoint: 'blocks.delete'
                });
            if (flags.json) {
                this.log(JSON.stringify(cliError.toJSON(), null, 2));
            }
            else {
                this.error(cliError.toHumanString());
            }
            process.exit(1);
        }
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
    {
        description: 'Delete a block and output JSON for automation',
        command: `$ notion-cli block delete BLOCK_ID --json`,
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
    ...base_flags_1.AutomationFlags,
};
