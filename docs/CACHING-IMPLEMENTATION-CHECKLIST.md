# Database Caching Implementation Checklist

This checklist breaks down the implementation of the database caching system into manageable tasks.

## Phase 1: MVP (Minimum Viable Product)

### 1. Core Cache Infrastructure

- [ ] **Create cache types** (`src/cache-db.ts`)
  ```typescript
  interface DatabaseCache {
    version: string
    lastSync: string
    syncStatus: SyncStatus
    databases: CachedDatabase[]
    metadata: CacheMetadata
  }

  interface CachedDatabase {
    id: string
    title: string
    titleNormalized: string
    aliases: string[]
    properties: Record<string, PropertyInfo>
    // ... other fields
  }
  ```

- [ ] **Create cache manager class** (`src/cache-db.ts`)
  ```typescript
  class DatabaseCacheManager {
    async loadCache(): Promise<DatabaseCache>
    async writeCache(cache: DatabaseCache): Promise<void>
    async syncCache(force?: boolean): Promise<void>
    async resolveDatabase(input: string): Promise<string>
    // ... other methods
  }
  ```

- [ ] **Implement cache file I/O**
  - Get cache directory path (`~/.notion-cli/`)
  - Create directory if not exists
  - Atomic write with temp file + rename
  - Handle corrupted cache gracefully
  - Backup old cache on corruption

- [ ] **Implement sync lock mechanism**
  - Create/check lock file (`.sync.lock`)
  - Detect stale locks (>5 minutes)
  - Auto-cleanup stale locks
  - Prevent concurrent syncs

### 2. Alias Generation

- [ ] **Implement alias generation** (`src/utils/alias-generator.ts`)
  ```typescript
  function generateAliases(title: string): string[]
  ```
  - Normalize title (lowercase, trim)
  - Remove common suffixes (database, db, table, list)
  - Add singular/plural variants
  - Generate acronyms for multi-word titles
  - Return unique aliases

- [ ] **Add tests for alias generation**
  - Test simple titles: "Tasks" → ["tasks", "task", "tasks db"]
  - Test multi-word: "Meeting Notes" → [..., "mn"]
  - Test edge cases: special chars, numbers, etc.

### 3. Sync Logic

- [ ] **Implement database fetching** (`src/cache-db.ts`)
  ```typescript
  async function fetchAllDatabases(): Promise<DataSourceObjectResponse[]>
  ```
  - Use search API with `object = 'data_source'` filter
  - Handle pagination (100 items per page)
  - Respect rate limits (3 req/sec)
  - Use existing retry infrastructure

- [ ] **Implement cache entry building**
  ```typescript
  async function buildCacheEntry(
    database: DataSourceObjectResponse
  ): Promise<CachedDatabase>
  ```
  - Extract id, title, url, parent
  - Retrieve full data source details for properties
  - Generate aliases
  - Normalize property names

- [ ] **Implement full sync**
  ```typescript
  async function syncCache(force = false): Promise<void>
  ```
  - Check if sync needed (force or cache stale)
  - Acquire sync lock
  - Fetch all databases
  - Build cache entries
  - Write cache atomically
  - Release lock
  - Handle errors (save partial results)

### 4. Name Resolution

- [ ] **Implement basic resolver** (`src/utils/database-resolver.ts`)
  ```typescript
  async function resolveDatabase(
    input: string,
    options?: ResolveOptions
  ): Promise<string>
  ```
  - Check if URL → extract ID
  - Check if ID → validate and return
  - Check if name → search cache
    - Exact match on title
    - Exact match on alias
  - Not found → error

- [ ] **Add tests for name resolution**
  - Test URL extraction
  - Test ID validation
  - Test exact title match
  - Test alias match
  - Test error cases

### 5. CLI Commands

- [ ] **Create `db sync` command** (`src/commands/db/sync.ts`)
  ```bash
  notion-cli db sync [--force] [--json]
  ```
  - Check if sync already in progress
  - Run sync operation
  - Show progress (databases fetched)
  - Output results (count, duration)
  - Handle errors gracefully

- [ ] **Create `db list` command** (`src/commands/db/list.ts`)
  ```bash
  notion-cli db list [--filter <pattern>] [--json]
  ```
  - Load cache
  - Filter by pattern (optional)
  - Display table/json output
  - Show cache age warning if stale

- [ ] **Create `db resolve` command** (`src/commands/db/resolve.ts`)
  ```bash
  notion-cli db resolve <name> [--json]
  ```
  - Resolve name to ID
  - Show match type (exact, alias)
  - Output ID and metadata
  - Useful for debugging

- [ ] **Update existing commands to use resolver**
  - `db retrieve` - Accept name or ID
  - `db query` - Accept name or ID
  - `db schema` - Accept name or ID
  - `db update` - Accept name or ID

### 6. Integration & Testing

- [ ] **Integration tests**
  - Test full sync workflow
  - Test name resolution E2E
  - Test error recovery
  - Test concurrent operations

- [ ] **Documentation updates**
  - Update README with cache commands
  - Add examples for natural language lookups
  - Document cache file structure
  - Add troubleshooting guide

- [ ] **Manual testing**
  - Test on Windows, Mac, Linux
  - Test with empty workspace
  - Test with 100+ databases
  - Test network failures
  - Test corrupted cache

## Phase 2: Fuzzy Matching

### 1. Fuzzy Matching Algorithm

- [ ] **Implement Levenshtein distance** (`src/utils/fuzzy-match.ts`)
  ```typescript
  function levenshtein(a: string, b: string): number
  function fuzzyScore(query: string, target: string): number
  ```
  - Calculate edit distance
  - Normalize to 0.0-1.0 score
  - Optimize for performance

- [ ] **Add fuzzy matching to resolver**
  - Search cache with fuzzy matching
  - Filter by threshold (default: 0.7)
  - Sort by score (best match first)
  - Return top match or error

- [ ] **Add tests for fuzzy matching**
  - Test exact matches (score = 1.0)
  - Test close matches (score > 0.8)
  - Test poor matches (score < 0.5)
  - Test edge cases (empty strings, etc.)

### 2. Configuration

- [ ] **Add environment variables**
  ```bash
  NOTION_CLI_DB_FUZZY_THRESHOLD=0.7
  NOTION_CLI_DB_CACHE_TTL=3600000
  NOTION_CLI_DB_AUTO_SYNC=true
  ```

- [ ] **Document configuration options**
  - Update `.env.example`
  - Document in README
  - Add to architecture docs

## Phase 3: Advanced Features

### 1. Auto-Sync

- [ ] **Implement auto-sync on cache miss**
  - Detect cache miss
  - Check if cache is stale
  - Trigger sync automatically
  - Retry resolution
  - Configurable via env var

- [ ] **Implement background sync**
  - Check cache age on every operation
  - Trigger sync in background if stale
  - Don't block current operation
  - Update cache when complete

### 2. Cache Statistics

- [ ] **Create `db cache stats` command**
  ```bash
  notion-cli db cache stats [--json]
  ```
  - Show total databases
  - Show cache age
  - Show cache size
  - Show sync history
  - Show error count

- [ ] **Create `db cache clear` command**
  ```bash
  notion-cli db cache clear [--json]
  ```
  - Remove cache file
  - Remove lock file
  - Confirm action
  - Output success/error

### 3. Enhanced Sync

- [ ] **Implement incremental sync** (future)
  - Track last sync time
  - Only fetch modified databases
  - Merge with existing cache
  - Requires API support for time filters

- [ ] **Implement parallel fetching**
  - Fetch database details in parallel
  - Respect rate limits (3 req/sec)
  - Use promise concurrency control
  - Show progress

## Testing Checklist

### Unit Tests

- [ ] Cache file I/O
- [ ] Alias generation
- [ ] Fuzzy matching
- [ ] Name resolution
- [ ] Sync lock mechanism

### Integration Tests

- [ ] Full sync workflow
- [ ] Concurrent sync prevention
- [ ] Cache corruption recovery
- [ ] API error handling
- [ ] Rate limit handling

### E2E Tests

- [ ] CLI commands
- [ ] Natural language lookups
- [ ] Error messages
- [ ] JSON output format
- [ ] Cross-platform compatibility

### Performance Tests

- [ ] Sync 100 databases
- [ ] Resolve 1000 names
- [ ] Cache file load time
- [ ] Fuzzy search speed

## Documentation Checklist

- [x] Architecture document (CACHING-ARCHITECTURE.md)
- [x] Quick reference guide (CACHING-QUICK-REF.md)
- [x] Implementation checklist (this file)
- [ ] Update main README
- [ ] Update AI Agent Cookbook
- [ ] Add examples to docs
- [ ] Create troubleshooting guide

## Deployment Checklist

- [ ] Bump version to 5.4.0
- [ ] Update CHANGELOG.md
- [ ] Create GitHub release
- [ ] Publish to npm
- [ ] Test installation on all platforms
- [ ] Update documentation website

## Rollout Strategy

### Week 1: MVP
1. Implement core cache infrastructure
2. Implement alias generation
3. Implement basic sync
4. Add unit tests

### Week 2: Integration
1. Implement name resolution
2. Create CLI commands
3. Update existing commands
4. Add integration tests

### Week 3: Polish
1. Implement fuzzy matching
2. Add auto-sync
3. Create cache management commands
4. Update documentation

### Week 4: Release
1. E2E testing
2. Performance testing
3. Bug fixes
4. Release v5.4.0

## Success Metrics

- [ ] Cache hit rate > 95%
- [ ] Name resolution < 10ms
- [ ] Sync 100 databases < 60s
- [ ] Zero cache corruption reports
- [ ] User adoption > 50% (using names vs IDs)

## Known Issues & Considerations

### API Limitations
- Search API doesn't support time-based filters (for incremental sync)
- Rate limit: 3 requests/second
- Max page size: 100 items

### Edge Cases
- Empty workspace (0 databases)
- Very large workspace (>1000 databases)
- Duplicate database names
- Special characters in titles
- Very long database titles

### Performance
- Fuzzy matching on 1000+ databases
- Cache file size >10MB
- Concurrent access to cache file
- Disk I/O on slow filesystems

### Security
- Cache file permissions (chmod 600)
- Lock file cleanup
- Sensitive data in titles
- Token storage (never cache!)

---

## Getting Started

1. Read [CACHING-ARCHITECTURE.md](./CACHING-ARCHITECTURE.md) for design details
2. Review [CACHING-QUICK-REF.md](./CACHING-QUICK-REF.md) for usage examples
3. Start with Phase 1, Task 1: Core Cache Infrastructure
4. Write tests first (TDD approach recommended)
5. Commit frequently with descriptive messages
6. Update this checklist as you progress

Good luck! The architecture is solid, now it's time to build it.
