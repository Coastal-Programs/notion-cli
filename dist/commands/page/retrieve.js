"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const table_formatter_1 = require("../../utils/table-formatter");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const notion_to_md_1 = require("notion-to-md");
const base_flags_1 = require("../../base-flags");
const notion_resolver_1 = require("../../utils/notion-resolver");
const errors_1 = require("../../errors");
class PageRetrieve extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(PageRetrieve);
        try {
            // Resolve ID from URL, direct ID, or name (future)
            const pageId = await (0, notion_resolver_1.resolveNotionId)(args.page_id, 'page');
            // Handle map flag (fast structure discovery with parallel fetching)
            if (flags.map) {
                const mapData = await notion.mapPageStructure(pageId);
                // Handle JSON output for automation (takes precedence)
                if (flags.json) {
                    this.log(JSON.stringify({
                        success: true,
                        data: mapData,
                        timestamp: new Date().toISOString()
                    }, null, 2));
                    process.exit(0);
                    return;
                }
                // Handle compact JSON output
                if (flags['compact-json']) {
                    (0, helper_1.outputCompactJson)(mapData);
                    process.exit(0);
                    return;
                }
                // Default: pretty JSON output for map
                this.log(JSON.stringify(mapData, null, 2));
                process.exit(0);
                return;
            }
            // Handle page content as markdown (uses NotionToMarkdown)
            if (flags.markdown) {
                const n2m = new notion_to_md_1.NotionToMarkdown({ notionClient: notion.client });
                const mdBlocks = await n2m.pageToMarkdown(pageId);
                const mdString = n2m.toMarkdownString(mdBlocks);
                console.log(mdString.parent);
                process.exit(0);
                return;
            }
            // Handle recursive fetching
            if (flags.recursive) {
                const recursiveData = await notion.retrievePageRecursive(pageId, 0, flags['max-depth']);
                // Handle JSON output for automation (takes precedence)
                if (flags.json) {
                    this.log(JSON.stringify({
                        success: true,
                        data: recursiveData,
                        timestamp: new Date().toISOString()
                    }, null, 2));
                    process.exit(0);
                    return;
                }
                // Handle compact JSON output
                if (flags['compact-json']) {
                    (0, helper_1.outputCompactJson)(recursiveData);
                    process.exit(0);
                    return;
                }
                // Handle raw JSON output
                if (flags.raw) {
                    (0, helper_1.outputRawJson)(recursiveData);
                    process.exit(0);
                    return;
                }
                // For other formats, show a message that they're not supported with recursive
                this.error('Recursive mode only supports --json, --compact-json, or --raw output formats');
                process.exit(1);
                return;
            }
            const pageProps = {
                page_id: pageId,
            };
            let res = await notion.retrievePage(pageProps);
            // Apply minimal flag to strip metadata
            if (flags.minimal) {
                res = (0, helper_1.stripMetadata)(res);
            }
            // Handle JSON output for automation (takes precedence)
            if (flags.json) {
                this.log(JSON.stringify({
                    success: true,
                    data: res,
                    timestamp: new Date().toISOString()
                }, null, 2));
                process.exit(0);
                return;
            }
            // Define columns for table output
            const columns = {
                title: {
                    get: (row) => {
                        return (0, helper_1.getPageTitle)(row);
                    },
                },
                object: {},
                id: {},
                url: {},
            };
            // Handle compact JSON output
            if (flags['compact-json']) {
                (0, helper_1.outputCompactJson)(res);
                process.exit(0);
                return;
            }
            // Handle pretty table output
            if (flags.pretty) {
                (0, helper_1.outputPrettyTable)([res], columns);
                // Show hint after table output
                (0, helper_1.showRawFlagHint)(1, res);
                process.exit(0);
                return;
            }
            // Handle raw JSON output
            if (flags.raw) {
                (0, helper_1.outputRawJson)(res);
                process.exit(0);
                return;
            }
            // Handle table output (default)
            const options = {
                printLine: this.log.bind(this),
                ...flags,
            };
            (0, table_formatter_1.formatTable)([res], columns, options);
            // Show hint after table output to make -r flag discoverable
            (0, helper_1.showRawFlagHint)(1, res);
        }
        catch (error) {
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    resourceType: 'page',
                    attemptedId: args.page_id,
                    endpoint: 'pages.retrieve'
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
PageRetrieve.description = 'Retrieve a page';
PageRetrieve.aliases = ['page:r'];
PageRetrieve.examples = [
    {
        description: 'Retrieve a page with full data (recommended for AI assistants)',
        command: `$ notion-cli page retrieve PAGE_ID -r`,
    },
    {
        description: 'Fast structure overview (90% faster than full fetch)',
        command: `$ notion-cli page retrieve PAGE_ID --map`,
    },
    {
        description: 'Fast structure overview with compact JSON',
        command: `$ notion-cli page retrieve PAGE_ID --map --compact-json`,
    },
    {
        description: 'Retrieve entire page tree with all nested content (35% token reduction)',
        command: `$ notion-cli page retrieve PAGE_ID --recursive --compact-json`,
    },
    {
        description: 'Retrieve page tree with custom depth limit',
        command: `$ notion-cli page retrieve PAGE_ID -R --max-depth 5 --json`,
    },
    {
        description: 'Retrieve a page and output table',
        command: `$ notion-cli page retrieve PAGE_ID`,
    },
    {
        description: 'Retrieve a page via URL',
        command: `$ notion-cli page retrieve https://notion.so/PAGE_ID`,
    },
    {
        description: 'Retrieve a page and output raw json',
        command: `$ notion-cli page retrieve PAGE_ID -r`,
    },
    {
        description: 'Retrieve a page and output markdown',
        command: `$ notion-cli page retrieve PAGE_ID -m`,
    },
    {
        description: 'Retrieve a page metadata and output as markdown table',
        command: `$ notion-cli page retrieve PAGE_ID --markdown`,
    },
    {
        description: 'Retrieve a page metadata and output as compact JSON',
        command: `$ notion-cli page retrieve PAGE_ID --compact-json`,
    },
    {
        description: 'Retrieve a page and output JSON for automation',
        command: `$ notion-cli page retrieve PAGE_ID --json`,
    },
];
PageRetrieve.args = {
    page_id: core_1.Args.string({
        required: true,
        description: 'Page ID or full Notion URL (e.g., https://notion.so/...)',
    }),
};
PageRetrieve.flags = {
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json (recommended for AI assistants - returns all fields)',
    }),
    markdown: core_1.Flags.boolean({
        char: 'm',
        description: 'output page content as markdown',
    }),
    map: core_1.Flags.boolean({
        description: 'fast structure discovery (returns minimal info: titles, types, IDs)',
        default: false,
        exclusive: ['raw', 'markdown'],
    }),
    recursive: core_1.Flags.boolean({
        char: 'R',
        description: 'recursively fetch all blocks and nested pages (reduces API calls)',
        default: false,
    }),
    'max-depth': core_1.Flags.integer({
        description: 'maximum recursion depth for --recursive (default: 3)',
        default: 3,
        min: 1,
        max: 10,
        dependsOn: ['recursive'],
    }),
    ...table_formatter_1.tableFlags,
    ...base_flags_1.OutputFormatFlags,
    ...base_flags_1.AutomationFlags,
};
exports.default = PageRetrieve;
