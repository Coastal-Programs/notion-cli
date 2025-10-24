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
const property_expander_1 = require("../../utils/property-expander");
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
            // Build properties object
            let properties = {};
            // Handle properties flag
            if (flags.properties) {
                try {
                    const parsedProps = JSON.parse(flags.properties);
                    if (flags['simple-properties']) {
                        // User provided simple format - expand to Notion format
                        // Need to get database schema first
                        if (!flags.parent_data_source_id) {
                            throw new Error('The --simple-properties flag requires --parent_data_source_id (-d) to be set. ' +
                                'Simple properties need the database schema for validation.');
                        }
                        const parentDataSourceId = await (0, notion_resolver_1.resolveNotionId)(flags.parent_data_source_id, 'database');
                        const dbSchema = await notion.retrieveDataSource(parentDataSourceId);
                        properties = await (0, property_expander_1.expandSimpleProperties)(parsedProps, dbSchema.properties);
                    }
                    else {
                        // Use raw Notion format
                        properties = parsedProps;
                    }
                }
                catch (error) {
                    if (error.message.includes('Unexpected token') || error.message.includes('JSON')) {
                        throw new Error(`Invalid JSON in --properties flag: ${error.message}\n` +
                            `Example: --properties '{"Name": "Task", "Status": "Done"}'`);
                    }
                    throw error;
                }
            }
            if (flags.file_path) {
                const p = path.join('./', flags.file_path);
                const fileName = path.basename(flags.file_path);
                const md = fs.readFileSync(p, { encoding: 'utf-8' });
                const blocks = (0, martian_1.markdownToBlocks)(md);
                // Extract title from H1 heading or use filename without extension
                const extractTitle = (markdown, filename) => {
                    const h1Match = markdown.match(/^#\s+(.+)$/m);
                    if (h1Match && h1Match[1]) {
                        return h1Match[1].trim();
                    }
                    // Fallback: use filename without extension
                    return filename.replace(/\.md$/, '');
                };
                const pageTitle = extractTitle(md, fileName);
                // If no properties were provided via flag, use extracted title
                if (!flags.properties) {
                    properties = {
                        [flags.title_property]: {
                            title: [{ text: { content: pageTitle } }],
                        },
                    };
                }
                else {
                    // Merge with existing properties, but ensure title is set
                    if (!properties[flags.title_property]) {
                        properties[flags.title_property] = {
                            title: [{ text: { content: pageTitle } }],
                        };
                    }
                }
                pageProps = {
                    parent: pageParent,
                    properties,
                    children: blocks,
                };
            }
            else {
                pageProps = {
                    parent: pageParent,
                    properties,
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
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    resourceType: 'page',
                    endpoint: 'pages.create'
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
        description: 'Create a page with simple properties (recommended for AI agents)',
        command: `$ notion-cli page create -d DATA_SOURCE_ID -S --properties '{"Name": "My Task", "Status": "In Progress", "Due Date": "2025-12-31"}'`,
    },
    {
        description: 'Create a page with simple properties using relative dates',
        command: `$ notion-cli page create -d DATA_SOURCE_ID -S --properties '{"Name": "Review", "Due Date": "tomorrow", "Priority": "High"}'`,
    },
    {
        description: 'Create a page with simple properties and multi-select',
        command: `$ notion-cli page create -d DATA_SOURCE_ID -S --properties '{"Name": "Bug Fix", "Tags": ["urgent", "bug"], "Status": "Done"}'`,
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
    properties: core_1.Flags.string({
        description: 'Page properties as JSON string',
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
