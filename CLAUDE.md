# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

notion-cli is a non-interactive command-line interface for Notion's API v5.2.1, optimized for AI agents and automation scripts. Built with oclif framework and TypeScript, targeting Node.js >=18.0.0.

**Key differentiators:**
- Non-interactive: All arguments required, no prompts
- Automation-first: JSON output mode, structured errors, exit codes
- Enhanced reliability: Exponential backoff retry with circuit breaker
- High performance: In-memory caching (up to 100x faster for reads)

## Development Commands

```bash
# Build (compile TypeScript to dist/)
npm run build

# Lint
npm run lint

# Run tests
npm test

# Test in development (use ts-node, no compilation)
./bin/dev [command]

# Test compiled version (requires build first)
./bin/run [command]

# Generate oclif manifest and update README
npm run prepack

# Clean up after pack
npm run postpack
```

## Architecture

### Core Layer Structure

**3-Layer Architecture:**

1. **API Wrapper Layer** (`src/notion.ts`)
   - Thin wrapper around `@notionhq/client` v5.2.1
   - All API calls go through `cachedFetch()` helper
   - Automatic retry + caching integration
   - Exports wrapped functions: `createDb`, `retrieveDataSource`, `fetchAllPagesInDS`, etc.

2. **Infrastructure Layer**
   - `src/cache.ts`: In-memory LRU cache with TTL
   - `src/retry.ts`: Exponential backoff with jitter, circuit breaker
   - `src/errors.ts`: Structured error handling (`NotionCLIError`, `wrapNotionError`)
   - `src/helper.ts`: Output formatters, data transformers

3. **Command Layer** (`src/commands/`)
   - oclif command classes extending `Command`
   - Structure: `{resource}/{action}.ts` (e.g., `db/query.ts`, `page/create.ts`)
   - Nested commands: `{resource}/{action}/{subaction}.ts` (e.g., `block/retrieve/children.ts`)

### Critical Concepts

**Notion API v5.2.1 Terminology Change:**
- OLD: "Database" = table with schema + rows
- NEW: "Database" = container, "Data Source" = table with schema
- Code uses `dataSources` API methods (not `databases`)
- Function names: `fetchAllPagesInDS` (DS = Data Source)

**Caching Strategy:**
```typescript
// Cached reads (via cachedFetch in notion.ts):
- Data sources: 10 min TTL (schemas rarely change)
- Users: 1 hour TTL (very stable)
- Pages: 1 min TTL (frequently updated)
- Blocks: 30 sec TTL (most dynamic)

// Write operations auto-invalidate related cache:
- updateDataSource() → invalidates dataSource cache
- createPage() → invalidates parent database cache
```

**Retry Strategy:**
```typescript
// Automatic retry on:
- 429 (rate limit) - respects Retry-After header
- 500-504 (server errors)
- Network errors (ECONNRESET, ETIMEDOUT)

// NO retry on:
- 400-499 (client errors like auth, validation)
```

**Output Formats:**
- All commands support: `--json`, `--raw`, `--output` (table/csv/yaml)
- New formats: `--markdown`, `--compact-json`, `--pretty`
- Flags in `src/base-flags.ts`: `AutomationFlags`, `OutputFormatFlags`

### Key Files

**`src/notion.ts`** (417 lines)
- Central API wrapper with automatic caching + retry
- `cachedFetch<T>()`: Internal helper combining cache + retry
- All Notion API methods exported as async functions
- Cache invalidation logic embedded in write operations

**`src/cache.ts`** (240 lines)
- `CacheManager` class: singleton instance exported as `cacheManager`
- LRU eviction when `maxSize` exceeded
- TTL per resource type (configurable via env vars)
- Methods: `get()`, `set()`, `invalidate()`, `getStats()`, `getHitRate()`

**`src/retry.ts`** (320 lines)
- `fetchWithRetry()`: Main retry function with exponential backoff
- `CircuitBreaker` class: Prevents cascading failures
- `isRetryableError()`: Error categorization
- `calculateDelay()`: Exponential backoff + jitter formula

**`src/helper.ts`**
- Output formatters: `outputRawJson()`, `outputMarkdownTable()`, `outputCompactJson()`, `outputPrettyTable()`
- Data extractors: `getDataSourceTitle()`, `getPageTitle()`, `getDbTitle()`
- Filter builders for database queries

**`src/base-flags.ts`**
- Reusable flag sets for commands
- `AutomationFlags`: `--json`, `--page-size`, `--retry`, `--timeout`, `--no-cache`
- `OutputFormatFlags`: `--markdown`, `--compact-json`, `--pretty` (mutually exclusive)

**`src/errors.ts`**
- `NotionCLIError`: Structured error class with `toJSON()` for automation
- `wrapNotionError()`: Maps HTTP status codes to semantic error codes
- ErrorCode enum: RATE_LIMITED, NOT_FOUND, UNAUTHORIZED, etc.

## Command Development Pattern

When adding a new command:

1. Create file in `src/commands/{resource}/{action}.ts`
2. Extend `Command` from `@oclif/core`
3. Define static properties:
   ```typescript
   static description = 'Command description'
   static aliases: string[] = ['shortcut']
   static examples = [{ description: '...', command: '...' }]
   static args = { arg_name: Args.string({ required: true }) }
   static flags = {
     ...ux.table.flags(),
     ...AutomationFlags,
     ...OutputFormatFlags,
   }
   ```
4. Implement `run()` method with try-catch
5. Handle output formats: check flags in order (compact-json, markdown, pretty, json, raw, table)
6. Use `wrapNotionError()` for consistent error handling
7. Always `process.exit(0)` on success, `process.exit(1)` on error

## Configuration via Environment Variables

All configurable via `.env` or environment:

**Required:**
- `NOTION_TOKEN`: Notion integration token

**Optional - Debug:**
- `DEBUG=true`: Enable debug logging (shows cache hits/misses, retries)

**Optional - Retry:**
- `NOTION_CLI_MAX_RETRIES`: Max retry attempts (default: 3)
- `NOTION_CLI_BASE_DELAY`: Base delay in ms (default: 1000)
- `NOTION_CLI_MAX_DELAY`: Max delay cap in ms (default: 30000)
- `NOTION_CLI_EXP_BASE`: Exponential base (default: 2)
- `NOTION_CLI_JITTER_FACTOR`: Jitter 0-1 (default: 0.1)

**Optional - Cache:**
- `NOTION_CLI_CACHE_ENABLED`: Enable/disable (default: true)
- `NOTION_CLI_CACHE_MAX_SIZE`: Max entries (default: 1000)
- `NOTION_CLI_CACHE_TTL`: Default TTL in ms (default: 300000)
- `NOTION_CLI_CACHE_DS_TTL`: Data source TTL (default: 600000)
- `NOTION_CLI_CACHE_USER_TTL`: User TTL (default: 3600000)
- `NOTION_CLI_CACHE_PAGE_TTL`: Page TTL (default: 60000)
- `NOTION_CLI_CACHE_BLOCK_TTL`: Block TTL (default: 30000)

## Common Patterns

**Adding a new API method:**
```typescript
// In src/notion.ts
export const newMethod = async (params: SomeParams) => {
  // For reads - use cachedFetch
  return cachedFetch(
    'resourceType',
    resourceId,
    () => client.someApi.method(params)
  )

  // For writes - use enhancedFetchWithRetry + invalidate cache
  const result = await enhancedFetchWithRetry(
    () => client.someApi.method(params),
    { context: 'newMethod' }
  )
  cacheManager.invalidate('resourceType', resourceId)
  return result
}
```

**Output format handling in commands:**
```typescript
// Order matters - check specific formats before defaults
if (flags['compact-json']) {
  outputCompactJson(data)
  process.exit(0)
}
if (flags.markdown) {
  outputMarkdownTable(data, columns)
  process.exit(0)
}
if (flags.pretty) {
  outputPrettyTable(data, columns)
  process.exit(0)
}
if (flags.json) {
  this.log(JSON.stringify({ success: true, data }, null, 2))
  process.exit(0)
}
if (flags.raw) {
  outputRawJson(data)
  process.exit(0)
}
// Default: table
ux.table(data, columns, flags)
```

## Important Notes

- **Non-interactive only**: Never use `prompts` library or interactive input
- **Exit codes matter**: 0 = success, 1 = error (automation relies on this)
- **Cache invalidation**: Write operations MUST invalidate related cache entries
- **Pagination**: Use `fetchAllPagesInDS()` for complete data source queries
- **Error wrapping**: Always use `wrapNotionError()` in command catch blocks
- **Type safety**: Use official `@notionhq/client` types from `/build/src/api-endpoints`

## Testing

Tests use Mocha + Chai. Located in `test/` directory.

To run specific test file:
```bash
npx mocha test/specific-test.test.ts
```

Note: Current test coverage focuses on cache and retry logic (`test/cache-retry.test.ts`). Command integration tests use `@oclif/test` framework.

## Documentation

- `README.md`: User-facing documentation
- `ENHANCEMENTS.md`: Deep dive on caching/retry features
- `OUTPUT_FORMATS.md`: Output format reference
- `.env.example`: Configuration examples
- `docs/`: Internal API reference documents (not user-facing)
