`notion-cli markdown`
=====================

Read and write a Notion page's content as enhanced markdown using the
`/v1/pages/{page_id}/markdown` endpoints introduced in API version
`2026-03-11`.

* [`notion-cli markdown get PAGE`](#notion-cli-markdown-get-page)
* [`notion-cli markdown set PAGE`](#notion-cli-markdown-set-page)

The page argument accepts a Notion page ID or any Notion URL —
`internal/resolver.ExtractID` is applied.

## `notion-cli markdown get`

Retrieve a page as enhanced markdown.

```
USAGE
  $ notion-cli markdown get <page-id-or-url> [--file PATH] [--json | --csv | ...]

FLAGS
  --file=<path>     Write markdown to this file path (instead of stdout)
  -o, --output=<f>  Output format: json, compact-json, csv, markdown, table, raw, pretty
  --json            Output as JSON envelope
  ...

DESCRIPTION
  Prints the raw markdown body to stdout by default. When --file is supplied,
  the markdown is written to that path instead. Passing any structured output
  flag (--json, --output json, --csv, etc.) returns the response wrapped in
  the standard envelope, with the markdown body exposed under the `content`
  key alongside `id`, `truncated` and `unknown_block_ids`.
```

### Examples

```bash
# Print page content to stdout
notion-cli markdown get 11111111111111111111111111111111

# Save to a file
notion-cli markdown get https://notion.so/My-Page-1111... --file ./out.md

# JSON envelope (for scripts)
notion-cli markdown get <page> --json | jq -r .data.content
```

## `notion-cli markdown set`

Replace or append a page's content using enhanced markdown.

```
USAGE
  $ notion-cli markdown set <page-id-or-url> [--content STR | --file PATH | <stdin>]
                                            [--append] [--allow-deleting-content]

FLAGS
  --content=<str>             Markdown content as a literal string
  --file=<path>               Path to a markdown file
  --append                    Append content instead of replacing the page body
  --allow-deleting-content    Permit removal of child pages or databases
  -o, --output=<f>            Output format (json, table, ...)

DESCRIPTION
  Provide content via --content, --file, or stdin (default). --content and
  --file are mutually exclusive; if neither is supplied, stdin is read.

  Without --append, the entire page body is replaced (`replace_content`).
  With --append, the markdown is inserted at the end of the page
  (`insert_content` with no `after` selector).

  By default, the API refuses to delete child pages or databases. Pass
  --allow-deleting-content to permit it.
```

### Examples

```bash
# Replace the page body with a literal string
notion-cli markdown set <page> --content "# Hello\n\nWorld"

# Replace from a file
notion-cli markdown set <page> --file ./README.md

# Pipe via stdin
echo "# Updated" | notion-cli markdown set <page>

# Append a section
notion-cli markdown set <page> --append --content "## Changelog\n\n- v1.2.3"

# Round-trip a page through your editor
notion-cli markdown get <page> --file /tmp/page.md
$EDITOR /tmp/page.md
notion-cli markdown set <page> --file /tmp/page.md
```

## See also

- [Retrieve a page as markdown](https://developers.notion.com/reference/retrieve-page-markdown)
- [Update a page's content as markdown](https://developers.notion.com/reference/update-page-markdown)
