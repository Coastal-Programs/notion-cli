# JSON Envelope Quick Reference Card

## TL;DR

**What:** Standardized JSON output format for all CLI commands
**Why:** Consistent machine-readable output for automation and AI agents
**When:** Use `--json` or `--compact-json` flag
**Exit Codes:** 0 = success, 1 = API error, 2 = CLI error

## Envelope Structures

### Success Envelope

```typescript
{
  success: true,
  data: T,                    // Actual API response
  metadata: {
    timestamp: string,        // ISO 8601
    command: string,          // e.g., "page retrieve"
    execution_time_ms: number,
    version: string           // CLI version
  }
}
```

### Error Envelope

```typescript
{
  success: false,
  error: {
    code: ErrorCode,          // Semantic error code
    message: string,          // Human-readable message
    details?: any,            // Additional context
    suggestions?: string[],   // Actionable suggestions
    notionError?: any         // Original Notion API error
  },
  metadata: {
    timestamp: string,
    command: string,
    execution_time_ms?: number,
    version: string
  }
}
```

## Command Flags

| Flag | Envelope? | Format | Use Case |
|------|-----------|--------|----------|
| `--json` | ✓ | Pretty (indented) | Debugging, human-readable |
| `--compact-json` | ✓ | Single-line | Piping, logging |
| `--raw` | ✗ | Raw API response | Legacy scripts |
| (none) | ✗ | Table | Terminal output |

## Exit Codes

| Code | Type | When | Example |
|------|------|------|---------|
| 0 | Success | Command succeeded | Page retrieved successfully |
| 1 | API Error | Notion API issue | Auth failed, not found, rate limit |
| 2 | CLI Error | Input validation | Missing arg, invalid JSON |

## Error Codes

### API Errors (Exit 1)
- `UNAUTHORIZED` - Authentication failed
- `NOT_FOUND` - Resource not accessible
- `RATE_LIMITED` - API rate limit exceeded
- `API_ERROR` - Generic Notion API error
- `UNKNOWN` - Unexpected error

### CLI Errors (Exit 2)
- `VALIDATION_ERROR` - Invalid input/arguments
- `CLI_ERROR` - CLI-specific error
- `CONFIG_ERROR` - Configuration error
- `INVALID_ARGUMENT` - Invalid argument value

## Usage Examples

### Basic Success

```bash
$ notion-cli page retrieve abc-123 --json
{
  "success": true,
  "data": { "object": "page", "id": "abc-123", ... },
  "metadata": { "timestamp": "...", "command": "page retrieve", ... }
}
$ echo $?
0
```

### Compact Output

```bash
$ notion-cli db query db-123 --compact-json | jq '.data | length'
{"success":true,"data":[...],"metadata":{...}}
42
```

### Error Handling

```bash
$ notion-cli page retrieve invalid-id --json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found",
    "suggestions": ["Verify the resource ID", "Try: notion-cli sync"]
  },
  "metadata": { ... }
}
$ echo $?
1
```

### Automation Script

```bash
OUTPUT=$(notion-cli page retrieve abc-123 --json)
if [ $? -eq 0 ]; then
  PAGE_ID=$(echo "$OUTPUT" | jq -r '.data.id')
  echo "Success: $PAGE_ID"
else
  ERROR=$(echo "$OUTPUT" | jq -r '.error.code')
  echo "Error: $ERROR"
fi
```

## Implementation Pattern

### For New Commands

```typescript
import { BaseCommand } from '../../base-command'
import { Args, ux } from '@oclif/core'
import * as notion from '../../notion'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'

export default class MyCommand extends BaseCommand {
  static flags = {
    ...ux.table.flags(),
    ...AutomationFlags,
    ...OutputFormatFlags,
  }

  async run(): Promise<void> {
    const { args, flags } = await this.parse(MyCommand)

    try {
      const result = await notion.someOperation(args)

      // Envelope output (auto-exits)
      if (flags.json || flags['compact-json']) {
        this.outputSuccess(result, flags, {
          /* optional metadata */
        })
      }

      // Raw output
      if (flags.raw) {
        this.log(JSON.stringify(result, null, 2))
        process.exit(0)
      }

      // Table output
      ux.table([result], columns, flags)
      process.exit(0)

    } catch (error) {
      this.outputError(error, flags, {
        /* optional context */
      })
    }
  }
}
```

### Migration Checklist

- [ ] Change `extends Command` → `extends BaseCommand`
- [ ] Replace manual JSON → `this.outputSuccess()`
- [ ] Replace error handling → `this.outputError()`
- [ ] Add `flags.json || flags['compact-json']` check
- [ ] Test all output modes
- [ ] Verify exit codes

## Key Methods

### EnvelopeFormatter

```typescript
// Create formatter (auto-done by BaseCommand)
const formatter = new EnvelopeFormatter(commandName, version)

// Wrap success
const envelope = formatter.wrapSuccess(data, { metadata })

// Wrap error
const envelope = formatter.wrapError(error, { context })

// Output envelope
formatter.outputEnvelope(envelope, flags, logFn)

// Get exit code
const exitCode = formatter.getExitCode(envelope) // 0, 1, or 2
```

### BaseCommand

```typescript
// Output success and exit
this.outputSuccess(data, flags, { additionalMetadata })

// Output error and exit
this.outputError(error, flags, { additionalContext })
```

### Diagnostics

```typescript
// Write to stderr (won't pollute JSON)
EnvelopeFormatter.writeDiagnostic('Processing...', 'info')
EnvelopeFormatter.logRetry(attempt, maxRetries, delay)
EnvelopeFormatter.logCacheHit(cacheKey)
```

## Testing

### Unit Test

```typescript
import { EnvelopeFormatter } from '../src/envelope'
import { expect } from 'chai'

it('should create success envelope', () => {
  const formatter = new EnvelopeFormatter('test', '1.0.0')
  const envelope = formatter.wrapSuccess({ id: 'abc' })

  expect(envelope.success).to.be.true
  expect(envelope.data).to.deep.equal({ id: 'abc' })
  expect(envelope.metadata.command).to.equal('test')
})
```

### Integration Test

```typescript
import { expect, test } from '@oclif/test'

test
  .stdout()
  .command(['page', 'retrieve', 'abc-123', '--json'])
  .it('returns success envelope', (ctx) => {
    const output = JSON.parse(ctx.stdout)
    expect(output.success).to.be.true
    expect(output.metadata.command).to.equal('page retrieve')
  })
```

## Common Patterns

### With Pagination Metadata

```typescript
this.outputSuccess(results, flags, {
  total_results: results.length,
  page_size: flags.pageSize,
  has_more: hasMore,
})
```

### With Operation Info

```typescript
this.outputSuccess(createdPage, flags, {
  operation: 'create',
  resource_type: 'page',
})
```

### With Custom Error Context

```typescript
this.outputError(error, flags, {
  database_id: args.database_id,
  filter_used: !!flags.filter,
})
```

## Type Safety

```typescript
import {
  SuccessEnvelope,
  ErrorEnvelope,
  isSuccessEnvelope
} from './envelope'

// Type-safe success envelope
const envelope: SuccessEnvelope<PageObjectResponse> =
  formatter.wrapSuccess(page)

// Type guard
if (isSuccessEnvelope(envelope)) {
  console.log(envelope.data.id) // Typed as PageObjectResponse
}
```

## Stdout/Stderr Separation

```typescript
// Good: JSON to stdout, diagnostics to stderr
if (flags.json) {
  EnvelopeFormatter.writeDiagnostic('Cache hit', 'info') // stderr
  this.outputSuccess(data, flags) // stdout
}

// Bad: Mixing output
console.log('Processing...') // Goes to stdout - DON'T DO THIS
```

## Performance

| Operation | Overhead | Impact |
|-----------|----------|--------|
| Envelope creation | <1ms | Negligible |
| JSON serialization | 2-5ms | Minimal |
| **Total** | **<10ms** | **<5% typical** |

## Troubleshooting

**Q: Command hangs after outputSuccess()**
A: Remove code after `outputSuccess()` - it auto-exits

**Q: Envelope missing execution_time_ms**
A: Ensure BaseCommand.init() is called (automatic)

**Q: JSON output polluted with logs**
A: Use `EnvelopeFormatter.writeDiagnostic()` for logs

**Q: Suggestions not appearing**
A: Use standard ErrorCode enum values

## Resources

- **Full Spec:** `docs/ENVELOPE_SPECIFICATION.md`
- **Integration Guide:** `docs/ENVELOPE_INTEGRATION_GUIDE.md`
- **Testing Strategy:** `docs/ENVELOPE_TESTING_STRATEGY.md`
- **Summary:** `docs/ENVELOPE_SYSTEM_SUMMARY.md`
- **Source:** `src/envelope.ts`, `src/base-command.ts`
- **Tests:** `test/envelope.test.ts`

## Version

- **Envelope System Version:** 1.0.0
- **CLI Version:** 5.4.0+
- **Last Updated:** 2025-10-23
