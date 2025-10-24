"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
const client_1 = require("@notionhq/client");
const readline = require("readline");
class BatchRetrieve extends core_1.Command {
    /**
     * Read IDs from stdin
     */
    async readStdin() {
        return new Promise((resolve, reject) => {
            const ids = [];
            const rl = readline.createInterface({
                input: process.stdin,
                output: process.stdout,
                terminal: false,
            });
            rl.on('line', (line) => {
                const trimmed = line.trim();
                if (trimmed) {
                    ids.push(trimmed);
                }
            });
            rl.on('close', () => {
                resolve(ids);
            });
            rl.on('error', (err) => {
                reject(err);
            });
            // Timeout after 5 seconds if no input
            setTimeout(() => {
                rl.close();
                resolve(ids);
            }, 5000);
        });
    }
    /**
     * Retrieve a single resource and handle errors
     */
    async retrieveResource(id, type) {
        try {
            let data;
            switch (type) {
                case 'page': {
                    const pageResponse = await notion.retrievePage({ page_id: id });
                    if (!(0, client_1.isFullPage)(pageResponse)) {
                        throw new errors_1.NotionCLIError(errors_1.NotionCLIErrorCode.API_ERROR, 'Received partial page response instead of full page', [], { attemptedId: id });
                    }
                    data = pageResponse;
                    break;
                }
                case 'block': {
                    const blockResponse = await notion.retrieveBlock(id);
                    if (!(0, client_1.isFullBlock)(blockResponse)) {
                        throw new errors_1.NotionCLIError(errors_1.NotionCLIErrorCode.API_ERROR, 'Received partial block response instead of full block', [], { attemptedId: id });
                    }
                    data = blockResponse;
                    break;
                }
                case 'database':
                    data = await notion.retrieveDataSource(id);
                    break;
                default:
                    throw new errors_1.NotionCLIError(errors_1.NotionCLIErrorCode.VALIDATION_ERROR, `Invalid resource type: ${type}`, [], { userInput: type, resourceType: type });
            }
            return {
                id,
                success: true,
                data,
            };
        }
        catch (error) {
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    attemptedId: id,
                    userInput: id
                });
            return {
                id,
                success: false,
                error: cliError.code,
                message: cliError.userMessage,
            };
        }
    }
    async run() {
        const { args, flags } = await this.parse(BatchRetrieve);
        try {
            // Get IDs from args, flags, or stdin
            let ids = [];
            if (args.ids) {
                // From positional argument
                ids = args.ids.split(',').map(id => id.trim()).filter(id => id);
            }
            else if (flags.ids) {
                // From --ids flag
                ids = flags.ids.split(',').map(id => id.trim()).filter(id => id);
            }
            else if (!process.stdin.isTTY) {
                // From stdin
                ids = await this.readStdin();
            }
            if (ids.length === 0) {
                throw new errors_1.NotionCLIError(errors_1.NotionCLIErrorCode.VALIDATION_ERROR, 'No IDs provided. Use --ids flag, positional argument, or pipe IDs via stdin', [
                    {
                        description: 'Provide IDs via --ids flag',
                        command: 'notion-cli batch retrieve --ids ID1,ID2,ID3'
                    },
                    {
                        description: 'Or pipe IDs from a file',
                        command: 'cat ids.txt | notion-cli batch retrieve'
                    }
                ]);
            }
            // Fetch all resources in parallel
            const results = await Promise.all(ids.map(id => this.retrieveResource(id, flags.type)));
            // Count successes and failures
            const successCount = results.filter(r => r.success).length;
            const failureCount = results.filter(r => !r.success).length;
            // Handle JSON output for automation (takes precedence)
            if (flags.json) {
                this.log(JSON.stringify({
                    success: successCount > 0,
                    total: results.length,
                    succeeded: successCount,
                    failed: failureCount,
                    results: results,
                    timestamp: new Date().toISOString(),
                }, null, 2));
                process.exit(failureCount === 0 ? 0 : 1);
                return;
            }
            // Handle compact JSON output
            if (flags['compact-json']) {
                (0, helper_1.outputCompactJson)({
                    total: results.length,
                    succeeded: successCount,
                    failed: failureCount,
                    results: results,
                });
                process.exit(failureCount === 0 ? 0 : 1);
                return;
            }
            // Handle raw JSON output
            if (flags.raw) {
                (0, helper_1.outputRawJson)(results);
                process.exit(failureCount === 0 ? 0 : 1);
                return;
            }
            // Handle table output (default)
            const tableData = results.map(result => {
                if (result.success && result.data) {
                    let title = '';
                    if ('object' in result.data) {
                        if (result.data.object === 'page') {
                            title = (0, helper_1.getPageTitle)(result.data);
                        }
                        else if (result.data.object === 'data_source') {
                            title = (0, helper_1.getDataSourceTitle)(result.data);
                        }
                        else if (result.data.object === 'block') {
                            title = (0, helper_1.getBlockPlainText)(result.data);
                        }
                    }
                    return {
                        id: result.id,
                        status: 'success',
                        type: result.data.object || flags.type,
                        title: title || '-',
                    };
                }
                else {
                    return {
                        id: result.id,
                        status: 'failed',
                        type: flags.type,
                        title: result.message || result.error || 'Unknown error',
                    };
                }
            });
            const columns = {
                id: {},
                status: {},
                type: {},
                title: {},
            };
            const options = {
                printLine: this.log.bind(this),
                ...flags,
            };
            core_1.ux.table(tableData, columns, options);
            // Print summary
            this.log(`\nTotal: ${results.length} | Succeeded: ${successCount} | Failed: ${failureCount}`);
            process.exit(failureCount === 0 ? 0 : 1);
        }
        catch (error) {
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    endpoint: 'batch.retrieve'
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
exports.default = BatchRetrieve;
BatchRetrieve.description = 'Batch retrieve multiple pages, blocks, or data sources';
BatchRetrieve.aliases = ['batch:r'];
BatchRetrieve.examples = [
    {
        description: 'Retrieve multiple pages via --ids flag',
        command: '$ notion-cli batch retrieve --ids PAGE_ID_1,PAGE_ID_2,PAGE_ID_3 --compact-json',
    },
    {
        description: 'Retrieve multiple pages from stdin (one ID per line)',
        command: '$ cat page_ids.txt | notion-cli batch retrieve --compact-json',
    },
    {
        description: 'Retrieve multiple blocks',
        command: '$ notion-cli batch retrieve --ids BLOCK_ID_1,BLOCK_ID_2 --type block --json',
    },
    {
        description: 'Retrieve multiple data sources',
        command: '$ notion-cli batch retrieve --ids DS_ID_1,DS_ID_2 --type database --json',
    },
    {
        description: 'Retrieve with raw output',
        command: '$ notion-cli batch retrieve --ids ID1,ID2,ID3 -r',
    },
];
BatchRetrieve.args = {
    ids: core_1.Args.string({
        required: false,
        description: 'Comma-separated list of IDs to retrieve (or use --ids flag or stdin)',
    }),
};
BatchRetrieve.flags = {
    ids: core_1.Flags.string({
        description: 'Comma-separated list of IDs to retrieve',
    }),
    type: core_1.Flags.string({
        description: 'Resource type to retrieve (page, block, database)',
        options: ['page', 'block', 'database'],
        default: 'page',
    }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json (recommended for AI assistants - returns all fields)',
    }),
    ...core_1.ux.table.flags(),
    ...base_flags_1.OutputFormatFlags,
    ...base_flags_1.AutomationFlags,
};
