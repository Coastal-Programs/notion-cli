`notion-cli files`
==================

Upload, retrieve, and list Notion file uploads.

The `files` command group wraps the Notion File Uploads API:

| Endpoint                              | Subcommand        |
| ------------------------------------- | ----------------- |
| `POST   /v1/file_uploads`             | (internal)        |
| `POST   /v1/file_uploads/{id}/send`   | `files upload`    |
| `POST   /v1/file_uploads/{id}/complete` | `files upload`  |
| `GET    /v1/file_uploads/{id}`        | `files retrieve`  |
| `GET    /v1/file_uploads`             | `files list`      |

> **Single-part vs multi-part** — files ≤20 MB are uploaded in a single
> request. Files >20 MB are split into chunks (default 10 MB each) and
> uploaded using the multi-part API: create → send each chunk → complete.

* [`notion-cli files upload <path>`](#notion-cli-files-upload-path)
* [`notion-cli files retrieve <id>`](#notion-cli-files-retrieve-id)
* [`notion-cli files list`](#notion-cli-files-list)

---

## `notion-cli files upload <path>`

Upload a local file to Notion. The `file_upload_id` is printed to stdout for
piping into other commands (e.g. `comment create --attach-file`). Use `--json`
to get the full response envelope.

```
USAGE
  $ notion-cli files upload <path> [--chunk-size <size>] [--json | --csv | --markdown | --pretty | --raw]

FLAGS
  --chunk-size <size>   Chunk size for multi-part uploads, e.g. 10MB (default: 10MB; range: 5MB–20MB)
```

Files ≤20 MB use a single-part upload (one `send` request, no `complete`
call). Files >20 MB are split into chunks and require a `complete` call after
all parts are sent. A progress bar is shown on stderr when stderr is a TTY.

EXAMPLES

```bash
# Upload a small file and pipe the ID to comment create
notion-cli files upload ./screenshot.png | xargs -I{} notion-cli comment create \
  --page <page_id> --text "See attached" --attach-file {}

# Upload a large file with a custom chunk size and inspect the full response
notion-cli files upload ./backup.zip --chunk-size 20MB --json

# Upload and capture the file upload ID into a shell variable
FILE_ID=$(notion-cli files upload ./report.pdf)
echo "Uploaded: $FILE_ID"
```

---

## `notion-cli files retrieve <id>`

Retrieve the status and metadata of a Notion file upload by ID.

```
USAGE
  $ notion-cli files retrieve <file_upload_id> [--json | --csv | --markdown | --pretty | --raw]
```

EXAMPLES

```bash
notion-cli files retrieve abc123 --json
```

---

## `notion-cli files list`

List Notion file uploads. Supports manual pagination via `--page-size` or
automatic full pagination via `--all`.

```
USAGE
  $ notion-cli files list [--page-size N] [--all] [--json | --csv | --markdown | --pretty | --raw]

FLAGS
  --page-size <int>   Number of results per page (default: 50)
  --all               Paginate until all file uploads are retrieved
```

EXAMPLES

```bash
# List the first page of file uploads
notion-cli files list --json

# Retrieve all file uploads
notion-cli files list --all --json

# List with a custom page size
notion-cli files list --page-size 10 --json
```
