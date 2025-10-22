# Database Query Filtering Guide

This guide explains how to filter database queries using the `db query` command in notion-cli.

## Quick Reference

### JSON Filter (Primary Method)

Use `--filter` with a JSON object following Notion's filter API format:

```bash
notion-cli db query DS_ID --filter '{"property": "Status", "select": {"equals": "Done"}}'
```

### Text Search (Convenience)

Use `--search` for simple text matching across common properties:

```bash
notion-cli db query DS_ID --search "urgent"
```

### File Filter (Complex Queries)

Use `--file-filter` to load complex filters from a JSON file:

```bash
# filter.json:
# {
#   "and": [
#     {"property": "Status", "select": {"equals": "Done"}},
#     {"property": "Priority", "number": {"greater_than": 5}}
#   ]
# }

notion-cli db query DS_ID --file-filter ./filter.json
```

## Filter Flags

### `--filter` (Recommended)

- **Description**: JSON object matching Notion's filter API format
- **Use case**: Programmatic filtering by AI agents and automation scripts
- **Shorthand**: `-f`

```bash
notion-cli db query DS_ID -f '{"property": "Status", "select": {"equals": "Done"}}' --json
```

### `--search`

- **Description**: Simple text search across common properties (Name, Title, Description)
- **Use case**: Quick human searches without writing JSON
- **Shorthand**: `-s`

```bash
notion-cli db query DS_ID -s "urgent" --json
```

### `--file-filter`

- **Description**: Load filter from a JSON file
- **Use case**: Complex filters that are reused or too long for command line
- **Shorthand**: `-F`

```bash
notion-cli db query DS_ID -F ./my-filter.json --json
```

## Filter Examples by Property Type

### Select Property

Filter by select property value:

```json
{
  "property": "Status",
  "select": {
    "equals": "In Progress"
  }
}
```

Available operators: `equals`, `does_not_equal`, `is_empty`, `is_not_empty`

### Multi-Select Property

Filter by multi-select property:

```json
{
  "property": "Tags",
  "multi_select": {
    "contains": "urgent"
  }
}
```

Available operators: `contains`, `does_not_contain`, `is_empty`, `is_not_empty`

### Number Property

Filter by number:

```json
{
  "property": "Priority",
  "number": {
    "greater_than_or_equal_to": 5
  }
}
```

Available operators: `equals`, `does_not_equal`, `greater_than`, `less_than`, `greater_than_or_equal_to`, `less_than_or_equal_to`, `is_empty`, `is_not_empty`

### Checkbox Property

Filter by checkbox state:

```json
{
  "property": "Completed",
  "checkbox": {
    "equals": true
  }
}
```

Available operators: `equals`, `does_not_equal`

### Date Property

Filter by date:

```json
{
  "property": "Due Date",
  "date": {
    "on_or_before": "2025-12-31"
  }
}
```

Available operators: `equals`, `before`, `after`, `on_or_before`, `on_or_after`, `is_empty`, `is_not_empty`, `past_week`, `past_month`, `past_year`, `next_week`, `next_month`, `next_year`

### Text/Title Property

Filter by text content (title or rich_text):

```json
{
  "property": "Name",
  "title": {
    "contains": "Project"
  }
}
```

```json
{
  "property": "Description",
  "rich_text": {
    "contains": "bug"
  }
}
```

Available operators: `equals`, `does_not_equal`, `contains`, `does_not_contain`, `starts_with`, `ends_with`, `is_empty`, `is_not_empty`

### People Property

Filter by person:

```json
{
  "property": "Assigned To",
  "people": {
    "contains": "user-id-here"
  }
}
```

Available operators: `contains`, `does_not_contain`, `is_empty`, `is_not_empty`

### Files Property

Filter by file presence:

```json
{
  "property": "Attachments",
  "files": {
    "is_empty": true
  }
}
```

Available operators: `is_empty`, `is_not_empty`

### Relation Property

Filter by relation:

```json
{
  "property": "Related Items",
  "relation": {
    "contains": "page-id-here"
  }
}
```

Available operators: `contains`, `does_not_contain`, `is_empty`, `is_not_empty`

## Combining Filters

### AND Logic

All conditions must be true:

```json
{
  "and": [
    {"property": "Status", "select": {"equals": "Done"}},
    {"property": "Priority", "number": {"greater_than": 5}}
  ]
}
```

Example command:
```bash
notion-cli db query DS_ID --filter '{"and": [{"property": "Status", "select": {"equals": "Done"}}, {"property": "Priority", "number": {"greater_than": 5}}]}' --json
```

### OR Logic

At least one condition must be true:

```json
{
  "or": [
    {"property": "Status", "select": {"equals": "Urgent"}},
    {"property": "Priority", "number": {"equals": 10}}
  ]
}
```

Example command:
```bash
notion-cli db query DS_ID --filter '{"or": [{"property": "Status", "select": {"equals": "Urgent"}}, {"property": "Priority", "number": {"equals": 10}}]}' --json
```

### Nested Logic

Combine AND and OR for complex queries:

```json
{
  "and": [
    {
      "or": [
        {"property": "Status", "select": {"equals": "In Progress"}},
        {"property": "Status", "select": {"equals": "Review"}}
      ]
    },
    {"property": "Priority", "number": {"greater_than_or_equal_to": 7}}
  ]
}
```

This finds items that are either "In Progress" OR "Review", AND have priority >= 7.

## Common Use Cases

### Find Completed High Priority Tasks

```bash
notion-cli db query DS_ID \
  --filter '{"and": [{"property": "Status", "select": {"equals": "Done"}}, {"property": "Priority", "number": {"greater_than": 7}}]}' \
  --json
```

### Find Items Due This Week

```bash
notion-cli db query DS_ID \
  --filter '{"property": "Due Date", "date": {"next_week": {}}}' \
  --json
```

### Find Unassigned Tasks

```bash
notion-cli db query DS_ID \
  --filter '{"property": "Assigned To", "people": {"is_empty": true}}' \
  --json
```

### Find Items Tagged as Urgent or Bug

```bash
notion-cli db query DS_ID \
  --filter '{"or": [{"property": "Tags", "multi_select": {"contains": "urgent"}}, {"property": "Tags", "multi_select": {"contains": "bug"}}]}' \
  --json
```

### Find Items Without Attachments

```bash
notion-cli db query DS_ID \
  --filter '{"property": "Attachments", "files": {"is_empty": true}}' \
  --json
```

## Using Filter Files

For complex or reusable filters, save them to a file:

**filter-high-priority.json:**
```json
{
  "and": [
    {
      "or": [
        {"property": "Status", "select": {"equals": "In Progress"}},
        {"property": "Status", "select": {"equals": "Review"}}
      ]
    },
    {"property": "Priority", "number": {"greater_than_or_equal_to": 8}},
    {"property": "Assigned To", "people": {"is_not_empty": true}}
  ]
}
```

**Usage:**
```bash
notion-cli db query DS_ID --file-filter ./filter-high-priority.json --json
```

## Text Search Behavior

The `--search` flag is a convenience wrapper that searches across common properties:

```bash
notion-cli db query DS_ID --search "urgent"
```

This is equivalent to:
```bash
notion-cli db query DS_ID --filter '{"or": [{"property": "Name", "title": {"contains": "urgent"}}, {"property": "Title", "title": {"contains": "urgent"}}, {"property": "Description", "rich_text": {"contains": "urgent"}}, {"property": "Name", "rich_text": {"contains": "urgent"}}]}'
```

**Limitations:**
- Only searches predefined properties: Name, Title, Description
- Case-sensitive matching
- For more control, use `--filter` with explicit property names

## Migration from Old Flags

If you're using deprecated flags, here's how to migrate:

**Old (deprecated):**
```bash
notion-cli db query DS_ID --rawFilter '{"property": "Status", "select": {"equals": "Done"}}'
notion-cli db query DS_ID --fileFilter ./filter.json
```

**New:**
```bash
notion-cli db query DS_ID --filter '{"property": "Status", "select": {"equals": "Done"}}'
notion-cli db query DS_ID --file-filter ./filter.json
```

The old flags (`--rawFilter`, `--fileFilter`) will continue to work but will show deprecation warnings. They will be removed in v6.0.0.

## Troubleshooting

### Invalid JSON Error

If you get "Invalid JSON in --filter", check:
1. Use single quotes around the filter: `--filter '...'`
2. Use double quotes for JSON keys and values: `{"property": "Name"}`
3. Escape special characters if using double quotes in bash

### Property Not Found

If filtering fails with property not found:
1. Check the exact property name in Notion (case-sensitive)
2. Use `notion-cli db retrieve DS_ID --json` to see all property names
3. Ensure the property type matches the filter type (e.g., `select` filter for select property)

### No Results

If you expect results but get none:
1. Test without filter to see all data: `notion-cli db query DS_ID`
2. Simplify the filter to isolate which condition is failing
3. Check property values match exactly (case-sensitive)

## Additional Resources

- [Notion Filter API Reference](https://developers.notion.com/reference/post-database-query-filter)
- [notion-cli README](../README.md)
- [Output Formats Guide](../OUTPUT_FORMATS.md)

## Examples Library

Save these to files for quick reuse:

**active-tasks.json** - Active tasks assigned to someone:
```json
{
  "and": [
    {"property": "Status", "select": {"does_not_equal": "Done"}},
    {"property": "Status", "select": {"does_not_equal": "Cancelled"}},
    {"property": "Assigned To", "people": {"is_not_empty": true}}
  ]
}
```

**overdue-items.json** - Items past their due date:
```json
{
  "and": [
    {"property": "Due Date", "date": {"before": "2025-10-23"}},
    {"property": "Status", "select": {"does_not_equal": "Done"}}
  ]
}
```

**needs-review.json** - Items in review without reviewers:
```json
{
  "and": [
    {"property": "Status", "select": {"equals": "Review"}},
    {"property": "Reviewer", "people": {"is_empty": true}}
  ]
}
```
