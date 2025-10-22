import { Flags } from '@oclif/core'

export const AutomationFlags = {
  json: Flags.boolean({
    char: 'j',
    description: 'Output as JSON (recommended for automation)',
    default: false,
  }),
  'page-size': Flags.integer({
    description: 'Items per page (1-100, default: 100 for automation)',
    min: 1,
    max: 100,
    default: 100,
  }),
  retry: Flags.boolean({
    description: 'Auto-retry on rate limit (respects Retry-After header)',
    default: true,
  }),
  timeout: Flags.integer({
    description: 'Request timeout in milliseconds',
    default: 30000,
  }),
  'no-cache': Flags.boolean({
    description: 'Bypass cache and force fresh API calls',
    default: false,
  }),
  verbose: Flags.boolean({
    char: 'v',
    description: 'Enable verbose logging to stderr (retry events, cache stats) - never pollutes stdout',
    default: false,
    env: 'NOTION_CLI_VERBOSE',
  }),
}

export const OutputFormatFlags = {
  markdown: Flags.boolean({
    char: 'm',
    description: 'Output as markdown table (GitHub-flavored)',
    default: false,
    exclusive: ['compact-json', 'pretty'],
  }),
  'compact-json': Flags.boolean({
    char: 'c',
    description: 'Output as compact JSON (single-line, ideal for piping)',
    default: false,
    exclusive: ['markdown', 'pretty'],
  }),
  pretty: Flags.boolean({
    char: 'P',
    description: 'Output as pretty table with borders',
    default: false,
    exclusive: ['markdown', 'compact-json'],
  }),
}
