# notion-cli

Unofficial CLI for the Notion API (`@coastal-programs/notion-cli`), optimized for AI agents and automation. Built in Go with Cobra, distributed as a single binary via npm.

## Quick Reference

```bash
make build             # Build Go binary to build/notion-cli
make test              # Run Go test suite
make lint              # go vet + golangci-lint
make release           # Cross-compile for all platforms
make fmt               # Format Go code
make tidy              # go mod tidy
```

## Project Structure

```
cmd/notion-cli/main.go          # Entry point
internal/
  cli/
    root.go                      # Cobra root command + global flags
    commands/
      db.go                      # db query, retrieve, create, update, schema
      page.go                    # page create, retrieve, update, property_item
      block.go                   # block append, retrieve, delete, update, children
      user.go                    # user list, retrieve, bot
      search.go                  # search command
      sync.go                    # workspace sync
      list.go                    # list cached databases
      batch.go                   # batch retrieve
      whoami.go                  # connectivity check
      doctor.go                  # health checks
      config.go                  # config get/set/path/list
      cache_cmd.go               # cache info/stats
  notion/
    client.go                    # HTTP client, auth, request/response
  cache/
    cache.go                     # In-memory TTL cache
    workspace.go                 # Workspace database cache
  retry/
    retry.go                     # Exponential backoff with jitter
  errors/
    errors.go                    # NotionCLIError with codes, suggestions
  config/
    config.go                    # Config loading (env vars + JSON file)
  resolver/
    resolver.go                  # URL/ID/name resolution
pkg/
  output/
    output.go                    # JSON/text/table/CSV formatting
    envelope.go                  # Envelope wrapper
    table.go                     # Table formatter
go.mod
go.sum
Makefile
```

## Code Patterns (Always Follow)

- All commands use Cobra; register via `Register*Commands(root *cobra.Command)`
- Use `pkg/output.Printer` for all output, never `fmt.Println` directly
- Use `internal/errors.NotionCLIError` for errors, never raw errors
- Use envelope format for JSON output: `{success, data, metadata}`
- Use `internal/resolver.ExtractID()` for all ID/URL inputs
- Use `context.Context` for all API calls

## Git Workflow

- Conventional commits: `feat:`, `fix:`, `test:`, `docs:`, `chore:`, `refactor:`
- Feature branches: `git checkout -b feature/description`
- PRs required for all changes to main
- Never force push to main

## Before Completing Any Task

1. All tests pass: `make test`
2. Build succeeds: `make build`
3. Lint passes: `make lint`
4. CHANGELOG.md updated with changes

## Key Dependencies

- `github.com/spf13/cobra` (CLI framework)
- No Notion SDK - raw HTTP client
- Standard library only for everything else

## Architecture Notes

- **Caching**: In-memory TTL cache. TTLs by resource type (blocks: 30s, pages: 1min, users: 1hr, databases: 10min)
- **Retry**: Exponential backoff with jitter for 408/429/5xx
- **HTTP**: Raw net/http client with gzip support
- **Distribution**: npm wrapper package with platform-specific binary packages (esbuild pattern)

## Environment

```bash
NOTION_TOKEN=secret_...  # Required for all API calls
```

## Additional Docs

- @PUBLISHING.md - npm publishing guide
- @CONTRIBUTING.md - contribution guidelines
- @CHANGELOG.md - release history
- @docs/ - command docs and user guides
