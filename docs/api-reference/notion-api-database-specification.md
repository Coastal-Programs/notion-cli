# Notion API Database & Data Source Reference
## Complete API Specification for v2025-09-03

---

## Table of Contents
1. [Overview & Version Information](#overview--version-information)
2. [Authentication & Headers](#authentication--headers)
3. [Rate Limits & Constraints](#rate-limits--constraints)
4. [Object Schemas](#object-schemas)
5. [Database Endpoints](#database-endpoints)
6. [Data Source Endpoints](#data-source-endpoints)
7. [Property Types](#property-types)
8. [Filtering](#filtering)
9. [Sorting](#sorting)
10. [Migration Guide](#migration-guide)

---

## Overview & Version Information

### Current API Version
**Version:** `2025-09-03` (Latest as of documentation date)

### Major Changes in v2025-09-03

The API underwent a significant architectural restructuring that separates two previously unified concepts:

- **Databases**: Container objects that organize and manage data
- **Data Sources**: Individual tables within database containers

**Key Impact:**
- Enables multi-source databases (one database containing multiple data sources)
- `/v1/databases` endpoints now split into:
  - `/v1/databases` - for managing database containers
  - `/v1/data_sources` - for managing individual data sources
- Existing database IDs remain unchanged
- New data source IDs required for managing data source properties
- Backwards compatible: existing single-source database integrations continue working

**SDK Requirements:**
- JavaScript/TypeScript SDK v5.0+ requires API version `2025-09-03` or later
- SDK v5.1.0 added `is_locked` parameter support

---

## Authentication & Headers

### Required Headers

All REST API requests must include:

```http
Authorization: Bearer {integration_token}
Content-Type: application/json
Notion-Version: 2025-09-03
```

**Header Details:**
- `Authorization`: Bearer token obtained from integration settings
- `Content-Type`: Must be `application/json`
- `Notion-Version`: **REQUIRED** - Must be included in all REST API requests

### Authentication Endpoints

- **Create a token**: `POST /v1/oauth/token`
- **Introspect token**: `POST /v1/oauth/introspect`
- **Revoke token**: `POST /v1/oauth/revoke`
- **Refresh a token**: `POST /v1/oauth/refresh`

### Base URL
```
https://api.notion.com
```

---

## Rate Limits & Constraints

### Rate Limits

**Request Rate:** Average of 3 requests per second per integration
- Burst allowance permitted
- Rate-limited requests return HTTP 429 with `rate_limited` error code
- Response includes `Retry-After` header (in seconds)

**Recommended Strategies:**
- Implement exponential backoff
- Use queue-based request management
- Respect `Retry-After` header values

### Size Constraints

**Request Payload:**
- Maximum: 1000 block elements
- Maximum: 500KB overall per request

**Property Value Limits:**
- Rich text content: 2000 characters
- URLs: 2000 characters
- Equations: 1000 characters
- Email addresses: 200 characters
- Phone numbers: 200 characters
- Multi-select options: 100 items
- Relations: 100 related pages
- People mentions: 100 users
- Rich text arrays: 100 elements

**Database Schema:**
- Recommended maximum: 50KB per data source schema

**Validation:**
- Oversized requests return HTTP 400 with `validation_error` code
- Includes detailed error messaging

---

## Object Schemas

### Database Object

**Type:** `database`

```json
{
  "object": "database",
  "id": "2f26ee68-df30-4251-aad4-8ddc420cba3d",
  "data_sources": [
    {
      "id": "248104cd-477e-80af-bc30-000bd28de8f9",
      "name": "Data Source Name"
    }
  ],
  "created_time": "2020-03-17T19:10:04.968Z",
  "created_by": {
    "object": "user",
    "id": "45ee8d13-687b-47ce-a5ca-6e2e45548c4b"
  },
  "last_edited_time": "2020-03-17T21:49:37.913Z",
  "last_edited_by": {
    "object": "user",
    "id": "45ee8d13-687b-47ce-a5ca-6e2e45548c4b"
  },
  "title": [
    {
      "type": "text",
      "text": {
        "content": "Database Title"
      }
    }
  ],
  "description": [
    {
      "type": "text",
      "text": {
        "content": "Database description"
      }
    }
  ],
  "icon": {
    "type": "emoji",
    "emoji": "📊"
  },
  "cover": {
    "type": "external",
    "external": {
      "url": "https://example.com/cover.png"
    }
  },
  "parent": {
    "type": "page_id",
    "page_id": "98ad959b-2b6a-4774-80ee-00246fb0ea9b"
  },
  "url": "https://www.notion.so/database-id",
  "archived": false,
  "in_trash": false,
  "is_inline": false,
  "public_url": null
}
```

**Field Definitions:**

| Field | Type | Description |
|-------|------|-------------|
| `object` | string | Always `"database"` |
| `id` | UUID string | Unique database identifier |
| `data_sources` | array | List of child data sources with `id` and `name` |
| `created_time` | ISO 8601 string | Creation timestamp |
| `created_by` | Partial User object | Creator information |
| `last_edited_time` | ISO 8601 string | Last modification timestamp |
| `last_edited_by` | Partial User object | Last editor information |
| `title` | Rich text array | Database name |
| `description` | Rich text array | Database description |
| `icon` | File or Emoji object | Visual identifier |
| `cover` | File object | Cover image |
| `parent` | object | Parent container reference |
| `url` | string | Notion database URL |
| `archived` | boolean | Archive status |
| `in_trash` | boolean | Deletion status |
| `is_inline` | boolean | `true` if inline, `false` if full-page |
| `public_url` | string or null | Public URL if published |

---

### Data Source Object

**Type:** `data_source`

```json
{
  "object": "data_source",
  "id": "248104cd-477e-80af-bc30-000bd28de8f9",
  "properties": {
    "Title": {
      "id": "title",
      "name": "Title",
      "type": "title",
      "title": {}
    },
    "Price": {
      "id": "%3B%3F%3C",
      "name": "Price",
      "type": "number",
      "number": {
        "format": "dollar"
      }
    },
    "Last ordered": {
      "id": "aBcD",
      "name": "Last ordered",
      "type": "date",
      "date": {}
    }
  },
  "parent": {
    "type": "database_id",
    "database_id": "2f26ee68-df30-4251-aad4-8ddc420cba3d"
  },
  "database_parent": {
    "type": "page_id",
    "page_id": "98ad959b-2b6a-4774-80ee-00246fb0ea9b"
  },
  "created_time": "2020-03-17T19:10:04.968Z",
  "created_by": {
    "object": "user",
    "id": "45ee8d13-687b-47ce-a5ca-6e2e45548c4b"
  },
  "last_edited_time": "2020-03-17T21:49:37.913Z",
  "last_edited_by": {
    "object": "user",
    "id": "45ee8d13-687b-47ce-a5ca-6e2e45548c4b"
  },
  "title": [
    {
      "type": "text",
      "text": {
        "content": "Data Source Name"
      }
    }
  ],
  "description": [
    {
      "type": "text",
      "text": {
        "content": "Data source description"
      }
    }
  ],
  "icon": null,
  "archived": false,
  "in_trash": false
}
```

**Field Definitions:**

| Field | Type | Description |
|-------|------|-------------|
| `object` | string | Always `"data_source"` |
| `id` | UUID string | Unique data source identifier |
| `properties` | object | Schema defining fields/columns |
| `parent` | object | Parent database reference |
| `database_parent` | object | Database's parent (grandparent) |
| `created_time` | ISO 8601 string | Creation timestamp |
| `created_by` | Partial User object | Creator information |
| `last_edited_time` | ISO 8601 string | Last modification timestamp |
| `last_edited_by` | Partial User object | Last editor information |
| `title` | Rich text array | Data source name |
| `description` | Rich text array | Data source description |
| `icon` | File or Emoji object | Visual identifier |
| `archived` | boolean | Archive status |
| `in_trash` | boolean | Deletion status |

---

## Database Endpoints

### 1. Create a Database

**Endpoint:** `POST /v1/databases`

**Description:** Creates a database and its initial data source as a subpage within a parent page or workspace.

**Required Capabilities:** Insert content

**Request Body:**

```json
{
  "parent": {
    "type": "page_id",
    "page_id": "98ad959b-2b6a-4774-80ee-00246fb0ea9b"
  },
  "properties": {
    "Name": {
      "title": {}
    },
    "Description": {
      "rich_text": {}
    },
    "Status": {
      "select": {
        "options": [
          {
            "name": "Active",
            "color": "green"
          },
          {
            "name": "Inactive",
            "color": "red"
          }
        ]
      }
    }
  }
}
```

**Required Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `parent` | object | Must be `page_id` type or workspace reference |
| `properties` | object | Database schema for initial data source |

**Optional Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `title` | Rich text array | Database title |
| `description` | Rich text array | Database description |
| `icon` | File or Emoji object | Database icon |
| `cover` | File object | Cover image |

**Response:**

Returns a database object with:
- Database `id`
- Initial data source reference in `data_sources` array
- All database properties

**Status Codes:**
- `200`: Success
- `400`: Malformed request
- `403`: Missing insert content capabilities
- `404`: Parent page doesn't exist or integration lacks access
- `429`: Rate limit exceeded

**Limitations:**
- Creating `status` database properties is not currently supported
- Parent must be a Notion page or wiki database

**Example Request:**

```bash
curl -X POST https://api.notion.com/v1/databases \
  -H 'Authorization: Bearer secret_token' \
  -H 'Content-Type: application/json' \
  -H 'Notion-Version: 2025-09-03' \
  --data '{
    "parent": {
      "type": "page_id",
      "page_id": "98ad959b-2b6a-4774-80ee-00246fb0ea9b"
    },
    "properties": {
      "Name": {"title": {}},
      "Status": {
        "select": {
          "options": [
            {"name": "Active", "color": "green"},
            {"name": "Inactive", "color": "red"}
          ]
        }
      }
    }
  }'
```

---

### 2. Retrieve a Database

**Endpoint:** `GET /v1/databases/{database_id}`

**Description:** Retrieves a database object containing information about the database container and its data sources.

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `database_id` | UUID string | Yes | 32-character alphanumeric identifier |

**Response:**

Returns a database object including:
- Database metadata
- `data_sources` array with IDs and names of all data sources

**Status Codes:**
- `200`: Success
- `404`: Database doesn't exist or integration lacks access

**Example Request:**

```bash
curl -X GET https://api.notion.com/v1/databases/2f26ee68df304251aad48ddc420cba3d \
  -H 'Authorization: Bearer secret_token' \
  -H 'Notion-Version: 2025-09-03'
```

**Example Response:**

```json
{
  "object": "database",
  "id": "2f26ee68-df30-4251-aad4-8ddc420cba3d",
  "data_sources": [
    {
      "id": "248104cd-477e-80af-bc30-000bd28de8f9",
      "name": "Main Table"
    }
  ],
  "title": [{"type": "text", "text": {"content": "My Database"}}],
  "parent": {
    "type": "page_id",
    "page_id": "98ad959b-2b6a-4774-80ee-00246fb0ea9b"
  },
  "created_time": "2020-03-17T19:10:04.968Z",
  "last_edited_time": "2020-03-17T21:49:37.913Z",
  "url": "https://www.notion.so/database-id"
}
```

**Notes:**
- To retrieve data source structure/properties, use `GET /v1/data_sources/{data_source_id}`
- To retrieve rows, use `POST /v1/data_sources/{data_source_id}/query`
- Database ID can be found in the URL between workspace name and query parameter

---

### 3. Update a Database

**Endpoint:** `PATCH /v1/databases/{database_id}`

**Description:** Updates database attributes (title, description, icon, cover) but NOT the properties schema.

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `database_id` | UUID string | Yes | Database identifier |

**Request Body:**

```json
{
  "title": [
    {
      "type": "text",
      "text": {
        "content": "Updated Database Name"
      }
    }
  ],
  "description": [
    {
      "type": "text",
      "text": {
        "content": "Updated description"
      }
    }
  ],
  "icon": {
    "type": "emoji",
    "emoji": "🎯"
  }
}
```

**Optional Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `title` | Rich text array | Database title |
| `description` | Rich text array | Database description |
| `icon` | File or Emoji object | Database icon |
| `cover` | File object | Cover image |
| `archived` | boolean | Archive status |

**Response:**

Returns the updated database object.

**Important Notes:**
- **DOES NOT** update data source properties
- To update properties schema, use `PATCH /v1/data_sources/{data_source_id}`
- To update page values, use `PATCH /v1/pages/{page_id}`

**Example Request:**

```bash
curl -X PATCH https://api.notion.com/v1/databases/2f26ee68df304251aad48ddc420cba3d \
  -H 'Authorization: Bearer secret_token' \
  -H 'Content-Type: application/json' \
  -H 'Notion-Version: 2025-09-03' \
  --data '{
    "title": [{"type": "text", "text": {"content": "Renamed Database"}}]
  }'
```

---

## Data Source Endpoints

### 1. Create a Data Source

**Endpoint:** `POST /v1/data_sources`

**Description:** Adds an additional data source to an existing database.

**Request Body:**

```json
{
  "parent": {
    "type": "database_id",
    "database_id": "2f26ee68-df30-4251-aad4-8ddc420cba3d"
  },
  "properties": {
    "Name": {
      "title": {}
    },
    "Email": {
      "email": {}
    },
    "Phone": {
      "phone_number": {}
    }
  }
}
```

**Required Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `parent` | object | Must contain `database_id` |
| `properties` | object | Data source schema |

**Optional Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `title` | Rich text array | Data source name |
| `description` | Rich text array | Data source description |

**Response:**

Returns the created data source object.

**Behavior:**
- Automatically creates a "table" view for the new data source
- View customization must be done through Notion UI (not available via API)

**Example Request:**

```bash
curl -X POST https://api.notion.com/v1/data_sources \
  -H 'Authorization: Bearer secret_token' \
  -H 'Content-Type: application/json' \
  -H 'Notion-Version: 2025-09-03' \
  --data '{
    "parent": {
      "type": "database_id",
      "database_id": "2f26ee68-df30-4251-aad4-8ddc420cba3d"
    },
    "properties": {
      "Title": {"title": {}},
      "Status": {"select": {"options": [{"name": "Active", "color": "green"}]}}
    }
  }'
```

---

### 2. Retrieve a Data Source

**Endpoint:** `GET /v1/data_sources/{data_source_id}`

**Description:** Retrieves structure and column definitions for a specific data source.

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data_source_id` | UUID string | Yes | 32-character identifier |

**Response:**

Returns a data source object including:
- Data source metadata
- Complete `properties` schema

**Example Request:**

```bash
curl -X GET https://api.notion.com/v1/data_sources/248104cd477e80afbc30000bd28de8f9 \
  -H 'Authorization: Bearer secret_token' \
  -H 'Notion-Version: 2025-09-03'
```

**Example Response:**

```json
{
  "object": "data_source",
  "id": "248104cd-477e-80af-bc30-000bd28de8f9",
  "properties": {
    "Name": {
      "id": "title",
      "name": "Name",
      "type": "title",
      "title": {}
    },
    "Status": {
      "id": "%3B%3F%3C",
      "name": "Status",
      "type": "select",
      "select": {
        "options": [
          {"id": "1", "name": "Active", "color": "green"},
          {"id": "2", "name": "Inactive", "color": "red"}
        ]
      }
    }
  },
  "parent": {
    "type": "database_id",
    "database_id": "2f26ee68-df30-4251-aad4-8ddc420cba3d"
  },
  "title": [{"type": "text", "text": {"content": "Main Table"}}]
}
```

**Notes:**
- Related databases must be shared with integration to access relation properties
- For retrieving rows, use `POST /v1/data_sources/{data_source_id}/query`

---

### 3. Update a Data Source

**Endpoint:** `PATCH /v1/data_sources/{data_source_id}`

**Description:** Updates data source properties (schema), title, description, or trash status.

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data_source_id` | UUID string | Yes | Data source identifier |

**Request Body:**

```json
{
  "properties": {
    "New Column": {
      "rich_text": {}
    },
    "Status": {
      "select": {
        "options": [
          {"name": "Active", "color": "green"},
          {"name": "Pending", "color": "yellow"},
          {"name": "Inactive", "color": "red"}
        ]
      }
    }
  },
  "title": [
    {
      "type": "text",
      "text": {
        "content": "Updated Table Name"
      }
    }
  ]
}
```

**Optional Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `properties` | object | Schema updates (add/modify properties) |
| `title` | Rich text array | Data source title |
| `description` | Rich text array | Data source description |
| `parent` | object | Move to different database via `database_id` |
| `in_trash` | boolean | Trash status |

**Response:**

Returns the updated data source object.

**Constraints:**

**Non-Updatable Property Types:**
- `formula`
- `status`
- Synced content
- `place`

**Requirements:**
- Related databases must be shared with integration
- Schema size should not exceed 50KB (recommended)

**Example Request:**

```bash
curl -X PATCH https://api.notion.com/v1/data_sources/248104cd477e80afbc30000bd28de8f9 \
  -H 'Authorization: Bearer secret_token' \
  -H 'Content-Type: application/json' \
  -H 'Notion-Version: 2025-09-03' \
  --data '{
    "properties": {
      "Priority": {
        "select": {
          "options": [
            {"name": "High", "color": "red"},
            {"name": "Medium", "color": "yellow"},
            {"name": "Low", "color": "green"}
          ]
        }
      }
    }
  }'
```

**Notes:**
- This endpoint modifies schema/structure, not row data
- To update row values, use `PATCH /v1/pages/{page_id}`
- To add rows, use `POST /v1/pages`

---

### 4. Query a Data Source

**Endpoint:** `POST /v1/data_sources/{data_source_id}/query`

**Description:** Retrieves paginated list of pages from a data source with filtering and sorting.

**Path Parameters:**

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `data_source_id` | UUID string | Yes | Data source identifier |

**Query Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `filter_properties[]` | string array | Limit returned properties by ID (e.g., `?filter_properties[]=title&filter_properties[]=status`) |

**Request Body:**

```json
{
  "filter": {
    "and": [
      {
        "property": "Status",
        "select": {
          "equals": "Active"
        }
      },
      {
        "property": "Priority",
        "select": {
          "equals": "High"
        }
      }
    ]
  },
  "sorts": [
    {
      "property": "Created",
      "direction": "descending"
    }
  ],
  "start_cursor": null,
  "page_size": 100
}
```

**Optional Parameters:**

| Parameter | Type | Description |
|-----------|------|-------------|
| `filter` | object | Filter conditions (see Filtering section) |
| `sorts` | array | Sort criteria (see Sorting section) |
| `start_cursor` | string | Pagination cursor from previous response |
| `page_size` | number | Number of results per page |

**Response:**

```json
{
  "object": "list",
  "results": [
    {
      "object": "page",
      "id": "page-id-1",
      "properties": {
        "Name": {
          "type": "title",
          "title": [{"type": "text", "text": {"content": "Item 1"}}]
        },
        "Status": {
          "type": "select",
          "select": {"name": "Active", "color": "green"}
        }
      }
    }
  ],
  "next_cursor": "cursor-token",
  "has_more": true
}
```

**Response Fields:**

| Field | Type | Description |
|-------|------|-------------|
| `object` | string | Always `"list"` |
| `results` | array | Array of page objects |
| `next_cursor` | string or null | Token for next page, null if no more results |
| `has_more` | boolean | Whether more results exist |

**Example Request:**

```bash
curl -X POST https://api.notion.com/v1/data_sources/248104cd477e80afbc30000bd28de8f9/query \
  -H 'Authorization: Bearer secret_token' \
  -H 'Content-Type: application/json' \
  -H 'Notion-Version: 2025-09-03' \
  --data '{
    "filter": {
      "property": "Last ordered",
      "date": {"past_week": {}}
    },
    "sorts": [
      {"timestamp": "created_time", "direction": "descending"}
    ]
  }'
```

**Performance Optimization:**
- Use `filter_properties[]` query parameter to retrieve only needed properties
- Especially beneficial for databases with complex formulas, rollups, or relations

**Requirements:**
- Integration must have read content capabilities (403 without)
- Parent database must be shared with integration (404 if not)

**Limitations:**
- Formula and rollup relations exceeding 25 references evaluate only first 25
- Multi-layer relation rollups may produce inaccurate results

**For Wikis:**
- Returns data sources instead of direct database children

---

## Property Types

Data source properties define the schema and render as columns in Notion. Each property has:
- `id`: Property identifier
- `name`: Property name
- `type`: Property type
- Type-specific configuration object

### Property Types Overview

#### Basic Properties (No Configuration)

These property types have empty configuration objects `{}`:

| Type | Description |
|------|-------------|
| `checkbox` | Boolean checkbox column |
| `created_by` | Auto-populated with page creator |
| `created_time` | Auto-populated creation timestamp |
| `date` | Date value column |
| `email` | Email address column |
| `files` | File upload or external links |
| `last_edited_by` | Auto-populated with last editor |
| `last_edited_time` | Auto-populated modification timestamp |
| `people` | People mention column |
| `phone_number` | Phone number column |
| `rich_text` | Text value column |
| `title` | **Required** - One per data source, appears as page title |
| `url` | URL value column |

**Example:**

```json
{
  "Name": {
    "title": {}
  },
  "Email": {
    "email": {}
  },
  "Last Modified": {
    "last_edited_time": {}
  }
}
```

---

#### Number Property

**Configuration:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `format` | enum | Yes | Display format |

**Format Options:**
- `number`
- `number_with_commas`
- `percent`
- `dollar`
- `canadian_dollar`
- `singapore_dollar`
- `euro`
- `pound`
- `yen`
- `ruble`
- `rupee`
- `won`
- `yuan`
- `real`
- `lira`
- `rupiah`
- `franc`
- `hong_kong_dollar`
- `new_zealand_dollar`
- `krona`
- `norwegian_krone`
- `mexican_peso`
- `rand`
- `new_taiwan_dollar`
- `danish_krone`
- `zloty`
- `baht`
- `forint`
- `koruna`
- `shekel`
- `chilean_peso`
- `philippine_peso`
- `dirham`
- `colombian_peso`
- `riyal`
- `ringgit`
- `leu`
- `argentine_peso`
- `uruguayan_peso`

**Example:**

```json
{
  "Price": {
    "number": {
      "format": "dollar"
    }
  },
  "Completion": {
    "number": {
      "format": "percent"
    }
  }
}
```

---

#### Select Property

**Configuration:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `options` | array | Yes | Array of option objects |

**Option Object:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | No | Option identifier (auto-generated if omitted) |
| `name` | string | Yes | Option name (unique, no commas) |
| `color` | enum | Yes | Option color |

**Color Options:**
- `default`
- `gray`
- `brown`
- `orange`
- `yellow`
- `green`
- `blue`
- `purple`
- `pink`
- `red`

**Example:**

```json
{
  "Status": {
    "select": {
      "options": [
        {
          "name": "Not started",
          "color": "gray"
        },
        {
          "name": "In progress",
          "color": "yellow"
        },
        {
          "name": "Completed",
          "color": "green"
        }
      ]
    }
  }
}
```

---

#### Multi-Select Property

Same configuration as Select, but allows multiple values.

**Example:**

```json
{
  "Tags": {
    "multi_select": {
      "options": [
        {"name": "urgent", "color": "red"},
        {"name": "bug", "color": "orange"},
        {"name": "feature", "color": "blue"}
      ]
    }
  }
}
```

---

#### Status Property

**Configuration:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `options` | array | Yes | Array of status option objects |
| `groups` | array | Yes | Array of status group objects |

**Option Object:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | No | Option identifier |
| `name` | string | Yes | Option name |
| `color` | enum | Yes | Option color |

**Group Object:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | No | Group identifier |
| `name` | string | Yes | Group name |
| `color` | enum | Yes | Group color |
| `option_ids` | array | Yes | Array of option IDs in this group |

**Example:**

```json
{
  "Status": {
    "status": {
      "options": [
        {"id": "1", "name": "Not started", "color": "gray"},
        {"id": "2", "name": "In progress", "color": "yellow"},
        {"id": "3", "name": "Done", "color": "green"}
      ],
      "groups": [
        {
          "id": "g1",
          "name": "To do",
          "color": "gray",
          "option_ids": ["1"]
        },
        {
          "id": "g2",
          "name": "In progress",
          "color": "yellow",
          "option_ids": ["2"]
        },
        {
          "id": "g3",
          "name": "Complete",
          "color": "green",
          "option_ids": ["3"]
        }
      ]
    }
  }
}
```

**Note:** Creating new status properties via API is not currently supported.

---

#### Formula Property

**Configuration:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `expression` | string | Yes | Formula expression syntax |

**Example:**

```json
{
  "Half Price": {
    "formula": {
      "expression": "{{notion:block_property:price_property_id}}/2"
    }
  }
}
```

**Note:** Formula properties cannot be updated after creation.

---

#### Relation Property

**Configuration:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `data_source_id` | UUID string | Yes | ID of related data source |
| `synced_property_id` | string | Yes | Synced property identifier |
| `synced_property_name` | string | Yes | Synced property name |

**Example:**

```json
{
  "Related Items": {
    "relation": {
      "data_source_id": "other-data-source-id",
      "synced_property_id": "sync-id",
      "synced_property_name": "Sync Name"
    }
  }
}
```

**Requirements:**
- Related data source must be shared with integration

**Version Note:**
- v2022-06-28 introduced `single_property` and `dual_property` relation types

---

#### Rollup Property

**Configuration:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `function` | enum | Yes | Aggregation function |
| `relation_property_id` | string | Yes | Relation property to roll up from |
| `relation_property_name` | string | Yes | Relation property name |
| `rollup_property_id` | string | Yes | Property to aggregate |
| `rollup_property_name` | string | Yes | Property name to aggregate |

**Function Options:**
- `count`
- `count_values`
- `empty`
- `not_empty`
- `unique`
- `show_unique`
- `percent_empty`
- `percent_not_empty`
- `sum`
- `average`
- `median`
- `min`
- `max`
- `range`
- `earliest_date`
- `latest_date`
- `date_range`
- `checked`
- `unchecked`
- `percent_checked`
- `percent_unchecked`
- `count_per_group`
- `percent_per_group`
- `show_original`

**Example:**

```json
{
  "Total Price": {
    "rollup": {
      "function": "sum",
      "relation_property_id": "rel-id",
      "relation_property_name": "Related Items",
      "rollup_property_id": "price-id",
      "rollup_property_name": "Price"
    }
  }
}
```

**Limitations:**
- Relations exceeding 25 references evaluate only first 25
- Multi-layer rollups may produce inaccurate results

---

#### Unique ID Property

**Configuration:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `prefix` | string | No | Prefix for unique ID (creates special lookup URL) |

**Example:**

```json
{
  "ID": {
    "unique_id": {
      "prefix": "TASK-"
    }
  }
}
```

---

## Filtering

Filters operate on database properties with support for single and compound logic.

### Filter Structure

**Single Filter:**

```json
{
  "property": "PropertyName",
  "property_type": {
    "condition": "value"
  }
}
```

**Compound Filters:**

```json
{
  "and": [
    {
      "property": "Status",
      "select": {"equals": "Active"}
    },
    {
      "property": "Priority",
      "select": {"equals": "High"}
    }
  ]
}
```

```json
{
  "or": [
    {
      "property": "Tags",
      "multi_select": {"contains": "urgent"}
    },
    {
      "property": "Tags",
      "multi_select": {"contains": "important"}
    }
  ]
}
```

**Nested Compound Filters:**

```json
{
  "and": [
    {
      "property": "Status",
      "select": {"equals": "Active"}
    },
    {
      "or": [
        {"property": "Tags", "multi_select": {"contains": "A"}},
        {"property": "Tags", "multi_select": {"contains": "B"}}
      ]
    }
  ]
}
```

**Nesting Limit:** Up to 2 levels deep

---

### Filter Conditions by Property Type

#### Checkbox

```json
{
  "property": "Completed",
  "checkbox": {
    "equals": true
  }
}
```

**Operators:**
- `equals: boolean`
- `does_not_equal: boolean`

---

#### Date

```json
{
  "property": "Due Date",
  "date": {
    "on_or_after": "2025-01-01"
  }
}
```

**Operators:**
- `after: string` (ISO 8601)
- `before: string` (ISO 8601)
- `equals: string` (ISO 8601)
- `on_or_after: string` (ISO 8601)
- `on_or_before: string` (ISO 8601)
- `is_empty: true`
- `is_not_empty: true`
- `past_week: {}`
- `past_month: {}`
- `past_year: {}`
- `next_week: {}`
- `next_month: {}`
- `next_year: {}`
- `this_week: {}`

**Example - Relative Date:**

```json
{
  "property": "Created",
  "date": {
    "past_week": {}
  }
}
```

---

#### Files

```json
{
  "property": "Attachments",
  "files": {
    "is_not_empty": true
  }
}
```

**Operators:**
- `is_empty: true`
- `is_not_empty: true`

---

#### Formula

Formula filters use nested conditions matching the formula result type.

```json
{
  "property": "Formula Result",
  "formula": {
    "number": {
      "greater_than": 100
    }
  }
}
```

**Nested Types:**
- `checkbox` - Use checkbox operators
- `date` - Use date operators
- `number` - Use number operators
- `string` - Use rich_text operators

---

#### Multi-Select

```json
{
  "property": "Tags",
  "multi_select": {
    "contains": "urgent"
  }
}
```

**Operators:**
- `contains: string`
- `does_not_contain: string`
- `is_empty: true`
- `is_not_empty: true`

---

#### Number

```json
{
  "property": "Price",
  "number": {
    "greater_than": 100
  }
}
```

**Operators:**
- `equals: number`
- `does_not_equal: number`
- `greater_than: number`
- `greater_than_or_equal_to: number`
- `less_than: number`
- `less_than_or_equal_to: number`
- `is_empty: true`
- `is_not_empty: true`

---

#### People

```json
{
  "property": "Assigned To",
  "people": {
    "contains": "user-uuid"
  }
}
```

**Operators:**
- `contains: string` (UUIDv4)
- `does_not_contain: string` (UUIDv4)
- `is_empty: true`
- `is_not_empty: true`

---

#### Phone Number

```json
{
  "property": "Contact",
  "phone_number": {
    "equals": "+1234567890"
  }
}
```

**Operators:**
- `equals: string`
- `does_not_equal: string`
- `contains: string`
- `does_not_contain: string`
- `starts_with: string`
- `ends_with: string`
- `is_empty: true`
- `is_not_empty: true`

---

#### Relation

```json
{
  "property": "Related Pages",
  "relation": {
    "contains": "page-uuid"
  }
}
```

**Operators:**
- `contains: string` (UUIDv4)
- `does_not_contain: string` (UUIDv4)
- `is_empty: true`
- `is_not_empty: true`

---

#### Rich Text

```json
{
  "property": "Description",
  "rich_text": {
    "contains": "important"
  }
}
```

**Operators:**
- `equals: string`
- `does_not_equal: string`
- `contains: string`
- `does_not_contain: string`
- `starts_with: string`
- `ends_with: string`
- `is_empty: true`
- `is_not_empty: true`

---

#### Select

```json
{
  "property": "Status",
  "select": {
    "equals": "Active"
  }
}
```

**Operators:**
- `equals: string`
- `does_not_equal: string`
- `is_empty: true`
- `is_not_empty: true`

---

#### Status

```json
{
  "property": "Status",
  "status": {
    "equals": "In Progress"
  }
}
```

**Operators:**
- `equals: string`
- `does_not_equal: string`
- `is_empty: true`
- `is_not_empty: true`

---

#### Timestamp

```json
{
  "property": "Created Time",
  "timestamp": "created_time",
  "created_time": {
    "past_week": {}
  }
}
```

**Configuration:**
- `timestamp: "created_time" | "last_edited_time"`
- Uses date filter conditions

**Example:**

```json
{
  "timestamp": "last_edited_time",
  "last_edited_time": {
    "after": "2025-01-01T00:00:00.000Z"
  }
}
```

---

#### Verification

```json
{
  "property": "Email Verified",
  "verification": {
    "status": "verified"
  }
}
```

**Operators:**
- `status: "verified" | "expired" | "none"`

---

#### ID (Unique ID)

```json
{
  "property": "ID",
  "id": {
    "greater_than": 1000
  }
}
```

**Operators:**
- `equals: number`
- `does_not_equal: number`
- `greater_than: number`
- `greater_than_or_equal_to: number`
- `less_than: number`
- `less_than_or_equal_to: number`

---

## Sorting

Sorts can be applied to database properties or page timestamps.

### Sort Structure

**Property Sort:**

```json
{
  "property": "PropertyName",
  "direction": "ascending"
}
```

**Timestamp Sort:**

```json
{
  "timestamp": "created_time",
  "direction": "descending"
}
```

### Direction Options

- `ascending`
- `descending`

### Multiple Sorts

Sorts can be combined, with earlier sorts taking precedence:

```json
{
  "sorts": [
    {
      "property": "Status",
      "direction": "ascending"
    },
    {
      "property": "Priority",
      "direction": "descending"
    },
    {
      "timestamp": "created_time",
      "direction": "ascending"
    }
  ]
}
```

**Behavior:**
- First sort groups results
- Subsequent sorts order within each group
- Example: Sort by Status (ascending), then Priority (descending) groups by status, then orders by priority within each status group

### Timestamp Options

- `created_time`
- `last_edited_time`

### Complete Example

```json
{
  "filter": {
    "and": [
      {"property": "Status", "select": {"equals": "Active"}},
      {"property": "Priority", "select": {"equals": "High"}}
    ]
  },
  "sorts": [
    {"property": "Due Date", "direction": "ascending"},
    {"timestamp": "created_time", "direction": "descending"}
  ]
}
```

---

## Migration Guide

### Upgrading from Pre-2025-09-03

**Key Changes:**

1. **Database vs Data Source Separation**
   - Old: Database object contained `properties`
   - New: Database contains `data_sources` array; each data source has `properties`

2. **Endpoint Migration**

| Old Endpoint | New Endpoint |
|-------------|-------------|
| `POST /v1/databases/{id}/query` | `POST /v1/data_sources/{id}/query` |
| `PATCH /v1/databases/{id}` (properties) | `PATCH /v1/data_sources/{id}` |
| N/A | `POST /v1/data_sources` |
| N/A | `GET /v1/data_sources/{id}` |

3. **Create Database Workflow**

**Old:**
```bash
POST /v1/databases
{
  "parent": {...},
  "properties": {...}
}
```

**New:**
```bash
# Creates database AND initial data source
POST /v1/databases
{
  "parent": {...},
  "properties": {...}  # For initial data source
}

# To add more data sources:
POST /v1/data_sources
{
  "parent": {"database_id": "..."},
  "properties": {...}
}
```

4. **Query Workflow**

**Old:**
```bash
POST /v1/databases/{database_id}/query
```

**New:**
```bash
# Get database to find data source IDs
GET /v1/databases/{database_id}

# Query specific data source
POST /v1/data_sources/{data_source_id}/query
```

5. **Update Schema Workflow**

**Old:**
```bash
PATCH /v1/databases/{database_id}
{
  "properties": {...}
}
```

**New:**
```bash
# Update database attributes (title, description, icon)
PATCH /v1/databases/{database_id}
{
  "title": [...]
}

# Update data source properties (schema)
PATCH /v1/data_sources/{data_source_id}
{
  "properties": {...}
}
```

### Backwards Compatibility

- Existing database IDs remain unchanged
- Single-source databases continue working
- Old endpoints marked as deprecated but still functional
- Integration behavior unchanged for single-source databases

### Action Items

1. **Update API Version Header**
   ```
   Notion-Version: 2025-09-03
   ```

2. **Update SDK**
   - JavaScript/TypeScript: Upgrade to v5.0+

3. **Review Code**
   - Replace `/v1/databases/{id}/query` with `/v1/data_sources/{id}/query`
   - Update property management to use `/v1/data_sources/{id}`
   - Add logic to handle multiple data sources if needed

4. **Test**
   - Verify queries work with new endpoints
   - Test schema updates
   - Confirm multi-source database support if applicable

### Multi-Source Database Example

```bash
# 1. Create database with first data source
curl -X POST https://api.notion.com/v1/databases \
  -H 'Authorization: Bearer secret_token' \
  -H 'Content-Type: application/json' \
  -H 'Notion-Version: 2025-09-03' \
  --data '{
    "parent": {"type": "page_id", "page_id": "parent-id"},
    "properties": {
      "Name": {"title": {}},
      "Status": {"select": {"options": [{"name": "Active", "color": "green"}]}}
    }
  }'

# Response includes database_id and first data_source_id

# 2. Add second data source to same database
curl -X POST https://api.notion.com/v1/data_sources \
  -H 'Authorization: Bearer secret_token' \
  -H 'Content-Type: application/json' \
  -H 'Notion-Version: 2025-09-03' \
  --data '{
    "parent": {"type": "database_id", "database_id": "database-id"},
    "properties": {
      "Title": {"title": {}},
      "Category": {"select": {"options": [{"name": "Type A", "color": "blue"}]}}
    }
  }'

# 3. Query each data source independently
curl -X POST https://api.notion.com/v1/data_sources/data-source-id-1/query \
  -H 'Authorization: Bearer secret_token' \
  -H 'Content-Type: application/json' \
  -H 'Notion-Version: 2025-09-03' \
  --data '{"filter": {"property": "Status", "select": {"equals": "Active"}}}'
```

---

## Complete Examples

### Creating a Task Database

```bash
curl -X POST https://api.notion.com/v1/databases \
  -H 'Authorization: Bearer secret_YOUR_TOKEN' \
  -H 'Content-Type: application/json' \
  -H 'Notion-Version: 2025-09-03' \
  --data '{
    "parent": {
      "type": "page_id",
      "page_id": "98ad959b2b6a477480ee00246fb0ea9b"
    },
    "title": [
      {
        "type": "text",
        "text": {
          "content": "Task Manager"
        }
      }
    ],
    "icon": {
      "type": "emoji",
      "emoji": "✅"
    },
    "properties": {
      "Task Name": {
        "title": {}
      },
      "Status": {
        "select": {
          "options": [
            {"name": "Not Started", "color": "gray"},
            {"name": "In Progress", "color": "yellow"},
            {"name": "Completed", "color": "green"},
            {"name": "Blocked", "color": "red"}
          ]
        }
      },
      "Priority": {
        "select": {
          "options": [
            {"name": "Low", "color": "green"},
            {"name": "Medium", "color": "yellow"},
            {"name": "High", "color": "red"}
          ]
        }
      },
      "Due Date": {
        "date": {}
      },
      "Assigned To": {
        "people": {}
      },
      "Tags": {
        "multi_select": {
          "options": [
            {"name": "bug", "color": "red"},
            {"name": "feature", "color": "blue"},
            {"name": "documentation", "color": "purple"}
          ]
        }
      },
      "Estimated Hours": {
        "number": {
          "format": "number"
        }
      },
      "Completed": {
        "checkbox": {}
      }
    }
  }'
```

### Querying with Complex Filters

```bash
curl -X POST https://api.notion.com/v1/data_sources/248104cd477e80afbc30000bd28de8f9/query \
  -H 'Authorization: Bearer secret_YOUR_TOKEN' \
  -H 'Content-Type: application/json' \
  -H 'Notion-Version: 2025-09-03' \
  --data '{
    "filter": {
      "and": [
        {
          "property": "Status",
          "select": {
            "does_not_equal": "Completed"
          }
        },
        {
          "or": [
            {
              "property": "Priority",
              "select": {
                "equals": "High"
              }
            },
            {
              "property": "Due Date",
              "date": {
                "this_week": {}
              }
            }
          ]
        },
        {
          "property": "Assigned To",
          "people": {
            "is_not_empty": true
          }
        }
      ]
    },
    "sorts": [
      {
        "property": "Priority",
        "direction": "descending"
      },
      {
        "property": "Due Date",
        "direction": "ascending"
      }
    ],
    "page_size": 50
  }'
```

### Adding a Page to a Data Source

```bash
curl -X POST https://api.notion.com/v1/pages \
  -H 'Authorization: Bearer secret_YOUR_TOKEN' \
  -H 'Content-Type: application/json' \
  -H 'Notion-Version: 2025-09-03' \
  --data '{
    "parent": {
      "type": "data_source_id",
      "data_source_id": "248104cd477e80afbc30000bd28de8f9"
    },
    "properties": {
      "Task Name": {
        "type": "title",
        "title": [
          {
            "type": "text",
            "text": {
              "content": "Implement user authentication"
            }
          }
        ]
      },
      "Status": {
        "type": "select",
        "select": {
          "name": "In Progress"
        }
      },
      "Priority": {
        "type": "select",
        "select": {
          "name": "High"
        }
      },
      "Due Date": {
        "type": "date",
        "date": {
          "start": "2025-02-01"
        }
      },
      "Tags": {
        "type": "multi_select",
        "multi_select": [
          {"name": "feature"},
          {"name": "security"}
        ]
      },
      "Estimated Hours": {
        "type": "number",
        "number": 16
      },
      "Completed": {
        "type": "checkbox",
        "checkbox": false
      }
    }
  }'
```

---

## Error Codes

| Code | Description |
|------|-------------|
| `200` | Success |
| `400` | Bad request (malformed request, validation error) |
| `403` | Forbidden (missing capabilities) |
| `404` | Not found (resource doesn't exist or integration lacks access) |
| `429` | Rate limited (too many requests) |
| `500` | Internal server error |
| `503` | Service unavailable |

---

## Additional Resources

- [Notion API Reference](https://developers.notion.com/reference)
- [Working with Databases Guide](https://developers.notion.com/docs/working-with-databases)
- [Property Object Reference](https://developers.notion.com/reference/property-object)
- [Data Source Object Reference](https://developers.notion.com/reference/data-source)
- [Database Object Reference](https://developers.notion.com/reference/database)
- [Versioning Documentation](https://developers.notion.com/reference/versioning)
- [Status Codes](https://developers.notion.com/reference/status-codes)

---

**Document Version:** 1.0
**API Version:** 2025-09-03
**Last Updated:** 2025-10-22
