# Architecture

High-level map of `notion-cli`. For directory-by-directory layout see [`.github/PROJECT_STRUCTURE.md`](./.github/PROJECT_STRUCTURE.md); for code conventions see [`CLAUDE.md`](./CLAUDE.md).

## Overview

`notion-cli` is a single-binary Go CLI that wraps the Notion HTTP API. It is distributed via npm (wrapper package + platform-specific binary packages, esbuild-style) and optimized for AI-agent and automation use: structured JSON envelopes, deterministic exit codes, and machine-readable errors.

## Layered design

```
┌─────────────────────────────────────────────┐
│  cmd/notion-cli/main.go                     │  Entry point — wires Cobra root.
└─────────────────────────────────────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│  internal/cli + internal/cli/commands       │  Cobra commands (db, page, block, …).
│                                             │  Parse flags, call resolver, invoke client,
│                                             │  format via pkg/output.
└─────────────────┬───────────────────────────┘
                  │
       ┌──────────┼──────────┬─────────────┐
       ▼          ▼          ▼             ▼
┌───────────┐ ┌────────┐ ┌──────────┐ ┌──────────┐
│ resolver  │ │ cache  │ │  retry   │ │  errors  │
│ URL/ID/   │ │ TTL    │ │ backoff  │ │ NotionCLI│
│ name → ID │ │ in-mem │ │ + jitter │ │  Error   │
└─────┬─────┘ └───┬────┘ └────┬─────┘ └────┬─────┘
      └──────┬────┴───────────┘            │
             ▼                             │
      ┌──────────────┐                     │
      │ notion.Client│◄────────────────────┘
      │ raw net/http │   (errors flow up wrapped)
      └──────────────┘
                  │
┌─────────────────▼───────────────────────────┐
│  pkg/output                                 │  JSON envelope / text / table / CSV.
│  {success, data, metadata}                  │
└─────────────────────────────────────────────┘
```

## Key design decisions

- **No Notion SDK.** Raw `net/http` keeps deps minimal (`cobra` + stdlib) and makes wire-level debugging straightforward.
- **Envelope output.** Every JSON response is `{success, data, metadata}` so agents can parse without per-command schemas.
- **In-memory TTL cache.** Per-resource TTLs (blocks 30s, pages 1m, users 1h, databases 10m). No persistent cache by default; workspace sync writes to disk.
- **Exponential backoff with jitter** for `408/429/5xx`. Retry policy lives in `internal/retry/`.
- **Resolver layer.** All ID-accepting commands route input through `internal/resolver.ExtractID()` so users can paste URLs, IDs, or cached names interchangeably.
- **Structured errors.** `internal/errors.NotionCLIError` carries a code + suggestion; surfaced verbatim in JSON output and used to drive non-zero exit codes.

## Distribution

- Go binary built per-platform via `make release` (`darwin/{arm64,amd64}`, `linux/{amd64,arm64}`, `windows/amd64`).
- npm wrapper (`package.json`) lists each platform binary as an `optionalDependencies` entry; `install.js` selects the right one at install time.
- Release secrets (`NOTION_OAUTH_CLIENT_ID`, `NOTION_OAUTH_SECRET`) are baked in via `-ldflags -X` for OAuth flows.

## Configuration

- `NOTION_TOKEN` (required) — env var, also readable from a config file via `internal/config/`.
- Config file resolution and `config get/set/path/list` commands live in `internal/cli/commands/config.go`.

## Testing

- Pure Go test suite (`go test ./...`); race detector enabled in CI.
- Live API verification is opt-in via `NOTION_TOKEN` and the perception probes in `.gg/eyes/`.
