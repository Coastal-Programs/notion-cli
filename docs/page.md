`notion-cli page`
=================

Create a page

* [`notion-cli page create`](#notion-cli-page-create)
* [`notion-cli page retrieve PAGE_ID`](#notion-cli-page-retrieve-page_id)
* [`notion-cli page retrieve property_item PAGE_ID PROPERTY_ID`](#notion-cli-page-retrieve-property_item-page_id-property_id)
* [`notion-cli page update PAGE_ID`](#notion-cli-page-update-page_id)

## `notion-cli page create`

Create a page

```
USAGE
  $ notion-cli page create [-p <value>] [-d <value>] [-f <value>] [-t <value>] [--properties <value>] [-S] [-r]
    [--columns <value> | -x] [--sort <value>] [--filter <value>] [--csv | --no-truncate] [--no-header] [-j] [--page-size
    <value>] [--retry] [--timeout <value>] [--no-cache] [-v] [--minimal]

FLAGS
  -S, --simple-properties              Use simplified property format (flat key-value pairs, recommended for AI agents)
  -d, --parent_data_source_id=<value>  Parent data source ID or URL (to create a page in a table)
  -f, --file_path=<value>              Path to a source markdown file
  -j, --json                           Output as JSON (recommended for automation)
  -p, --parent_page_id=<value>         Parent page ID or URL (to create a sub-page)
  -r, --raw                            output raw json
  -t, --title_property=<value>         [default: Name] Name of the title property (defaults to "Name" if not specified)
  -v, --verbose                        [env: NOTION_CLI_VERBOSE] Enable verbose logging to stderr (retry events, cache
                                       stats) - never pollutes stdout
  -x, --extended                       Show extra columns
      --columns=<value>                Only show provided columns (comma-separated)
      --csv                            Output in CSV format
      --filter=<value>                 Filter property by substring match
      --minimal                        Strip unnecessary metadata (created_by, last_edited_by, object fields,
                                       request_id, etc.) - reduces response size by ~40%
      --no-cache                       Bypass cache and force fresh API calls
      --no-header                      Hide table header from output
      --no-truncate                    Do not truncate output to fit screen
      --page-size=<value>              [default: 100] Items per page (1-100, default: 100 for automation)
      --properties=<value>             Page properties as JSON string
      --retry                          Auto-retry on rate limit (respects Retry-After header)
      --sort=<value>                   Property to sort by (prepend with - for descending)
      --timeout=<value>                [default: 30000] Request timeout in milliseconds

DESCRIPTION
  Create a page

ALIASES
  $ notion-cli page c

EXAMPLES
  Create a page via interactive mode

    $ notion-cli page create

  Create a page with a specific parent_page_id

    $ notion-cli page create -p PARENT_PAGE_ID

  Create a page with a parent page URL

    $ notion-cli page create -p https://notion.so/PARENT_PAGE_ID

  Create a page with a specific parent_db_id

    $ notion-cli page create -d PARENT_DB_ID

  Create a page with simple properties (recommended for AI agents)

    $ notion-cli page create -d DATA_SOURCE_ID -S --properties '{"Name": "My Task", "Status": "In Progress", "Due \
      Date": "2025-12-31"}'

  Create a page with simple properties using relative dates

    $ notion-cli page create -d DATA_SOURCE_ID -S --properties '{"Name": "Review", "Due Date": "tomorrow", \
      "Priority": "High"}'

  Create a page with simple properties and multi-select

    $ notion-cli page create -d DATA_SOURCE_ID -S --properties '{"Name": "Bug Fix", "Tags": ["urgent", "bug"], \
      "Status": "Done"}'

  Create a page with a specific source markdown file and parent_page_id

    $ notion-cli page create -f ./path/to/source.md -p PARENT_PAGE_ID

  Create a page with a specific source markdown file and parent_db_id

    $ notion-cli page create -f ./path/to/source.md -d PARENT_DB_ID

  Create a page with a specific source markdown file and output raw json with parent_page_id

    $ notion-cli page create -f ./path/to/source.md -p PARENT_PAGE_ID -r

  Create a page and output JSON for automation

    $ notion-cli page create -p PARENT_PAGE_ID --json
```



## `notion-cli page retrieve PAGE_ID`

Retrieve a page

```
USAGE
  $ notion-cli page retrieve PAGE_ID [--map | -r | [-m | -c | -P]] [--max-depth <value> -R] [--columns <value> |
    -x] [--sort <value>] [--filter <value>] [--csv | --no-truncate] [--no-header] [-j] [--page-size <value>] [--retry]
    [--timeout <value>] [--no-cache] [-v] [--minimal]

ARGUMENTS
  PAGE_ID  Page ID or full Notion URL (e.g., https://notion.so/...)

FLAGS
  -P, --pretty             Output as pretty table with borders
  -R, --recursive          recursively fetch all blocks and nested pages (reduces API calls)
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
      --map                fast structure discovery (returns minimal info: titles, types, IDs)
      --max-depth=<value>  [default: 3] maximum recursion depth for --recursive (default: 3)
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
  Retrieve a page

ALIASES
  $ notion-cli page r

EXAMPLES
  Retrieve a page with full data (recommended for AI assistants)

    $ notion-cli page retrieve PAGE_ID -r

  Fast structure overview (90% faster than full fetch)

    $ notion-cli page retrieve PAGE_ID --map

  Fast structure overview with compact JSON

    $ notion-cli page retrieve PAGE_ID --map --compact-json

  Retrieve entire page tree with all nested content (35% token reduction)

    $ notion-cli page retrieve PAGE_ID --recursive --compact-json

  Retrieve page tree with custom depth limit

    $ notion-cli page retrieve PAGE_ID -R --max-depth 5 --json

  Retrieve a page and output table

    $ notion-cli page retrieve PAGE_ID

  Retrieve a page via URL

    $ notion-cli page retrieve https://notion.so/PAGE_ID

  Retrieve a page and output raw json

    $ notion-cli page retrieve PAGE_ID -r

  Retrieve a page and output markdown

    $ notion-cli page retrieve PAGE_ID -m

  Retrieve a page metadata and output as markdown table

    $ notion-cli page retrieve PAGE_ID --markdown

  Retrieve a page metadata and output as compact JSON

    $ notion-cli page retrieve PAGE_ID --compact-json

  Retrieve a page and output JSON for automation

    $ notion-cli page retrieve PAGE_ID --json
```



## `notion-cli page retrieve property_item PAGE_ID PROPERTY_ID`

Retrieve a page property item

```
USAGE
  $ notion-cli page retrieve property_item PAGE_ID PROPERTY_ID [-r] [-j] [--page-size <value>] [--retry] [--timeout <value>]
    [--no-cache] [-v] [--minimal]

FLAGS
  -j, --json               Output as JSON (recommended for automation)
  -r, --raw                output raw json
  -v, --verbose            [env: NOTION_CLI_VERBOSE] Enable verbose logging to stderr (retry events, cache stats) -
                           never pollutes stdout
      --minimal            Strip unnecessary metadata (created_by, last_edited_by, object fields, request_id, etc.) -
                           reduces response size by ~40%
      --no-cache           Bypass cache and force fresh API calls
      --page-size=<value>  [default: 100] Items per page (1-100, default: 100 for automation)
      --retry              Auto-retry on rate limit (respects Retry-After header)
      --timeout=<value>    [default: 30000] Request timeout in milliseconds

DESCRIPTION
  Retrieve a page property item

ALIASES
  $ notion-cli page r pi

EXAMPLES
  Retrieve a page property item

    $ notion-cli page retrieve:property_item PAGE_ID PROPERTY_ID

  Retrieve a page property item and output raw json

    $ notion-cli page retrieve:property_item PAGE_ID PROPERTY_ID -r

  Retrieve a page property item and output JSON for automation

    $ notion-cli page retrieve:property_item PAGE_ID PROPERTY_ID --json
```



## `notion-cli page update PAGE_ID`

Update a page

```
USAGE
  $ notion-cli page update PAGE_ID [-a] [-u] [--properties <value>] [-S] [-r] [--columns <value> | -x] [--sort
    <value>] [--filter <value>] [--csv | --no-truncate] [--no-header] [-j] [--page-size <value>] [--retry] [--timeout
    <value>] [--no-cache] [-v] [--minimal]

ARGUMENTS
  PAGE_ID  Page ID or full Notion URL (e.g., https://notion.so/...)

FLAGS
  -S, --simple-properties   Use simplified property format (flat key-value pairs, recommended for AI agents)
  -a, --archived            Archive the page
  -j, --json                Output as JSON (recommended for automation)
  -r, --raw                 output raw json
  -u, --unarchive           Unarchive the page
  -v, --verbose             [env: NOTION_CLI_VERBOSE] Enable verbose logging to stderr (retry events, cache stats) -
                            never pollutes stdout
  -x, --extended            Show extra columns
      --columns=<value>     Only show provided columns (comma-separated)
      --csv                 Output in CSV format
      --filter=<value>      Filter property by substring match
      --minimal             Strip unnecessary metadata (created_by, last_edited_by, object fields, request_id, etc.) -
                            reduces response size by ~40%
      --no-cache            Bypass cache and force fresh API calls
      --no-header           Hide table header from output
      --no-truncate         Do not truncate output to fit screen
      --page-size=<value>   [default: 100] Items per page (1-100, default: 100 for automation)
      --properties=<value>  Page properties to update as JSON string
      --retry               Auto-retry on rate limit (respects Retry-After header)
      --sort=<value>        Property to sort by (prepend with - for descending)
      --timeout=<value>     [default: 30000] Request timeout in milliseconds

DESCRIPTION
  Update a page

ALIASES
  $ notion-cli page u

EXAMPLES
  Update a page and output table

    $ notion-cli page update PAGE_ID

  Update a page via URL

    $ notion-cli page update https://notion.so/PAGE_ID -a

  Update page properties with simple format (recommended for AI agents)

    $ notion-cli page update PAGE_ID -S --properties '{"Status": "Done", "Priority": "High"}'

  Update page properties with relative date

    $ notion-cli page update PAGE_ID -S --properties '{"Due Date": "tomorrow", "Status": "In Progress"}'

  Update page with multi-select tags

    $ notion-cli page update PAGE_ID -S --properties '{"Tags": ["urgent", "bug"], "Status": "Done"}'

  Update a page and output raw json

    $ notion-cli page update PAGE_ID -r

  Update a page and archive

    $ notion-cli page update PAGE_ID -a

  Update a page and unarchive

    $ notion-cli page update PAGE_ID -u

  Update a page and archive and output raw json

    $ notion-cli page update PAGE_ID -a -r

  Update a page and unarchive and output raw json

    $ notion-cli page update PAGE_ID -u -r

  Update a page and output JSON for automation

    $ notion-cli page update PAGE_ID -a --json
```


