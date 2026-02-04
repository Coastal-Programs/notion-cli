`notion-cli init`
=================

Interactive first-time setup wizard for Notion CLI

* [`notion-cli init`](#notion-cli-init)

## `notion-cli init`

Interactive first-time setup wizard for Notion CLI

```
USAGE
  $ notion-cli init [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache] [-v] [--minimal]

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
  Interactive first-time setup wizard for Notion CLI

EXAMPLES
  Run interactive setup wizard

    $ notion-cli init

  Run setup with automated JSON output

    $ notion-cli init --json
```


