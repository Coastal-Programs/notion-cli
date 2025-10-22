"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../../notion");
const helper_1 = require("../../../helper");
const base_flags_1 = require("../../../base-flags");
const errors_1 = require("../../../errors");
class BlockRetrieveChildren extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(BlockRetrieveChildren);
        try {
            // TODO: Add support start_cursor, page_size
            const res = await notion.retrieveBlockChildren(args.block_id);
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
    {
        description: 'Retrieve block children and output JSON for automation',
        command: `$ notion-cli block retrieve:children BLOCK_ID --json`,
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
    ...base_flags_1.AutomationFlags,
};
