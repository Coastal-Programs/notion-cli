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
  yellow: '\x1b[33m',
  magenta: '\x1b[35m',
};

// ASCII Art Banner
const banner = `
${colors.cyan}███╗   ██╗ ██████╗ ████████╗██╗ ██████╗ ███╗   ██╗     ██████╗██╗     ██╗
████╗  ██║██╔═══██╗╚══██╔══╝██║██╔═══██╗████╗  ██║    ██╔════╝██║     ██║
██╔██╗ ██║██║   ██║   ██║   ██║██║   ██║██╔██╗ ██║    ██║     ██║     ██║
██║╚██╗██║██║   ██║   ██║   ██║██║   ██║██║╚██╗██║    ██║     ██║     ██║
██║ ╚████║╚██████╔╝   ██║   ██║╚██████╔╝██║ ╚████║    ╚██████╗███████╗██║
╚═╝  ╚═══╝ ╚═════╝    ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝     ╚═════╝╚══════╝╚═╝${colors.reset}
`;

// Graceful error handling - don't break installation
try {
  const packageJson = require('../package.json');
  const version = packageJson.version;

  // Welcome message with banner and clear next steps
  console.log(banner);
  console.log(`${colors.green}✓${colors.reset} Version ${colors.bright}${version}${colors.reset} installed successfully!\n`);
  console.log(`${colors.blue}Quick Start:${colors.reset}`);
  console.log(`  ${colors.cyan}notion-cli init${colors.reset}          ${colors.dim}# Interactive setup wizard${colors.reset}`);
  console.log(`  ${colors.cyan}notion-cli --help${colors.reset}        ${colors.dim}# View all commands${colors.reset}\n`);
  console.log(`${colors.blue}Resources:${colors.reset}`);
  console.log(`  ${colors.gray}•${colors.reset} Documentation: ${colors.dim}https://github.com/Coastal-Programs/notion-cli${colors.reset}`);
  console.log(`  ${colors.gray}•${colors.reset} Get API Token:  ${colors.dim}https://developers.notion.com/docs/create-a-notion-integration${colors.reset}`);
  console.log(`  ${colors.gray}•${colors.reset} Report Issues:  ${colors.dim}https://github.com/Coastal-Programs/notion-cli/issues${colors.reset}`);
  console.log('');
} catch (error) {
  // Fallback to simple message if anything goes wrong
  console.log(`
Notion CLI installed successfully!

Quick Start:
  notion-cli init          # Interactive setup wizard
  notion-cli --help        # View all commands

Get help: https://github.com/Coastal-Programs/notion-cli
`);
}
