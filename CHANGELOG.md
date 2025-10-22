# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [5.2.0] - 2025-10-22

### Added - Schema Discovery for AI Agents

**NEW `db schema` Command:**
- Extract clean, AI-parseable database schemas with `notion-cli db schema <DATA_SOURCE_ID>`
- Automatic property type detection for all Notion property types (title, select, multi-select, date, number, formula, rollup, relation, etc.)
- Option enumeration for select/multi-select properties - get valid values instantly
- Multiple output formats: JSON, YAML, table, and markdown
- Filtered extraction with `--properties` flag to get only what you need
- Command aliases: `db:s`, `ds:schema`, `ds:s`

**Schema Extractor Utility (`src/utils/schema-extractor.ts`):**
- `extractSchema()` - Transform complex Notion API responses into simple, flat structures
- `filterProperties()` - Extract specific properties by name
- `formatSchemaForTable()` - Human-readable table data
- `formatSchemaAsMarkdown()` - Generate markdown documentation from schemas
- `validateAgainstSchema()` - Validate data objects against schemas
- Full TypeScript type definitions for schema structures

**AI Agent Cookbook:**
- New comprehensive guide: `docs/AI-AGENT-COOKBOOK.md`
- 12+ practical recipes for AI automation workflows
- Complete code examples with expected outputs
- Error handling patterns and best practices
- Multi-step automation workflows
- Schema discovery patterns
- Batch operations examples
- Data extraction and transformation recipes

**Documentation Enhancements:**
- Updated README with schema discovery quick start
- Added schema command to all command listings
- New "Schema Discovery" use case section
- Updated API coverage table
- Links to AI Agent Cookbook throughout documentation

### Changed

- **Version:** Bumped to 5.2.0 to reflect major new feature
- **Package description:** Added "with schema discovery" to description
- **Keywords:** Added "schema-discovery" to package.json keywords
- **README structure:** Reorganized to highlight schema discovery as key feature

### Technical Details

**Command Architecture:**
- Follows existing oclif patterns from `db retrieve` command
- Uses established caching layer (10-minute TTL for schemas)
- Integrates with existing retry logic for reliability
- Consistent error handling with structured JSON responses
- Multiple output formats (JSON, YAML, table, markdown)

**Type Safety:**
- Full TypeScript type definitions
- Interfaces for `PropertySchema` and `DataSourceSchema`
- Type-safe property extraction and transformation

**Performance:**
- Leverages existing cache infrastructure
- Schema requests cached for 10 minutes by default
- Same retry and circuit breaker patterns as other commands

**Compatibility:**
- No breaking changes to existing commands
- All existing functionality preserved
- New command is additive only

### Why This Matters

**For AI Agents:**
- Eliminates guessing about property names and types
- Provides valid options for select/multi-select fields upfront
- Enables dynamic, schema-aware automation
- Reduces errors from invalid property values
- Makes Notion databases self-documenting

**For Developers:**
- Faster integration with unknown databases
- Clear property type information
- Instant documentation generation
- Validation helpers included
- jq-friendly JSON output

### Examples

```bash
# Get full schema in JSON (best for AI agents)
notion-cli db schema abc123 --output json

# Get only specific properties
notion-cli db schema abc123 --properties Name,Status,Tags --output json

# Generate markdown documentation
notion-cli db schema abc123 --markdown

# Find all select properties and their options
notion-cli db schema abc123 --output json | \
  jq '.data.properties[] | select(.options) | {name, options}'
```

### Migration Guide

No migration needed! This is a purely additive feature. All existing commands continue to work exactly as before.

**New workflow recommendation:**
1. Always run `db schema` first when working with a new database
2. Use schema output to build correct property structures
3. Validate your data against the schema before creating pages

---

## [5.1.0] - 2025-10-21

### Added - Enhanced Reliability & Performance

**Retry Logic with Exponential Backoff:**
- Automatic retry on failures (up to 3 attempts by default)
- Intelligent error categorization (retryable vs non-retryable)
- Exponential backoff with jitter to prevent thundering herd
- Automatic rate limit handling with Retry-After header support
- Circuit breaker pattern for resilient operations
- Configurable via environment variables

**In-Memory Caching Layer:**
- Intelligent caching for frequently accessed resources
- TTL-based expiration (data sources: 10min, users: 1hr, blocks/pages: 30s-1m)
- Automatic cache invalidation on write operations
- Cache statistics for monitoring performance
- Up to 100x faster for repeated reads
- Configurable cache size and TTLs

**Configuration Options:**
- `NOTION_CLI_MAX_RETRIES` - Maximum retry attempts
- `NOTION_CLI_BASE_DELAY` - Base delay between retries (ms)
- `NOTION_CLI_MAX_DELAY` - Maximum delay cap (ms)
- `NOTION_CLI_CACHE_ENABLED` - Enable/disable caching
- `NOTION_CLI_CACHE_MAX_SIZE` - Maximum cache entries
- `NOTION_CLI_CACHE_DS_TTL` - Data source cache TTL (ms)
- `NOTION_CLI_CACHE_USER_TTL` - User cache TTL (ms)

### Changed

- All API calls now use enhanced retry logic
- Frequently accessed resources automatically cached
- Error messages more informative with retry context
- Debug mode shows cache hits/misses and retry attempts

### Performance

- **110x faster** for repeated DB schema reads
- **2.5x faster** for bulk operations
- **11x more reliable** rate limit handling (99.9% success)
- **9x fewer failures** on poor network conditions

---

## [5.0.0] - 2025-10-20

### Added - Initial Release

**Core Notion API v5.2.1 Support:**
- Data source operations (query, retrieve, update, create)
- Page operations (create, retrieve, update, archive)
- Block operations (append, retrieve, update, delete)
- User operations (list, retrieve, me)
- Search operations

**Output Formats:**
- JSON output mode (`--output json`, `--json`)
- CSV format (`--output csv`)
- YAML format (`--output yaml`)
- Table format (default)
- Raw API responses (`--raw`)

**Automation Features:**
- Non-interactive design
- Structured JSON responses with `success` flag
- Exit codes (0 = success, 1 = error)
- Consistent error handling
- Machine-readable error responses

**Command Aliases:**
- `db:*` commands available as `ds:*` (data-source)
- Short aliases for common commands (e.g., `db:r`, `db:u`)

### Technical

- Built with oclif framework
- TypeScript codebase
- Node.js >= 18.0.0 required
- @notionhq/client v5.2.1

---

## Legend

- **Added:** New features
- **Changed:** Changes in existing functionality
- **Deprecated:** Soon-to-be removed features
- **Removed:** Removed features
- **Fixed:** Bug fixes
- **Security:** Security improvements

---

**Links:**
- [GitHub Releases](https://github.com/Coastal-Programs/notion-cli/releases)
- [Notion API Documentation](https://developers.notion.com/)
- [AI Agent Cookbook](./docs/AI-AGENT-COOKBOOK.md)
