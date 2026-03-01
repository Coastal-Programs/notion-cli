# Documentation Index

This directory contains comprehensive documentation for the Notion CLI.

> **Note:** As of v6.0.0, notion-cli has been completely rewritten from TypeScript to Go.
> It is distributed as a single ~8MB binary via npm (platform-specific packages) or built from source with `make build`.
> All 26 commands from v5.x are fully ported with identical syntax and output formats.

## Command Reference

Individual command documentation:

- **[db](db.md)** - Database commands (query, retrieve, create, update, schema)
- **[page](page.md)** - Page commands (create, retrieve, update, property_item)
- **[block](block.md)** - Block commands (append, retrieve, children, update, delete)
- **[user](user.md)** - User commands (list, retrieve, bot)
- **[search](search.md)** - Search command
- **[batch](batch.md)** - Batch retrieve command
- **[sync](sync.md)** - Workspace sync command
- **[list](list.md)** - List cached databases
- **[whoami](whoami.md)** - Connectivity check
- **[doctor](doctor.md)** - Health checks and diagnostics
- **[config](config.md)** - Configuration management (set-token, get, path, list)
- **[cache](cache.md)** - Cache info and statistics
- **[help](help.md)** - Built-in help system

## User Guides

Essential guides for using the CLI effectively:

- **[AI Agent Guide](user-guides/ai-agent-guide.md)** - Quick reference for AI assistants using this CLI
- **[AI Agent Cookbook](user-guides/ai-agent-cookbook.md)** - Common patterns and recipes for AI agents
- **[Output Formats](user-guides/output-formats.md)** - JSON, CSV, YAML, and table output options
- **[Filter Guide](user-guides/filter-guide.md)** - Database query filtering syntax
- **[Verbose Logging](user-guides/verbose-logging.md)** - Debug mode and troubleshooting
- **[Envelope System](user-guides/envelope-index.md)** - Standardized response format
- **[Error Handling](user-guides/error-handling-examples.md)** - Understanding and handling errors

## Architecture

Deep dives into internal systems:

- **[Caching](architecture/caching.md)** - In-memory TTL caching strategy
- **[Envelopes](architecture/envelopes.md)** - Response envelope architecture
- **[Error Handling](architecture/error-handling.md)** - Enhanced error system architecture
- **[Smart ID Resolution](architecture/smart-id-resolution.md)** - Automatic database_id / data_source_id conversion

## API Reference

Notion API documentation:

- **[Notion API v5.2.1](api-reference/notion-api-reference-v5.2.1.md)** - Complete API reference
- **[Quick Reference](api-reference/notion-api-quick-reference.md)** - Common API patterns
- **[Database Specification](api-reference/notion-api-database-specification.md)** - Database API details
- **[Search & Users](api-reference/notion-api-search-users-spec.md)** - Search and user endpoints

## Development

Building and contributing:

```bash
make build     # Build Go binary to build/notion-cli
make test      # Run Go test suite
make lint      # go vet + golangci-lint
make release   # Cross-compile for all platforms
make fmt       # Format Go code
make tidy      # go mod tidy
```

- **[Claude Guide](development/claude.md)** - Instructions for Claude Code when contributing

## Phase 2 Features (Planned)

The following v5.x features are planned for a future release:

- Interactive setup wizard (`init` command)
- Simple properties (`-S` flag) for page create/update
- Recursive page retrieval (`-R` flag)
- Markdown output from page content
- Disk cache and request deduplication
- Circuit breaker
- Update notifications

---

For the main README with installation and quick start, see [../README.md](../README.md).
