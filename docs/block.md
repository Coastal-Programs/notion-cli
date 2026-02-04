`notion-cli block`
==================

Append block children

* [`notion-cli block append`](#notion-cli-block-append)
* [`notion-cli block delete BLOCK_ID`](#notion-cli-block-delete-block_id)
* [`notion-cli block retrieve BLOCK_ID`](#notion-cli-block-retrieve-block_id)
* [`notion-cli block retrieve children BLOCK_ID`](#notion-cli-block-retrieve-children-block_id)
* [`notion-cli block update BLOCK_ID`](#notion-cli-block-update-block_id)

## `notion-cli block append`

Append block children

```
USAGE
  $ notion-cli block append -b <value> [-c <value>] [--text <value>] [--heading-1 <value>] [--heading-2 <value>]
    [--heading-3 <value>] [--bullet <value>] [--numbered <value>] [--todo <value>] [--toggle <value>] [--code <value>]
    [--language <value>] [--quote <value>] [--callout <value>] [-a <value>] [-r] [--columns <value> | -x] [--sort
    <value>] [--filter <value>] [--csv | --no-truncate] [--no-header] [-j] [--page-size <value>] [--retry] [--timeout
    <value>] [--no-cache] [-v] [--minimal]

FLAGS
  -a, --after=<value>      Block ID or URL to append after (optional)
  -b, --block_id=<value>   (required) Parent block ID or URL
  -c, --children=<value>   Block children (JSON array) - for complex cases
  -j, --json               Output as JSON (recommended for automation)
  -r, --raw                output raw json
  -v, --verbose            [env: NOTION_CLI_VERBOSE] Enable verbose logging to stderr (retry events, cache stats) -
                           never pollutes stdout
  -x, --extended           Show extra columns
      --bullet=<value>     Bulleted list item text
      --callout=<value>    Callout block text
      --code=<value>       Code block content
      --columns=<value>    Only show provided columns (comma-separated)
      --csv                Output in CSV format
      --filter=<value>     Filter property by substring match
      --heading-1=<value>  H1 heading text
      --heading-2=<value>  H2 heading text
      --heading-3=<value>  H3 heading text
      --language=<value>   [default: plain text] Code block language (used with --code)
      --minimal            Strip unnecessary metadata (created_by, last_edited_by, object fields, request_id, etc.) -
                           reduces response size by ~40%
      --no-cache           Bypass cache and force fresh API calls
      --no-header          Hide table header from output
      --no-truncate        Do not truncate output to fit screen
      --numbered=<value>   Numbered list item text
      --page-size=<value>  [default: 100] Items per page (1-100, default: 100 for automation)
      --quote=<value>      Quote block text
      --retry              Auto-retry on rate limit (respects Retry-After header)
      --sort=<value>       Property to sort by (prepend with - for descending)
      --text=<value>       Paragraph text
      --timeout=<value>    [default: 30000] Request timeout in milliseconds
      --todo=<value>       To-do item text
      --toggle=<value>     Toggle block text

DESCRIPTION
  Append block children

ALIASES
  $ notion-cli block a

EXAMPLES
  Append a simple paragraph

    $ notion-cli block append -b BLOCK_ID --text "Hello world!"

  Append a heading

    $ notion-cli block append -b BLOCK_ID --heading-1 "Chapter Title"

  Append a bullet point

    $ notion-cli block append -b BLOCK_ID --bullet "First item"

  Append a code block

    $ notion-cli block append -b BLOCK_ID --code "console.log('test')" --language javascript

  Append block children with complex JSON (for advanced cases)

    $ notion-cli block append -b BLOCK_ID -c \
      '[{"object":"block","type":"paragraph","paragraph":{"rich_text":[{"type":"text","text":{"content":"Hello \
      world!"}}]}}]'

  Append block children via URL

    $ notion-cli block append -b https://notion.so/BLOCK_ID --text "Hello world!"

  Append block children after a block

    $ notion-cli block append -b BLOCK_ID --text "Hello world!" -a AFTER_BLOCK_ID

  Append block children and output raw json

    $ notion-cli block append -b BLOCK_ID --text "Hello world!" -r

  Append block children and output JSON for automation

    $ notion-cli block append -b BLOCK_ID --text "Hello world!" --json
```



## `notion-cli block delete BLOCK_ID`

Delete a block

```
USAGE
  $ notion-cli block delete BLOCK_ID [-r] [--columns <value> | -x] [--sort <value>] [--filter <value>] [--csv |
    --no-truncate] [--no-header] [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache] [-v] [--minimal]

FLAGS
  -j, --json               Output as JSON (recommended for automation)
  -r, --raw                output raw json
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
  Delete a block

ALIASES
  $ notion-cli block d

EXAMPLES
  Delete a block

    $ notion-cli block delete BLOCK_ID

  Delete a block and output raw json

    $ notion-cli block delete BLOCK_ID -r

  Delete a block and output JSON for automation

    $ notion-cli block delete BLOCK_ID --json
```



## `notion-cli block retrieve BLOCK_ID`

Retrieve a block

```
USAGE
  $ notion-cli block retrieve BLOCK_ID [-r] [--columns <value> | -x] [--sort <value>] [--filter <value>] [--csv |
    --no-truncate] [--no-header] [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache] [-v] [--minimal]

FLAGS
  -j, --json               Output as JSON (recommended for automation)
  -r, --raw                output raw json
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
  Retrieve a block

ALIASES
  $ notion-cli block r

EXAMPLES
  Retrieve a block

    $ notion-cli block retrieve BLOCK_ID

  Retrieve a block and output raw json

    $ notion-cli block retrieve BLOCK_ID -r

  Retrieve a block and output JSON for automation

    $ notion-cli block retrieve BLOCK_ID --json
```



## `notion-cli block retrieve children BLOCK_ID`

Retrieve block children (supports database discovery via --show-databases)

```
USAGE
  $ notion-cli block retrieve children BLOCK_ID [-r] [-d] [--columns <value> | -x] [--sort <value>] [--filter <value>] [--csv
    | --no-truncate] [--no-header] [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache] [-v]
    [--minimal]

ARGUMENTS
  BLOCK_ID  block_id or page_id

FLAGS
  -d, --show-databases     show only child databases with their queryable IDs (data_source_id)
  -j, --json               Output as JSON (recommended for automation)
  -r, --raw                output raw json
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
  Retrieve block children (supports database discovery via --show-databases)

ALIASES
  $ notion-cli block r c

EXAMPLES
  Retrieve block children

    $ notion-cli block retrieve:children BLOCK_ID

  Retrieve block children and output raw json

    $ notion-cli block retrieve:children BLOCK_ID -r

  Retrieve block children and output JSON for automation

    $ notion-cli block retrieve:children BLOCK_ID --json

  Discover databases on a page with queryable IDs

    $ notion-cli block retrieve:children PAGE_ID --show-databases

  Get databases as JSON for automation

    $ notion-cli block retrieve:children PAGE_ID --show-databases --json
```



## `notion-cli block update BLOCK_ID`

Update a block

```
USAGE
  $ notion-cli block update BLOCK_ID [-a] [-c <value>] [--text <value>] [--heading-1 <value>] [--heading-2
    <value>] [--heading-3 <value>] [--bullet <value>] [--numbered <value>] [--todo <value>] [--toggle <value>] [--code
    <value>] [--language <value>] [--quote <value>] [--callout <value>] [--color
    default|gray|brown|orange|yellow|green|blue|purple|pink|red] [-r] [--columns <value> | -x] [--sort <value>]
    [--filter <value>] [--csv | --no-truncate] [--no-header] [-j] [--page-size <value>] [--retry] [--timeout <value>]
    [--no-cache] [-v] [--minimal]

ARGUMENTS
  BLOCK_ID  Block ID or URL

FLAGS
  -a, --archived           Archive the block
  -c, --content=<value>    Updated block content (JSON object with block type properties) - for complex cases
  -j, --json               Output as JSON (recommended for automation)
  -r, --raw                output raw json
  -v, --verbose            [env: NOTION_CLI_VERBOSE] Enable verbose logging to stderr (retry events, cache stats) -
                           never pollutes stdout
  -x, --extended           Show extra columns
      --bullet=<value>     Update bulleted list item text
      --callout=<value>    Update callout block text
      --code=<value>       Update code block content
      --color=<option>     Block color (for supported block types)
                           <options: default|gray|brown|orange|yellow|green|blue|purple|pink|red>
      --columns=<value>    Only show provided columns (comma-separated)
      --csv                Output in CSV format
      --filter=<value>     Filter property by substring match
      --heading-1=<value>  Update H1 heading text
      --heading-2=<value>  Update H2 heading text
      --heading-3=<value>  Update H3 heading text
      --language=<value>   [default: plain text] Update code block language (used with --code)
      --minimal            Strip unnecessary metadata (created_by, last_edited_by, object fields, request_id, etc.) -
                           reduces response size by ~40%
      --no-cache           Bypass cache and force fresh API calls
      --no-header          Hide table header from output
      --no-truncate        Do not truncate output to fit screen
      --numbered=<value>   Update numbered list item text
      --page-size=<value>  [default: 100] Items per page (1-100, default: 100 for automation)
      --quote=<value>      Update quote block text
      --retry              Auto-retry on rate limit (respects Retry-After header)
      --sort=<value>       Property to sort by (prepend with - for descending)
      --text=<value>       Update paragraph text
      --timeout=<value>    [default: 30000] Request timeout in milliseconds
      --todo=<value>       Update to-do item text
      --toggle=<value>     Update toggle block text

DESCRIPTION
  Update a block

ALIASES
  $ notion-cli block u

EXAMPLES
  Update block with simple text

    $ notion-cli block update BLOCK_ID --text "Updated content"

  Update heading content

    $ notion-cli block update BLOCK_ID --heading-1 "New Title"

  Update code block

    $ notion-cli block update BLOCK_ID --code "const x = 42;" --language javascript

  Archive a block

    $ notion-cli block update BLOCK_ID -a

  Archive a block via URL

    $ notion-cli block update https://notion.so/BLOCK_ID -a

  Update block content with complex JSON (for advanced cases)

    $ notion-cli block update BLOCK_ID -c '{"paragraph":{"rich_text":[{"text":{"content":"Updated text"}}]}}'

  Update block color

    $ notion-cli block update BLOCK_ID --color blue

  Update a block and output raw json

    $ notion-cli block update BLOCK_ID --text "Updated" -r

  Update a block and output JSON for automation

    $ notion-cli block update BLOCK_ID --text "Updated" --json
```


