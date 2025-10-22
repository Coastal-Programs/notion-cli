"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
class PageUpdate extends core_1.Command {
    // NOTE: Support only archived or un archive property for now
    // TODO: Add support for updating a page properties, icon, cover
    async run() {
        const { args, flags } = await this.parse(PageUpdate);
        const pageProps = {
            page_id: args.page_id,
        };
        if (flags.archived) {
            pageProps.archived = true;
        }
        if (flags.unarchive) {
            pageProps.archived = false;
        }
        const res = await notion.updatePageProps(pageProps);
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
exports.default = PageUpdate;
PageUpdate.description = 'Update a page';
PageUpdate.aliases = ['page:u'];
PageUpdate.examples = [
    {
        description: 'Update a page and output table',
        command: `$ notion-cli page update PAGE_ID`,
    },
    {
        description: 'Update a page and output raw json',
        command: `$ notion-cli page update PAGE_ID -r`,
    },
    {
        description: 'Update a page and archive',
        command: `$ notion-cli page update PAGE_ID -a`,
    },
    {
        description: 'Update a page and unarchive',
        command: `$ notion-cli page update PAGE_ID -u`,
    },
    {
        description: 'Update a page and archive and output raw json',
        command: `$ notion-cli page update PAGE_ID -a -r`,
    },
    {
        description: 'Update a page and unarchive and output raw json',
        command: `$ notion-cli page update PAGE_ID -u -r`,
    },
];
PageUpdate.args = {
    page_id: core_1.Args.string({ required: true }),
};
PageUpdate.flags = {
    archived: core_1.Flags.boolean({ char: 'a', description: 'Archive the page' }),
    unarchive: core_1.Flags.boolean({ char: 'u', description: 'Unarchive the page' }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
};
