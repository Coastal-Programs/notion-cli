#!/usr/bin/env node

/**
 * Post-install script for @coastal-programs/notion-cli
 * Shows welcome message and next steps after installation
 */

// Respect npm's --silent flag
const isSilent = process.env.npm_config_loglevel === 'silent';
if (isSilent) {
  process.exit(0);
}

// ANSI color codes for cross-platform compatibility
const colors = {
  reset: '\x1b[0m',
  bright: '\x1b[1m',
  dim: '\x1b[2m',
  green: '\x1b[32m',
  blue: '\x1b[34m',
  cyan: '\x1b[36m',
  gray: '\x1b[90m',
};

// Graceful error handling - don't break installation
try {
  const packageJson = require('../package.json');
  const version = packageJson.version;

  // Welcome message with clear next steps
  console.log(`
${colors.green}✓${colors.reset} Notion CLI ${colors.bright}v${version}${colors.reset} installed successfully!

${colors.blue}Next steps:${colors.reset}
  ${colors.gray}1.${colors.reset} Set your token: ${colors.cyan}notion-cli config set-token${colors.reset}
  ${colors.gray}2.${colors.reset} Test connection: ${colors.cyan}notion-cli whoami${colors.reset}
  ${colors.gray}3.${colors.reset} Sync workspace: ${colors.cyan}notion-cli sync${colors.reset}

${colors.blue}Resources:${colors.reset}
  Documentation: https://github.com/Coastal-Programs/notion-cli
  Report issues: https://github.com/Coastal-Programs/notion-cli/issues
  Get started:   ${colors.cyan}notion-cli --help${colors.reset}
`);
} catch (error) {
  // Fallback to simple message if anything goes wrong
  console.log(`
Notion CLI installed successfully!

Next steps:
  1. Set your token: notion-cli config set-token
  2. Test connection: notion-cli whoami
  3. Sync workspace: notion-cli sync

Get help: notion-cli --help
`);
}
