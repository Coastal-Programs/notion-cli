`notion-cli doctor`
===================

Run health checks and diagnostics for Notion CLI

* [`notion-cli doctor`](#notion-cli-doctor)

## `notion-cli doctor`

Run health checks and diagnostics for Notion CLI

```
USAGE
  $ notion-cli doctor [-j]

FLAGS
  -j, --json  Output as JSON

DESCRIPTION
  Run health checks and diagnostics for Notion CLI

  Performs the following checks:
  - Go runtime version
  - API token configuration (env var or config file)
  - Token format validation (secret_ or ntn_ prefix)
  - Network connectivity to api.notion.com
  - API connection test with latency measurement
  - Data directory status
  - Workspace cache freshness

ALIASES
  $ notion-cli diagnose
  $ notion-cli healthcheck

EXAMPLES
  Run all health checks

    $ notion-cli doctor

  Run health checks with JSON output

    $ notion-cli doctor --json
```

