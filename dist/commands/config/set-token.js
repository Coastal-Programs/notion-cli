"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
const core_1 = require("@oclif/core");
const fs = require("fs/promises");
const path = require("path");
const os = require("os");
const readline = require("readline");
const base_flags_1 = require("../../base-flags");
const errors_1 = require("../../errors");
class ConfigSetToken extends core_1.Command {
    async run() {
        const { args, flags } = await this.parse(ConfigSetToken);
        try {
            // Get token from args or prompt
            let token = args.token;
            if (!token) {
                if (flags.json) {
                    throw new errors_1.NotionCLIError(errors_1.NotionCLIErrorCode.TOKEN_MISSING, 'Token required in JSON mode', [
                        {
                            description: 'Provide the token as an argument',
                            command: 'notion-cli config set-token secret_your_token_here --json'
                        }
                    ]);
                }
                // Interactive prompt
                const rl = readline.createInterface({
                    input: process.stdin,
                    output: process.stdout,
                });
                token = await new Promise((resolve) => {
                    rl.question('Enter your Notion integration token: ', (answer) => {
                        rl.close();
                        resolve(answer.trim());
                    });
                });
            }
            // Validate token format
            if (!token || !token.startsWith('secret_')) {
                throw new errors_1.NotionCLIError(errors_1.NotionCLIErrorCode.TOKEN_INVALID, 'Invalid token format - Notion tokens must start with "secret_"', [
                    {
                        description: 'Get your integration token from Notion',
                        link: 'https://developers.notion.com/docs/create-a-notion-integration'
                    },
                    {
                        description: 'Tokens should look like: secret_abc123...',
                    }
                ], {
                    userInput: token,
                    metadata: { tokenFormat: 'invalid' }
                });
            }
            // Detect shell and rc file
            const shell = this.detectShell();
            const rcFile = this.getRcFilePath(shell);
            // Read existing rc file
            let rcContent = '';
            try {
                rcContent = await fs.readFile(rcFile, 'utf-8');
            }
            catch (error) {
                if (error && typeof error === 'object' && 'code' in error && error.code !== 'ENOENT') {
                    throw error;
                }
                // File doesn't exist, will create it
            }
            // Check if NOTION_TOKEN already exists
            const tokenLineRegex = /^export\s+NOTION_TOKEN=.*/gm;
            const newTokenLine = `export NOTION_TOKEN="${token}"`;
            let updatedContent;
            if (tokenLineRegex.test(rcContent)) {
                // Replace existing token
                updatedContent = rcContent.replace(tokenLineRegex, newTokenLine);
            }
            else {
                // Add new token
                updatedContent = rcContent.trim() + '\n\n# Notion CLI Token\n' + newTokenLine + '\n';
            }
            // Write updated rc file
            await fs.writeFile(rcFile, updatedContent, 'utf-8');
            if (flags.json) {
                this.log(JSON.stringify({
                    success: true,
                    message: 'Token saved successfully',
                    rcFile,
                    shell,
                    nextSteps: [
                        `Reload your shell: source ${rcFile}`,
                        'Run: notion-cli sync',
                    ],
                }, null, 2));
            }
            else {
                this.log(`\n✓ Token saved to ${rcFile}`);
                this.log('\nNext steps:');
                this.log(`  1. Reload your shell: source ${rcFile}`);
                this.log(`  2. Or restart your terminal`);
                this.log(`  3. Run: notion-cli sync`);
                this.log('\nWould you like to sync your workspace now? (y/n)');
                const rl = readline.createInterface({
                    input: process.stdin,
                    output: process.stdout,
                });
                const answer = await new Promise((resolve) => {
                    rl.question('> ', (answer) => {
                        rl.close();
                        resolve(answer.trim().toLowerCase());
                    });
                });
                if (answer === 'y' || answer === 'yes') {
                    // Set token in current process
                    process.env.NOTION_TOKEN = token;
                    // Run sync command - dynamic import to avoid circular dependencies
                    this.log('\nRunning sync...\n');
                    const { default: Sync } = await Promise.resolve().then(() => require('../sync.js'));
                    await Sync.run([]);
                }
                else {
                    this.log('\nSkipping sync. You can run it manually with: notion-cli sync');
                }
            }
            process.exit(0);
        }
        catch (error) {
            const cliError = error instanceof errors_1.NotionCLIError
                ? error
                : (0, errors_1.wrapNotionError)(error instanceof Error ? error : new Error(String(error)), {
                    endpoint: 'config.set-token'
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
    /**
     * Detect the current shell
     */
    detectShell() {
        const shell = process.env.SHELL || '';
        if (shell.includes('zsh'))
            return 'zsh';
        if (shell.includes('bash'))
            return 'bash';
        if (shell.includes('fish'))
            return 'fish';
        // Default to bash on Unix, powershell on Windows
        return process.platform === 'win32' ? 'powershell' : 'bash';
    }
    /**
     * Get the rc file path for the detected shell
     */
    getRcFilePath(shell) {
        const home = os.homedir();
        switch (shell) {
            case 'zsh':
                return path.join(home, '.zshrc');
            case 'bash':
                return path.join(home, '.bashrc');
            case 'fish':
                return path.join(home, '.config', 'fish', 'config.fish');
            case 'powershell':
                return path.join(home, 'Documents', 'PowerShell', 'Microsoft.PowerShell_profile.ps1');
            default:
                return path.join(home, '.bashrc');
        }
    }
}
exports.default = ConfigSetToken;
ConfigSetToken.description = 'Set NOTION_TOKEN in your shell configuration file';
ConfigSetToken.aliases = ['config:token'];
ConfigSetToken.examples = [
    {
        description: 'Set Notion token interactively',
        command: 'notion-cli config set-token',
    },
    {
        description: 'Set Notion token directly',
        command: 'notion-cli config set-token secret_abc123...',
    },
    {
        description: 'Set token with JSON output',
        command: 'notion-cli config set-token secret_abc123... --json',
    },
];
ConfigSetToken.args = {
    token: core_1.Args.string({
        description: 'Notion integration token (starts with secret_)',
        required: false,
    }),
};
ConfigSetToken.flags = {
    ...base_flags_1.AutomationFlags,
};
