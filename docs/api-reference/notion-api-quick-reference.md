# Notion API Quick Reference Guide
## Version 2025-09-03

### Base Information
- **URL**: `https://api.notion.com`
- **Auth**: Bearer token in `Authorization` header
- **Version Header**: `Notion-Version: 2025-09-03` (required)
- **Format**: JSON (snake_case properties)
- **Rate Limit**: ~3 requests/second

---

## Complete Endpoint List

### Pages (6 endpoints)
```
POST   /v1/pages                                    # Create page
GET    /v1/pages/{page_id}                         # Get page
PATCH  /v1/pages/{page_id}                         # Update page
GET    /v1/pages/{page_id}/properties/{prop_id}   # Get page property (paginated)
```

### Databases (2 endpoints - Container level)
```
POST   /v1/databases                               # Create database
GET    /v1/databases/{database_id}                 # Get database container
```

### Data Sources (4 endpoints - NEW in v2025-09-03)
```
POST   /v1/data_sources                            # Create data source
GET    /v1/data_sources/{data_source_id}          # Get data source
PATCH  /v1/data_sources/{data_source_id}          # Update data source
POST   /v1/data_sources/{data_source_id}/query    # Query with filters/sorts
```

### Blocks (5 endpoints)
```
GET    /v1/blocks/{block_id}                      # Get block
PATCH  /v1/blocks/{block_id}                      # Update block
DELETE /v1/blocks/{block_id}                      # Archive block
GET    /v1/blocks/{block_id}/children             # Get children (paginated)
PATCH  /v1/blocks/{block_id}/children             # Append children
```

### Users (3 endpoints)
```
GET    /v1/users                                   # List users (paginated)
GET    /v1/users/{user_id}                        # Get user
GET    /v1/users/me                               # Get bot user
```

### Comments (2 endpoints)
```
POST   /v1/comments                                # Create comment
GET    /v1/comments                                # List comments (paginated)
```

### Search (1 endpoint)
```
POST   /v1/search                                  # Search pages/data sources
```

**Total Active Endpoints**: 23

---

## Object Types

### Core Objects
1. **Page** - Individual Notion pages
2. **Database** - Container for data sources
3. **Data Source** - Individual table within database (NEW in v2025-09-03)
4. **Block** - Content blocks (28 types)
5. **User** - People and bots
6. **Comment** - Page/block comments

### Supporting Objects
7. **Rich Text** - Formatted text with annotations
8. **File** - Notion-hosted, uploaded, or external files
9. **Emoji** - Standard or custom emojis
10. **Parent** - Hierarchical relationships

---

## Block Types (28 Total)

### Content
- paragraph, heading_1, heading_2, heading_3, quote, callout, code

### Lists
- bulleted_list_item, numbered_list_item, to_do, toggle

### Media
- image, video, audio, file, pdf, embed, link_preview, bookmark

### Layout
- table, table_row, column_list, column, divider

### Special
- child_page, child_database, synced_block, template, table_of_contents, breadcrumb, equation, unsupported

---

## Property Types

### Simple Properties (20)
checkbox, created_by, created_time, date, email, files, last_edited_by, last_edited_time, people, phone_number, rich_text, title, url, unique_id

### Complex Properties (6)
- **formula** - Computed values with expression
- **number** - Numeric with 50+ formats
- **select** - Single option from list
- **multi_select** - Multiple options
- **status** - Workflow states with groups
- **relation** - Links to other data sources
- **rollup** - Aggregated data from relations

---

## Filter Operators by Type

### Text (rich_text, title, email, phone_number, url)
contains, does_not_contain, equals, does_not_equal, starts_with, ends_with, is_empty, is_not_empty

### Number (number, unique_id)
equals, does_not_equal, greater_than, greater_than_or_equal_to, less_than, less_than_or_equal_to, is_empty, is_not_empty

### Date/Timestamp
after, before, equals, on_or_after, on_or_before, is_empty, is_not_empty, next_month, next_week, next_year, past_month, past_week, past_year, this_week

### Select/Status
equals, does_not_equal, is_empty, is_not_empty

### Multi-select
contains, does_not_contain, is_empty, is_not_empty

### Checkbox
equals, does_not_equal

### People/Relation
contains, does_not_contain, is_empty, is_not_empty

### Files
is_empty, is_not_empty

### Compound
- **and** - All conditions must match
- **or** - Any condition must match
- **Nesting**: Up to 2 levels

---

## Error Codes

### 4xx Client Errors
- 400: invalid_json, invalid_request_url, invalid_request, validation_error, missing_version
- 401: unauthorized
- 403: restricted_resource
- 404: object_not_found
- 409: conflict_error
- 429: rate_limited

### 5xx Server Errors
- 500: internal_server_error
- 502: bad_gateway
- 503: service_unavailable, database_connection_unavailable
- 504: gateway_timeout

---

## Limits & Constraints

### Rate Limits
- 3 requests/second average
- Burst allowance available
- HTTP 429 when exceeded

### Size Limits
| Item | Limit |
|------|-------|
| Rich text | 2,000 chars |
| Equation | 1,000 chars |
| URL | 2,000 chars |
| Email | 200 chars |
| Phone | 200 chars |
| Multi-select options | 100 |
| Relations | 100 pages |
| People | 100 users |
| Block children/request | 100 |
| Blocks total | 1,000 |
| Payload size | 500KB |
| Schema size | 50KB (recommended) |

### Pagination
- Default: 100 items
- Maximum: 100 items per page
- Cursor-based pagination

---

## Authentication Requirements

### Required Headers
```
Authorization: Bearer {token}
Notion-Version: 2025-09-03
```

### Integration Capabilities
- Read content
- Update content
- Insert content
- Read comments
- Insert comments
- Read user information

**Note**: Comment capabilities disabled by default

---

## Version History

### 2025-09-03 (Latest)
- Introduced Data Sources
- Databases as containers
- Multi-source database support
- Split database APIs

### 2022-06-28
- Dedicated page property endpoint
- Direct parent references
- New relation types

### 2021-08-16
- Block children return new blocks
- Standardized rollup types
- URL-safe property IDs

### 2021-05-13
- Renamed text to rich_text

---

## Key Differences: v2025-09-03

### Before (Pre-2025-09-03)
```
GET  /v1/databases/{id}           # Returns properties
POST /v1/databases/{id}/query     # Query database
```

### After (v2025-09-03)
```
GET  /v1/databases/{id}                    # Returns data_sources array
GET  /v1/data_sources/{data_source_id}    # Returns properties
POST /v1/data_sources/{data_source_id}/query  # Query data source
```

**Migration**: Get data source ID from database's `data_sources` array

---

## Common Patterns

### Create Page in Data Source
```json
POST /v1/pages
{
  "parent": { "data_source_id": "uuid" },
  "properties": {
    "Name": { "title": [{ "text": { "content": "Page Title" }}]}
  }
}
```

### Query with Filter & Sort
```json
POST /v1/data_sources/{id}/query
{
  "filter": {
    "property": "Status",
    "select": { "equals": "Done" }
  },
  "sorts": [
    { "property": "Priority", "direction": "descending" }
  ]
}
```

### Append Blocks
```json
PATCH /v1/blocks/{id}/children
{
  "children": [
    {
      "type": "paragraph",
      "paragraph": {
        "rich_text": [{ "text": { "content": "Text" }}]
      }
    }
  ]
}
```

---

## Best Practices

1. Use `filter_properties` to reduce response size
2. Implement pagination for large datasets
3. Keep schemas under 50KB
4. Use property endpoint for >25 references
5. Validate before sending requests
6. Handle rate limiting with exponential backoff
7. Cache file URLs (<1 hour expiration)
8. Enable minimum required capabilities

---

## Unsupported Operations

- Creating status properties
- Updating formula/rollup values
- Creating inline discussion comments
- Moving existing blocks
- Changing page parent
- Retrieving linked databases
- Filtering users by email/name
- Managing database views via API

---

## Quick Start

1. Create integration: https://www.notion.so/my-integrations
2. Get integration token
3. Share pages/databases with integration
4. Enable required capabilities
5. Make API calls with Bearer token

---

**Document Created**: 2025-10-22
**API Version**: 2025-09-03
**Status**: Current
