"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.OutputFormatFlags = exports.AutomationFlags = void 0;
const core_1 = require("@oclif/core");
exports.AutomationFlags = {
    json: core_1.Flags.boolean({
        char: 'j',
        description: 'Output as JSON (recommended for automation)',
        default: false,
    }),
    'page-size': core_1.Flags.integer({
        description: 'Items per page (1-100, default: 100 for automation)',
        min: 1,
        max: 100,
        default: 100,
    }),
    retry: core_1.Flags.boolean({
        description: 'Auto-retry on rate limit (respects Retry-After header)',
        default: true,
    }),
    timeout: core_1.Flags.integer({
        description: 'Request timeout in milliseconds',
        default: 30000,
    }),
    'no-cache': core_1.Flags.boolean({
        description: 'Bypass cache and force fresh API calls',
        default: false,
    }),
    verbose: core_1.Flags.boolean({
        char: 'v',
        description: 'Enable verbose logging to stderr (retry events, cache stats) - never pollutes stdout',
        default: false,
        env: 'NOTION_CLI_VERBOSE',
    }),
    minimal: core_1.Flags.boolean({
        description: 'Strip unnecessary metadata (created_by, last_edited_by, object fields, request_id, etc.) - reduces response size by ~40%',
        default: false,
    }),
};
exports.OutputFormatFlags = {
    markdown: core_1.Flags.boolean({
        char: 'm',
        description: 'Output as markdown table (GitHub-flavored)',
        default: false,
        exclusive: ['compact-json', 'pretty'],
    }),
    'compact-json': core_1.Flags.boolean({
        char: 'c',
        description: 'Output as compact JSON (single-line, ideal for piping)',
        default: false,
        exclusive: ['markdown', 'pretty'],
    }),
    pretty: core_1.Flags.boolean({
        char: 'P',
        description: 'Output as pretty table with borders',
        default: false,
        exclusive: ['markdown', 'compact-json'],
    }),
};
