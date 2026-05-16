# API Command

`notion-cli api` makes authenticated Notion API requests from the terminal with
syntax compatible with the official `ntn api` command.

## Examples

```bash
# GET request
notion-cli api v1/users/me

# POST request with typed inline JSON
notion-cli api v1/search query=roadmap page_size:=10

# PATCH request
notion-cli api "v1/pages/$PAGE_ID" -X PATCH archived:=true

# Query parameters and headers
notion-cli api v1/file_uploads page_size==25 Accept:application/json

# JSON body from stdin or --data
jq -n '{query:"roadmap",page_size:10}' | notion-cli api v1/search
notion-cli api v1/search --data '{"query":"roadmap","page_size":10}'

# Multipart file send
notion-cli api "v1/file_uploads/$FILE_UPLOAD_ID/send" --file ./chunk.bin part_number:=1
```

## Inline Syntax

- `path=value`: body field with a string value.
- `path:=json`: body field parsed as JSON.
- `name==value`: query parameter.
- `Header:Value`: request header.

Body paths support bracket notation, dot notation, explicit array indexes, and
`[]` appends.

## Introspection

```bash
notion-cli api ls
notion-cli api ls --json
notion-cli api v1/comments --spec -X POST
notion-cli api v1/comments --docs -X POST
```

The introspection catalog is embedded for deterministic offline output. Refresh
the source material with `scripts/refresh_notion_api_catalog.sh` before updating
the catalog.

## Compatibility Notes

`NOTION_TOKEN` remains the primary automation token. `NOTION_API_TOKEN` is also
accepted for official CLI compatibility when `NOTION_TOKEN` is unset.

`NOTION_API_VERSION` and `--notion-version` override the `Notion-Version` header.
