"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion = require("../../../notion");
const helper_1 = require("../../../helper");
class PageRetrievePropertyItem extends core_1.Command {
    async run() {
        const { args } = await this.parse(PageRetrievePropertyItem);
        const res = await notion.retrievePageProperty(args.page_id, args.property_id);
        (0, helper_1.outputRawJson)(res);
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
];
PageRetrievePropertyItem.args = {
    page_id: core_1.Args.string({ required: true }),
    property_id: core_1.Args.string({ required: true }),
};
