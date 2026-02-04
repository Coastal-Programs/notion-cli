`notion-cli whoami`
===================

Verify API connectivity and show workspace context

* [`notion-cli whoami`](#notion-cli-whoami)

## `notion-cli whoami`

Verify API connectivity and show workspace context

```
USAGE
  $ notion-cli whoami [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache] [-v] [--minimal]

FLAGS
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
  Verify API connectivity and show workspace context

ALIASES
  $ notion-cli test
  $ notion-cli health
  $ notion-cli connectivity

EXAMPLES
  Check connection and show bot info

    $ notion-cli whoami

  Check connection and output as JSON

    $ notion-cli whoami --json

  Bypass cache for fresh connectivity test

    $ notion-cli whoami --no-cache
```


