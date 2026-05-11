`notion-cli view`
=================

Create, list, retrieve, update, delete, and query Notion database views.

The `view` command group wraps the Notion Views API
(`Notion-Version: 2026-03-11`):

| Endpoint                                         | Subcommand             |
| ------------------------------------------------ | ---------------------- |
| `POST   /v1/views`                               | `view create`          |
| `GET    /v1/views`                               | `view list`            |
| `GET    /v1/views/{view_id}`                     | `view retrieve`        |
| `PATCH  /v1/views/{view_id}`                     | `view update`          |
| `DELETE /v1/views/{view_id}`                     | `view delete`          |
| `POST   /v1/views/{view_id}/queries`             | `view query`           |
| `GET    /v1/views/{view_id}/queries/{query_id}`  | `view results`         |
| `DELETE /v1/views/{view_id}/queries/{query_id}`  | `view delete-query`    |

* [`notion-cli view create`](#notion-cli-view-create)
* [`notion-cli view list`](#notion-cli-view-list)
* [`notion-cli view retrieve <view_id>`](#notion-cli-view-retrieve-view_id)
* [`notion-cli view update <view_id>`](#notion-cli-view-update-view_id)
* [`notion-cli view delete <view_id>`](#notion-cli-view-delete-view_id)
* [`notion-cli view query <view_id>`](#notion-cli-view-query-view_id)
* [`notion-cli view results <view_id> <query_id>`](#notion-cli-view-results-view_id-query_id)
* [`notion-cli view delete-query <view_id> <query_id>`](#notion-cli-view-delete-query-view_id-query_id)

---

## `notion-cli view create`

Create a new Notion database view.

```
USAGE
  $ notion-cli view create
      --data-source <data_source_id>
      --name <name>
      --type <type>
      [--database <database_id>]
      [--filter <json-or-@file>]
      [--sorts <json-or-@file>]
      [--json | --csv | --markdown | --pretty | --raw]

FLAGS
  --data-source <id>    Data source ID the view belongs to (required)
  --name <string>       Display name for the view (required)
  --type <string>       View type (required): table | board | list | calendar |
                        timeline | gallery | form | chart | map | dashboard
  --database <id>       Database ID to associate with the view
  --filter <json|@file> Filter as a JSON object, or @<path> to load from file
  --sorts <json|@file>  Sorts as a JSON array, or @<path> to load from file
```

EXAMPLES

```bash
# Simple table view
notion-cli view create \
  --data-source <data_source_id> \
  --name "My Table" \
  --type table

# Board view with a filter loaded from a file
notion-cli view create \
  --data-source <data_source_id> \
  --name "Active Boards" \
  --type board \
  --filter @./filter.json
```

---

## `notion-cli view list`

List all views for a database or data source. Exactly one of
`--data-source` or `--database` is required.

```
USAGE
  $ notion-cli view list ( --data-source <id> | --database <id> )
      [--page-size N] [--start-cursor <cursor>]
      [--json | --csv | --markdown | --pretty | --raw]

FLAGS
  --data-source <id>      Data source ID (lists all linked views workspace-wide)
  --database <id>         Database ID (lists views on that database)
  --page-size <int>       Page size, max 100
  --start-cursor <str>    Continue from a previous cursor
```

EXAMPLES

```bash
notion-cli view list --database <database_id> --json
notion-cli view list --data-source <data_source_id>
```

---

## `notion-cli view retrieve <view_id>`

Retrieve a single view by ID.

```
USAGE
  $ notion-cli view retrieve <view_id>
```

EXAMPLES

```bash
notion-cli view retrieve 5c6a2821-6bb1-4a7e-b6e1-c50111515c3d --json
```

---

## `notion-cli view update <view_id>`

Update an existing Notion view. At least one of `--name`, `--filter`, or
`--sorts` must be provided.

```
USAGE
  $ notion-cli view update <view_id>
      [--name <string>]
      [--filter <json-or-@file>]
      [--sorts <json-or-@file>]
      [--json | --csv | --markdown | --pretty | --raw]

FLAGS
  --name <string>       New display name for the view
  --filter <json|@file> Replacement filter as a JSON object, or @<path>
  --sorts <json|@file>  Replacement sorts as a JSON array, or @<path>
```

EXAMPLES

```bash
# Rename a view
notion-cli view update <view_id> --name "Renamed View"

# Replace filter from a file
notion-cli view update <view_id> --filter @./filter.json
```

---

## `notion-cli view delete <view_id>`

Delete a Notion view by ID.

```
USAGE
  $ notion-cli view delete <view_id>
```

EXAMPLES

```bash
notion-cli view delete 5c6a2821-6bb1-4a7e-b6e1-c50111515c3d
```

---

## `notion-cli view query <view_id>`

Execute a view's filter and sort configuration against its data source,
caching the result set on the server.

Without `--all`, returns the first page of results and a `query_id` you
can pass to `view results` for manual pagination. Cached results expire
15 minutes after creation.

With `--all`, paginates through every page automatically and deletes the
server-side query cache when done.

```
USAGE
  $ notion-cli view query <view_id>
      [--page-size N]
      [--all]
      [--json | --csv | --markdown | --pretty | --raw]

FLAGS
  --page-size <int>   Results per page (max 100)
  --all               Paginate through all results and clean up the query
```

EXAMPLES

```bash
# Single-page query — returns first page + query_id
notion-cli view query <view_id> --json

# Full auto-paginated query — returns all results in one envelope
notion-cli view query <view_id> --all --json
```

---

## `notion-cli view results <view_id> <query_id>`

Paginate through results of a previously created view query.

Cached results expire 15 minutes after the query was created; expired
queries return a 404. Use `view query` to re-create the cache.

```
USAGE
  $ notion-cli view results <view_id> <query_id>
      [--page-size N] [--start-cursor <cursor>]

FLAGS
  --page-size <int>       Results per page (max 100)
  --start-cursor <str>    Continue from a previous cursor
```

EXAMPLES

```bash
notion-cli view results <view_id> <query_id> --start-cursor <cursor> --json
```

---

## `notion-cli view delete-query <view_id> <query_id>`

Delete a cached view query. Idempotent — returns success even if the
query has already expired or never existed.

```
USAGE
  $ notion-cli view delete-query <view_id> <query_id>
```

EXAMPLES

```bash
notion-cli view delete-query <view_id> <query_id>
```
