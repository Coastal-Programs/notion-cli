---
paths:
  - "src/commands/**/*.ts"
  - "src/base-command.ts"
  - "src/base-flags.ts"
---

# CLI Command Rules

## Command Structure

All commands extend `BaseCommand` (src/base-command.ts), which provides:
- Automatic JSON envelope wrapping
- Execution time measurement
- Disk cache init/shutdown
- HTTP agent cleanup
- `outputSuccess<T>()` and `outputError()` methods

## Required Pattern

```typescript
import { Args, Flags } from '@oclif/core'
import { BaseCommand } from '../../base-command'

export default class MyCommand extends BaseCommand {
  static description = 'What this command does'
  static aliases = ['shortcut']
  static examples = ['<%= config.bin %> my-command ARG --flag']

  static args = {
    id: Args.string({ description: 'Resource ID', required: true }),
  }

  static flags = {
    ...AutomationFlags,
    ...OutputFormatFlags,
    // command-specific flags
  }

  async run(): Promise<void> {
    const { args, flags } = await this.parse(MyCommand)
    // implementation
  }
}
```

## Output Rules

- Use `this.log()` for user-facing output, never `console.log`
- Structured logs go to stderr, command output to stdout
- Support output flags: `--json`, `--compact-json`, `--markdown`, `--pretty`, `--csv`, `--raw`
- Use `formatTable()` for all tabular output
- Use envelope format: `{success: boolean, data: any, metadata: any}`

## Error Handling

- Wrap Notion API errors with `wrapNotionError()` for consistent formatting
- Use `NotionCLIErrorFactory` for common error scenarios
- Exit codes: 0 = success, 1 = API error, 2 = CLI error
- Always include actionable suggestions in error messages

## Shared Flags

Import from `src/base-flags.ts`:
- `AutomationFlags`: --json, --page-size, --retry, --timeout, --no-cache, --verbose, --minimal
- `OutputFormatFlags`: --markdown, --compact-json, --pretty
