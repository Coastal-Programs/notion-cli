# JSON Envelope Specification

## Overview

The JSON envelope standardization system ensures consistent, machine-readable output across all Notion CLI commands. This specification defines the structure, behavior, and integration patterns for envelope-based output.

**Version:** 1.0.0
**Status:** Implementation Ready
**Target API:** Notion API v5.2.1
**CLI Version:** 5.4.0+

## Motivation

### Current Problems

1. **Inconsistent JSON Output**
   - Some commands return `{success: true, data: ...}` with AutomationFlags
   - Some return raw arrays `[]` with `--output json`
   - Some return raw objects directly
   - No standardized metadata (timestamp, execution time, command name)

2. **Error Handling Inconsistency**
   - Error formats vary between commands
   - No standard error codes or suggestions
   - Mixing stderr and stdout in JSON mode

3. **Exit Code Ambiguity**
   - Only 0 (success) and 1 (error) currently used
   - No distinction between API errors and CLI validation errors
   - Automation scripts can't differentiate error types

### Goals

1. **Standardization**: All commands return consistent envelope structure
2. **Machine-Readable**: Structured metadata for automation and AI agents
3. **Debugging**: Execution time, version, and command tracking
4. **Error Handling**: Semantic error codes with actionable suggestions
5. **Backward Compatibility**: Preserve `--raw` flag for legacy users
6. **Clean Output**: Proper stdout/stderr separation

## Envelope Structure

### Success Envelope

```typescript
interface SuccessEnvelope<T> {
  success: true
  data: T
  metadata: {
    timestamp: string          // ISO 8601 format
    command: string            // e.g., "page retrieve"
    execution_time_ms: number  // Execution duration
    version: string            // CLI version (e.g., "5.4.0")
  }
}
```

**Example:**

```json
{
  "success": true,
  "data": {
    "object": "page",
    "id": "abc-123",
    "properties": { ... }
  },
  "metadata": {
    "timestamp": "2025-10-23T14:23:45.123Z",
    "command": "page retrieve",
    "execution_time_ms": 234,
    "version": "5.4.0"
  }
}
```

### Error Envelope

```typescript
interface ErrorEnvelope {
  success: false
  error: {
    code: ErrorCode | string   // Semantic error code
    message: string            // Human-readable message
    details?: any              // Additional context
    suggestions?: string[]     // Actionable suggestions
    notionError?: any          // Original Notion API error
  }
  metadata: {
    timestamp: string
    command: string
    execution_time_ms?: number // May not be available for early failures
    version: string
  }
}
```

**Example:**

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found",
    "details": {
      "resourceId": "abc-123",
      "resourceType": "database"
    },
    "suggestions": [
      "Verify the resource ID is correct",
      "Ensure your integration has access to the resource",
      "Try running: notion-cli sync"
    ]
  },
  "metadata": {
    "timestamp": "2025-10-23T14:23:45.123Z",
    "command": "db retrieve",
    "execution_time_ms": 156,
    "version": "5.4.0"
  }
}
```

## Error Codes

### Semantic Error Codes

| Code | Description | Exit Code | Example Scenarios |
|------|-------------|-----------|-------------------|
| `UNAUTHORIZED` | Authentication failure | 1 | Missing/invalid NOTION_TOKEN |
| `NOT_FOUND` | Resource not accessible | 1 | Page/DB doesn't exist or no access |
| `RATE_LIMITED` | API rate limit exceeded | 1 | Too many requests |
| `VALIDATION_ERROR` | Invalid input/arguments | 2 | Missing required arg, invalid format |
| `API_ERROR` | Generic Notion API error | 1 | Server error, network timeout |
| `CLI_ERROR` | CLI-specific error | 2 | Config file issue, parse error |
| `UNKNOWN` | Unexpected error | 1 | Unhandled exception |

### Error Code Mapping

```typescript
// CLI/Validation Errors (Exit Code 2)
const cliErrors = [
  'VALIDATION_ERROR',
  'CLI_ERROR',
  'CONFIG_ERROR',
  'INVALID_ARGUMENT'
]

// API/Notion Errors (Exit Code 1)
const apiErrors = [
  'UNAUTHORIZED',
  'NOT_FOUND',
  'RATE_LIMITED',
  'API_ERROR',
  'UNKNOWN'
]
```

## Exit Codes

| Exit Code | Meaning | When to Use | Example |
|-----------|---------|-------------|---------|
| 0 | Success | Command completed successfully | Data retrieved, page created |
| 1 | API Error | Notion API or network error | Auth failed, resource not found, rate limit |
| 2 | CLI Error | Input validation or CLI config error | Missing argument, invalid JSON filter |

**Automation Pattern:**

```bash
#!/bin/bash
notion-cli page retrieve abc-123 --json
EXIT_CODE=$?

case $EXIT_CODE in
  0)
    echo "Success - process the JSON output"
    ;;
  1)
    echo "API error - check authentication and resource access"
    ;;
  2)
    echo "CLI error - check command syntax and arguments"
    ;;
esac
```

## Output Flags

### Flag Behavior

| Flag | Envelope | Format | Use Case |
|------|----------|--------|----------|
| `--json` / `-j` | Yes | Pretty (2-space indent) | Human-readable JSON for debugging |
| `--compact-json` / `-c` | Yes | Single-line | Piping to jq, logging, storage |
| `--raw` / `-r` | No | Pretty JSON (data only) | Legacy mode, raw API response |
| `--output json` | Yes | Pretty (2-space indent) | Consistent with `--json` |
| (none) | N/A | Table | Human-readable terminal output |

### Flag Precedence

The command evaluates flags in this order:

1. `--compact-json` - Compact envelope (single line)
2. `--markdown` - Markdown table (no envelope)
3. `--pretty` - Pretty table (no envelope)
4. `--json` - Pretty envelope (2-space indent)
5. `--raw` - Raw data (no envelope)
6. Default - Table output (no envelope)

**Important:** `--raw` bypasses the envelope system entirely.

### Examples

```bash
# Pretty envelope (recommended for automation)
$ notion-cli page retrieve abc-123 --json
{
  "success": true,
  "data": { ... },
  "metadata": { ... }
}

# Compact envelope (for piping)
$ notion-cli db query abc-123 --compact-json | jq '.data.results | length'
{"success":true,"data":{"results":[...]},"metadata":{...}}
42

# Raw mode (legacy, no envelope)
$ notion-cli page retrieve abc-123 --raw
{
  "object": "page",
  "id": "abc-123",
  ...
}

# Error with suggestions
$ notion-cli db retrieve invalid-id --json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Resource not found",
    "suggestions": [
      "Verify the resource ID is correct",
      "Try running: notion-cli sync"
    ]
  },
  "metadata": { ... }
}
```

## Stdout/Stderr Separation

### Rules

1. **stdout**: JSON envelope ONLY (when --json or --compact-json is used)
2. **stderr**: All diagnostic messages (logs, warnings, retry info, cache hits)
3. **Never mix**: Human-readable text never appears on stdout in JSON mode

### Implementation

```typescript
// Good: Diagnostic to stderr
EnvelopeFormatter.writeDiagnostic('Retrying request...', 'warn')

// Good: JSON to stdout via this.log()
this.log(JSON.stringify(envelope, null, 2))

// Bad: Mixing stderr messages on stdout
console.log('Processing...') // DON'T DO THIS in JSON mode
```

### Diagnostic Helpers

```typescript
// Write to stderr (won't pollute JSON output)
EnvelopeFormatter.writeDiagnostic(message, 'info' | 'warn' | 'error')

// Log retry attempts to stderr
EnvelopeFormatter.logRetry(attempt, maxRetries, delay)

// Log cache hits to stderr (if DEBUG=true)
EnvelopeFormatter.logCacheHit(cacheKey)
```

### Example Output

```bash
$ notion-cli page retrieve abc-123 --json 2>/dev/null
{
  "success": true,
  "data": { ... },
  "metadata": { ... }
}

$ notion-cli page retrieve abc-123 --json
[INFO] Cache hit: page:abc-123
{
  "success": true,
  "data": { ... },
  "metadata": { ... }
}

# Redirect stderr to see only JSON
$ notion-cli page retrieve abc-123 --json 2>/dev/null > output.json
```

## Metadata Fields

### Standard Metadata

Every envelope includes these metadata fields:

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `timestamp` | string | ISO 8601 timestamp | `"2025-10-23T14:23:45.123Z"` |
| `command` | string | Full command name | `"page retrieve"`, `"db query"` |
| `execution_time_ms` | number | Execution duration in ms | `234` |
| `version` | string | CLI version | `"5.4.0"` |

### Additional Metadata (Optional)

Commands can add custom metadata:

```typescript
// Add pagination info
envelope.wrapSuccess(data, {
  page_size: 100,
  has_more: true,
  next_cursor: "abc123"
})

// Add query info
envelope.wrapSuccess(data, {
  total_results: 42,
  filter_applied: true
})
```

**Example with Additional Metadata:**

```json
{
  "success": true,
  "data": { ... },
  "metadata": {
    "timestamp": "2025-10-23T14:23:45.123Z",
    "command": "db query",
    "execution_time_ms": 567,
    "version": "5.4.0",
    "total_results": 42,
    "page_size": 100,
    "has_more": false
  }
}
```

## Type Safety

### TypeScript Interfaces

```typescript
import {
  SuccessEnvelope,
  ErrorEnvelope,
  Envelope,
  EnvelopeMetadata
} from './envelope'

// Type-safe success envelope
const envelope: SuccessEnvelope<PageObjectResponse> = {
  success: true,
  data: pageResponse,
  metadata: { ... }
}

// Type guards
if (isSuccessEnvelope(envelope)) {
  // envelope.data is typed as T
  console.log(envelope.data.id)
}

if (isErrorEnvelope(envelope)) {
  // envelope.error is typed as ErrorDetails
  console.log(envelope.error.code)
}
```

### Generic Type Support

```typescript
// Database query result
type QueryResult = {
  results: PageObjectResponse[]
  has_more: boolean
  next_cursor?: string
}

const envelope: SuccessEnvelope<QueryResult> =
  formatter.wrapSuccess(queryResult)

// User retrieval
const envelope: SuccessEnvelope<UserObjectResponse> =
  formatter.wrapSuccess(user)
```

## Backward Compatibility

### Preserving `--raw` Flag

The `--raw` flag is preserved for backward compatibility:

- **Bypasses envelope system completely**
- **Returns raw API response** (no metadata, no success field)
- **Useful for:** Legacy scripts, direct API response inspection

### Migration Path

Users can migrate gradually:

```bash
# Old script (still works)
PAGE=$(notion-cli page retrieve abc-123 --raw | jq '.id')

# New script (recommended)
PAGE=$(notion-cli page retrieve abc-123 --json | jq '.data.id')
```

### Detection

Scripts can detect envelope vs. raw mode:

```bash
# Check if output is enveloped
OUTPUT=$(notion-cli page retrieve abc-123 --json)
if echo "$OUTPUT" | jq -e '.success' > /dev/null 2>&1; then
  echo "Envelope format detected"
  DATA=$(echo "$OUTPUT" | jq '.data')
else
  echo "Raw format detected"
  DATA="$OUTPUT"
fi
```

## Performance Considerations

### Minimal Overhead

- **Timestamp**: Single `Date.now()` call at start
- **Execution time**: Subtraction on completion
- **Envelope creation**: Plain object construction (no serialization cost until output)
- **Memory**: Negligible (~1KB per envelope)

### Benchmarks

```
| Operation | Overhead | Notes |
|-----------|----------|-------|
| Envelope creation | <1ms | Object construction |
| JSON serialization | ~2-5ms | Depends on data size |
| Total overhead | <10ms | For typical responses |
```

### Caching Integration

Envelopes work seamlessly with the caching system:

- Cache stores raw API responses (not envelopes)
- Envelope wrapping happens after cache retrieval
- Execution time includes cache lookup time

## Security Considerations

### Sensitive Data

Never include sensitive data in error details:

```typescript
// Good: Generic error
error: {
  code: "UNAUTHORIZED",
  message: "Authentication failed",
  suggestions: ["Check your NOTION_TOKEN"]
}

// Bad: Leaking token
error: {
  code: "UNAUTHORIZED",
  message: "Token abc123xyz is invalid", // DON'T DO THIS
  details: {
    token: process.env.NOTION_TOKEN // DON'T DO THIS
  }
}
```

### Stack Traces

Stack traces are included only in non-production mode:

```typescript
if (process.env.NODE_ENV !== 'production') {
  errorDetails.details.stack = error.stack
}
```

## Testing Strategy

### Unit Tests

```typescript
describe('EnvelopeFormatter', () => {
  it('should create success envelope with metadata', () => {
    const formatter = new EnvelopeFormatter('test command', '1.0.0')
    const envelope = formatter.wrapSuccess({ test: 'data' })

    expect(envelope.success).to.be.true
    expect(envelope.data).to.deep.equal({ test: 'data' })
    expect(envelope.metadata.command).to.equal('test command')
    expect(envelope.metadata.version).to.equal('1.0.0')
  })

  it('should track execution time', () => {
    const formatter = new EnvelopeFormatter('test', '1.0.0')
    // Wait 100ms
    setTimeout(() => {
      const envelope = formatter.wrapSuccess({})
      expect(envelope.metadata.execution_time_ms).to.be.gte(100)
    }, 100)
  })

  it('should generate suggestions for errors', () => {
    const formatter = new EnvelopeFormatter('test', '1.0.0')
    const error = new NotionCLIError('NOT_FOUND', 'Resource not found')
    const envelope = formatter.wrapError(error)

    expect(envelope.error.suggestions).to.be.an('array')
    expect(envelope.error.suggestions).to.include.match(/notion-cli sync/)
  })
})
```

### Integration Tests

```typescript
describe('page retrieve with envelope', () => {
  it('should return success envelope with --json', async () => {
    const result = await runCommand(['page', 'retrieve', 'abc-123', '--json'])
    const envelope = JSON.parse(result.stdout)

    expect(envelope.success).to.be.true
    expect(envelope.data.object).to.equal('page')
    expect(envelope.metadata.command).to.equal('page retrieve')
  })

  it('should return error envelope on not found', async () => {
    const result = await runCommand(['page', 'retrieve', 'invalid-id', '--json'])
    const envelope = JSON.parse(result.stdout)

    expect(envelope.success).to.be.false
    expect(envelope.error.code).to.equal('NOT_FOUND')
    expect(result.exitCode).to.equal(1)
  })
})
```

### Manual Testing Checklist

- [ ] Success envelope has all metadata fields
- [ ] Error envelope includes suggestions
- [ ] Exit codes are correct (0, 1, 2)
- [ ] No stdout pollution in JSON mode
- [ ] Stderr contains diagnostic messages
- [ ] `--raw` bypasses envelope
- [ ] `--compact-json` outputs single line
- [ ] Execution time is accurate
- [ ] Version is populated correctly

## References

- **Source Files:**
  - `src/envelope.ts` - Core envelope system
  - `src/base-command.ts` - Base command with envelope support
  - `src/errors.ts` - Error types and codes

- **Related Documentation:**
  - `ENHANCEMENTS.md` - Caching and retry features
  - `OUTPUT_FORMATS.md` - Output format reference
  - `CLAUDE.md` - Development patterns

- **External Standards:**
  - [JSON API Specification](https://jsonapi.org/)
  - [HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)
  - [Exit Status Codes](https://tldp.org/LDP/abs/html/exitcodes.html)
