`notion-cli custom-emoji`
========================

List and retrieve workspace custom emojis (Notion API 2026-03-11).

* [`notion-cli custom-emoji list`](#notion-cli-custom-emoji-list)
* [`notion-cli custom-emoji retrieve <id>`](#notion-cli-custom-emoji-retrieve-id)

## `notion-cli custom-emoji list`

List custom emojis in the workspace.

```
USAGE
  $ notion-cli custom-emoji list [--page-size <value>] [--start-cursor <value>] [--all] [-j]

FLAGS
  -j, --json                 Output as JSON (recommended for automation)
      --all                  Paginate until all custom emojis are retrieved
      --page-size=<value>    Number of results per page (max 100)
      --start-cursor=<value> Pagination cursor for resuming from a specific page

DESCRIPTION
  List custom emojis in the workspace. Output is a table with columns: id, name, url.
  Use --all to automatically paginate through every page of results.

ALIASES
  $ notion-cli custom-emojis list
  $ notion-cli emoji list

EXAMPLES
  List the first page of custom emojis

    $ notion-cli custom-emoji list

  List all custom emojis (auto-paginate)

    $ notion-cli custom-emoji list --all

  List custom emojis with a specific page size

    $ notion-cli custom-emoji list --page-size 10

  Output as JSON for automation

    $ notion-cli custom-emoji list --json
```



## `notion-cli custom-emoji retrieve <id>`

Retrieve a single workspace custom emoji by ID.

```
USAGE
  $ notion-cli custom-emoji retrieve <custom_emoji_id> [-j]

ARGUMENTS
  custom_emoji_id  The ID of the custom emoji to retrieve

FLAGS
  -j, --json  Output as JSON (recommended for automation)

DESCRIPTION
  Retrieve a single workspace custom emoji by its ID.

ALIASES
  $ notion-cli custom-emoji r <id>
  $ notion-cli custom-emoji get <id>

EXAMPLES
  Retrieve a custom emoji

    $ notion-cli custom-emoji retrieve abc-123

  Retrieve and output as JSON

    $ notion-cli custom-emoji retrieve abc-123 --json
```
