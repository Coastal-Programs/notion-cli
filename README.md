<div align="center">
<pre>
███╗   ██╗ ██████╗ ████████╗██╗ ██████╗ ███╗   ██╗     ██████╗██╗     ██╗
████╗  ██║██╔═══██╗╚══██╔══╝██║██╔═══██╗████╗  ██║    ██╔════╝██║     ██║
██╔██╗ ██║██║   ██║   ██║   ██║██║   ██║██╔██╗ ██║    ██║     ██║     ██║
██║╚██╗██║██║   ██║   ██║   ██║██║   ██║██║╚██╗██║    ██║     ██║     ██║
██║ ╚████║╚██████╔╝   ██║   ██║╚██████╔╝██║ ╚████║    ╚██████╗███████╗██║
╚═╝  ╚═══╝ ╚═════╝    ╚═╝   ╚═╝ ╚═════╝ ╚═╝  ╚═══╝     ╚═════╝╚══════╝╚═╝
</pre>
</div>

> Notion CLI for AI Agents & Automation (API v5.2.1)

A non-interactive command-line interface for Notion's API, optimized for AI coding assistants and automation scripts. **Skip the MCP server** - get direct, fast access to Notion via CLI.

**Why Use This Instead of MCP Servers:**
- 🚀 **Faster**: Direct API calls without MCP overhead
- 🤖 **AI-First**: JSON output mode, structured errors, exit codes
- ⚡ **Non-Interactive**: No prompts - perfect for automation
- 📊 **Flexible Output**: JSON, CSV, YAML, or raw API responses
- ✅ **Latest API**: Notion API v5.2.1 with data sources support
- 🔄 **Enhanced Reliability**: Automatic retry with exponential backoff
- ⚡ **High Performance**: In-memory + persistent caching
- 🔍 **Schema Discovery**: AI-friendly database schema extraction
- 🗄️ **Workspace Caching**: Fast database lookups without API calls
- 🧠 **Smart ID Resolution**: Automatic database_id → data_source_id conversion

## What's New in v5.4.0

### NEW: Smart ID Resolution
- **Automatic conversion** - Use `database_id` or `data_source_id` interchangeably
- **No more confusion** - System detects and converts wrong ID type automatically
- **Helpful messaging** - Shows when conversion happens and explains the difference
- **Zero friction** - Works with all database commands (retrieve, query, update)

[📖 Read the Smart ID Resolution Guide](./docs/smart-id-resolution.md)

### Workspace Database Caching
- **`sync` command** - Cache all workspace databases locally for instant lookups
- **`list` command** - Browse cached databases with rich metadata
- **`config set-token` command** - Easy token setup with guided workflow
- **Persistent cache** at `~/.notion-cli/databases.json`
- **Alias generation** - Find databases by name, nickname, or acronym
- **Zero API calls** for name resolution after sync

### Schema Discovery Command
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
- **Persistent workspace cache** for database metadata
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

2. **Set your API token** (new easy way!):
   ```bash
   # Interactive setup (recommended for first-time users)
   notion-cli config set-token

   # Or set manually (ask the human for the token):
   # Mac/Linux
   export NOTION_TOKEN="secret_your_token_here"

   # Windows (Command Prompt)
   set NOTION_TOKEN=secret_your_token_here

   # Windows (PowerShell)
   $env:NOTION_TOKEN="secret_your_token_here"
   ```

3. **Sync your workspace** (one-time setup):
   ```bash
   notion-cli sync
   ```

4. **Verify it works**:
   ```bash
   notion-cli list --json
   ```

5. **Discover database schema**:
   ```bash
   notion-cli db schema <DATA_SOURCE_ID> --json
   ```

6. **All commands support** `--json` for machine-readable responses.

**Get your API token**: https://developers.notion.com/docs/create-a-notion-integration

## Key Features for AI Agents

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

### Database Commands

```bash
# Retrieve database metadata (works with any ID type!)
notion-cli db retrieve <DATABASE_ID>
notion-cli db retrieve <DATA_SOURCE_ID>
notion-cli db retrieve "Tasks"

# Query database with filters
notion-cli db query <ID> \
  --filter status equals "Done" \
  --json

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

# Configure token
notion-cli config set-token
```

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
1. Use filters to reduce data: `--filter status equals "Active"`
2. Enable caching: `notion-cli sync`
3. Use `--compact-json` for faster output

### Authentication Errors

**Problem**: 401 Unauthorized

**Solution**:
```bash
# Verify token is set
echo $NOTION_TOKEN

# Re-configure token
notion-cli config set-token

# Check integration has access
# Visit: https://www.notion.so/my-integrations
```

## Development

```bash
# Clone repository
git clone https://github.com/Coastal-Programs/notion-cli
cd notion-cli

# Install dependencies
npm install

# Build TypeScript
npm run build

# Run tests
npm test

# Link for local development
npm link
```

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
