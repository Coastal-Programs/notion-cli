# Notion API - Search and Users Endpoints Specification

## Overview
This document provides comprehensive specifications for Notion API's Search and Users endpoints based on the official API reference documentation.

API Version: `2025-09-03`

---

## 1. Search API

### 1.1 Search API Capabilities

The Search API allows searching across all pages and data sources that have been shared with an integration.

**Endpoint:** `POST /v1/search`

**Capabilities:**
- Searches all parent or child pages and data sources shared with the integration
- Returns pages or data sources with titles that include the query parameter
- If no query is provided, returns all pages/data sources shared with the integration
- Excludes duplicated linked databases from results
- Supports pagination for large result sets
- Results adhere to integration's capabilities limitations

**Important Notes:**
- To search a specific data source (not all sources), use the "Query a data source" endpoint instead
- The API only searches by title; full-text content search is not supported
- Guests' content may be included if shared with the integration

---

### 1.2 Search API Parameters

#### Request Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `query` | string | No | The text that the API compares page and data source titles against. If not provided, returns all accessible pages/data sources |
| `sort` | object | No | Criteria to order results by `direction` and `timestamp` |
| `filter` | object | No | Criteria to limit results to pages or data sources only |
| `start_cursor` | string | No | Cursor value from previous response for pagination |
| `page_size` | int32 | No | Number of items to return (max: 100, default: 100) |

#### Headers

| Header | Type | Required | Description |
|--------|------|----------|-------------|
| `Notion-Version` | string | Yes | API version to use (e.g., `2025-09-03`) |

---

### 1.3 Search Filters and Sorting

#### Sort Object

The `sort` parameter accepts an object with the following structure:

```json
{
  "direction": "ascending" | "descending",
  "timestamp": "last_edited_time"
}
```

**Properties:**
- `direction`: Sort order - `"ascending"` or `"descending"`
- `timestamp`: Currently **only** `"last_edited_time"` is supported

**Default Behavior:** If sort is not provided, the most recently edited results are returned first.

#### Filter Object

The `filter` parameter limits results to specific object types:

```json
{
  "property": "object",
  "value": "page" | "data_source"
}
```

**Properties:**
- `property`: Currently **only** `"object"` is supported
- `value`: Either `"page"` or `"data_source"` to filter by type

---

### 1.4 Search Response Structure

#### Success Response (200)

```json
{
  "object": "list",
  "results": [
    {
      "object": "page" | "database",
      "id": "string (UUID)",
      "created_time": "string (ISO 8601)",
      "last_edited_time": "string (ISO 8601)",
      "created_by": {
        "object": "user",
        "id": "string"
      },
      "last_edited_by": {
        "object": "user",
        "id": "string"
      },
      "cover": {},
      "icon": {},
      "parent": {},
      "archived": false,
      "properties": {},
      "url": "string"
    }
  ],
  "next_cursor": "string | null",
  "has_more": false,
  "type": "page_or_database",
  "page_or_database": {}
}
```

**Response Fields:**
- `object`: Always `"list"`
- `results`: Array of page or database objects
- `next_cursor`: Cursor for next page of results (null if no more results)
- `has_more`: Boolean indicating if more results are available
- `type`: Type of results returned
- `page_or_database`: Additional metadata about results

---

### 1.5 Search Pagination

The Search API supports cursor-based pagination:

1. Make initial request without `start_cursor`
2. Check `has_more` in response
3. If `true`, use `next_cursor` value in subsequent request's `start_cursor` parameter
4. Repeat until `has_more` is `false`

**Maximum page size:** 100 items per request

---

### 1.6 Search Error Responses

#### 400 Bad Request

```json
{
  "object": "error",
  "status": 400,
  "code": "string",
  "message": "string"
}
```

Returned when the request parameters are invalid.

#### 429 Rate Limited

```json
{
  "object": "error",
  "status": 429,
  "code": "string",
  "message": "string"
}
```

Returned when rate limits are exceeded.

---

### 1.7 Search Limitations

- **Title-only search**: The API only searches page and data source titles
- **No filtering by creation date or other metadata**
- **No full-text content search**
- **Results limited by integration capabilities**: Only returns content the integration has access to
- **No advanced query syntax**: Simple string matching only
- **Excludes duplicated linked databases**

---

## 2. Users API

### 2.1 User Object Structure

The User object represents a user in a Notion workspace, including:
- Full workspace members
- Guests
- Integrations (bots)

---

### 2.2 User Object - Common Properties

All user objects contain these fields:

| Property | Type | Required | Description | Example |
|----------|------|----------|-------------|---------|
| `object` | string | Yes | Always `"user"` | `"user"` |
| `id` | string (UUID) | Yes | Unique identifier for the user | `"e79a0b74-3aba-4149-9f74-0bb5791a6ee6"` |
| `type` | string | No | Type of user: `"person"` or `"bot"` | `"person"` |
| `name` | string | No | User's display name in Notion | `"Avocado Lovelace"` |
| `avatar_url` | string | No | URL to user's avatar image | `"https://secure.notion-static.com/e6a352a8.jpg"` |

**Note:** Fields marked as "No" for Required may not appear if the bot lacks proper capabilities or the user context doesn't require them.

---

### 2.3 People vs Bot Users

#### People User Objects

User objects with `type: "person"` represent actual people and include:

```json
{
  "object": "user",
  "id": "e79a0b74-3aba-4149-9f74-0bb5791a6ee6",
  "type": "person",
  "name": "Avocado Lovelace",
  "avatar_url": "https://secure.notion-static.com/e6a352a8.jpg",
  "person": {
    "email": "avo@example.org"
  }
}
```

**Person-specific Properties:**

| Property | Type | Description | Access Requirements |
|----------|------|-------------|---------------------|
| `person` | object | Container for person-specific properties | Always present for person type |
| `person.email` | string | Email address of the person | Requires user information capabilities with email access |

---

#### Bot User Objects

User objects with `type: "bot"` represent integrations and include:

```json
{
  "object": "user",
  "id": "9188c6a5-7381-452f-b3dc-d4865aa89bdf",
  "name": "Test Integration",
  "avatar_url": null,
  "type": "bot",
  "bot": {
    "owner": {
      "type": "workspace",
      "workspace": true
    },
    "workspace_name": "Ada Lovelace's Notion"
  },
  "workspace_id": "17ab3186-873d-418f-b899-c3f6a43f68de",
  "workspace_limits": {
    "max_file_upload_size_in_bytes": 5242880
  }
}
```

**Bot-specific Properties:**

| Property | Type | Description | Example |
|----------|------|-------------|---------|
| `bot` | object | Container for bot-specific properties | See above |
| `owner` | object | Information about bot ownership | `{"type": "workspace", "workspace": true}` |
| `owner.type` | string | Owner type: `"workspace"` or `"user"` | `"workspace"` |
| `workspace_name` | string | Name of workspace owning the bot (null if user-owned) | `"Ada Lovelace's Notion"` |
| `workspace_id` | string | ID of the bot's workspace | `"17ab3186-873d-418f-b899-c3f6a43f68de"` |
| `workspace_limits` | object | Workspace limits and restrictions | See below |
| `workspace_limits.max_file_upload_size_in_bytes` | integer | Maximum file upload size in bytes | `5242880` |

---

### 2.4 List All Users Endpoint

**Endpoint:** `GET /v1/users`

**Purpose:** Returns a paginated list of users for the workspace.

#### Request Parameters

| Parameter | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `start_cursor` | string | No | - | Cursor from previous response for pagination |
| `page_size` | int32 | No | 100 | Number of items to return (max: 100) |

#### Headers

| Header | Type | Required | Description |
|--------|------|----------|-------------|
| `Notion-Version` | string | Yes | API version (e.g., `2025-09-03`) |

#### Important Notes

- **Guests are NOT included** in the response
- Requires **user information capabilities** on the integration
- Returns 403 error if integration lacks required capabilities
- **No filtering by email or name** is currently supported
- Response may contain fewer than `page_size` results

---

### 2.5 List Users Response

#### Success Response (200)

```json
{
  "object": "list",
  "results": [
    {
      "object": "user",
      "id": "string",
      "type": "person",
      "name": "string",
      "avatar_url": "string",
      "person": {
        "email": "string"
      }
    }
  ],
  "next_cursor": "string | null",
  "has_more": false
}
```

#### Error Response (400)

```json
{
  "object": "error",
  "status": 400,
  "code": "string",
  "message": "string"
}
```

---

### 2.6 Retrieve a User Endpoint

**Endpoint:** `GET /v1/users/{user_id}`

**Purpose:** Retrieves a specific user by ID.

#### Path Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `user_id` | string (UUID) | Yes | The ID of the user to retrieve |

#### Headers

| Header | Type | Required | Description |
|--------|------|----------|-------------|
| `Notion-Version` | string | Yes | API version (e.g., `2025-09-03`) |

#### Requirements

- Requires **user information capabilities** on the integration
- Returns 403 error if integration lacks required capabilities

#### Response

Returns a single User object (same structure as described in section 2.2-2.3).

---

### 2.7 User Objects in Other API Contexts

User objects appear throughout the Notion API in various contexts:

#### Block Objects
- `created_by` - User who created the block
- `last_edited_by` - User who last edited the block

#### Page Objects
- `created_by` - User who created the page
- `last_edited_by` - User who last edited the page
- `people` property items - Users assigned or mentioned

#### Database Objects
- `created_by` - User who created the database
- `last_edited_by` - User who last edited the database

#### Rich Text Objects
- As user mentions (@mentions)

#### Property Objects
- When property type is `people`

**Guaranteed Fields:** User objects will **always** contain `object` and `id` keys. Other properties may or may not appear depending on:
- Bot's capabilities
- Context where user object appears
- Whether integration has user information capabilities

---

### 2.8 User Capabilities Requirements

To access user information via the Users API:

1. **Integration must have "User information" capabilities enabled**
2. **For email addresses:** Additional permission required beyond basic user info
3. **403 Forbidden** error returned if capabilities are missing

**Capability Levels:**
- **No user capabilities:** Can only see user `id` and `object` type
- **Basic user capabilities:** Can see `id`, `object`, `type`, `name`, `avatar_url`
- **Email access:** Can additionally see `person.email` for person users

---

## 3. Integration Capabilities Notes

### 3.1 Search API Capabilities
- Results filtered based on what integration has access to
- Only returns pages/databases shared with the integration
- Respects all workspace permissions and sharing settings

### 3.2 Users API Capabilities
- Requires explicit "user information" capability
- Email access requires additional permissions
- Guest users excluded from list endpoint
- Bot can retrieve information about itself via `/v1/users/me`

---

## 4. Pagination Best Practices

Both Search and Users endpoints use cursor-based pagination:

### Standard Pagination Flow

```javascript
let allResults = [];
let hasMore = true;
let startCursor = undefined;

while (hasMore) {
  const response = await notion.search({
    start_cursor: startCursor,
    page_size: 100
  });

  allResults = allResults.concat(response.results);
  hasMore = response.has_more;
  startCursor = response.next_cursor;
}
```

### Key Points
- Always check `has_more` before making next request
- Use `next_cursor` value as `start_cursor` for next request
- Maximum page size is 100 items
- Don't assume page size matches requested amount

---

## 5. Error Handling

### Common Error Codes

| Status Code | Error Type | Description | Resolution |
|-------------|------------|-------------|------------|
| 400 | Bad Request | Invalid parameters or malformed request | Check parameter values and types |
| 403 | Forbidden | Insufficient capabilities | Enable required capabilities in integration settings |
| 429 | Rate Limited | Too many requests | Implement exponential backoff retry logic |
| 500 | Internal Server Error | Notion API issue | Retry with exponential backoff |

### Error Response Structure

```json
{
  "object": "error",
  "status": 400,
  "code": "validation_error",
  "message": "Detailed error message"
}
```

---

## 6. Example Use Cases

### 6.1 Search for Specific Pages

```javascript
const response = await notion.search({
  query: "Project Plan",
  filter: {
    property: "object",
    value: "page"
  },
  sort: {
    direction: "descending",
    timestamp: "last_edited_time"
  }
});
```

### 6.2 Get All Accessible Content

```javascript
// No query parameter returns everything
const response = await notion.search({
  page_size: 100
});
```

### 6.3 List All Workspace Users

```javascript
const response = await notion.users.list({
  page_size: 100
});
```

### 6.4 Get Specific User Details

```javascript
const user = await notion.users.retrieve({
  user_id: "e79a0b74-3aba-4149-9f74-0bb5791a6ee6"
});
```

---

## 7. Limitations and Constraints

### Search API Limitations
- Title-only search (no full-text content search)
- Only `last_edited_time` supported for sorting
- No date range filtering
- No filtering by user, creation date, or custom properties
- Maximum 100 results per request

### Users API Limitations
- Cannot filter users by email or name
- Guests not included in list endpoint
- Requires user information capabilities
- Email addresses require additional permissions
- Maximum 100 users per request

---

## 8. Versioning

**Current API Version:** `2025-09-03`

All requests must include the `Notion-Version` header with a valid API version. The API version controls:
- Available features
- Response structure
- Behavior changes

Always specify the version explicitly to ensure consistent behavior.

---

## Document Information

- **Created:** 2025-10-22
- **API Version:** 2025-09-03
- **Sources:**
  - https://developers.notion.com/reference/search
  - https://developers.notion.com/reference/get-users
  - https://developers.notion.com/reference/get-user
  - https://developers.notion.com/reference/user

---

## Summary

This specification covers:

1. **Search API** - Capabilities, parameters, filtering, sorting, pagination, and limitations
2. **Search Filters and Sorting** - Detailed filter and sort object structures
3. **User Object Structure** - Common properties for all users
4. **People vs Bot Users** - Distinct properties and use cases
5. **User Listing and Retrieval** - Methods to get user information
6. **Integration Capabilities** - Required permissions and access levels
7. **Error Handling** - Common errors and resolution strategies
8. **Best Practices** - Pagination, error handling, and usage patterns
