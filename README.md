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
- 🔍 **Schema Discovery**: AI-friendly database schema extraction

## What's New in v5.2.0

### NEW: Schema Discovery Command
- **`db schema` command** - Extract clean, AI-parseable database schemas
- **Property type detection** - Automatic identification of all Notion property types
- **Option enumeration** - Get valid values for select/multi-select properties
- **Multiple formats** - JSON, YAML, table, or markdown output
- **Filtered extraction** - Get only the properties you need

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

[📖 Full Enhancement Documentation](./ENHANCEMENTS.md) | [📊 Output Formats Guide](./OUTPUT_FORMATS.md) | [👨‍💻 AI Agent Cookbook](./docs/AI-AGENT-COOKBOOK.md)

## Quick Start for AI Agents

**If you're Claude, GPT, or another AI assistant**, here's everything you need:

1. **Install** (choose based on platform):
   ```bash
   # Mac/Linux: GitHub install works
   npm install -g Coastal-Programs/notion-cli

   # Windows: Use local install (GitHub has symlink issues)
   git clone https://github.com/Coastal-Programs/notion-cli
   cd notion-cli
   npm install -g .
   ```

2. **Set your API token** (ask the human for this):
   ```bash
   # Mac/Linux
   export NOTION_TOKEN="ntn_your_token_here"

   # Windows (Command Prompt)
   set NOTION_TOKEN=ntn_your_token_here

   # Windows (PowerShell)
   $env:NOTION_TOKEN="ntn_your_token_here"
   ```

3. **Verify it works**:
   ```bash
   notion-cli user retrieve bot --output json
   ```

4. **Discover database schema** (new!):
   ```bash
   notion-cli db schema <DATA_SOURCE_ID> --output json
   ```

5. **All commands support** `--output json` for machine-readable responses.

**Get your API token**: https://developers.notion.com/docs/create-a-notion-integration

**Learn by example**: See [AI Agent Cookbook](./docs/AI-AGENT-COOKBOOK.md) for 12+ practical recipes

---

## Installation

### Mac/Linux Users

**GitHub Install (Recommended):**
```bash
npm install -g Coastal-Programs/notion-cli
```

**Local Install (Alternative):**
```bash
git clone https://github.com/Coastal-Programs/notion-cli
cd notion-cli
npm install -g .
```

### Windows Users

**⚠️ Windows Note:** GitHub installations create broken symlinks on Windows. Use local install:

```bash
git clone https://github.com/Coastal-Programs/notion-cli
cd notion-cli
npm install -g .
```

### Using npx (No Install)

**Mac/Linux:**
```bash
npx Coastal-Programs/notion-cli page retrieve <PAGE_ID> --json
```

**Windows:** npx has the same symlink issue. Use local install method above.

---

## Setup Your API Token

After installation, you need to configure your Notion Integration Token.

### Get Your Token
1. Go to https://developers.notion.com/
2. Click "Create new integration"
3. Copy your "Internal Integration Token" (starts with `ntn_`)

### Set the Token

**Mac/Linux (bash/zsh):**
```bash
export NOTION_TOKEN="ntn_your_token_here"

# Make it permanent (add to ~/.bashrc or ~/.zshrc)
echo 'export NOTION_TOKEN="ntn_your_token_here"' >> ~/.bashrc
source ~/.bashrc
```

**Windows (Command Prompt):**
```cmd
set NOTION_TOKEN=ntn_your_token_here

REM Make it permanent (system-wide)
setx NOTION_TOKEN "ntn_your_token_here"
```

**Windows (PowerShell):**
```powershell
$env:NOTION_TOKEN="ntn_your_token_here"

# Make it permanent (current user)
[System.Environment]::SetEnvironmentVariable('NOTION_TOKEN', 'ntn_your_token_here', 'User')
```

### Verify Setup

```bash
notion-cli user retrieve bot --output json
```

**Expected output:**
```json
[
  {
    "id": "your-bot-id",
    "name": "Your Integration Name",
    "type": "bot",
    ...
  }
]
```

**If you see an error:**
```json
{
  "success": false,
  "error": {
    "code": "unauthorized",
    "message": "API token is invalid"
  }
}
```
→ Double-check your `NOTION_TOKEN` is set correctly.

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

### Discover Database Schema (NEW!)
Understand database structure before creating/updating pages:
```bash
# Get schema in JSON format (best for AI agents)
notion-cli db schema <DATA_SOURCE_ID> --output json

# Get schema as formatted table
notion-cli db schema <DATA_SOURCE_ID>

# Get only specific properties
notion-cli db schema <DATA_SOURCE_ID> --properties Name,Status,Tags --output json

# Export as markdown documentation
notion-cli db schema <DATA_SOURCE_ID> --markdown
```

**Example output:**
```json
{
  "success": true,
  "data": {
    "id": "abc123...",
    "title": "Tasks",
    "properties": [
      {
        "name": "Name",
        "type": "title",
        "required": true,
        "description": "Title (required)"
      },
      {
        "name": "Status",
        "type": "select",
        "options": ["Not Started", "In Progress", "Done"],
        "description": "Select one: Not Started, In Progress, Done"
      },
      {
        "name": "Tags",
        "type": "multi_select",
        "options": ["urgent", "bug", "feature", "docs"]
      }
    ]
  }
}
```

**Use with jq:**
```bash
# Get all property names
notion-cli db schema <DATA_SOURCE_ID> --output json | jq -r '.data.properties[].name'

# Get properties with options (select/multi-select)
notion-cli db schema <DATA_SOURCE_ID> --output json | \
  jq '.data.properties[] | select(.options) | {name, options}'
```

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
Page title,page,c77dbaf2-4017-4ea1-ac1e93a87269f3ea,https://www.notion.so/Page-title-c77dbaf240174ea1ac1e93a87269f3ea
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
- `notion-cli db schema <DATA_SOURCE_ID>` - **NEW:** Extract AI-friendly schema
- `notion-cli db update <DATA_SOURCE_ID>` - Update data source title/properties

**Aliases:**
- `db:*` commands also available as `ds:*` (data-source)
- `db:r` = `db:retrieve` = `ds:r` = `ds:retrieve`
- `db:s` = `db:schema` = `ds:s` = `ds:schema` (new!)
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

### 1. Schema Discovery (NEW!)
AI agents can understand database structure instantly:
```bash
# Discover what properties exist and their types
notion-cli db schema <DATA_SOURCE_ID> --output json | jq '.data.properties'

# Find valid options for select fields
notion-cli db schema <DATA_SOURCE_ID> --output json | \
  jq '.data.properties[] | select(.type=="select") | {name, options}'
```

See [AI Agent Cookbook](./docs/AI-AGENT-COOKBOOK.md) for complete recipes.

### 2. Knowledge Base Queries
AI agents can query your Notion workspace:
```bash
notion-cli search --query "API documentation" --json
```

### 3. Task Management
Create and update tasks programmatically:
```bash
notion-cli page create -d <TASKS_DATA_SOURCE_ID> --json
notion-cli page update <TASK_PAGE_ID> --archive --json
```

### 4. Content Generation
Generate content with AI, save to Notion:
```bash
echo "# Generated Content" > output.md
notion-cli page create -f output.md -p <PARENT_PAGE_ID> --json
```

### 5. Data Extraction
Extract structured data for AI processing:
```bash
notion-cli db query <DATA_SOURCE_ID> --json | jq '.data.results'
```

### 6. Workflow Automation
Chain commands in scripts:
```bash
#!/bin/bash
RESULT=$(notion-cli db query $DATA_SOURCE_ID --json)
if [ $? -eq 0 ]; then
  echo "$RESULT" | jq '.data.results[] | .properties'
fi
```

**See complete examples:** [AI Agent Cookbook](./docs/AI-AGENT-COOKBOOK.md) has 12+ practical automation recipes.

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
**Extract schema (NEW!)** | ✅ | ✅ | ✅ (10m)
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

- [AI Agent Cookbook](./docs/AI-AGENT-COOKBOOK.md) - **NEW:** Practical recipes for AI automation
- [Enhancement Documentation](./ENHANCEMENTS.md) - Complete reference for retry & caching
- [Output Formats Guide](./OUTPUT_FORMATS.md) - All available output formats
- [Configuration Examples](.env.example) - Environment variable templates
- [Code Examples](./src/examples/cache-retry-examples.ts) - Programmatic usage

## Author

**Jake Schepis** - [Coastal Programs](https://github.com/Coastal-Programs)

**Key Features:**
- ✅ Notion API v5.2.1 with data sources support
- ✅ **NEW: Schema discovery for AI agents**
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
