# Database Caching Quick Reference

## TL;DR

notion-cli now supports **natural language database lookups** via a persistent cache in `~/.notion-cli/databases.json`.

```bash
# Instead of this:
notion-cli db query 1fb79d4c71bb8032b722c82305b63a00 --json

# You can do this:
notion-cli db query "tasks" --json
notion-cli db query "task db" --json
notion-cli db query "Tasks Database" --json
```

## Quick Setup

```bash
# 1. Sync your databases to cache (run once)
notion-cli db sync

# 2. Use natural names in any command
notion-cli db query "meeting notes" --json
notion-cli db schema "tasks" --output json
```

## How It Works

```
Input: "tasks db"
    ↓
1. Is it a URL? → Extract ID
2. Is it an ID? → Validate
3. Is it a name? → Search cache
    ↓
Cache Lookup (in order):
    - Exact title match: "tasks db"
    - Alias match: ["tasks", "task", "tasks db"]
    - Fuzzy match: score >= 0.7
    ↓
Not found? → Auto-sync → Retry
    ↓
Still not found? → Search API
    ↓
Return: "1fb79d4c71bb8032b722c82305b63a00"
```

## Cache Schema Highlights

```json
{
  "version": "1.0.0",
  "lastSync": "2025-10-22T10:30:00.000Z",
  "databases": [
    {
      "id": "1fb79d4c71bb8032b722c82305b63a00",
      "title": "Tasks Database",
      "titleNormalized": "tasks database",
      "aliases": ["tasks", "task", "tasks db", "td"],
      "properties": {
        "Name": { "type": "title" },
        "Status": { "type": "select", "options": [...] }
      }
    }
  ]
}
```

## Key Features

### Alias Generation

Titles automatically generate searchable aliases:

| Title | Aliases |
|-------|---------|
| "Tasks Database" | tasks database, tasks, task, tasks db, td |
| "Meeting Notes" | meeting notes, meeting note, meeting, mn |
| "Weekly Sprint Log" | weekly sprint log, weekly sprint, wsl |

### Fuzzy Matching

Handles typos and variations:

| Query | Matches | Score |
|-------|---------|-------|
| "task" | "Tasks Database" | 0.80 |
| "meeting" | "Meeting Notes" | 0.89 |
| "sprint log" | "Weekly Sprint Log" | 0.75 |

### Auto-Sync

Cache auto-syncs when:
- Cache is older than 1 hour (configurable)
- Database name not found in cache
- User runs `notion-cli db sync`

## Commands

```bash
# Sync database cache
notion-cli db sync [--force] [--json]

# List cached databases
notion-cli db list [--filter <pattern>] [--json]

# Resolve name to ID (debugging)
notion-cli db resolve "tasks" [--json]

# Cache statistics
notion-cli db cache stats [--json]

# Clear cache
notion-cli db cache clear [--json]
```

## Configuration

```bash
# Cache TTL (default: 1 hour)
export NOTION_CLI_DB_CACHE_TTL=3600000

# Fuzzy match threshold (default: 0.7 = 70% match)
export NOTION_CLI_DB_FUZZY_THRESHOLD=0.7

# Sync concurrency (default: 3 requests/sec)
export NOTION_CLI_DB_SYNC_CONCURRENCY=3

# Auto-sync on cache miss (default: true)
export NOTION_CLI_DB_AUTO_SYNC=true

# Custom cache location (default: ~/.notion-cli)
export NOTION_CLI_DB_CACHE_PATH="~/.notion-cli"
```

## Performance

| Operation | Without Cache | With Cache | Speedup |
|-----------|--------------|------------|---------|
| Exact name match | ~200ms | ~1ms | 200x |
| Fuzzy name match | ~300ms | ~5ms | 60x |
| Schema lookup | ~200ms | ~2ms | 100x |

**Cache size:**
- 10 databases: ~15 KB
- 100 databases: ~150 KB
- 1000 databases: ~1.5 MB

**Sync time:**
- 10 databases: ~5s
- 100 databases: ~40s
- 1000 databases: ~6min

## Error Handling

### Database Not Found

```bash
$ notion-cli db query "nonexistent"
Error: Could not find database matching: "nonexistent"

Tried:
  - URL extraction
  - ID validation
  - Cache lookup (exact, alias, fuzzy)
  - Cache sync
  - API search

Suggestions:
  - Verify the database exists and is shared with your integration
  - Try using the full database ID or URL
  - Run 'notion-cli db sync' to refresh the cache
```

### Stale Cache Warning

```bash
$ notion-cli db query "tasks" --json
Warning: Database cache is stale. Run "notion-cli db sync" to refresh.

{
  "success": true,
  "data": {...}
}
```

### Corrupted Cache

Automatically recovers:
1. Backs up corrupted cache to `.bak` file
2. Creates fresh empty cache
3. Triggers sync

## Integration Examples

### Use in Scripts

```bash
#!/bin/bash

# Resolve database name to ID
DB_ID=$(notion-cli db resolve "tasks" --json | jq -r '.data.id')

# Use resolved ID
notion-cli db query "$DB_ID" --json | jq '.data.results'
```

### Pipe with jq

```bash
# Get all task database IDs
notion-cli db list --json | jq -r '.data[].id'

# Find databases matching "task"
notion-cli db list --json | \
  jq '.data[] | select(.title | contains("Task"))'

# Get schema for multiple databases
for db in "tasks" "meetings" "notes"; do
  echo "=== $db ==="
  notion-cli db schema "$db" --output json | \
    jq '.data.properties[] | {name, type, options}'
done
```

### AI Agent Usage

```typescript
// Discover database structure
const schema = await exec('notion-cli db schema "tasks" --output json')
const properties = JSON.parse(schema).data.properties

// Create page with correct properties
const pageData = {
  Name: "New Task",
  Status: properties.find(p => p.name === "Status").options[0],
  Tags: ["urgent"]
}

await exec(`notion-cli page create "tasks" --json`, {
  input: JSON.stringify(pageData)
})
```

## Troubleshooting

### Cache Not Working

```bash
# Check cache exists
ls -lh ~/.notion-cli/databases.json

# Check cache is valid JSON
cat ~/.notion-cli/databases.json | jq .

# Force refresh
notion-cli db sync --force

# Check cache stats
notion-cli db cache stats
```

### Sync Failing

```bash
# Check API token
echo $NOTION_TOKEN

# Test API access
notion-cli user retrieve bot --json

# Check rate limiting
# Wait 1 minute and retry

# Enable debug logging
DEBUG=true notion-cli db sync
```

### Database Not Found After Sync

Possible reasons:
1. Database not shared with integration → Share it in Notion
2. Database is archived → Use `--include-archived` flag
3. Typo in name → Use `notion-cli db list` to see all names

## Best Practices

### For End Users

1. **Run sync after install**: `notion-cli db sync`
2. **Use natural names**: "tasks" instead of long IDs
3. **Refresh periodically**: `notion-cli db sync` weekly
4. **Check cache age**: `notion-cli db cache stats`

### For AI Agents

1. **Cache warmup**: Run sync in initialization
2. **Error handling**: Catch "not found" and suggest sync
3. **Schema discovery**: Use cache for fast property lookup
4. **Batch operations**: Use cache to avoid repeated API calls

### For Scripts

1. **Check cache**: Verify cache exists before running
2. **Handle staleness**: Auto-sync if cache too old
3. **Parallel execution**: Share cache across processes
4. **Logging**: Log resolved IDs for debugging

## Migration Guide

### From ID-based to Name-based

**Before:**
```bash
# Hard-coded IDs in scripts
notion-cli db query 1fb79d4c71bb8032b722c82305b63a00
notion-cli page create -d 2a8c3d5e71bb8042b833d94316c74b11
```

**After:**
```bash
# Use descriptive names
notion-cli db query "tasks"
notion-cli page create -d "meeting notes"
```

### Updating Existing Scripts

1. Run `notion-cli db sync` to build cache
2. Replace hard-coded IDs with names
3. Add error handling for "not found"
4. Test with `--json` output

## Advanced Usage

### Custom Fuzzy Threshold

```bash
# Stricter matching (80% similarity required)
export NOTION_CLI_DB_FUZZY_THRESHOLD=0.8

# Looser matching (50% similarity required)
export NOTION_CLI_DB_FUZZY_THRESHOLD=0.5
```

### Sync Scheduling

```bash
# Cron job to sync daily at 2am
0 2 * * * /usr/local/bin/notion-cli db sync >> /var/log/notion-sync.log 2>&1

# systemd timer (Linux)
# Create /etc/systemd/system/notion-sync.service
[Unit]
Description=Notion Database Cache Sync

[Service]
Type=oneshot
ExecStart=/usr/local/bin/notion-cli db sync
User=your-user
Environment=NOTION_TOKEN=ntn_...

# Create /etc/systemd/system/notion-sync.timer
[Unit]
Description=Notion Database Cache Sync Timer

[Timer]
OnCalendar=daily
Persistent=true

[Install]
WantedBy=timers.target
```

### Multi-User Setup

```bash
# System-wide cache (Linux)
export NOTION_CLI_DB_CACHE_PATH="/var/cache/notion-cli"
sudo mkdir -p /var/cache/notion-cli
sudo chown your-user:your-group /var/cache/notion-cli
sudo chmod 775 /var/cache/notion-cli

# Per-user cache (default)
export NOTION_CLI_DB_CACHE_PATH="~/.notion-cli"
```

## Debugging

### Enable Debug Logging

```bash
DEBUG=true notion-cli db query "tasks" --json
```

**Output:**
```
[DB-CACHE] Loading cache from disk...
[DB-CACHE] Cache file: ~/.notion-cli/databases.json
[DB-CACHE] Cache age: 5 minutes (TTL: 60 minutes)
[DB-CACHE] Searching for: "tasks"
[DB-CACHE] Exact match: tasks (alias)
[DB-CACHE] Resolved: tasks → 1fb79d4c71bb8032b722c82305b63a00
[NOTION] Querying database: 1fb79d4c71bb8032b722c82305b63a00
[CACHE] Cache HIT: dataSource:1fb79d4c71bb8032b722c82305b63a00
```

### Inspect Cache

```bash
# Pretty-print cache
cat ~/.notion-cli/databases.json | jq .

# Show database titles
cat ~/.notion-cli/databases.json | jq -r '.databases[].title'

# Show aliases
cat ~/.notion-cli/databases.json | \
  jq -r '.databases[] | "\(.title): \(.aliases | join(", "))"'

# Find database by ID
cat ~/.notion-cli/databases.json | \
  jq '.databases[] | select(.id == "1fb79d4c...")'
```

## API Reference

See [CACHING-ARCHITECTURE.md](./CACHING-ARCHITECTURE.md) for complete technical documentation.

---

**Full Documentation**: [CACHING-ARCHITECTURE.md](./CACHING-ARCHITECTURE.md)
**Main README**: [../README.md](../README.md)
**AI Agent Cookbook**: [AI-AGENT-COOKBOOK.md](./AI-AGENT-COOKBOOK.md)
