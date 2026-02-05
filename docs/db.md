`notion-cli db`
===============

Create a database with an initial data source (table)

* [`notion-cli db create PAGE_ID`](#notion-cli-db-create-page_id)
* [`notion-cli db query DATABASE_ID`](#notion-cli-db-query-database_id)
* [`notion-cli db retrieve DATABASE_ID`](#notion-cli-db-retrieve-database_id)
* [`notion-cli db schema DATA_SOURCE_ID`](#notion-cli-db-schema-data_source_id)
* [`notion-cli db update DATABASE_ID`](#notion-cli-db-update-database_id)

## `notion-cli db create PAGE_ID`

Create a database with an initial data source (table)

```
USAGE
  $ notion-cli db create PAGE_ID -t <value> [-r] [--columns <value> | -x] [--sort <value>] [--filter <value>]
    [--csv | --no-truncate] [--no-header] [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache] [-v]
    [--minimal]

ARGUMENTS
  PAGE_ID  Parent page ID or URL where the database will be created

FLAGS
  -j, --json               Output as JSON (recommended for automation)
  -r, --raw                output raw json
  -t, --title=<value>      (required) Title for the database (and initial data source)
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
  Create a database with an initial data source (table)

ALIASES
  $ notion-cli db c

EXAMPLES
  Create a database with an initial data source

    $ notion-cli db create PAGE_ID -t 'My Database'

  Create a database using page URL

    $ notion-cli db create https://notion.so/PAGE_ID -t 'My Database'

  Create a database with an initial data source and output raw json

    $ notion-cli db create PAGE_ID -t 'My Database' -r
```



## `notion-cli db query DATABASE_ID`

Query a database

```
USAGE
  $ notion-cli db query DATABASE_ID [--page-size <value>] [-A] [--sort-property <value>] [--sort-direction
    asc|desc] [-r] [--columns <value> | -x] [--sort <value>] [-f <value> | -s <value> | -F <value> |  | ] [--csv |
    --no-truncate] [--no-header] [-j] [--retry] [--timeout <value>] [--no-cache] [-v] [--minimal] [-m | -c | -P]
    [--select <value>]

ARGUMENTS
  DATABASE_ID  Database or data source ID or URL (required for automation)

FLAGS
  -A, --page-all                 Get all pages (bypass pagination)
  -F, --file-filter=<value>      Load filter from JSON file
  -P, --pretty                   Output as pretty table with borders
  -c, --compact-json             Output as compact JSON (single-line, ideal for piping)
  -f, --filter=<value>           Filter as JSON object (Notion filter API format)
  -j, --json                     Output as JSON (recommended for automation)
  -m, --markdown                 Output as markdown table (GitHub-flavored)
  -r, --raw                      Output raw JSON (recommended for AI assistants - returns all page data)
  -s, --search=<value>           Simple text search (searches across title and common text properties)
  -v, --verbose                  [env: NOTION_CLI_VERBOSE] Enable verbose logging to stderr (retry events, cache stats)
                                 - never pollutes stdout
  -x, --extended                 Show extra columns
      --columns=<value>          Only show provided columns (comma-separated)
      --csv                      Output in CSV format
      --minimal                  Strip unnecessary metadata (created_by, last_edited_by, object fields, request_id,
                                 etc.) - reduces response size by ~40%
      --no-cache                 Bypass cache and force fresh API calls
      --no-header                Hide table header from output
      --no-truncate              Do not truncate output to fit screen
      --page-size=<value>        [default: 100] Items per page (1-100, default: 100 for automation)
      --retry                    Auto-retry on rate limit (respects Retry-After header)
      --select=<value>           Select specific properties to return (comma-separated). Reduces token usage by 60-80%.
      --sort=<value>             Property to sort by (prepend with - for descending)
      --sort-direction=<option>  [default: asc] The direction to sort results
                                 <options: asc|desc>
      --sort-property=<value>    The property to sort results by
      --timeout=<value>          [default: 30000] Request timeout in milliseconds

DESCRIPTION
  Query a database

ALIASES
  $ notion-cli db q

EXAMPLES
  Query a database with full data (recommended for AI assistants)

    $ notion-cli db query DATABASE_ID --raw

  Query all records as JSON

    $ notion-cli db query DATABASE_ID --json

  Filter with JSON object (recommended for AI agents)

    $ notion-cli db query DATABASE_ID --filter '{"property": "Status", "select": {"equals": "Done"}}' --json

  Simple text search across properties

    $ notion-cli db query DATABASE_ID --search "urgent" --json

  Load complex filter from file

    $ notion-cli db query DATABASE_ID --file-filter ./filter.json --json

  Query with AND filter

    $ notion-cli db query DATABASE_ID --filter '{"and": [{"property": "Status", "select": {"equals": "Done"}}, \
      {"property": "Priority", "number": {"greater_than": 5}}]}' --json

  Query using database URL

    $ notion-cli db query https://notion.so/DATABASE_ID --json

  Query with sorting

    $ notion-cli db query DATABASE_ID --sort-property Name --sort-direction desc

  Query with pagination

    $ notion-cli db query DATABASE_ID --page-size 50

  Get all pages (bypass pagination)

    $ notion-cli db query DATABASE_ID --page-all

  Output as CSV

    $ notion-cli db query DATABASE_ID --csv

  Output as markdown table

    $ notion-cli db query DATABASE_ID --markdown

  Output as compact JSON

    $ notion-cli db query DATABASE_ID --compact-json

  Output as pretty table

    $ notion-cli db query DATABASE_ID --pretty

  Select specific properties (60-80% token reduction)

    $ notion-cli db query DATABASE_ID --select "title,status,priority" --json
```



## `notion-cli db retrieve DATABASE_ID`

Retrieve a data source (table) schema and properties

```
USAGE
  $ notion-cli db retrieve DATABASE_ID [-r] [--columns <value> | -x] [--sort <value>] [--filter <value>] [--csv |
    --no-truncate] [--no-header] [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache] [-v] [--minimal]
    [-m | -c | -P]

ARGUMENTS
  DATABASE_ID  Data source ID or URL (the ID of the table whose schema you want to retrieve)

FLAGS
  -P, --pretty             Output as pretty table with borders
  -c, --compact-json       Output as compact JSON (single-line, ideal for piping)
  -j, --json               Output as JSON (recommended for automation)
  -m, --markdown           Output as markdown table (GitHub-flavored)
  -r, --raw                output raw json (recommended for AI assistants - returns full schema)
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
  Retrieve a data source (table) schema and properties

ALIASES
  $ notion-cli db r
  $ notion-cli ds retrieve
  $ notion-cli ds r

EXAMPLES
  Retrieve a data source with full schema (recommended for AI assistants)

    $ notion-cli db retrieve DATA_SOURCE_ID -r

  Retrieve a data source schema via data_source_id

    $ notion-cli db retrieve DATA_SOURCE_ID

  Retrieve a data source via URL

    $ notion-cli db retrieve https://notion.so/DATABASE_ID

  Retrieve a data source and output as markdown table

    $ notion-cli db retrieve DATA_SOURCE_ID --markdown

  Retrieve a data source and output as compact JSON

    $ notion-cli db retrieve DATA_SOURCE_ID --compact-json
```



## `notion-cli db schema DATA_SOURCE_ID`

Extract clean, AI-parseable schema from a Notion data source (table). This command is optimized for AI agents and automation - it returns property names, types, options (for select/multi-select), and configuration in an easy-to-parse format.

```
USAGE
  $ notion-cli db schema DATA_SOURCE_ID [-o json|yaml|table] [-p <value>] [-m] [-j] [-e]

ARGUMENTS
  DATA_SOURCE_ID  Data source ID or URL (the table whose schema you want to extract)

FLAGS
  -e, --with-examples       Include property payload examples for create/update operations
  -j, --json                Output as JSON (shorthand for --output json)
  -m, --markdown            Output as markdown documentation
  -o, --output=<option>     [default: table] Output format
                            <options: json|yaml|table>
  -p, --properties=<value>  Comma-separated list of properties to include (default: all)

DESCRIPTION
  Extract clean, AI-parseable schema from a Notion data source (table). This command is optimized for AI agents and
  automation - it returns property names, types, options (for select/multi-select), and configuration in an
  easy-to-parse format.

ALIASES
  $ notion-cli db s
  $ notion-cli ds schema
  $ notion-cli ds s

EXAMPLES
  Get full schema in JSON format (recommended for AI agents)

    $ notion-cli db schema abc123def456 --output json

  Get schema with property payload examples (recommended for AI agents)

    $ notion-cli db schema abc123def456 --with-examples --json

  Get schema using database URL

    $ notion-cli db schema https://notion.so/DATABASE_ID --output json

  Get schema as formatted table

    $ notion-cli db schema abc123def456

  Get schema with examples in human-readable format

    $ notion-cli db schema abc123def456 --with-examples

  Get schema in YAML format

    $ notion-cli db schema abc123def456 --output yaml

  Get only specific properties

    $ notion-cli db schema abc123def456 --properties Name,Status,Tags --output json

  Get schema as markdown documentation

    $ notion-cli db schema abc123def456 --markdown

  Parse schema with jq (extract property names)

    $ notion-cli db schema abc123def456 --output json | jq '.data.properties[].name'

  Find all select/multi-select properties and their options

    $ notion-cli db schema abc123def456 --output json | jq '.data.properties[] | select(.options) | {name, options}'
```



## `notion-cli db update DATABASE_ID`

Update a data source (table) title and properties

```
USAGE
  $ notion-cli db update DATABASE_ID -t <value> [-r] [--columns <value> | -x] [--sort <value>] [--filter
    <value>] [--csv | --no-truncate] [--no-header] [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache]
    [-v] [--minimal]

ARGUMENTS
  DATABASE_ID  Data source ID or URL (the ID of the table you want to update)

FLAGS
  -j, --json               Output as JSON (recommended for automation)
  -r, --raw                output raw json
  -t, --title=<value>      (required) New database title
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
  Update a data source (table) title and properties

ALIASES
  $ notion-cli db u
  $ notion-cli ds update
  $ notion-cli ds u

EXAMPLES
  Update a data source with a specific data_source_id and title

    $ notion-cli db update DATA_SOURCE_ID -t 'My Data Source'

  Update a data source via URL

    $ notion-cli db update https://notion.so/DATABASE_ID -t 'My Data Source'

  Update a data source with a specific data_source_id and output raw json

    $ notion-cli db update DATA_SOURCE_ID -t 'My Table' -r
```


