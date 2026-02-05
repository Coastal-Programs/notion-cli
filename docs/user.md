`notion-cli user`
=================

List all users

* [`notion-cli user list`](#notion-cli-user-list)
* [`notion-cli user retrieve [USER_ID]`](#notion-cli-user-retrieve-user_id)
* [`notion-cli user retrieve bot`](#notion-cli-user-retrieve-bot)

## `notion-cli user list`

List all users

```
USAGE
  $ notion-cli user list [-r] [--columns <value> | -x] [--sort <value>] [--filter <value>] [--csv |
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
  List all users

ALIASES
  $ notion-cli user l

EXAMPLES
  List all users

    $ notion-cli user list

  List all users and output raw json

    $ notion-cli user list -r

  List all users and output JSON for automation

    $ notion-cli user list --json
```



## `notion-cli user retrieve [USER_ID]`

Retrieve a user

```
USAGE
  $ notion-cli user retrieve [USER_ID] [-r] [--columns <value> | -x] [--sort <value>] [--filter <value>] [--csv |
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
  Retrieve a user

ALIASES
  $ notion-cli user r

EXAMPLES
  Retrieve a user

    $ notion-cli user retrieve USER_ID

  Retrieve a user and output raw json

    $ notion-cli user retrieve USER_ID -r

  Retrieve a user and output JSON for automation

    $ notion-cli user retrieve USER_ID --json
```



## `notion-cli user retrieve bot`

Retrieve a bot user

```
USAGE
  $ notion-cli user retrieve bot [-r] [--columns <value> | -x] [--sort <value>] [--filter <value>] [--csv |
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
  Retrieve a bot user

ALIASES
  $ notion-cli user r b

EXAMPLES
  Retrieve a bot user

    $ notion-cli user retrieve:bot

  Retrieve a bot user and output raw json

    $ notion-cli user retrieve:bot -r

  Retrieve a bot user and output JSON for automation

    $ notion-cli user retrieve:bot --json
```


