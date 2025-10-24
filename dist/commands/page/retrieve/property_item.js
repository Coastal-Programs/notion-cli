"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../../notion");
const helper_1 = require("../../../helper");
const base_flags_1 = require("../../../base-flags");
const errors_1 = require("../../../errors");
class PageRetrievePropertyItem extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(PageRetrievePropertyItem);
        try {
            const res = await notion.retrievePageProperty(args.page_id, args.property_id);
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
            // Handle raw JSON output (default for this command)
            (0, helper_1.outputRawJson)(res);
            process.exit(0);
        }
        catch (error) {
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    resourceType: 'page',
                    attemptedId: args.page_id,
                    endpoint: 'pages.properties.retrieve'
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
exports.default = PageRetrievePropertyItem;
PageRetrievePropertyItem.description = 'Retrieve a page property item';
PageRetrievePropertyItem.aliases = ['page:r:pi'];
PageRetrievePropertyItem.examples = [
    {
        description: 'Retrieve a page property item',
        command: `$ notion-cli page retrieve:property_item PAGE_ID PROPERTY_ID`,
    },
    {
        description: 'Retrieve a page property item and output raw json',
        command: `$ notion-cli page retrieve:property_item PAGE_ID PROPERTY_ID -r`,
    },
    {
        description: 'Retrieve a page property item and output JSON for automation',
        command: `$ notion-cli page retrieve:property_item PAGE_ID PROPERTY_ID --json`,
    },
];
PageRetrievePropertyItem.args = {
    page_id: core_1.Args.string({ required: true }),
    property_id: core_1.Args.string({ required: true }),
};
PageRetrievePropertyItem.flags = {
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'output raw json',
    }),
    ...base_flags_1.AutomationFlags,
};
