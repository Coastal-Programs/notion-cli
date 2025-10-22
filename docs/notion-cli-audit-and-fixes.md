# Notion CLI Audit & Fixes Required
## Notion API v5.2.1 (API Version 2025-09-03)

**Generated**: 2025-10-22
**Purpose**: Comprehensive audit findings and required fixes for notion-cli-fresh

---

## Table of Contents
1. [Critical Terminology Changes](#critical-terminology-changes)
2. [Critical Fixes Required](#critical-fixes-required)
3. [High Priority Fixes](#high-priority-fixes)
4. [Medium Priority Fixes](#medium-priority-fixes)
5. [Security & Best Practices](#security--best-practices)
6. [Reference Documentation](#reference-documentation)

---

## Critical Terminology Changes

### Database vs Data Source Paradigm (API v2025-09-03)

**OLD MODEL (pre-2025-09-03)**:
- Database = Table with schema + rows
- One database = One table

**NEW MODEL (v2025-09-03+)**:
- **Database** = Container that can hold multiple data sources
- **Data Source** = Individual table with schema (properties) and content (pages/rows)
- One database can contain multiple independent data sources

### Key Changes:
- Parameter change: `database_id` → `data_source_id` when creating pages
- Endpoint change: `/v1/databases/{id}/query` → `/v1/data_sources/{id}/query`
- `databases.retrieve()` now returns list of child data sources, NOT schema
- Use `dataSources.retrieve()` to get schema/properties of a specific table

---

## Critical Fixes Required

### 1. ❌ CRITICAL: db/retrieve.ts - Wrong Endpoint

**File**: `src/commands/db/retrieve.ts`
**Line**: 42
**Severity**: CRITICAL - Returns wrong data structure

**Current Code**:
```typescript
const res = await notion.retrieveDb(databaseId)
```

**Problem**:
- In API v2025-09-03, `databases.retrieve()` returns list of child data sources
- Does NOT return the schema/properties users expect
- Will fail or return incomplete data for multi-source databases

**Fix Required**:
```typescript
// Option 1: Retrieve specific data source schema
const res = await client.dataSources.retrieve({
  data_source_id: dataSourceId
})

// Option 2: List all data sources in database
const res = await client.databases.retrieve({
  database_id: databaseId
})
// This returns: { object: 'list', results: [data_source, data_source, ...] }
```

**Recommendation**: Rename command to `data-source:retrieve` and use `dataSources.retrieve()`, OR add flag to toggle between database listing and data source schema retrieval.

---

### 2. ❌ CRITICAL: db/update.ts - Wrong Endpoint

**File**: `src/commands/db/update.ts`
**Lines**: 57-69
**Severity**: CRITICAL - Updates wrong resource

**Current Code**:
```typescript
const updateDbProps: UpdateDatabaseParameters = {
  database_id: databaseId,  // ❌ Wrong parameter name
  // ... other properties
}
const res = await notion.updateDb(updateDbProps)
// Calls client.databases.update() internally
```

**Problem**:
- Uses `database_id` parameter (updates database container metadata)
- Should use `data_source_id` to update table schema/properties
- In v2025-09-03, updating schema requires `dataSources.update()`

**Fix Required**:
```typescript
const updateDataSourceProps: UpdateDataSourceParameters = {
  data_source_id: dataSourceId,  // ✅ Correct parameter
  title: flags.title,
  description: flags.description,
  properties: properties,
  // ... other properties
}
const res = await client.dataSources.update(updateDataSourceProps)
```

**Files to Update**:
- `src/commands/db/update.ts` - Change API call
- `src/notion.ts` - Update `updateDb()` wrapper or create new `updateDataSource()` wrapper

---

### 3. ❌ BREAKING: page/create.ts - Deprecated Parameter

**File**: `src/commands/page/create.ts`
**Lines**: 69-77
**Severity**: CRITICAL - Will fail with v2025-09-03 API

**Current Code**:
```typescript
if (flags.parent_page_id) {
  pageParent = {
    page_id: flags.parent_page_id,
  }
} else {
  pageParent = {
    database_id: flags.parent_db_id,  // ❌ DEPRECATED - Will be rejected
  }
}
```

**Problem**:
- `database_id` parameter no longer accepted in parent object
- Must use `data_source_id` to create page in a data source (table)

**Fix Required**:
```typescript
if (flags.parent_page_id) {
  pageParent = {
    page_id: flags.parent_page_id,
  }
} else {
  pageParent = {
    data_source_id: flags.parent_data_source_id,  // ✅ Correct parameter
  }
}
```

**Additional Changes**:
- Rename flag: `--parent_db_id` → `--parent_data_source_id`
- Update description to mention "data source" not "database"
- Update examples in help text

---

## High Priority Fixes

### 4. ⚠️ HIGH: db/create.ts - Terminology Confusion

**File**: `src/commands/db/create.ts`
**Lines**: 66-72
**Severity**: HIGH - Mixed terminology

**Current Code**:
```typescript
const res = await notion.createDb({
  parent: { page_id: pageId },
  initial_data_source: {  // ✅ Uses new parameter name
    properties: {
      Name: {
        title: {},
      },
    },
  },
})
```

**Problem**:
- Function name `createDb()` is legacy terminology
- Parameter `initial_data_source` is correct for v2025-09-03
- Internally calls `client.databases.create()` which is still valid but confusing

**Fix Required**:
1. Rename command: `db:create` → `data-source:create` OR keep both
2. Update help text to clarify: "Creates a new database with an initial data source"
3. Consider adding `--title` flag for database title (separate from data source title)

**Status**: PARTIAL - Uses correct API but misleading naming

---

### 5. ✅ CORRECT: db/query.ts - Already Fixed

**File**: `src/commands/db/query.ts`
**Line**: 165
**Severity**: NONE - This is the reference implementation

**Current Code**:
```typescript
const res = await client.dataSources.query(queryParams)
```

**Status**: ✅ **CORRECT** - This is the ONLY db command using the new API properly

**Note**: This should be the reference for how other commands should be updated

---

## Medium Priority Fixes

### 6. Typo: page/create.ts - Wrong Flag Name

**File**: `src/commands/page/create.ts`
**Line**: 79
**Severity**: MEDIUM - Code won't work

**Current Code**:
```typescript
if (flags.filePath) {  // ❌ Typo - should be file_path
```

**Problem**: Flag is defined as `file_path` (line 49-52) but referenced as `filePath`

**Fix Required**:
```typescript
if (flags.file_path) {  // ✅ Matches flag definition
```

---

### 7. Limited Functionality: page/create.ts - Hardcoded Property

**File**: `src/commands/page/create.ts`
**Lines**: 88-92
**Severity**: MEDIUM - Won't work for custom schemas

**Current Code**:
```typescript
properties: {
  Name: {  // ❌ Hardcoded - assumes "Name" property exists
    title: [
      {
        text: {
          content: flags.name,
        },
      },
    ],
  },
},
```

**Problem**:
- Hardcodes "Name" property
- Will fail if data source uses different title property name
- Should detect title property dynamically or require user to specify

**Fix Required**:
```typescript
// Option 1: Require explicit property name
properties: {
  [flags.title_property || 'Name']: {  // Allow user to specify
    title: [{ text: { content: flags.name } }],
  },
},

// Option 2: Fetch schema first and find title property
const dataSource = await client.dataSources.retrieve({
  data_source_id: flags.parent_data_source_id
})
const titleProp = Object.entries(dataSource.properties)
  .find(([_, prop]) => prop.type === 'title')?.[0]
```

---

### 8. Unsupported Parameter: page/retrieve.ts

**File**: `src/commands/page/retrieve.ts`
**Lines**: 40-43, 70-72
**Severity**: MEDIUM - Parameter doesn't exist in API

**Current Code**:
```typescript
static flags = {
  // ...
  filter_properties: Flags.string({
    description: 'Filter properties',
  }),
}

// Later...
if (flags.filter_properties) {
  retrievePageParams.filter_properties = flags.filter_properties.split(',')
}
```

**Problem**: `filter_properties` parameter does NOT exist in `pages.retrieve()` endpoint

**Fix Required**: Remove this flag and related logic entirely

---

### 9. Multiple Issues: page/update.ts

**File**: `src/commands/page/update.ts`
**Issues**:
- Line 26: Typo `descriptin` → `description`
- Line 44: Flag name `un_archive` should be `unarchive`
- Lines 52-53: TODO comment - only supports archiving

**Severity**: MEDIUM

**Fix Required**:
```typescript
// Fix typo
description: 'Update a Notion page',  // ✅

// Fix flag name
unarchive: Flags.boolean({  // ✅
  description: 'Unarchive the page',
  default: false,
}),

// Add property update support
if (flags.properties) {
  updatePageParams.properties = JSON.parse(flags.properties)
}
```

---

### 10. Parameter Issue: block/append.ts

**File**: `src/commands/block/append.ts`
**Lines**: 29-32
**Severity**: MEDIUM - CLI syntax doesn't match examples

**Current Code**:
```typescript
static args = {
  children: Args.string({
    description: 'Block children (JSON)',
    required: true,
  }),
  after: Args.string({
    description: 'Block id to append after',
  }),
}
```

**Problem**: `children` and `after` defined as positional arguments but should be flags for better UX

**Fix Required**:
```typescript
static flags = {
  block_id: Flags.string({
    char: 'b',
    description: 'Parent block ID',
    required: true,
  }),
  children: Flags.string({
    char: 'c',
    description: 'Block children (JSON)',
    required: true,
  }),
  after: Flags.string({
    char: 'a',
    description: 'Block ID to append after',
  }),
}
```

**New Usage**:
```bash
notion-cli block:append -b <block_id> -c '[{"paragraph":{"rich_text":[{"text":{"content":"Hello"}}]}}]'
```

---

### 11. Multiple Issues: block/update.ts

**File**: `src/commands/block/update.ts`
**Issues**:
- Line 26: Typo `descriptin` → `description`
- Line 38: TODO - only supports archiving
- Missing: Content updates, color changes, block conversion

**Severity**: MEDIUM - Very limited functionality

**Fix Required**:
```typescript
static flags = {
  block_id: Flags.string({
    description: 'Block ID to update',
    required: true,
  }),
  archived: Flags.boolean({
    description: 'Archive the block',
  }),
  // Add support for content updates
  content: Flags.string({
    description: 'Updated block content (JSON)',
  }),
  color: Flags.string({
    description: 'Block color',
    options: ['default', 'gray', 'brown', 'orange', 'yellow', 'green', 'blue', 'purple', 'pink', 'red'],
  }),
}

async run(): Promise<void> {
  const { flags } = await this.parse(Update)
  const client = getNotionClient()

  const updateParams: any = {
    block_id: flags.block_id,
  }

  if (flags.archived !== undefined) {
    updateParams.archived = flags.archived
  }

  if (flags.content) {
    // Parse and update block content based on type
    const content = JSON.parse(flags.content)
    Object.assign(updateParams, content)
  }

  if (flags.color) {
    // Update color for supported block types
    updateParams.color = flags.color
  }

  const res = await client.blocks.update(updateParams as UpdateBlockParameters)
  console.log(JSON.stringify(res, null, 2))
}
```

---

## Security & Best Practices

### 12. 🔒 SECURITY: Missing .env in .gitignore

**File**: `.gitignore`
**Severity**: CRITICAL SECURITY RISK

**Problem**: `.env` files not excluded from Git

**Risk**: API tokens could be accidentally committed to repository

**Fix Required**:
Add to `.gitignore`:
```gitignore
# Environment variables
.env
.env.local
.env.*.local
.env.development
.env.production

# API tokens and secrets
*.key
*.pem
secrets/
```

---

### 13. Version Requirement: package.json

**File**: `package.json`
**Severity**: MEDIUM - Dependency conflict

**Current Code**:
```json
{
  "engines": {
    "node": ">=12.0.0"  // ❌ Too low
  },
  "dependencies": {
    "@notionhq/client": "^5.2.1"  // Requires Node >=18
  }
}
```

**Problem**: @notionhq/client v5.2.1 requires Node.js >=18.0.0

**Fix Required**:
```json
{
  "engines": {
    "node": ">=18.0.0"  // ✅ Matches SDK requirement
  }
}
```

---

### 14. Missing Error Handling (ALL Commands)

**Files**: All command files in `src/commands/`
**Severity**: HIGH - Poor user experience

**Problem**: No try-catch blocks for API errors

**Common Errors Not Handled**:
- 400 Bad Request (invalid parameters)
- 401 Unauthorized (invalid token)
- 403 Forbidden (insufficient permissions)
- 404 Not Found (resource doesn't exist)
- 429 Too Many Requests (rate limit)
- 500+ Server errors

**Fix Required** (Apply to ALL commands):
```typescript
async run(): Promise<void> {
  try {
    const { flags } = await this.parse(CommandName)
    const client = getNotionClient()

    // ... existing code ...

    const res = await client.someMethod(params)
    console.log(JSON.stringify(res, null, 2))
  } catch (error: any) {
    // Handle Notion API errors
    if (error.code === 'notionhq_client_request_timeout') {
      this.error('Request timed out. Please try again.')
    } else if (error.status === 401) {
      this.error('Authentication failed. Please check your NOTION_TOKEN.')
    } else if (error.status === 403) {
      this.error('Permission denied. Check integration capabilities.')
    } else if (error.status === 404) {
      this.error('Resource not found. Please check the ID.')
    } else if (error.status === 429) {
      const retryAfter = error.headers?.['retry-after'] || '60'
      this.error(`Rate limited. Retry after ${retryAfter} seconds.`)
    } else {
      this.error(`API Error: ${error.message}`, { exit: 1 })
    }
  }
}
```

---

### 15. Missing Rate Limit Handling

**Files**: All commands making multiple API calls
**Severity**: MEDIUM - Could hit rate limits

**Problem**: No rate limit handling (3 requests/second limit)

**Best Practice**:
```typescript
// Add to src/utils/rateLimit.ts
const RATE_LIMIT_MS = 350 // ~3 requests/second with buffer

export async function waitForRateLimit() {
  return new Promise(resolve => setTimeout(resolve, RATE_LIMIT_MS))
}

// Use in commands with loops
for (const item of items) {
  await client.someMethod(item)
  await waitForRateLimit()  // ✅ Prevent rate limit errors
}
```

---

## Reference Documentation

All research findings compiled in:

1. **notion-api-reference-v5.2.1.md** (37KB)
   - Complete API reference
   - All endpoints with request/response schemas
   - Property types, filters, sorting
   - Best practices

2. **notion-api-quick-reference.md** (8KB)
   - Quick lookup guide
   - Common patterns
   - Code examples

3. **notion-api-database-specification.md**
   - Database vs Data Source detailed explanation
   - All CRUD operations
   - Property types with examples
   - Filter and sort specifications

4. **Other specification files**:
   - notion-api-search-users-spec.md
   - notion-api-pages-spec.md
   - notion-api-blocks-spec.md
   - notion-api-comments-spec.md
   - notion-api-files-media-spec.md
   - notion-api-best-practices.md

---

## Priority Summary

### MUST FIX (Before Release):
1. ❌ db/retrieve.ts - Wrong endpoint (CRITICAL)
2. ❌ db/update.ts - Wrong endpoint (CRITICAL)
3. ❌ page/create.ts - Deprecated database_id parameter (BREAKING)
4. 🔒 .gitignore - Add .env files (SECURITY)
5. ✏️ Multiple typos causing broken functionality

### SHOULD FIX (Next Sprint):
6. ⚠️ db/create.ts - Clarify database vs data source terminology
7. ⚠️ page/create.ts - Hardcoded "Name" property assumption
8. ⚠️ page/retrieve.ts - Remove unsupported filter_properties
9. ⚠️ block/append.ts - Fix parameter handling
10. ⚠️ block/update.ts - Add content update support
11. ⚠️ package.json - Update Node version requirement

### NICE TO HAVE (Future):
12. Error handling for all commands
13. Rate limit handling
14. Progress indicators for batch operations
15. Retry logic with exponential backoff

---

## Migration Path

For users upgrading from old version:

1. **Command renames** (breaking changes):
   - `--parent_db_id` → `--parent_data_source_id` in page:create
   - Consider: `db:*` → `data-source:*` for consistency

2. **Behavior changes**:
   - `db:retrieve` now returns different data structure
   - May need separate commands for database metadata vs data source schema

3. **New features available**:
   - Multi-source databases support
   - Improved query capabilities
   - Better property types

---

**End of Audit Report**

Next Steps:
1. Review and approve fixes
2. Create feature branch
3. Implement fixes in priority order
4. Update tests
5. Update README with correct terminology
6. Release v2.0.0 with breaking changes noted
