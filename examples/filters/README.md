# Filter Examples

This directory contains reusable filter examples for the `db query` command.

## Usage

```bash
notion-cli db query <DATABASE_ID> --file-filter ./examples/filters/<filter-name>.json --json
```

## Available Filters

### active-tasks.json
Finds tasks that are:
- Not done
- Not cancelled
- Assigned to someone

**Use case**: Daily standup, checking team workload

```bash
notion-cli db query <DB_ID> --file-filter ./examples/filters/active-tasks.json --json
```

### high-priority-tasks.json
Finds tasks that are:
- Not done
- Not cancelled
- Priority >= 8
- Assigned to someone

**Use case**: Critical task management, escalation tracking

```bash
notion-cli db query <DB_ID> --file-filter ./examples/filters/high-priority-tasks.json --json
```

### overdue-items.json
Finds items that are:
- Due date before today (2025-10-23)
- Not done

**Use case**: Overdue task reports, deadline tracking

```bash
notion-cli db query <DB_ID> --file-filter ./examples/filters/overdue-items.json --json
```

### needs-review.json
Finds items that are:
- Status is "Review"
- No reviewer assigned

**Use case**: Review queue management, unassigned work

```bash
notion-cli db query <DB_ID> --file-filter ./examples/filters/needs-review.json --json
```

## Creating Custom Filters

1. Create a JSON file with your filter definition
2. Follow Notion's filter API format (see [Filter Guide](../../docs/FILTER_GUIDE.md))
3. Test with `--file-filter` flag

Example custom filter:
```json
{
  "and": [
    {
      "property": "Status",
      "select": {
        "equals": "In Progress"
      }
    },
    {
      "property": "Assignee",
      "people": {
        "contains": "USER_ID_HERE"
      }
    }
  ]
}
```

## Common Property Types

- **Select**: `{"property": "Status", "select": {"equals": "Done"}}`
- **Multi-select**: `{"property": "Tags", "multi_select": {"contains": "urgent"}}`
- **Number**: `{"property": "Priority", "number": {"greater_than": 5}}`
- **Date**: `{"property": "Due Date", "date": {"before": "2025-12-31"}}`
- **People**: `{"property": "Assignee", "people": {"is_not_empty": true}}`
- **Checkbox**: `{"property": "Completed", "checkbox": {"equals": true}}`

## Resources

- [Full Filter Guide](../../docs/FILTER_GUIDE.md)
- [Notion Filter API Reference](https://developers.notion.com/reference/post-database-query-filter)
