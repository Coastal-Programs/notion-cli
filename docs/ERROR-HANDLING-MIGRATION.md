# Error Handling Migration Guide

**Version**: 1.0.0
**Last Updated**: 2025-10-22
**Status**: Ready for Implementation

Step-by-step guide for migrating existing commands to use the enhanced error handling system.

---

## Table of Contents

1. [Overview](#overview)
2. [Prerequisites](#prerequisites)
3. [Migration Steps](#migration-steps)
4. [Command-by-Command Checklist](#command-by-command-checklist)
5. [Breaking Changes](#breaking-changes)
6. [Testing Strategy](#testing-strategy)
7. [Rollback Plan](#rollback-plan)

---

## Overview

### Why Migrate?

**Current State:**
- Generic error messages ("Resource not found")
- No context or suggestions
- Inconsistent error handling across commands
- Difficult for AI assistants to debug

**After Migration:**
- Context-rich error messages with suggestions
- Consistent error structure across all commands
- JSON output mode for automation
- AI-friendly error codes and messages

### Migration Scope

**Files to Update:**
- `src/commands/**/*.ts` - All command files (25+ files)
- `src/notion.ts` - API wrapper functions
- `src/utils/notion-resolver.ts` - ID resolution (partially done)

**Files to Create:**
- ✅ `src/errors/enhanced-errors.ts` - Main error system
- ✅ `src/errors/index.ts` - Clean exports

**Files to Deprecate (Later):**
- `src/errors.ts` - Legacy error system (keep for compatibility)

---

## Prerequisites

### 1. Verify New Error System

```bash
# Check files exist
ls src/errors/enhanced-errors.ts
ls src/errors/index.ts

# Run TypeScript compiler to check for syntax errors
npm run build

# Run unit tests (if available)
npm test -- test/errors/enhanced-errors.test.ts
```

### 2. Create Test Database

Set up a test Notion workspace with:
- Sample database (shared with integration)
- Sample database (NOT shared with integration)
- Various property types for validation testing

### 3. Backup Current Implementation

```bash
# Create backup branch
git checkout -b backup/pre-error-migration
git push origin backup/pre-error-migration

# Return to main branch
git checkout main
```

---

## Migration Steps

### Step 1: Update Imports (Per Command)

**Before:**
```typescript
import { wrapNotionError } from '../errors'
```

**After:**
```typescript
import {
  handleCliError,
  wrapNotionError,
  NotionCLIErrorFactory,
  ErrorContext
} from '../errors/enhanced-errors'
```

Or use the clean export:
```typescript
import {
  handleCliError,
  wrapNotionError,
  NotionCLIErrorFactory,
  ErrorContext
} from '../errors'
```

---

### Step 2: Add Context Object

**Before:**
```typescript
try {
  const result = await someOperation()
} catch (error) {
  throw wrapNotionError(error)
}
```

**After:**
```typescript
try {
  const result = await someOperation()
} catch (error) {
  const context: ErrorContext = {
    resourceType: 'database',
    attemptedId: args.database_id,
    userInput: args.database_id,
    endpoint: 'dataSources.query'
  }
  throw wrapNotionError(error, context)
}
```

---

### Step 3: Use handleCliError for Top-Level Catches

**Before:**
```typescript
try {
  // Command logic
} catch (error) {
  const cliError = wrapNotionError(error)
  if (flags.json) {
    this.log(JSON.stringify(cliError.toJSON()))
  } else {
    this.error(cliError.message)
  }
  process.exit(1)
}
```

**After:**
```typescript
try {
  // Command logic
} catch (error) {
  const context: ErrorContext = {
    resourceType: 'database',
    attemptedId: args.database_id,
    userInput: args.database_id
  }
  handleCliError(error, flags.json, context)
  // handleCliError never returns - exits process
}
```

---

### Step 4: Add Validation Before API Calls

**Before:**
```typescript
async run() {
  const { args, flags } = await this.parse(MyCommand)

  try {
    // Direct API call without validation
    const result = await client.dataSources.query({
      data_source_id: args.database_id
    })
  } catch (error) {
    // Handle error
  }
}
```

**After:**
```typescript
async run() {
  const { args, flags } = await this.parse(MyCommand)

  try {
    // Validate ID format BEFORE API call
    const cleanId = this.validateNotionId(args.database_id)

    // Now make API call with clean ID
    const result = await client.dataSources.query({
      data_source_id: cleanId
    })
  } catch (error) {
    // Handle error with context
    handleCliError(error, flags.json, {
      resourceType: 'database',
      attemptedId: args.database_id,
      userInput: args.database_id
    })
  }
}

private validateNotionId(input: string): string {
  const cleaned = input.replace(/-/g, '')
  if (!/^[a-f0-9]{32}$/i.test(cleaned)) {
    throw NotionCLIErrorFactory.invalidIdFormat(input, 'database')
  }
  return cleaned
}
```

---

### Step 5: Add Property Validation (Database Commands)

**Before:**
```typescript
// No validation - let API fail
queryParams.sorts = [{
  property: flags.sortProperty,
  direction: 'ascending'
}]
```

**After:**
```typescript
// Validate property exists before query
if (flags.sortProperty) {
  await this.validateProperty(databaseId, flags.sortProperty)

  queryParams.sorts = [{
    property: flags.sortProperty,
    direction: 'ascending'
  }]
}

private async validateProperty(dbId: string, propName: string): Promise<void> {
  const db = await notion.retrieveDb(dbId)
  if (!db.properties?.[propName]) {
    throw NotionCLIErrorFactory.invalidProperty(propName, dbId)
  }
}
```

---

### Step 6: Add JSON Parsing Validation

**Before:**
```typescript
if (flags.rawFilter) {
  const filter = JSON.parse(flags.rawFilter) // Can throw generic error
  queryParams.filter = filter
}
```

**After:**
```typescript
if (flags.rawFilter) {
  try {
    const filter = JSON.parse(flags.rawFilter)
    queryParams.filter = filter
  } catch (parseError) {
    throw NotionCLIErrorFactory.invalidJson(
      flags.rawFilter,
      parseError as Error
    )
  }
}
```

---

### Step 7: Update Tests

**Add to each command test:**
```typescript
describe('Error Handling', () => {
  it('returns proper error for missing token', async () => {
    // Test implementation
  })

  it('returns INVALID_ID_FORMAT for bad ID', async () => {
    // Test implementation
  })

  it('supports JSON error output', async () => {
    // Test implementation
  })
})
```

---

## Command-by-Command Checklist

### High Priority (Week 1)

#### ✅ db query
- [ ] Update imports
- [ ] Add ErrorContext to catch blocks
- [ ] Validate database ID before query
- [ ] Add property validation for sorts
- [ ] Add JSON filter validation
- [ ] Use handleCliError for top-level catch
- [ ] Test human output
- [ ] Test JSON output
- [ ] Add unit tests

#### ✅ db retrieve
- [ ] Update imports
- [ ] Add ErrorContext
- [ ] Validate database ID
- [ ] Use handleCliError
- [ ] Test outputs
- [ ] Add unit tests

#### ✅ page retrieve
- [ ] Update imports
- [ ] Add ErrorContext
- [ ] Validate page ID
- [ ] Use handleCliError
- [ ] Test outputs
- [ ] Add unit tests

#### ✅ list
- [ ] Update imports
- [ ] Add ErrorContext
- [ ] Handle empty cache gracefully
- [ ] Use handleCliError
- [ ] Test outputs
- [ ] Add unit tests

### Medium Priority (Week 2)

#### ✅ db create
- [ ] Update imports
- [ ] Add ErrorContext
- [ ] Validate parent page ID
- [ ] Validate property definitions
- [ ] Use handleCliError
- [ ] Test outputs
- [ ] Add unit tests

#### ✅ db update
- [ ] Update imports
- [ ] Add ErrorContext
- [ ] Validate database ID
- [ ] Validate property updates
- [ ] Use handleCliError
- [ ] Test outputs
- [ ] Add unit tests

#### ✅ page create
- [ ] Update imports
- [ ] Add ErrorContext
- [ ] Validate database ID
- [ ] Validate property values
- [ ] Validate JSON content
- [ ] Use handleCliError
- [ ] Test outputs
- [ ] Add unit tests

#### ✅ page update
- [ ] Update imports
- [ ] Add ErrorContext
- [ ] Validate page ID
- [ ] Validate property updates
- [ ] Use handleCliError
- [ ] Test outputs
- [ ] Add unit tests

#### ✅ block append
- [ ] Update imports
- [ ] Add ErrorContext
- [ ] Validate parent block/page ID
- [ ] Validate block content JSON
- [ ] Use handleCliError
- [ ] Test outputs
- [ ] Add unit tests

#### ✅ block update
- [ ] Update imports
- [ ] Add ErrorContext
- [ ] Validate block ID
- [ ] Validate update content
- [ ] Use handleCliError
- [ ] Test outputs
- [ ] Add unit tests

### Low Priority (Week 3-4)

#### ✅ sync
- [ ] Update error handling
- [ ] Add context to API errors
- [ ] Test failure scenarios

#### ✅ config set-token
- [ ] Update error handling
- [ ] Validate token format
- [ ] Test error cases

#### ✅ db schema
- [ ] Update error handling
- [ ] Validate database ID
- [ ] Test error outputs

#### ✅ Other commands
- [ ] Audit all remaining commands
- [ ] Apply consistent error handling
- [ ] Test error scenarios

---

## Breaking Changes

### For End Users

**None** - This is a backward-compatible enhancement.

**What Changes:**
- Error messages become more detailed
- New suggestions appear in errors
- JSON error format becomes more structured

**What Stays Same:**
- Command syntax
- Exit codes
- Core functionality

### For Developers

**Import Paths:**
```typescript
// Old (still works)
import { wrapNotionError } from '../errors'

// New (recommended)
import { wrapNotionError } from '../errors/enhanced-errors'

// Clean (best)
import { wrapNotionError } from '../errors'
```

**Error Structure:**
```typescript
// Old - Basic structure
{
  success: false,
  error: {
    code: 'NOT_FOUND',
    message: 'Resource not found'
  }
}

// New - Rich structure
{
  success: false,
  error: {
    code: 'DATABASE_NOT_FOUND',
    message: 'Database not found: test-id',
    suggestions: [...],
    context: {...},
    timestamp: '2025-10-22T10:30:00.000Z'
  }
}
```

---

## Testing Strategy

### Unit Tests (Per Command)

```bash
# Test error factories
npm test -- test/errors/enhanced-errors.test.ts

# Test individual command
npm test -- test/commands/db/query.test.ts

# Test all commands
npm test
```

### Integration Tests

```bash
# Test CLI with real scenarios
./test/integration/error-handling.sh

# Test JSON output
notion-cli db query invalid-id --json

# Test human output
notion-cli db query invalid-id
```

### Manual Testing Checklist

For each command:
- [ ] Run with invalid ID → expect INVALID_ID_FORMAT
- [ ] Run with unshared database → expect INTEGRATION_NOT_SHARED
- [ ] Run without token → expect TOKEN_MISSING
- [ ] Run with invalid JSON → expect INVALID_JSON
- [ ] Run with invalid property → expect INVALID_PROPERTY
- [ ] Run with --json flag → expect JSON output
- [ ] Run without --json flag → expect human output

---

## Rollback Plan

### If Migration Fails

1. **Identify Issue**
   ```bash
   # Check what broke
   npm test
   notion-cli db query [test-id]
   ```

2. **Quick Rollback**
   ```bash
   # Revert to backup branch
   git checkout backup/pre-error-migration

   # Or revert specific commits
   git revert [commit-hash]
   ```

3. **Fix and Re-attempt**
   ```bash
   # Fix the issue
   git checkout main
   # Make fixes
   git commit -m "fix: error handling issue"
   ```

### Emergency Bypass

If enhanced errors are causing issues, temporarily bypass:

```typescript
// In src/errors/index.ts
export {
  ErrorCode as NotionCLIErrorCode,
  NotionCLIError,
  wrapNotionError as legacyWrapNotionError
} from '../errors'

// In commands, temporarily use:
import { legacyWrapNotionError as wrapNotionError } from '../errors'
```

---

## Progress Tracking

### Week 1: Core Commands
- [ ] db query
- [ ] db retrieve
- [ ] page retrieve
- [ ] list
- [ ] Integration tests

### Week 2: Write Commands
- [ ] db create
- [ ] db update
- [ ] page create
- [ ] page update
- [ ] block append
- [ ] block update
- [ ] Integration tests

### Week 3: Utility Commands
- [ ] sync
- [ ] config set-token
- [ ] db schema
- [ ] Other minor commands
- [ ] Documentation updates

### Week 4: Polish
- [ ] Full test coverage
- [ ] User acceptance testing
- [ ] Performance testing
- [ ] Documentation review
- [ ] Release preparation

---

## Success Metrics

### Quantitative
- [ ] 100% of commands use enhanced error system
- [ ] 0 breaking changes for end users
- [ ] >90% test coverage for error scenarios
- [ ] <20ms overhead per error (target: <10ms)

### Qualitative
- [ ] AI assistants can debug errors independently
- [ ] Users report clearer error messages
- [ ] Reduced support requests for common errors
- [ ] Developers find error handling easy to use

---

## Common Pitfalls

### Pitfall 1: Forgetting Context

❌ **Bad:**
```typescript
throw wrapNotionError(error)
```

✅ **Good:**
```typescript
throw wrapNotionError(error, {
  resourceType: 'database',
  attemptedId: databaseId
})
```

### Pitfall 2: Not Handling JSON Output

❌ **Bad:**
```typescript
console.error(error.message)
```

✅ **Good:**
```typescript
handleCliError(error, flags.json)
```

### Pitfall 3: Validating After API Call

❌ **Bad:**
```typescript
await client.query(params) // Let API fail
```

✅ **Good:**
```typescript
validateParams(params)     // Fail fast
await client.query(params)
```

### Pitfall 4: Generic Error Codes

❌ **Bad:**
```typescript
throw new NotionCLIError('VALIDATION_ERROR', 'Invalid input')
```

✅ **Good:**
```typescript
throw NotionCLIErrorFactory.invalidIdFormat(input, 'database')
```

---

## Support

### Getting Help

1. **Check Documentation:**
   - [Architecture](./ERROR-HANDLING-ARCHITECTURE.md)
   - [Quick Reference](./ERROR-HANDLING-QUICK-REF.md)
   - [Examples](./ERROR-HANDLING-EXAMPLES.md)

2. **Review Examples:**
   - Look at migrated commands in `src/commands/`
   - Check test files in `test/errors/`

3. **Ask Questions:**
   - Create GitHub issue
   - Tag with `error-handling` label

---

## Conclusion

This migration will significantly improve the CLI's usability for both humans and AI assistants. By following this guide step-by-step and testing thoroughly, we can migrate all commands without breaking changes.

**Remember:**
- Take it one command at a time
- Test each command before moving to the next
- Use the provided examples as templates
- Don't hesitate to ask for help

---

**Document Version**: 1.0.0
**Last Updated**: 2025-10-22
**Status**: Ready for Use
**Maintainer**: Backend Architecture Team
