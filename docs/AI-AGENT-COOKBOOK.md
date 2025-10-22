# AI Agent Cookbook for Notion CLI

> **Practical recipes for AI agents working with Notion via CLI**

This cookbook contains battle-tested patterns for AI coding assistants (Claude, GPT, etc.) and automation scripts working with Notion. Each recipe includes complete code examples, expected outputs, and error handling.

---

## Table of Contents

1. [Quick Start: First 5 Minutes](#quick-start-first-5-minutes)
2. [Schema Discovery Pattern](#schema-discovery-pattern)
3. [Create Page from AI-Generated Content](#create-page-from-ai-generated-content)
4. [Query and Filter Database](#query-and-filter-database)
5. [Update Task Status Workflow](#update-task-status-workflow)
6. [Search and Retrieve Pattern](#search-and-retrieve-pattern)
7. [Batch Operations](#batch-operations)
8. [Error Handling and Retry](#error-handling-and-retry)
9. [Working with Markdown Files](#working-with-markdown-files)
10. [Multi-Step Automation Workflows](#multi-step-automation-workflows)
11. [Dynamic Schema Discovery](#dynamic-schema-discovery)
12. [Data Extraction and Transformation](#data-extraction-and-transformation)

---

## Quick Start: First 5 Minutes

**Goal:** Verify setup and understand basic commands.

```bash
# 1. Verify installation
notion-cli --version

# 2. Test authentication (get bot info)
notion-cli user retrieve bot --output json

# 3. Expected output:
# {
#   "success": true,
#   "data": {
#     "object": "user",
#     "id": "...",
#     "type": "bot",
#     "name": "Your Integration Name"
#   }
# }
```

**If you get an error:**
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "NOTION_TOKEN environment variable is not set..."
  }
}
```

**Fix:** Set your token:
```bash
# Mac/Linux
export NOTION_TOKEN="ntn_your_token_here"

# Windows CMD
set NOTION_TOKEN=ntn_your_token_here

# Windows PowerShell
$env:NOTION_TOKEN="ntn_your_token_here"
```

---

## Schema Discovery Pattern

**Problem:** You need to understand a database's structure before creating/updating pages.

**Solution:** Use the `db schema` command to get clean, AI-parseable property information.

### Basic Schema Discovery

```bash
# Get schema in JSON format (best for AI parsing)
notion-cli db schema <DATA_SOURCE_ID> --output json
```

**Output:**
```json
{
  "success": true,
  "data": {
    "id": "abc123...",
    "title": "Tasks",
    "properties": [
      {
        "name": "Name",
        "type": "title",
        "required": true,
        "description": "Title (required)"
      },
      {
        "name": "Status",
        "type": "select",
        "options": ["Not Started", "In Progress", "Done"],
        "description": "Select one: Not Started, In Progress, Done"
      },
      {
        "name": "Priority",
        "type": "select",
        "options": ["High", "Medium", "Low"],
        "description": "Select one: High, Medium, Low"
      },
      {
        "name": "Tags",
        "type": "multi_select",
        "options": ["urgent", "bug", "feature", "docs"],
        "description": "Select multiple: urgent, bug, feature, docs"
      },
      {
        "name": "Due Date",
        "type": "date"
      }
    ]
  }
}
```

### Extract Specific Information with jq

```bash
# Get all property names
notion-cli db schema <DATA_SOURCE_ID> --output json | jq -r '.data.properties[].name'

# Output:
# Name
# Status
# Priority
# Tags
# Due Date

# Get properties with options (select/multi-select)
notion-cli db schema <DATA_SOURCE_ID> --output json | \
  jq '.data.properties[] | select(.options) | {name, options}'

# Output:
# {
#   "name": "Status",
#   "options": ["Not Started", "In Progress", "Done"]
# }
# {
#   "name": "Priority",
#   "options": ["High", "Medium", "Low"]
# }
```

### Find Required Properties

```bash
# Get required properties only
notion-cli db schema <DATA_SOURCE_ID> --output json | \
  jq '.data.properties[] | select(.required) | .name'

# Output:
# Name
```

### Human-Readable Table View

```bash
# Get schema as formatted table
notion-cli db schema <DATA_SOURCE_ID>
```

**Output:**
```
📋 Tasks
   ID: abc123def456...
   URL: https://notion.so/...

   Name          | Type         | Req | Details
   --------------+--------------+-----+----------------------------------
   Name          | title        |  ✓  | Title (required)
   Status        | select       |     | Not Started, In Progress, Done
   Priority      | select       |     | High, Medium, Low
   Tags          | multi_select |     | urgent, bug, feature...
   Due Date      | date         |     |
```

---

## Create Page from AI-Generated Content

**Problem:** You've generated content with AI and need to save it to Notion.

**Solution:** Create pages with proper property mapping.

### Method 1: Create in Database (Table)

```bash
# Assuming you have a Tasks database
DATA_SOURCE_ID="abc123..."

# Create with inline properties (JSON format)
notion-cli page create \
  -d "$DATA_SOURCE_ID" \
  --properties '{
    "Name": {"title": [{"text": {"content": "Complete documentation"}}]},
    "Status": {"select": {"name": "In Progress"}},
    "Priority": {"select": {"name": "High"}},
    "Tags": {"multi_select": [{"name": "docs"}, {"name": "urgent"}]}
  }' \
  --output json
```

**Output:**
```json
{
  "success": true,
  "data": {
    "object": "page",
    "id": "new-page-id-123...",
    "url": "https://notion.so/...",
    "properties": { ... }
  }
}
```

### Method 2: Create from Markdown File

```bash
# Create markdown file
cat > task.md << 'EOF'
# Complete Documentation

## Overview
This task involves writing comprehensive docs for the API.

## Action Items
- [ ] Write intro section
- [ ] Add code examples
- [ ] Review with team
EOF

# Create page from file
notion-cli page create \
  -d "$DATA_SOURCE_ID" \
  -f task.md \
  --output json
```

### Method 3: Create Sub-Page (Not in Database)

```bash
# Create a page under another page
PARENT_PAGE_ID="parent-page-id..."

notion-cli page create \
  -p "$PARENT_PAGE_ID" \
  --properties '{
    "title": [{"text": {"content": "Meeting Notes - Oct 22"}}]
  }' \
  --output json
```

---

## Query and Filter Database

**Problem:** You need to find specific pages in a database.

**Solution:** Use `db query` with filters.

### Basic Query (All Pages)

```bash
notion-cli db query <DATA_SOURCE_ID> --output json
```

### Filter by Single Property

```bash
# Get only "Done" tasks
notion-cli db query <DATA_SOURCE_ID> \
  --filter '{
    "property": "Status",
    "select": {
      "equals": "Done"
    }
  }' \
  --output json
```

### Filter by Multiple Conditions (AND)

```bash
# Get high-priority tasks that are in progress
notion-cli db query <DATA_SOURCE_ID> \
  --filter '{
    "and": [
      {
        "property": "Priority",
        "select": {"equals": "High"}
      },
      {
        "property": "Status",
        "select": {"equals": "In Progress"}
      }
    ]
  }' \
  --output json
```

### Filter with OR Logic

```bash
# Get tasks that are either high priority OR urgent
notion-cli db query <DATA_SOURCE_ID> \
  --filter '{
    "or": [
      {
        "property": "Priority",
        "select": {"equals": "High"}
      },
      {
        "property": "Tags",
        "multi_select": {"contains": "urgent"}
      }
    ]
  }' \
  --output json
```

### Sort Results

```bash
# Get tasks sorted by due date (ascending)
notion-cli db query <DATA_SOURCE_ID> \
  --sorts '[
    {
      "property": "Due Date",
      "direction": "ascending"
    }
  ]' \
  --output json
```

### Extract Specific Fields with jq

```bash
# Get just page IDs and titles
notion-cli db query <DATA_SOURCE_ID> --output json | \
  jq '.data.results[] | {
    id: .id,
    title: .properties.Name.title[0].plain_text
  }'
```

---

## Update Task Status Workflow

**Problem:** Common workflow - mark a task as complete.

**Solution:** Query, extract ID, update status.

### Complete Workflow

```bash
#!/bin/bash
# update-task-status.sh

DATA_SOURCE_ID="your-database-id"
TASK_NAME="Complete documentation"
NEW_STATUS="Done"

# Step 1: Find the task
echo "Finding task: $TASK_NAME..."
RESULT=$(notion-cli db query "$DATA_SOURCE_ID" \
  --filter "{
    \"property\": \"Name\",
    \"title\": {
      \"equals\": \"$TASK_NAME\"
    }
  }" \
  --output json)

# Step 2: Extract page ID
PAGE_ID=$(echo "$RESULT" | jq -r '.data.results[0].id')

if [ "$PAGE_ID" = "null" ]; then
  echo "Error: Task not found"
  exit 1
fi

echo "Found task with ID: $PAGE_ID"

# Step 3: Update status
echo "Updating status to: $NEW_STATUS..."
notion-cli page update "$PAGE_ID" \
  --properties "{
    \"Status\": {
      \"select\": {\"name\": \"$NEW_STATUS\"}
    }
  }" \
  --output json

echo "✓ Task updated successfully"
```

### Archive Completed Tasks

```bash
#!/bin/bash
# archive-completed-tasks.sh

DATA_SOURCE_ID="your-database-id"

# Find all done tasks
DONE_TASKS=$(notion-cli db query "$DATA_SOURCE_ID" \
  --filter '{
    "property": "Status",
    "select": {"equals": "Done"}
  }' \
  --output json)

# Extract page IDs
PAGE_IDS=$(echo "$DONE_TASKS" | jq -r '.data.results[].id')

# Archive each
for PAGE_ID in $PAGE_IDS; do
  echo "Archiving page: $PAGE_ID"
  notion-cli page update "$PAGE_ID" --archive --output json
done

echo "✓ All completed tasks archived"
```

---

## Search and Retrieve Pattern

**Problem:** You need to find pages by title across your workspace.

**Solution:** Use search, then retrieve full details.

```bash
# Search for pages
notion-cli search --query "meeting notes" --output json | \
  jq '.data.results[] | {id: .id, title: .properties.title.title[0].plain_text}'

# Get specific page details
PAGE_ID="found-page-id..."
notion-cli page retrieve "$PAGE_ID" --output json
```

---

## Batch Operations

**Problem:** You need to create/update multiple pages efficiently.

**Solution:** Loop with error handling.

### Batch Create Pages

```bash
#!/bin/bash
# batch-create-tasks.sh

DATA_SOURCE_ID="your-database-id"

# Array of tasks to create
TASKS=(
  "Review pull request #123"
  "Update dependencies"
  "Write tests for auth module"
  "Deploy to staging"
)

for TASK in "${TASKS[@]}"; do
  echo "Creating task: $TASK"

  notion-cli page create \
    -d "$DATA_SOURCE_ID" \
    --properties "{
      \"Name\": {\"title\": [{\"text\": {\"content\": \"$TASK\"}}]},
      \"Status\": {\"select\": {\"name\": \"Not Started\"}},
      \"Priority\": {\"select\": {\"name\": \"Medium\"}}
    }" \
    --output json > /dev/null

  if [ $? -eq 0 ]; then
    echo "  ✓ Created"
  else
    echo "  ✗ Failed"
  fi
done
```

### Batch Update from CSV

```bash
#!/bin/bash
# batch-update-from-csv.sh

# tasks.csv format:
# id,status,priority
# page-id-1,Done,High
# page-id-2,In Progress,Medium

while IFS=',' read -r page_id status priority; do
  [ "$page_id" = "id" ] && continue  # Skip header

  echo "Updating page: $page_id"
  notion-cli page update "$page_id" \
    --properties "{
      \"Status\": {\"select\": {\"name\": \"$status\"}},
      \"Priority\": {\"select\": {\"name\": \"$priority\"}}
    }" \
    --output json
done < tasks.csv
```

---

## Error Handling and Retry

**Problem:** API calls can fail - you need robust error handling.

**Solution:** The CLI has built-in retry logic, but you should still handle errors.

### Check Exit Codes

```bash
#!/bin/bash

notion-cli page retrieve "$PAGE_ID" --output json

if [ $? -eq 0 ]; then
  echo "Success"
else
  echo "Failed - check the JSON error output"
  exit 1
fi
```

### Parse Error Response

```bash
#!/bin/bash

RESULT=$(notion-cli page retrieve "invalid-id" --output json 2>&1)

# Check if operation succeeded
if echo "$RESULT" | jq -e '.success' > /dev/null; then
  echo "Operation succeeded"
  # Process data
  echo "$RESULT" | jq '.data'
else
  echo "Operation failed"
  # Extract error details
  ERROR_CODE=$(echo "$RESULT" | jq -r '.error.code')
  ERROR_MSG=$(echo "$RESULT" | jq -r '.error.message')

  echo "Error code: $ERROR_CODE"
  echo "Error message: $ERROR_MSG"

  # Handle specific errors
  case "$ERROR_CODE" in
    "UNAUTHORIZED")
      echo "Check your NOTION_TOKEN"
      ;;
    "NOT_FOUND")
      echo "Resource doesn't exist or isn't shared with integration"
      ;;
    "RATE_LIMITED")
      echo "Rate limited - retry after a delay"
      ;;
  esac
fi
```

### Retry with Exponential Backoff

The CLI has built-in retry, but here's how to add your own:

```bash
#!/bin/bash
# retry-wrapper.sh

MAX_RETRIES=5
RETRY_COUNT=0
DELAY=1

while [ $RETRY_COUNT -lt $MAX_RETRIES ]; do
  notion-cli db query "$DATA_SOURCE_ID" --output json > result.json

  if [ $? -eq 0 ] && jq -e '.success' result.json > /dev/null; then
    echo "Success!"
    cat result.json
    exit 0
  fi

  RETRY_COUNT=$((RETRY_COUNT + 1))
  echo "Attempt $RETRY_COUNT failed, retrying in ${DELAY}s..."
  sleep $DELAY
  DELAY=$((DELAY * 2))  # Exponential backoff
done

echo "Failed after $MAX_RETRIES attempts"
exit 1
```

---

## Working with Markdown Files

**Problem:** You generate content in Markdown and need to convert it to Notion blocks.

**Solution:** The CLI automatically converts Markdown to Notion blocks.

### Create Page from Markdown

```bash
# Generate markdown content
cat > article.md << 'EOF'
# Building APIs with Node.js

## Introduction
This guide covers best practices for API development.

## Key Concepts
- RESTful design
- Authentication
- Error handling

## Code Example
```javascript
app.get('/api/users', (req, res) => {
  res.json({ users: [] })
})
```

## Conclusion
Follow these patterns for maintainable APIs.
EOF

# Create page from markdown
notion-cli page create \
  -d "$DATA_SOURCE_ID" \
  -f article.md \
  --output json
```

### Update Page Content

```bash
# Get page blocks (to understand structure)
notion-cli block retrieve "$PAGE_ID" --output json

# Append new content as markdown
cat > addition.md << 'EOF'
## Update Log
- Added new section
- Fixed typos
EOF

# Note: Full page content update requires block manipulation
# For now, use append:
notion-cli block append \
  -b "$PAGE_ID" \
  -c '[{"paragraph": {"rich_text": [{"text": {"content": "New paragraph"}}]}}]' \
  --output json
```

---

## Multi-Step Automation Workflows

**Problem:** Complex workflows require multiple API calls in sequence.

**Solution:** Chain commands with proper error handling.

### Workflow: Create Task from GitHub Issue

```bash
#!/bin/bash
# github-issue-to-notion.sh

GITHUB_ISSUE_URL="$1"
DATA_SOURCE_ID="your-tasks-db-id"

# Extract issue details (pseudo-code - use GitHub API)
ISSUE_TITLE=$(curl -s "$GITHUB_ISSUE_URL" | jq -r '.title')
ISSUE_BODY=$(curl -s "$GITHUB_ISSUE_URL" | jq -r '.body')
ISSUE_LABELS=$(curl -s "$GITHUB_ISSUE_URL" | jq -r '.labels[].name' | tr '\n' ',')

# Step 1: Discover schema to ensure labels are valid
echo "Checking database schema..."
SCHEMA=$(notion-cli db schema "$DATA_SOURCE_ID" --output json)
VALID_TAGS=$(echo "$SCHEMA" | jq -r '.data.properties[] | select(.name=="Tags") | .options[]')

# Step 2: Create page in Notion
echo "Creating task in Notion..."
RESULT=$(notion-cli page create \
  -d "$DATA_SOURCE_ID" \
  --properties "{
    \"Name\": {\"title\": [{\"text\": {\"content\": \"$ISSUE_TITLE\"}}]},
    \"Status\": {\"select\": {\"name\": \"Not Started\"}},
    \"Priority\": {\"select\": {\"name\": \"Medium\"}}
  }" \
  --output json)

# Step 3: Get created page ID
PAGE_ID=$(echo "$RESULT" | jq -r '.data.id')

# Step 4: Add issue body as page content
cat > temp_content.md << EOF
# GitHub Issue

$ISSUE_BODY

---
[View original issue]($GITHUB_ISSUE_URL)
EOF

notion-cli block append \
  -b "$PAGE_ID" \
  -f temp_content.md \
  --output json

echo "✓ Task created: $PAGE_ID"
rm temp_content.md
```

### Workflow: Daily Status Report

```bash
#!/bin/bash
# daily-status-report.sh

DATA_SOURCE_ID="tasks-db-id"
REPORT_PAGE_ID="report-page-id"

# Get tasks completed today
TODAY=$(date +%Y-%m-%d)

COMPLETED_TASKS=$(notion-cli db query "$DATA_SOURCE_ID" \
  --filter "{
    \"and\": [
      {\"property\": \"Status\", \"select\": {\"equals\": \"Done\"}},
      {\"property\": \"Last Edited\", \"date\": {\"on_or_after\": \"$TODAY\"}}
    ]
  }" \
  --output json)

# Format as markdown
TASK_LIST=$(echo "$COMPLETED_TASKS" | \
  jq -r '.data.results[] | "- " + .properties.Name.title[0].plain_text')

# Create report
cat > report.md << EOF
# Daily Status Report - $(date +%Y-%m-%d)

## Completed Tasks
$TASK_LIST

## Stats
- Total completed: $(echo "$COMPLETED_TASKS" | jq '.data.results | length')
EOF

# Append to report page
notion-cli block append -b "$REPORT_PAGE_ID" -f report.md --output json

echo "✓ Daily report generated"
rm report.md
```

---

## Dynamic Schema Discovery

**Problem:** You're working with databases you've never seen before.

**Solution:** Always discover schema first, then adapt your automation.

```bash
#!/bin/bash
# smart-create.sh - Discovers schema and creates page dynamically

DATA_SOURCE_ID="$1"

# Step 1: Get schema
echo "Discovering schema..."
SCHEMA=$(notion-cli db schema "$DATA_SOURCE_ID" --output json)

# Step 2: Find title property name (might not be "Name")
TITLE_PROP=$(echo "$SCHEMA" | jq -r '.data.properties[] | select(.type=="title") | .name')

# Step 3: Find status property (if exists)
STATUS_PROP=$(echo "$SCHEMA" | jq -r '.data.properties[] | select(.type=="select" and .name=="Status") | .name')
DEFAULT_STATUS=$(echo "$SCHEMA" | jq -r ".data.properties[] | select(.name==\"$STATUS_PROP\") | .options[0]")

# Step 4: Build properties dynamically
PROPERTIES="{\"$TITLE_PROP\": {\"title\": [{\"text\": {\"content\": \"New Item\"}}]}}"

if [ "$STATUS_PROP" != "null" ] && [ "$DEFAULT_STATUS" != "null" ]; then
  PROPERTIES=$(echo "$PROPERTIES" | jq ". + {\"$STATUS_PROP\": {\"select\": {\"name\": \"$DEFAULT_STATUS\"}}}")
fi

# Step 5: Create page
echo "Creating page with discovered schema..."
notion-cli page create -d "$DATA_SOURCE_ID" --properties "$PROPERTIES" --output json
```

---

## Data Extraction and Transformation

**Problem:** Extract data from Notion for analysis or sync to other tools.

**Solution:** Query + jq for powerful data transformation.

### Extract to CSV

```bash
#!/bin/bash
# export-to-csv.sh

DATA_SOURCE_ID="$1"

# Get all pages
DATA=$(notion-cli db query "$DATA_SOURCE_ID" --output json)

# Convert to CSV
echo "Name,Status,Priority,Tags" > export.csv

echo "$DATA" | jq -r '.data.results[] | [
  .properties.Name.title[0].plain_text,
  .properties.Status.select.name,
  .properties.Priority.select.name,
  (.properties.Tags.multi_select | map(.name) | join(";"))
] | @csv' >> export.csv

echo "✓ Exported to export.csv"
```

### Sync to External System

```bash
#!/bin/bash
# sync-to-external.sh

DATA_SOURCE_ID="notion-db-id"
EXTERNAL_API="https://api.example.com/tasks"

# Get all Notion tasks
TASKS=$(notion-cli db query "$DATA_SOURCE_ID" --output json)

# Transform and send to external API
echo "$TASKS" | jq -c '.data.results[]' | while read task; do
  EXTERNAL_FORMAT=$(echo "$task" | jq '{
    external_id: .id,
    title: .properties.Name.title[0].plain_text,
    status: .properties.Status.select.name,
    url: .url
  }')

  curl -X POST "$EXTERNAL_API" \
    -H "Content-Type: application/json" \
    -d "$EXTERNAL_FORMAT"
done

echo "✓ Sync complete"
```

---

## Best Practices for AI Agents

### 1. Always Use --output json Flag

Makes parsing predictable and reliable.

```bash
# Good
notion-cli db query "$ID" --output json | jq '.data.results'

# Avoid (table output varies)
notion-cli db query "$ID"
```

### 2. Discover Schema Before Creating/Updating

Don't assume property names or types.

```bash
# Always do this first
notion-cli db schema "$DATA_SOURCE_ID" --output json
```

### 3. Check Success Before Processing

```bash
RESULT=$(notion-cli page create ... --output json)

if echo "$RESULT" | jq -e '.success' > /dev/null; then
  # Process data
else
  # Handle error
fi
```

### 4. Use Environment Variables for IDs

```bash
# .env
export TASKS_DB_ID="abc123..."
export NOTES_DB_ID="def456..."

# script.sh
notion-cli db query "$TASKS_DB_ID" --output json
```

### 5. Cache Schema Queries

The CLI caches automatically, but you can too:

```bash
# Cache schema for 10 minutes
if [ ! -f schema.json ] || [ $(find schema.json -mmin +10) ]; then
  notion-cli db schema "$DATA_SOURCE_ID" --output json > schema.json
fi

# Use cached schema
SCHEMA=$(cat schema.json)
```

---

## Common Pitfalls and Solutions

### Pitfall 1: Assuming Property Names

**Wrong:**
```bash
# Assumes property is named "Name"
--properties '{"Name": {...}}'
```

**Right:**
```bash
# Discover first
TITLE_PROP=$(notion-cli db schema "$ID" --output json | \
  jq -r '.data.properties[] | select(.type=="title") | .name')

--properties "{\"$TITLE_PROP\": {...}}"
```

### Pitfall 2: Not Handling Pagination

**Wrong:**
```bash
# Only gets first 100 results
notion-cli db query "$ID" --output json
```

**Right:**
```bash
# The CLI handles pagination automatically in db query
# But always check has_more if you're concerned:
RESULT=$(notion-cli db query "$ID" --output json)
HAS_MORE=$(echo "$RESULT" | jq -r '.data.has_more')
```

### Pitfall 3: Ignoring Error Responses

**Wrong:**
```bash
notion-cli page create ... --output json
# Assumes it worked
```

**Right:**
```bash
RESULT=$(notion-cli page create ... --output json)
if ! echo "$RESULT" | jq -e '.success' > /dev/null; then
  echo "Error:" $(echo "$RESULT" | jq -r '.error.message')
  exit 1
fi
```

---

## Performance Tips

1. **Use caching:** The CLI caches schemas automatically (10 min default)
2. **Batch when possible:** Group similar operations to reduce API calls
3. **Filter at query time:** Use `--filter` instead of filtering results
4. **Request only what you need:** Use `--properties` flag on schema command
5. **Parallel operations:** Run independent operations in parallel (bash `&`)

---

## Resources

- [Notion API Documentation](https://developers.notion.com/)
- [jq Manual](https://stedolan.github.io/jq/manual/)
- [Notion Filter Reference](https://developers.notion.com/reference/post-database-query-filter)
- [CLI Help](../README.md)

---

## Contributing

Found a useful pattern? Submit a PR to add it to this cookbook!

**Format:**
```markdown
## Pattern Name

**Problem:** Clear description of the problem

**Solution:** High-level approach

### Code Example
```bash
# Complete, working example
```

**Output:**
```json
// Expected output
```
```

---

**Last Updated:** 2025-10-22
**Version:** 1.0.0 (Initial Release)
