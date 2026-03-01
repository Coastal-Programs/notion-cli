# Contributing to notion-cli

Thank you for your interest in contributing to notion-cli! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Commit Message Format](#commit-message-format)
- [Project Structure](#project-structure)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

This project follows a simple code of conduct: be respectful, constructive, and collaborative. We welcome contributions from everyone.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/notion-cli.git
   cd notion-cli
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/Coastal-Programs/notion-cli.git
   ```
4. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Setup

### Prerequisites

- Go 1.21 or later
- Make
- Git
- golangci-lint (optional, for extended linting)

### Installation

```bash
# Download Go module dependencies
go mod download

# Build the binary
make build

# Run the test suite
make test
```

The built binary is placed at `build/notion-cli`. You can also install it directly into your `$GOPATH/bin`:

```bash
make install
```

### Environment Setup

Set the Notion API token as an environment variable:

```bash
export NOTION_TOKEN="secret_your_token_here"
```

Get your token from: https://www.notion.so/my-integrations

## Code Style Guidelines

### Go Conventions

- Follow standard Go conventions as described in [Effective Go](https://go.dev/doc/effective_go)
- All code must be formatted with `gofmt` (run `make fmt`)
- All code must pass `go vet` and `golangci-lint` (run `make lint`)
- Keep functions focused and short
- Return errors rather than panicking
- Use `context.Context` for all API calls

### Code Patterns

- All commands use Cobra; register via `Register*Commands(root *cobra.Command)`
- Use `pkg/output.Printer` for all output, never `fmt.Println` directly
- Use `internal/errors.NotionCLIError` for errors, never raw errors
- Use envelope format for JSON output: `{success, data, metadata}`
- Use `internal/resolver.ExtractID()` for all ID/URL inputs

**Example:**

```go
// RegisterPageCommands adds all page subcommands to the root command.
func RegisterPageCommands(root *cobra.Command) {
	pageCmd := &cobra.Command{
		Use:   "page",
		Short: "Page operations",
	}

	pageCmd.AddCommand(newPageCreateCmd())
	pageCmd.AddCommand(newPageRetrieveCmd())
	root.AddCommand(pageCmd)
}
```

### Naming Conventions

- **Files:** snake_case (`cache_cmd.go`, `workspace.go`)
- **Exported functions/types:** PascalCase (`NewCache`, `NotionCLIError`)
- **Unexported functions/types:** camelCase (`doRequest`, `parseResponse`)
- **Constants:** PascalCase for exported, camelCase for unexported (`DefaultTimeout`, `maxRetries`)
- **Acronyms:** ALL_CAPS within identifiers (`ExtractID`, `ParseJSON`, `HTTPClient`)
- **Packages:** lowercase, single word when possible (`cache`, `retry`, `errors`)

### Documentation

All exported functions, types, and packages must have GoDoc comments. Comments should start with the name of the thing being documented:

```go
// NotionCLIError represents a structured error with an error code,
// user-facing message, and optional suggestions for resolution.
type NotionCLIError struct {
	Code       string
	Message    string
	Suggestions []string
}

// NewCache creates a new in-memory TTL cache with the given maximum
// number of entries. If maxSize is zero or negative, a default of
// 1000 is used.
func NewCache(maxSize int) *Cache {
	// Implementation
}

// ExtractID parses a Notion URL or raw ID string and returns
// the normalized UUID. It returns an error if the input cannot
// be resolved to a valid Notion ID.
func ExtractID(input string) (string, error) {
	// Implementation
}
```

## Testing Requirements

### Running Tests

```bash
# Run all tests
make test

# Run tests for a specific package
go test ./internal/cache/... -v

# Run a specific test function
go test ./internal/cache/... -run TestSetAndGet -v

# Run tests with race detection
go test -race ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Coverage

- All new features must include tests
- Aim for 80%+ code coverage
- Test both success and error cases
- Use `net/http/httptest` for mocking HTTP API calls

### Test Structure

Tests use Go's built-in `testing` package. Test files live alongside the code they test with a `_test.go` suffix:

```go
package cache

import (
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	if c.Size() != 0 {
		t.Errorf("expected empty cache, got size %d", c.Size())
	}
}

func TestSetAndGet(t *testing.T) {
	c := NewCache(100)
	defer c.Stop()

	c.Set("key1", "value1", 1*time.Minute)

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if val != "value1" {
		t.Errorf("expected value1, got %v", val)
	}
}
```

### Test Guidelines

1. **Mock external dependencies** - Use `net/http/httptest` for HTTP calls, never make real API calls
2. **Use descriptive test names** - `TestSetAndGet`, `TestNewCacheInvalidSize`, `TestRetryOnRateLimit`
3. **Use table-driven tests** where appropriate for testing multiple inputs
4. **Test edge cases** - Empty inputs, nil values, zero values, boundary conditions
5. **Keep tests isolated** - No shared mutable state between tests
6. **Use `t.Helper()`** in test helper functions for better error reporting
7. **Use `t.Parallel()`** where safe to speed up the test suite

## Pull Request Process

### Before Submitting

1. **Update from upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks**:
   ```bash
   make build
   make test
   make lint
   ```

3. **Update documentation** if needed:
   - Update README.md for new features
   - Add CHANGELOG.md entry
   - Update GoDoc comments

### Submitting

1. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create Pull Request** on GitHub with:
   - Clear title describing the change
   - Detailed description of what changed and why
   - Reference any related issues (`Fixes #123`)

3. **Fill out PR template** completely

### PR Review Process

- Maintainers will review within 1-2 weeks
- Address review feedback promptly
- Keep PRs focused on a single feature/fix
- Be open to suggestions and changes

### PR Checklist

- [ ] Code follows Go style guidelines
- [ ] All tests pass (`make test`)
- [ ] New tests added for new features
- [ ] Lint passes (`make lint`)
- [ ] Code is formatted (`make fmt`)
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Commit messages follow conventional format
- [ ] No merge conflicts
- [ ] Build succeeds (`make build`)

## Commit Message Format

We follow **Conventional Commits** specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples

```
feat(db): add schema discovery command

Implement new 'db schema' command to extract database schemas
in AI-parseable format. Supports JSON and table output.

Closes #42
```

```
fix(cache): resolve race condition in cache invalidation

Fixed issue where concurrent writes could corrupt cache state.
Added mutex lock for write operations.

Fixes #56
```

```
refactor(notion): simplify HTTP client retry logic

Consolidate retry and backoff into a single configurable function
to reduce code duplication across request methods.
```

### Commit Guidelines

- Use present tense ("add feature" not "added feature")
- Use imperative mood ("move cursor to..." not "moves cursor to...")
- Keep subject line under 72 characters
- Reference issues and PRs in footer
- Explain "what" and "why", not "how"

## Project Structure

```
notion-cli/
├── cmd/
│   └── notion-cli/
│       └── main.go              # Entry point
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
│   │       ├── config.go        # config get/set/path/list
│   │       └── cache_cmd.go     # cache info/stats
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
├── docs/                        # Documentation
├── go.mod                       # Go module definition
├── go.sum                       # Dependency checksums
└── Makefile                     # Build, test, lint, release targets
```

### Key Directories

- **cmd/** - Application entry points. Each subdirectory is a separate binary.
- **internal/** - Private application code. Cannot be imported by other modules.
- **pkg/** - Public library code. Can be imported by external projects.
- **docs/** - User-facing documentation and guides.

## Reporting Issues

### Bug Reports

Include:
- Clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Environment (Go version, OS, architecture)
- Error messages and stack traces
- Minimal reproduction example

### Feature Requests

Include:
- Clear description of the feature
- Use case and motivation
- Example usage
- Potential implementation approach

### Security Vulnerabilities

**Do not open public issues for security vulnerabilities.**

See [SECURITY.md](SECURITY.md) for reporting instructions.

## Development Tips

### Debugging

Use the `--verbose` flag to enable debug output:

```bash
./build/notion-cli db query <id> --verbose
```

### Testing Local Changes

```bash
# Build and test the binary
make build
./build/notion-cli --version
./build/notion-cli whoami

# Or install to $GOPATH/bin
make install
notion-cli db query <id>
```

### Working with Cobra

This project uses the **Cobra** CLI framework:

- Commands are created with `&cobra.Command{}`
- Flags are registered with `cmd.Flags()` (local) or `cmd.PersistentFlags()` (inherited)
- Use `RunE` (not `Run`) so commands can return errors
- Register subcommands via `Register*Commands(root *cobra.Command)` functions

### Common Tasks

```bash
# Build the binary
make build

# Run the full test suite
make test

# Run linters (go vet + golangci-lint)
make lint

# Format all Go code
make fmt

# Tidy module dependencies
make tidy

# Cross-compile for all platforms
make release

# Clean build artifacts
make clean

# Run a specific test with verbose output
go test ./internal/retry/... -run TestExponentialBackoff -v

# Check test coverage for a package
go test -coverprofile=coverage.out ./internal/cache/...
go tool cover -func=coverage.out

# Clear local config/cache during development
rm -rf ~/.config/notion-cli/
```

### Adding a New Command

1. Create a new file in `internal/cli/commands/` (e.g., `newcmd.go`)
2. Define the command using `&cobra.Command{}`
3. Create a `RegisterNewCmdCommands(root *cobra.Command)` function
4. Call the register function from `internal/cli/root.go`
5. Add tests in a corresponding `_test.go` file
6. Use `pkg/output.Printer` for all output
7. Use `internal/errors.NotionCLIError` for error handling

## Questions?

- Check existing issues and PRs first
- Open a discussion on GitHub Discussions
- Review documentation in `/docs` folder

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to notion-cli!
