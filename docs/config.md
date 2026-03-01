`notion-cli config`
===================

Configuration management

* [`notion-cli config set-token [TOKEN]`](#notion-cli-config-set-token-token)
* [`notion-cli config get KEY`](#notion-cli-config-get-key)
* [`notion-cli config path`](#notion-cli-config-path)
* [`notion-cli config list`](#notion-cli-config-list)

## `notion-cli config set-token [TOKEN]`

Save the Notion API token to the config file

```
USAGE
  $ notion-cli config set-token [TOKEN]

ARGUMENTS
  [TOKEN]  Notion integration token (starts with secret_ or ntn_)

DESCRIPTION
  Save the Notion API token to the config file.

  If no argument is provided, the token is read from stdin.
  You can pipe a token via stdin to avoid exposing it in process listings:

    echo "$NOTION_TOKEN" | notion-cli config set-token
    notion-cli config set-token < token-file.txt

ALIASES
  $ notion-cli config token

EXAMPLES
  Set Notion token interactively (prompts for input)

    $ notion-cli config set-token

  Set Notion token directly (warning: exposes token in process listing)

    $ notion-cli config set-token secret_abc123...

  Set token via pipe (recommended for security)

    $ echo "$NOTION_TOKEN" | notion-cli config set-token
```



## `notion-cli config get KEY`

Get a configuration value by key

```
USAGE
  $ notion-cli config get KEY [--show-secret]

ARGUMENTS
  KEY  Configuration key to retrieve (e.g., token, base_url, max_retries)

FLAGS
  --show-secret  Show unmasked token value (token is masked by default)

DESCRIPTION
  Get a configuration value by key.

EXAMPLES
  Get the configured token (masked)

    $ notion-cli config get token

  Get the configured token (unmasked)

    $ notion-cli config get token --show-secret

  Get the API base URL

    $ notion-cli config get base_url
```



## `notion-cli config path`

Show the config file path

```
USAGE
  $ notion-cli config path

DESCRIPTION
  Show the full path to the configuration file.

EXAMPLES
  Show config file path

    $ notion-cli config path
```



## `notion-cli config list`

List all configuration values

```
USAGE
  $ notion-cli config list [-j]

FLAGS
  -j, --json  Output as JSON

DESCRIPTION
  List all configuration values. Token values are masked for security.

  Displays: token, base_url, max_retries, base_delay_ms, max_delay_ms,
  cache_enabled, cache_max_size, disk_cache_enabled, http_keep_alive, verbose.

ALIASES
  $ notion-cli config ls

EXAMPLES
  List all config values

    $ notion-cli config list

  List all config values as JSON

    $ notion-cli config list --json
```

