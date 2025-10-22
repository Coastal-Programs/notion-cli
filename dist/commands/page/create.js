"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const fs = require("fs");
const path = require("path");
const martian_1 = require("@tryfabric/martian");
const helper_1 = require("../../helper");
const notion_resolver_1 = require("../../utils/notion-resolver");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
class PageCreate extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(PageCreate);
        try {
            let pageProps;
            let pageParent;
            if (flags.parent_page_id) {
                // Resolve parent page ID from URL, direct ID, or name (future)
                const parentPageId = await (0, notion_resolver_1.resolveNotionId)(flags.parent_page_id, 'page');
                pageParent = {
                    page_id: parentPageId,
                };
            }
            else {
                // Resolve parent database ID from URL, direct ID, or name (future)
                const parentDataSourceId = await (0, notion_resolver_1.resolveNotionId)(flags.parent_data_source_id, 'database');
                pageParent = {
                    data_source_id: parentDataSourceId,
                };
            }
            if (flags.file_path) {
                const p = path.join('./', flags.file_path);
                const fileName = path.basename(flags.file_path);
                const md = fs.readFileSync(p, { encoding: 'utf-8' });
                const blocks = (0, martian_1.markdownToBlocks)(md);
                // TODO: Add support for creating a page from a template
                pageProps = {
                    parent: pageParent,
                    properties: {
                        [flags.title_property]: {
                            title: [{ text: { content: fileName } }],
                        },
                    },
                    children: blocks,
                };
            }
            else {
                pageProps = {
                    parent: pageParent,
                    properties: {},
                };
            }
            const res = await notion.createPage(pageProps);
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
                title: {
                    get: (row) => {
                        return (0, helper_1.getPageTitle)(row);
                    },
                },
                object: {},
                id: {},
                url: {},
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
exports.default = PageCreate;
PageCreate.description = 'Create a page';
PageCreate.aliases = ['page:c'];
PageCreate.examples = [
    {
        description: 'Create a page via interactive mode',
        command: `$ notion-cli page create`,
    },
    {
        description: 'Create a page with a specific parent_page_id',
        command: `$ notion-cli page create -p PARENT_PAGE_ID`,
    },
    {
        description: 'Create a page with a parent page URL',
        command: `$ notion-cli page create -p https://notion.so/PARENT_PAGE_ID`,
    },
    {
        description: 'Create a page with a specific parent_db_id',
        command: `$ notion-cli page create -d PARENT_DB_ID`,
    },
    {
        description: 'Create a page with a specific source markdown file and parent_page_id',
        command: `$ notion-cli page create -f ./path/to/source.md -p PARENT_PAGE_ID`,
    },
    {
        description: 'Create a page with a specific source markdown file and parent_db_id',
        command: `$ notion-cli page create -f ./path/to/source.md -d PARENT_DB_ID`,
    },
    {
        description: 'Create a page with a specific source markdown file and output raw json with parent_page_id',
        command: `$ notion-cli page create -f ./path/to/source.md -p PARENT_PAGE_ID -r`,
    },
    {
        description: 'Create a page and output JSON for automation',
        command: `$ notion-cli page create -p PARENT_PAGE_ID --json`,
    },
];
PageCreate.flags = {
    parent_page_id: core_1.Flags.string({
        char: 'p',
        description: 'Parent page ID or URL (to create a sub-page)',
    }),
    parent_data_source_id: core_1.Flags.string({
        char: 'd',
        description: 'Parent data source ID or URL (to create a page in a table)',
    }),
    file_path: core_1.Flags.string({
        char: 'f',
        description: 'Path to a source markdown file',
    }),
    title_property: core_1.Flags.string({
        char: 't',
        description: 'Name of the title property (defaults to "Name" if not specified)',
        default: 'Name',
    }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
    ...base_flags_1.AutomationFlags,
};
