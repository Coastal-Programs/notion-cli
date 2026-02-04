`notion-cli search`
===================

Search by title

* [`notion-cli search`](#notion-cli-search)

## `notion-cli search`

Search by title

```
USAGE
  $ notion-cli search [-q <value>] [-d asc|desc] [-p data_source|page] [-c <value>] [-s <value>] [--database
    <value>] [--created-after <value>] [--created-before <value>] [--edited-after <value>] [--edited-before <value>]
    [--limit <value>] [-r] [--columns <value> | -x] [--sort <value>] [--filter <value>] [--csv | --no-truncate]
    [--no-header] [-m | -c | -P] [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache] [-v] [--minimal]

FLAGS
  -P, --pretty                   Output as pretty table with borders
  -c, --compact-json             Output as compact JSON (single-line, ideal for piping)
  -c, --start_cursor=<value>
  -d, --sort_direction=<option>  [default: desc] The direction to sort results. The only supported timestamp value is
                                 "last_edited_time"
                                 <options: asc|desc>
  -j, --json                     Output as JSON (recommended for automation)
  -m, --markdown                 Output as markdown table (GitHub-flavored)
  -p, --property=<option>        <options: data_source|page>
  -q, --query=<value>            The text that the API compares page and database titles against
  -r, --raw                      output raw json (recommended for AI assistants - returns all search results)
  -s, --page_size=<value>        [default: 5] The number of results to return. The default is 5, with a minimum of 1 and
                                 a maximum of 100.
  -v, --verbose                  [env: NOTION_CLI_VERBOSE] Enable verbose logging to stderr (retry events, cache stats)
                                 - never pollutes stdout
  -x, --extended                 Show extra columns
      --columns=<value>          Only show provided columns (comma-separated)
      --created-after=<value>    Filter results created after this date (ISO 8601 format: YYYY-MM-DD)
      --created-before=<value>   Filter results created before this date (ISO 8601 format: YYYY-MM-DD)
      --csv                      Output in CSV format
      --database=<value>         Limit search to pages within a specific database (data source ID)
      --edited-after=<value>     Filter results edited after this date (ISO 8601 format: YYYY-MM-DD)
      --edited-before=<value>    Filter results edited before this date (ISO 8601 format: YYYY-MM-DD)
      --filter=<value>           Filter property by substring match
      --limit=<value>            Maximum number of results to return (applied after filters)
      --minimal                  Strip unnecessary metadata (created_by, last_edited_by, object fields, request_id,
                                 etc.) - reduces response size by ~40%
      --no-cache                 Bypass cache and force fresh API calls
      --no-header                Hide table header from output
      --no-truncate              Do not truncate output to fit screen
      --page-size=<value>        [default: 100] Items per page (1-100, default: 100 for automation)
      --retry                    Auto-retry on rate limit (respects Retry-After header)
      --sort=<value>             Property to sort by (prepend with - for descending)
      --timeout=<value>          [default: 30000] Request timeout in milliseconds

DESCRIPTION
  Search by title

EXAMPLES
  Search with full data (recommended for AI assistants)

    $ notion-cli search -q 'My Page' -r

  Search by title

    $ notion-cli search -q 'My Page'

  Search only within a specific database

    $ notion-cli search -q 'meeting' --database DB_ID

  Search with created date filter

    $ notion-cli search -q 'report' --created-after 2025-10-01

  Search with edited date filter

    $ notion-cli search -q 'project' --edited-after 2025-10-20

  Limit number of results

    $ notion-cli search -q 'task' --limit 20

  Combined filters

    $ notion-cli search -q 'project' -d DB_ID --edited-after 2025-10-20 --limit 10

  Search by title and output csv

    $ notion-cli search -q 'My Page' --csv

  Search by title and output raw json

    $ notion-cli search -q 'My Page' -r

  Search by title and output markdown table

    $ notion-cli search -q 'My Page' --markdown

  Search by title and output compact JSON

    $ notion-cli search -q 'My Page' --compact-json

  Search by title and output pretty table

    $ notion-cli search -q 'My Page' --pretty

  Search by title and output table with specific columns

    $ notion-cli search -q 'My Page' --columns=title,object

  Search by title and output table with specific columns and sort direction

    $ notion-cli search -q 'My Page' --columns=title,object -d asc

  Search by title and output table with specific columns and sort direction and page size

    $ notion-cli search -q 'My Page' -columns=title,object -d asc -s 10

  Search by title and output table with specific columns and sort direction and page size and start cursor

    $ notion-cli search -q 'My Page' --columns=title,object -d asc -s 10 -c START_CURSOR_ID

  Search by title and output table with specific columns and sort direction and page size and start cursor and
  property

    $ notion-cli search -q 'My Page' --columns=title,object -d asc -s 10 -c START_CURSOR_ID -p page

  Search and output JSON for automation

    $ notion-cli search -q 'My Page' --json
```


