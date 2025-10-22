"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const fs = require("fs");
const path = require("path");
const martian_1 = require("@tryfabric/martian");
const helper_1 = require("../../helper");
class PageCreate extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(PageCreate);
        let pageProps;
        let pageParent;
        if (flags.parent_page_id) {
            pageParent = {
                page_id: flags.parent_page_id,
            };
        }
        else {
            pageParent = {
                data_source_id: flags.parent_data_source_id,
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
        if (flags.raw) {
            (0, helper_1.outputRawJson)(res);
            this.exit(0);
        }
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
];
PageCreate.flags = {
    parent_page_id: core_1.Flags.string({
        char: 'p',
        description: 'Parent page ID (to create a sub-page)',
    }),
    parent_data_source_id: core_1.Flags.string({
        char: 'd',
        description: 'Parent data source ID (to create a page in a table)',
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
};
