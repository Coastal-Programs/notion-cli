"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const readline = require("readline");
const base_flags_1 = require("../base-flags");
const errors_1 = require("../errors");
const token_validator_1 = require("../utils/token-validator");
const notion_1 = require("../notion");
const workspace_cache_1 = require("../utils/workspace-cache");
const terminal_banner_1 = require("../utils/terminal-banner");
/**
 * Interactive first-time setup wizard for Notion CLI
 *
 * Guides new users through:
 * 1. Token configuration
 * 2. Connection testing
 * 3. Workspace synchronization
 *
 * Designed to provide a welcoming, educational experience that sets users up for success.
 */
class Init extends core_1.Command {
    constructor() {
        super(...arguments);
        this.isJsonMode = false;
    }
    async run() {
        const { flags } = await this.parse(Init);
        this.isJsonMode = flags.json;
        try {
            // Check if already configured
            const alreadyConfigured = await this.checkExistingSetup();
            if (alreadyConfigured && !this.isJsonMode) {
                const shouldReconfigure = await this.promptReconfigure();
                if (!shouldReconfigure) {
                    this.log('\nSetup cancelled. Your existing configuration is unchanged.');
                    process.exit(0);
                }
            }
            // Welcome message
            if (!this.isJsonMode) {
                this.showWelcome();
            }
            // Step 1: Configure token
            const tokenResult = await this.setupToken();
            // Step 2: Test connection
            const connectionResult = await this.testConnection();
            // Step 3: Sync workspace
            const syncResult = await this.syncWorkspace();
            // Success summary
            await this.showSuccess(tokenResult, connectionResult, syncResult);
            process.exit(0);
        }
        catch (error) {
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error, {
                    endpoint: 'init'
                });
            if (this.isJsonMode) {
                this.log(JSON.stringify(cliError.toJSON(), null, 2));
            }
            else {
                this.error(cliError.toHumanString());
            }
            process.exit(1);
        }
    }
    /**
     * Check if user already has a configured token
     */
    async checkExistingSetup() {
        if (!process.env.NOTION_TOKEN) {
            return false;
        }
        try {
            // Try to validate token
            (0, token_validator_1.validateNotionToken)();
            await (0, notion_1.botUser)();
            return true;
        }
        catch {
            // Token exists but is invalid
            return false;
        }
    }
    /**
     * Prompt user if they want to reconfigure
     */
    async promptReconfigure() {
        this.log('\nYou already have a configured Notion token.');
        this.log('Running init again will update your configuration.');
        this.log('');
        const rl = readline.createInterface({
            input: process.stdin,
            output: process.stdout,
        });
        const answer = await new Promise((resolve) => {
            rl.question('Do you want to reconfigure? (y/n): ', (answer) => {
                rl.close();
                resolve(answer.trim().toLowerCase());
            });
        });
        return answer === 'y' || answer === 'yes';
    }
    /**
     * Show welcome message
     */
    showWelcome() {
        this.log(terminal_banner_1.ASCII_BANNER);
        this.log(`${terminal_banner_1.colors.blue}Welcome to Notion CLI Setup!${terminal_banner_1.colors.reset}\n`);
        this.log('This wizard will help you set up your Notion CLI in 3 steps:');
        this.log(`  ${terminal_banner_1.colors.dim}1.${terminal_banner_1.colors.reset} Configure your Notion integration token`);
        this.log(`  ${terminal_banner_1.colors.dim}2.${terminal_banner_1.colors.reset} Test the connection to Notion API`);
        this.log(`  ${terminal_banner_1.colors.dim}3.${terminal_banner_1.colors.reset} Sync your workspace databases`);
        this.log('');
        this.log('Let\'s get started!');
        this.log('');
    }
    /**
     * Step 1: Setup token
     */
    async setupToken() {
        const stepNum = 1;
        const stepTotal = 3;
        if (!this.isJsonMode) {
            this.log('='.repeat(60));
            this.log(`Step ${stepNum}/${stepTotal}: Set your Notion token`);
            this.log('='.repeat(60));
            this.log('');
            this.log('You need a Notion integration token to use this CLI.');
            this.log('Get one at: https://www.notion.so/my-integrations');
            this.log('');
        }
        // Check if token already exists in environment
        if (process.env.NOTION_TOKEN && !this.isJsonMode) {
            this.log('Found existing NOTION_TOKEN in environment.');
            const rl = readline.createInterface({
                input: process.stdin,
                output: process.stdout,
            });
            const useExisting = await new Promise((resolve) => {
                rl.question('Use existing token? (y/n): ', (answer) => {
                    rl.close();
                    resolve(answer.trim().toLowerCase());
                });
            });
            if (useExisting === 'y' || useExisting === 'yes') {
                if (!this.isJsonMode) {
                    this.log('Using existing token from environment.');
                    this.log('');
                }
                return {
                    source: 'environment',
                    updated: false
                };
            }
        }
        // Get token from user
        let token;
        if (this.isJsonMode) {
            // In JSON mode, token must be in environment
            if (!process.env.NOTION_TOKEN) {
                throw new errors_1.NotionCLIError(errors_1.NotionCLIErrorCode.TOKEN_MISSING, 'NOTION_TOKEN required in JSON mode', [
                    {
                        description: 'Set token in environment before running init',
                        command: 'export NOTION_TOKEN="secret_your_token_here"'
                    }
                ]);
            }
            token = process.env.NOTION_TOKEN;
        }
        else {
            // Interactive token input
            const rl = readline.createInterface({
                input: process.stdin,
                output: process.stdout,
            });
            token = await new Promise((resolve) => {
                rl.question('Enter your Notion integration token (paste with or without "secret_" prefix): ', (answer) => {
                    rl.close();
                    resolve(answer.trim());
                });
            });
            // Validate token is not empty
            if (!token) {
                throw new errors_1.NotionCLIError(errors_1.NotionCLIErrorCode.TOKEN_INVALID, 'Token cannot be empty', [
                    {
                        description: 'Get your integration token from Notion',
                        link: 'https://developers.notion.com/docs/create-a-notion-integration'
                    }
                ]);
            }
            // Auto-prepend "secret_" if user didn't include it
            if (!token.startsWith('secret_')) {
                token = `secret_${token}`;
                this.log('');
                this.log(`${terminal_banner_1.colors.dim}Note: Automatically added "secret_" prefix to token${terminal_banner_1.colors.reset}`);
            }
            // Set token in current process for subsequent steps
            process.env.NOTION_TOKEN = token;
            this.log('');
            this.log('Token set for this session.');
            this.log('');
            this.log('Note: To persist this token, add it to your shell configuration:');
            this.log(`  export NOTION_TOKEN="${token}"`);
            this.log('');
            this.log('Or use: notion-cli config set-token');
            this.log('');
        }
        if (!this.isJsonMode) {
            this.log('Step 1 complete!');
            this.log('');
        }
        return {
            source: 'user_input',
            updated: true,
            tokenLength: token.length
        };
    }
    /**
     * Step 2: Test connection
     */
    async testConnection() {
        const stepNum = 2;
        const stepTotal = 3;
        if (!this.isJsonMode) {
            this.log('='.repeat(60));
            this.log(`Step ${stepNum}/${stepTotal}: Test connection`);
            this.log('='.repeat(60));
            this.log('');
            core_1.ux.action.start('Connecting to Notion API');
        }
        const startTime = Date.now();
        try {
            // Validate token and fetch bot info
            (0, token_validator_1.validateNotionToken)();
            const user = await (0, notion_1.botUser)();
            const latency = Date.now() - startTime;
            // Extract bot info
            const botInfo = {
                id: user.id,
                name: user.name || 'Unnamed Bot',
                type: user.type
            };
            let workspaceInfo = null;
            if (user.type === 'bot') {
                const botUser = user;
                if (botUser.bot && typeof botUser.bot === 'object' && 'owner' in botUser.bot) {
                    if (botUser.bot.workspace_name) {
                        workspaceInfo = {
                            name: botUser.bot.workspace_name,
                            id: botUser.bot.workspace_id,
                        };
                    }
                }
            }
            if (!this.isJsonMode) {
                core_1.ux.action.stop('connected');
                this.log('');
                this.log(`Bot Name: ${botInfo.name}`);
                this.log(`Bot ID: ${botInfo.id}`);
                if (workspaceInfo) {
                    this.log(`Workspace: ${workspaceInfo.name}`);
                }
                this.log(`Connection latency: ${latency}ms`);
                this.log('');
                this.log('Step 2 complete!');
                this.log('');
            }
            return {
                success: true,
                bot: botInfo,
                workspace: workspaceInfo,
                latency_ms: latency
            };
        }
        catch (error) {
            if (!this.isJsonMode) {
                core_1.ux.action.stop('failed');
            }
            throw (0, errors_1.wrapNotionError)(error, {
                endpoint: 'users.botUser',
                resourceType: 'user'
            });
        }
    }
    /**
     * Step 3: Sync workspace
     */
    async syncWorkspace() {
        var _a;
        const stepNum = 3;
        const stepTotal = 3;
        if (!this.isJsonMode) {
            this.log('='.repeat(60));
            this.log(`Step ${stepNum}/${stepTotal}: Sync workspace`);
            this.log('='.repeat(60));
            this.log('');
            this.log('This will index all databases your integration can access.');
            this.log('');
            core_1.ux.action.start('Syncing databases');
        }
        const startTime = Date.now();
        try {
            // Fetch all databases
            const databases = [];
            let cursor = undefined;
            while (true) {
                const response = await notion_1.client.search({
                    filter: {
                        value: 'data_source',
                        property: 'object',
                    },
                    start_cursor: cursor,
                    page_size: 100,
                });
                databases.push(...response.results);
                if (!this.isJsonMode && response.has_more) {
                    core_1.ux.action.start(`Syncing databases (found ${databases.length} so far)`);
                }
                if (!response.has_more || !response.next_cursor) {
                    break;
                }
                cursor = response.next_cursor;
            }
            const syncTime = Date.now() - startTime;
            if (!this.isJsonMode) {
                core_1.ux.action.stop(`found ${databases.length}`);
                this.log('');
                this.log(`Synced ${databases.length} database${databases.length === 1 ? '' : 's'} in ${(syncTime / 1000).toFixed(2)}s`);
                this.log('');
                if (databases.length > 0) {
                    this.log('Your integration has access to these databases:');
                    databases.slice(0, 5).forEach((db) => {
                        var _a, _b;
                        const title = ((_b = (_a = db.title) === null || _a === void 0 ? void 0 : _a[0]) === null || _b === void 0 ? void 0 : _b.plain_text) || 'Untitled';
                        this.log(`  - ${title}`);
                    });
                    if (databases.length > 5) {
                        this.log(`  ... and ${databases.length - 5} more`);
                    }
                    this.log('');
                }
                else {
                    this.log('No databases found.');
                    this.log('Make sure you\'ve shared databases with your integration.');
                    this.log('Learn more: https://developers.notion.com/docs/create-a-notion-integration#give-your-integration-page-permissions');
                    this.log('');
                }
                this.log('Step 3 complete!');
                this.log('');
            }
            // Try to load cache to show it's working
            const cache = await (0, workspace_cache_1.loadCache)();
            return {
                success: true,
                databases_found: databases.length,
                sync_time_ms: syncTime,
                cached: ((_a = cache === null || cache === void 0 ? void 0 : cache.databases) === null || _a === void 0 ? void 0 : _a.length) || 0
            };
        }
        catch (error) {
            if (!this.isJsonMode) {
                core_1.ux.action.stop('failed');
            }
            throw (0, errors_1.wrapNotionError)(error, {
                endpoint: 'search',
                resourceType: 'database'
            });
        }
    }
    /**
     * Show success summary
     */
    async showSuccess(tokenResult, connectionResult, syncResult) {
        if (this.isJsonMode) {
            this.log(JSON.stringify({
                success: true,
                message: 'Notion CLI setup complete',
                data: {
                    token: tokenResult,
                    connection: connectionResult,
                    sync: syncResult
                },
                next_steps: [
                    'notion-cli list - List all databases',
                    'notion-cli db query <name-or-id> - Query a database',
                    'notion-cli whoami - Check connection status',
                    'notion-cli sync - Refresh workspace cache',
                ],
                metadata: {
                    timestamp: new Date().toISOString(),
                    command: 'init'
                }
            }, null, 2));
        }
        else {
            this.log('='.repeat(60));
            this.log('  Setup Complete!');
            this.log('='.repeat(60));
            this.log('');
            this.log('Your Notion CLI is ready to use!');
            this.log('');
            this.log('Quick Start Commands:');
            this.log('  notion-cli list              - List all databases');
            this.log('  notion-cli db query <name>   - Query a database');
            this.log('  notion-cli whoami            - Check connection status');
            this.log('  notion-cli sync              - Refresh workspace cache');
            this.log('');
            this.log('Documentation:');
            this.log('  https://github.com/Coastal-Programs/notion-cli');
            this.log('');
            this.log('Need help? Run any command with --help flag');
            this.log('');
            this.log('Happy building with Notion!');
            this.log('');
        }
    }
}
exports.default = Init;
Init.description = 'Interactive first-time setup wizard for Notion CLI';
Init.examples = [
    {
        description: 'Run interactive setup wizard',
        command: '$ notion-cli init',
    },
    {
        description: 'Run setup with automated JSON output',
        command: '$ notion-cli init --json',
    },
];
Init.flags = {
    ...base_flags_1.AutomationFlags,
};
