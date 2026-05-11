`notion-cli data-source`
=======================

Low-level commands for the Notion data sources API (2025-09-03).  
Aliases: `ds`, `datasource`, `data_source`.

Every Notion database has one or more data sources. These commands address
them directly by ID, bypassing the `db` command layer.

* [`notion-cli data-source retrieve DATA_SOURCE_ID`](#retrieve)
* [`notion-cli data-source create`](#create)
* [`notion-cli data-source update DATA_SOURCE_ID`](#update)
* [`notion-cli data-source query DATA_SOURCE_ID`](#query)
* [`notion-cli data-source templates DATA_SOURCE_ID`](#templates)
* [`notion-cli data-source properties update DATA_SOURCE_ID`](#properties-update)

---

## `retrieve`

Retrieve a data source by ID.

```
USAGE
  $ notion-cli data-source retrieve <data_source_id> [--json | --raw | --csv | --markdown | --pretty]

ARGUMENTS
  DATA_SOURCE_ID  32-char hex ID, hyphenated UUID, or Notion URL (with dataSource= param)
```

**Example**
```bash
notion-cli data-source retrieve aabbccdd00112233aabbccdd00112233 --json
```

---

## `create`

Create a new data source attached to an existing database.

```
USAGE
  $ notion-cli data-source create --parent-database <db_id> --properties <json> [--title <title>]

FLAGS
  --parent-database  (required) Parent database ID
  --properties       (required) Properties schema as a JSON string
  --title            Optional display title for the data source
```

---

## `update`

Update a data source's title, properties, or trash state.

```
USAGE
  $ notion-cli data-source update <data_source_id> [--title <title>] [--properties <json>] [--in-trash true|false]

FLAGS
  --title       New data source title
  --properties  Updated properties schema as JSON
  --in-trash    Move to (true) or restore from (false) trash
```

At least one flag is required.

---

## `query`

Query a data source directly with optional filters, sorts, and pagination.

```
USAGE
  $ notion-cli data-source query <data_source_id> [--filter <json>] [--sorts <json>]
      [--page-size N] [--start-cursor <cursor>]
```

**Example**
```bash
notion-cli data-source query aabbccdd00112233aabbccdd00112233 \
  --filter '{"property":"Status","select":{"equals":"Done"}}' \
  --page-size 50 \
  --json
```

---

## `templates`

List the page templates associated with a data source.

```
USAGE
  $ notion-cli data-source templates <data_source_id> [--page-size N] [--all] [--start-cursor <cursor>]

FLAGS
  --page-size N     Results per page (1–100)
  --all             Auto-paginate through all results
  --start-cursor    Resume from a pagination cursor
```

---

## `properties update`

Update the properties schema of a data source.  
The `--schema` flag accepts a JSON string or a `@file` path.

```
USAGE
  $ notion-cli data-source properties update <data_source_id> --schema <json|@file>

FLAGS
  --schema  (required) Properties map as JSON string or @path/to/schema.json
```

**Example using a file**
```bash
notion-cli data-source properties update aabbccdd00112233aabbccdd00112233 \
  --schema @schema.json \
  --json
```

---

## Migration from `db query`

`db query <database_id>` resolves the primary data source automatically, but
now prints a deprecation notice to stderr. Migrate to:

```bash
# 1. Find the data source ID
notion-cli db retrieve <database_id> --json | jq '.data.data_sources[0].id'

# 2. Query directly
notion-cli data-source query <data_source_id>
```

To suppress the deprecation notice in existing scripts while you migrate:

```bash
notion-cli db query <database_id> --quiet
```
