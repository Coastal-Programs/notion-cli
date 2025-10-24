# Notion API Reference Documentation
## API Surface Map for Version 2025-09-03

---

## Table of Contents
1. [Overview](#overview)
2. [Authentication](#authentication)
3. [API Versioning](#api-versioning)
4. [Request/Response Formats](#requestresponse-formats)
5. [Error Handling](#error-handling)
6. [Rate Limits & Constraints](#rate-limits--constraints)
7. [Pagination](#pagination)
8. [API Endpoints](#api-endpoints)
9. [Object Types & Schemas](#object-types--schemas)
10. [Filter & Sort Options](#filter--sort-options)

---

## Overview

### Base Information
- **Base URL**: `https://api.notion.com`
- **Protocol**: HTTPS (required)
- **API Style**: RESTful
- **Data Format**: JSON

### Core Capabilities
The Notion API enables integrations to:
- Access and manage Notion pages, databases, and users
- Connect external services to Notion
- Build interactive experiences within Notion
- Query and filter database content
- Manage blocks and page content

---

## Authentication

### Authentication Method
**Bearer Token Authentication** via HTTP Authorization header

### Token Types
1. **Integration Tokens**: Generated when creating an integration in Notion
2. **OAuth Tokens**: Issued to users who complete the OAuth authorization flow

### Required Headers
```
Authorization: Bearer {ACCESS_TOKEN}
Notion-Version: 2025-09-03
```

### Authorization Header Format
```
Authorization: Bearer [ACCESS_TOKEN]
```

### Integration Capabilities
Integrations require specific capabilities enabled:
- Read content
- Update content
- Insert content
- Read comments
- Insert comments
- Read user information

Operations without proper capabilities return HTTP 403.

---

## API Versioning

### Current Version
**Latest**: `2025-09-03`

### Version Header
**Required**: All requests must include `Notion-Version` header
```
Notion-Version: 2025-09-03
```

### Versioning Policy
- New versions released only for **backwards-incompatible changes**
- New features and endpoints do NOT trigger version updates
- Existing features remain accessible without version updates

### Major Version History

#### 2025-09-03 (Latest)
- Introduced **Data Sources** as separate entities from databases
- Databases now act as containers for one or more data sources
- Split `/v1/databases` APIs into:
  - `/v1/databases` - Managing database containers
  - `/v1/data_sources` - Managing individual data sources
- Support for multi-source databases
- Data source IDs now required for property management
- Existing database IDs remain unchanged

#### 2022-06-28
- Page properties retrieved via dedicated endpoint
- Parents reference direct parents only
- Database relations use `single_property` and `dual_property` types

#### 2021-08-16
- Block children endpoint returns new blocks
- Array rollup types standardized
- Property IDs made URL-safe

#### 2021-05-13
- Rich text properties renamed from `text` to `rich_text`

---

## Request/Response Formats

### JSON Conventions

#### Property Naming
- **Snake Case**: All property names use `snake_case` (not `camelCase` or `kebab-case`)

#### Object Identification
- **Object Type**: Top-level resources have an `"object"` property (e.g., `"database"`, `"user"`, `"page"`)
- **UUID**: Top-level resources have a UUIDv4 `"id"` property
- **ID Format**: Dashes may be omitted from IDs in requests

#### Temporal Values
- **Datetimes**: ISO 8601 format with time (e.g., `2020-08-12T02:12:33.231Z`)
- **Dates**: ISO 8601 format date only (e.g., `2020-08-12`)

#### String Values
- **Empty Strings**: Not supported - use explicit `null` instead
- **Nullability**: Use `null` to unset string values like URLs

---

## Error Handling

### Error Response Format
```json
{
  "code": "error_code",
  "message": "Detailed error message"
}
```

### HTTP Status Codes

#### Success
| Code | Description |
|------|-------------|
| 200 | Request successfully processed |

#### Client Errors (4xx)
| Code | Error Code | Description |
|------|-----------|-------------|
| 400 | `invalid_json` | Request body cannot be decoded as JSON |
| 400 | `invalid_request_url` | Request URL format is invalid |
| 400 | `invalid_request` | Unsupported request type |
| 400 | `invalid_grant` | Authorization grant or refresh token invalid/expired |
| 400 | `validation_error` | Request body doesn't match schema |
| 400 | `missing_version` | Missing required `Notion-Version` header |
| 401 | `unauthorized` | Invalid bearer token |
| 403 | `restricted_resource` | Insufficient permissions |
| 404 | `object_not_found` | Resource doesn't exist or isn't shared |
| 409 | `conflict_error` | Data collision - retry with updated parameters |
| 429 | `rate_limited` | Rate limit exceeded - slow down and retry |

#### Server Errors (5xx)
| Code | Error Code | Description |
|------|-----------|-------------|
| 500 | `internal_server_error` | Unexpected server error |
| 502 | `bad_gateway` | Server connection issue |
| 503 | `service_unavailable` | Notion temporarily unavailable |
| 503 | `database_connection_unavailable` | Database unqueryable |
| 504 | `gateway_timeout` | Request timeout |

### Error Handling Best Practices
- Check for specific error codes in responses
- Implement retry logic for 409, 429, and 5xx errors
- Respect `Retry-After` header for rate limiting
- Validate data before sending to avoid 400 errors

---

## Rate Limits & Constraints

### Rate Limits
- **Request Rate**: Average of 3 requests per second per integration
- **Burst Allowance**: Some burst capacity permitted
- **Error Response**: HTTP 429 with `rate_limited` error code
- **Retry Header**: `Retry-After` header indicates wait time in seconds

### Size Limits

#### Property Value Limits
| Property Type | Limit |
|--------------|-------|
| Rich text content | 2,000 characters |
| Text URLs | 2,000 characters |
| Equations | 1,000 characters |
| Block arrays | 100 elements |
| URLs (general) | 2,000 characters |
| Email addresses | 200 characters |
| Phone numbers | 200 characters |
| Multi-select options | 100 |
| Relations | 100 related pages |
| People mentions | 100 users |

#### Payload Limits
- **Maximum block elements**: 1,000 per request
- **Maximum payload size**: 500KB
- **Database schema size**: Recommended maximum 50KB

#### Block Children Limits
- **Maximum children per request**: 100
- **Maximum nesting depth**: 2 levels per request

### Pagination Defaults
- **Default page size**: 10 items (for queries that don't specify)
- **Default page size**: 100 items (configurable)
- **Maximum page size**: 100 items

---

## Pagination

### Supported Endpoints
Cursor-based pagination is available on:
- List all users (GET)
- Retrieve block children (GET)
- Retrieve a comment (GET)
- Retrieve a page property item (GET)
- Query a database (POST)
- Query a data source (POST)
- Search (POST)

### Response Fields
```json
{
  "object": "list",
  "type": "page|block|user|database|data_source|comment|property_item",
  "results": [...],
  "has_more": true|false,
  "next_cursor": "cursor_string"
}
```

| Field | Type | Description |
|-------|------|-------------|
| `object` | string | Always `"list"` |
| `type` | string | Type of objects in results |
| `results` | array | List of results |
| `has_more` | boolean | `true` if more results available |
| `next_cursor` | string | Cursor for next page (when `has_more` is true) |

### Request Parameters
| Parameter | Type | Description |
|-----------|------|-------------|
| `page_size` | number | Number of items to return (max: 100, default: 100) |
| `start_cursor` | string | Cursor from previous response |

**Note**: Parameter location varies:
- **GET requests**: Query string parameters
- **POST requests**: Request body parameters

### Pagination Workflow
1. Send initial request to endpoint
2. Check `has_more` in response
3. If `true`, use `next_cursor` value in next request
4. Repeat until `has_more` is `false`

---

## API Endpoints

### Pages

| Method | Endpoint | Description | Capabilities Required |
|--------|----------|-------------|----------------------|
| POST | `/v1/pages` | Create a page | Insert content |
| GET | `/v1/pages/{page_id}` | Retrieve a page | Read content |
| PATCH | `/v1/pages/{page_id}` | Update page properties | Update content |
| GET | `/v1/pages/{page_id}/properties/{property_id}` | Retrieve a page property item | Read content |

**Note**: Use page property endpoint when properties exceed 25 references (affects people, relations, rich text, title fields).

### Databases (Container Level - v2025-09-03)

| Method | Endpoint | Description | Capabilities Required |
|--------|----------|-------------|----------------------|
| POST | `/v1/databases` | Create a database with initial data source | Insert content |
| GET | `/v1/databases/{database_id}` | Retrieve database container info | Read content |

**Note**: Returns database container with `data_sources` array containing IDs and names.

### Data Sources (v2025-09-03)

| Method | Endpoint | Description | Capabilities Required |
|--------|----------|-------------|----------------------|
| POST | `/v1/data_sources` | Create additional data source in database | Insert content |
| GET | `/v1/data_sources/{data_source_id}` | Retrieve data source structure | Read content |
| PATCH | `/v1/data_sources/{data_source_id}` | Update data source properties/schema | Update content |
| POST | `/v1/data_sources/{data_source_id}/query` | Query data source with filters/sorts | Read content |

### Blocks

| Method | Endpoint | Description | Capabilities Required |
|--------|----------|-------------|----------------------|
| GET | `/v1/blocks/{block_id}` | Retrieve a block | Read content |
| PATCH | `/v1/blocks/{block_id}` | Update block content | Update content |
| DELETE | `/v1/blocks/{block_id}` | Archive a block (move to trash) | Update content |
| GET | `/v1/blocks/{block_id}/children` | Retrieve block children | Read content |
| PATCH | `/v1/blocks/{block_id}/children` | Append child blocks | Insert content |

**Notes**:
- Page content accessed using page ID as `block_id`
- Maximum 100 children per append request
- Maximum 2 levels of nesting per request

### Users

| Method | Endpoint | Description | Capabilities Required |
|--------|----------|-------------|----------------------|
| GET | `/v1/users` | List all workspace users (paginated) | User information |
| GET | `/v1/users/{user_id}` | Retrieve specific user | User information |
| GET | `/v1/users/me` | Retrieve bot user information | None (any capability) |

**Note**: List users excludes guest accounts.

### Comments

| Method | Endpoint | Description | Capabilities Required |
|--------|----------|-------------|----------------------|
| POST | `/v1/comments` | Create comment on page/block/discussion | Insert comment |
| GET | `/v1/comments` | List unresolved comments | Read comment |

**Note**: Comment capabilities are disabled by default and must be enabled in integration settings.

### Search

| Method | Endpoint | Description | Capabilities Required |
|--------|----------|-------------|----------------------|
| POST | `/v1/search` | Search all shared pages and data sources | Read content |

**Parameters**:
- `query` (optional): Search term to filter by title
- `filter` (optional): Constrain to pages or data sources only

### Legacy Database Endpoints (Deprecated in v2025-09-03)

| Method | Endpoint | Status |
|--------|----------|--------|
| POST | `/v1/databases` | Deprecated - use new database/data source endpoints |
| GET | `/v1/databases/{database_id}` | Deprecated - use `/v1/databases/{database_id}` (container) or `/v1/data_sources/{data_source_id}` |
| PATCH | `/v1/databases/{database_id}` | Deprecated - use `/v1/data_sources/{data_source_id}` |
| POST | `/v1/databases/{database_id}/query` | Deprecated - use `/v1/data_sources/{data_source_id}/query` |

---

## Object Types & Schemas

### Page Object

```json
{
  "object": "page",
  "id": "uuid-string",
  "created_time": "2020-08-12T02:12:33.231Z",
  "created_by": { /* User object */ },
  "last_edited_time": "2020-08-12T02:12:33.231Z",
  "last_edited_by": { /* User object */ },
  "archived": false,
  "in_trash": false,
  "icon": { /* File or Emoji object */ },
  "cover": { /* File object */ },
  "properties": { /* Property values */ },
  "parent": { /* Parent object */ },
  "url": "https://notion.so/...",
  "public_url": "https://notion.so/..." // or null
}
```

| Field | Type | Description |
|-------|------|-------------|
| `object` | string | Always `"page"` |
| `id` | UUIDv4 | Unique page identifier |
| `created_time` | ISO 8601 | Creation timestamp |
| `created_by` | Partial User | Creator |
| `last_edited_time` | ISO 8601 | Last modification timestamp |
| `last_edited_by` | Partial User | Last editor |
| `archived` | boolean | Archive status |
| `in_trash` | boolean | Trash status |
| `icon` | File/Emoji | Page icon |
| `cover` | File | Cover image |
| `properties` | object | Property values (schema depends on parent) |
| `parent` | object | Parent information |
| `url` | string | Notion page URL |
| `public_url` | string | Published URL (or null) |

### Database Object (Container - v2025-09-03)

```json
{
  "object": "database",
  "id": "uuid-string",
  "data_sources": [
    { "id": "uuid", "name": "Data Source Name" }
  ],
  "created_time": "2020-08-12T02:12:33.231Z",
  "created_by": { /* User object */ },
  "last_edited_time": "2020-08-12T02:12:33.231Z",
  "last_edited_by": { /* User object */ },
  "title": [ /* Rich text array */ ],
  "description": [ /* Rich text array */ ],
  "icon": { /* File or Emoji object */ },
  "cover": { /* File object */ },
  "parent": { /* Parent object */ },
  "url": "https://notion.so/...",
  "archived": false,
  "in_trash": false,
  "is_inline": false,
  "public_url": "https://notion.so/..." // or null
}
```

| Field | Type | Description |
|-------|------|-------------|
| `object` | string | Always `"database"` |
| `id` | UUID | Unique identifier |
| `data_sources` | array | List of child data sources with `id` and `name` |
| `created_time` | ISO 8601 | Creation timestamp |
| `created_by` | Partial User | Creator |
| `last_edited_time` | ISO 8601 | Last modification |
| `last_edited_by` | Partial User | Last editor |
| `title` | Rich text array | Database name |
| `description` | Rich text array | Database description |
| `icon` | File/Emoji | Database icon |
| `cover` | File | Cover image |
| `parent` | object | Parent (page, workspace, etc.) |
| `url` | string | Notion database URL |
| `archived` | boolean | Archive status |
| `in_trash` | boolean | Deletion status |
| `is_inline` | boolean | Inline vs full-page display |
| `public_url` | string | Public URL if published |

### Data Source Object (v2025-09-03)

```json
{
  "object": "data_source",
  "id": "uuid-string",
  "properties": { /* Property schema */ },
  "parent": { /* Database parent */ },
  "database_parent": { /* Database's parent */ },
  "created_time": "2020-08-12T02:12:33.231Z",
  "created_by": { /* User object */ },
  "last_edited_time": "2020-08-12T02:12:33.231Z",
  "last_edited_by": { /* User object */ },
  "title": [ /* Rich text array */ ],
  "description": [ /* Rich text array */ ],
  "icon": { /* File or Emoji object */ },
  "archived": false,
  "in_trash": false
}
```

| Field | Type | Description |
|-------|------|-------------|
| `object` | string | Always `"data_source"` |
| `id` | UUID | Unique identifier |
| `properties` | object | Property schema for data source |
| `parent` | object | Parent database reference |
| `database_parent` | object | Database's parent (grandparent) |
| `created_time` | ISO 8601 | Creation timestamp |
| `created_by` | Partial User | Creator |
| `last_edited_time` | ISO 8601 | Last modification |
| `last_edited_by` | Partial User | Last editor |
| `title` | Rich text array | Data source name |
| `description` | Rich text array | Description |
| `icon` | File/Emoji | Icon |
| `archived` | boolean | Archive status |
| `in_trash` | boolean | Deletion status |

### Block Object

```json
{
  "object": "block",
  "id": "uuid-string",
  "parent": { /* Parent object */ },
  "type": "paragraph",
  "created_time": "2020-08-12T02:12:33.231Z",
  "created_by": { /* User object */ },
  "last_edited_time": "2020-08-12T02:12:33.231Z",
  "last_edited_by": { /* User object */ },
  "archived": false,
  "in_trash": false,
  "has_children": false,
  "paragraph": { /* Type-specific properties */ }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `object` | string | Always `"block"` |
| `id` | UUIDv4 | Unique block identifier |
| `parent` | object | Parent container |
| `type` | string | Block category/type |
| `created_time` | ISO 8601 | Creation timestamp |
| `created_by` | Partial User | Creator |
| `last_edited_time` | ISO 8601 | Last modification |
| `last_edited_by` | Partial User | Last editor |
| `archived` | boolean | Archive status |
| `in_trash` | boolean | Deletion status |
| `has_children` | boolean | Contains nested blocks |
| `{type}` | object | Type-specific properties |

#### Block Types (28 total)

**Content Blocks**:
- `paragraph` - Rich text with optional children and color
- `heading_1`, `heading_2`, `heading_3` - Formatted headings (toggleable)
- `quote` - Rich text with color
- `callout` - Rich text with icon and color
- `code` - Syntax-highlighted code with language and caption

**List Blocks**:
- `bulleted_list_item` - Unordered list items
- `numbered_list_item` - Ordered list items
- `to_do` - Checkbox items with checked status

**Media Blocks**:
- `image` - Image files (external/uploaded)
- `video` - Video files (external/uploaded/YouTube)
- `audio` - Audio files
- `file` - Generic files with captions
- `pdf` - PDF files with captions
- `embed` - External website embeds
- `link_preview` - Originally pasted URLs (read-only)
- `bookmark` - URLs with optional captions

**Layout Blocks**:
- `table` - Grid with rows/columns and headers
- `table_row` - Table row with cell array
- `column_list`, `column` - Multi-column layouts
- `divider` - Visual separator (no properties)

**Special Blocks**:
- `child_page` - Nested pages with title
- `child_database` - Embedded databases
- `synced_block` - Original or duplicate mirrored content
- `toggle` - Collapsible sections
- `template` - Duplication buttons (deprecated for creation)
- `table_of_contents` - Auto-generated navigation
- `breadcrumb` - Navigation breadcrumbs
- `equation` - LaTeX mathematical expressions
- `unsupported` - Unsupported block types

### User Object

```json
{
  "object": "user",
  "id": "uuid-string",
  "type": "person",
  "name": "John Doe",
  "avatar_url": "https://...",
  "person": {
    "email": "john@example.com"
  }
}
```

**All Users**:
| Field | Type | Description |
|-------|------|-------------|
| `object` | string | Always `"user"` |
| `id` | UUID | User identifier |
| `type` | string | `"person"` or `"bot"` |
| `name` | string | Display name |
| `avatar_url` | string | Avatar image URL |

**Person-Specific**:
| Field | Type | Description |
|-------|------|-------------|
| `person.email` | string | Email (requires user capability permissions) |

**Bot-Specific**:
| Field | Type | Description |
|-------|------|-------------|
| `bot.owner.type` | string | `"workspace"` or `"user"` |
| `workspace_name` | string | Owning workspace name |
| `workspace_id` | UUID | Bot's workspace ID |
| `workspace_limits.max_file_upload_size_in_bytes` | integer | Max upload size |

### Rich Text Object

```json
{
  "type": "text",
  "text": {
    "content": "Example text",
    "link": { "url": "https://..." }
  },
  "annotations": {
    "bold": false,
    "italic": false,
    "strikethrough": false,
    "underline": false,
    "code": false,
    "color": "default"
  },
  "plain_text": "Example text",
  "href": null
}
```

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `"text"`, `"mention"`, or `"equation"` |
| `text`/`mention`/`equation` | object | Type-specific configuration |
| `annotations` | object | Styling properties |
| `plain_text` | string | Unformatted text |
| `href` | string | URL for links/mentions (optional) |

**Rich Text Types**:
1. **Text**: Contains `content` and optional `link` with `url`
2. **Mention**: Six subtypes - database, date, link preview, page, template mention, user
3. **Equation**: Contains LaTeX `expression`

**Annotations**:
- `bold`, `italic`, `strikethrough`, `underline`, `code` (boolean)
- `color` (string): Standard colors + background variants

### File Object

```json
{
  "type": "external",
  "external": {
    "url": "https://..."
  }
}
```

**File Types**:
1. **Notion-hosted** (`type: "file"`):
   - `url`: Authenticated URL (expires in 1 hour)
   - `expiry_time`: Expiration timestamp

2. **API-uploaded** (`type: "file_upload"`):
   - `id`: UUID reference to uploaded file

3. **External** (`type: "external"`):
   - `url`: HTTPS link to external content (never expires)

### Emoji Object

```json
{
  "type": "emoji",
  "emoji": "😻"
}
```

**Standard Emoji**:
| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `"emoji"` |
| `emoji` | string | Emoji character |

**Custom Emoji**:
| Field | Type | Description |
|-------|------|-------------|
| `type` | string | `"custom_emoji"` |
| `custom_emoji.id` | UUID | Custom emoji identifier |
| `custom_emoji.name` | string | Display name |
| `custom_emoji.url` | string | Image URL |

### Parent Object

**Parent Types**:

1. **Database Parent**:
```json
{
  "type": "database_id",
  "database_id": "uuid"
}
```

2. **Data Source Parent**:
```json
{
  "type": "data_source_id",
  "data_source_id": "uuid",
  "database_id": "uuid"
}
```

3. **Page Parent**:
```json
{
  "type": "page_id",
  "page_id": "uuid"
}
```

4. **Workspace Parent**:
```json
{
  "type": "workspace",
  "workspace": true
}
```

5. **Block Parent**:
```json
{
  "type": "block_id",
  "block_id": "uuid"
}
```

### Comment Object

Referenced in comment endpoints with:
- Comment attachment
- Comment display name
- Comment discussion threads

### Property Schema Types (Data Source Properties)

**Simple Properties** (empty configuration):
- `checkbox` - Boolean values
- `created_by` - Auto-populated creator
- `created_time` - Auto-populated timestamp
- `date` - Date values
- `email` - Email addresses
- `files` - File uploads/links
- `last_edited_by` - Auto-populated editor
- `last_edited_time` - Auto-populated edit timestamp
- `people` - People mentions
- `phone_number` - Phone values
- `rich_text` - Text content
- `title` - Page title (required, one per data source)
- `url` - Web addresses

**Complex Properties**:
- `formula` - Contains `expression` field
- `number` - Includes `format` field (50+ format options)
- `select` - Array of `options` with id/name/color
- `multi_select` - Multiple selection options
- `status` - Contains `options` and `groups` for workflow states
- `relation` - References data source via `data_source_id`
- `rollup` - Aggregates via `function`, `relation_property_id`, `rollup_property_id`
- `unique_id` - Optional `prefix` for auto-increment

### Property Value Types (Page Properties)

**Read-Only Properties**:
- `created_by` - User object
- `created_time` - ISO 8601 timestamp
- `last_edited_by` - User object
- `last_edited_time` - ISO 8601 timestamp
- `formula` - Object with type and calculated value
- `rollup` - Object with type, function, and calculated value
- `unique_id` - Object with `number` and optional `prefix`
- `verification` - Object with `state`, `verified_by`, optional `date`

**Writable Properties**:
- `checkbox` - Boolean value
- `date` - Object with `start` (required) and `end` (optional)
- `email` - String
- `files` - Array of file objects with `name` and file reference
- `multi_select` - Array of option objects (id/name/color)
- `number` - Numeric value
- `people` - Array of user objects
- `phone_number` - String
- `relation` - Array of page references with `id` and `has_more`
- `rich_text` - Array of rich text objects
- `select` - Single option object (id/name/color)
- `status` - Option object (id/name/color)
- `title` - Array of rich text objects
- `url` - String

---

## Filter & Sort Options

### Database Query Filters

#### Supported Property Types for Filtering
- Checkbox
- Date
- Files
- Formula
- Multi-select
- Number
- People
- Phone number
- Relation
- Rich text
- Select
- Status
- Timestamp (created_time, last_edited_time)
- Verification
- ID (unique_id)

#### Filter Operators by Property Type

**Checkbox**:
- `equals`
- `does_not_equal`

**Date/Timestamp**:
- `after`
- `before`
- `equals`
- `on_or_after`
- `on_or_before`
- `is_empty`
- `is_not_empty`
- `next_month`
- `next_week`
- `next_year`
- `past_month`
- `past_week`
- `past_year`
- `this_week`

**Files**:
- `is_empty`
- `is_not_empty`

**Multi-select**:
- `contains`
- `does_not_contain`
- `is_empty`
- `is_not_empty`

**Number/ID**:
- `equals`
- `does_not_equal`
- `greater_than`
- `greater_than_or_equal_to`
- `less_than`
- `less_than_or_equal_to`
- `is_empty`
- `is_not_empty`

**People/Relation**:
- `contains`
- `does_not_contain`
- `is_empty`
- `is_not_empty`

**Rich Text**:
- `contains`
- `does_not_contain`
- `equals`
- `does_not_equal`
- `starts_with`
- `ends_with`
- `is_empty`
- `is_not_empty`

**Select/Status**:
- `equals`
- `does_not_equal`
- `is_empty`
- `is_not_empty`

**Verification**:
- `status` (values: `verified`, `expired`, `none`)

#### Compound Filters

**AND Filter**:
```json
{
  "and": [
    { "property": "Done", "checkbox": { "equals": true } },
    { "property": "Priority", "select": { "equals": "High" } }
  ]
}
```

**OR Filter**:
```json
{
  "or": [
    { "property": "Status", "select": { "equals": "In Progress" } },
    { "property": "Status", "select": { "equals": "Done" } }
  ]
}
```

**Nesting**: Supported up to 2 levels deep

### Database Query Sorts

#### Sort Types

**Property Value Sort**:
```json
{
  "property": "Name",
  "direction": "ascending"
}
```

**Timestamp Sort**:
```json
{
  "timestamp": "created_time",
  "direction": "descending"
}
```

#### Sort Directions
- `ascending` - Smallest to largest, earliest to latest
- `descending` - Largest to smallest, latest to earliest

#### Multiple Sorts
Sorts can be combined in arrays. Earlier sorts take precedence:
```json
{
  "sorts": [
    { "property": "Priority", "direction": "descending" },
    { "timestamp": "created_time", "direction": "ascending" }
  ]
}
```

---

## Key Features by Version

### Version 2025-09-03 Features
- Multi-source databases
- Separate data source management
- Data source IDs for property operations
- Enhanced database container API
- Backward compatibility with existing database IDs

### Deprecated Features (v2025-09-03)
- Direct database property management via `/v1/databases/{id}`
- Single-source database queries via `/v1/databases/{id}/query`
- Database schema updates via database endpoint

### Unsupported Features
- Creating `status` properties via API
- Creating inline discussion comments (only responses supported)
- Updating formula, status, and synced content properties
- Retrieving linked databases (must access original source)
- Filtering users by email/name
- Managing database views programmatically

---

## SDK & Code Examples

### Official SDKs
- **JavaScript SDK**: https://github.com/makenotion/notion-sdk-js
- **Open Source**: SDKs are open source projects

### Sample Request Formats
Documentation includes samples in:
- JavaScript SDK
- cURL

### Integration Setup
1. Create integration at https://www.notion.so/my-integrations
2. Obtain integration token from settings page
3. Share pages/databases with integration
4. Enable required capabilities
5. Include token in Authorization header

---

## Important Notes

### Best Practices
- Keep database schemas under 50KB for optimal performance
- Use `filter_properties` parameter to reduce response size
- Implement pagination for large datasets
- Validate data before sending to avoid validation errors
- Use dedicated property endpoint for properties exceeding 25 references
- Enable only required capabilities for security

### Common Limitations
- Cannot move existing blocks (only append)
- Cannot update parent of a page
- Cannot create status properties
- Cannot update rollup values
- Rich text properties limited to 2,000 characters
- Maximum 100 block children per append request
- Page properties conform to parent data source schema

### Data Considerations
- File URLs from Notion expire in 1 hour
- Empty strings not supported - use `null`
- All properties stored as rich text, converted based on schema
- Related databases must be shared for relation properties
- Guest users excluded from user list endpoint

---

## Version Migration Guide

### Upgrading to v2025-09-03

**Breaking Changes**:
- Database property operations now require data source ID
- Database query endpoint moved to data sources
- Database objects no longer contain `properties` field directly

**Migration Steps**:
1. Update `Notion-Version` header to `2025-09-03`
2. Retrieve data source IDs from database object's `data_sources` array
3. Update query calls from `/v1/databases/{id}/query` to `/v1/data_sources/{data_source_id}/query`
4. Update property operations to use data source endpoints
5. Test multi-source database support if applicable

**Backward Compatibility**:
- Existing database IDs remain valid
- Database retrieval endpoint still works at container level
- Single-source databases function as before

---

## Additional Resources

- Getting Started Guide: https://developers.notion.com/docs/getting-started
- Integration Settings: https://www.notion.so/my-integrations
- Authorization Guide: Referenced in authentication documentation
- Working with Comments Guide: https://developers.notion.com/docs/working-with-comments
- Working with Page Content Guide: https://developers.notion.com/docs/working-with-page-content
- Upgrading to 2025-09-03 Guide: Version-specific migration documentation

---

**Document Version**: Based on Notion API version 2025-09-03
**Last Updated**: 2025-10-22
**Base URL**: https://api.notion.com
**Documentation Source**: https://developers.notion.com/reference
