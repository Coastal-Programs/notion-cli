`notion-cli config`
===================

Set NOTION_TOKEN in your shell configuration file

* [`notion-cli config set-token [TOKEN]`](#notion-cli-config-set-token-token)

## `notion-cli config set-token [TOKEN]`

Set NOTION_TOKEN in your shell configuration file

```
USAGE
  $ notion-cli config set-token [TOKEN] [-j] [--page-size <value>] [--retry] [--timeout <value>] [--no-cache] [-v]
    [--minimal]

ARGUMENTS
  [TOKEN]  Notion integration token (starts with secret_)

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
  Set NOTION_TOKEN in your shell configuration file

ALIASES
  $ notion-cli config token

EXAMPLES
  Set Notion token interactively

    $ notion-cli config set-token

  Set Notion token directly

    $ notion-cli config set-token secret_abc123...

  Set token with JSON output

    $ notion-cli config set-token secret_abc123... --json
```


