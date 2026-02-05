`notion-cli list`
=================

List all cached databases from your workspace

* [`notion-cli list`](#notion-cli-list)

## `notion-cli list`

List all cached databases from your workspace

```
USAGE
  $ notion-cli list [--columns <value> | -x] [--sort <value>] [--filter <value>] [--csv | --no-truncate]
    [--no-header] [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache] [-v] [--minimal] [-m | -c | -P]

FLAGS
  -P, --pretty             Output as pretty table with borders
  -c, --compact-json       Output as compact JSON (single-line, ideal for piping)
  -j, --json               Output as JSON (recommended for automation)
  -m, --markdown           Output as markdown table (GitHub-flavored)
  -v, --verbose            [env: NOTION_CLI_VERBOSE] Enable verbose logging to stderr (retry events, cache stats) -
                           never pollutes stdout
  -x, --extended           Show extra columns
      --columns=<value>    Only show provided columns (comma-separated)
      --csv                Output in CSV format
      --filter=<value>     Filter property by substring match
      --minimal            Strip unnecessary metadata (created_by, last_edited_by, object fields, request_id, etc.) -
                           reduces response size by ~40%
      --no-cache           Bypass cache and force fresh API calls
      --no-header          Hide table header from output
      --no-truncate        Do not truncate output to fit screen
      --page-size=<value>  [default: 100] Items per page (1-100, default: 100 for automation)
      --retry              Auto-retry on rate limit (respects Retry-After header)
      --sort=<value>       Property to sort by (prepend with - for descending)
      --timeout=<value>    [default: 30000] Request timeout in milliseconds

DESCRIPTION
  List all cached databases from your workspace

ALIASES
  $ notion-cli db list
  $ notion-cli ls

EXAMPLES
  List all cached databases

    $ notion-cli list

  List databases in markdown format

    $ notion-cli list --markdown

  List databases in JSON format

    $ notion-cli list --json

  List databases in pretty table format

    $ notion-cli list --pretty
```


