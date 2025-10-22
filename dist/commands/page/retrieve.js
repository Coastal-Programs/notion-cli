"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const notion_to_md_1 = require("notion-to-md");
const base_flags_1 = require("../../base-flags");
class PageRetrieve extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(PageRetrieve);
        // Handle page content as markdown (uses NotionToMarkdown)
        if (flags.markdown) {
            const n2m = new notion_to_md_1.NotionToMarkdown({ notionClient: notion.client });
            const mdBlocks = await n2m.pageToMarkdown(args.page_id);
            const mdString = n2m.toMarkdownString(mdBlocks);
            console.log(mdString.parent);
            this.exit(0);
        }
        const pageProps = {
            page_id: args.page_id,
        };
        const res = await notion.retrievePage(pageProps);
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
            this.exit(0);
        }
        // Handle pretty table output
        if (flags.pretty) {
            (0, helper_1.outputPrettyTable)([res], columns);
            this.exit(0);
        }
        // Handle raw JSON output
        if (flags.raw) {
            (0, helper_1.outputRawJson)(res);
            this.exit(0);
        }
        // Handle table output (default)
        const options = {
            printLine: this.log.bind(this),
            ...flags,
        };
        core_1.ux.table([res], columns, options);
    }
}
exports.default = PageRetrieve;
PageRetrieve.description = 'Retrieve a page';
PageRetrieve.aliases = ['page:r'];
PageRetrieve.examples = [
    {
        description: 'Retrieve a page and output table',
        command: `$ notion-cli page retrieve PAGE_ID`,
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
];
PageRetrieve.args = {
    page_id: core_1.Args.string({ required: true }),
};
PageRetrieve.flags = {
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    markdown: core_1.Flags.boolean({
        char: 'm',
        description: 'output page content as markdown',
    }),
    ...core_1.ux.table.flags(),
    ...base_flags_1.OutputFormatFlags,
};
