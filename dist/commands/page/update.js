"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../notion");
const helper_1 = require("../../helper");
const notion_resolver_1 = require("../../utils/notion-resolver");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
const property_expander_1 = require("../../utils/property-expander");
class PageUpdate extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(PageUpdate);
        try {
            // Resolve ID from URL, direct ID, or name (future)
            const pageId = await (0, notion_resolver_1.resolveNotionId)(args.page_id, 'page');
            const pageProps = {
                page_id: pageId,
            };
            // Handle archived flags
            if (flags.archived) {
                pageProps.archived = true;
            }
            if (flags.unarchive) {
                pageProps.archived = false;
            }
            // Handle properties update
            if (flags.properties) {
                try {
                    const parsedProps = JSON.parse(flags.properties);
                    if (flags['simple-properties']) {
                        // User provided simple format - expand to Notion format
                        // Need to get the page first to find its parent database
                        const page = await notion.retrievePage({ page_id: pageId });
                        // Check if page is in a database
                        if (!('parent' in page) || !('data_source_id' in page.parent)) {
                            throw new Error('The --simple-properties flag can only be used with pages in a database. ' +
                                'This page does not have a parent database.');
                        }
                        // Get the database schema
                        const parentDataSourceId = page.parent.data_source_id;
                        const dbSchema = await notion.retrieveDataSource(parentDataSourceId);
                        // Expand simple properties to Notion format
                        pageProps.properties = await (0, property_expander_1.expandSimpleProperties)(parsedProps, dbSchema.properties);
                    }
                    else {
                        // Use raw Notion format
                        pageProps.properties = parsedProps;
                    }
                }
                catch (error) {
                    if (error.message.includes('Unexpected token') || error.message.includes('JSON')) {
                        throw new Error(`Invalid JSON in --properties flag: ${error.message}\n` +
                            `Example: --properties '{"Status": "Done", "Priority": "High"}'`);
                    }
                    throw error;
                }
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
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    resourceType: 'page',
                    attemptedId: args.page_id,
                    endpoint: 'pages.update'
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
        description: 'Update page properties with simple format (recommended for AI agents)',
        command: `$ notion-cli page update PAGE_ID -S --properties '{"Status": "Done", "Priority": "High"}'`,
    },
    {
        description: 'Update page properties with relative date',
        command: `$ notion-cli page update PAGE_ID -S --properties '{"Due Date": "tomorrow", "Status": "In Progress"}'`,
    },
    {
        description: 'Update page with multi-select tags',
        command: `$ notion-cli page update PAGE_ID -S --properties '{"Tags": ["urgent", "bug"], "Status": "Done"}'`,
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
    properties: core_1.Flags.string({
        description: 'Page properties to update as JSON string',
    }),
    'simple-properties': core_1.Flags.boolean({
        char: 'S',
        description: 'Use simplified property format (flat key-value pairs, recommended for AI agents)',
        default: false,
    }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...core_1.ux.table.flags(),
    ...base_flags_1.AutomationFlags,
};
