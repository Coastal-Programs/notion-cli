# Name Resolution Testing Plan

## Overview

This document outlines the testing strategy for the name resolution feature in notion-cli v5.3.0. The name resolution system allows users to reference databases by:
- **URLs**: Full Notion URLs (e.g., `https://notion.so/abc123...`)
- **Direct IDs**: Raw database IDs (e.g., `abc123...`)
- **Names**: Database titles (e.g., `"Tasks Database"`)
- **Aliases**: Generated nicknames (e.g., `"tasks"`, `"td"`)
- **Partial Matches**: Substring matches (e.g., `"task"`)

## Architecture Components

### Core Files
- `src/utils/notion-resolver.ts` - Main resolution logic with fallback stages
- `src/utils/workspace-cache.ts` - Persistent cache management
- `src/utils/notion-url-parser.ts` - URL parsing and validation
- `src/commands/sync.ts` - Workspace synchronization command
- `src/commands/list.ts` - Cache browsing command

### Resolution Stages
1. **URL Extraction** - Parse Notion URLs to extract IDs
2. **Direct ID Validation** - Validate and normalize raw IDs
3. **Cache Lookup** - Search cached databases by exact name, aliases, and partial matches
4. **API Fallback** - Query Notion API if cache lookup fails

---

## Unit Tests (if time permits)

### Test Suite: `resolveNotionId()`

#### URL Input Tests
```typescript
describe('resolveNotionId - URL input', () => {
  test('should extract ID from full Notion URL', async () => {
    const input = 'https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00'
    const result = await resolveNotionId(input)
    expect(result).toBe('1fb79d4c71bb8032b722c82305b63a00')
  })

  test('should extract ID from URL with dashes', async () => {
    const input = 'https://www.notion.so/1fb79d4c-71bb-8032-b722-c82305b63a00'
    const result = await resolveNotionId(input)
    expect(result).toBe('1fb79d4c71bb8032b722c82305b63a00')
  })

  test('should extract ID from URL with page title', async () => {
    const input = 'https://www.notion.so/Tasks-Database-1fb79d4c71bb8032b722c82305b63a00'
    const result = await resolveNotionId(input)
    expect(result).toBe('1fb79d4c71bb8032b722c82305b63a00')
  })

  test('should throw error for invalid Notion URL', async () => {
    const input = 'https://example.com/not-notion'
    await expect(resolveNotionId(input)).rejects.toThrow('Invalid Notion URL')
  })
})
```

#### Raw ID Input Tests
```typescript
describe('resolveNotionId - Raw ID input', () => {
  test('should accept ID without dashes', async () => {
    const input = '1fb79d4c71bb8032b722c82305b63a00'
    const result = await resolveNotionId(input)
    expect(result).toBe('1fb79d4c71bb8032b722c82305b63a00')
  })

  test('should accept ID with dashes', async () => {
    const input = '1fb79d4c-71bb-8032-b722-c82305b63a00'
    const result = await resolveNotionId(input)
    expect(result).toBe('1fb79d4c71bb8032b722c82305b63a00')
  })

  test('should reject ID with invalid characters', async () => {
    const input = '1fb79d4c71bb8032b722c82305b63aXX' // XX not valid hex
    await expect(resolveNotionId(input)).rejects.toThrow('Invalid')
  })

  test('should reject ID with wrong length', async () => {
    const input = '1fb79d4c71bb8032' // Too short
    await expect(resolveNotionId(input)).rejects.toThrow('Invalid')
  })
})
```

#### Cache Lookup Tests (Name Resolution)
```typescript
describe('resolveNotionId - Name lookup', () => {
  beforeEach(async () => {
    // Mock cache with test data
    await mockCache({
      databases: [
        {
          id: '1fb79d4c71bb8032b722c82305b63a00',
          title: 'Tasks Database',
          aliases: ['tasks database', 'tasks', 'task', 'td'],
        },
        {
          id: '2a8c3d5e71bb8042b833d94316c74b11',
          title: 'Meeting Notes',
          aliases: ['meeting notes', 'meetings', 'notes', 'mn'],
        },
      ],
    })
  })

  test('should resolve by exact title match', async () => {
    const result = await resolveNotionId('Tasks Database')
    expect(result).toBe('1fb79d4c71bb8032b722c82305b63a00')
  })

  test('should be case-insensitive', async () => {
    const result = await resolveNotionId('tasks database')
    expect(result).toBe('1fb79d4c71bb8032b722c82305b63a00')
  })

  test('should resolve by alias', async () => {
    const result = await resolveNotionId('tasks')
    expect(result).toBe('1fb79d4c71bb8032b722c82305b63a00')
  })

  test('should resolve by acronym', async () => {
    const result = await resolveNotionId('td')
    expect(result).toBe('1fb79d4c71bb8032b722c82305b63a00')
  })

  test('should resolve by partial match', async () => {
    const result = await resolveNotionId('task')
    expect(result).toBe('1fb79d4c71bb8032b722c82305b63a00')
  })

  test('should throw helpful error for not found', async () => {
    await expect(resolveNotionId('nonexistent')).rejects.toThrow(
      /not found.*Run.*notion-cli sync/i
    )
  })
})
```

#### Error Handling Tests
```typescript
describe('resolveNotionId - Error cases', () => {
  test('should reject null input', async () => {
    await expect(resolveNotionId(null as any)).rejects.toThrow('Invalid input')
  })

  test('should reject empty string', async () => {
    await expect(resolveNotionId('')).rejects.toThrow('Invalid input')
  })

  test('should reject undefined', async () => {
    await expect(resolveNotionId(undefined as any)).rejects.toThrow('Invalid input')
  })

  test('should trim whitespace', async () => {
    const input = '  1fb79d4c71bb8032b722c82305b63a00  '
    const result = await resolveNotionId(input)
    expect(result).toBe('1fb79d4c71bb8032b722c82305b63a00')
  })
})
```

### Test Suite: `searchCache()`

```typescript
describe('searchCache', () => {
  test('should find exact title match first', async () => {
    // Cache has "Tasks" and "Task Manager"
    const result = await searchCache('Tasks')
    expect(result.title).toBe('Tasks') // Not "Task Manager"
  })

  test('should rank exact matches higher than partial', async () => {
    const results = await searchCache('task')
    expect(results[0].title).toBe('Tasks') // Exact alias match
    expect(results[1].title).toBe('Task Manager') // Partial match
  })

  test('should handle multi-word queries', async () => {
    const result = await searchCache('meeting notes')
    expect(result.title).toBe('Meeting Notes')
  })

  test('should handle acronyms', async () => {
    const result = await searchCache('mn')
    expect(result.title).toBe('Meeting Notes')
  })
})
```

### Test Suite: `generateAliases()`

```typescript
describe('generateAliases', () => {
  test('should generate basic aliases', () => {
    const aliases = generateAliases('Tasks Database')
    expect(aliases).toContain('tasks database')
    expect(aliases).toContain('tasks')
    expect(aliases).toContain('task')
    expect(aliases).toContain('td')
  })

  test('should handle plural/singular', () => {
    const aliases = generateAliases('Meeting Notes')
    expect(aliases).toContain('meeting notes')
    expect(aliases).toContain('meeting note')
  })

  test('should remove common suffixes', () => {
    const aliases = generateAliases('Customer Database')
    expect(aliases).toContain('customer')
    expect(aliases).toContain('customer db')
  })

  test('should create acronyms from multi-word titles', () => {
    const aliases = generateAliases('Customer Relationship Management')
    expect(aliases).toContain('crm')
  })

  test('should handle single word titles', () => {
    const aliases = generateAliases('Tasks')
    expect(aliases).toContain('tasks')
    expect(aliases).toContain('task')
  })
})
```

---

## Integration Tests (PRIORITY)

### Test Workflow: End-to-End Name Resolution

#### Setup
```bash
# Prerequisites
export NOTION_TOKEN="secret_your_test_token"
cd notion-cli
npm run build
```

#### Test 1: Sync Workspace → List Databases
```bash
# Step 1: Sync workspace
./bin/run sync --json > sync-result.json

# Verify:
# - success: true
# - count > 0
# - databases array populated
# - cachePath points to ~/.notion-cli/databases.json
cat sync-result.json | jq '.success'
# Expected: true

# Step 2: List cached databases
./bin/run list --json > list-result.json

# Verify:
# - databases array matches sync result
# - each entry has: id, title, aliases
cat list-result.json | jq '.databases | length'
# Expected: Same as sync count
```

#### Test 2: Retrieve by Exact Title
```bash
# Get a database title from cache
DB_TITLE=$(./bin/run list --json | jq -r '.databases[0].title')

# Try to retrieve using exact title
./bin/run db retrieve "$DB_TITLE" --json > retrieve-exact.json

# Verify:
# - success: true
# - data.id matches expected database
cat retrieve-exact.json | jq '.success'
# Expected: true
```

#### Test 3: Retrieve by Alias
```bash
# Get first alias for a database
DB_ALIAS=$(./bin/run list --json | jq -r '.databases[0].aliases[0]')

# Try to retrieve using alias
./bin/run db retrieve "$DB_ALIAS" --json > retrieve-alias.json

# Verify:
# - success: true
# - resolves to same database as exact title
cat retrieve-alias.json | jq '.success'
# Expected: true
```

#### Test 4: Retrieve by Partial Match
```bash
# Get partial name (first 4 chars of title)
DB_PARTIAL=$(./bin/run list --json | jq -r '.databases[0].title[:4]')

# Try to retrieve using partial match
./bin/run db retrieve "$DB_PARTIAL" --json > retrieve-partial.json

# Verify:
# - success: true
# - resolves correctly
cat retrieve-partial.json | jq '.success'
# Expected: true
```

#### Test 5: Retrieve by URL (Backward Compatibility)
```bash
# Get URL from cache
DB_URL=$(./bin/run list --json | jq -r '.databases[0].url')

# Try to retrieve using URL
./bin/run db retrieve "$DB_URL" --json > retrieve-url.json

# Verify:
# - success: true
# - URL parsing still works
cat retrieve-url.json | jq '.success'
# Expected: true
```

#### Test 6: Retrieve by Raw ID (Backward Compatibility)
```bash
# Get ID from cache
DB_ID=$(./bin/run list --json | jq -r '.databases[0].id')

# Try to retrieve using raw ID
./bin/run db retrieve "$DB_ID" --json > retrieve-id.json

# Verify:
# - success: true
# - Direct ID access still works
cat retrieve-id.json | jq '.success'
# Expected: true
```

#### Test 7: Cache Miss → API Fallback
```bash
# Clear cache
rm ~/.notion-cli/databases.json

# Try to retrieve without cache (should fallback to API)
./bin/run db retrieve "$DB_ID" --json > retrieve-no-cache.json

# Verify:
# - success: true
# - Falls back to API search
# - Shows helpful message about syncing
cat retrieve-no-cache.json | jq '.success'
# Expected: true (with warning about cache)
```

#### Test 8: Helpful Error Messages
```bash
# Try to retrieve nonexistent database
./bin/run db retrieve "ThisDatabaseDoesNotExist" --json > retrieve-error.json

# Verify error structure:
# - success: false
# - error.code: "not_found"
# - error.message contains helpful guidance
# - Suggests running "notion-cli sync"
cat retrieve-error.json | jq '.error.message'
# Expected: Helpful message with next steps
```

### Test Results Template

```markdown
## Integration Test Results

Date: [DATE]
Version: 5.3.0
Tester: [NAME]

### Environment
- OS: [Windows/Mac/Linux]
- Node.js: [VERSION]
- NOTION_TOKEN: [Set/Not Set]

### Test Results

| Test | Status | Notes |
|------|--------|-------|
| Sync workspace | PASS/FAIL | |
| List databases | PASS/FAIL | |
| Retrieve by exact title | PASS/FAIL | |
| Retrieve by alias | PASS/FAIL | |
| Retrieve by partial match | PASS/FAIL | |
| Retrieve by URL | PASS/FAIL | |
| Retrieve by raw ID | PASS/FAIL | |
| Cache miss fallback | PASS/FAIL | |
| Error messages | PASS/FAIL | |

### Issues Found
- [Issue 1]
- [Issue 2]

### Performance Notes
- Sync time: [X seconds for Y databases]
- Cache lookup time: [< 1ms expected]
- API fallback time: [200-500ms typical]
```

---

## Manual Testing Checklist

### Build & Setup
- [ ] Build succeeds: `npm run build`
- [ ] No TypeScript errors
- [ ] No lint errors: `npm run lint`
- [ ] Help text shows: `./bin/run --help`
- [ ] Version displays: `./bin/run --version`

### Token Configuration
- [ ] Token setup works: `./bin/run config set-token`
- [ ] Token validation: `./bin/run user retrieve bot --json`
- [ ] Environment variable set: `echo $NOTION_TOKEN`

### Sync Command
- [ ] Sync works: `./bin/run sync`
- [ ] Shows progress indicators
- [ ] Creates cache file: `~/.notion-cli/databases.json`
- [ ] JSON output: `./bin/run sync --json`
- [ ] Force resync: `./bin/run sync --force`

### List Command
- [ ] List shows cached databases: `./bin/run list`
- [ ] Shows titles, IDs, aliases
- [ ] JSON format: `./bin/run list --json`
- [ ] Markdown format: `./bin/run list --markdown`
- [ ] Pretty table: `./bin/run list --pretty`

### Name Resolution
- [ ] Retrieve by exact name: `./bin/run db retrieve "Database Name"`
- [ ] Case insensitive: `./bin/run db retrieve "database name"`
- [ ] Retrieve by alias: `./bin/run db retrieve "alias"`
- [ ] Retrieve by partial: `./bin/run db retrieve "part"`
- [ ] Retrieve by URL: `./bin/run db retrieve https://notion.so/...`
- [ ] Retrieve by ID: `./bin/run db retrieve abc123...`

### Error Handling
- [ ] Nonexistent name: `./bin/run db retrieve "NonexistentDB"`
- [ ] Error message helpful
- [ ] Suggests running sync
- [ ] JSON error format correct
- [ ] Exit code is 1 on error

### Edge Cases
- [ ] Empty cache behavior
- [ ] Cache corruption recovery
- [ ] Multiple databases with similar names
- [ ] Databases with special characters in names
- [ ] Very long database names (>100 chars)
- [ ] Databases with emoji in names

### Performance
- [ ] Sync completes in reasonable time (< 10s for 50 DBs)
- [ ] List is instant (< 100ms)
- [ ] Name resolution is instant (< 10ms)
- [ ] No memory leaks with large caches (500+ DBs)

### Cross-Platform
- [ ] Works on Windows
- [ ] Works on macOS
- [ ] Works on Linux
- [ ] Cache path correct for each OS

---

## Performance Benchmarks

### Expected Performance

| Operation | Expected Time | Acceptable Range |
|-----------|---------------|------------------|
| Sync (50 DBs) | 3-5s | < 10s |
| List cached DBs | < 100ms | < 200ms |
| Name resolution | < 10ms | < 50ms |
| URL resolution | < 1ms | < 5ms |
| ID resolution | < 1ms | < 5ms |
| API fallback | 200-500ms | < 1s |

### Benchmark Script

```bash
#!/bin/bash
# performance-benchmark.sh

echo "=== Name Resolution Performance Benchmark ==="
echo ""

# Test 1: Sync time
echo "Test 1: Sync workspace..."
START=$(date +%s%N)
./bin/run sync --json > /dev/null
END=$(date +%s%N)
SYNC_TIME=$(( (END - START) / 1000000 ))
echo "Sync time: ${SYNC_TIME}ms"
echo ""

# Test 2: List time
echo "Test 2: List databases..."
START=$(date +%s%N)
./bin/run list --json > /dev/null
END=$(date +%s%N)
LIST_TIME=$(( (END - START) / 1000000 ))
echo "List time: ${LIST_TIME}ms"
echo ""

# Test 3: Name resolution time
echo "Test 3: Resolve by name..."
DB_NAME=$(./bin/run list --json | jq -r '.databases[0].title')
START=$(date +%s%N)
./bin/run db retrieve "$DB_NAME" --json > /dev/null
END=$(date +%s%N)
NAME_TIME=$(( (END - START) / 1000000 ))
echo "Name resolution time: ${NAME_TIME}ms"
echo ""

# Test 4: ID resolution time
echo "Test 4: Resolve by ID..."
DB_ID=$(./bin/run list --json | jq -r '.databases[0].id')
START=$(date +%s%N)
./bin/run db retrieve "$DB_ID" --json > /dev/null
END=$(date +%s%N)
ID_TIME=$(( (END - START) / 1000000 ))
echo "ID resolution time: ${ID_TIME}ms"
echo ""

echo "=== Summary ==="
echo "Sync: ${SYNC_TIME}ms"
echo "List: ${LIST_TIME}ms"
echo "Name resolution: ${NAME_TIME}ms"
echo "ID resolution: ${ID_TIME}ms"
```

---

## Test Data Scenarios

### Scenario 1: Simple Names
```json
{
  "databases": [
    {
      "id": "1fb79d4c71bb8032b722c82305b63a00",
      "title": "Tasks",
      "aliases": ["tasks", "task"]
    },
    {
      "id": "2a8c3d5e71bb8042b833d94316c74b11",
      "title": "Notes",
      "aliases": ["notes", "note"]
    }
  ]
}
```

### Scenario 2: Complex Names
```json
{
  "databases": [
    {
      "id": "3b9d4e6f82bc8053c944e05427d85c22",
      "title": "Customer Relationship Management",
      "aliases": ["customer relationship management", "crm", "customer"]
    },
    {
      "id": "4c0e5f7g93cd9164d055f16538e96d33",
      "title": "Meeting Notes & Action Items",
      "aliases": ["meeting notes & action items", "meeting", "mnai"]
    }
  ]
}
```

### Scenario 3: Similar Names (Ambiguity)
```json
{
  "databases": [
    {
      "id": "5d1f6g8h04de0275e166g27649f07e44",
      "title": "Tasks Database",
      "aliases": ["tasks database", "tasks", "task", "td"]
    },
    {
      "id": "6e2g7h9i15ef1386f277h38750g18f55",
      "title": "Tasks Archive",
      "aliases": ["tasks archive", "tasks", "task", "ta"]
    }
  ]
}
```

### Scenario 4: Special Characters
```json
{
  "databases": [
    {
      "id": "7f3h8i0j26fg2497g388i49861h29g66",
      "title": "Q1 2025 - Revenue Tracking 📊",
      "aliases": ["q1 2025 - revenue tracking 📊", "q1 2025", "revenue", "q125rt"]
    }
  ]
}
```

---

## Troubleshooting Guide

### Issue: Cache not created after sync
**Symptoms:**
- `sync` command succeeds but cache file missing
- `list` command shows no databases

**Debug Steps:**
1. Check cache directory exists: `ls -la ~/.notion-cli/`
2. Check file permissions: `ls -la ~/.notion-cli/databases.json`
3. Run sync with debug: `DEBUG=true ./bin/run sync`
4. Check for write errors in output

**Fix:**
- Ensure home directory is writable
- Delete and recreate cache directory
- Check disk space

### Issue: Name resolution not working
**Symptoms:**
- ID/URL resolution works
- Name resolution returns "not found"

**Debug Steps:**
1. Verify cache exists: `cat ~/.notion-cli/databases.json`
2. Check database in cache: `./bin/run list | grep "Database Name"`
3. Verify aliases generated: `./bin/run list --json | jq '.databases[0].aliases'`
4. Test with exact title (case-sensitive): `./bin/run db retrieve "Exact Title"`

**Fix:**
- Run `notion-cli sync --force`
- Verify database title exactly matches
- Check for typos in database name

### Issue: Slow sync performance
**Symptoms:**
- Sync takes > 30 seconds
- Progress indicators lag

**Debug Steps:**
1. Count databases: `./bin/run sync --json | jq '.count'`
2. Check network: `ping notion.so`
3. Test API directly: `curl -H "Authorization: Bearer $NOTION_TOKEN" https://api.notion.com/v1/search`

**Fix:**
- Check internet connection
- Verify API rate limits not hit
- Try again later
- Use `--force` to bypass stale cache

---

## Acceptance Criteria

### Must Have (P0)
- [ ] `sync` command creates cache successfully
- [ ] `list` command shows cached databases
- [ ] Name resolution works for exact titles
- [ ] URL resolution still works (backward compatibility)
- [ ] ID resolution still works (backward compatibility)
- [ ] Error messages are helpful and actionable
- [ ] JSON output format consistent across commands

### Should Have (P1)
- [ ] Alias resolution works
- [ ] Partial match resolution works
- [ ] Cache automatically refreshed when stale
- [ ] Multiple output formats (table, markdown, pretty)
- [ ] Case-insensitive matching

### Nice to Have (P2)
- [ ] Fuzzy matching for typos
- [ ] Disambiguation for similar names
- [ ] Cache statistics (hit rate, size)
- [ ] Progress bars for long operations
- [ ] Color-coded output

---

## Next Steps

After testing is complete:

1. **Document Results** - Fill in test results template
2. **File Issues** - Create GitHub issues for any bugs found
3. **Update Documentation** - Add findings to README and examples
4. **Performance Optimization** - If benchmarks fail, optimize slow paths
5. **Release Notes** - Document what works and known limitations
