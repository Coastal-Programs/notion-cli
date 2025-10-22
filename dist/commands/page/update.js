"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const notion_resolver_1 = require("../../utils/notion-resolver");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
class PageUpdate extends core_1.Command {
    // NOTE: Support only archived or un archive property for now
    // TODO: Add support for updating a page properties, icon, cover
    async run() {
        const { args, flags } = await this.parse(PageUpdate);
        try {
            // Resolve ID from URL, direct ID, or name (future)
            const pageId = await (0, notion_resolver_1.resolveNotionId)(args.page_id, 'page');
            const pageProps = {
                page_id: pageId,
            };
            if (flags.archived) {
                pageProps.archived = true;
            }
            if (flags.unarchive) {
                pageProps.archived = false;
            }
            const res = await notion.updatePageProps(pageProps);
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
exports.default = PageUpdate;
PageUpdate.description = 'Update a page';
PageUpdate.aliases = ['page:u'];
PageUpdate.examples = [
    {
        description: 'Update a page and output table',
        command: `$ notion-cli page update PAGE_ID`,
    },
    {
        description: 'Update a page via URL',
        command: `$ notion-cli page update https://notion.so/PAGE_ID -a`,
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
    {
        description: 'Update a page and output JSON for automation',
        command: `$ notion-cli page update PAGE_ID -a --json`,
    },
];
PageUpdate.args = {
    page_id: core_1.Args.string({
        required: true,
        description: 'Page ID or full Notion URL (e.g., https://notion.so/...)',
    }),
};
PageUpdate.flags = {
    archived: core_1.Flags.boolean({ char: 'a', description: 'Archive the page' }),
    unarchive: core_1.Flags.boolean({ char: 'u', description: 'Unarchive the page' }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
    ...base_flags_1.AutomationFlags,
};
