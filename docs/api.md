`notion-cli api`
================

Make an authenticated request to any Notion API endpoint

* [`notion-cli api <endpoint>`](#notion-cli-api-endpoint)

## `notion-cli api <endpoint>`

Make an authenticated HTTP request to the Notion API and print the response.

```
USAGE
  $ notion-cli api <endpoint> [-X <method>] [-f key=value]... [-F key=json]... [-H 'Name: value']... [--input <file|->] [--paginate] [-o <format>]

FLAGS
  -X, --method=<value>      HTTP method (default GET, or POST when a body is given)
  -f, --raw-field=<value>   Add a string field: key=value (repeatable)
  -F, --field=<value>       Add a typed field: key=value, value is JSON; @file or @- reads it (repeatable)
  -H, --header=<value>      Add an HTTP header: 'Name: value' (repeatable)
      --input=<value>       Read the JSON request body from a file, or '-' for stdin
      --paginate            Follow next_cursor and merge all pages' results
  -o, --output=<value>      Output format: json, compact-json, csv, markdown, table, raw, pretty
      --raw                 Output raw API response without envelope

DESCRIPTION
  The endpoint is a path under https://api.notion.com/v1, such as "users/me" or
  "data_sources/<id>/query". Authentication, Notion-Version, gzip, retries, and
  token refresh are handled exactly as for every other command, so endpoints
  this CLI has no dedicated command for (agents, sessions, meeting notes, async
  tasks, betas) are reachable today.

  The method defaults to GET, or POST when any body input is supplied. Body
  fields come from --input (a JSON file, or "-" for stdin) and/or repeated
  -f / -F flags; flags are merged on top of --input. With an explicit -X GET or
  -X DELETE, -f / -F values become query parameters instead. Keys may use dots
  to nest (parent.page_id=...). -f values are strings; -F values are parsed as
  JSON (numbers, booleans, null, objects), and "@file" or "@-" reads the value
  from a file or stdin.

  --paginate follows next_cursor until has_more is false and merges every
  page's results into one list. Notion reads the cursor from the query on GET
  and from the body on POST; the command places it accordingly. Against a
  non-list endpoint the single response is printed unchanged.

  Absolute URLs are rejected: the bearer token is only ever sent to the
  configured API host. Non-2xx responses are reported with the same structured
  error envelope and exit code as every other command.

EXAMPLES
  Who am I?

    $ notion-cli api users/me

  List users, 10 per page (fields on an explicit GET become query params)

    $ notion-cli api users -X GET -f page_size=10

  Query a data source with a filter from a file, fetching every page

    $ notion-cli api data_sources/<id>/query --input filter.json --paginate

  Create a page with nested fields

    $ notion-cli api pages -f parent.page_id=<id> -F 'properties={"title":[{"text":{"content":"Hi"}}]}'

  Retrieve an async task

    $ notion-cli api async_tasks/<task_id>

  Opt into a beta for one call

    $ notion-cli api -H 'Notion-Beta: workspace-as-code' workspaces/<id>/export

  Move a page to trash

    $ notion-cli api -X PATCH pages/<id> -F in_trash=true
```
