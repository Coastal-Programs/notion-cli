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

Aliases are auto-generated from the title to improve fuzzy matching. The implementation lives in `internal/cache/workspace.go`.

**Example: "Tasks Database"**
Generates: `["tasks database", "tasks", "task", "tasks db", "task db", "td"]`

**Example: "Meeting Notes"**
Generates: `["meeting notes", "meeting note", "meeting", "meeting db", "mn"]`

The algorithm:
1. Normalize title (lowercase, trim)
2. Strip common suffixes (database, db, table, list, tracker, log)
3. Add common suffix variants (db, database)
4. Add singular/plural variants
5. Generate acronym for multi-word titles

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

The sync implementation in `internal/cli/commands/sync.go` handles Notion API pagination, fetching up to 100 databases per page and continuing until all results are retrieved. Retries use the infrastructure in `internal/retry/retry.go`.

#### Atomic File Write

The workspace cache (`internal/cache/workspace.go`) writes to a temporary file first, then performs an atomic rename to prevent corruption if the process crashes during write.

#### Sync Lock File

Concurrent sync operations are prevented using a lock file mechanism. Stale locks (older than 5 minutes) are automatically removed.

## Name Resolution Algorithm

### Resolution Flow

The name resolution algorithm is implemented in `internal/resolver/resolver.go`. It follows a cascading strategy:

1. **URL extraction**: If input is a Notion URL, extract the ID
2. **Direct ID validation**: If input is a 32-character hex string, use it directly
3. **Cache lookup**: Search the workspace cache:
   - 3a. Exact match on normalized title
   - 3b. Exact match on alias
   - 3c. Fuzzy match using Levenshtein distance (threshold: 0.7)
4. **Sync and retry**: If not found in cache, trigger a sync and retry
5. **API search**: Fall back to the Notion search API
6. **Error**: Provide a helpful error message with suggestions

### Fuzzy Matching Algorithm

Uses normalized Levenshtein distance: `score = 1 - (distance / maxLength)`

Example scores:
- `fuzzyScore("tasks", "tasks database")` = 0.73
- `fuzzyScore("task", "tasks")` = 0.80
- `fuzzyScore("meeting", "meetings")` = 0.89
- `fuzzyScore("td", "tasks database")` = 0.14 (too low, below threshold)

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

Default TTL is 1 hour, configurable via `NOTION_CLI_DB_CACHE_TTL`. Maximum absolute TTL is 24 hours. The in-memory cache (`internal/cache/cache.go`) uses per-resource-type TTLs: blocks (30s), pages (1min), users (1hr), databases (10min).

### Incremental Updates

On database mutation operations, the cache is updated immediately. The workspace cache (`internal/cache/workspace.go`) handles adding/updating/removing entries after create, update, or delete operations.

## API Rate Limiting Considerations

### Search API Rate Limits

Notion API rate limits (as of 2025):
- **General requests**: 3 requests per second
- **Search endpoint**: Higher cost (counts as multiple requests)
- **Rate limit response**: HTTP 429 with Retry-After header

### Sync Optimization Strategies

1. **Batch Processing**: Fetch 100 databases per search page (max allowed)

2. **Parallel Detail Fetching**: Fetch database details in parallel with a concurrency limit (default: 3 concurrent requests to respect rate limits)

3. **Smart Retry**: Uses the retry infrastructure in `internal/retry/retry.go` with exponential backoff and jitter

4. **Partial Results**: If sync fails midway, partial results are cached with a warning flag

## Error Handling Patterns

### Error Hierarchy

Error handling uses `internal/errors/errors.go` with the `NotionCLIError` type, which includes error codes and suggestions. Key error codes for cache operations:

- `SYNC_ERROR` - Sync operation failed
- `DATABASE_NOT_FOUND` - Database not found in cache or API
- `CACHE_CORRUPTED` - Cache file is invalid or corrupted

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

The cache loading logic (`internal/cache/workspace.go`) handles failure gracefully:

1. **Missing cache file**: Creates an empty cache (first-time setup)
2. **Corrupted cache**: Backs up the corrupted file and creates a new empty cache
3. **Stale cache**: Warns the user and suggests running `notion-cli sync`

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

The persistent database cache complements the in-memory TTL cache (`internal/cache/cache.go`):

- **In-memory cache**: Fast, temporary, per-process - stores full API responses
- **Persistent cache**: File-based, cross-process - stores database metadata for name resolution

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

Only fetch databases modified since last sync. Currently falls back to full sync since the Notion API doesn't support time-based filters on the search endpoint.

### Phase 4: Vector Search

Use embeddings for semantic search (e.g., "tasks due this week" finds "Weekly Tasks" database).

## Testing Strategy

Tests are implemented as Go test files alongside their source code. Run with `make test`.

### Key Test Areas

- **Alias Generation**: Verifies aliases for simple titles, multi-word titles, and acronyms
- **Name Resolution**: Tests exact match, alias match, fuzzy match, and not-found scenarios
- **Fuzzy Matching**: Validates score calculations for identical, similar, and dissimilar strings
- **Sync**: Tests API pagination, concurrent sync prevention, and partial result handling
- **Cache Recovery**: Tests handling of missing, corrupted, and stale cache files

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

The cache tracks hits, misses, sync count, sync errors, average sync duration, cache size, and cache age. These are accessible via the `cache info` command.

### Logging

Debug logging is available via the `--verbose` flag, showing cache hits/misses, sync progress, and timing information.

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
