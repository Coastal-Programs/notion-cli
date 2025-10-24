"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const notion_1 = require("../notion");
const workspace_cache_1 = require("../utils/workspace-cache");
const fs = require("fs/promises");
const https = require("https");
class Doctor extends core_1.Command {
    async run() {
        const { flags } = await this.parse(Doctor);
        const checks = [];
        // Run all health checks
        await this.checkNodeVersion(checks);
        await this.checkTokenSet(checks);
        await this.checkTokenFormat(checks);
        await this.checkNetworkConnectivity(checks);
        await this.checkApiConnection(checks);
        await this.checkCacheExists(checks);
        await this.checkCacheFreshness(checks);
        // Calculate summary
        const summary = {
            total: checks.length,
            passed: checks.filter(c => c.passed).length,
            failed: checks.filter(c => !c.passed).length,
        };
        const result = {
            success: summary.failed === 0,
            checks,
            summary,
        };
        // Output results
        if (flags.json) {
            this.log(JSON.stringify(result, null, 2));
        }
        else {
            this.printHumanReadable(result);
        }
        // Exit with appropriate code
        process.exit(result.success ? 0 : 1);
    }
    /**
     * Check Node.js version meets requirement (>=18.0.0)
     */
    async checkNodeVersion(checks) {
        try {
            const version = process.version;
            const major = parseInt(version.split('.')[0].replace('v', ''));
            const passed = major >= 18;
            checks.push({
                name: 'nodejs_version',
                passed,
                value: version,
                message: passed ? undefined : 'Node.js version must be >= 18.0.0',
                recommendation: passed ? undefined : 'Please upgrade Node.js to version 18 or higher',
            });
        }
        catch (error) {
            checks.push({
                name: 'nodejs_version',
                passed: false,
                message: 'Failed to check Node.js version',
            });
        }
    }
    /**
     * Check if NOTION_TOKEN environment variable is set
     */
    async checkTokenSet(checks) {
        const tokenSet = !!process.env.NOTION_TOKEN;
        checks.push({
            name: 'token_set',
            passed: tokenSet,
            message: tokenSet ? undefined : 'NOTION_TOKEN environment variable is not set',
            recommendation: tokenSet ? undefined : "Run 'notion-cli config set-token' or 'notion-cli init'",
        });
    }
    /**
     * Check if token format is valid
     * Accepts both "secret_" prefix (internal integrations) and "ntn_" prefix (OAuth tokens)
     */
    async checkTokenFormat(checks) {
        const token = process.env.NOTION_TOKEN;
        if (!token) {
            // Skip if token not set (already handled by checkTokenSet)
            checks.push({
                name: 'token_format',
                passed: false,
                message: 'Cannot check format - token not set',
            });
            return;
        }
        // Check for valid token formats
        // Internal integrations: secret_*
        // OAuth tokens: ntn_*
        // Also accept tokens that look like valid base64 or hex strings (length >= 32)
        const isValidFormat = token.startsWith('secret_') ||
            token.startsWith('ntn_') ||
            (token.length >= 32 && /^[A-Za-z0-9_-]+$/.test(token));
        if (!isValidFormat) {
            checks.push({
                name: 'token_format',
                passed: false,
                message: 'Token format appears invalid',
                recommendation: 'Notion tokens typically start with "secret_" or "ntn_". Please verify your token.',
            });
        }
        else {
            checks.push({
                name: 'token_format',
                passed: true,
            });
        }
    }
    /**
     * Check network connectivity to api.notion.com
     */
    async checkNetworkConnectivity(checks) {
        try {
            await this.checkHttpsConnection('api.notion.com', 443);
            checks.push({
                name: 'network_connectivity',
                passed: true,
            });
        }
        catch (error) {
            checks.push({
                name: 'network_connectivity',
                passed: false,
                message: 'Cannot reach api.notion.com',
                recommendation: 'Check your internet connection and firewall settings',
            });
        }
    }
    /**
     * Check if can connect to Notion API (whoami check)
     */
    async checkApiConnection(checks) {
        const token = process.env.NOTION_TOKEN;
        if (!token) {
            // Skip if token not set
            checks.push({
                name: 'api_connection',
                passed: false,
                message: 'Cannot test API - token not set',
            });
            return;
        }
        try {
            const user = await notion_1.client.users.me({});
            let botName = 'Unknown Bot';
            let workspaceName;
            if (user.type === 'bot') {
                const botUser = user;
                botName = user.name || 'Unnamed Bot';
                if (botUser.bot && typeof botUser.bot === 'object' && 'workspace_name' in botUser.bot) {
                    workspaceName = botUser.bot.workspace_name;
                }
            }
            checks.push({
                name: 'api_connection',
                passed: true,
                bot_name: botName,
                workspace_name: workspaceName,
            });
        }
        catch (error) {
            let message = 'Failed to connect to Notion API';
            let recommendation = 'Verify your NOTION_TOKEN is valid and active';
            if (error.code === 'unauthorized' || error.status === 401) {
                message = 'Authentication failed - invalid token';
                recommendation = "Run 'notion-cli config set-token' to update your token";
            }
            else if (error.code === 'ENOTFOUND' || error.code === 'ETIMEDOUT') {
                message = 'Network error - cannot reach Notion API';
                recommendation = 'Check your internet connection';
            }
            checks.push({
                name: 'api_connection',
                passed: false,
                message,
                recommendation,
            });
        }
    }
    /**
     * Check if workspace cache exists
     */
    async checkCacheExists(checks) {
        try {
            const cachePath = await (0, workspace_cache_1.getCachePath)();
            try {
                await fs.access(cachePath);
                checks.push({
                    name: 'cache_exists',
                    passed: true,
                    value: cachePath,
                });
            }
            catch {
                checks.push({
                    name: 'cache_exists',
                    passed: false,
                    message: 'Workspace cache does not exist',
                    recommendation: "Run 'notion-cli sync' to create cache",
                });
            }
        }
        catch (error) {
            checks.push({
                name: 'cache_exists',
                passed: false,
                message: 'Failed to check cache existence',
            });
        }
    }
    /**
     * Check if cache is fresh (< 24 hours old) or needs sync
     */
    async checkCacheFreshness(checks) {
        try {
            const cache = await (0, workspace_cache_1.loadCache)();
            if (!cache || !cache.lastSync) {
                checks.push({
                    name: 'cache_fresh',
                    passed: false,
                    message: 'Cache is empty or corrupted',
                    recommendation: "Run 'notion-cli sync' to rebuild cache",
                });
                return;
            }
            const lastSyncTime = new Date(cache.lastSync).getTime();
            const now = Date.now();
            const ageMs = now - lastSyncTime;
            const ageHours = ageMs / (1000 * 60 * 60);
            const ageDays = Math.floor(ageHours / 24);
            const remainingHours = Math.floor(ageHours % 24);
            const isFresh = ageHours < 24;
            let ageString;
            if (ageDays === 0) {
                ageString = `${Math.floor(ageHours)} hours ago`;
            }
            else if (ageDays === 1) {
                ageString = remainingHours === 0 ? '1 day ago' : `1 day, ${remainingHours} hours ago`;
            }
            else {
                ageString = remainingHours === 0 ? `${ageDays} days ago` : `${ageDays} days, ${remainingHours} hours ago`;
            }
            checks.push({
                name: 'cache_fresh',
                passed: isFresh,
                age_hours: Math.round(ageHours * 10) / 10,
                value: ageString,
                message: isFresh ? undefined : `Cache is outdated (last sync: ${ageString})`,
                recommendation: isFresh ? undefined : "Run 'notion-cli sync' to refresh",
            });
        }
        catch (error) {
            checks.push({
                name: 'cache_fresh',
                passed: false,
                message: 'Failed to check cache freshness',
                recommendation: "Run 'notion-cli sync' to refresh cache",
            });
        }
    }
    /**
     * Test HTTPS connection to a host
     */
    async checkHttpsConnection(host, port) {
        return new Promise((resolve, reject) => {
            const options = {
                host,
                port,
                method: 'GET',
                path: '/',
                timeout: 5000,
            };
            const req = https.request(options, (res) => {
                resolve();
            });
            req.on('error', (error) => {
                reject(error);
            });
            req.on('timeout', () => {
                req.destroy();
                reject(new Error('Connection timeout'));
            });
            req.end();
        });
    }
    /**
     * Print human-readable output
     */
    printHumanReadable(result) {
        this.log('\nNotion CLI Health Check');
        this.log('━'.repeat(50));
        this.log('');
        // Print each check
        for (const check of result.checks) {
            const icon = check.passed ? '✓' : '✗';
            const color = check.passed ? '\x1b[32m' : '\x1b[31m'; // Green or Red
            const reset = '\x1b[0m';
            switch (check.name) {
                case 'nodejs_version':
                    if (check.passed) {
                        this.log(`${color}${icon}${reset} Node.js version: ${check.value}`);
                    }
                    else {
                        this.log(`${color}${icon}${reset} Node.js version: ${check.value || 'unknown'} (${check.message})`);
                    }
                    break;
                case 'token_set':
                    if (check.passed) {
                        this.log(`${color}${icon}${reset} NOTION_TOKEN is set`);
                    }
                    else {
                        this.log(`${color}${icon}${reset} NOTION_TOKEN is not set`);
                    }
                    break;
                case 'token_format':
                    if (check.passed) {
                        this.log(`${color}${icon}${reset} Token format is valid`);
                    }
                    else {
                        this.log(`${color}${icon}${reset} Token format is invalid (${check.message})`);
                    }
                    break;
                case 'network_connectivity':
                    if (check.passed) {
                        this.log(`${color}${icon}${reset} Network connectivity to api.notion.com`);
                    }
                    else {
                        this.log(`${color}${icon}${reset} Cannot reach api.notion.com`);
                    }
                    break;
                case 'api_connection':
                    if (check.passed) {
                        this.log(`${color}${icon}${reset} API connection successful`);
                        if (check.bot_name) {
                            const workspaceInfo = check.workspace_name ? ` (${check.workspace_name})` : '';
                            this.log(`${color}${icon}${reset} Connected as: ${check.bot_name}${workspaceInfo}`);
                        }
                    }
                    else {
                        this.log(`${color}${icon}${reset} API connection failed (${check.message})`);
                    }
                    break;
                case 'cache_exists':
                    if (check.passed) {
                        this.log(`${color}${icon}${reset} Workspace cache exists`);
                    }
                    else {
                        this.log(`${color}${icon}${reset} Workspace cache does not exist`);
                    }
                    break;
                case 'cache_fresh':
                    if (check.passed) {
                        this.log(`${color}${icon}${reset} Cache is fresh (last sync: ${check.value})`);
                    }
                    else {
                        if (check.value) {
                            const warningIcon = '⚠';
                            const warningColor = '\x1b[33m'; // Yellow
                            this.log(`${warningColor}${warningIcon}${reset} Cache is outdated (last sync: ${check.value})`);
                        }
                        else {
                            this.log(`${color}${icon}${reset} ${check.message}`);
                        }
                    }
                    break;
            }
        }
        // Print recommendations for failed checks
        const failedChecks = result.checks.filter(c => !c.passed && c.recommendation);
        if (failedChecks.length > 0) {
            this.log('');
            const infoColor = '\x1b[36m'; // Cyan
            const reset = '\x1b[0m';
            for (const check of failedChecks) {
                if (check.recommendation) {
                    this.log(`${infoColor}ℹ${reset} ${check.recommendation}`);
                }
            }
        }
        // Print summary
        this.log('');
        this.log('━'.repeat(50));
        if (result.success) {
            const greenColor = '\x1b[32m';
            const reset = '\x1b[0m';
            this.log(`${greenColor}Overall: All ${result.summary.total} checks passed${reset}`);
        }
        else {
            const redColor = '\x1b[31m';
            const reset = '\x1b[0m';
            this.log(`${redColor}Overall: ${result.summary.passed}/${result.summary.total} checks passed${reset}`);
        }
        this.log('');
    }
}
exports.default = Doctor;
Doctor.description = 'Run health checks and diagnostics for Notion CLI';
Doctor.aliases = ['diagnose', 'healthcheck'];
Doctor.examples = [
    {
        description: 'Run all health checks',
        command: '$ notion-cli doctor',
    },
    {
        description: 'Run health checks with JSON output',
        command: '$ notion-cli doctor --json',
    },
];
Doctor.flags = {
    json: core_1.Flags.boolean({
        char: 'j',
        description: 'Output as JSON',
        default: false,
    }),
};
