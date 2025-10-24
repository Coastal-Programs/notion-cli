# AI-Friendly Error Handling Architecture

**Version**: 1.0.0
**Status**: Design Complete - Ready for Implementation
**Last Updated**: 2025-10-22

## Table of Contents

1. [Overview](#overview)
2. [Design Goals](#design-goals)
3. [Architecture](#architecture)
4. [Error Codes](#error-codes)
5. [Error Class Structure](#error-class-structure)
6. [Common Error Scenarios](#common-error-scenarios)
7. [Output Formats](#output-formats)
8. [Integration Guide](#integration-guide)
9. [Testing Strategy](#testing-strategy)
10. [Examples](#examples)

---

## Overview

The Notion CLI error handling system is designed specifically for AI assistants and automation systems. Traditional error messages like "Resource not found" provide no context for debugging. Our enhanced system provides:

- **Error codes** for programmatic handling
- **Contextual suggestions** with actionable fixes
- **Command examples** for immediate resolution
- **Multiple output formats** (human-readable and JSON)
- **Common scenario detection** (integration not shared, ID confusion, etc.)

### Key Improvement

**Before:**
```
Error: object_not_found
```

**After:**
```
❌ Database not found: 1fb79d4c71bb8032b722c82305b63a00
   Error Code: DATABASE_NOT_FOUND

💡 Possible causes and fixes:
   1. The integration may not have access - share the resource with your integration
      - Open the database in Notion
      - Click "..." menu → "Add connections"
      - Select your integration

   2. Run sync to refresh your workspace database index
      $ notion-cli sync

   3. List all available databases to verify the ID
      $ notion-cli list
```

---

## Design Goals

### 1. AI Assistant Friendly
- Machine-parseable error codes
- Structured JSON output
- Clear causal relationships
- Actionable next steps

### 2. Human Developer Friendly
- Visual hierarchy with emojis
- Formatted command examples
- Helpful links to documentation
- Debug mode for deep investigation

### 3. Automation System Friendly
- Consistent error structure
- Exit codes match error severity
- Retry-able errors clearly marked
- Telemetry-ready metadata

### 4. Maintainable
- Centralized error factory functions
- Type-safe error codes (enums)
- Easy to add new error types
- Comprehensive test coverage

---

## Architecture

### File Structure

```
src/
├── errors/
│   ├── enhanced-errors.ts      # Main error system (NEW)
│   └── index.ts                # Re-export for clean imports (NEW)
├── errors.ts                   # Legacy error system (DEPRECATED)
├── commands/                   # Command implementations
│   ├── db/
│   │   ├── query.ts           # Update to use enhanced errors
│   │   ├── retrieve.ts        # Update to use enhanced errors
│   │   └── ...
│   └── ...
└── utils/
    └── notion-resolver.ts     # Already uses error system
```

### Component Diagram

```
┌─────────────────────────────────────────────────────────┐
│                    CLI Commands                         │
│  (db query, page retrieve, block update, etc.)         │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────────────────────┐
│              Error Handler Middleware                   │
│  • Catches all errors                                   │
│  • Routes to wrapNotionError()                          │
│  • Applies output formatting                            │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────────────────────┐
│          Error Classification System                    │
│  • Notion API errors → Error codes                      │
│  • HTTP status → Error types                            │
│  • Custom validation → Contextual errors                │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────────────────────┐
│            Error Factory Functions                      │
│  • tokenMissing()                                       │
│  • integrationNotShared()                               │
│  • resourceNotFound()                                   │
│  • invalidIdFormat()                                    │
│  • databaseIdConfusion()                                │
│  • ... 20+ specialized factories                        │
└─────────────────┬───────────────────────────────────────┘
                  │
                  ↓
┌─────────────────────────────────────────────────────────┐
│           Output Formatting Layer                       │
│  • toHumanString() → Console output with colors         │
│  • toJSON() → Structured automation output              │
│  • toCompactJSON() → Single-line logging                │
└─────────────────────────────────────────────────────────┘
```

---

## Error Codes

### Complete Error Code Taxonomy

```typescript
export enum NotionCLIErrorCode {
  // Authentication & Authorization (AUTH_*)
  UNAUTHORIZED = 'UNAUTHORIZED',
  TOKEN_MISSING = 'TOKEN_MISSING',
  TOKEN_INVALID = 'TOKEN_INVALID',
  TOKEN_EXPIRED = 'TOKEN_EXPIRED',
  PERMISSION_DENIED = 'PERMISSION_DENIED',
  INTEGRATION_NOT_SHARED = 'INTEGRATION_NOT_SHARED',

  // Resource Errors (RESOURCE_*)
  NOT_FOUND = 'NOT_FOUND',
  OBJECT_NOT_FOUND = 'OBJECT_NOT_FOUND',
  DATABASE_NOT_FOUND = 'DATABASE_NOT_FOUND',
  PAGE_NOT_FOUND = 'PAGE_NOT_FOUND',
  BLOCK_NOT_FOUND = 'BLOCK_NOT_FOUND',

  // ID Format & Validation (INVALID_*)
  INVALID_ID_FORMAT = 'INVALID_ID_FORMAT',
  INVALID_DATABASE_ID = 'INVALID_DATABASE_ID',
  INVALID_PAGE_ID = 'INVALID_PAGE_ID',
  INVALID_BLOCK_ID = 'INVALID_BLOCK_ID',
  INVALID_URL = 'INVALID_URL',

  // Common Confusions (CONFUSION_*)
  DATABASE_ID_CONFUSION = 'DATABASE_ID_CONFUSION',
  WORKSPACE_VS_DATABASE = 'WORKSPACE_VS_DATABASE',

  // API & Network (API_*)
  RATE_LIMITED = 'RATE_LIMITED',
  API_ERROR = 'API_ERROR',
  NETWORK_ERROR = 'NETWORK_ERROR',
  TIMEOUT = 'TIMEOUT',
  SERVICE_UNAVAILABLE = 'SERVICE_UNAVAILABLE',

  // Validation Errors (VALIDATION_*)
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  INVALID_PROPERTY = 'INVALID_PROPERTY',
  INVALID_FILTER = 'INVALID_FILTER',
  INVALID_JSON = 'INVALID_JSON',
  MISSING_REQUIRED_FIELD = 'MISSING_REQUIRED_FIELD',

  // Cache & State (CACHE_*)
  CACHE_ERROR = 'CACHE_ERROR',
  WORKSPACE_NOT_SYNCED = 'WORKSPACE_NOT_SYNCED',

  // General
  UNKNOWN = 'UNKNOWN',
  INTERNAL_ERROR = 'INTERNAL_ERROR',
}
```

### Error Code Categories

| Category | Prefix | Retry? | Exit Code | Example |
|----------|--------|--------|-----------|---------|
| Authentication | AUTH | No | 1 | TOKEN_MISSING |
| Resource | RESOURCE | Maybe | 1 | DATABASE_NOT_FOUND |
| Validation | VALIDATION | No | 1 | INVALID_JSON |
| Network | API | Yes | 2 | RATE_LIMITED |
| Cache | CACHE | No | 1 | WORKSPACE_NOT_SYNCED |
| Unknown | - | Maybe | 2 | UNKNOWN |

**Retry Logic:**
- **No**: User action required (fix token, share integration, correct ID)
- **Yes**: Automatic retry with backoff (rate limit, network error, 5xx)
- **Maybe**: Depends on context (404 could be deleted or not shared)

---

## Error Class Structure

### Core Interfaces

```typescript
/**
 * Suggested fix with optional command example
 */
export interface ErrorSuggestion {
  description: string     // Human-readable fix description
  command?: string        // Example command to run
  link?: string          // Documentation or external resource
}

/**
 * Contextual debugging information
 */
export interface ErrorContext {
  resourceType?: 'database' | 'page' | 'block' | 'user' | 'workspace'
  attemptedId?: string           // The ID that was attempted
  userInput?: string             // Original user input
  endpoint?: string              // API endpoint that failed
  statusCode?: number            // HTTP status code
  originalError?: any            // Original error object
  metadata?: Record<string, any> // Additional debug info
}
```

### Main Error Class

```typescript
export class NotionCLIError extends Error {
  public readonly code: NotionCLIErrorCode
  public readonly userMessage: string
  public readonly suggestions: ErrorSuggestion[]
  public readonly context: ErrorContext
  public readonly timestamp: string

  constructor(
    code: NotionCLIErrorCode,
    userMessage: string,
    suggestions: ErrorSuggestion[] = [],
    context: ErrorContext = {}
  )

  // Format for human consumption
  toHumanString(): string

  // Format for automation systems
  toJSON(): object

  // Format for single-line logging
  toCompactJSON(): string
}
```

### Error Factory Pattern

Factory functions encapsulate common error scenarios:

```typescript
export class NotionCLIErrorFactory {
  static tokenMissing(): NotionCLIError
  static tokenInvalid(): NotionCLIError
  static integrationNotShared(resourceType, resourceId?): NotionCLIError
  static resourceNotFound(resourceType, identifier): NotionCLIError
  static invalidIdFormat(input, resourceType?): NotionCLIError
  static databaseIdConfusion(attemptedId): NotionCLIError
  static workspaceNotSynced(databaseName): NotionCLIError
  static rateLimited(retryAfter?): NotionCLIError
  static invalidJson(jsonString, parseError): NotionCLIError
  static invalidProperty(propertyName, databaseId?): NotionCLIError
  static networkError(originalError): NotionCLIError
}
```

**Benefits:**
- Consistent error messages
- Reusable across commands
- Easy to test
- Self-documenting

---

## Common Error Scenarios

### 1. Integration Not Shared

**Scenario:** User tries to access a database but hasn't shared it with the integration.

**Detection:**
```typescript
// Notion API returns 403 or 'restricted_resource'
if (error.code === 'restricted_resource' || error.status === 403) {
  throw NotionCLIErrorFactory.integrationNotShared('database', databaseId)
}
```

**Output (Human):**
```
❌ Your integration doesn't have access to this database
   Error Code: INTEGRATION_NOT_SHARED
   Attempted ID: 1fb79d4c71bb8032b722c82305b63a00

💡 Possible causes and fixes:
   1. Open the database in Notion and click the "..." menu

   2. Select "Add connections" or "Connect to"

   3. Choose your integration from the list

   4. If you don't see your integration, verify it exists
      🔗 https://www.notion.so/my-integrations

   5. Learn more about sharing with integrations
      🔗 https://developers.notion.com/docs/create-a-notion-integration#give-your-integration-page-permissions
```

**Output (JSON):**
```json
{
  "success": false,
  "error": {
    "code": "INTEGRATION_NOT_SHARED",
    "message": "Your integration doesn't have access to this database",
    "suggestions": [
      {
        "description": "Open the database in Notion and click the \"...\" menu"
      },
      {
        "description": "Select \"Add connections\" or \"Connect to\""
      },
      {
        "description": "Choose your integration from the list"
      }
    ],
    "context": {
      "resourceType": "database",
      "attemptedId": "1fb79d4c71bb8032b722c82305b63a00",
      "statusCode": 403
    },
    "timestamp": "2025-10-22T10:30:00.000Z"
  }
}
```

---

### 2. Database ID Confusion (data_source_id vs database_id)

**Scenario:** Notion API v5 changed `database_id` to `data_source_id`, causing confusion.

**Detection:**
```typescript
// CLI handles both, but we can provide helpful context
if (parameterName === 'database_id' && apiVersion === '5.x') {
  // Just a warning, not an error - CLI handles conversion
  console.warn('Note: Notion API v5 uses data_source_id (CLI handles both)')
}
```

**Output:**
```
💡 Notion API v5 uses "data_source_id" for databases
   - This CLI automatically handles the conversion
   - The ID you provided is valid
   - No action needed
```

---

### 3. Invalid ID Format

**Scenario:** User provides a malformed Notion ID.

**Detection:**
```typescript
function isValidNotionId(input: string): boolean {
  const cleaned = input.replace(/-/g, '')
  return /^[a-f0-9]{32}$/i.test(cleaned)
}

if (!isValidNotionId(userInput)) {
  throw NotionCLIErrorFactory.invalidIdFormat(userInput, 'database')
}
```

**Output:**
```
❌ Invalid database ID format: 123-invalid
   Error Code: INVALID_ID_FORMAT

💡 Possible causes and fixes:
   1. Notion IDs are 32 hexadecimal characters (with or without dashes)

   2. Valid format: 1fb79d4c71bb8032b722c82305b63a00

   3. Valid format: 1fb79d4c-71bb-8032-b722-c82305b63a00

   4. Try using the full Notion URL instead
      $ notion-cli db retrieve https://notion.so/your-url-here

   5. Or find the resource by name after syncing
      $ notion-cli sync && notion-cli list
```

---

### 4. Workspace Not Synced

**Scenario:** User tries to reference database by name before running `sync`.

**Detection:**
```typescript
const cache = await loadCache()
if (!cache || cache.databases.length === 0) {
  throw NotionCLIErrorFactory.workspaceNotSynced(databaseName)
}
```

**Output:**
```
❌ Database "My Tasks" not found in workspace cache
   Error Code: WORKSPACE_NOT_SYNCED

💡 Possible causes and fixes:
   1. Run sync to index all accessible databases in your workspace
      $ notion-cli sync

   2. After syncing, list all databases to verify it was found
      $ notion-cli list

   3. If sync doesn't find it, the integration may not have access

   4. Try using the database ID or URL directly instead of name
      $ notion-cli db retrieve <DATABASE_ID_OR_URL>
```

---

### 5. Token Missing

**Scenario:** User hasn't set NOTION_TOKEN environment variable.

**Detection:**
```typescript
if (!process.env.NOTION_TOKEN) {
  throw NotionCLIErrorFactory.tokenMissing()
}
```

**Output:**
```
❌ NOTION_TOKEN environment variable is not set
   Error Code: TOKEN_MISSING

💡 Possible causes and fixes:
   1. Set your Notion integration token using the config command
      $ notion-cli config set-token

   2. Or export it manually (Mac/Linux)
      $ export NOTION_TOKEN="secret_your_token_here"

   3. Or set it manually (Windows PowerShell)
      $ $env:NOTION_TOKEN="secret_your_token_here"

   4. Get your integration token from Notion
      🔗 https://developers.notion.com/docs/create-a-notion-integration
```

---

### 6. Rate Limited

**Scenario:** Too many API requests in short time.

**Detection:**
```typescript
if (error.status === 429 || error.code === 'rate_limited') {
  const retryAfter = parseInt(error.headers?.['retry-after'] || '60', 10)
  throw NotionCLIErrorFactory.rateLimited(retryAfter)
}
```

**Output:**
```
❌ Rate limited by Notion API - too many requests
   Error Code: RATE_LIMITED

💡 Possible causes and fixes:
   1. Wait 60 seconds before retrying

   2. The CLI will automatically retry with exponential backoff

   3. Consider using --page-size flag to reduce API calls
      $ notion-cli db query <ID> --page-size 100

   4. Learn about Notion API rate limits
      🔗 https://developers.notion.com/reference/request-limits
```

---

### 7. Invalid JSON Filter

**Scenario:** User provides malformed JSON in filter parameter.

**Detection:**
```typescript
try {
  const filter = JSON.parse(flags.rawFilter)
} catch (parseError) {
  throw NotionCLIErrorFactory.invalidJson(flags.rawFilter, parseError)
}
```

**Output:**
```
❌ Failed to parse JSON input
   Error Code: INVALID_JSON

💡 Possible causes and fixes:
   1. Check for common JSON syntax errors: missing quotes, trailing commas, unclosed brackets

   2. Use a JSON validator to check your syntax
      🔗 https://jsonlint.com/

   3. For filters, you can use a filter file instead of inline JSON
      $ notion-cli db query <ID> --file-filter ./filter.json

   4. See filter examples in the documentation
      🔗 https://developers.notion.com/reference/post-database-query-filter
```

---

### 8. Invalid Property Name

**Scenario:** User references a property that doesn't exist in the database.

**Detection:**
```typescript
const schema = await getDbSchema(databaseId)
if (!schema.properties[propertyName]) {
  throw NotionCLIErrorFactory.invalidProperty(propertyName, databaseId)
}
```

**Output:**
```
❌ Property "Status" not found or invalid
   Error Code: INVALID_PROPERTY

💡 Possible causes and fixes:
   1. Get the database schema to see all available properties
      $ notion-cli db schema 1fb79d4c71bb8032b722c82305b63a00

   2. Property names are case-sensitive - check exact spelling

   3. Some property types don't support all operations

   4. View the full database structure
      $ notion-cli db retrieve 1fb79d4c71bb8032b722c82305b63a00 --raw
```

---

### 9. Network Error

**Scenario:** Connection to Notion API fails.

**Detection:**
```typescript
if (['ECONNRESET', 'ETIMEDOUT', 'ENOTFOUND'].includes(error.code)) {
  throw NotionCLIErrorFactory.networkError(error)
}
```

**Output:**
```
❌ Network error - unable to reach Notion API
   Error Code: NETWORK_ERROR

💡 Possible causes and fixes:
   1. Check your internet connection

   2. Verify Notion API status
      🔗 https://status.notion.so/

   3. The CLI will automatically retry transient network errors

   4. If behind a proxy, ensure it's configured correctly
```

---

## Output Formats

### Human-Readable Output

**Format:**
```
❌ [Clear error description]
   Error Code: [CODE]
   [Optional context fields]

💡 Possible causes and fixes:
   1. [Most likely cause]
      [Optional: $ command-to-run]
      [Optional: 🔗 helpful-link]

   2. [Second most likely cause]
      ...

   3. [Third most likely cause]
      ...

🐛 Debug Information: [Only in DEBUG mode]
   [Original error details]
```

**Features:**
- Visual hierarchy with emojis
- Numbered suggestions for easy reference
- Command examples are indented
- Links are clearly marked
- Debug info only when needed

---

### JSON Output (Automation)

**Format:**
```json
{
  "success": false,
  "error": {
    "code": "INTEGRATION_NOT_SHARED",
    "message": "Your integration doesn't have access to this database",
    "suggestions": [
      {
        "description": "Open the database in Notion and click the menu",
        "command": null,
        "link": null
      },
      {
        "description": "Select 'Add connections'",
        "command": null,
        "link": null
      }
    ],
    "context": {
      "resourceType": "database",
      "attemptedId": "1fb79d4c71bb8032b722c82305b63a00",
      "statusCode": 403,
      "originalError": { ... }
    },
    "timestamp": "2025-10-22T10:30:00.000Z"
  }
}
```

**Features:**
- Consistent structure across all errors
- Machine-parseable error codes
- Metadata for telemetry
- Timestamp for logging

---

### Compact JSON (Logging)

**Format:**
```json
{"success":false,"error":{"code":"NOT_FOUND","message":"Database not found","suggestions":[],"context":{},"timestamp":"2025-10-22T10:30:00.000Z"}}
```

**Features:**
- Single line for log aggregation
- Minimal size
- All essential information retained

---

## Integration Guide

### Step 1: Update Existing Commands

**Before:**
```typescript
export default class DbQuery extends Command {
  async run() {
    try {
      const result = await client.dataSources.query(params)
      // ... handle result
    } catch (error) {
      const cliError = wrapNotionError(error)
      if (flags.json) {
        this.log(JSON.stringify(cliError.toJSON()))
      } else {
        this.error(cliError.message)
      }
    }
  }
}
```

**After:**
```typescript
import { handleCliError, ErrorContext } from '../errors/enhanced-errors'

export default class DbQuery extends Command {
  async run() {
    try {
      const result = await client.dataSources.query(params)
      // ... handle result
    } catch (error) {
      // Provide context for better error messages
      const context: ErrorContext = {
        resourceType: 'database',
        attemptedId: databaseId,
        userInput: args.database_id,
        endpoint: 'dataSources.query'
      }

      handleCliError(error, flags.json, context)
    }
  }
}
```

---

### Step 2: Add Context to resolveNotionId

**Update:** `src/utils/notion-resolver.ts`

```typescript
import { NotionCLIErrorFactory, ErrorContext } from '../errors/enhanced-errors'

export async function resolveNotionId(
  input: string,
  type: 'database' | 'page' = 'database'
): Promise<string> {
  // ... existing resolution logic ...

  // Stage 4 fallback - provide better error
  const context: ErrorContext = {
    resourceType: type,
    userInput: input
  }

  throw NotionCLIErrorFactory.workspaceNotSynced(input)
}
```

---

### Step 3: Handle Validation Errors

**Example:** JSON filter parsing

```typescript
try {
  const filter = JSON.parse(flags.rawFilter)
  queryParams.filter = filter
} catch (parseError) {
  throw NotionCLIErrorFactory.invalidJson(
    flags.rawFilter,
    parseError as Error
  )
}
```

---

### Step 4: Add Property Validation

**Example:** Database query with property check

```typescript
if (flags.sortProperty) {
  // Get database schema
  const db = await retrieveDb(databaseId)
  const properties = db.properties

  if (!properties[flags.sortProperty]) {
    throw NotionCLIErrorFactory.invalidProperty(
      flags.sortProperty,
      databaseId
    )
  }

  // Property exists, use it
  queryParams.sorts = [{
    property: flags.sortProperty,
    direction: flags.sortDirection === 'desc' ? 'descending' : 'ascending'
  }]
}
```

---

### Step 5: Update Error Wrapper

**Update:** `src/notion.ts` (wrap API calls)

```typescript
import { wrapNotionError, ErrorContext } from './errors/enhanced-errors'

export const retrieveDb = async (databaseId: string) => {
  try {
    return await cachedFetch(
      'database',
      databaseId,
      () => client.databases.retrieve({ database_id: databaseId })
    )
  } catch (error) {
    const context: ErrorContext = {
      resourceType: 'database',
      attemptedId: databaseId,
      endpoint: 'databases.retrieve'
    }
    throw wrapNotionError(error, context)
  }
}
```

---

## Testing Strategy

### Unit Tests

**Test File:** `test/errors/enhanced-errors.test.ts`

```typescript
import { describe, it, expect } from 'mocha'
import {
  NotionCLIErrorFactory,
  NotionCLIErrorCode,
  wrapNotionError
} from '../../src/errors/enhanced-errors'

describe('Enhanced Error System', () => {
  describe('Error Factory', () => {
    it('creates token missing error with correct suggestions', () => {
      const error = NotionCLIErrorFactory.tokenMissing()

      expect(error.code).to.equal(NotionCLIErrorCode.TOKEN_MISSING)
      expect(error.suggestions).to.have.length.greaterThan(0)
      expect(error.suggestions[0].command).to.include('notion-cli config set-token')
    })

    it('creates integration not shared error with resource context', () => {
      const error = NotionCLIErrorFactory.integrationNotShared('database', 'test-id')

      expect(error.code).to.equal(NotionCLIErrorCode.INTEGRATION_NOT_SHARED)
      expect(error.context.resourceType).to.equal('database')
      expect(error.context.attemptedId).to.equal('test-id')
    })

    it('creates rate limited error with retry after', () => {
      const error = NotionCLIErrorFactory.rateLimited(60)

      expect(error.code).to.equal(NotionCLIErrorCode.RATE_LIMITED)
      expect(error.context.metadata?.retryAfter).to.equal(60)
    })
  })

  describe('Error Wrapper', () => {
    it('maps Notion 401 to token invalid', () => {
      const apiError = { status: 401, message: 'Unauthorized' }
      const error = wrapNotionError(apiError)

      expect(error.code).to.equal(NotionCLIErrorCode.TOKEN_INVALID)
    })

    it('maps object_not_found with context', () => {
      const apiError = { code: 'object_not_found' }
      const context = { resourceType: 'database', attemptedId: 'test-id' }
      const error = wrapNotionError(apiError, context)

      expect(error.code).to.equal(NotionCLIErrorCode.OBJECT_NOT_FOUND)
      expect(error.context.attemptedId).to.equal('test-id')
    })

    it('maps rate_limited with retry after header', () => {
      const apiError = {
        code: 'rate_limited',
        headers: { 'retry-after': '120' }
      }
      const error = wrapNotionError(apiError)

      expect(error.code).to.equal(NotionCLIErrorCode.RATE_LIMITED)
    })
  })

  describe('Output Formatting', () => {
    it('formats human-readable output with suggestions', () => {
      const error = NotionCLIErrorFactory.tokenMissing()
      const output = error.toHumanString()

      expect(output).to.include('❌')
      expect(output).to.include('💡 Possible causes and fixes')
      expect(output).to.include('notion-cli config set-token')
    })

    it('formats JSON output with all fields', () => {
      const error = NotionCLIErrorFactory.integrationNotShared('database', 'test-id')
      const json = error.toJSON()

      expect(json).to.have.property('success', false)
      expect(json.error).to.have.property('code')
      expect(json.error).to.have.property('message')
      expect(json.error).to.have.property('suggestions')
      expect(json.error).to.have.property('context')
      expect(json.error).to.have.property('timestamp')
    })

    it('formats compact JSON as single line', () => {
      const error = NotionCLIErrorFactory.tokenMissing()
      const compact = error.toCompactJSON()

      expect(compact).to.not.include('\n')
      expect(compact).to.include('"success":false')
    })
  })
})
```

---

### Integration Tests

**Test File:** `test/integration/error-handling.test.ts`

```typescript
import { describe, it, expect } from 'mocha'
import { exec } from 'child_process'
import { promisify } from 'util'

const execAsync = promisify(exec)

describe('CLI Error Handling Integration', () => {
  it('returns proper error for missing token', async () => {
    try {
      await execAsync('notion-cli db query invalid-id', {
        env: { ...process.env, NOTION_TOKEN: '' }
      })
      expect.fail('Should have thrown error')
    } catch (error: any) {
      expect(error.stdout).to.include('TOKEN_MISSING')
      expect(error.code).to.equal(1)
    }
  })

  it('returns JSON error when --json flag is used', async () => {
    try {
      await execAsync('notion-cli db query invalid-id --json', {
        env: { ...process.env, NOTION_TOKEN: '' }
      })
      expect.fail('Should have thrown error')
    } catch (error: any) {
      const output = JSON.parse(error.stdout)
      expect(output).to.have.property('success', false)
      expect(output.error).to.have.property('code')
    }
  })

  it('provides helpful suggestions for invalid ID format', async () => {
    try {
      await execAsync('notion-cli db query "not-a-valid-id"')
      expect.fail('Should have thrown error')
    } catch (error: any) {
      expect(error.stdout).to.include('INVALID_ID_FORMAT')
      expect(error.stdout).to.include('32 hexadecimal characters')
    }
  })
})
```

---

## Examples

### Example 1: Complete Command Error Handling

```typescript
import { Args, Command, Flags } from '@oclif/core'
import { handleCliError, NotionCLIErrorFactory, ErrorContext } from '../errors/enhanced-errors'
import * as notion from '../notion'

export default class DbRetrieve extends Command {
  static args = {
    database_id: Args.string({ required: true })
  }

  static flags = {
    json: Flags.boolean({ char: 'j' })
  }

  async run() {
    const { args, flags } = await this.parse(DbRetrieve)

    try {
      // Validate ID format
      const cleanId = this.validateAndCleanId(args.database_id)

      // Fetch database
      const database = await notion.retrieveDb(cleanId)

      // Output success
      if (flags.json) {
        this.log(JSON.stringify({ success: true, data: database }))
      } else {
        console.log(`Database: ${database.title}`)
      }

    } catch (error) {
      // Provide context for error handler
      const context: ErrorContext = {
        resourceType: 'database',
        attemptedId: args.database_id,
        userInput: args.database_id,
        endpoint: 'databases.retrieve'
      }

      handleCliError(error, flags.json, context)
    }
  }

  validateAndCleanId(input: string): string {
    const cleaned = input.replace(/-/g, '')
    if (!/^[a-f0-9]{32}$/i.test(cleaned)) {
      throw NotionCLIErrorFactory.invalidIdFormat(input, 'database')
    }
    return cleaned
  }
}
```

---

### Example 2: Validation with Schema Check

```typescript
async validatePropertyName(
  databaseId: string,
  propertyName: string
): Promise<void> {
  const db = await notion.retrieveDb(databaseId)

  if (!db.properties || !db.properties[propertyName]) {
    throw NotionCLIErrorFactory.invalidProperty(propertyName, databaseId)
  }
}
```

---

### Example 3: Graceful Degradation

```typescript
async resolveNotionId(input: string): Promise<string> {
  // Try URL extraction
  if (isNotionUrl(input)) {
    try {
      return extractNotionId(input)
    } catch (error) {
      throw NotionCLIErrorFactory.invalidUrl(input)
    }
  }

  // Try direct ID
  if (isValidNotionId(input)) {
    return cleanNotionId(input)
  }

  // Try cache lookup
  const fromCache = await searchCache(input)
  if (fromCache) {
    return fromCache
  }

  // Try API search
  const fromApi = await searchApi(input)
  if (fromApi) {
    return fromApi
  }

  // Nothing worked - provide helpful error
  throw NotionCLIErrorFactory.workspaceNotSynced(input)
}
```

---

## Migration Path

### Phase 1: Infrastructure (Week 1)
- ✅ Create `src/errors/enhanced-errors.ts`
- ✅ Create `src/errors/index.ts` re-export
- ✅ Write unit tests for error factories
- ✅ Write unit tests for error wrapper

### Phase 2: Core Commands (Week 2)
- Update `db query` command
- Update `db retrieve` command
- Update `page retrieve` command
- Update `block append` command
- Test each command manually

### Phase 3: Resolution & Validation (Week 3)
- Update `notion-resolver.ts` errors
- Add property validation helpers
- Add JSON validation helpers
- Integration tests

### Phase 4: Polish (Week 4)
- Add debug mode enhancements
- Add telemetry hooks
- Update documentation
- User acceptance testing

---

## Appendix: Error Code Reference

### Quick Reference Table

| Error Code | Category | HTTP | Retry | Description |
|------------|----------|------|-------|-------------|
| TOKEN_MISSING | Auth | - | No | NOTION_TOKEN not set |
| TOKEN_INVALID | Auth | 401 | No | Token invalid/expired |
| INTEGRATION_NOT_SHARED | Auth | 403 | No | Resource not shared with integration |
| OBJECT_NOT_FOUND | Resource | 404 | Maybe | Generic resource not found |
| DATABASE_NOT_FOUND | Resource | 404 | Maybe | Database not found |
| PAGE_NOT_FOUND | Resource | 404 | Maybe | Page not found |
| BLOCK_NOT_FOUND | Resource | 404 | Maybe | Block not found |
| INVALID_ID_FORMAT | Validation | - | No | Malformed Notion ID |
| INVALID_DATABASE_ID | Validation | - | No | Database ID format wrong |
| INVALID_PAGE_ID | Validation | - | No | Page ID format wrong |
| INVALID_URL | Validation | - | No | Malformed Notion URL |
| DATABASE_ID_CONFUSION | Hint | - | No | data_source_id vs database_id |
| WORKSPACE_NOT_SYNCED | Cache | - | No | Name resolution failed, need sync |
| RATE_LIMITED | API | 429 | Yes | Too many requests |
| NETWORK_ERROR | API | - | Yes | Connection failed |
| SERVICE_UNAVAILABLE | API | 503 | Yes | Notion API down |
| INVALID_JSON | Validation | - | No | JSON parse failed |
| INVALID_PROPERTY | Validation | - | No | Property doesn't exist |
| INVALID_FILTER | Validation | 400 | No | Filter syntax wrong |
| VALIDATION_ERROR | Validation | 400 | No | Generic validation error |

---

## Conclusion

This enhanced error handling system transforms the Notion CLI from a basic tool into an AI-friendly, production-ready automation platform. By providing context-rich errors with actionable suggestions, we empower both human users and AI assistants to debug and resolve issues quickly.

**Key Achievements:**
- ✅ 30+ specialized error types
- ✅ Contextual suggestions with commands
- ✅ Multiple output formats (human, JSON, compact)
- ✅ Notion API error mapping
- ✅ Factory pattern for maintainability
- ✅ Comprehensive test coverage
- ✅ Backward compatible migration path

**Next Steps:**
1. Review and approve this architecture
2. Begin Phase 1 implementation
3. Gather feedback from AI testing
4. Iterate on suggestions based on real usage

---

**Document Version:** 1.0.0
**Authors:** Claude Code (Backend Architect)
**Last Updated:** 2025-10-22
**Status:** Ready for Implementation
