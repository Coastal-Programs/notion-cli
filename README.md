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
  <a href="https://www.npmjs.com/package/@coastal-programs/notion-cli">
    <img src="https://img.shields.io/npm/v/@coastal-programs/notion-cli.svg" alt="npm version">
  </a>
  <a href="https://go.dev/">
    <img src="https://img.shields.io/badge/go-%3E%3D1.21-00ADD8.svg" alt="Go Version">
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

> Notion CLI for AI Agents & Automation -- a single Go binary, no runtime dependencies.

A powerful command-line interface for the Notion API, optimized for AI coding assistants and automation scripts. Built in Go with Cobra, distributed as a single ~8MB binary.

**Key Features:**
- **Single binary**: ~8MB, zero runtime dependencies, instant startup
- **OAuth login**: `notion-cli auth login` -- authenticate in your browser, no token management
- **AI-first design**: JSON envelope output, structured errors, exit codes
- **Non-interactive**: Perfect for scripts and automation
- **Flexible output**: JSON, CSV, table, or raw API responses
- **Reliable**: Automatic retry with exponential backoff for 408/429/5xx
- **Intelligent caching**: In-memory TTL cache with per-resource-type TTLs
- **Schema discovery**: AI-friendly database schema extraction
- **Workspace caching**: Fast database lookups without API calls
- **Smart ID resolution**: Automatic database_id / data_source_id conversion
- **Cross-platform**: macOS (arm64, amd64), Linux (amd64, arm64), Windows (amd64)
- **Near-zero supply chain risk**: 2 Go dependencies (cobra, pflag) vs 573 npm packages in v5.x

## What's New in v6.0.0

**Complete rewrite from TypeScript/oclif to Go/Cobra.**

All 26 commands have been ported with identical syntax -- existing scripts work unchanged. The JSON envelope format (`{success, data, metadata}`) and environment variables (`NOTION_TOKEN`) are the same.

**Why Go?**
- Single binary distribution (~8MB) instead of 573 npm dependencies
- Instant startup with no Node.js runtime overhead
- Cross-compilation: one build produces 5 platform binaries
- Near-zero supply chain risk: 2 Go dependencies vs hundreds of npm packages

**v6.1.0: OAuth Authentication** -- `notion-cli auth login` opens your browser, you authorize, done. No more copying tokens manually.

**Technical details:**
- 36 Go source files, ~9,800 lines of code
- 196 tests across 9 test suites, all passing
- ~8MB binary (stripped, darwin/arm64)

**Deferred to Phase 2:** Disk cache, request deduplication, circuit breaker, simple properties (`-S` flag), recursive page retrieval, markdown output from page content, interactive init wizard, update notifications.

See [CHANGELOG.md](./CHANGELOG.md) for full details.

## Quick Start

### Installation

**Option 1: npm (recommended for most users)**
```bash
npm install -g @coastal-programs/notion-cli
```

This installs a thin npm wrapper that downloads the correct platform-specific binary automatically. Requires Node.js >= 18 at install time only -- the binary itself has no Node.js dependency.

**Option 2: Go install (from source)**
```bash
go install github.com/Coastal-Programs/notion-cli/cmd/notion-cli@latest
```

Requires Go 1.21+.

**Option 3: Build from source**
```bash
git clone https://github.com/Coastal-Programs/notion-cli.git
cd notion-cli
make build
# Binary is at build/notion-cli
```

### Setup

**Option A: OAuth login (recommended)**
```bash
# Authenticate via your browser -- no token needed
notion-cli auth login

# Verify connectivity
notion-cli whoami
```

**Option B: Manual token (CI/automation)**
```bash
# Set your Notion API token
export NOTION_TOKEN="secret_your_token_here"

# Or save it to the config file
notion-cli config set-token <YOUR_TOKEN>
```

**Get a manual token:** https://developers.notion.com/docs/create-a-notion-integration

```bash
# Sync your workspace for local database lookups
notion-cli sync
```

### Common Commands

```bash
# List your databases
notion-cli list --output json

# Discover database schema
notion-cli db schema <DATABASE_ID> --output json

# Query a database
notion-cli db query <DATABASE_ID> --output json

# Create a page
notion-cli page create --database-id <DATABASE_ID> \
  --properties '{"Name": {"title": [{"text": {"content": "My Task"}}]}}'

# Search the workspace
notion-cli search "project" --output json
```

All commands support `--output json` for machine-readable responses.

## Commands

### Authentication

```bash
# Log in via OAuth (opens browser)
notion-cli auth login

# Check current auth status
notion-cli auth status

# Log out (clear stored OAuth tokens)
notion-cli auth logout
```

### Setup and Diagnostics

```bash
# Test connectivity and show bot info
notion-cli whoami

# Health check and diagnostics
notion-cli doctor

# Configure token manually (for CI/scripts)
notion-cli config set-token <TOKEN>

# Get config value
notion-cli config get <KEY>

# Show config file path
notion-cli config path

# View cache statistics
notion-cli cache info
```

### Database Commands

```bash
# Retrieve database metadata
notion-cli db retrieve <DATABASE_ID>

# Query database with filters
notion-cli db query <DATABASE_ID> --output json
notion-cli db query <DATABASE_ID> --filter '{"property": "Status", "select": {"equals": "Done"}}'

# Create new database
notion-cli db create \
  --parent-page <PAGE_ID> \
  --title "My Database" \
  --properties '{"Name": {"type": "title"}}'

# Update database
notion-cli db update <DATABASE_ID> --title "New Title"

# Extract schema (AI-friendly)
notion-cli db schema <DATABASE_ID> --output json
```

### Page Commands

```bash
# Create page in database
notion-cli page create \
  --database-id <DATABASE_ID> \
  --properties '{"Name": {"title": [{"text": {"content": "Task"}}]}}'

# Retrieve page
notion-cli page retrieve <PAGE_ID> --output json

# Update page properties
notion-cli page update <PAGE_ID> \
  --properties '{"Status": {"select": {"name": "Done"}}}'

# Get page property item
notion-cli page property-item <PAGE_ID> --property-id <PROPERTY_ID>
```

### Block Commands

```bash
# Retrieve block
notion-cli block retrieve <BLOCK_ID>

# Get block children
notion-cli block children <BLOCK_ID> --output json

# Append children to block
notion-cli block append <BLOCK_ID> \
  --children '[{"object": "block", "type": "paragraph", "paragraph": {"rich_text": [{"text": {"content": "Hello"}}]}}]'

# Update block
notion-cli block update <BLOCK_ID> --content "Updated text"

# Delete block
notion-cli block delete <BLOCK_ID>
```

### User Commands

```bash
# List all users
notion-cli user list --output json

# Retrieve user
notion-cli user retrieve <USER_ID>

# Get bot user info
notion-cli user bot
```

### Search

```bash
# Search workspace
notion-cli search "project" --output json

# Search with filter
notion-cli search "docs" --filter page
```

### Workspace

```bash
# Sync workspace (cache all accessible databases)
notion-cli sync

# List cached databases
notion-cli list --output json

# Batch retrieve multiple objects
notion-cli batch retrieve --ids <ID1>,<ID2>,<ID3>
```

## Database Query Filtering

Use `--filter` with JSON matching the Notion API filter format:

```bash
# Filter by select property
notion-cli db query <ID> \
  --filter '{"property": "Status", "select": {"equals": "Done"}}' \
  --output json

# Compound AND filter
notion-cli db query <ID> \
  --filter '{"and": [{"property": "Status", "select": {"equals": "Done"}}, {"property": "Priority", "number": {"greater_than": 5}}]}' \
  --output json

# OR filter
notion-cli db query <ID> \
  --filter '{"or": [{"property": "Tags", "multi_select": {"contains": "urgent"}}, {"property": "Tags", "multi_select": {"contains": "bug"}}]}' \
  --output json

# Date filter
notion-cli db query <ID> \
  --filter '{"property": "Due Date", "date": {"on_or_before": "2026-12-31"}}' \
  --output json
```

## Output Formats

All commands support multiple output formats via the `--output` flag:

```bash
# JSON (structured envelope format)
notion-cli db query <ID> --output json

# Table (default, human-readable)
notion-cli db query <ID> --output table

# CSV
notion-cli db query <ID> --output csv

# Raw API response (no envelope wrapping)
notion-cli db query <ID> --raw
```

### JSON Envelope Format

All JSON output uses a consistent envelope:

```json
{
  "success": true,
  "data": { ... },
  "metadata": {
    "object": "database",
    "request_id": "abc-123"
  }
}
```

Error responses follow the same structure:

```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Database not found",
    "suggestion": "Verify the database ID and check that your integration has access."
  }
}
```

### Exit Codes

- `0` -- Success
- `1` -- Notion API error
- `2` -- CLI error (invalid flags, missing arguments, etc.)

## Key Features for AI Agents

### JSON Mode

Every command supports `--output json` for structured, parseable output:

```bash
# Get structured data
notion-cli db query <ID> --output json | jq '.data.results[].properties'

# Error responses are also structured JSON
notion-cli db retrieve invalid-id --output json
```

### Schema Discovery

Extract complete database schemas in AI-friendly formats:

```bash
# Get full schema
notion-cli db schema <DATABASE_ID> --output json

# Example output:
# {
#   "success": true,
#   "data": {
#     "database_id": "...",
#     "title": "Tasks",
#     "properties": {
#       "Name": { "type": "title" },
#       "Status": {
#         "type": "select",
#         "options": ["Not Started", "In Progress", "Done"]
#       }
#     }
#   }
# }
```

### Smart ID Resolution

No need to worry about `database_id` vs `data_source_id` confusion. The CLI automatically detects and converts between them:

```bash
# Both work -- use whichever ID you have
notion-cli db retrieve 1fb79d4c71bb8032b722c82305b63a00  # database_id
notion-cli db retrieve 2gc80e5d82cc9043c833d93416c74b11  # data_source_id
```

### Workspace Caching

Cache your entire workspace locally for instant database lookups:

```bash
# One-time sync
notion-cli sync

# Now use database names instead of IDs
notion-cli db query "Tasks Database" --output json

# Browse all cached databases
notion-cli list --output json
```

### Cache Metadata

AI agents can check data freshness before operations:

```bash
notion-cli cache info --output json
```

**Cache TTLs by resource type:**
- Databases: 10 minutes
- Pages: 1 minute
- Users: 1 hour
- Blocks: 30 seconds

## Authentication

The CLI supports three authentication methods, checked in this order:

| Priority | Method | Use case |
|----------|--------|----------|
| 1 | `NOTION_TOKEN` env var | CI/CD, scripts, automation |
| 2 | OAuth token (from `auth login`) | Interactive / daily use |
| 3 | Manual token (from `config set-token`) | Fallback |

```bash
# Recommended: OAuth login
notion-cli auth login

# For CI/automation: environment variable
export NOTION_TOKEN=secret_your_token_here

# Check which method is active
notion-cli auth status
```

### Configuration

The CLI reads configuration from `~/.config/notion-cli/config.json`. You can manage it with:

```bash
# Set a value
notion-cli config set-token <TOKEN>

# Get a value
notion-cli config get <KEY>

# Show config file path
notion-cli config path
```

## Real-World Examples

### Automated Task Management

```bash
#!/bin/bash
# Create a task and mark it complete

TASK_ID=$(notion-cli page create \
  --database-id "$TASKS_DB_ID" \
  --properties '{
    "Name": {"title": [{"text": {"content": "Review PR"}}]},
    "Status": {"select": {"name": "In Progress"}}
  }' \
  --output json | jq -r '.data.id')

echo "Created task: $TASK_ID"

# Do work...

# Mark complete
notion-cli page update "$TASK_ID" \
  --properties '{"Status": {"select": {"name": "Done"}}}' \
  --output json
```

### Database Schema Migration

```bash
#!/bin/bash
# Export schema from one database, create another

notion-cli db schema "$SOURCE_DB" --output json > schema.json

notion-cli db create \
  --parent-page "$TARGET_PAGE" \
  --title "Migrated Database" \
  --properties "$(jq '.data.properties' schema.json)" \
  --output json
```

### Daily Sync Script

```bash
#!/bin/bash
# Sync workspace and generate report

notion-cli sync

notion-cli list --output json > databases.json

echo "# Database Report - $(date)" > report.md
jq -r '.data[] | "- **\(.title)** (\(.id))"' databases.json >> report.md
```

## Troubleshooting

### "Database not found" Error

The CLI auto-resolves `database_id` vs `data_source_id`. If it still fails, verify your integration has access to the database in Notion's integration settings.

### Rate Limiting (429 errors)

The CLI handles this automatically with exponential backoff and jitter. Retry behavior covers HTTP 408, 429, and 5xx responses.

### Authentication Errors

```bash
# Check auth status
notion-cli auth status

# Re-authenticate via OAuth
notion-cli auth login

# Or verify your token env var is set
echo $NOTION_TOKEN

# Test connectivity
notion-cli whoami

# Ensure integration has access to the pages/databases you need:
# https://www.notion.so/my-integrations
```

### Slow Queries

1. Use filters to reduce data: `--filter '{"property": "Status", "select": {"equals": "Active"}}'`
2. Sync your workspace first: `notion-cli sync`
3. Use `--output json` for faster parsing than table output

## Architecture

notion-cli is built in Go with a focus on simplicity, reliability, and minimal dependencies.

- **CLI framework**: [Cobra](https://github.com/spf13/cobra) for command parsing and flag handling
- **HTTP client**: Raw `net/http` with gzip support -- no Notion SDK dependency
- **Caching**: In-memory TTL cache with per-resource-type expiration
- **Retry**: Exponential backoff with jitter for transient failures (408/429/5xx)
- **Errors**: 40+ structured error codes with human-readable suggestions
- **Output**: JSON envelope, ASCII table, CSV via `pkg/output.Printer`
- **Config**: Environment variables + JSON config file (`~/.config/notion-cli/config.json`)
- **Distribution**: npm wrapper with platform-specific binary packages (esbuild-style pattern)

### Dependencies

| Dependency | Purpose |
|---|---|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/spf13/pflag` | Flag parsing (indirect, via cobra) |
| Go standard library | Everything else |

## Development

### Prerequisites

- Go 1.21+ (the go.mod specifies 1.25, but the project targets 1.21+ compatibility)
- Git
- Make
- (Optional) [golangci-lint](https://golangci-lint.run/) for extended linting

### Setup

```bash
git clone https://github.com/Coastal-Programs/notion-cli.git
cd notion-cli
make build
```

### Development Workflow

```bash
# Build the binary to build/notion-cli
make build

# Run tests
make test

# Lint (go vet + golangci-lint if installed)
make lint

# Format code
make fmt

# Tidy module dependencies
make tidy

# Install to $GOPATH/bin
make install

# Cross-compile for all platforms
make release

# Clean build artifacts
make clean
```

### Project Structure

```
notion-cli/
├── cmd/notion-cli/
│   └── main.go                  # Entry point
├── internal/
│   ├── cli/
│   │   ├── root.go              # Cobra root command + global flags
│   │   └── commands/
│   │       ├── db.go            # db query, retrieve, create, update, schema
│   │       ├── page.go          # page create, retrieve, update, property_item
│   │       ├── block.go         # block append, retrieve, delete, update, children
│   │       ├── user.go          # user list, retrieve, bot
│   │       ├── search.go        # search command
│   │       ├── sync.go          # workspace sync
│   │       ├── list.go          # list cached databases
│   │       ├── batch.go         # batch retrieve
│   │       ├── whoami.go        # connectivity check
│   │       ├── doctor.go        # health checks
│   │       ├── auth.go           # auth login, logout, status (OAuth)
│   │       ├── config.go        # config get/set/path
│   │       └── cache_cmd.go     # cache info/stats
│   ├── oauth/
│   │   └── oauth.go             # OAuth flow (localhost server, token exchange)
│   ├── notion/
│   │   └── client.go            # HTTP client, auth, request/response
│   ├── cache/
│   │   ├── cache.go             # In-memory TTL cache
│   │   └── workspace.go         # Workspace database cache
│   ├── retry/
│   │   └── retry.go             # Exponential backoff with jitter
│   ├── errors/
│   │   └── errors.go            # NotionCLIError with codes, suggestions
│   ├── config/
│   │   └── config.go            # Config loading (env vars + JSON file)
│   └── resolver/
│       └── resolver.go          # URL/ID/name resolution
├── pkg/
│   └── output/
│       ├── output.go            # JSON/text/table/CSV formatting
│       ├── envelope.go          # Envelope wrapper
│       └── table.go             # Table formatter
├── go.mod
├── go.sum
├── Makefile
├── package.json                 # npm distribution wrapper
└── docs/                        # Documentation
```

### Code Patterns

- All commands use Cobra; registered via `Register*Commands(root *cobra.Command)`
- Use `pkg/output.Printer` for all output -- never `fmt.Println` directly
- Use `internal/errors.NotionCLIError` for errors -- never raw errors
- Use envelope format for JSON output: `{success, data, metadata}`
- Use `internal/resolver.ExtractID()` for all ID/URL inputs
- Use `context.Context` for all API calls

### Testing

```bash
# Run all tests with verbose output
make test

# Run a specific test package
go test ./internal/cache/... -v

# Run a specific test
go test ./internal/cache/... -v -run TestCacheExpiry
```

196 tests across 9 test suites.

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on:
- Code style and conventions
- Test requirements
- Pull request process
- Commit message format (conventional commits: `feat:`, `fix:`, `test:`, etc.)

## Legal and Compliance

### Trademark Notice
"Notion" is a registered trademark of Notion Labs, Inc. This project is an independent,
unofficial tool and is not affiliated with, endorsed by, or sponsored by Notion Labs, Inc.

### License
This project is licensed under the MIT License -- see the [LICENSE](LICENSE) file for details.

### Third-Party Licenses
This project uses open-source dependencies. See [NOTICE](NOTICE) for complete
third-party license information.

## Support

- **Issues**: Report bugs on [GitHub Issues](https://github.com/Coastal-Programs/notion-cli/issues)
- **Discussions**: Ask questions in [GitHub Discussions](https://github.com/Coastal-Programs/notion-cli/discussions)
- **Documentation**: Full guides in the `/docs` folder

## Related Projects

- **Notion API**: https://developers.notion.com
- **Cobra**: https://github.com/spf13/cobra

---

**Built for AI agents, optimized for automation. A single Go binary -- no runtime dependencies.**
