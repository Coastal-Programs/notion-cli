"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
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
            // Handle page content as markdown (uses NotionToMarkdown)
            if (flags.markdown) {
                const n2m = new notion_to_md_1.NotionToMarkdown({ notionClient: notion.client });
                const mdBlocks = await n2m.pageToMarkdown(pageId);
                const mdString = n2m.toMarkdownString(mdBlocks);
                console.log(mdString.parent);
                process.exit(0);
                return;
            }
            const pageProps = {
                page_id: pageId,
            };
            const res = await notion.retrievePage(pageProps);
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
            core_1.ux.table([res], columns, options);
            // Show hint after table output to make -r flag discoverable
            (0, helper_1.showRawFlagHint)(1, res);
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
exports.default = PageRetrieve;
PageRetrieve.description = 'Retrieve a page';
PageRetrieve.aliases = ['page:r'];
PageRetrieve.examples = [
    {
        description: 'Retrieve a page with full data (recommended for AI assistants)',
        command: `$ notion-cli page retrieve PAGE_ID -r`,
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
    ...core_1.ux.table.flags(),
    ...base_flags_1.OutputFormatFlags,
    ...base_flags_1.AutomationFlags,
};
