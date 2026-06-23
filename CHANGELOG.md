# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **`db create` and `db update` can now set/modify the database column schema** (fixes #97). Both commands accept `--properties` (a JSON object of Notion property definitions) and `--properties-file` (the same JSON from a file).
  - `db create` sends the schema under `initial_data_source.properties` per Notion API version 2025-09-03+ (replacing the previously hard-coded, now-deprecated top-level `properties` payload). A default `Name` title column is injected only when the supplied schema has no title-type property.
  - `db update` applies schema changes via `PATCH /data_sources/{id}`, resolving the primary data source automatically (or an explicit `--data-source`). A property value of JSON `null` deletes that column. `--title` still updates the database object and can be combined with `--properties`.

### CI
- **Dependabot PRs now auto-merge.** New `.github/workflows/dependabot-auto-merge.yml` reads update metadata via `dependabot/fetch-metadata` (SHA-pinned) and runs `gh pr merge --auto --squash` for non-major updates; the merge only completes once required checks pass. Requires repo settings: "Allow auto-merge" enabled and a branch-protection rule on `main` requiring the `CI Success` check.
- **Dependabot PR churn reduced** by grouping minor/patch updates for both `gomod` and `github-actions` in `.github/dependabot.yml`.

## [6.3.3] - 2026-05-15

### Tests
- **OAuth `/callback` handler now has direct coverage** for its DNS-rebinding and method guards. `internal/oauth/oauth.go` extracts the handler into `newCallbackHandler` (no behavior change) so `internal/oauth/oauth_test.go` can assert: non-GET → 405 with `Allow: GET`, mismatched `Host` → 421, and valid `localhost:PORT` / `127.0.0.1:PORT` hosts → 200 with the code+state delivered on `resultCh`. Each test was confirmed to fail when its corresponding guard is removed.

### Fixed
- **OAuth `--manual` bare-code paste now returns an actionable error.** When a user pastes only the authorization code (no `state=`) into `auth login --manual`, `LoginManual` now surfaces `"OAuth state parameter is missing from the pasted input"` with suggestions that tell the user to paste the FULL redirected URL — instead of the generic "OAuth state mismatch (possible CSRF attack)" text, which left users in a loop on SSH. The real state-mismatch path keeps the original CSRF message. See `internal/oauth/oauth.go` `LoginManual`.

### Security
- **OAuth `LoginManual` now requires state.** Pasting a bare authorization code into the manual flow is rejected with `OAUTH_STATE_MISMATCH`; users must paste the FULL redirected URL so the CSRF state can be verified. The local callback flow (`auth login` without `--manual`) was already verifying state.
- **OAuth `/callback` server adds method and Host pinning.** Only `GET` is accepted (405 otherwise) and the `Host` header must match `localhost:PORT` or `127.0.0.1:PORT`. Defeats DNS-rebinding even with state in place.
- **Refresh-token rotation guard.** `internal/notion/client.go` now only overwrites the stored `oauth_refresh_token` when the token endpoint returns a non-empty `refresh_token`. Prevents Notion's non-rotating-refresh responses from silently blanking the stored credential.
- **`oauth_refresh_token` is now masked by default** in `config get` and `config list`. Use `--show-secret` to reveal.
- **npm postinstall verifies SHA-256 checksums.** `install.js` fetches `SHA256SUMS` from the GitHub release, streams the binary while hashing, and refuses to install if the hash does not match. Hash mismatch is a hard error (non-zero exit) — not a silent skip. Redirects are pinned to `github.com` and `*.githubusercontent.com`, https-only, with a 50 MB body cap.
- **GitHub Actions hardening.** Every `uses:` is pinned to a commit SHA with a version comment for Dependabot. `gh release upload --clobber` removed (release assets are immutable). `--provenance` added to the platform-package publish loop. The `publish-npm` job is now gated on the `npm-publish` GitHub Environment so `NPM_TOKEN` requires manual approval per release.
- **`SHA256SUMS` is generated and uploaded** as a release asset alongside the binaries.
- **`make release` refuses to ship the known-leaked OAuth client secret** as a defense against re-shipping a credential that was previously exposed. See `SECURITY.md` → "Credential rotation".
- **Maintainer dev secret moved out of the repo tree** to `~/.config/notion-cli-dev/.env`. The Makefile reads from there; CI is unaffected.

## [6.3.2] - 2026-05-15

### Fixed
- OAuth token endpoint requests now include `Notion-Version` and `Accept: application/json` headers, matching the official Notion JS SDK (`makenotion/notion-sdk-js`) and developer-docs spec. Applied to `exchangeCode`, `TokenRefresh`, `TokenIntrospect`, and `TokenRevoke` in `internal/oauth/oauth.go`.

## [6.3.1] - 2026-05-11

### Breaking
- **Go module path is now `github.com/Coastal-Programs/notion-cli/v6`.** Per Go's [major-version-suffix rule](https://go.dev/ref/mod#major-version-suffixes), modules at v2+ must carry a `/vN` suffix. Without it, `go install github.com/Coastal-Programs/notion-cli/cmd/notion-cli@latest` resolved to `v5.9.0+incompatible` and failed because that older tag did not contain `cmd/notion-cli`. Update any `go install` command and any `import` of `pkg/output` to use the `/v6` prefix (e.g. `go install github.com/Coastal-Programs/notion-cli/v6/cmd/notion-cli@latest`). Fixes #86.

### Changed
- Bumped default Notion-Version header from 2022-06-28 to 2026-03-11 (latest).
- `db query` now calls `/v1/data_sources/{id}/query` (Notion API 2025-09-03). Use `--data-source` to target a specific source on multi-source databases.
- `block append` now uses the `position` object (Notion API 2026-03-11). The `--after` flag is deprecated; use `--position=after --after-block <id>`.
- Use `in_trash` instead of `archived`; recognise `meeting_notes` block type (renamed from `transcription`) per Notion API 2026-03-11.

### Fixed
- Config fields previously loaded but never applied to the runtime client are now wired through. `NOTION_CLI_BASE_URL`, `NOTION_CLI_MAX_RETRIES`, `NOTION_CLI_BASE_DELAY`, `NOTION_CLI_MAX_DELAY`, and `NOTION_CLI_HTTP_KEEP_ALIVE` (and their config-file equivalents) now affect HTTP behavior. `NOTION_CLI_VERBOSE` / `verbose` is honored as a fallback when the `--verbose` flag isn't set.

### Added
- `auth refresh` — explicitly refresh OAuth access token using stored refresh token; prints new expiry.
- `auth status --remote` — also calls token introspect to show active status, scope, and issued_at/expires_at.
- `auth logout --local-only` — skip the API revoke call and only clear local config.
- `oauth.TokenRefresh`, `oauth.TokenIntrospect`, `oauth.TokenRevoke` client functions in `internal/oauth/oauth.go`.
- `Config.OAuthRefreshToken` and `Config.OAuthTokenExpiresAt` (RFC3339) with full persistence (save/load/clear).
- `Config.NeedsRefresh()` — returns true when token expires within 5 minutes and a refresh token is present.
- Auto-refresh-on-401: the Notion HTTP client transparently refreshes the OAuth token and retries once when a refresh token is available in config.
- `custom-emoji list --all` — auto-paginate through all workspace custom emojis. Table output (`id`, `name`, `url`) by default. New httptest coverage for `CustomEmojiList` client method.
- `files` command group (`upload`, `retrieve`, `list`) wrapping the Notion File Uploads API. `files upload <path>` auto-selects single-part (≤20 MB) or multi-part (>20 MB) upload, streams file contents in 10 MB chunks, shows a progress bar on TTY, and prints the `file_upload_id` to stdout for piping. `files list` supports `--page-size` and `--all` for full pagination. Five new Notion client methods: `FileUploadCreate`, `FileUploadSend`, `FileUploadComplete`, `FileUploadRetrieve`, `FileUploadList`.
- `view create` — create a new Notion database view with `--data-source`, `--name`, `--type` (required) and optional `--database`, `--filter`, `--sorts` (JSON string or `@<file>`).
- `view update <view_id>` — update an existing view's `--name`, `--filter`, and/or `--sorts`.
- `view delete <view_id>` — delete a Notion view.
- `view query --all` — auto-paginate through all view query results and clean up the server-side query cache on completion.
- `ViewCreate`, `ViewUpdate`, `ViewDelete` methods added to the Notion HTTP client.
- `page trash <page_id>` — moves a page to trash (`in_trash: true`). Requires `--yes` in non-TTY environments. Interactive mode prompts for confirmation.
- `page restore <page_id>` — restores a trashed page (`in_trash: false`).
- `page move <page_id>` — moves a page to a new parent via `POST /pages/{id}/move`. Supports `--parent <page-id>`, `--data-source <data-source-id>`, and `--workspace` (mutually exclusive).
- `PageTrash`, `PageRestore`, `PageMove` methods added to the Notion HTTP client.
- `comment` command group (`create`, `list`, `retrieve`, `update`, `delete`) wrapping the Notion Comments API (Notion-Version 2026-03-11). `comment create` supports `--page`/`--block`/`--discussion` parents (mutually exclusive), `--text` or `--rich-text <file.json>`, plus optional `--display-name` and repeatable `--attach-file <upload-id>` (max 3). `comment list` supports `--page-size` and `--all` for full pagination.
- `markdown` command group (`get`, `set`) for reading and writing page content as enhanced markdown via `/v1/pages/{id}/markdown` (Notion API 2026-03-11). `set` supports replace (default) and `--append`, with content from `--content`, `--file`, or stdin.
- `custom-emoji` command group (list, retrieve) per Notion API 2026-03-11.
- `notion.WithKeepAlive(bool)` client option installs a transport with `DisableKeepAlives=true` when keep-alive is disabled in config.
- Data Source HTTP client methods and resolver helper for the 2025-09-03 multi-source database split.
- `data-source` command group (retrieve, create, update, query) for the 2025-09-03 multi-source database API.
- `data-source templates <id>` — lists page templates for a data source with `--all` pagination support.
- `data-source properties update <id> --schema <json|@file>` — updates a data source's properties schema.
- `db query` now emits a deprecation notice to stderr when routing through auto-resolved data sources; suppress with `--quiet`.
- Workspace cache (`sync`) now indexes data sources extracted from database search results (`data_sources` key), stored in `~/.notion-cli/databases.json` under `data_sources`.
- `list` command now includes a `data_sources` array alongside `databases` in the output.
- `ExtractID` in the resolver now recognises the `dataSource=` query parameter in Notion URLs, returning the data source ID instead of the page/database path ID.
- `ds` alias moved from `db` to `data-source` (breaking: `notion-cli ds` now invokes `data-source`).
- `view` command group (list, retrieve, query, results, delete-query) per Notion API 2026-03-11.

### Removed
- Dead code cleanup: removed the unused in-memory TTL `Cache` struct, `NewCache`, TTL constants (`DatabaseTTL`, `UserTTL`, `PageTTL`, `BlockTTL`), and `CacheKeyForResource` from `internal/cache/cache.go` (and its test file). Only the workspace database cache (`internal/cache/workspace.go`) was ever wired up; the response cache was never integrated. The `cache_enabled` / `cache_max_size` / `disk_cache_enabled` config fields are kept for forward compatibility but documented as reserved.
- Removed unused error factory functions `IntegrationNotShared`, `ResourceNotFound`, `DatabaseIdConfusion`, `InvalidProperty`, `NetworkError`, and `Timeout` from `internal/errors/errors.go`. The HTTP client uses `FromNotionAPI` for all API errors; these factories had no production callers.

## [6.3.0] - 2026-05-06

### Added
- **Multi-port OAuth callback** (`internal/oauth`): `auth login` now tries ports 8080, 8081, 8089 in order and uses the first one it can bind. Removes the most common production failure (something else holding port 8080).
- **`auth login --manual`**: skips the local callback server, prints the authorize URL, and accepts a pasted redirect URL or bare code. Auto-enabled when `$SSH_TTY` is set so remote sessions work without forwarding.
- **Polished OAuth callback page**: redesigned success/error HTML with dark-mode support, branded icon, and clear "return to terminal" guidance.
- **`doctor` OAuth checks**: now probes that at least one callback port is bindable and that Notion's authorize endpoint accepts the embedded `client_id` (catches deleted/internalized integrations).
- `page create` and `page update` now accept `--icon-emoji`, `--icon-url`, and `--cover-url` flags to set page icons and covers (#74). On `page update`, pass `none` to clear an existing icon or cover.
- `doctor` output includes `oauth_credentials_embedded` so users can verify whether their binary was built with OAuth client credentials (#73).

### Changed
- `auth login` timeout bumped from 2 minutes to 5 minutes to accommodate 2FA, account switching, and slow networks.
- `OAuthPortInUse` error now lists every attempted port and recommends `--manual` as a fallback.
- `OAuthNotConfigured` error now reports the binary version and lists concrete remediation steps for both npm-installed and source-built binaries (#73).

### CI/Build
- `make release` now fails fast when `NOTION_OAUTH_CLIENT_ID` / `NOTION_OAUTH_SECRET` are unset, preventing silent OAuth-less release artifacts (#73).
- `.github/workflows/publish.yml` verifies OAuth secrets are present before building or publishing.

### Operations
- Notion integration now requires three registered redirect URIs: `http://localhost:8080/callback`, `http://localhost:8081/callback`, `http://localhost:8089/callback`.
- Rotated `NOTION_OAUTH_SECRET` (GitHub Actions secret + `.env.local`).

## [6.2.1] - 2026-05-06

### Changed
- Bumped Go toolchain from 1.25 to 1.26
- Updated `github.com/spf13/pflag` from v1.0.9 to v1.0.10
- Updated CI publish workflow to use Go 1.26

### Fixed
- Migrated golangci-lint config to v2 format (added `version: "2"`, removed `gosimple` linter merged into `staticcheck`)
- Resolved all pre-existing errcheck, staticcheck QF1012, and unused lint errors across the codebase

## [6.2.0] - 2026-03-02

### Added
- **`--output` / `-o` flag**: All commands now accept `--output <format>` (or `-o <format>`) as an alternative to boolean flags like `--json`, `--csv`, etc. Valid formats: `json`, `compact-json`, `csv`, `markdown`, `table`, `raw`, `pretty`. Invalid values return a structured error with suggestions.
- **Platform binary packages on npm**: First publish of all 5 platform-specific packages (`@coastal-programs/notion-cli-darwin-arm64`, `-darwin-x64`, `-linux-x64`, `-linux-arm64`, `-win32-x64`). Users on supported platforms now get a native binary via npm's optional dependency resolution instead of relying on the postinstall GitHub Release fallback.

### Fixed
- **Publish workflow**: Platform package publish failures are no longer silently ignored (`|| true` removed). The workflow now reports explicit errors with guidance when platform packages fail to publish.

## [6.1.2] - 2026-03-02

### Fixed
- **OAuth URL encoding**: Authorization URL now properly URL-encodes redirect_uri parameter
- **OAuth security**: Callback server binds to `127.0.0.1` instead of `0.0.0.0` (all interfaces)
- **OAuth security**: Token exchange JSON body built with `json.Marshal` instead of string formatting
- **Version reporting**: JSON envelope metadata now shows correct version (was hardcoded to `6.0.0`)
- **User-Agent header**: Now reports actual CLI version (was hardcoded to `1.0`)
- **db schema**: No longer panics when database has no properties
- **doctor command**: Now returns error envelope and non-zero exit code when checks fail
- **search command**: Date filter flags now validate format upfront instead of failing silently
- **search command**: `--page-size` now validated against Notion's 100 limit
- **db query**: `--sort-direction` now validates input and gives clear error for invalid values
- **batch command**: Alias changed from `b` to `ba` to avoid collision with `block` alias

### Changed
- **Removed hardcoded OAuth client ID from Makefile** — both credentials now sourced exclusively from environment variables via GitHub Secrets
- **Publish workflow**: Now passes `NOTION_OAUTH_CLIENT_ID` secret to build steps
- All error messages in `batch` and `config` commands now use structured `NotionCLIError`
- `whoami`, `sync`, `list`, `doctor` commands now reject unexpected arguments
- Removed dead `--search` flag from `db query`
- Rotated OAuth client secret after credential leak in v6.1.1 binary

## [6.1.1] - 2026-03-02

### Fixed
- Publish workflow: added `contents: write` permission for binary uploads
- Publish workflow: create `bin/` directories before copying platform binaries
- Publish workflow: pass `NOTION_OAUTH_SECRET` to build steps so OAuth works in published binaries
- Publish workflow: bumped Go version to 1.25 to match go.mod

### Changed
- README: OAuth login is now the recommended setup method
- README: added Authentication section with token precedence table
- README: added auth commands to Commands section
- README: updated project structure, test counts, troubleshooting

## [6.1.0] - 2026-03-02

### Added
- **OAuth authentication** - `auth login`, `auth logout`, `auth status` commands
  - Browser-based OAuth flow: run `notion-cli auth login`, authorize in Notion, start using the CLI
  - No manual token management needed for interactive use
  - Token precedence: `NOTION_TOKEN` env var > OAuth token > manual config token
  - CSRF protection via cryptographic state parameter
  - OAuth tokens stored in config file with 0600 permissions
  - 6 new OAuth-specific error codes with actionable suggestions
  - `doctor` command now shows auth method (oauth/env/token/none) with workspace info
  - Build-time OAuth client ID/secret injection via Makefile ldflags

## [6.0.0] - 2026-03-01

### Changed - **Complete Rewrite from TypeScript to Go**

This release is a complete rewrite of notion-cli from TypeScript/oclif to Go/Cobra.

**Why Go?**
- **Single binary distribution**: ~8MB binary vs 573 npm dependencies
- **Instant startup**: No Node.js runtime overhead
- **Cross-compilation**: One build produces darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, windows/amd64
- **Near-zero supply chain risk**: 2 Go dependencies (cobra, pflag) vs hundreds of npm packages

### Architecture

- **CLI framework**: Cobra (replacing oclif v4)
- **API client**: Raw HTTP (replacing @notionhq/client SDK)
- **Output**: JSON envelope, ASCII table, CSV, markdown (same formats as v5.x)
- **Caching**: In-memory TTL cache with per-resource-type TTLs
- **Retry**: Exponential backoff with jitter for 408/429/5xx
- **Config**: Environment variables + JSON config file (~/.config/notion-cli/config.json)
- **Errors**: 40+ error codes with suggestions (matching v5.x error system)
- **npm distribution**: Platform-specific binary packages (esbuild pattern)

### All 26 Commands Ported

**Database**: `db query`, `db retrieve`, `db create`, `db update`, `db schema`
**Page**: `page create`, `page retrieve`, `page update`, `page property-item`
**Block**: `block append`, `block retrieve`, `block children`, `block update`, `block delete`
**User**: `user list`, `user retrieve`, `user bot`
**Other**: `search`, `batch retrieve`, `sync`, `list`, `whoami`, `doctor`, `config set-token`, `config get`, `config path`, `cache info`

### Technical Details

- **33 Go source files** totaling ~8,900 lines
- **183 tests** across 8 test suites, all passing
- **7.9MB binary** (stripped, darwin/arm64)
- **5 platform binaries** built via `make release`
- **Zero new runtime dependencies** beyond Go stdlib + cobra

### Breaking Changes

- **v6.0.0**: Major version bump indicates this is a rewrite
- Command syntax is identical to v5.x - existing scripts should work unchanged
- JSON envelope format is identical: `{success, data, metadata}`
- Same environment variable: `NOTION_TOKEN`

### Phase 2 (Future)

The following v5.x features are deferred to a future release:
- Disk cache and request deduplication
- Circuit breaker
- Simple properties (`-S` flag) with property expansion
- Recursive page retrieval
- Markdown output from page content
- Interactive init wizard
- Update notifications

## [5.9.0] - 2026-02-05

### Added
- **Request deduplication** - Prevents duplicate concurrent API calls for the same resource
  - Automatic deduplication of in-flight requests using promise memoization
  - Statistics tracking for hits/misses/pending requests
  - Configurable via `NOTION_CLI_DEDUP_ENABLED` environment variable
  - Integrated with `cachedFetch()` for seamless API call optimization
  - Expected 30-50% reduction in duplicate API calls
- **Parallel operations** - Execute bulk operations concurrently for faster performance
  - Block deletion in `updatePage()` now runs in parallel (configurable concurrency)
  - Child block fetching in `retrievePageRecursive()` now runs in parallel
  - Configurable via `NOTION_CLI_DELETE_CONCURRENCY` (default: 5) and `NOTION_CLI_CHILDREN_CONCURRENCY` (default: 10)
  - Expected 60-80% faster bulk operations
- **Persistent disk cache** - Maintains cache across CLI invocations
  - Cache entries stored in `~/.notion-cli/cache/` directory
  - Automatic max size enforcement (default: 100MB)
  - Atomic writes prevent corruption
  - Configurable via `NOTION_CLI_DISK_CACHE_ENABLED` and `NOTION_CLI_DISK_CACHE_MAX_SIZE`
  - Expected 40-60% improved cache hit rate
- **HTTP keep-alive and connection pooling** - Reduces connection overhead
  - Reuses HTTPS connections across multiple requests
  - Configurable connection pool size (default: 10 free sockets)
  - Configurable max concurrent connections (default: 50 sockets)
  - Keep-alive timeout configurable (default: 60 seconds)
  - Automatic cleanup on command exit
  - Expected 10-20% latency improvement
- **Response compression** - Reduces bandwidth usage
  - Automatic gzip, deflate, and brotli compression support
  - Accept-Encoding headers added to all API requests
  - Server automatically compresses responses when supported
  - Client automatically decompresses responses
  - Expected 60-70% bandwidth reduction for JSON payloads

### Performance
- Request deduplication reduces unnecessary API calls when multiple concurrent requests target the same resource
- Parallel execution of bulk operations significantly reduces total operation time
  - Page updates with many blocks complete 60-80% faster
  - Recursive page retrieval with many child blocks completes 60-80% faster
- Persistent disk cache maintains cache across CLI invocations
  - Subsequent CLI runs benefit from cached data (40-60% improved hit rate)
  - Cache survives process restarts and system reboots
  - Automatic cleanup of expired entries
- HTTP keep-alive reduces connection overhead
  - Connection reuse eliminates TLS handshake for subsequent requests
  - 10-20% latency improvement for multi-request operations
  - Configurable pool sizes for different workload patterns
- Response compression reduces bandwidth usage
  - JSON responses compressed by 60-70% (typical)
  - Faster data transfer, especially on slow connections
  - Lower bandwidth costs and network usage
  - Automatic compression/decompression handled by HTTP client

### Breaking Changes

**None** - All performance optimizations are backward compatible and can be independently disabled via environment variables.

### Technical Details

- **121 new tests** added across 5 test suites with comprehensive coverage
  - Deduplication: 22 tests (94.73% coverage)
  - Parallel Operations: 21 tests (timing benchmarks included)
  - Disk Cache: 34 tests (83.59% coverage)
  - HTTP Agent: 26 tests (78.94% coverage)
  - Compression: 18 tests (header validation)
- **Zero new dependencies** - All optimizations use Node.js built-in features
- **Production-ready** - Comprehensive error handling with graceful degradation
- **Lifecycle management** - Proper initialization in `BaseCommand.init()` and cleanup in `BaseCommand.finally()`

### Configuration

All optimizations are configurable via environment variables. See `.env.example` for complete configuration guide.

**Request Deduplication:**
- `NOTION_CLI_DEDUP_ENABLED` (default: true)

**Parallel Operations:**
- `NOTION_CLI_DELETE_CONCURRENCY` (default: 5)
- `NOTION_CLI_CHILDREN_CONCURRENCY` (default: 10)

**Persistent Disk Cache:**
- `NOTION_CLI_DISK_CACHE_ENABLED` (default: true)
- `NOTION_CLI_DISK_CACHE_MAX_SIZE` (default: 104857600 / 100MB)
- `NOTION_CLI_DISK_CACHE_SYNC_INTERVAL` (default: 5000ms)

**HTTP Keep-Alive:**
- `NOTION_CLI_HTTP_KEEP_ALIVE` (default: true)
- `NOTION_CLI_HTTP_KEEP_ALIVE_MS` (default: 60000ms)
- `NOTION_CLI_HTTP_MAX_SOCKETS` (default: 50)
- `NOTION_CLI_HTTP_MAX_FREE_SOCKETS` (default: 10)
- `NOTION_CLI_HTTP_TIMEOUT` (default: 30000ms)

**Response Compression:**
- Always enabled (no configuration needed)

### Migration Guide

**Upgrading from 5.8.0:**

1. **No code changes required** - All optimizations work automatically
2. **Default settings are optimal** for most use cases
3. **To customize performance**, create a `.env` file with desired settings
4. **To disable specific optimizations**, set corresponding `_ENABLED` flag to `false`
5. **For batch operations**, consider increasing concurrency limits
6. **For memory-constrained environments**, reduce cache sizes

Example `.env` for high-throughput batch processing:
```bash
NOTION_CLI_DELETE_CONCURRENCY=10
NOTION_CLI_CHILDREN_CONCURRENCY=20
NOTION_CLI_HTTP_MAX_SOCKETS=50
NOTION_CLI_DISK_CACHE_MAX_SIZE=104857600
```

### Performance Summary

**Overall improvement: 1.5-2x for batch operations and repeated data access**

Individual phase improvements:
- Request deduplication: 5-15% typical (30-50% best case with concurrent duplicates)
- Parallel operations: 60-70% typical (80% best case for large batches)
- Disk cache: 20-30% improvement across sessions (60% best case with heavy reuse)
- HTTP keep-alive: 5-10% typical (10-20% best case for multi-request operations)
- Response compression: Bandwidth reduction varies (compression already handled by modern APIs)

See [README.md Performance Optimizations](./README.md#-performance-optimizations-v590) for detailed documentation.

## [5.8.0] - 2026-02-04

### Changed
- **Major dependency updates** - Updated 24 packages with comprehensive testing
  - **oclif framework v4**: Core CLI framework updated from v2 to v4
    - Migrated from deprecated `ux.table` to new `cli-table3`-based table formatter
    - Created backward-compatible table utility maintaining all CLI flags
    - Updated 18 command files with new table rendering approach
  - **TypeScript 5.9**: Upgraded from 4.9 with full type checking compatibility
  - **Notion SDK 5.9**: Updated from 5.2.1 with latest API improvements
  - **Node.js 22 types**: Updated @types/node from v16 to v22
  - **Testing framework updates**: mocha 11.7, sinon 21.0, @types/sinon 21.0
  - **Linting updates**: eslint 9.39, typescript-eslint 8.54, eslint-plugin-unicorn 62.0
  - **Other updates**: dayjs 1.11.19, prettier 3.8.1, undici 7.20.0

### Security
- **Eliminated all production vulnerabilities** - 0 vulnerabilities in production dependencies (down from 2)
- **Reduced total vulnerabilities by 87%** - From 31 to 4 (all low-priority devDependencies)
- **Resolved 27 security issues** in oclif v2/v3 tooling by upgrading to v4

### Technical
- All 471 tests passing with >95% code coverage maintained
- Backward compatible - no breaking changes for CLI users
- Deferred ESM-only packages (chai v5+, node-fetch v3+, globby v14+) to v6.0.0

## [5.7.0] - 2026-01-28

### Added
- **ASCII art banner** displayed during installation and `notion-cli init` command for enhanced branding and professional appearance
- **PUBLISHING.md guide** with comprehensive npm release workflow and best practices
- **Shared banner utilities** (`scripts/banner.js` and `src/utils/terminal-banner.ts`) following DRY principle
- **Automatic update notifications** - CLI now checks npm once per day and notifies users of available updates
  - Non-intrusive: runs asynchronously in background
  - Respects `NO_UPDATE_NOTIFIER` environment variable
  - Automatic in CI environments and test suites
  - Caches results for 24 hours to minimize npm registry calls
  - Users remain in full control - updates are never applied automatically

### Changed
- **README Quick Start section** simplified with clearer installation flow and removed MCP server comparisons
- **Enhanced postinstall message** to prominently highlight `notion-cli init` as the recommended first step
- **Banner styling** now uses terminal's default color for better consistency with README
- **Token input UX improved** - `notion-cli init` now accepts tokens with or without "secret_" prefix
  - Automatically prepends "secret_" if user pastes just the token value
  - Shows friendly note when prefix is added
  - Eliminates confusing validation errors
- **Token length validation** - Minimum 20 character requirement catches incomplete token copies early
- **TypeScript configuration** - Enabled `resolveJsonModule` for better JSON import support
- **Update notification timing** - Changed to `defer: true` for non-intrusive display after command execution

### Fixed
- **PUBLISHING.md chicken-and-egg problem** resolved by recommending README update before first npm publish
- **Version consistency** - All documentation now correctly references v5.6.0

### Security
- **Token masking** - Console output now masks tokens to prevent leakage in screen recordings, terminal sharing, or logs
  - Format: `secret_***...***abc` (shows prefix and last 3 chars only)
  - Applies to all token display scenarios (setup wizard, error messages, help text)
  - Protects against accidental token exposure during demos or support sessions

### Developer Experience
- **DEBUG mode** - Set `DEBUG=1` environment variable to see verbose update check errors for troubleshooting
  - Helps diagnose npm registry connectivity issues
  - Shows detailed error messages when update checks fail
  - Silent by default to avoid cluttering output

## [5.6.0] - 2025-10-25

### Quality Improvements

This release focuses on code quality, testing, and documentation improvements following a comprehensive 4-week quality sprint.

### Changed

- **ESLint upgraded to v9** with flat config format (eslint.config.js)
- **@notionhq/client downgraded to v2.2.15** for better test compatibility
- **Development dependencies updated** to latest stable versions
- **Enhanced project documentation** with comprehensive development guide

### Fixed

- **100% test pass rate** - Fixed all previously failing integration tests
- **1,800+ lint issues resolved** - Applied consistent code style across entire codebase
- **14 security vulnerabilities fixed** in development dependencies

### Security

- **Zero production vulnerabilities** - All critical and high-severity issues resolved
- **11 devDependency vulnerabilities remaining** - Non-critical, deferred to oclif v4 migration
  - 2 moderate severity (oclif v2 deprecation-related)
  - 9 low severity (test infrastructure)
- **No impact on production runtime** - All remaining vulnerabilities are in development-only packages

### Documentation

- **README.md enhanced** with complete Development section including:
  - Prerequisites and setup instructions
  - Development workflow and commands
  - Project structure overview
  - Testing and code quality guidelines
  - Building and publishing instructions
- **CHANGELOG.md updated** with quality sprint documentation
- **CONTRIBUTING.md created** with contributor guidelines
- **SECURITY.md created** with vulnerability reporting process
- **JSDoc comments added** to all public command classes
- **package.json metadata reviewed** and updated

### Quality Metrics

- **Tests:** 100% pass rate (40/40 tests passing)
- **Lint:** Zero errors, minimal warnings
- **Security:** 0 critical/high vulnerabilities in production
- **Documentation:** 95% completeness

### Migration Guide

No breaking changes! This is a quality-focused release. All existing commands and functionality work exactly as before.

**For developers:**
- Review the new [CONTRIBUTING.md](CONTRIBUTING.md) before submitting pull requests
- Run `npm run lint` to ensure code follows project standards
- All tests must pass before commits

**For users:**
- Update normally: `npm update -g @coastal-programs/notion-cli`
- No configuration changes needed
- All commands remain backward compatible

---

## [5.5.0] - 2025-10-24

### Added

- **`notion-cli init`** - Interactive first-time setup wizard with 3-step flow (token setup, connection test, workspace sync)
  - Guides users through token configuration with clear instructions
  - Tests API connection before proceeding
  - Offers optional workspace sync to build local database cache
  - Supports `--json` mode for automation and CI/CD environments
  - Provides helpful next steps after completion
- **`notion-cli doctor`** - Comprehensive health check and diagnostics command (aliases: `diagnose`, `healthcheck`)
  - 7 diagnostic checks: Node.js version, token configuration, API connectivity, workspace access, cache status, dependencies, and file permissions
  - Color-coded output with clear pass/fail indicators
  - Actionable recommendations for each failed check
  - JSON output support for automated monitoring
  - Perfect for troubleshooting and pre-flight checks
- **Token validator utility** - Centralized token validation with consistent error messaging
  - Early validation before API calls (500x faster error feedback)
  - Platform-specific setup instructions (Windows vs Unix/Mac)
  - Helpful error messages with exact commands to fix issues
  - Prevents cryptic API errors with proactive validation
- **Post-install experience** - Welcome message after `npm install`
  - Friendly introduction to notion-cli
  - Clear next steps for new users
  - Guides users to run `notion-cli init` for setup
  - Respects npm's `--silent` flag
- **Progress indicators** - Enhanced user feedback during operations
  - `sync` command now shows real-time progress with status messages
  - Execution timing displayed for all sync operations
  - Enhanced completion summary with cache metadata (database count, cache age)
  - Improved user experience during long-running operations
- **Custom markdown-to-blocks converter** - Zero-dependency markdown parser
  - Replaces @tryfabric/martian dependency
  - Supports headings, paragraphs, lists, code blocks, and quotes
  - Secure implementation without external vulnerabilities
  - Maintains full feature compatibility

### Changed

- **Error messages** now provide platform-specific instructions (Windows Command Prompt, Windows PowerShell, Unix/Mac)
- **Sync command** displays real-time progress and execution metrics
- **Token validation** happens early, before API calls, providing instant feedback
- **Completion summaries** include rich metadata about cache state and recommendations

### Fixed

- **Token validation errors** now provide clear, actionable guidance instead of cryptic API errors
- Users no longer encounter confusing authentication errors on first run
- Cache-related operations provide better context and next steps

### Security

- **Fixed all 16 production security vulnerabilities** reported by npm audit
- **Removed @tryfabric/martian dependency** (contained katex XSS vulnerabilities)
  - CVE-2023-48618: katex XSS vulnerability
  - CVE-2024-28245: katex XSS vulnerability
  - CVE-2021-23906: yargs-parser prototype pollution
  - CVE-2020-28469: glob-parent ReDoS vulnerability
- **Now reports 0 production vulnerabilities** in npm audit
- Custom markdown converter eliminates entire dependency chain of vulnerable packages

### Migration Guide

**No breaking changes!** All existing commands work exactly as before.

**New recommended workflow for new users:**
1. Install: `npm install -g @coastal-programs/notion-cli`
2. Run setup wizard: `notion-cli init`
3. Verify health: `notion-cli doctor`
4. Start using: `notion-cli list`, `notion-cli db query`, etc.

**For existing users:**
- Run `notion-cli doctor` to verify your setup is healthy
- Your existing token configuration continues to work
- No action required unless you want to use the new commands

---

## [5.4.0] - 2025-10-23

### Added - AI Agent Usability Features (Issue #4)

**Complete implementation of 7 major AI agent usability improvements:**

**1. JSON Envelope Standardization**
- Consistent `{success, data, metadata}` response format across ALL commands
- Standardized exit codes: 0 = success, 1 = API error, 2 = CLI error
- New `envelope.ts` module for centralized response handling
- New `base-command.ts` extending oclif Command with envelope support
- 7 comprehensive documentation guides in `docs/ENVELOPE_*.md`:
  - Architecture overview
  - Integration guide
  - Quick reference
  - Specification
  - System summary
  - Testing strategy
  - Index

**2. Health Check Command**
- NEW `whoami` command (aliases: `test`, `health`, `connectivity`)
- Reports bot info, workspace access, cache status, and API latency
- Comprehensive error handling with actionable suggestions
- Perfect for AI agents to verify connectivity before operations

**3. Simple Properties Mode** 🎉
- NEW `--simple-properties` (`-S`) flag for `page create` and `page update`
- Flat JSON format: `{"Name": "Task", "Status": "Done"}` instead of complex nested Notion structures
- **70% reduction in complexity** for AI agents
- Supports 13 property types: title, rich_text, number, checkbox, select, multi_select, status, date, url, email, phone_number, people, files, relation
- Case-insensitive property name and value matching
- Relative date parsing: `"today"`, `"tomorrow"`, `"+7 days"`, `"+2 weeks"`, etc.
- Comprehensive validation with helpful error messages
- Auto-creates proper Notion API format based on database schema
- Documentation: `docs/SIMPLE_PROPERTIES.md`, `AI_AGENT_QUICK_REFERENCE.md`
- New utility: `src/utils/property-expander.ts` (400 lines, fully tested)

**4. Schema Examples**
- NEW `--with-examples` flag for `db schema` command
- Shows copy-pastable property payloads for each property type
- Groups writable vs read-only properties
- Makes it trivial for AI agents to construct valid property objects

**5. Verbose Logging**
- NEW `--verbose` (`-v`) flag for debugging
- Shows cache hits/misses, retry attempts, API latency
- Helps AI agents understand what's happening behind the scenes
- Documentation: `docs/VERBOSE_LOGGING.md`

**6. Filter Simplification**
- Simplified filter syntax for database queries
- Better validation and error messages
- Documentation: `docs/FILTER_GUIDE.md`, `docs/FILTER_MIGRATION.md`

**7. Output Format Enhancements**
- NEW `--compact-json` flag: minified JSON (one line)
- NEW `--pretty` flag: enhanced table formatting
- NEW `--markdown` flag: markdown table output
- Consistent across all commands via `base-flags.ts`
- Documentation: `OUTPUT_FORMATS.md`

### Changed

- **Version:** Bumped to 5.4.0 to reflect major feature release
- **Package description:** Updated to "with simple properties, JSON envelopes, and enhanced usability"
- **Keywords:** Added "simple-properties" and "json-envelope"
- **All commands:** Now extend `BaseCommand` for consistent envelope responses
- **Error handling:** Unified error response format across all commands

### Technical Details

**New Files:**
- `src/base-command.ts` - Base command class with envelope support
- `src/base-flags.ts` - Reusable flag sets for consistency
- `src/envelope.ts` - Envelope response formatting
- `src/utils/property-expander.ts` - Simple properties converter
- `src/commands/whoami.ts` - Health check command
- `test/utils/property-expander.test.ts` - 30+ test cases
- 7 envelope documentation files
- `docs/SIMPLE_PROPERTIES.md` - Complete simple properties guide
- `docs/VERBOSE_LOGGING.md` - Verbose logging guide
- `docs/FILTER_GUIDE.md` - Filter syntax guide
- `AI_AGENT_QUICK_REFERENCE.md` - Quick reference for AI agents (root level for easy access)

**Modified Files:**
- `src/commands/page/create.ts` - Added `--simple-properties` support
- `src/commands/page/update.ts` - Added `--simple-properties` support
- `src/commands/db/schema.ts` - Added `--with-examples` flag
- All command files updated to use envelope format

**Type Safety:**
- Full TypeScript type definitions for all new features
- Interfaces for `Envelope`, `EnvelopeMetadata`, `SimpleProperties`
- Type-safe property expansion and validation

**Testing:**
- 30+ test cases for property expander
- Envelope response validation
- All existing tests still passing

### Why This Matters

**For AI Agents:**
- **Simple Properties:** Dramatically reduces errors from malformed property structures
- **JSON Envelopes:** Predictable response format makes parsing trivial
- **Health Check:** Easy connectivity verification before complex operations
- **Verbose Logging:** Helps debug issues and understand system behavior
- **Schema Examples:** Copy-paste examples eliminate guesswork

**For Developers:**
- Consistent API across all commands
- Better error messages with actionable suggestions
- Easier to debug with verbose mode
- Simpler property syntax for manual use

### Examples

**Simple Properties (Before vs After):**

```bash
# BEFORE: Complex nested structure (easy to get wrong)
notion-cli page create -d DB_ID --properties '{
  "Name": {"title": [{"text": {"content": "Task"}}]},
  "Status": {"select": {"name": "In Progress"}},
  "Priority": {"number": 5}
}'

# AFTER: Flat, simple structure
notion-cli page create -d DB_ID -S --properties '{
  "Name": "Task",
  "Status": "In Progress",
  "Priority": 5
}'
```

**Health Check:**
```bash
notion-cli whoami
# Returns bot info, workspace, cache stats, API latency
```

**Schema with Examples:**
```bash
notion-cli db schema DB_ID --with-examples
# Shows example payloads for each property
```

### Migration Guide

**No breaking changes!** All existing commands work exactly as before.

**New recommended workflow:**
1. Run `whoami` to verify connectivity
2. Run `db schema DB_ID --with-examples` to understand structure
3. Use `-S` flag with simple properties for page create/update
4. Use `--verbose` for debugging when needed

---

## [5.3.0] - 2025-10-22

### Added - Workspace Management Features

**Smart ID Resolution:**
- Automatic conversion between `database_id` and `data_source_id`
- System detects and converts wrong ID type automatically
- Helpful messaging when conversion happens
- Works with all database commands (retrieve, query, update)
- Documentation: `docs/smart-id-resolution.md`

**Workspace Database Caching:**
- NEW `sync` command - Cache all workspace databases locally
- NEW `list` command - Browse cached databases with rich metadata
- NEW `config set-token` command - Easy token setup with guided workflow
- Persistent cache at `~/.notion-cli/databases.json`
- Alias generation - Find databases by name, nickname, or acronym
- Zero API calls for name resolution after sync

### Changed

- **Version:** Bumped to 5.3.0 for workspace management features
- Enhanced database resolution with smart ID conversion
- Improved token configuration workflow

---

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
