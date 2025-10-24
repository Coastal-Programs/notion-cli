# Caching & Sync Architecture for Natural Language Database Lookups

## Overview

This document describes the persistent caching system for notion-cli that enables natural language database lookups. The system combines file-based persistent storage with in-memory caching to provide fast, AI-friendly database resolution.

## Architecture Goals

1. **Fast Lookups**: Resolve database names to IDs in <10ms from cache
2. **Natural Language**: Support fuzzy matching (e.g., "task db" matches "Tasks Database")
3. **Hybrid Approach**: Cache-first with API fallback for freshness
4. **Zero Config**: Work out-of-the-box with sensible defaults
5. **Resilient**: Handle API failures gracefully with stale cache fallback

## System Components

```
┌─────────────────────────────────────────────────────────────┐
│                     USER INPUT                               │
│  "Tasks DB" | "1fb79d4c..." | "notion.so/xyz" | "Tasks"    │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│              Name Resolution Layer                           │
│  extractNotionId() + resolveDatabase()                      │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
         ┌────────┴────────┐
         │   Is it a URL?  │──Yes──> Extract ID from URL
         └────────┬────────┘
                  │ No
                  ▼
         ┌────────┴────────┐
         │  Is it an ID?   │──Yes──> Return clean ID
         └────────┬────────┘
                  │ No
                  ▼
         ┌────────┴────────┐
         │  Is it a name?  │──Yes──> Search cache → fuzzy match
         └────────┬────────┘              │
                  │                       │ Not found
                  ▼                       ▼
           Return error          Sync cache → Search again
                                          │
                                          │ Still not found
                                          ▼
                                   Search API → Return error
```

## Cache Schema

### Primary Cache File: `~/.notion-cli/databases.json`

```json
{
  "version": "1.0.0",
  "lastSync": "2025-10-22T10:30:00.000Z",
  "syncStatus": {
    "inProgress": false,
    "lastAttempt": "2025-10-22T10:30:00.000Z",
    "lastSuccess": "2025-10-22T10:30:00.000Z",
    "errors": []
  },
  "databases": [
    {
      "id": "1fb79d4c71bb8032b722c82305b63a00",
      "object": "data_source",
      "title": "Tasks Database",
      "titleNormalized": "tasks database",
      "url": "https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00",
      "parent": {
        "type": "workspace",
        "workspace": true
      },
      "created_time": "2025-01-15T09:00:00.000Z",
      "last_edited_time": "2025-10-20T14:30:00.000Z",
      "archived": false,
      "properties": {
        "Name": { "type": "title" },
        "Status": {
          "type": "select",
          "options": ["Not Started", "In Progress", "Done"]
        },
        "Tags": {
          "type": "multi_select",
          "options": ["urgent", "bug", "feature"]
        }
      },
      "propertyNames": ["name", "status", "tags"],
      "aliases": ["tasks", "task db", "todo", "tasks db"]
    },
    {
      "id": "2a8c3d5e71bb8042b833d94316c74b11",
      "object": "data_source",
      "title": "Meeting Notes",
      "titleNormalized": "meeting notes",
      "url": "https://www.notion.so/2a8c3d5e71bb8042b833d94316c74b11",
      "parent": {
        "type": "page_id",
        "page_id": "abc123..."
      },
      "created_time": "2025-02-01T10:00:00.000Z",
      "last_edited_time": "2025-10-21T16:45:00.000Z",
      "archived": false,
      "properties": {
        "Title": { "type": "title" },
        "Date": { "type": "date" },
        "Participants": { "type": "people" }
      },
      "propertyNames": ["title", "date", "participants"],
      "aliases": ["meetings", "meeting", "notes", "meeting db"]
    }
  ],
  "metadata": {
    "totalDatabases": 2,
    "workspaceDatabases": 1,
    "pageDatabases": 1,
    "archivedDatabases": 0
  }
}
```

### Cache Entry Fields

| Field | Type | Description | Usage |
|-------|------|-------------|-------|
| `id` | string | Clean Notion ID (32 hex chars) | Primary key for lookups |
| `object` | string | Always "data_source" | Type validation |
| `title` | string | Original database title | Display to user |
| `titleNormalized` | string | Lowercase, trimmed title | Fast exact matching |
| `url` | string | Full Notion URL | Direct browser access |
| `parent` | object | Parent workspace/page info | Context and organization |
| `created_time` | ISO8601 | Database creation timestamp | Metadata |
| `last_edited_time` | ISO8601 | Last modification timestamp | Freshness tracking |
| `archived` | boolean | Archive status | Filter active databases |
| `properties` | object | Schema with types and options | Schema discovery |
| `propertyNames` | string[] | Normalized property names | Fast property lookup |
| `aliases` | string[] | Auto-generated search terms | Fuzzy matching |

### Alias Generation Rules

Aliases are auto-generated from the title to improve fuzzy matching:

```typescript
function generateAliases(title: string): string[] {
  const aliases = new Set<string>()
  const normalized = title.toLowerCase().trim()

  // Add full title (normalized)
  aliases.add(normalized)

  // Add title without common suffixes
  const withoutSuffixes = normalized
    .replace(/\s+(database|db|table|list|tracker|log)$/i, '')
  if (withoutSuffixes !== normalized) {
    aliases.add(withoutSuffixes)
  }

  // Add title with common suffixes
  aliases.add(`${withoutSuffixes} db`)
  aliases.add(`${withoutSuffixes} database`)

  // Add singular/plural variants
  if (withoutSuffixes.endsWith('s')) {
    aliases.add(withoutSuffixes.slice(0, -1)) // Remove 's'
  } else {
    aliases.add(`${withoutSuffixes}s`) // Add 's'
  }

  // Add acronym if multi-word (e.g., "Meeting Notes" → "mn")
  const words = withoutSuffixes.split(/\s+/)
  if (words.length > 1) {
    const acronym = words.map(w => w[0]).join('')
    if (acronym.length >= 2) {
      aliases.add(acronym)
    }
  }

  return Array.from(aliases)
}

// Example: "Tasks Database"
// Generates: ["tasks database", "tasks", "task", "tasks db", "task db", "td"]

// Example: "Meeting Notes"
// Generates: ["meeting notes", "meeting note", "meeting", "meeting db", "mn"]
```

## Sync Logic

### Sync Strategy

**Sync Types:**
1. **Full Sync**: Fetch all databases from API, replace cache
2. **Incremental Sync**: Only fetch databases modified since last sync (future enhancement)
3. **On-Demand Sync**: Triggered when name resolution fails

**Sync Triggers:**
1. **Manual**: User runs `notion-cli db sync`
2. **Auto (Stale Cache)**: Cache older than TTL (default: 1 hour)
3. **Auto (Not Found)**: Name resolution fails in cache

### Sync Workflow

```
┌─────────────────────────────────────────────────────────────┐
│                     Sync Triggered                           │
│  (manual / stale cache / not found)                         │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
         ┌────────────────┐
         │ Check Lock File│──Locked──> Return (sync in progress)
         └────────┬───────┘
                  │ Not Locked
                  ▼
         ┌────────────────┐
         │  Create Lock   │
         └────────┬───────┘
                  │
                  ▼
         ┌────────────────────────────────────────┐
         │   Search API for Data Sources          │
         │   Filter: object = 'data_source'       │
         │   Pagination: Handle all pages         │
         └────────┬───────────────────────────────┘
                  │
                  ▼
┌─────────────────────────────────────────────────────────────┐
│              For Each Data Source Found                      │
├─────────────────────────────────────────────────────────────┤
│  1. Retrieve full data source details (API call)            │
│  2. Extract metadata (id, title, url, parent, etc)          │
│  3. Extract properties schema                                │
│  4. Generate aliases from title                              │
│  5. Build cache entry                                        │
└─────────────────┬───────────────────────────────────────────┘
                  │
                  ▼
         ┌────────────────┐
         │  Build Index   │  (titleNormalized → id mapping)
         └────────┬───────┘
                  │
                  ▼
         ┌────────────────┐
         │ Write to File  │  (atomic write with .tmp + rename)
         └────────┬───────┘
                  │
                  ▼
         ┌────────────────┐
         │  Release Lock  │
         └────────┬───────┘
                  │
                  ▼
         ┌────────────────┐
         │   Success!     │
         └────────────────┘
```

### Sync Implementation Details

#### Pagination Handling

```typescript
async function fetchAllDatabases(): Promise<DataSourceObjectResponse[]> {
  const databases: DataSourceObjectResponse[] = []
  let cursor: string | undefined = undefined

  while (true) {
    const response = await enhancedFetchWithRetry(
      () => client.search({
        filter: {
          value: 'data_source',
          property: 'object',
        },
        start_cursor: cursor,
        page_size: 100, // Max allowed by API
      }),
      {
        context: 'sync:fetchAllDatabases',
        config: { maxRetries: 5 } // Higher retries for sync
      }
    )

    databases.push(...response.results as DataSourceObjectResponse[])

    if (!response.has_more || !response.next_cursor) {
      break
    }

    cursor = response.next_cursor
  }

  return databases
}
```

#### Atomic File Write

Prevent corruption if process crashes during write:

```typescript
async function writeCache(cache: DatabaseCache): Promise<void> {
  const cachePath = getCachePath() // ~/.notion-cli/databases.json
  const tmpPath = `${cachePath}.tmp`

  // Write to temporary file
  await fs.writeFile(
    tmpPath,
    JSON.stringify(cache, null, 2),
    'utf8'
  )

  // Atomic rename (replaces old file)
  await fs.rename(tmpPath, cachePath)
}
```

#### Sync Lock File

Prevent concurrent sync operations:

```typescript
async function acquireSyncLock(): Promise<boolean> {
  const lockPath = path.join(getCacheDir(), '.sync.lock')

  try {
    // Create lock file (fails if exists)
    await fs.writeFile(lockPath, Date.now().toString(), { flag: 'wx' })
    return true
  } catch (error) {
    if (error.code === 'EEXIST') {
      // Check if lock is stale (>5 minutes old)
      const lockContent = await fs.readFile(lockPath, 'utf8')
      const lockTime = parseInt(lockContent, 10)
      const isStale = Date.now() - lockTime > 5 * 60 * 1000

      if (isStale) {
        // Remove stale lock and retry
        await fs.unlink(lockPath)
        return acquireSyncLock()
      }

      return false // Sync in progress
    }
    throw error
  }
}

async function releaseSyncLock(): Promise<void> {
  const lockPath = path.join(getCacheDir(), '.sync.lock')
  await fs.unlink(lockPath).catch(() => {}) // Ignore errors
}
```

## Name Resolution Algorithm

### Resolution Flow

```typescript
/**
 * Resolve database name/ID/URL to clean Notion ID
 *
 * @param input - Database name, ID, or URL
 * @param options - Resolution options
 * @returns Clean Notion ID
 */
async function resolveDatabase(
  input: string,
  options: {
    fuzzyThreshold?: number // Default: 0.7 (70% match)
    syncIfNotFound?: boolean // Default: true
    includeArchived?: boolean // Default: false
  } = {}
): Promise<string> {
  const {
    fuzzyThreshold = 0.7,
    syncIfNotFound = true,
    includeArchived = false,
  } = options

  // Step 1: Is it a URL? Extract ID
  if (isNotionUrl(input)) {
    return extractNotionId(input)
  }

  // Step 2: Is it a clean ID? (32 hex chars)
  if (/^[a-f0-9]{32}$/i.test(input.replace(/-/g, ''))) {
    return extractNotionId(input)
  }

  // Step 3: Treat as name - search cache
  const cache = await loadCache()
  const normalized = input.toLowerCase().trim()

  // 3a. Exact match on title
  let match = cache.databases.find(
    db => db.titleNormalized === normalized &&
          (includeArchived || !db.archived)
  )
  if (match) return match.id

  // 3b. Exact match on alias
  match = cache.databases.find(
    db => db.aliases.includes(normalized) &&
          (includeArchived || !db.archived)
  )
  if (match) return match.id

  // 3c. Fuzzy match
  const fuzzyMatches = cache.databases
    .filter(db => includeArchived || !db.archived)
    .map(db => ({
      db,
      score: Math.max(
        fuzzyScore(normalized, db.titleNormalized),
        ...db.aliases.map(alias => fuzzyScore(normalized, alias))
      )
    }))
    .filter(({ score }) => score >= fuzzyThreshold)
    .sort((a, b) => b.score - a.score)

  if (fuzzyMatches.length > 0) {
    return fuzzyMatches[0].db.id
  }

  // 3d. Not found in cache - sync and retry
  if (syncIfNotFound) {
    await syncCache()
    return resolveDatabase(input, {
      ...options,
      syncIfNotFound: false // Prevent infinite recursion
    })
  }

  // 3e. Still not found - search API directly
  const apiResult = await searchDatabaseByName(input)
  if (apiResult) {
    return apiResult.id
  }

  // 3f. Give up
  throw new Error(
    `Could not find database matching: "${input}"\n\n` +
    `Tried:\n` +
    `  - URL extraction\n` +
    `  - ID validation\n` +
    `  - Cache lookup (exact, alias, fuzzy)\n` +
    `  - Cache sync\n` +
    `  - API search\n\n` +
    `Suggestions:\n` +
    `  - Verify the database exists and is shared with your integration\n` +
    `  - Try using the full database ID or URL\n` +
    `  - Run 'notion-cli db sync' to refresh the cache`
  )
}
```

### Fuzzy Matching Algorithm

Using Levenshtein distance with normalization:

```typescript
/**
 * Calculate fuzzy match score (0.0 to 1.0)
 *
 * Uses normalized Levenshtein distance:
 * score = 1 - (distance / maxLength)
 */
function fuzzyScore(query: string, target: string): number {
  const distance = levenshtein(query, target)
  const maxLength = Math.max(query.length, target.length)

  if (maxLength === 0) return 1.0

  return 1 - (distance / maxLength)
}

/**
 * Levenshtein distance (edit distance)
 * Measures minimum number of edits to transform query into target
 */
function levenshtein(a: string, b: string): number {
  const matrix: number[][] = []

  // Initialize matrix
  for (let i = 0; i <= b.length; i++) {
    matrix[i] = [i]
  }
  for (let j = 0; j <= a.length; j++) {
    matrix[0][j] = j
  }

  // Fill matrix
  for (let i = 1; i <= b.length; i++) {
    for (let j = 1; j <= a.length; j++) {
      if (b.charAt(i - 1) === a.charAt(j - 1)) {
        matrix[i][j] = matrix[i - 1][j - 1]
      } else {
        matrix[i][j] = Math.min(
          matrix[i - 1][j - 1] + 1, // substitution
          matrix[i][j - 1] + 1,     // insertion
          matrix[i - 1][j] + 1      // deletion
        )
      }
    }
  }

  return matrix[b.length][a.length]
}

// Examples:
// fuzzyScore("tasks", "tasks database") = 0.73
// fuzzyScore("task", "tasks") = 0.80
// fuzzyScore("meeting", "meetings") = 0.89
// fuzzyScore("td", "tasks database") = 0.14 (too low)
```

## Cache Invalidation Strategy

### When to Invalidate

| Event | Action | Reason |
|-------|--------|--------|
| Database created | Add to cache | Keep cache fresh |
| Database updated | Update cache entry | Title/schema may change |
| Database archived | Update archived flag | Filter from searches |
| Database deleted | Remove from cache | No longer exists |
| Manual sync | Full refresh | User-initiated |
| Cache TTL expired | Stale until sync | Prevent serving old data |
| Name not found | Trigger sync | May be new database |

### Cache TTL (Time To Live)

```typescript
const CACHE_TTL = {
  default: 60 * 60 * 1000, // 1 hour (configurable)
  max: 24 * 60 * 60 * 1000, // 24 hours (absolute limit)
}

function isCacheStale(cache: DatabaseCache): boolean {
  if (!cache.lastSync) return true

  const age = Date.now() - new Date(cache.lastSync).getTime()
  const ttl = parseInt(
    process.env.NOTION_CLI_DB_CACHE_TTL || String(CACHE_TTL.default),
    10
  )

  return age > ttl
}
```

### Incremental Updates

On database mutation operations, update cache immediately:

```typescript
export async function createDb(
  dbProps: CreateDatabaseParameters
): Promise<CreateDatabaseResponse> {
  const result = await enhancedFetchWithRetry(
    () => client.databases.create(dbProps),
    { context: 'createDb' }
  )

  // Invalidate search cache (legacy in-memory)
  cacheManager.invalidate('search')

  // Update persistent database cache
  await updateDatabaseCache(result)

  return result
}

async function updateDatabaseCache(
  database: GetDatabaseResponse | CreateDatabaseResponse
): Promise<void> {
  const cache = await loadCache()

  // Remove old entry if exists
  cache.databases = cache.databases.filter(db => db.id !== database.id)

  // Add new entry
  const entry = await buildCacheEntry(database)
  cache.databases.push(entry)

  // Update metadata
  cache.metadata.totalDatabases = cache.databases.length
  cache.lastSync = new Date().toISOString()

  // Write back
  await writeCache(cache)
}
```

## API Rate Limiting Considerations

### Search API Rate Limits

Notion API rate limits (as of 2025):
- **General requests**: 3 requests per second
- **Search endpoint**: Higher cost (counts as multiple requests)
- **Rate limit response**: HTTP 429 with Retry-After header

### Sync Optimization Strategies

1. **Batch Processing**: Fetch 100 databases per search page (max allowed)

2. **Parallel Detail Fetching**: Fetch database details in parallel (with concurrency limit)

```typescript
async function syncWithRateLimit(): Promise<void> {
  const databases = await fetchAllDatabases() // Paginated search

  // Fetch details with concurrency limit
  const CONCURRENCY = 3 // 3 requests per second
  const cacheEntries: CacheEntry[] = []

  for (let i = 0; i < databases.length; i += CONCURRENCY) {
    const batch = databases.slice(i, i + CONCURRENCY)
    const entries = await Promise.all(
      batch.map(db => buildCacheEntryFromSearch(db))
    )
    cacheEntries.push(...entries)

    // Respect rate limit (wait 1 second between batches)
    if (i + CONCURRENCY < databases.length) {
      await sleep(1000)
    }
  }

  // Write cache
  await writeCache({
    version: '1.0.0',
    lastSync: new Date().toISOString(),
    databases: cacheEntries,
    metadata: buildMetadata(cacheEntries),
  })
}
```

3. **Smart Retry**: Use existing retry infrastructure with exponential backoff

4. **Partial Results**: If sync fails midway, cache partial results with warning flag

```typescript
try {
  await syncWithRateLimit()
} catch (error) {
  // Save partial results if we got some data
  if (cacheEntries.length > 0) {
    cache.syncStatus.errors.push({
      timestamp: new Date().toISOString(),
      message: error.message,
      partial: true,
    })
    await writeCache(cache)
  }
  throw error
}
```

## Error Handling Patterns

### Error Hierarchy

```typescript
class DatabaseCacheError extends Error {
  constructor(message: string, public code: string) {
    super(message)
    this.name = 'DatabaseCacheError'
  }
}

class CacheSyncError extends DatabaseCacheError {
  constructor(message: string, public cause?: Error) {
    super(message, 'SYNC_ERROR')
  }
}

class DatabaseNotFoundError extends DatabaseCacheError {
  constructor(query: string) {
    super(
      `Database not found: "${query}"`,
      'DATABASE_NOT_FOUND'
    )
  }
}

class CacheCorruptedError extends DatabaseCacheError {
  constructor(message: string) {
    super(message, 'CACHE_CORRUPTED')
  }
}
```

### Error Recovery Strategies

| Error Type | Recovery Strategy | User Impact |
|------------|-------------------|-------------|
| Cache file missing | Create new cache, trigger sync | First-time setup |
| Cache file corrupted | Backup old cache, create new | Automatic recovery |
| Sync API failure | Use stale cache + warning | Degraded (stale data) |
| Rate limit hit | Exponential backoff + retry | Slow sync |
| Network timeout | Retry with backoff | Slow sync |
| Database not found | Sync + retry → API search → error | User sees error |
| Lock timeout | Force release stale lock | Automatic recovery |

### Graceful Degradation

```typescript
async function loadCache(): Promise<DatabaseCache> {
  try {
    const cache = await readCacheFile()

    // Validate cache structure
    if (!cache.version || !Array.isArray(cache.databases)) {
      throw new CacheCorruptedError('Invalid cache structure')
    }

    // Warn if cache is stale
    if (isCacheStale(cache)) {
      console.warn(
        'Warning: Database cache is stale. ' +
        'Run "notion-cli db sync" to refresh.'
      )
    }

    return cache
  } catch (error) {
    if (error.code === 'ENOENT') {
      // Cache doesn't exist - create empty cache
      return createEmptyCache()
    }

    if (error instanceof CacheCorruptedError) {
      // Backup corrupted cache
      await backupCorruptedCache()

      // Create new cache
      return createEmptyCache()
    }

    throw error
  }
}
```

## Performance Characteristics

### Cache Hit Performance

| Operation | Cold Cache | Warm Cache | Speedup |
|-----------|-----------|------------|---------|
| Exact name match | API call (~200ms) | Hash lookup (~1ms) | 200x |
| Fuzzy name match | API search (~300ms) | Fuzzy scan (~5ms) | 60x |
| ID validation | API call (~200ms) | Regex check (<1ms) | 200x+ |
| Schema lookup | API call (~200ms) | JSON read (~2ms) | 100x |

### Cache Size Estimates

| Workspace Size | Cache Size | Load Time |
|----------------|------------|-----------|
| 10 databases | ~15 KB | <5ms |
| 100 databases | ~150 KB | <20ms |
| 1000 databases | ~1.5 MB | <100ms |
| 10000 databases | ~15 MB | ~1s |

**Note**: Large workspaces (1000+ databases) should consider database organization and filtering strategies.

### Sync Performance

| Workspace Size | Full Sync Time | Incremental Sync |
|----------------|----------------|------------------|
| 10 databases | ~5s | ~1s |
| 100 databases | ~40s | ~5s |
| 1000 databases | ~6min | ~30s |

**Rate limiting**: 3 requests/sec × 1 search + 1 retrieve per DB

## Integration with Existing Systems

### Integration with In-Memory Cache

The persistent database cache complements the existing in-memory cache:

```typescript
// In-memory cache: Fast, temporary, per-process
cacheManager.get('dataSource', databaseId) // Retrieves full API response

// Persistent cache: Slower, permanent, cross-process
resolveDatabase('tasks db') // Resolves name → ID
getDatabaseSchema(databaseId) // Retrieves schema from cache
```

**When to use each:**

| Use Case | In-Memory Cache | Persistent Cache |
|----------|----------------|------------------|
| Full API responses | ✅ | ❌ |
| Database schemas | ✅ (via dataSource) | ✅ |
| Name → ID resolution | ❌ | ✅ |
| Cross-process sharing | ❌ | ✅ |
| Survives restart | ❌ | ✅ |
| Fast lookups (<1ms) | ✅ | ❌ (file I/O) |

### Cache Hierarchy

```
User Query: "tasks database"
     │
     ▼
1. Extract ID if URL/ID format
     │ (not URL/ID)
     ▼
2. Check persistent cache (name → ID)
     │ (ID found)
     ▼
3. Check in-memory cache (ID → full data)
     │ (cache miss)
     ▼
4. Fetch from API + cache in memory
     │
     ▼
Return data
```

## CLI Commands

### New Commands to Implement

```bash
# Sync database cache
notion-cli db sync [--force] [--json]

# List cached databases
notion-cli db list [--filter <pattern>] [--json]

# Clear database cache
notion-cli db cache clear [--json]

# Show cache stats
notion-cli db cache stats [--json]

# Resolve database name to ID (for debugging)
notion-cli db resolve <name> [--json]
```

### Command Examples

```bash
# Initial sync (run once after install)
$ notion-cli db sync
Syncing database cache...
Found 42 databases
Writing cache to ~/.notion-cli/databases.json
Sync complete in 25s

# List databases
$ notion-cli db list
Name                 ID                                URL
Tasks Database       1fb79d4c71bb8032b722c82305b63a00  https://notion.so/...
Meeting Notes        2a8c3d5e71bb8042b833d94316c74b11  https://notion.so/...
...

# Filter databases
$ notion-cli db list --filter "task*"
Name                 ID                                URL
Tasks Database       1fb79d4c71bb8032b722c82305b63a00  https://notion.so/...

# Resolve database name
$ notion-cli db resolve "tasks"
Database: Tasks Database
ID: 1fb79d4c71bb8032b722c82305b63a00
Match: exact (alias)

$ notion-cli db resolve "task db"
Database: Tasks Database
ID: 1fb79d4c71bb8032b722c82305b63a00
Match: fuzzy (score: 0.85)

# Cache stats
$ notion-cli db cache stats
Database Cache Statistics
─────────────────────────
Total databases: 42
Cached: 42
Stale: 0
Last sync: 2025-10-22 10:30:00
Cache age: 5 minutes
Cache size: 125 KB

# Clear cache
$ notion-cli db cache clear
Cache cleared successfully

# Use resolved names in other commands
$ notion-cli db query "tasks" --json
$ notion-cli db schema "meeting notes" --output json
```

## Environment Variables

```bash
# Database cache configuration
export NOTION_CLI_DB_CACHE_TTL=3600000      # Cache TTL (1 hour)
export NOTION_CLI_DB_CACHE_PATH="~/.notion-cli"  # Cache directory
export NOTION_CLI_DB_FUZZY_THRESHOLD=0.7    # Fuzzy match threshold (0.0-1.0)
export NOTION_CLI_DB_SYNC_CONCURRENCY=3     # Parallel requests during sync
export NOTION_CLI_DB_AUTO_SYNC=true         # Auto-sync on cache miss
```

## Future Enhancements

### Phase 2: Page Indexing

Extend cache to include pages for full-text search:

```json
{
  "pages": [
    {
      "id": "abc123",
      "title": "Project Alpha",
      "parent": { "database_id": "1fb79d4c..." },
      "created_time": "2025-01-15T09:00:00.000Z",
      "last_edited_time": "2025-10-20T14:30:00.000Z"
    }
  ]
}
```

### Phase 3: Incremental Sync

Only fetch databases modified since last sync:

```typescript
async function incrementalSync(lastSync: Date): Promise<void> {
  // Search with last_edited_time filter
  const databases = await client.search({
    filter: {
      property: 'object',
      value: 'data_source',
    },
    // Note: Notion API doesn't support time-based filters on search yet
    // This is a future enhancement when API supports it
  })

  // For now, fall back to full sync
  await fullSync()
}
```

### Phase 4: Vector Search

Use embeddings for semantic search:

```typescript
// "tasks due this week" → finds "Weekly Tasks" database
async function semanticResolve(query: string): Promise<string> {
  const embedding = await getEmbedding(query)
  const matches = cache.databases.map(db => ({
    db,
    score: cosineSimilarity(embedding, db.embedding)
  }))
  return matches[0].db.id
}
```

## Testing Strategy

### Unit Tests

```typescript
describe('Database Cache', () => {
  describe('Alias Generation', () => {
    it('generates aliases for simple title', () => {
      expect(generateAliases('Tasks')).toContain('tasks')
      expect(generateAliases('Tasks')).toContain('task')
      expect(generateAliases('Tasks')).toContain('tasks db')
    })

    it('generates acronyms for multi-word titles', () => {
      expect(generateAliases('Meeting Notes')).toContain('mn')
    })
  })

  describe('Name Resolution', () => {
    it('resolves exact title match', async () => {
      const id = await resolveDatabase('Tasks Database')
      expect(id).toBe('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('resolves alias match', async () => {
      const id = await resolveDatabase('tasks')
      expect(id).toBe('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('resolves fuzzy match', async () => {
      const id = await resolveDatabase('task db')
      expect(id).toBe('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('throws error for non-existent database', async () => {
      await expect(
        resolveDatabase('nonexistent', { syncIfNotFound: false })
      ).rejects.toThrow('Could not find database')
    })
  })

  describe('Fuzzy Matching', () => {
    it('calculates fuzzy scores correctly', () => {
      expect(fuzzyScore('tasks', 'tasks')).toBe(1.0)
      expect(fuzzyScore('task', 'tasks')).toBeGreaterThan(0.8)
      expect(fuzzyScore('xyz', 'tasks')).toBeLessThan(0.5)
    })
  })
})
```

### Integration Tests

```typescript
describe('Database Cache Integration', () => {
  it('syncs databases from API', async () => {
    await syncCache()
    const cache = await loadCache()
    expect(cache.databases.length).toBeGreaterThan(0)
  })

  it('handles concurrent sync attempts', async () => {
    const promises = [syncCache(), syncCache(), syncCache()]
    const results = await Promise.allSettled(promises)
    expect(results.filter(r => r.status === 'fulfilled')).toHaveLength(1)
  })

  it('recovers from corrupted cache', async () => {
    await writeCacheFile('invalid json{')
    const cache = await loadCache()
    expect(cache.databases).toEqual([])
  })
})
```

## Security Considerations

1. **File Permissions**: Cache file should be readable only by user (chmod 600)

2. **Sensitive Data**: Cache contains database IDs and titles (not content)
   - IDs are not secret (needed for API calls)
   - Titles may be sensitive (business info)
   - Don't cache in shared/public directories

3. **API Token**: Never cache the NOTION_TOKEN
   - Always read from environment variable
   - Never write to cache file

4. **Lock File**: Prevent race conditions in multi-process environments
   - Use atomic operations (write with 'wx' flag)
   - Clean up stale locks (>5 minutes)

## Monitoring & Observability

### Cache Metrics

```typescript
interface CacheMetrics {
  hits: number
  misses: number
  syncCount: number
  syncErrors: number
  avgSyncDuration: number
  lastSyncDuration: number
  cacheSize: number
  cacheAge: number
}
```

### Logging

```typescript
// Debug logging (when DEBUG=true)
console.log('[DB-CACHE] Loading cache from disk...')
console.log('[DB-CACHE] Cache hit: "tasks" → 1fb79d4c...')
console.log('[DB-CACHE] Cache miss: "unknown db" → triggering sync')
console.log('[DB-CACHE] Sync started: fetching databases...')
console.log('[DB-CACHE] Sync complete: 42 databases in 25s')
```

### Telemetry (optional)

```typescript
// Send metrics to monitoring service
trackEvent('db_cache_hit', { query: 'tasks', matchType: 'exact' })
trackEvent('db_cache_miss', { query: 'unknown db' })
trackEvent('db_cache_sync', { duration: 25000, count: 42 })
```

## Summary

This caching architecture provides:

1. **Fast Lookups**: Sub-millisecond name-to-ID resolution
2. **Natural Language**: Fuzzy matching with aliases
3. **Resilient**: Graceful degradation with stale cache fallback
4. **Scalable**: Efficient sync with rate limiting
5. **User-Friendly**: "Just works" with zero configuration
6. **Developer-Friendly**: Clear error messages and debugging tools

The system balances performance, reliability, and ease of use - perfect for AI agents that need quick, natural database access without complex setup.

---

## File Structure

```
~/.notion-cli/
├── databases.json      # Main cache file
├── databases.json.bak  # Backup (rotated on sync)
├── .sync.lock          # Sync lock file
└── logs/
    └── sync.log        # Sync history (optional)
```

## Implementation Priority

1. **Phase 1 (MVP)**: Name resolution + basic sync
   - Cache schema (databases.json)
   - Sync command
   - Name resolution with exact/alias matching
   - Integration with existing commands

2. **Phase 2**: Fuzzy matching + auto-sync
   - Levenshtein distance
   - Auto-sync on cache miss
   - Cache stats command

3. **Phase 3**: Advanced features
   - Page indexing
   - Incremental sync
   - Vector/semantic search

Start with Phase 1 to prove the concept, then iterate based on user feedback.
