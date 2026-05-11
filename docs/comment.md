`notion-cli comment`
====================

Create, list, retrieve, update, and delete Notion comments.

The `comment` command group wraps the Notion Comments API
(`Notion-Version: 2026-03-11`):

| Endpoint                     | Subcommand           |
| ---------------------------- | -------------------- |
| `POST   /v1/comments`        | `comment create`     |
| `GET    /v1/comments`        | `comment list`       |
| `GET    /v1/comments/{id}`   | `comment retrieve`   |
| `PATCH  /v1/comments/{id}`   | `comment update`     |
| `DELETE /v1/comments/{id}`   | `comment delete`     |

> **Capabilities** — your connection must have **read comments** and/or
> **insert comments** capabilities enabled in the Notion Connections
> dashboard. Without them, the API returns `403`.

* [`notion-cli comment create`](#notion-cli-comment-create)
* [`notion-cli comment list`](#notion-cli-comment-list)
* [`notion-cli comment retrieve <id>`](#notion-cli-comment-retrieve-id)
* [`notion-cli comment update <id>`](#notion-cli-comment-update-id)
* [`notion-cli comment delete <id>`](#notion-cli-comment-delete-id)

---

## `notion-cli comment create`

Create a comment on a page, block, or existing discussion thread.

```
USAGE
  $ notion-cli comment create
      ( --page <page_id> | --block <block_id> | --discussion <discussion_id> )
      ( --text "..." | --rich-text <file.json> )
      [--display-name "..."]
      [--attach-file <upload_id>]...
      [--json | --csv | --markdown | --pretty | --raw]

FLAGS
  --page <id>           Add a top-level comment to this page (mutually exclusive)
  --block <id>          Add a top-level comment to this block (mutually exclusive)
  --discussion <id>     Reply to an existing discussion thread (mutually exclusive)
  --text <string>       Comment body as plain text (becomes a single rich_text run)
  --rich-text <file>    Path to a JSON file containing a rich_text array
  --display-name <s>    Custom author display name (integration-defined)
  --attach-file <id>    File upload ID to attach (repeatable, max 3)
```

Exactly one of `--page`, `--block`, or `--discussion` must be provided.
Exactly one of `--text` or `--rich-text` must be provided.

The `--rich-text` flag loads a JSON array of rich text objects from a
file. Example contents:

```json
[
  { "type": "text", "text": { "content": "Hello " } },
  { "type": "text", "text": { "content": "world", "link": { "url": "https://notion.so" } } }
]
```

`--attach-file` may be passed multiple times (up to 3) to attach
previously-uploaded files. The argument is the `id` returned by the
File Upload API.

EXAMPLES

```bash
# Top-level comment on a page
notion-cli comment create --page <page_id> --text "Looks good!"

# Reply to a discussion
notion-cli comment create --discussion <discussion_id> --text "Agreed."

# Rich-text body loaded from a file, with custom display name and an attachment
notion-cli comment create \
  --block <block_id> \
  --rich-text ./body.json \
  --display-name "Build Bot" \
  --attach-file <file_upload_id>
```

---

## `notion-cli comment list`

List open (unresolved) comments for a block (or page — pages are blocks).

```
USAGE
  $ notion-cli comment list --block <id> [--page-size N] [--start-cursor <cursor>] [--all]

FLAGS
  --block <id>            Block or page ID whose comments to list (required)
  --page-size <int>       Page size, max 100
  --start-cursor <str>    Continue from a previous cursor
  --all                   Paginate until all comments are retrieved
```

The Notion API's `GET /v1/comments` endpoint uses a single `block_id`
query parameter for both blocks and pages, since pages are technically
blocks. Pass either a page or block ID via `--block`.

EXAMPLES

```bash
# Single page of results
notion-cli comment list --block <page_id> --json

# Iterate the entire thread
notion-cli comment list --block <block_id> --all --json
```

---

## `notion-cli comment retrieve <id>`

Retrieve a single comment by ID.

```
USAGE
  $ notion-cli comment retrieve <comment_id>
```

EXAMPLES

```bash
notion-cli comment retrieve 7a793800-3e55-4d5e-8009-2261de026179 --json
```

---

## `notion-cli comment update <id>`

Update an existing comment's body. Exactly one of `--text` or
`--rich-text` must be provided.

```
USAGE
  $ notion-cli comment update <comment_id> ( --text "..." | --rich-text <file.json> )
```

A connection can only update comments that it created.

EXAMPLES

```bash
notion-cli comment update <comment_id> --text "Edited body"
notion-cli comment update <comment_id> --rich-text ./body.json
```

---

## `notion-cli comment delete <id>`

Delete a comment. A connection can only delete comments that it created.

```
USAGE
  $ notion-cli comment delete <comment_id>
```

EXAMPLES

```bash
notion-cli comment delete 7a793800-3e55-4d5e-8009-2261de026179
```
