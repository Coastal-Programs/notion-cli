# Envelope Integration Guide

## Overview

This guide provides step-by-step instructions for integrating the JSON envelope system into existing and new commands. It includes code examples, migration patterns, and best practices.

**Target Audience:** Developers modifying or adding commands to the Notion CLI

## Quick Start

### For New Commands

```typescript
import { BaseCommand } from '../../base-command'
import { Args, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'

export default class PageRetrieve extends BaseCommand {
  static description = 'Retrieve a page'

  static args = {
    page_id: Args.string({ required: true, description: 'Page ID' })
  }

  static flags = {
    ...ux.table.flags(),
    ...AutomationFlags,
    ...OutputFormatFlags,
  }

  async run(): Promise<void> {
    const { args, flags } = await this.parse(PageRetrieve)

    try {
      // Fetch data
      const page = await notion.retrievePage({ page_id: args.page_id })

      // Handle JSON output with envelope
      if (flags.json || flags['compact-json']) {
        this.outputSuccess(page, flags)
        // Never reached - outputSuccess calls process.exit
      }

      // Handle raw output (bypasses envelope)
      if (flags.raw) {
        this.log(JSON.stringify(page, null, 2))
        process.exit(0)
      }

      // Handle other formats (table, markdown, etc.)
      // ... existing table output logic ...

    } catch (error) {
      this.outputError(error, flags)
      // Never reached - outputError calls process.exit
    }
  }
}
```

### For Existing Commands

See the [Migration Checklist](#migration-checklist) below.

## Command Patterns

### Pattern 1: Simple Retrieve Command

**Use Case:** Commands that retrieve a single resource

```typescript
import { BaseCommand } from '../../base-command'
import { Args, ux } from '@oclif/core'
import * as notion from '../../notion'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'
import { getPageTitle, outputRawJson } from '../../helper'
import { PageObjectResponse } from '@notionhq/client/build/src/api-endpoints'

export default class PageRetrieve extends BaseCommand {
  static description = 'Retrieve a page'
  static args = {
    page_id: Args.string({ required: true })
  }
  static flags = {
    ...ux.table.flags(),
    ...AutomationFlags,
    ...OutputFormatFlags,
  }

  async run(): Promise<void> {
    const { args, flags } = await this.parse(PageRetrieve)

    try {
      const page = await notion.retrievePage({ page_id: args.page_id })

      // Envelope output (--json or --compact-json)
      if (flags.json || flags['compact-json']) {
        this.outputSuccess(page, flags)
      }

      // Raw output (--raw)
      if (flags.raw) {
        outputRawJson(page)
        process.exit(0)
      }

      // Table output (default)
      const columns = {
        title: { get: (row: PageObjectResponse) => getPageTitle(row) },
        object: {},
        id: {},
        url: {},
      }
      ux.table([page], columns, { ...flags, printLine: this.log.bind(this) })
      process.exit(0)

    } catch (error) {
      this.outputError(error, flags)
    }
  }
}
```

### Pattern 2: Query/List Command with Pagination

**Use Case:** Commands that return multiple results with optional pagination

```typescript
import { BaseCommand } from '../../base-command'
import { Args, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'

export default class DbQuery extends BaseCommand {
  static description = 'Query a database'
  static args = {
    database_id: Args.string({ required: true })
  }
  static flags = {
    pageSize: Flags.integer({ char: 'p', default: 10, min: 1, max: 100 }),
    pageAll: Flags.boolean({ char: 'A', default: false }),
    ...ux.table.flags(),
    ...AutomationFlags,
    ...OutputFormatFlags,
  }

  async run(): Promise<void> {
    const { args, flags } = await this.parse(DbQuery)

    try {
      // Fetch pages
      let pages = []
      if (flags.pageAll) {
        pages = await notion.fetchAllPagesInDS(args.database_id)
      } else {
        const res = await notion.client.dataSources.query({
          data_source_id: args.database_id,
          page_size: flags.pageSize,
        })
        pages = res.results
      }

      // Envelope output with pagination metadata
      if (flags.json || flags['compact-json']) {
        this.outputSuccess(pages, flags, {
          total_results: pages.length,
          page_size: flags.pageSize,
        })
      }

      // Raw output
      if (flags.raw) {
        this.log(JSON.stringify(pages, null, 2))
        process.exit(0)
      }

      // Table output
      const columns = { /* ... */ }
      ux.table(pages, columns, { ...flags, printLine: this.log.bind(this) })
      process.exit(0)

    } catch (error) {
      this.outputError(error, flags)
    }
  }
}
```

### Pattern 3: Create/Update Command

**Use Case:** Commands that modify resources

```typescript
import { BaseCommand } from '../../base-command'
import { Args, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'

export default class PageCreate extends BaseCommand {
  static description = 'Create a page'
  static args = {
    parent_id: Args.string({ required: true })
  }
  static flags = {
    title: Flags.string({ char: 't', required: true }),
    ...ux.table.flags(),
    ...AutomationFlags,
    ...OutputFormatFlags,
  }

  async run(): Promise<void> {
    const { args, flags } = await this.parse(PageCreate)

    try {
      const page = await notion.createPage({
        parent: { database_id: args.parent_id },
        properties: {
          title: { title: [{ text: { content: flags.title } }] }
        }
      })

      // Envelope output
      if (flags.json || flags['compact-json']) {
        this.outputSuccess(page, flags, {
          operation: 'create',
          resource_type: 'page',
        })
      }

      // Raw output
      if (flags.raw) {
        this.log(JSON.stringify(page, null, 2))
        process.exit(0)
      }

      // Human-readable confirmation
      this.log(`✓ Page created: ${page.id}`)
      this.log(`  URL: ${page.url}`)
      process.exit(0)

    } catch (error) {
      this.outputError(error, flags)
    }
  }
}
```

### Pattern 4: Search Command

**Use Case:** Commands with complex filtering and search

```typescript
import { BaseCommand } from '../../base-command'
import { Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'

export default class Search extends BaseCommand {
  static description = 'Search across workspace'
  static flags = {
    query: Flags.string({ char: 'q', description: 'Search query' }),
    filter: Flags.string({ char: 'f', options: ['page', 'database'] }),
    ...ux.table.flags(),
    ...AutomationFlags,
    ...OutputFormatFlags,
  }

  async run(): Promise<void> {
    const { flags } = await this.parse(Search)

    try {
      const results = await notion.client.search({
        query: flags.query,
        filter: flags.filter ? { value: flags.filter, property: 'object' } : undefined,
      })

      // Envelope output with search metadata
      if (flags.json || flags['compact-json']) {
        this.outputSuccess(results.results, flags, {
          query: flags.query || 'all',
          filter: flags.filter || 'none',
          total_results: results.results.length,
          has_more: results.has_more,
        })
      }

      // Raw output
      if (flags.raw) {
        this.log(JSON.stringify(results, null, 2))
        process.exit(0)
      }

      // Table output
      const columns = { /* ... */ }
      ux.table(results.results, columns, { ...flags, printLine: this.log.bind(this) })
      process.exit(0)

    } catch (error) {
      this.outputError(error, flags)
    }
  }
}
```

### Pattern 5: Utility Command (No API Call)

**Use Case:** Local commands like config, cache stats, etc.

```typescript
import { BaseCommand } from '../../base-command'
import { Flags } from '@oclif/core'
import { cacheManager } from '../../cache'
import { AutomationFlags } from '../../base-flags'

export default class CacheStats extends BaseCommand {
  static description = 'Show cache statistics'
  static flags = {
    ...AutomationFlags,
  }

  async run(): Promise<void> {
    const { flags } = await this.parse(CacheStats)

    try {
      const stats = cacheManager.getStats()

      // Envelope output
      if (flags.json || flags['compact-json']) {
        this.outputSuccess(stats, flags)
      }

      // Human-readable output
      this.log('Cache Statistics:')
      this.log(`  Size: ${stats.size}/${stats.maxSize}`)
      this.log(`  Hit Rate: ${(stats.hitRate * 100).toFixed(2)}%`)
      this.log(`  Hits: ${stats.hits}`)
      this.log(`  Misses: ${stats.misses}`)
      process.exit(0)

    } catch (error) {
      this.outputError(error, flags)
    }
  }
}
```

## Error Handling Patterns

### Pattern 1: Standard Error Handling

```typescript
try {
  const result = await notion.retrievePage({ page_id: args.page_id })

  if (flags.json || flags['compact-json']) {
    this.outputSuccess(result, flags)
  }

  // ... other output formats ...

} catch (error) {
  // BaseCommand.outputError automatically:
  // 1. Wraps error in NotionCLIError
  // 2. Creates error envelope (if JSON mode)
  // 3. Adds suggestions based on error code
  // 4. Exits with appropriate code (1 or 2)
  this.outputError(error, flags)
}
```

### Pattern 2: Custom Error Context

```typescript
try {
  const result = await notion.queryDataSource(queryParams)

  if (flags.json || flags['compact-json']) {
    this.outputSuccess(result, flags)
  }

} catch (error) {
  // Add custom context to error
  this.outputError(error, flags, {
    database_id: args.database_id,
    filter_used: !!flags.filter,
    page_size: flags.pageSize,
  })
}
```

### Pattern 3: Validation Errors

```typescript
import { NotionCLIError, ErrorCode } from '../../errors'

async run(): Promise<void> {
  const { args, flags } = await this.parse(MyCommand)

  try {
    // Validate input before API call
    if (!args.page_id || args.page_id.length < 32) {
      throw new NotionCLIError(
        ErrorCode.VALIDATION_ERROR,
        'Invalid page ID format',
        { provided: args.page_id, expected: '32-character UUID' }
      )
    }

    const result = await notion.retrievePage({ page_id: args.page_id })

    if (flags.json || flags['compact-json']) {
      this.outputSuccess(result, flags)
    }

  } catch (error) {
    this.outputError(error, flags)
  }
}
```

## Migration Checklist

### Step 1: Update Command Class

- [ ] Change `extends Command` to `extends BaseCommand`
- [ ] Import `BaseCommand` from `'../../base-command'`
- [ ] Keep existing imports for `AutomationFlags` and `OutputFormatFlags`

```diff
- import { Command } from '@oclif/core'
+ import { BaseCommand } from '../../base-command'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'

- export default class PageRetrieve extends Command {
+ export default class PageRetrieve extends BaseCommand {
```

### Step 2: Update JSON Output Handling

- [ ] Replace manual envelope construction with `this.outputSuccess()`
- [ ] Remove `process.exit(0)` after `this.outputSuccess()` (it exits automatically)
- [ ] Ensure JSON check uses `flags.json || flags['compact-json']`

```diff
- if (flags.json) {
-   this.log(JSON.stringify({
-     success: true,
-     data: result,
-     timestamp: new Date().toISOString()
-   }, null, 2))
-   process.exit(0)
- }
+ if (flags.json || flags['compact-json']) {
+   this.outputSuccess(result, flags)
+ }
```

### Step 3: Update Error Handling

- [ ] Replace manual error envelope with `this.outputError()`
- [ ] Remove `process.exit(1)` after `this.outputError()` (it exits automatically)
- [ ] Remove conditional error JSON output

```diff
  } catch (error) {
-   const cliError = wrapNotionError(error)
-   if (flags.json) {
-     this.log(JSON.stringify(cliError.toJSON(), null, 2))
-   } else {
-     this.error(cliError.message)
-   }
-   process.exit(1)
+   this.outputError(error, flags)
  }
```

### Step 4: Add Metadata (Optional)

- [ ] Identify metadata worth including (pagination, query info, etc.)
- [ ] Pass as third argument to `outputSuccess()`

```diff
  if (flags.json || flags['compact-json']) {
-   this.outputSuccess(pages, flags)
+   this.outputSuccess(pages, flags, {
+     total_results: pages.length,
+     page_size: flags.pageSize,
+     has_more: hasMore,
+   })
  }
```

### Step 5: Test

- [ ] Test `--json` flag (pretty envelope)
- [ ] Test `--compact-json` flag (single-line envelope)
- [ ] Test `--raw` flag (bypasses envelope)
- [ ] Test error cases (should show suggestions)
- [ ] Verify exit codes (0, 1, or 2)
- [ ] Check stderr doesn't pollute stdout in JSON mode

## Complete Migration Example

### Before (Original Command)

```typescript
import { Args, Command, Flags, ux } from '@oclif/core'
import * as notion from '../../notion'
import { PageObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import { getPageTitle, outputRawJson } from '../../helper'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'
import { wrapNotionError } from '../../errors'

export default class PageRetrieve extends Command {
  static description = 'Retrieve a page'
  static args = {
    page_id: Args.string({ required: true, description: 'Page ID' }),
  }
  static flags = {
    raw: Flags.boolean({ char: 'r', description: 'output raw json' }),
    ...ux.table.flags(),
    ...AutomationFlags,
    ...OutputFormatFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(PageRetrieve)

    try {
      const res = await notion.retrievePage({ page_id: args.page_id })

      // Handle JSON output for automation
      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          data: res,
          timestamp: new Date().toISOString()
        }, null, 2))
        process.exit(0)
        return
      }

      // Handle raw JSON output
      if (flags.raw) {
        outputRawJson(res)
        process.exit(0)
        return
      }

      // Define columns for table output
      const columns = {
        title: {
          get: (row: PageObjectResponse) => {
            return getPageTitle(row)
          },
        },
        object: {},
        id: {},
        url: {},
      }

      // Handle table output (default)
      const options = {
        printLine: this.log.bind(this),
        ...flags,
      }
      ux.table([res], columns, options)

    } catch (error) {
      const cliError = wrapNotionError(error)
      if (flags.json) {
        this.log(JSON.stringify(cliError.toJSON(), null, 2))
      } else {
        this.error(cliError.message)
      }
      process.exit(1)
    }
  }
}
```

### After (With Envelope Support)

```typescript
import { Args, Flags, ux } from '@oclif/core'
import { BaseCommand } from '../../base-command'
import * as notion from '../../notion'
import { PageObjectResponse } from '@notionhq/client/build/src/api-endpoints'
import { getPageTitle, outputRawJson } from '../../helper'
import { AutomationFlags, OutputFormatFlags } from '../../base-flags'

export default class PageRetrieve extends BaseCommand {
  static description = 'Retrieve a page'
  static args = {
    page_id: Args.string({ required: true, description: 'Page ID' }),
  }
  static flags = {
    raw: Flags.boolean({ char: 'r', description: 'output raw json' }),
    ...ux.table.flags(),
    ...AutomationFlags,
    ...OutputFormatFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(PageRetrieve)

    try {
      const res = await notion.retrievePage({ page_id: args.page_id })

      // Handle JSON output with envelope (auto-exits)
      if (flags.json || flags['compact-json']) {
        this.outputSuccess(res, flags)
      }

      // Handle raw JSON output (bypasses envelope)
      if (flags.raw) {
        outputRawJson(res)
        process.exit(0)
        return
      }

      // Define columns for table output
      const columns = {
        title: {
          get: (row: PageObjectResponse) => {
            return getPageTitle(row)
          },
        },
        object: {},
        id: {},
        url: {},
      }

      // Handle table output (default)
      const options = {
        printLine: this.log.bind(this),
        ...flags,
      }
      ux.table([res], columns, options)
      process.exit(0)

    } catch (error) {
      // Automatic envelope wrapping and error handling (auto-exits)
      this.outputError(error, flags)
    }
  }
}
```

### Changes Summary

1. **Import change**: `Command` → `BaseCommand`
2. **Class extension**: `extends Command` → `extends BaseCommand`
3. **JSON output**: Manual envelope → `this.outputSuccess()`
4. **Error handling**: Manual wrapping → `this.outputError()`
5. **Compact JSON**: Added support via `flags['compact-json']` check
6. **Exit calls**: Removed after `outputSuccess/outputError` (auto-exit)

## Phased Migration Plan

### Phase 1: Core Commands (Week 1)

**Priority: HIGH** - Most commonly used commands

- [ ] `src/commands/page/retrieve.ts`
- [ ] `src/commands/page/create.ts`
- [ ] `src/commands/page/update.ts`
- [ ] `src/commands/db/retrieve.ts`
- [ ] `src/commands/db/query.ts`
- [ ] `src/commands/db/create.ts`
- [ ] `src/commands/user/retrieve.ts`
- [ ] `src/commands/user/list.ts`

### Phase 2: Block and Search Commands (Week 2)

**Priority: MEDIUM** - Frequently used by automation

- [ ] `src/commands/block/retrieve.ts`
- [ ] `src/commands/block/retrieve/children.ts`
- [ ] `src/commands/block/append.ts`
- [ ] `src/commands/block/update.ts`
- [ ] `src/commands/block/delete.ts`
- [ ] `src/commands/search.ts`

### Phase 3: Utility Commands (Week 3)

**Priority: LOW** - Less critical but complete coverage

- [ ] `src/commands/sync.ts`
- [ ] `src/commands/list.ts`
- [ ] `src/commands/config/set-token.ts`
- [ ] `src/commands/db/schema.ts`
- [ ] `src/commands/page/retrieve/property_item.ts`

### Phase 4: Testing and Documentation (Week 4)

**Priority: HIGH** - Ensure quality and usability

- [ ] Write integration tests for all migrated commands
- [ ] Update README with envelope examples
- [ ] Create OUTPUT_FORMATS.md updates
- [ ] Add migration guide to CLAUDE.md
- [ ] Test automation scripts with new envelope format

## Testing Strategy

### Unit Tests

```typescript
// test/envelope.test.ts
import { expect } from 'chai'
import { EnvelopeFormatter } from '../src/envelope'
import { NotionCLIError, ErrorCode } from '../src/errors'

describe('EnvelopeFormatter', () => {
  describe('wrapSuccess', () => {
    it('should create success envelope with data', () => {
      const formatter = new EnvelopeFormatter('test command', '1.0.0')
      const data = { id: 'abc-123', object: 'page' }
      const envelope = formatter.wrapSuccess(data)

      expect(envelope.success).to.be.true
      expect(envelope.data).to.deep.equal(data)
      expect(envelope.metadata.command).to.equal('test command')
      expect(envelope.metadata.version).to.equal('1.0.0')
      expect(envelope.metadata.timestamp).to.be.a('string')
      expect(envelope.metadata.execution_time_ms).to.be.a('number')
    })

    it('should include additional metadata', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const envelope = formatter.wrapSuccess({}, {
        page_size: 100,
        has_more: false,
      })

      expect(envelope.metadata).to.have.property('page_size', 100)
      expect(envelope.metadata).to.have.property('has_more', false)
    })

    it('should track execution time accurately', (done) => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')

      setTimeout(() => {
        const envelope = formatter.wrapSuccess({})
        expect(envelope.metadata.execution_time_ms).to.be.gte(50)
        expect(envelope.metadata.execution_time_ms).to.be.lte(150)
        done()
      }, 100)
    })
  })

  describe('wrapError', () => {
    it('should create error envelope from NotionCLIError', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(
        ErrorCode.NOT_FOUND,
        'Resource not found',
        { resourceId: 'abc-123' }
      )
      const envelope = formatter.wrapError(error)

      expect(envelope.success).to.be.false
      expect(envelope.error.code).to.equal('NOT_FOUND')
      expect(envelope.error.message).to.equal('Resource not found')
      expect(envelope.error.details).to.have.property('resourceId', 'abc-123')
      expect(envelope.error.suggestions).to.be.an('array')
    })

    it('should generate suggestions based on error code', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.UNAUTHORIZED, 'Auth failed')
      const envelope = formatter.wrapError(error)

      expect(envelope.error.suggestions).to.include.match(/NOTION_TOKEN/)
    })

    it('should handle standard Error objects', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new Error('Something went wrong')
      const envelope = formatter.wrapError(error)

      expect(envelope.success).to.be.false
      expect(envelope.error.code).to.equal('UNKNOWN')
      expect(envelope.error.message).to.equal('Something went wrong')
    })
  })

  describe('getExitCode', () => {
    it('should return 0 for success', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const envelope = formatter.wrapSuccess({})

      expect(formatter.getExitCode(envelope)).to.equal(0)
    })

    it('should return 2 for validation errors', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.VALIDATION_ERROR, 'Invalid input')
      const envelope = formatter.wrapError(error)

      expect(formatter.getExitCode(envelope)).to.equal(2)
    })

    it('should return 1 for API errors', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.NOT_FOUND, 'Not found')
      const envelope = formatter.wrapError(error)

      expect(formatter.getExitCode(envelope)).to.equal(1)
    })
  })
})
```

### Integration Tests

```typescript
// test/commands/page/retrieve.test.ts
import { expect, test } from '@oclif/test'

describe('page retrieve with envelope', () => {
  test
    .stdout()
    .command(['page', 'retrieve', 'abc-123', '--json'])
    .it('returns success envelope', (ctx) => {
      const output = JSON.parse(ctx.stdout)

      expect(output.success).to.be.true
      expect(output.data).to.have.property('object', 'page')
      expect(output.metadata).to.have.property('command', 'page retrieve')
      expect(output.metadata).to.have.property('version')
      expect(output.metadata).to.have.property('timestamp')
      expect(output.metadata).to.have.property('execution_time_ms')
    })

  test
    .stdout()
    .command(['page', 'retrieve', 'invalid-id', '--json'])
    .catch((error) => {
      const output = JSON.parse(error.message)
      expect(output.success).to.be.false
      expect(output.error.code).to.be.oneOf(['NOT_FOUND', 'VALIDATION_ERROR'])
      expect(output.error.suggestions).to.be.an('array')
    })
    .it('returns error envelope on failure')

  test
    .stdout()
    .command(['page', 'retrieve', 'abc-123', '--compact-json'])
    .it('returns compact envelope', (ctx) => {
      // Should be single line (no newlines except final)
      const lines = ctx.stdout.trim().split('\n')
      expect(lines.length).to.equal(1)

      const output = JSON.parse(ctx.stdout)
      expect(output.success).to.be.true
    })

  test
    .stdout()
    .command(['page', 'retrieve', 'abc-123', '--raw'])
    .it('bypasses envelope with --raw', (ctx) => {
      const output = JSON.parse(ctx.stdout)

      // Raw mode: no success field, no metadata
      expect(output).to.not.have.property('success')
      expect(output).to.not.have.property('metadata')
      expect(output).to.have.property('object', 'page')
    })
})
```

## Best Practices

### DO

1. **Always extend BaseCommand** for new commands
2. **Use outputSuccess/outputError** for JSON output
3. **Check both --json and --compact-json** flags
4. **Add meaningful metadata** (pagination, query info, operation type)
5. **Write to stderr for diagnostics** using `EnvelopeFormatter.writeDiagnostic()`
6. **Include suggestions in custom errors**
7. **Test all output modes** (json, compact-json, raw, table)
8. **Document envelope structure** in command examples

### DON'T

1. **Don't manually construct envelopes** - use EnvelopeFormatter
2. **Don't call process.exit() after outputSuccess/outputError** - they exit automatically
3. **Don't write logs to stdout in JSON mode** - use stderr
4. **Don't include sensitive data in error details** (tokens, credentials)
5. **Don't wrap errors multiple times** - outputError handles it
6. **Don't forget --raw flag** - it must bypass envelope
7. **Don't hardcode versions or timestamps** - use metadata
8. **Don't skip error suggestions** - they help users

## Troubleshooting

### Issue: Command hangs after outputSuccess()

**Cause:** Code execution continues after `outputSuccess()`

**Fix:** Remove any code after `outputSuccess()` - it never returns

```typescript
// Bad
if (flags.json) {
  this.outputSuccess(data, flags)
  console.log('This will never execute')
}

// Good
if (flags.json || flags['compact-json']) {
  this.outputSuccess(data, flags)
}
```

### Issue: Envelope missing execution_time_ms

**Cause:** EnvelopeFormatter not initialized properly

**Fix:** Ensure BaseCommand.init() is called (automatic with BaseCommand)

### Issue: Suggestions not appearing in errors

**Cause:** Error code not recognized by suggestion generator

**Fix:** Use standard ErrorCode enum values or add custom suggestions

```typescript
throw new NotionCLIError(
  ErrorCode.VALIDATION_ERROR,
  'Custom error message',
  { details: 'context' }
)
// Suggestions auto-generated based on VALIDATION_ERROR code
```

### Issue: JSON output polluted with logs

**Cause:** Writing to stdout instead of stderr

**Fix:** Use EnvelopeFormatter.writeDiagnostic() for logs

```typescript
// Bad
console.log('Processing...') // Goes to stdout

// Good
EnvelopeFormatter.writeDiagnostic('Processing...', 'info') // Goes to stderr
```

## FAQ

**Q: Should I migrate all commands at once?**
A: No, follow the phased migration plan. Start with core commands.

**Q: What if a command doesn't use the Notion API?**
A: Still use BaseCommand and envelope for consistency (e.g., cache stats, config).

**Q: Can I add custom metadata fields?**
A: Yes! Pass them as third argument to `outputSuccess()`.

**Q: What about backward compatibility?**
A: The `--raw` flag is preserved for legacy scripts. Document both formats.

**Q: Should I remove the old --json code?**
A: Yes, after migration. BaseCommand handles it automatically.

**Q: How do I test envelope output?**
A: Use `@oclif/test` - see Integration Tests section above.

## Additional Resources

- **Envelope Specification:** `docs/ENVELOPE_SPECIFICATION.md`
- **Base Command Source:** `src/base-command.ts`
- **Envelope Source:** `src/envelope.ts`
- **Error Codes:** `src/errors.ts`
- **Example Commands:** `src/commands/page/retrieve.ts` (after migration)

## Support

For questions or issues:
1. Check this guide and the specification
2. Review example commands in `src/commands/`
3. Run tests to validate implementation
4. Open an issue on GitHub if needed
