"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
class BlockRetrieve extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(BlockRetrieve);
        try {
            let res = await notion.retrieveBlock(args.block_id);
            // Apply minimal flag to strip metadata
            if (flags.minimal) {
                res = (0, helper_1.stripMetadata)(res);
            }
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
                    endpoint: 'blocks.retrieve'
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
    {
        description: 'Retrieve a block and output JSON for automation',
        command: `$ notion-cli block retrieve BLOCK_ID --json`,
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
    ...base_flags_1.AutomationFlags,
};
