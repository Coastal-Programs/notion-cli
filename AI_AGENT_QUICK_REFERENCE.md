# AI Agent Quick Reference - Simple Properties

## TLDR

Use `--simple-properties` (or `-S`) flag with flat JSON instead of nested Notion structures.

## Quick Examples

### Create Page
```bash
notion-cli page create -d DATABASE_ID -S --properties '{
  "Name": "Task Title",
  "Status": "In Progress",
  "Priority": 5,
  "Due Date": "tomorrow",
  "Tags": ["urgent", "bug"]
}'
```

### Update Page
```bash
notion-cli page update PAGE_ID -S --properties '{
  "Status": "Done",
  "Completed": true
}'
```

## Property Type Cheat Sheet

| Type | Example | Notes |
|------|---------|-------|
| **title** | `"Name": "Task"` | Simple string |
| **rich_text** | `"Description": "Text"` | Simple string |
| **number** | `"Priority": 5` | Number value |
| **checkbox** | `"Done": true` | true/false, yes/no, 1/0 |
| **select** | `"Status": "Done"` | Case-insensitive |
| **multi_select** | `"Tags": ["a", "b"]` | Array of strings |
| **status** | `"Status": "Active"` | Case-insensitive |
| **date** | `"Due": "2025-12-31"` | ISO or relative |
| **url** | `"Link": "https://..."` | Must start with http:// |
| **email** | `"Email": "a@b.com"` | Valid email format |
| **phone_number** | `"Phone": "+1-555-..."` | Any string |
| **people** | `"Users": ["user-id"]` | Array of user IDs |
| **relation** | `"Related": ["page-id"]` | Array of page IDs |
| **files** | `"Files": ["https://..."]` | Array of URLs |

## Relative Dates

| String | Result |
|--------|--------|
| `"today"` | Current date |
| `"tomorrow"` | Next day |
| `"yesterday"` | Previous day |
| `"+7 days"` | 7 days from now |
| `"-3 days"` | 3 days ago |
| `"+2 weeks"` | 2 weeks from now |
| `"+1 month"` | 1 month from now |
| `"+1 year"` | 1 year from now |

## Common Patterns

### Create with Multiple Properties
```json
{
  "Name": "Bug Fix: Login Error",
  "Status": "In Progress",
  "Priority": 8,
  "Due Date": "+3 days",
  "Tags": ["urgent", "bug"],
  "Assignee": ["user-id-123"],
  "Completed": false,
  "Description": "User cannot log in with special characters"
}
```

### Update Status and Date
```json
{
  "Status": "Done",
  "Completed": true,
  "Completion Date": "today"
}
```

### Clear a Property
```json
{
  "Description": null
}
```

## Error Handling

If you get an error, the message will tell you:
1. **What went wrong** - Invalid property, invalid value, etc.
2. **Valid options** - List of valid select options
3. **How to fix** - Suggestions for correction

Example error:
```
Error: Invalid select value: "Completed"
Valid options: Not Started, In Progress, Done
Tip: Values are case-insensitive
```

## Pro Tips

1. **Case doesn't matter** - `"name"`, `"Name"`, `"NAME"` all work
2. **Select values are case-insensitive** - `"done"`, `"Done"`, `"DONE"` all work
3. **Use relative dates** - `"tomorrow"` is easier than calculating dates
4. **Get user IDs** - Use `notion-cli user list` to get user IDs for people properties
5. **Get schema** - Use `notion-cli ds schema DB_ID` to see available properties

## Requirements

- Must use `-d DATABASE_ID` when creating pages with simple properties
- Page must be in a database when updating with simple properties
- Cannot use standalone pages (pages without a database parent)

## Don't Use Simple Properties For

- Formula properties (read-only)
- Rollup properties (read-only)
- System properties (created_time, created_by, etc.)

## Full Documentation

See `docs/SIMPLE_PROPERTIES.md` for complete documentation with all features and examples.
