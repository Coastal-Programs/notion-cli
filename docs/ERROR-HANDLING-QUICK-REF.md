# Error Handling Quick Reference

**Version**: 1.0.0
**Last Updated**: 2025-10-22

Quick reference guide for using the enhanced error handling system in Notion CLI.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [Common Patterns](#common-patterns)
3. [Error Code Reference](#error-code-reference)
4. [Factory Functions](#factory-functions)
5. [Output Examples](#output-examples)
6. [Testing Checklist](#testing-checklist)

---

## Quick Start

### Import Error System

```typescript
import {
  NotionCLIError,
  NotionCLIErrorCode,
  NotionCLIErrorFactory,
  handleCliError,
  wrapNotionError,
  ErrorContext
} from '../errors'
```

### Basic Command Error Handling

```typescript
export default class MyCommand extends Command {
  async run() {
    const { args, flags } = await this.parse(MyCommand)

    try {
      // Your command logic
      const result = await someOperation()

      // Success output
      if (flags.json) {
        this.log(JSON.stringify({ success: true, data: result }))
      } else {
        console.log('Operation successful')
      }

    } catch (error) {
      // Provide context
      const context: ErrorContext = {
        resourceType: 'database',
        attemptedId: args.database_id,
        userInput: args.database_id
      }

      // Handle error with proper formatting
      handleCliError(error, flags.json, context)
    }
  }
}
```

---

## Common Patterns

### Pattern 1: Validate ID Format

```typescript
function validateNotionId(input: string, type: 'database' | 'page' = 'database'): string {
  const cleaned = input.replace(/-/g, '')

  if (!/^[a-f0-9]{32}$/i.test(cleaned)) {
    throw NotionCLIErrorFactory.invalidIdFormat(input, type)
  }

  return cleaned
}
```

### Pattern 2: Check Token

```typescript
function ensureToken(): void {
  if (!process.env.NOTION_TOKEN) {
    throw NotionCLIErrorFactory.tokenMissing()
  }
}
```

### Pattern 3: Validate JSON Input

```typescript
function parseFilterJson(jsonString: string): object {
  try {
    return JSON.parse(jsonString)
  } catch (parseError) {
    throw NotionCLIErrorFactory.invalidJson(jsonString, parseError as Error)
  }
}
```

### Pattern 4: Check Property Exists

```typescript
async function validateProperty(
  databaseId: string,
  propertyName: string
): Promise<void> {
  const db = await retrieveDb(databaseId)

  if (!db.properties || !db.properties[propertyName]) {
    throw NotionCLIErrorFactory.invalidProperty(propertyName, databaseId)
  }
}
```

### Pattern 5: Wrap API Calls

```typescript
async function safeDatabaseQuery(params: QueryParams) {
  try {
    return await client.dataSources.query(params)
  } catch (error) {
    const context: ErrorContext = {
      resourceType: 'database',
      attemptedId: params.data_source_id,
      endpoint: 'dataSources.query'
    }
    throw wrapNotionError(error, context)
  }
}
```

### Pattern 6: Check Cache State

```typescript
async function ensureWorkspaceSynced(databaseName: string): Promise<void> {
  const cache = await loadCache()

  if (!cache || cache.databases.length === 0) {
    throw NotionCLIErrorFactory.workspaceNotSynced(databaseName)
  }
}
```

---

## Error Code Reference

### By Category

#### Authentication (Use when...)
- `TOKEN_MISSING` - NOTION_TOKEN env var not set
- `TOKEN_INVALID` - API returns 401/403 with invalid token
- `TOKEN_EXPIRED` - Token no longer valid
- `INTEGRATION_NOT_SHARED` - 403 error on resource access

#### Resource Not Found (Use when...)
- `OBJECT_NOT_FOUND` - Generic 404, don't know resource type
- `DATABASE_NOT_FOUND` - Specific database 404
- `PAGE_NOT_FOUND` - Specific page 404
- `BLOCK_NOT_FOUND` - Specific block 404

#### Validation (Use when...)
- `INVALID_ID_FORMAT` - ID doesn't match 32 hex chars pattern
- `INVALID_DATABASE_ID` - Database-specific ID validation failed
- `INVALID_PAGE_ID` - Page-specific ID validation failed
- `INVALID_URL` - Notion URL parsing failed
- `INVALID_JSON` - JSON.parse() threw error
- `INVALID_PROPERTY` - Property name doesn't exist in schema
- `INVALID_FILTER` - Filter syntax error (400 from API)

#### Network & API (Use when...)
- `RATE_LIMITED` - 429 status or rate_limited error code
- `NETWORK_ERROR` - ECONNRESET, ETIMEDOUT, etc.
- `SERVICE_UNAVAILABLE` - 503/504 from Notion API
- `TIMEOUT` - Request exceeded timeout threshold

#### Cache & State (Use when...)
- `WORKSPACE_NOT_SYNCED` - Name lookup failed, no cache
- `CACHE_ERROR` - Cache read/write failure

#### Special Cases (Use when...)
- `DATABASE_ID_CONFUSION` - Educational hint about API v5 changes
- `WORKSPACE_VS_DATABASE` - User passed workspace ID instead of DB ID

---

## Factory Functions

### Authentication Errors

```typescript
// Token not set
throw NotionCLIErrorFactory.tokenMissing()

// Token invalid or expired
throw NotionCLIErrorFactory.tokenInvalid()

// Resource not shared with integration
throw NotionCLIErrorFactory.integrationNotShared('database', databaseId)
throw NotionCLIErrorFactory.integrationNotShared('page', pageId)
```

### Resource Errors

```typescript
// Resource not found (with context-aware suggestions)
throw NotionCLIErrorFactory.resourceNotFound('database', identifier)
throw NotionCLIErrorFactory.resourceNotFound('page', identifier)
throw NotionCLIErrorFactory.resourceNotFound('block', identifier)
```

### Validation Errors

```typescript
// Invalid ID format
throw NotionCLIErrorFactory.invalidIdFormat(userInput, 'database')

// Invalid JSON
throw NotionCLIErrorFactory.invalidJson(jsonString, parseError)

// Invalid property name
throw NotionCLIErrorFactory.invalidProperty(propertyName, databaseId)
```

### API Errors

```typescript
// Rate limited
throw NotionCLIErrorFactory.rateLimited(60) // with retry after seconds
throw NotionCLIErrorFactory.rateLimited()   // without retry after

// Network error
throw NotionCLIErrorFactory.networkError(originalError)
```

### Cache Errors

```typescript
// Workspace not synced
throw NotionCLIErrorFactory.workspaceNotSynced(databaseName)
```

### Special Cases

```typescript
// Database ID confusion (API v5 change)
throw NotionCLIErrorFactory.databaseIdConfusion(attemptedId)
```

---

## Output Examples

### Human-Readable (Default)

```bash
$ notion-cli db query invalid-id

❌ Invalid database ID format: invalid-id
   Error Code: INVALID_ID_FORMAT

💡 Possible causes and fixes:
   1. Notion IDs are 32 hexadecimal characters (with or without dashes)

   2. Valid format: 1fb79d4c71bb8032b722c82305b63a00

   3. Valid format: 1fb79d4c-71bb-8032-b722-c82305b63a00

   4. Try using the full Notion URL instead
      $ notion-cli db query https://notion.so/your-url-here

   5. Or find the resource by name after syncing
      $ notion-cli sync && notion-cli list
```

### JSON Output (--json flag)

```json
{
  "success": false,
  "error": {
    "code": "INVALID_ID_FORMAT",
    "message": "Invalid database ID format: invalid-id",
    "suggestions": [
      {
        "description": "Notion IDs are 32 hexadecimal characters (with or without dashes)"
      },
      {
        "description": "Valid format: 1fb79d4c71bb8032b722c82305b63a00"
      },
      {
        "description": "Try using the full Notion URL instead",
        "command": "notion-cli db query https://notion.so/your-url-here"
      }
    ],
    "context": {
      "resourceType": "database",
      "userInput": "invalid-id",
      "attemptedId": "invalid-id"
    },
    "timestamp": "2025-10-22T10:30:00.000Z"
  }
}
```

### Compact JSON (Logging)

```json
{"success":false,"error":{"code":"INVALID_ID_FORMAT","message":"Invalid database ID format: invalid-id","suggestions":[{"description":"Notion IDs are 32 hexadecimal characters"}],"context":{"resourceType":"database","userInput":"invalid-id"},"timestamp":"2025-10-22T10:30:00.000Z"}}
```

---

## Testing Checklist

### Unit Tests (For Each Error Factory)

```typescript
describe('NotionCLIErrorFactory.tokenMissing', () => {
  it('has correct error code')
  it('includes config set-token command in suggestions')
  it('includes export command for Mac/Linux')
  it('includes PowerShell command for Windows')
  it('includes link to Notion integration docs')
})
```

### Integration Tests (For Each Command)

```typescript
describe('db query command', () => {
  it('returns TOKEN_MISSING when token not set')
  it('returns INVALID_ID_FORMAT for bad ID')
  it('returns INTEGRATION_NOT_SHARED for 403 error')
  it('returns RATE_LIMITED for 429 error')
  it('outputs JSON format when --json flag used')
})
```

### Manual Test Scenarios

1. **No Token Set**
   ```bash
   unset NOTION_TOKEN
   notion-cli db query test-id
   # Should show TOKEN_MISSING with setup instructions
   ```

2. **Invalid ID Format**
   ```bash
   notion-cli db query "not-a-valid-id"
   # Should show INVALID_ID_FORMAT with examples
   ```

3. **Resource Not Found**
   ```bash
   notion-cli db query 00000000000000000000000000000000
   # Should show DATABASE_NOT_FOUND with sync suggestion
   ```

4. **Integration Not Shared**
   ```bash
   notion-cli db query [valid-id-not-shared]
   # Should show INTEGRATION_NOT_SHARED with sharing steps
   ```

5. **Workspace Not Synced**
   ```bash
   rm ~/.notion-cli/databases.json
   notion-cli db query "My Database"
   # Should show WORKSPACE_NOT_SYNCED with sync command
   ```

6. **Invalid JSON Filter**
   ```bash
   notion-cli db query test-id --raw-filter '{invalid json'
   # Should show INVALID_JSON with validator link
   ```

7. **Rate Limited**
   ```bash
   # Make many rapid requests
   for i in {1..100}; do notion-cli db query test-id; done
   # Should show RATE_LIMITED with retry info
   ```

8. **JSON Output Mode**
   ```bash
   notion-cli db query invalid-id --json
   # Should output valid JSON error structure
   ```

---

## Debug Mode

Enable debug output for additional context:

```bash
export DEBUG=1
notion-cli db query test-id
```

Output includes:
- Original error stack trace
- Full API response
- Request parameters
- Cache state

---

## Migration Checklist

### For Each Command

- [ ] Import enhanced error system
- [ ] Add try-catch with context
- [ ] Replace generic errors with factory functions
- [ ] Add validation before API calls
- [ ] Test human output format
- [ ] Test JSON output format
- [ ] Add unit tests
- [ ] Add integration tests
- [ ] Update command documentation

### Priority Order

1. **High Traffic Commands** (Week 1)
   - db query
   - db retrieve
   - page retrieve
   - list

2. **Core Commands** (Week 2)
   - db create
   - page create
   - block append
   - block update

3. **Utility Commands** (Week 3)
   - sync
   - config set-token
   - db schema

4. **Remaining Commands** (Week 4)
   - All other commands
   - Cleanup legacy error system

---

## Quick Decision Tree

```
Error occurred
├── Is it authentication related?
│   ├── Token not set? → tokenMissing()
│   ├── 401/403? → tokenInvalid()
│   └── Integration not shared? → integrationNotShared()
│
├── Is it validation related?
│   ├── Invalid ID format? → invalidIdFormat()
│   ├── JSON parse failed? → invalidJson()
│   ├── Property doesn't exist? → invalidProperty()
│   └── Bad filter syntax? → Use wrapNotionError()
│
├── Is it a resource error?
│   ├── 404 for database? → resourceNotFound('database', id)
│   ├── 404 for page? → resourceNotFound('page', id)
│   ├── 404 for block? → resourceNotFound('block', id)
│   └── Name not in cache? → workspaceNotSynced()
│
├── Is it API/Network related?
│   ├── 429 rate limit? → rateLimited()
│   ├── Network failure? → networkError()
│   ├── 503/504? → Use wrapNotionError()
│   └── Unknown API error? → Use wrapNotionError()
│
└── Unknown error? → Use wrapNotionError()
```

---

## Performance Impact

### Minimal Overhead

- Error object creation: <1ms
- Suggestion generation: <1ms
- JSON serialization: <5ms
- String formatting: <10ms

**Total**: <20ms overhead per error (negligible for CLI)

### Memory Usage

- Base error object: ~2KB
- With full context: ~5KB
- JSON output: ~3KB

**Impact**: Minimal - errors are short-lived

---

## Best Practices

### DO

✅ Provide context when wrapping errors
✅ Use specific factory functions when available
✅ Test both human and JSON output modes
✅ Include debug information in originalError
✅ Add helpful links to documentation
✅ Keep suggestion lists focused (3-5 items)

### DON'T

❌ Throw generic Error() objects
❌ Lose original error information
❌ Forget to handle --json flag
❌ Make suggestions too technical
❌ Skip validation in favor of API errors
❌ Create new error codes without updating docs

---

## Resources

- [Full Architecture Document](./ERROR-HANDLING-ARCHITECTURE.md)
- [Implementation File](../src/errors/enhanced-errors.ts)
- [Notion API Error Codes](https://developers.notion.com/reference/errors)
- [HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status)

---

**Document Version**: 1.0.0
**Last Updated**: 2025-10-22
**Maintained By**: Backend Architecture Team
