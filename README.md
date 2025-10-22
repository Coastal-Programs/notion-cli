# notion-cli

> Notion CLI for AI Agents & Automation (API v5.2.1)

A non-interactive command-line interface for Notion's API, optimized for AI coding assistants and automation scripts. **Skip the MCP server** - get direct, fast access to Notion via CLI.

**Why Use This Instead of MCP Servers:**
- 🚀 **Faster**: Direct API calls without MCP overhead
- 🤖 **AI-First**: JSON output mode, structured errors, exit codes
- ⚡ **Non-Interactive**: No prompts - perfect for automation
- 📊 **Flexible Output**: JSON, CSV, YAML, or raw API responses
- ✅ **Latest API**: Notion API v5.2.1 with data sources support
- 🔄 **Enhanced Reliability**: Automatic retry with exponential backoff
- ⚡ **High Performance**: In-memory caching for faster operations

## What's New in v5.1.0

### Enhanced Retry Logic
- **Exponential backoff** with jitter to prevent thundering herd
- **Intelligent error categorization** (retryable vs non-retryable)
- **Automatic rate limit handling** with Retry-After header support
- **Circuit breaker pattern** for resilient operations
- **Configurable via environment variables**

### Caching Layer
- **In-memory caching** for frequently accessed resources
- **TTL-based expiration** (data sources: 10min, users: 1hr)
- **Automatic invalidation** on write operations
- **Cache statistics** for monitoring performance
- **Up to 100x faster** for repeated reads

[📖 Full Enhancement Documentation](./ENHANCEMENTS.md) | [📊 Output Formats Guide](./OUTPUT_FORMATS.md)

## Quick Start

```sh
$ npm install -g Coastal-Programs/notion-cli
$ export NOTION_TOKEN=secret_xxx...
$ notion-cli page retrieve <PAGE_ID> --json
```

* **Get your `NOTION_TOKEN`**: https://developers.notion.com/docs/create-a-notion-integration
* **Find `PAGE_ID`**: In page URL `https://www.notion.so/Page-title-<PAGE_ID>`

## Installation

**npm (Recommended):**
```sh
npm install -g Coastal-Programs/notion-cli
```

**npx (No install):**
```sh
npx Coastal-Programs/notion-cli page retrieve <PAGE_ID> --json
```

## Performance & Reliability Configuration

### Zero Configuration (Recommended)
Works great out of the box with:
- Automatic retry on failures (up to 3 attempts)
- Smart caching for schemas and metadata
- Exponential backoff with jitter

### Custom Configuration
Create a `.env` file for your use case:

```bash
# Production setup (balanced)
NOTION_CLI_MAX_RETRIES=5
NOTION_CLI_CACHE_ENABLED=true
NOTION_CLI_CACHE_MAX_SIZE=2000

# Read-heavy workload (maximum caching)
NOTION_CLI_CACHE_DS_TTL=1800000  # 30 minutes
NOTION_CLI_CACHE_USER_TTL=7200000  # 2 hours

# Unstable network (aggressive retry)
NOTION_CLI_MAX_RETRIES=10
NOTION_CLI_BASE_DELAY=2000
NOTION_CLI_MAX_DELAY=60000

# Memory-constrained (minimal cache)
NOTION_CLI_CACHE_MAX_SIZE=100
```

See [.env.example](./.env.example) for all options.

## Key Concepts: Database vs Data Source

In Notion API v5.2.1 (2025-09-03), terminology changed:

**OLD (pre-2025-09-03):**
- Database = Table with schema + rows

**NEW (v5.2.1+):**
- **Database** = Container that can hold multiple data sources
- **Data Source** = Individual table with schema (properties) and content (pages/rows)

**Commands use data source IDs** for querying/updating tables.

## Common Commands

### Query a Data Source (Table)
Get pages/rows from a table with optional filters:
```bash
# Basic query
notion-cli db query <DATA_SOURCE_ID> --json

# With filter (JSON)
notion-cli db query <DATA_SOURCE_ID> --filter '{"property":"Status","select":{"equals":"Done"}}' --json

# With sorting
notion-cli db query <DATA_SOURCE_ID> --sorts '[{"property":"Name","direction":"ascending"}]' --json
```

### Retrieve Data Source Schema
Get table properties and metadata (cached for 10 minutes by default):
```bash
notion-cli db retrieve <DATA_SOURCE_ID> --json
```

### Update Data Source
Change table title or properties:
```bash
notion-cli db update <DATA_SOURCE_ID> --title "Updated Table Name" --json
```

### Create Database with Data Source
Creates a new database container with initial table:
```bash
notion-cli db create <PARENT_PAGE_ID> --title "My Database" --json
```

### Create Page in Data Source (Table Row)
```bash
# In a data source (table)
notion-cli page create -d <DATA_SOURCE_ID> --json

# From markdown file
notion-cli page create -f ./content.md -d <DATA_SOURCE_ID> --json

# Custom title property name (if not "Name")
notion-cli page create -f ./content.md -d <DATA_SOURCE_ID> -t "Title" --json
```

### Create Sub-Page
```bash
notion-cli page create -p <PARENT_PAGE_ID> --json
```

### Retrieve Page
```bash
notion-cli page retrieve <PAGE_ID> --json
```

### Update Page
```bash
# Archive a page
notion-cli page update <PAGE_ID> --archive --json

# Unarchive a page
notion-cli page update <PAGE_ID> --unarchive --json
```

### Append Blocks to Page
```bash
notion-cli block append -b <BLOCK_ID> -c '[{"paragraph":{"rich_text":[{"text":{"content":"Hello"}}]}}]' --json
```

### Update Block
```bash
# Archive block
notion-cli block update <BLOCK_ID> --archived true --json

# Update block content (JSON)
notion-cli block update <BLOCK_ID> --content '{"paragraph":{"rich_text":[{"text":{"content":"Updated"}}]}}' --json

# Update block color
notion-cli block update <BLOCK_ID> --color blue --json
```

### Search
```bash
notion-cli search --query "meeting notes" --json
```

### List Users
```bash
notion-cli user list --json
```

## Automation Features

### JSON Output Mode
Every command supports `--json` flag for structured output:
```bash
notion-cli db query <DATA_SOURCE_ID> --json
```

**Output:**
```json
{
  "success": true,
  "data": {
    "object": "list",
    "results": [...],
    "has_more": false,
    "next_cursor": null
  },
  "timestamp": "2025-10-22T10:30:00.000Z"
}
```

### Exit Codes
- `0`: Success
- `1`: Error (check JSON error output)

### Error Handling
Errors return structured JSON with `--json` flag:
```json
{
  "success": false,
  "error": {
    "code": "object_not_found",
    "message": "Could not find database with ID: abc123",
    "status": 404
  },
  "timestamp": "2025-10-22T10:30:00.000Z"
}
```

**Enhanced Error Recovery:**
- Automatic retry on rate limits (429) and server errors (500-504)
- Exponential backoff prevents API overwhelming
- Network errors (timeouts, connection resets) automatically retried

### Automation Flags

Available on all commands:
- `--json` / `-j`: JSON output (recommended for AI agents)
- `--raw` / `-r`: Raw Notion API response
- `--no-color`: Disable colored output
- `--output <format>`: Output format (`table`, `csv`, `json`, `yaml`)

## Output Formats

### Table (Default)
```sh
$ notion-cli page retrieve c77dbaf240174ea1ac1e93a87269f3ea
 Title      Object Id                                   Url
 ────────── ────── ──────────────────────────────────── ─────────────────────────────────────────────────────────────────
 Page title page   c77dbaf2-4017-4ea1-ac1e-93a87269f3ea https://www.notion.so/Page-title-c77dbaf240174ea1ac1e93a87269f3ea
```

### CSV
```sh
$ notion-cli page retrieve c77dbaf240174ea1ac1e93a87269f3ea --output csv
Title,Object,Id,Url
Page title,page,c77dbaf2-4017-4ea1-ac1e-93a87269f3ea,https://www.notion.so/Page-title-c77dbaf240174ea1ac1e93a87269f3ea
```

### JSON (Formatted)
```sh
$ notion-cli page retrieve c77dbaf240174ea1ac1e93a87269f3ea --output json
[
  {
    "title": "Page title",
    "object": "page",
    "id": "c77dbaf2-4017-4ea1-ac1e93a87269f3ea",
    "url": "https://www.notion.so/Page-title-c77dbaf240174ea1ac1e93a87269f3ea"
  }
]
```

### YAML
```sh
$ notion-cli page retrieve c77dbaf240174ea1ac1e93a87269f3ea --output yaml
- title: Page title
  object: page
  id: c77dbaf2-4017-4ea1-ac1e93a87269f3ea
  url: 'https://www.notion.so/Page-title-c77dbaf240174ea1ac1e93a87269f3ea'
```

### Raw API Response
Get the complete Notion API response (pipe to `jq` for processing):
```sh
$ notion-cli page retrieve c77dbaf240174ea1ac1e93a87269f3ea --raw | jq
{
  "object": "page",
  "id": "c77dbaf2-4017-4ea1-ac1e93a87269f3ea",
  "created_time": "2023-05-07T09:08:00.000Z",
  "last_edited_time": "2023-08-15T01:08:00.000Z",
  "created_by": {
    "object": "user",
    "id": "3555ae80-4588-4514-bb6b-2ece534157de"
  },
  ...
}
```

## Available Commands

### Data Source (Database) Commands
- `notion-cli db create <PAGE_ID>` - Create database with initial data source
- `notion-cli db query <DATA_SOURCE_ID>` - Query data source (table) for pages
- `notion-cli db retrieve <DATA_SOURCE_ID>` - Get data source schema (cached)
- `notion-cli db update <DATA_SOURCE_ID>` - Update data source title/properties

**Aliases:**
- `db:*` commands also available as `ds:*` (data-source)
- `db:r` = `db:retrieve` = `ds:r` = `ds:retrieve`
- `db:u` = `db:update` = `ds:u` = `ds:update`

### Page Commands
- `notion-cli page create` - Create page in data source or as sub-page
- `notion-cli page retrieve <PAGE_ID>` - Get page details (cached)
- `notion-cli page update <PAGE_ID>` - Update page (archive/unarchive)

### Block Commands
- `notion-cli block append` - Append blocks to parent
- `notion-cli block delete <BLOCK_ID>` - Delete block
- `notion-cli block retrieve <BLOCK_ID>` - Get block details (cached)
- `notion-cli block update <BLOCK_ID>` - Update block content/color/archive

### User Commands
- `notion-cli user list` - List all users (cached for 1 hour)
- `notion-cli user retrieve <USER_ID>` - Get user details (cached)
- `notion-cli user me` - Get current bot user (cached)

### Search Commands
- `notion-cli search --query <QUERY>` - Search pages by title

### Help
- `notion-cli help` - Show all commands
- `notion-cli help <command>` - Show command help

## Use Cases for AI Agents

### 1. Knowledge Base Queries
AI agents can query your Notion workspace:
```bash
notion-cli search --query "API documentation" --json
```

### 2. Task Management
Create and update tasks programmatically:
```bash
notion-cli page create -d <TASKS_DATA_SOURCE_ID> --json
notion-cli page update <TASK_PAGE_ID> --archive --json
```

### 3. Content Generation
Generate content with AI, save to Notion:
```bash
echo "# Generated Content" > output.md
notion-cli page create -f output.md -p <PARENT_PAGE_ID> --json
```

### 4. Data Extraction
Extract structured data for AI processing:
```bash
notion-cli db query <DATA_SOURCE_ID> --json | jq '.data.results'
```

### 5. Workflow Automation
Chain commands in scripts:
```bash
#!/bin/bash
RESULT=$(notion-cli db query $DATA_SOURCE_ID --json)
if [ $? -eq 0 ]; then
  echo "$RESULT" | jq '.data.results[] | .properties'
fi
```

## Performance Benchmarks

| Scenario | Before v5.1.0 | After v5.1.0 | Improvement |
|----------|---------------|--------------|-------------|
| Repeat DB schema reads (100x) | 33s | 0.3s | **110x faster** |
| Bulk operations (50x) | 15s | 6s | **2.5x faster** |
| Rate limit success rate | 90% | 99.9% | **11x more reliable** |
| Poor network success rate | 95% | 99.5% | **9x fewer failures** |

*Results with default configuration. See [ENHANCEMENTS.md](./ENHANCEMENTS.md) for details.*

## API Coverage

Endpoint | Supported | JSON Output | Cached
-- | -- | -- | --
**Blocks** |  |  |
Append block children | ✅ | ✅ | -
Retrieve block | ✅ | ✅ | ✅ (30s)
Retrieve block children | ✅ | ✅ | ✅ (30s)
Update block | ✅ | ✅ | -
Delete block | ✅ | ✅ | -
**Pages** |  |  |
Create page | ✅ | ✅ | -
Retrieve page | ✅ | ✅ | ✅ (1m)
Update page | ✅ | ✅ | -
**Data Sources** |  |  |
Create database | ✅ | ✅ | -
Retrieve data source | ✅ | ✅ | ✅ (10m)
Update data source | ✅ | ✅ | -
Query data source | ✅ | ✅ | -
**Users** |  |  |
List users | ✅ | ✅ | ✅ (1h)
Retrieve user | ✅ | ✅ | ✅ (1h)
Retrieve bot user | ✅ | ✅ | ✅ (1h)
**Search** |  |  |
Search by title | ✅ | ✅ | -
**Comments** |  |  |
Create comment | ❌ | - | -
Retrieve comments | ❌ | - | -

## Requirements

- **Node.js**: >=18.0.0 (required by @notionhq/client v5.2.1)
- **Notion Integration**: Create at https://developers.notion.com/

## Environment Variables

```sh
# Required
export NOTION_TOKEN=secret_xxx...

# Optional - Debug
export DEBUG=true  # Enable debug logging (shows cache hits/misses, retries)

# Optional - Retry Configuration
export NOTION_CLI_MAX_RETRIES=3        # Max retry attempts
export NOTION_CLI_BASE_DELAY=1000      # Base delay in ms
export NOTION_CLI_MAX_DELAY=30000      # Max delay cap in ms

# Optional - Cache Configuration
export NOTION_CLI_CACHE_ENABLED=true   # Enable/disable cache
export NOTION_CLI_CACHE_MAX_SIZE=1000  # Max cached entries
export NOTION_CLI_CACHE_DS_TTL=600000  # Data source TTL (10 min)
export NOTION_CLI_CACHE_USER_TTL=3600000  # User TTL (1 hour)
```

See [.env.example](./.env.example) for complete configuration options.

## Documentation

- [Enhancement Documentation](./ENHANCEMENTS.md) - Complete reference for retry & caching
- [Output Formats Guide](./OUTPUT_FORMATS.md) - All available output formats
- [Configuration Examples](.env.example) - Environment variable templates
- [Code Examples](./src/examples/cache-retry-examples.ts) - Programmatic usage

## Author

**Jake Schepis** - [Coastal Programs](https://github.com/Coastal-Programs)

**Key Features:**
- ✅ Notion API v5.2.1 with data sources support
- ✅ Non-interactive CLI for automation and AI agents
- ✅ JSON output mode with structured responses
- ✅ Advanced error handling with exit codes
- ✅ Enhanced retry with exponential backoff
- ✅ Intelligent in-memory caching (up to 100x faster)
- ✅ Multiple output formats (JSON, CSV, YAML, table)
- ✅ Circuit breaker for resilient operations

## License

MIT

## Contributing

Issues and pull requests welcome at https://github.com/Coastal-Programs/notion-cli

## Resources

- [Notion API Documentation](https://developers.notion.com/)
- [Notion API v5.2.1 Upgrade Guide](https://developers.notion.com/docs/upgrade-guide-2025-09-03)
- [Create Notion Integration](https://developers.notion.com/docs/create-a-notion-integration)
