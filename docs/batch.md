`notion-cli batch`
==================

Batch retrieve multiple pages, blocks, or data sources

* [`notion-cli batch retrieve [IDS]`](#notion-cli-batch-retrieve-ids)

## `notion-cli batch retrieve [IDS]`

Batch retrieve multiple pages, blocks, or data sources

```
USAGE
  $ notion-cli batch retrieve [IDS] [--ids <value>] [--type page|block|database] [-r] [--columns <value> | -x]
    [--sort <value>] [--filter <value>] [--csv | --no-truncate] [--no-header] [-m | -c | -P] [-j] [--page-size <value>]
    [--retry] [--timeout <value>] [--no-cache] [-v] [--minimal]

ARGUMENTS
  [IDS]  Comma-separated list of IDs to retrieve (or use --ids flag or stdin)

FLAGS
  -P, --pretty             Output as pretty table with borders
  -c, --compact-json       Output as compact JSON (single-line, ideal for piping)
  -j, --json               Output as JSON (recommended for automation)
  -m, --markdown           Output as markdown table (GitHub-flavored)
  -r, --raw                output raw json (recommended for AI assistants - returns all fields)
  -v, --verbose            [env: NOTION_CLI_VERBOSE] Enable verbose logging to stderr (retry events, cache stats) -
                           never pollutes stdout
  -x, --extended           Show extra columns
      --columns=<value>    Only show provided columns (comma-separated)
      --csv                Output in CSV format
      --filter=<value>     Filter property by substring match
      --ids=<value>        Comma-separated list of IDs to retrieve
      --minimal            Strip unnecessary metadata (created_by, last_edited_by, object fields, request_id, etc.) -
                           reduces response size by ~40%
      --no-cache           Bypass cache and force fresh API calls
      --no-header          Hide table header from output
      --no-truncate        Do not truncate output to fit screen
      --page-size=<value>  [default: 100] Items per page (1-100, default: 100 for automation)
      --retry              Auto-retry on rate limit (respects Retry-After header)
      --sort=<value>       Property to sort by (prepend with - for descending)
      --timeout=<value>    [default: 30000] Request timeout in milliseconds
      --type=<option>      [default: page] Resource type to retrieve (page, block, database)
                           <options: page|block|database>

DESCRIPTION
  Batch retrieve multiple pages, blocks, or data sources

ALIASES
  $ notion-cli batch r

EXAMPLES
  Retrieve multiple pages via --ids flag

    $ notion-cli batch retrieve --ids PAGE_ID_1,PAGE_ID_2,PAGE_ID_3 --compact-json

  Retrieve multiple pages from stdin (one ID per line)

    $ cat page_ids.txt | notion-cli batch retrieve --compact-json

  Retrieve multiple blocks

    $ notion-cli batch retrieve --ids BLOCK_ID_1,BLOCK_ID_2 --type block --json

  Retrieve multiple data sources

    $ notion-cli batch retrieve --ids DS_ID_1,DS_ID_2 --type database --json

  Retrieve with raw output

    $ notion-cli batch retrieve --ids ID1,ID2,ID3 -r
```


