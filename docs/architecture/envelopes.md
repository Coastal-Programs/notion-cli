# Envelope System Architecture

## System Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         User / Script                            │
└─────────────────────────┬───────────────────────────────────────┘
                          │
                          ▼
                    ┌──────────┐
                    │   CLI    │  notion-cli [command] --json
                    └─────┬────┘
                          │
                          ▼
        ┌─────────────────────────────────────┐
        │         BaseCommand (oclif)          │
        │  - Parses flags                      │
        │  - Creates EnvelopeFormatter         │
        │  - Routes to command logic           │
        └─────────────┬───────────────────────┘
                      │
        ┌─────────────┴─────────────┐
        │                           │
        ▼                           ▼
  ┌──────────┐              ┌──────────────┐
  │  Success │              │    Error     │
  │   Path   │              │     Path     │
  └────┬─────┘              └──────┬───────┘
       │                           │
       ▼                           ▼
  ┌──────────────────┐      ┌────────────────────┐
  │ outputSuccess()  │      │  outputError()     │
  │ - Wrap data      │      │  - Wrap error      │
  │ - Add metadata   │      │  - Add suggestions │
  │ - Exit code 0    │      │  - Exit code 1/2   │
  └────┬─────────────┘      └──────┬─────────────┘
       │                           │
       └───────────┬───────────────┘
                   │
                   ▼
         ┌──────────────────┐
         │ EnvelopeFormatter │
         │ - Format output   │
         │ - Track timing    │
         │ - Version info    │
         └─────────┬─────────┘
                   │
         ┌─────────┴─────────┐
         │                   │
         ▼                   ▼
    ┌────────┐         ┌─────────┐
    │ stdout │         │ stderr  │
    │  JSON  │         │  Logs   │
    └────────┘         └─────────┘
```

## Component Architecture

### Layer 1: User Interface

**Components:**
- CLI invocation (`notion-cli [command] [args] [flags]`)
- Flag parsing (`--json`, `--compact-json`, `--raw`)
- Command routing (oclif framework)

**Responsibilities:**
- Parse command-line arguments
- Validate required arguments
- Route to appropriate command handler

### Layer 2: Command Layer (BaseCommand)

**File:** `src/base-command.ts`

**Responsibilities:**
- Extend oclif `Command` class
- Initialize `EnvelopeFormatter` with command name and version
- Provide `outputSuccess()` and `outputError()` convenience methods
- Determine envelope usage based on flags
- Handle process exit with appropriate code

**Key Methods:**

```typescript
class BaseCommand extends Command {
  envelope: EnvelopeFormatter

  init(): void
    // Initialize envelope formatter
    // Extract command name from this.id
    // Get version from this.config

  outputSuccess<T>(data: T, flags: any, metadata?: object): never
    // Wrap data in success envelope
    // Output envelope to stdout
    // Exit with code 0

  outputError(error: any, flags: any, context?: object): never
    // Wrap error in error envelope
    // Add suggestions based on error code
    // Output envelope to stdout (or error message to stderr)
    // Exit with code 1 or 2
}
```

### Layer 3: Envelope System

**File:** `src/envelope.ts`

**Components:**

#### EnvelopeFormatter Class

```typescript
class EnvelopeFormatter {
  private startTime: number
  private commandName: string
  private version: string

  constructor(commandName: string, version: string)
    // Record start time for execution tracking
    // Store command name and version

  wrapSuccess<T>(data: T, metadata?: object): SuccessEnvelope<T>
    // Create success envelope
    // Add standard metadata (timestamp, command, execution_time_ms, version)
    // Merge additional metadata

  wrapError(error: any, context?: object): ErrorEnvelope
    // Wrap NotionCLIError, Error, or raw error object
    // Generate suggestions based on error code
    // Add standard metadata

  outputEnvelope(envelope: Envelope, flags: OutputFlags, logFn?: Function): void
    // Handle --raw flag (bypass envelope)
    // Handle --compact-json flag (single-line)
    // Default: pretty JSON (2-space indent)

  getExitCode(envelope: Envelope): ExitCode
    // Return 0 for success
    // Return 2 for CLI errors (VALIDATION_ERROR)
    // Return 1 for API errors

  static writeDiagnostic(message: string, level: string): void
    // Write diagnostic messages to stderr
    // Prevents pollution of JSON output on stdout
}
```

#### Type Definitions

```typescript
interface SuccessEnvelope<T> {
  success: true
  data: T
  metadata: EnvelopeMetadata
}

interface ErrorEnvelope {
  success: false
  error: ErrorDetails
  metadata: EnvelopeMetadata
}

interface EnvelopeMetadata {
  timestamp: string
  command: string
  execution_time_ms: number
  version: string
}

interface ErrorDetails {
  code: ErrorCode | string
  message: string
  details?: any
  suggestions?: string[]
  notionError?: any
}

enum ExitCode {
  SUCCESS = 0
  API_ERROR = 1
  CLI_ERROR = 2
}
```

### Layer 4: Error System

**File:** `src/errors.ts` (existing, enhanced)

**Components:**

```typescript
enum ErrorCode {
  RATE_LIMITED = 'RATE_LIMITED'
  NOT_FOUND = 'NOT_FOUND'
  UNAUTHORIZED = 'UNAUTHORIZED'
  VALIDATION_ERROR = 'VALIDATION_ERROR'
  API_ERROR = 'API_ERROR'
  UNKNOWN = 'UNKNOWN'
}

class NotionCLIError extends Error {
  code: ErrorCode
  details?: any
  notionError?: any

  toJSON(): object
    // Legacy method - now superseded by envelope
}

function wrapNotionError(error: any): NotionCLIError
  // Map HTTP status codes to ErrorCode
  // 401/403 → UNAUTHORIZED
  // 404 → NOT_FOUND
  // 429 → RATE_LIMITED
  // 400 → VALIDATION_ERROR
  // Other → API_ERROR
```

## Data Flow

### Success Path

```
1. User runs: notion-cli page retrieve abc-123 --json

2. BaseCommand.run()
   ├─ Parse arguments: { page_id: 'abc-123' }
   ├─ Parse flags: { json: true }
   └─ Initialize EnvelopeFormatter('page retrieve', '5.4.0')

3. Command logic
   ├─ Call Notion API: retrievePage({ page_id: 'abc-123' })
   └─ Receive: PageObjectResponse

4. Output handling
   ├─ Check flags.json → true
   └─ Call this.outputSuccess(pageData, flags)

5. EnvelopeFormatter.wrapSuccess()
   ├─ Create envelope structure
   ├─ Add data: pageData
   ├─ Calculate execution_time_ms: Date.now() - startTime
   └─ Add metadata: { timestamp, command, execution_time_ms, version }

6. EnvelopeFormatter.outputEnvelope()
   ├─ Check flags.raw → false
   ├─ Check flags['compact-json'] → false
   └─ Output pretty JSON to stdout

7. Process exit
   └─ getExitCode(envelope) → 0
```

### Error Path

```
1. User runs: notion-cli page retrieve invalid-id --json

2. BaseCommand.run()
   ├─ Parse arguments: { page_id: 'invalid-id' }
   ├─ Parse flags: { json: true }
   └─ Initialize EnvelopeFormatter('page retrieve', '5.4.0')

3. Command logic
   ├─ Call Notion API: retrievePage({ page_id: 'invalid-id' })
   └─ Notion API throws: { status: 404, message: 'Not found' }

4. Error catch block
   └─ Call this.outputError(error, flags)

5. wrapNotionError()
   ├─ Check error.status → 404
   └─ Return NotionCLIError(NOT_FOUND, 'Resource not found')

6. EnvelopeFormatter.wrapError()
   ├─ Extract error code: NOT_FOUND
   ├─ Generate suggestions based on error code
   └─ Create error envelope with metadata

7. EnvelopeFormatter.outputEnvelope()
   ├─ Check flags.json → true
   └─ Output error envelope to stdout

8. Process exit
   └─ getExitCode(envelope) → 1 (API_ERROR)
```

## Integration Points

### With Notion API Wrapper

```typescript
// src/notion.ts
export const retrievePage = async (params: GetPageParameters) => {
  return cachedFetch(
    'page',
    params.page_id,
    () => client.pages.retrieve(params)
  )
}

// If API call fails, error propagates to command
// Command catch block wraps it in envelope
```

### With Cache System

```typescript
// Cache operates transparently
// Envelope metadata includes cache time
// Cache diagnostics go to stderr

cachedFetch('page', pageId, fetcher)
  ├─ Check cache
  ├─ If hit: EnvelopeFormatter.logCacheHit(key) → stderr
  └─ Return data (timing includes cache lookup)
```

### With Retry System

```typescript
// Retry operates before envelope wrapping
// Retry diagnostics go to stderr

fetchWithRetry(fetcher)
  ├─ Attempt 1: fails
  ├─ EnvelopeFormatter.logRetry(1, 3, 1000) → stderr
  ├─ Attempt 2: succeeds
  └─ Return data (timing includes all attempts)
```

## Output Routing

### Stdout (JSON Output Only)

```
┌────────────────────────────────┐
│   Envelope JSON (stdout)       │
│                                │
│  --json:                       │
│    Pretty formatted (indented) │
│                                │
│  --compact-json:               │
│    Single-line (no whitespace) │
│                                │
│  --raw:                        │
│    Raw data only (no envelope) │
└────────────────────────────────┘
```

### Stderr (Diagnostics)

```
┌────────────────────────────────┐
│   Diagnostics (stderr)         │
│                                │
│  - Cache hit/miss              │
│  - Retry attempts              │
│  - Debug messages              │
│  - Warnings                    │
│  - Non-JSON error messages     │
└────────────────────────────────┘
```

**Example:**

```bash
$ notion-cli page retrieve abc-123 --json 2>err.log >out.json

# out.json (stdout):
{"success":true,"data":{...},"metadata":{...}}

# err.log (stderr):
[INFO] Cache hit: page:abc-123
```

## State Management

### EnvelopeFormatter State

```
┌─────────────────────────┐
│   EnvelopeFormatter     │
├─────────────────────────┤
│ - startTime: number     │  Set in constructor
│ - commandName: string   │  Set in constructor
│ - version: string       │  Set in constructor
└─────────────────────────┘

Lifecycle:
1. Created in BaseCommand.init()
2. Used throughout command execution
3. Destroyed after process.exit()
```

### No Shared State

- Each command invocation creates new EnvelopeFormatter
- No global state or singletons
- Thread-safe (each process is isolated)
- No race conditions

## Error Handling Flow

```
┌────────────────────────────────────────────────────┐
│                   Error Source                      │
└──────────────┬─────────────────────────────────────┘
               │
     ┌─────────┴──────────┐
     │                    │
     ▼                    ▼
┌─────────┐         ┌──────────┐
│ Notion  │         │ CLI/User │
│   API   │         │  Input   │
└────┬────┘         └────┬─────┘
     │                   │
     ▼                   ▼
┌─────────────┐    ┌────────────────┐
│ HTTP Error  │    │ Validation     │
│ - 401, 403  │    │ - Missing arg  │
│ - 404       │    │ - Invalid JSON │
│ - 429       │    │ - Parse error  │
│ - 500-504   │    └────────┬───────┘
└─────┬───────┘             │
      │                     │
      ▼                     ▼
┌────────────────┐   ┌──────────────────┐
│ wrapNotionError│   │ NotionCLIError   │
│ - Map status   │   │ - VALIDATION_    │
│ - Create       │   │   ERROR          │
│   NotionCLIError│  └────────┬─────────┘
└────────┬───────┘            │
         │                    │
         └──────────┬─────────┘
                    │
                    ▼
          ┌──────────────────┐
          │ EnvelopeFormatter│
          │   .wrapError()   │
          │ - Add suggestions│
          │ - Add metadata   │
          └────────┬─────────┘
                   │
                   ▼
             ┌──────────┐
             │ Error    │
             │ Envelope │
             └──────────┘
                   │
         ┌─────────┴──────────┐
         │                    │
         ▼                    ▼
    ┌────────┐          ┌─────────┐
    │Exit 1  │          │ Exit 2  │
    │API Err │          │ CLI Err │
    └────────┘          └─────────┘
```

## Metadata Flow

```
┌────────────────────────────────────┐
│     Command Invocation             │
│  notion-cli page retrieve abc-123  │
└─────────────┬──────────────────────┘
              │
              ▼
    ┌──────────────────┐
    │ BaseCommand.init()│
    │                   │
    │ Extract:          │
    │ - commandName     │ From this.id ('page:retrieve')
    │   → "page retrieve"│ (replace ':' with ' ')
    │                   │
    │ - version         │ From this.config.version
    │   → "5.4.0"       │
    │                   │
    │ - startTime       │ From Date.now()
    │   → 1698234567890 │
    └─────────┬─────────┘
              │
              ▼
    ┌──────────────────────┐
    │ EnvelopeFormatter    │
    │   constructor()      │
    │                      │
    │ Store:               │
    │ - this.commandName   │
    │ - this.version       │
    │ - this.startTime     │
    └──────────┬───────────┘
               │
               ▼
    ┌─────────────────────┐
    │ Command execution    │
    │ (API calls, logic)   │
    └──────────┬──────────┘
               │
               ▼
    ┌──────────────────────┐
    │ wrapSuccess() or     │
    │ wrapError()          │
    │                      │
    │ Calculate:           │
    │ - execution_time_ms  │ Date.now() - this.startTime
    │   → 234              │
    │                      │
    │ - timestamp          │ new Date().toISOString()
    │   → "2025-10-23T..." │
    │                      │
    │ Create metadata:     │
    │ {                    │
    │   timestamp,         │
    │   command,           │
    │   execution_time_ms, │
    │   version,           │
    │   ...additional      │ Optional metadata from caller
    │ }                    │
    └──────────┬───────────┘
               │
               ▼
         ┌──────────┐
         │ Envelope │
         └──────────┘
```

## Type Flow

```
┌─────────────────────────────────────────┐
│         Generic Type Flow               │
└─────────────────┬───────────────────────┘
                  │
                  ▼
    ┌───────────────────────────┐
    │ Command returns data      │
    │ Type: PageObjectResponse  │
    └─────────────┬─────────────┘
                  │
                  ▼
    ┌──────────────────────────────────────┐
    │ wrapSuccess<PageObjectResponse>(data)│
    │                                      │
    │ Returns:                             │
    │ SuccessEnvelope<PageObjectResponse>  │
    └─────────────┬────────────────────────┘
                  │
                  ▼
    ┌─────────────────────────────────────┐
    │ Type-safe access to envelope.data   │
    │                                     │
    │ envelope.data.id         // string  │
    │ envelope.data.properties // Props   │
    └─────────────────────────────────────┘

┌─────────────────────────────────────────┐
│        Type Guards                      │
└─────────────┬───────────────────────────┘
              │
              ▼
    if (isSuccessEnvelope(envelope)) {
      // TypeScript knows: envelope is SuccessEnvelope<T>
      console.log(envelope.data)
    } else {
      // TypeScript knows: envelope is ErrorEnvelope
      console.log(envelope.error.code)
    }
```

## Performance Characteristics

### Time Complexity

| Operation | Complexity | Typical Time |
|-----------|------------|--------------|
| EnvelopeFormatter creation | O(1) | <0.1ms |
| wrapSuccess() | O(1) | <1ms |
| wrapError() | O(1) | <1ms |
| JSON.stringify() | O(n) | 2-5ms |
| Process exit | O(1) | <1ms |
| **Total overhead** | **O(n)** | **<10ms** |

*n = size of data being serialized*

### Space Complexity

| Component | Size | Notes |
|-----------|------|-------|
| EnvelopeFormatter instance | ~100 bytes | Per command invocation |
| Envelope object | ~1KB + data size | Metadata is small |
| Suggestions array | ~200 bytes | 3-5 suggestions |
| **Total overhead** | **~1.5KB** | Negligible |

### Scalability

- **Single command:** <10ms overhead
- **Concurrent commands:** Independent processes, no interference
- **Large responses (1000+ items):** JSON serialization dominates (10-50ms)
- **Small responses (<100 items):** Overhead is <5% of total time

## Extension Points

### Adding New Error Codes

```typescript
// 1. Add to ErrorCode enum (src/errors.ts)
export enum ErrorCode {
  // ... existing codes ...
  NETWORK_ERROR = 'NETWORK_ERROR',
}

// 2. Add suggestion generation (src/envelope.ts)
function generateSuggestions(errorCode: ErrorCode) {
  switch (errorCode) {
    // ... existing cases ...
    case ErrorCode.NETWORK_ERROR:
      return ['Check your internet connection', 'Retry the request']
  }
}

// 3. Add exit code mapping if needed (src/envelope.ts)
// Network errors are API errors → exit code 1
```

### Adding Custom Metadata

```typescript
// In command
this.outputSuccess(results, flags, {
  // Standard pagination
  page_size: flags.pageSize,
  has_more: hasMore,

  // Custom metadata
  filter_applied: !!flags.filter,
  sort_applied: !!flags.sort,
  cache_hit: wasCacheHit,
})
```

### Adding New Output Formats

```typescript
// In EnvelopeFormatter.outputEnvelope()
if (flags.yaml) {
  // Convert envelope to YAML
  logFn(yaml.stringify(envelope))
  return
}

if (flags.xml) {
  // Convert envelope to XML
  logFn(xmlBuilder.build(envelope))
  return
}
```

## Security Considerations

### Sensitive Data

```typescript
// DON'T: Leak tokens in error details
error.details = {
  token: process.env.NOTION_TOKEN // ❌
}

// DO: Generic error messages
error.details = {
  tokenSet: !!process.env.NOTION_TOKEN // ✓
}
```

### Stack Traces

```typescript
// Production: No stack traces
if (process.env.NODE_ENV !== 'production') {
  errorDetails.details.stack = error.stack
}
```

### Input Sanitization

- User input in error messages is not sanitized
- JSON.stringify handles special characters safely
- No XSS risk (CLI output, not web)

## Monitoring and Observability

### Metrics Available

```json
{
  "metadata": {
    "timestamp": "...",           // When
    "command": "page retrieve",   // What
    "execution_time_ms": 234,     // How long
    "version": "5.4.0"            // Which version
  }
}
```

### Log Aggregation

```bash
# Collect all JSON responses
notion-cli page retrieve abc-123 --json >> logs/commands.jsonl

# Parse for analytics
cat logs/commands.jsonl | jq '.metadata.execution_time_ms' | stats
```

### Error Tracking

```bash
# Collect errors
notion-cli db query db-123 --json 2>&1 | \
  jq 'select(.success == false)' >> logs/errors.jsonl

# Aggregate error codes
cat logs/errors.jsonl | jq -r '.error.code' | sort | uniq -c
```

## Backward Compatibility Strategy

```
┌────────────────────────────────────┐
│   Legacy Mode (--raw flag)         │
│                                    │
│  Input:  notion-cli page retrieve  │
│          abc-123 --raw             │
│                                    │
│  Output: { "object": "page", ... } │
│                                    │
│  No envelope, no metadata          │
└────────────────────────────────────┘

┌────────────────────────────────────┐
│   Envelope Mode (--json flag)      │
│                                    │
│  Input:  notion-cli page retrieve  │
│          abc-123 --json            │
│                                    │
│  Output: {                         │
│    "success": true,                │
│    "data": { "object": "page" },   │
│    "metadata": { ... }             │
│  }                                 │
└────────────────────────────────────┘

Migration path: Users can use both flags in parallel
```

## Future Architecture Enhancements

### Streaming Envelopes

```typescript
// For long-running operations
{
  "success": "pending",
  "progress": 0.45,
  "data": { /* partial results */ },
  "metadata": { ... }
}
```

### Envelope Versioning

```typescript
{
  "envelope_version": "1.0.0",
  "success": true,
  ...
}
```

### Custom Error Codes per Command

```typescript
// Allow commands to register custom error codes
ErrorRegistry.register('page', {
  PAGE_LOCKED: { message: '...', suggestions: [...] }
})
```

## Summary

The envelope system provides a clean, layered architecture that:

1. **Separates concerns:** Command logic, envelope formatting, error handling
2. **Maintains type safety:** Generic types, type guards, TypeScript interfaces
3. **Enables observability:** Metadata, execution time, version tracking
4. **Supports automation:** Consistent structure, exit codes, suggestions
5. **Scales efficiently:** O(1) overhead, minimal memory, concurrent-safe
6. **Extends easily:** New error codes, custom metadata, output formats

All while maintaining backward compatibility and following SOLID principles.
