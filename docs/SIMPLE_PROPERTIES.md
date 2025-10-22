# Simple Properties Feature

## Overview

The `--simple-properties` (or `-S`) flag simplifies property creation and updates by allowing flat key-value mappings instead of complex nested Notion API structures.

## Problem

Creating or updating Notion properties traditionally requires deeply nested structures:

```json
{
  "Name": {
    "title": [{"text": {"content": "Task"}}]
  },
  "Status": {
    "select": {"name": "In Progress"}
  },
  "Priority": {
    "number": 5
  }
}
```

AI agents frequently get this structure wrong, leading to API errors.

## Solution

With `--simple-properties`, use flat mappings:

```json
{
  "Name": "Task",
  "Status": "In Progress",
  "Priority": 5
}
```

The CLI automatically expands these to the correct Notion format based on the database schema.

## Usage

### Creating Pages

```bash
# Simple properties (recommended for AI agents)
notion-cli page create -d DATABASE_ID -S --properties '{"Name": "My Task", "Status": "In Progress"}'

# Traditional Notion format (still supported)
notion-cli page create -d DATABASE_ID --properties '{"Name": {"title": [{"text": {"content": "My Task"}}]}}'
```

### Updating Pages

```bash
# Simple properties
notion-cli page update PAGE_ID -S --properties '{"Status": "Done", "Priority": "High"}'

# Traditional format
notion-cli page update PAGE_ID --properties '{"Status": {"select": {"name": "Done"}}}'
```

## Supported Property Types

### Basic Types

- **title**: Simple string
  ```json
  {"Name": "Task Title"}
  ```

- **rich_text**: Simple string
  ```json
  {"Description": "This is a description"}
  ```

- **number**: Number value
  ```json
  {"Priority": 5}
  ```

- **checkbox**: Boolean or string (true/false, yes/no, 1/0)
  ```json
  {"Completed": true}
  {"Completed": "yes"}
  ```

### Selection Types

- **select**: String value (case-insensitive)
  ```json
  {"Status": "In Progress"}
  {"Status": "in progress"}  // Also works
  ```

- **multi_select**: Array of strings (case-insensitive)
  ```json
  {"Tags": ["urgent", "bug"]}
  ```

- **status**: String value (case-insensitive)
  ```json
  {"Status": "Done"}
  ```

### Date Type

- **date**: ISO date or relative date
  ```json
  {"Due Date": "2025-12-31"}           // ISO date
  {"Due Date": "today"}                // Today
  {"Due Date": "tomorrow"}             // Tomorrow
  {"Due Date": "yesterday"}            // Yesterday
  {"Due Date": "+7 days"}              // 7 days from now
  {"Due Date": "-3 days"}              // 3 days ago
  {"Due Date": "+2 weeks"}             // 2 weeks from now
  {"Due Date": "+1 month"}             // 1 month from now
  {"Due Date": "+1 year"}              // 1 year from now
  ```

### Contact Types

- **email**: Email address (validated)
  ```json
  {"Email": "user@example.com"}
  ```

- **phone_number**: Phone number string
  ```json
  {"Phone": "+1-555-1234"}
  ```

- **url**: URL string (validated, must start with http:// or https://)
  ```json
  {"Website": "https://example.com"}
  ```

### Advanced Types

- **people**: Array of user IDs (UUIDs)
  ```json
  {"Assignees": ["user-id-1", "user-id-2"]}
  ```

  Note: Email addresses are NOT supported. Use `notion-cli user list` to get user IDs.

- **relation**: Array of page IDs
  ```json
  {"Related Pages": ["page-id-1", "page-id-2"]}
  ```

- **files**: Array of external file URLs
  ```json
  {"Attachments": ["https://example.com/file.pdf"]}
  ```

## Features

### Case-Insensitive Property Names

Property names are matched case-insensitively:

```bash
# All of these work
--properties '{"Name": "Task"}'
--properties '{"name": "Task"}'
--properties '{"NAME": "Task"}'
```

The actual property name from the schema is used in the output.

### Case-Insensitive Select Values

Select and multi-select values are matched case-insensitively:

```bash
# Schema has "In Progress" option
--properties '{"Status": "in progress"}'   // Works
--properties '{"Status": "IN PROGRESS"}'   // Works
--properties '{"Status": "In Progress"}'   // Works
```

The exact option name from the schema is used in the API call.

### Validation with Clear Error Messages

Invalid values provide helpful error messages:

```bash
$ notion-cli page create -d DB_ID -S --properties '{"Status": "Invalid"}'

Error: Error expanding property "Status": Invalid select value: "Invalid"
Valid options: Not Started, In Progress, Done
Tip: Values are case-insensitive
```

### Null Values

Use `null` to clear a property:

```json
{"Description": null}
```

## Examples

### Simple Task Creation

```bash
notion-cli page create -d my-tasks-db -S --properties '{
  "Name": "Fix bug in login",
  "Status": "In Progress",
  "Priority": 8,
  "Due Date": "+3 days",
  "Tags": ["urgent", "bug"],
  "Assignee": ["user-id-123"]
}'
```

### Update Multiple Properties

```bash
notion-cli page update page-123 -S --properties '{
  "Status": "Done",
  "Completed": true,
  "Due Date": "today"
}'
```

### Complex Example with All Types

```bash
notion-cli page create -d db-id -S --properties '{
  "Name": "Project Proposal",
  "Status": "In Progress",
  "Priority": 9,
  "Due Date": "+2 weeks",
  "Tags": ["important", "feature"],
  "Completed": false,
  "Email": "contact@example.com",
  "Website": "https://example.com",
  "Description": "Detailed project description here",
  "Notes": "Additional notes"
}'
```

## Requirements

- The `--simple-properties` flag requires `-d` (parent_data_source_id) for page creation
- For page updates, the page must be in a database (not a standalone page)
- The database schema is fetched automatically to validate property types

## Error Handling

Common errors and solutions:

### Property Not Found
```
Error: Property "StatusXYZ" not found in database schema.
Available properties: Name, Status, Priority, Due Date
```
**Solution**: Use the exact property name (case-insensitive) from the database.

### Invalid Select Value
```
Error: Invalid select value: "Completed"
Valid options: Not Started, In Progress, Done
```
**Solution**: Use one of the valid options from the database schema.

### Invalid Email
```
Error: Invalid email: "not-an-email"
```
**Solution**: Provide a valid email address format.

### Invalid URL
```
Error: Invalid URL: "example.com". Must start with http:// or https://
```
**Solution**: Include the protocol (http:// or https://) in the URL.

### People Property with Email
```
Error: Cannot use email addresses for people property.
Use Notion user IDs instead. You can get user IDs with: notion-cli user list
```
**Solution**: Use user IDs instead of email addresses.

## Implementation Details

The simple properties feature:

1. **Fetches database schema** to understand property types
2. **Validates values** against schema (select options, types, etc.)
3. **Expands to Notion format** with proper nested structures
4. **Preserves exact casing** from schema for consistency
5. **Provides helpful errors** when validation fails

## Comparison

### Without --simple-properties

```bash
notion-cli page create -d db-id --properties '{
  "Name": {
    "title": [{"text": {"content": "Task"}}]
  },
  "Status": {
    "select": {"name": "In Progress"}
  },
  "Priority": {
    "number": 5
  },
  "Due Date": {
    "date": {"start": "2025-12-31"}
  },
  "Tags": {
    "multi_select": [
      {"name": "urgent"},
      {"name": "bug"}
    ]
  }
}'
```

### With --simple-properties

```bash
notion-cli page create -d db-id -S --properties '{
  "Name": "Task",
  "Status": "In Progress",
  "Priority": 5,
  "Due Date": "2025-12-31",
  "Tags": ["urgent", "bug"]
}'
```

Much simpler and less error-prone!

## API Reference

### expandSimpleProperties

```typescript
import { expandSimpleProperties } from './utils/property-expander'

const simple = {
  "Name": "Task",
  "Status": "Done"
}

const schema = await retrieveDataSource(databaseId)
const expanded = await expandSimpleProperties(simple, schema.properties)
```

### validateSimpleProperties

```typescript
import { validateSimpleProperties } from './utils/property-expander'

const result = validateSimpleProperties(simple, schema.properties)

if (!result.valid) {
  console.error('Validation errors:', result.errors)
}
```

## Best Practices for AI Agents

1. **Always use --simple-properties** flag for easier property handling
2. **Use relative dates** when appropriate ("today", "tomorrow", "+7 days")
3. **Case doesn't matter** for property names and select values
4. **Check error messages** for valid options when a value is rejected
5. **Use schema command** to discover available properties: `notion-cli ds schema DB_ID`

## Future Enhancements

Potential future improvements:

- Support for formula properties (read-only currently)
- Support for rollup properties (read-only currently)
- Auto-creation of new select/multi-select options
- Email-to-user-ID resolution for people properties
- Date range support (start and end dates)
