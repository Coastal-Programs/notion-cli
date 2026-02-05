`notion-cli sync`
=================

Sync workspace databases to local cache for fast lookups

* [`notion-cli sync`](#notion-cli-sync)

## `notion-cli sync`

Sync workspace databases to local cache for fast lookups

```
USAGE
  $ notion-cli sync [-f] [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache] [-v]
    [--minimal]

FLAGS
  -f, --force              Force resync even if cache is fresh
  -j, --json               Output as JSON (recommended for automation)
  -v, --verbose            [env: NOTION_CLI_VERBOSE] Enable verbose logging to stderr (retry events, cache stats) -
                           never pollutes stdout
      --minimal            Strip unnecessary metadata (created_by, last_edited_by, object fields, request_id, etc.) -
                           reduces response size by ~40%
      --no-cache           Bypass cache and force fresh API calls
      --page-size=<value>  [default: 100] Items per page (1-100, default: 100 for automation)
      --retry              Auto-retry on rate limit (respects Retry-After header)
      --timeout=<value>    [default: 30000] Request timeout in milliseconds

DESCRIPTION
  Sync workspace databases to local cache for fast lookups

ALIASES
  $ notion-cli db sync

EXAMPLES
  Sync all workspace databases

    $ notion-cli sync

  Force resync even if cache exists

    $ notion-cli sync --force

  Sync and output as JSON

    $ notion-cli sync --json
```


