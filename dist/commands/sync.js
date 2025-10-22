"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion_1 = require("../notion");
const retry_1 = require("../retry");
const workspace_cache_1 = require("../utils/workspace-cache");
const base_flags_1 = require("../base-flags");
const errors_1 = require("../errors");
class Sync extends core_1.Command {
    async run() {
        const { flags } = await this.parse(Sync);
        try {
            if (!flags.json) {
                core_1.ux.action.start('Syncing workspace databases');
            }
            // Fetch all databases from Notion API
            const databases = await this.fetchAllDatabases();
            if (!flags.json) {
                core_1.ux.action.stop(`Found ${databases.length} database${databases.length === 1 ? '' : 's'}`);
                core_1.ux.action.start('Generating search aliases');
            }
            // Build cache entries
            const cacheEntries = databases.map(db => (0, workspace_cache_1.buildCacheEntry)(db));
            if (!flags.json) {
                core_1.ux.action.stop();
                core_1.ux.action.start('Saving cache');
            }
            // Save to cache
            const cache = {
                version: '1.0.0',
                lastSync: new Date().toISOString(),
                databases: cacheEntries,
            };
            await (0, workspace_cache_1.saveCache)(cache);
            const cachePath = await (0, workspace_cache_1.getCachePath)();
            if (flags.json) {
                this.log(JSON.stringify({
                    success: true,
                    count: databases.length,
                    cachePath,
                    databases: cacheEntries.map(db => ({
                        id: db.id,
                        title: db.title,
                        aliases: db.aliases,
                        url: db.url,
                    })),
                    timestamp: new Date().toISOString(),
                }, null, 2));
            }
            else {
                core_1.ux.action.stop();
                this.log(`\n✓ Cache saved to ${cachePath}\n`);
                if (databases.length > 0) {
                    this.log('Indexed databases:');
                    cacheEntries.slice(0, 10).forEach(db => {
                        const aliasesStr = db.aliases.slice(0, 3).join(', ');
                        this.log(`  • ${db.title} (aliases: ${aliasesStr})`);
                    });
                    if (databases.length > 10) {
                        this.log(`  ... and ${databases.length - 10} more`);
                    }
                }
                else {
                    this.log('No databases found in workspace.');
                    this.log('Make sure your integration has access to databases.');
                }
            }
            process.exit(0);
        }
        catch (error) {
            const cliError = (0, errors_1.wrapNotionError)(error);
            if (flags.json) {
                this.log(JSON.stringify(cliError.toJSON(), null, 2));
            }
            else {
                core_1.ux.action.stop('failed');
                this.error(cliError.message);
            }
            process.exit(1);
        }
    }
    /**
     * Fetch all databases from Notion API with pagination
     */
    async fetchAllDatabases() {
        const databases = [];
        let cursor = undefined;
        while (true) {
            const response = await (0, retry_1.fetchWithRetry)(() => notion_1.client.search({
                filter: {
                    value: 'data_source',
                    property: 'object',
                },
                start_cursor: cursor,
                page_size: 100, // Max allowed by API
            }), {
                context: 'sync:fetchAllDatabases',
                config: { maxRetries: 5 }, // Higher retries for sync
            });
            databases.push(...response.results);
            if (!response.has_more || !response.next_cursor) {
                break;
            }
            cursor = response.next_cursor;
        }
        return databases;
    }
}
exports.default = Sync;
Sync.description = 'Sync workspace databases to local cache for fast lookups';
Sync.aliases = ['db:sync'];
Sync.examples = [
    {
        description: 'Sync all workspace databases',
        command: 'notion-cli sync',
    },
    {
        description: 'Force resync even if cache exists',
        command: 'notion-cli sync --force',
    },
    {
        description: 'Sync and output as JSON',
        command: 'notion-cli sync --json',
    },
];
Sync.flags = {
    force: core_1.Flags.boolean({
        char: 'f',
        description: 'Force resync even if cache is fresh',
        default: false,
    }),
    ...base_flags_1.AutomationFlags,
};
