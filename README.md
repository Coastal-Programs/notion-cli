<div align="center">
<pre>
███╗   ██╗ ██████╗ ████████╗██╗ ██████╗ ███╗   ██╗     ██████╗██╗     ██╗
████╗  ██║██╔═══██╗╚══██╔══╝██║██╔═══██╗████╗  ██║    ██╔════╝██║     ██║
██╔██╗ ██║██║   ██║   ██║   ██║██║   ██║██╔██╗ ██║    ██║     ██║     ██║
██║╚██╗██║██║   ██║   ██║   ██║██║   ██║██║╚██╗██║    ██║     ██║     ██║
██║ ╚████║╚██████╔╝   ██║   ██║╚██████╔╝██║ ╚████║    ╚██████╗███████╗██║
╚═╝  ╚═══╝ ╚═════╝    ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝     ╚═════╝╚══════╝╚═╝
</pre>

<p align="center">
  <a href="https://github.com/Coastal-Programs/notion-cli/actions/workflows/ci.yml">
    <img src="https://github.com/Coastal-Programs/notion-cli/actions/workflows/ci.yml/badge.svg" alt="CI/CD Pipeline">
  </a>
  <a href="https://codecov.io/gh/Coastal-Programs/notion-cli">
    <img src="https://codecov.io/gh/Coastal-Programs/notion-cli/branch/main/graph/badge.svg" alt="Code Coverage">
  </a>
  <a href="https://www.npmjs.com/package/@coastal-programs/notion-cli">
    <img src="https://img.shields.io/npm/v/@coastal-programs/notion-cli.svg" alt="npm version">
  </a>
  <a href="https://nodejs.org">
    <img src="https://img.shields.io/badge/node-%3E%3D22.0.0-brightgreen.svg" alt="Node.js Version">
  </a>
  <a href="https://github.com/Coastal-Programs/notion-cli/blob/main/LICENSE">
    <img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License">
  </a>
</p>
</div>

**IMPORTANT NOTICE:**

This is an independent, unofficial command-line tool for working with Notion's API.
This project is not affiliated with, endorsed by, or sponsored by Notion Labs, Inc.
"Notion" is a registered trademark of Notion Labs, Inc.

> Notion CLI for AI Agents & Automation (v5.9.0 with 5-Phase Performance Optimization)

A powerful command-line interface for Notion's API, optimized for AI coding assistants and automation scripts. Now with comprehensive performance optimizations delivering 1.5-2x improvement for batch operations.

**Key Features:**
- 🚀 **Fast & Direct**: Native API integration with intelligent caching
- 🤖 **AI-First Design**: JSON output mode, structured errors, exit codes
- ⚡ **Non-Interactive**: Perfect for scripts and automation
- 📊 **Flexible Output**: JSON, CSV, YAML, or raw API responses
- ✅ **Latest API**: Notion API v5.2.1 with data sources support
- 🔄 **Enhanced Reliability**: Automatic retry with exponential backoff
- ⚡ **High Performance**: 5-phase optimization (1.5-2x improvement)
  - Request deduplication, parallel operations, disk cache, keep-alive, compression
- 🔍 **Schema Discovery**: AI-friendly database schema extraction
- 🗄️ **Workspace Caching**: Fast database lookups without API calls
- 🧠 **Smart ID Resolution**: Automatic database_id → data_source_id conversion
- 🔒 **Secure**: 0 production vulnerabilities

## What's New in v5.6.0

**Improved First-Time Experience & Enhanced Security** 🎉

### 1. Interactive Setup Wizard
- **NEW `notion-cli init` command** - Guided first-time setup in 3 easy steps
- **Token configuration** - Clear instructions for getting and setting your API token
- **Connection testing** - Verify API access before proceeding
- **Workspace sync** - Optional local caching for faster operations
- **Supports `--json`** - Automation-friendly for CI/CD environments

### 2. Health Check & Diagnostics
- **NEW `notion-cli doctor` command** - Comprehensive health checks (aliases: `diagnose`, `healthcheck`)
- **7 diagnostic checks** - Node version, token config, API connection, workspace access, cache status, dependencies, file permissions
- **Color-coded output** - Clear pass/fail indicators with actionable recommendations
- **JSON support** - Monitor system health programmatically
- **Perfect for troubleshooting** - Quickly identify and fix issues

### 3. Enhanced Error Handling
- **Token validator** - Early validation before API calls (500x faster error feedback)
- **Platform-specific help** - Tailored instructions for Windows (CMD/PowerShell) and Unix/Mac
- **Actionable messages** - Clear next steps instead of cryptic API errors
- **Prevents confusion** - No more "unauthorized" errors on first run

### 4. Post-Install Experience
- **Welcome message** - Friendly introduction after installation
- **Clear next steps** - Guides new users to run `notion-cli init`
- **Respects `--silent`** - Honors npm's silent flag for automation

### 5. Progress Indicators
- **Real-time feedback** - `sync` command shows progress during execution
- **Execution timing** - See how long operations take
- **Enhanced summaries** - Rich metadata about cache state and recommendations

### 6. Security Improvements
- **0 production vulnerabilities** - Fixed all 16 security issues
- **Removed vulnerable dependency** - Replaced @tryfabric/martian with custom markdown converter
- **Zero-dependency parser** - No external security risks from markdown processing
- **CVE fixes** - Addressed XSS, prototype pollution, and ReDoS vulnerabilities

---

## 🚀 Performance Optimizations (v5.9.0)

**5-Phase Performance Enhancement** delivering **1.5-2x overall improvement** for batch operations and repeated data access.

### Overview

Version 5.9.0 introduces comprehensive performance optimizations across five key areas:

| Optimization | Best Case | Typical Case | When It Helps |
|--------------|-----------|--------------|---------------|
| **Request Deduplication** | 30-50% fewer calls | 5-15% fewer calls | Concurrent duplicate requests |
| **Parallel Operations** | 80% faster | 60-70% faster | Batch deletions, bulk retrievals |
| **Persistent Disk Cache** | 60% better hits | 20-30% better hits | Repeated CLI sessions |
| **HTTP Keep-Alive** | 20% faster | 5-10% faster | Multi-request operations |
| **Response Compression** | 70% less bandwidth | Varies | Large JSON responses |

### Phase 1: Request Deduplication

Automatically prevents duplicate concurrent API calls using promise memoization.

**How it works:**
- Multiple concurrent requests for the same resource share a single API call
- In-flight request tracking with automatic cleanup
- Statistics tracking for monitoring effectiveness

**When it helps:**
- ✅ Parallel execution of commands that fetch the same data
- ✅ Applications making concurrent duplicate requests
- ❌ Sequential CLI commands (typical usage)

**Configuration:**
```bash
# Enable/disable (default: true)
NOTION_CLI_DEDUP_ENABLED=true
```

**Example:**
```bash
# Without deduplication: 3 API calls
# With deduplication: 1 API call (3 requests share result)
notion-cli page:retrieve PAGE_ID &
notion-cli page:retrieve PAGE_ID &
notion-cli page:retrieve PAGE_ID &
```

### Phase 2: Parallel Operations

Executes bulk operations concurrently with configurable concurrency limits.

**How it works:**
- Block deletions run in parallel (default: 5 concurrent)
- Child block fetching runs in parallel (default: 10 concurrent)
- Batch processing with retry logic and error handling
- Respects Notion API rate limits (3 req/sec per integration)

**When it helps:**
- ✅ `page:update` with many existing blocks
- ✅ Recursive page retrieval with many children
- ✅ Batch operations on multiple resources
- ❌ Single block operations

**Configuration:**
```bash
# Block deletion concurrency (default: 5)
NOTION_CLI_DELETE_CONCURRENCY=5

# Child fetching concurrency (default: 10)
NOTION_CLI_CHILDREN_CONCURRENCY=10
```

**Example:**
```bash
# Sequential: 10 blocks × 100ms = 1000ms
# Parallel (5 concurrent): ~200ms (5x faster)
notion-cli page:update PAGE_ID --file content.md
```

**Performance test results:**
```
✓ Should be significantly faster than sequential execution (607ms)
  Sequential: 500ms | Parallel: ~100ms | Speedup: 5x
```

### Phase 3: Persistent Disk Cache

Maintains cache across CLI invocations for improved hit rates.

**How it works:**
- Cache stored in `~/.notion-cli/cache/` directory
- Automatic TTL-based expiration
- LRU eviction when max size reached (default: 100MB)
- Atomic writes prevent corruption
- SHA-256 key hashing for safe filenames

**When it helps:**
- ✅ Running the same query multiple times
- ✅ Repeated `db:query` on same database
- ✅ Schema lookups across sessions
- ❌ Always-fresh data requirements
- ❌ Single-use queries

**Configuration:**
```bash
# Enable/disable (default: true)
NOTION_CLI_DISK_CACHE_ENABLED=true

# Max cache size in bytes (default: 100MB)
NOTION_CLI_DISK_CACHE_MAX_SIZE=104857600

# Sync interval in ms (default: 5s)
NOTION_CLI_DISK_CACHE_SYNC_INTERVAL=5000
```

**Example:**
```bash
# First run: API call + disk write
notion-cli db:query DB_ID  # 250ms

# Subsequent runs: Disk cache hit
notion-cli db:query DB_ID  # 50ms (5x faster)
```

### Phase 4: HTTP Keep-Alive & Connection Pooling

Reuses HTTPS connections to eliminate TLS handshake overhead.

**How it works:**
- Connection pool with configurable size (default: 10 free sockets)
- Keep-alive timeout: 60 seconds
- Max concurrent connections: 50
- Automatic cleanup on command exit

**When it helps:**
- ✅ Multi-request operations (e.g., batch queries)
- ✅ Long-running scripts
- ✅ Repeated API calls in quick succession
- ❌ Single request per session
- ⚠️ **Note**: Effectiveness depends on Notion SDK's HTTP client implementation

**Configuration:**
```bash
# Enable/disable (default: true)
NOTION_CLI_HTTP_KEEP_ALIVE=true

# Keep-alive timeout in ms (default: 60s)
NOTION_CLI_HTTP_KEEP_ALIVE_MS=60000

# Max concurrent connections (default: 50)
NOTION_CLI_HTTP_MAX_SOCKETS=50

# Connection pool size (default: 10)
NOTION_CLI_HTTP_MAX_FREE_SOCKETS=10

# Request timeout in ms (default: 30s)
NOTION_CLI_HTTP_TIMEOUT=30000
```

**Performance impact:**
- TLS handshake typically adds 50-100ms per connection
- With keep-alive: 1 handshake for multiple requests
- Savings: 5-10% typical, 10-20% best case

### Phase 5: Response Compression

Enables gzip, deflate, and brotli compression for API responses.

**How it works:**
- Adds `Accept-Encoding: gzip, deflate, br` header to requests
- Server decides whether to compress responses
- Client automatically decompresses (transparent)

**When it helps:**
- ✅ Large JSON responses (>10KB)
- ✅ Slow network connections
- ✅ Bandwidth-constrained environments
- ❌ Small responses (<1KB)
- ⚠️ **Note**: Notion API may already compress responses by default

**Configuration:**
- Always enabled, no configuration needed

**Compression ratios:**
- JSON typically compresses 60-70%
- Actual performance impact varies (likely already compressed)

---

### Combined Performance Impact

**Real-world scenarios:**

**Scenario 1: Batch Operations**
```bash
# Update 5 pages in parallel with cached schemas
# Expected improvement: 2-2.5x faster
notion-cli batch:update --input pages.json
```

**Scenario 2: Repeated Queries**
```bash
# Run same query multiple times
# First run: 300ms | Subsequent runs: 50ms (6x faster via disk cache)
notion-cli db:query DB_ID --filter '{"status": "active"}'
```

**Scenario 3: Typical CLI Usage**
```bash
# Sequential commands on unique data
# Expected improvement: 1.2-1.5x (disk cache + compression)
notion-cli page:retrieve PAGE_ID
notion-cli db:query DB_ID
```

### Configuration Best Practices

**Development (fast iteration):**
```bash
NOTION_CLI_CACHE_TTL=30000                    # 30s cache
NOTION_CLI_DISK_CACHE_ENABLED=true            # Keep disk cache
NOTION_CLI_DELETE_CONCURRENCY=3               # Conservative
DEBUG=true                                     # See optimization activity
```

**Production (balanced performance):**
```bash
NOTION_CLI_CACHE_TTL=300000                   # 5min cache
NOTION_CLI_DISK_CACHE_MAX_SIZE=104857600      # 100MB
NOTION_CLI_DELETE_CONCURRENCY=5               # Default
NOTION_CLI_CHILDREN_CONCURRENCY=10            # Default
NOTION_CLI_HTTP_KEEP_ALIVE=true               # Enabled
```

**Batch Processing (maximum throughput):**
```bash
NOTION_CLI_DELETE_CONCURRENCY=10              # Higher concurrency
NOTION_CLI_CHILDREN_CONCURRENCY=20            # Higher concurrency
NOTION_CLI_HTTP_MAX_SOCKETS=50                # More connections
NOTION_CLI_DISK_CACHE_ENABLED=true            # Cache results
```

**Memory-Constrained (minimal footprint):**
```bash
NOTION_CLI_CACHE_MAX_SIZE=100                 # Small memory cache
NOTION_CLI_DISK_CACHE_MAX_SIZE=10485760       # 10MB disk cache
NOTION_CLI_HTTP_MAX_FREE_SOCKETS=2            # Fewer pooled connections
```

### Monitoring Performance

**Check optimization statistics:**
```bash
# View cache statistics
notion-cli doctor --json | jq '.checks[] | select(.name | contains("cache"))'

# Enable verbose logging to see:
# - Cache hits/misses
# - Deduplication hits
# - Disk cache activity
DEBUG=true notion-cli db:query DB_ID
```

**Expected verbose output:**
```
Cache MISS: dataSource:abc123
Dedup MISS: dataSource:abc123
[API Call] GET /v1/databases/abc123
Cache SET: dataSource:abc123 (TTL: 600000ms)
Disk cache WRITE: dataSource:abc123
```

### Performance Testing

All optimizations are thoroughly tested with 121 comprehensive tests:
- ✅ 22 deduplication tests (94.73% coverage)
- ✅ 21 parallel operations tests (timing benchmarks included)
- ✅ 34 disk cache tests (83.59% coverage)
- ✅ 26 HTTP agent tests (78.94% coverage)
- ✅ 18 compression tests

See [CHANGELOG.md](./CHANGELOG.md) for detailed implementation notes and [test directory](./test) for test suites.

---

### Earlier Features (v5.4.0)

**7 Major AI Agent Usability Features** (Issue #4)

### 1. Simple Properties Mode
- **NEW `--simple-properties` (`-S`) flag** - Use flat JSON instead of complex nested structures
- **70% complexity reduction** - `{"Name": "Task", "Status": "Done"}` vs verbose Notion format
- **13 property types supported** - title, rich_text, number, checkbox, select, multi_select, status, date, url, email, phone, people, files, relation
- **Case-insensitive matching** - Property names and select values work regardless of case
- **Relative dates** - Use `"today"`, `"tomorrow"`, `"+7 days"`, `"+2 weeks"`, etc.
- **Smart validation** - Helpful error messages with suggestions

[📖 Simple Properties Guide](./docs/SIMPLE_PROPERTIES.md) | [⚡ Quick Reference](./AI_AGENT_QUICK_REFERENCE.md)

### 2. JSON Envelope Standardization
- **Consistent response format** - All commands return `{success, data, metadata}`
- **Standardized exit codes** - 0 = success, 1 = API error, 2 = CLI error
- **Predictable parsing** - AI agents can reliably extract data

[📖 Envelope Documentation](./docs/ENVELOPE_INDEX.md)

### 3. Health Check Command
- **NEW `whoami` command** - Verify connectivity before operations (aliases: `test`, `health`)
- **Reports** - Bot info, workspace access, cache status, API latency
- **Error diagnostics** - Comprehensive troubleshooting suggestions

### 4. Schema Examples
- **NEW `--with-examples` flag** - Get copy-pastable property payloads
- **Works with `db schema`** - Shows example values for each property type
- **Groups properties** - Separates writable vs read-only

### 5. Verbose Logging
- **NEW `--verbose` (`-v`) flag** - Debug mode for troubleshooting
- **Shows** - Cache hits/misses, retry attempts, API latency
- **Helps AI agents** - Understand what's happening behind the scenes

[📖 Verbose Logging Guide](./docs/VERBOSE_LOGGING.md)

### 6. Filter Simplification
- **Improved filter syntax** - Easier database query filters
- **Better validation** - Clear error messages for invalid filters

[📖 Filter Guide](./docs/FILTER_GUIDE.md)

### 7. Output Format Enhancements
- **NEW `--compact-json`** - Minified single-line JSON output
- **NEW `--pretty`** - Enhanced table formatting
- **NEW `--markdown`** - Markdown table output

[📊 Output Formats Guide](./OUTPUT_FORMATS.md)

---

### Earlier Features (v5.2-5.3)

**Smart ID Resolution** - Automatic `database_id` ↔ `data_source_id` conversion • [Guide](./docs/smart-id-resolution.md)

**Workspace Caching** - `sync` and `list` commands for local database cache

**Schema Discovery** - `db schema` command for AI-parseable schemas • [AI Agent Cookbook](./docs/AI-AGENT-COOKBOOK.md)

**Enhanced Reliability** - Exponential backoff retry + circuit breaker • [Details](./ENHANCEMENTS.md)

**Performance** - In-memory caching (up to 100x faster for repeated reads)

## Quick Start

### Installation

```bash
# From npm (recommended)
npm install -g @coastal-programs/notion-cli

# Or from source
npm install -g Coastal-Programs/notion-cli
```

**Note**: Windows users installing from source should use the local clone method due to symlink limitations:
```bash
git clone https://github.com/Coastal-Programs/notion-cli
cd notion-cli
npm install
npm run build
npm link
```

### Updating

To update to the latest version:

```bash
# Update to latest version
npm update -g @coastal-programs/notion-cli

# Or reinstall for a specific version
npm install -g @coastal-programs/notion-cli@latest
```

Check your current version:
```bash
notion-cli --version
```

**Update Notifications:**
- The CLI automatically checks for updates once per day
- You'll see a notification when a new version is available
- Updates are never applied automatically - you stay in control
- To disable notifications: `export NO_UPDATE_NOTIFIER=1`

Check for new releases at:
- [npm package page](https://www.npmjs.com/package/@coastal-programs/notion-cli)
- [GitHub releases](https://github.com/Coastal-Programs/notion-cli/releases)

### First-Time Setup

The easiest way to get started:

```bash
# Run the interactive setup wizard
notion-cli init
```

This will guide you through:
1. 🔑 Setting your Notion API token
2. ✅ Testing the connection
3. 🔄 Syncing your workspace

**Manual Configuration (Optional):**

If you prefer to set up manually:

```bash
# Set your token
export NOTION_TOKEN="secret_your_token_here"

# Test the connection
notion-cli whoami

# Sync your workspace
notion-cli sync
```

### Common Commands

**List your databases:**
```bash
notion-cli list --json
```

**Discover database schema:**
```bash
# Get schema with examples for easy copy-paste
notion-cli db schema <DATA_SOURCE_ID> --with-examples --json
```

**Create a page** (using simple properties):
```bash
notion-cli page create -d <DATA_SOURCE_ID> -S --properties '{
  "Name": "My Task",
  "Status": "In Progress",
  "Priority": 5,
  "Due Date": "tomorrow"
}'
```

**All commands support** `--json` for machine-readable responses.

**Get your API token**: https://developers.notion.com/docs/create-a-notion-integration

## Key Features for AI Agents

### Simple Properties - 70% Less Complexity
Create and update Notion pages with flat JSON instead of complex nested structures:

```bash
# ❌ OLD WAY: Complex nested structure (error-prone)
notion-cli page create -d DB_ID --properties '{
  "Name": {
    "title": [{"text": {"content": "Task"}}]
  },
  "Status": {
    "select": {"name": "In Progress"}
  },
  "Priority": {
    "number": 5
  },
  "Tags": {
    "multi_select": [
      {"name": "urgent"},
      {"name": "bug"}
    ]
  }
}'

# ✅ NEW WAY: Simple properties with -S flag
notion-cli page create -d DB_ID -S --properties '{
  "Name": "Task",
  "Status": "In Progress",
  "Priority": 5,
  "Tags": ["urgent", "bug"],
  "Due Date": "tomorrow"
}'

# Update is just as easy
notion-cli page update PAGE_ID -S --properties '{
  "Status": "Done",
  "Completed": true
}'
```

**Features:**
- 🔤 **Case-insensitive** - Property names and select values work regardless of case
- 📅 **Relative dates** - Use `"today"`, `"tomorrow"`, `"+7 days"`, `"+2 weeks"`
- ✅ **Smart validation** - Clear error messages with valid options listed
- 🎯 **13 property types** - title, rich_text, number, checkbox, select, multi_select, status, date, url, email, phone, people, files, relation

[📖 Simple Properties Guide](./docs/SIMPLE_PROPERTIES.md) | [⚡ Quick Reference](./AI_AGENT_QUICK_REFERENCE.md)

### Smart Database ID Resolution
No need to worry about `database_id` vs `data_source_id` confusion anymore! The CLI automatically detects and converts between them:

```bash
# Both work! Use whichever ID you have
notion-cli db retrieve 1fb79d4c71bb8032b722c82305b63a00  # database_id
notion-cli db retrieve 2gc80e5d82cc9043c833d93416c74b11  # data_source_id

# When conversion happens, you'll see:
# Info: Resolved database_id to data_source_id
#   database_id:    1fb79d4c71bb8032b722c82305b63a00
#   data_source_id: 2gc80e5d82cc9043c833d93416c74b11
```

[📖 Learn more about Smart ID Resolution](./docs/smart-id-resolution.md)

### JSON Mode - Perfect for AI Processing
Every command supports `--json` for structured, parseable output:

```bash
# Get structured data
notion-cli db query <ID> --json | jq '.data.results[].properties'

# Error responses are also JSON
notion-cli db retrieve invalid-id --json
# {
#   "success": false,
#   "error": {
#     "code": "NOT_FOUND",
#     "message": "Database not found"
#   }
# }
```

### Schema Discovery - Know Your Data Structure
Extract complete database schemas in AI-friendly formats:

```bash
# Get full schema
notion-cli db schema <DATA_SOURCE_ID> --json

# Output:
# {
#   "database_id": "...",
#   "title": "Tasks",
#   "properties": {
#     "Name": { "type": "title", "required": true },
#     "Status": {
#       "type": "select",
#       "options": ["Not Started", "In Progress", "Done"]
#     }
#   }
# }

# Filter to specific properties
notion-cli db schema <ID> --properties Status,Priority --yaml
```

### Workspace Caching - Zero API Calls for Lookups
Cache your entire workspace locally for instant database lookups:

```bash
# One-time sync
notion-cli sync

# Now use database names instead of IDs
notion-cli db query "Tasks Database" --json

# Browse all cached databases
notion-cli list --json
```

### Cache Management - AI-Friendly Metadata
AI agents need to know when data is fresh. Get machine-readable cache metadata:

```bash
# Check cache status and TTLs
notion-cli cache:info --json

# Sample output:
# {
#   "success": true,
#   "data": {
#     "in_memory": {
#       "enabled": true,
#       "stats": { "hits": 42, "misses": 8, "hit_rate": 84.0 },
#       "ttls_ms": {
#         "data_source": 600000,  // 10 minutes
#         "page": 60000,          // 1 minute
#         "user": 3600000,        // 1 hour
#         "block": 30000          // 30 seconds
#       }
#     },
#     "workspace": {
#       "last_sync": "2025-10-23T14:30:00.000Z",
#       "cache_age_hours": 2.5,
#       "is_stale": false,
#       "databases_cached": 15
#     },
#     "recommendations": {
#       "sync_interval_hours": 24,
#       "next_sync": "2025-10-24T14:30:00.000Z",
#       "action_needed": "Cache is fresh"
#     }
#   }
# }

# List databases with cache age metadata
notion-cli list --json

# Sync with comprehensive metadata
notion-cli sync --json
```

**Cache TTLs:**
- **Workspace cache**: Persists until next `sync` (recommended: every 24 hours)
- **In-memory cache**:
  - Data sources: 10 minutes (schemas rarely change)
  - Pages: 1 minute (frequently updated)
  - Users: 1 hour (very stable)
  - Blocks: 30 seconds (most dynamic)

**AI Agent Best Practices:**
1. Run `cache:info --json` to check freshness before bulk operations
2. Parse `is_stale` flag to decide whether to re-sync
3. Use `cache_age_hours` for smart caching decisions
4. Respect TTL metadata when planning repeated reads


### Exit Codes - Script-Friendly
```bash
notion-cli db retrieve <ID> --json
if [ $? -eq 0 ]; then
  echo "Success!"
else
  echo "Failed!"
fi
```

- `0` = Success
- `1` = Notion API error
- `2` = CLI error (invalid flags, etc.)

## Core Commands

### Setup & Diagnostics

```bash
# First-time setup wizard
notion-cli init

# Health check and diagnostics
notion-cli doctor

# Test connectivity
notion-cli whoami
```

### Database Commands

```bash
# Retrieve database metadata (works with any ID type!)
notion-cli db retrieve <DATABASE_ID>
notion-cli db retrieve <DATA_SOURCE_ID>
notion-cli db retrieve "Tasks"

# Query database with filters
notion-cli db query <ID> --json

# Update database properties
notion-cli db update <ID> --title "New Title"

# Create new database
notion-cli db create \
  --parent-page <PAGE_ID> \
  --title "My Database" \
  --properties '{"Name": {"type": "title"}}'

# Extract schema
notion-cli db schema <ID> --json
```

### Page Commands

```bash
# Create page in database
notion-cli page create \
  --database-id <ID> \
  --properties '{"Name": {"title": [{"text": {"content": "Task"}}]}}'

# Retrieve page
notion-cli page retrieve <PAGE_ID> --json

# Update page properties
notion-cli page update <PAGE_ID> \
  --properties '{"Status": {"select": {"name": "Done"}}}'
```

### Block Commands

```bash
# Retrieve block
notion-cli block retrieve <BLOCK_ID>

# Append children to block
notion-cli block append <BLOCK_ID> \
  --children '[{"object": "block", "type": "paragraph", ...}]'

# Update block
notion-cli block update <BLOCK_ID> --content "Updated text"
```

### User Commands

```bash
# List all users
notion-cli user list --json

# Retrieve user
notion-cli user retrieve <USER_ID>

# Get bot user info
notion-cli user retrieve bot
```

### Search Command

```bash
# Search workspace
notion-cli search "project" --json

# Search with filters
notion-cli search "docs" --filter page
```

### Workspace Commands

```bash
# Sync workspace (cache all databases)
notion-cli sync

# List cached databases
notion-cli list --json

# Check cache status
notion-cli cache:info --json


# Configure token
notion-cli config set-token
```

## Database Query Filtering

Filter database queries with three powerful options optimized for AI agents and automation:

### JSON Filter (Primary Method - Recommended for AI)

Use `--filter` with JSON objects matching Notion's filter API format:

```bash
# Filter by select property
notion-cli db query <ID> \
  --filter '{"property": "Status", "select": {"equals": "Done"}}' \
  --json

# Complex AND filter
notion-cli db query <ID> \
  --filter '{"and": [{"property": "Status", "select": {"equals": "Done"}}, {"property": "Priority", "number": {"greater_than": 5}}]}' \
  --json

# OR filter for multiple conditions
notion-cli db query <ID> \
  --filter '{"or": [{"property": "Tags", "multi_select": {"contains": "urgent"}}, {"property": "Tags", "multi_select": {"contains": "bug"}}]}' \
  --json

# Date filter
notion-cli db query <ID> \
  --filter '{"property": "Due Date", "date": {"on_or_before": "2025-12-31"}}' \
  --json
```

### Text Search (Human Convenience)

Use `--search` for simple text matching across common properties:

```bash
# Quick text search (searches Name, Title, Description)
notion-cli db query <ID> --search "urgent" --json

# Case-sensitive matching
notion-cli db query <ID> --search "Project Alpha" --json
```

### File Filter (Complex Queries)

Use `--file-filter` to load complex filters from JSON files:

```bash
# Create filter file
cat > high-priority-filter.json << 'EOF'
{
  "and": [
    {"property": "Status", "select": {"equals": "In Progress"}},
    {"property": "Priority", "number": {"greater_than_or_equal_to": 8}},
    {"property": "Assigned To", "people": {"is_not_empty": true}}
  ]
}
EOF

# Use filter file
notion-cli db query <ID> --file-filter ./high-priority-filter.json --json
```

### Common Filter Examples

**Find completed high-priority tasks:**
```bash
notion-cli db query <ID> \
  --filter '{"and": [{"property": "Status", "select": {"equals": "Done"}}, {"property": "Priority", "number": {"greater_than": 7}}]}' \
  --json
```

**Find items due this week:**
```bash
notion-cli db query <ID> \
  --filter '{"property": "Due Date", "date": {"next_week": {}}}' \
  --json
```

**Find unassigned tasks:**
```bash
notion-cli db query <ID> \
  --filter '{"property": "Assigned To", "people": {"is_empty": true}}' \
  --json
```

**Find items without attachments:**
```bash
notion-cli db query <ID> \
  --filter '{"property": "Attachments", "files": {"is_empty": true}}' \
  --json
```

[📖 Full Filter Guide with Examples](./docs/FILTER_GUIDE.md)

## Output Formats

All commands support multiple output formats:

```bash
# JSON (default for --json flag)
notion-cli db query <ID> --json

# Compact JSON (single-line)
notion-cli db query <ID> --compact-json

# Markdown table
notion-cli db query <ID> --markdown

# Pretty table (with borders)
notion-cli db query <ID> --pretty

# Raw API response
notion-cli db query <ID> --raw
```

[📊 Full Output Formats Guide](./OUTPUT_FORMATS.md)

## Environment Variables

### Authentication
```bash
NOTION_TOKEN=secret_your_token_here
```

### Retry Configuration
```bash
NOTION_RETRY_MAX_ATTEMPTS=3        # Max retry attempts (default: 3)
NOTION_RETRY_INITIAL_DELAY=1000    # Initial delay in ms (default: 1000)
NOTION_RETRY_MAX_DELAY=30000       # Max delay in ms (default: 30000)
NOTION_RETRY_TIMEOUT=60000         # Request timeout in ms (default: 60000)
```

### Circuit Breaker
```bash
NOTION_CB_FAILURE_THRESHOLD=5      # Failures before opening (default: 5)
NOTION_CB_SUCCESS_THRESHOLD=2      # Successes to close (default: 2)
NOTION_CB_TIMEOUT=60000            # Reset timeout in ms (default: 60000)
```

### Caching
```bash
NOTION_CACHE_DISABLED=true         # Disable all caching
```

### Debug Mode
```bash
DEBUG=notion-cli:*                 # Enable debug logging
```

### Verbose Logging
```bash
# Enable structured event logging to stderr
NOTION_CLI_VERBOSE=true            # Logs retry events, cache stats to stderr
NOTION_CLI_DEBUG=true              # Enables DEBUG + VERBOSE modes
```

**Verbose Mode** provides machine-readable JSON events to stderr for observability:
- Retry events (rate limits, backoff delays, exhaustion)
- Cache events (hits, misses, evictions)
- Circuit breaker state changes
- Never pollutes stdout JSON output

```bash
# Enable verbose logging for debugging
notion-cli db query <ID> --json --verbose 2>debug.log

# View retry events
cat debug.log | jq 'select(.event == "retry")'

# Monitor rate limiting
notion-cli db query <ID> --verbose 2>&1 >/dev/null | jq 'select(.reason == "RATE_LIMITED")'
```

[📖 Full Verbose Logging Guide](./docs/VERBOSE_LOGGING.md)

## Real-World Examples

### Automated Task Management

```bash
#!/bin/bash
# Create and track a task

# Create task
TASK_ID=$(notion-cli page create \
  --database-id <TASKS_DB_ID> \
  --properties '{
    "Name": {"title": [{"text": {"content": "Review PR"}}]},
    "Status": {"select": {"name": "In Progress"}}
  }' \
  --json | jq -r '.data.id')

# Do work...
echo "Working on task: $TASK_ID"

# Mark complete
notion-cli page update $TASK_ID \
  --properties '{"Status": {"select": {"name": "Done"}}}' \
  --json
```

### Database Schema Migration

```bash
#!/bin/bash
# Export schema from one database, import to another

# Extract source schema
notion-cli db schema $SOURCE_DB --json > schema.json

# Parse and create new database
notion-cli db create \
  --parent-page $TARGET_PAGE \
  --title "Migrated Database" \
  --properties "$(jq '.properties' schema.json)" \
  --json
```

### Daily Sync Script

```bash
#!/bin/bash
# Sync workspace and generate report

# Refresh cache
notion-cli sync

# List all databases with stats
notion-cli list --json > databases.json

# Generate markdown report
echo "# Database Report - $(date)" > report.md
jq -r '.[] | "- **\(.title)** (\(.page_count) pages)"' databases.json >> report.md
```

## Performance Tips

1. **Use caching**: Run `notion-cli sync` before heavy operations
2. **Batch operations**: Combine multiple updates when possible
3. **Use --json**: Faster parsing than pretty output
4. **Filter early**: Use query filters to reduce data transfer
5. **Cache results**: Store query results for repeated access

## Troubleshooting

### Setup Issues

**Problem**: Not sure if everything is configured correctly

**Solution**: Run the health check command
```bash
notion-cli doctor
# Shows 7 diagnostic checks with clear pass/fail indicators
```

### "Database not found" Error

**Problem**: You're using a `database_id` instead of `data_source_id`

**Solution**: The CLI now auto-resolves this! But if it fails:
```bash
# Get the correct data_source_id
notion-cli page retrieve <PAGE_ID> --raw | jq '.parent.data_source_id'
```

### Rate Limiting

**Problem**: Getting 429 errors

**Solution**: The CLI handles this automatically with retry logic. To adjust:
```bash
export NOTION_RETRY_MAX_ATTEMPTS=5
export NOTION_RETRY_MAX_DELAY=60000
```

### Slow Queries

**Problem**: Database queries taking too long

**Solution**:
1. Use filters to reduce data: `--filter '{"property": "Status", "select": {"equals": "Active"}}'`
2. Enable caching: `notion-cli sync`
3. Use `--compact-json` for faster output

### Authentication Errors

**Problem**: 401 Unauthorized or token errors

**Solution**:
```bash
# Run the setup wizard
notion-cli init

# Or verify token is set
echo $NOTION_TOKEN

# Or manually configure token
notion-cli config set-token

# Check integration has access
# Visit: https://www.notion.so/my-integrations
```

## Development

### Prerequisites

- Node.js >= 22.0.0
- npm >= 8.0.0
- Git

### Setup

```bash
# Clone repository
git clone https://github.com/Coastal-Programs/notion-cli
cd notion-cli

# Install dependencies
npm install

# Build TypeScript
npm run build
```

### Development Workflow

```bash
# Build the project
npm run build

# Run tests
npm test

# Run linter
npm run lint

# Auto-fix linting issues
npm run lint -- --fix

# Link for local development (test CLI globally)
npm link
```

### Project Structure

```
notion-cli/
├── src/               # TypeScript source files
│   ├── commands/      # CLI command implementations
│   ├── utils/         # Utility functions
│   ├── base-command.ts   # Base command class
│   ├── base-flags.ts     # Reusable flag definitions
│   ├── envelope.ts       # JSON envelope formatting
│   ├── notion.ts         # Notion API client wrapper
│   └── cache.ts          # In-memory caching layer
├── test/              # Test files (mocha + chai)
├── dist/              # Compiled JavaScript (generated)
├── docs/              # Documentation
└── package.json       # Project configuration
```

### Testing

```bash
# Run all tests
npm test

# Run tests with verbose output
npm test -- --reporter spec

# Run specific test file
npm test -- test/commands/db/query.test.ts
```

### Code Quality

This project uses:
- **TypeScript** for type safety
- **ESLint v9** for code linting (flat config)
- **Prettier** for code formatting
- **Mocha + Chai** for testing

### Building and Publishing

```bash
# Build for production
npm run build

# Create package tarball (for testing)
npm pack

# Publish to npm (maintainers only)
npm publish
```

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Code style and conventions
- Test requirements
- Pull request process
- Commit message format

## Legal & Compliance

### Trademark Notice
"Notion" is a registered trademark of Notion Labs, Inc. This project is an independent,
unofficial tool and is not affiliated with, endorsed by, or sponsored by Notion Labs, Inc.

### License
This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

### Third-Party Licenses
This project uses various open-source dependencies. See [NOTICE](NOTICE) for complete
third-party license information.

## Contributing

Contributions welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Add tests for new features
4. Submit a pull request

## API Version

This CLI uses **Notion API v5.2.1** with full support for:
- Data sources (new database API)
- Enhanced properties
- Improved pagination
- Better error handling

## License

MIT License - see [LICENSE](LICENSE) file

## Support

- **Documentation**: Full guides in `/docs` folder
- **Issues**: Report bugs on GitHub Issues
- **Discussions**: Ask questions in GitHub Discussions
- **Examples**: See `/examples` folder for sample scripts

## Related Projects

- **Notion API**: https://developers.notion.com
- **@notionhq/client**: Official Notion SDK
- **notion-md**: Markdown converter for Notion

---

**Built for AI agents, optimized for automation, powered by Notion API v5.2.1**
