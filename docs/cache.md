`notion-cli cache`
==================

Show cache statistics and configuration

* [`notion-cli cache info`](#notion-cli-cache-info)

## `notion-cli cache info`

Show cache statistics and configuration

```
USAGE
  $ notion-cli cache info [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache] [-v] [--minimal]

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
  Show cache statistics and configuration

ALIASES
  $ notion-cli cache stats
  $ notion-cli cache status

EXAMPLES
  Show cache info in JSON format

    $ notion-cli cache:info --json

  Show cache statistics

    $ notion-cli cache:info
```


