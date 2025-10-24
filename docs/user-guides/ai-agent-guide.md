# AI Agent Guide for Notion CLI

> **Complete guide for AI coding assistants (Claude, GPT, etc.) working with Notion CLI**

This guide provides comprehensive instructions for AI agents to effectively use the Notion CLI for automation, task management, and Notion workspace operations.

---

## Table of Contents

1. [Quick Start](#quick-start)
2. [First-Time Setup](#first-time-setup)
3. [Health Checks & Diagnostics](#health-checks--diagnostics)
4. [Core Workflow](#core-workflow)
5. [Simple Properties Mode](#simple-properties-mode)
6. [Database Operations](#database-operations)
7. [Troubleshooting](#troubleshooting)
8. [Best Practices](#best-practices)

---

## Quick Start

### Installation

```bash
# Mac/Linux: Install from GitHub
npm install -g Coastal-Programs/notion-cli

# Windows: Use local install (GitHub has symlink issues)
git clone https://github.com/Coastal-Programs/notion-cli
cd notion-cli
npm install -g .
```

### Verify Installation

```bash
notion-cli --version
```

---

## First-Time Setup

### Option 1: Interactive Setup Wizard (Recommended)

The `init` command guides you through the complete setup process:

```bash
notion-cli init

# Step 1: Token Setup
#   - Prompts for Notion API token
#   - Tests token validity
#   - Saves to environment/config

# Step 2: Connection Test
#   - Verifies API connectivity
#   - Shows workspace information
#   - Confirms bot permissions

# Step 3: Workspace Sync
#   - Offers to cache all databases
#   - Builds local database index
#   - Enables name-based lookups
```

**Automation Mode:**
```bash
# For CI/CD and automated environments
notion-cli init --json
```

### Option 2: Manual Token Setup

```bash
# Mac/Linux
export NOTION_TOKEN="secret_your_token_here"

# Windows (Command Prompt)
set NOTION_TOKEN=secret_your_token_here

# Windows (PowerShell)
$env:NOTION_TOKEN="secret_your_token_here"

# Or use the config command
notion-cli config set-token
```

**Get your API token**: https://developers.notion.com/docs/create-a-notion-integration

---

## Health Checks & Diagnostics

### Doctor Command (Comprehensive Health Check)

Run comprehensive diagnostics to verify your setup:

```bash
notion-cli doctor

# Checks performed:
# ✓ Node.js version compatibility
# ✓ Token configuration
# ✓ API connectivity
# ✓ Workspace access
# ✓ Cache status
# ✓ Dependencies
# ✓ File permissions

# JSON output for automation
notion-cli doctor --json
```

**Aliases:**
- `notion-cli diagnose`
- `notion-cli healthcheck`

**Use cases:**
- First-time setup verification
- Troubleshooting errors
- Pre-flight checks before automation
- CI/CD health monitoring

### WhoAmI Command (Quick Connection Test)

Test API connectivity and get bot information:

```bash
notion-cli whoami

# Returns:
# - Bot info (name, ID, type)
# - Workspace details
# - Cache status
# - API latency
```

**Aliases:**
- `notion-cli test`
- `notion-cli health`

---

## Core Workflow

### Recommended AI Agent Workflow

**1. Verify Setup (First Run)**
```bash
# Run health check
notion-cli doctor

# If any checks fail, init command will guide you
notion-cli init
```

**2. Sync Workspace (One-Time Setup)**
```bash
# Cache all databases for faster lookups
notion-cli sync

# View cached databases
notion-cli list --json
```

**3. Discover Database Schema**
```bash
# Get schema before creating/updating pages
notion-cli db schema <DATA_SOURCE_ID> --with-examples --json

# Filter specific properties
notion-cli db schema <ID> --properties Name,Status,Priority --json
```

**4. Create/Update Pages**
```bash
# Use simple properties mode for easier syntax
notion-cli page create -d <DB_ID> -S --properties '{
  "Name": "Task Title",
  "Status": "In Progress",
  "Priority": 5,
  "Due Date": "tomorrow"
}'

# Update existing page
notion-cli page update <PAGE_ID> -S --properties '{
  "Status": "Done",
  "Completed": true
}'
```

**5. Query Databases**
```bash
# Query with filters
notion-cli db query <ID> --filter '{
  "property": "Status",
  "select": {"equals": "In Progress"}
}' --json

# Search text
notion-cli db query <ID> --search "urgent" --json
```

---

## Simple Properties Mode

### Overview

The `-S` or `--simple-properties` flag enables flat JSON syntax instead of complex Notion API structures.

**70% complexity reduction** - Perfect for AI agents!

### Property Type Reference

| Type | Simple Format | Example |
|------|---------------|---------|
| **title** | String | `"Name": "Task"` |
| **rich_text** | String | `"Description": "Details"` |
| **number** | Number | `"Priority": 5` |
| **checkbox** | Boolean | `"Done": true` |
| **select** | String | `"Status": "In Progress"` |
| **multi_select** | Array | `"Tags": ["urgent", "bug"]` |
| **status** | String | `"Status": "Active"` |
| **date** | String/Object | `"Due": "2025-12-31"` |
| **url** | String | `"Link": "https://..."` |
| **email** | String | `"Email": "user@example.com"` |
| **phone_number** | String | `"Phone": "+1-555-..."` |
| **people** | Array | `"Assignee": ["user-id"]` |
| **relation** | Array | `"Related": ["page-id"]` |
| **files** | Array | `"Attachments": ["https://..."]` |

### Relative Dates

```json
{
  "Due Date": "today",
  "Start Date": "tomorrow",
  "End Date": "+7 days",
  "Review Date": "+2 weeks",
  "Archive Date": "+1 month"
}
```

**Supported formats:**
- `"today"`, `"tomorrow"`, `"yesterday"`
- `"+N days"`, `"-N days"`
- `"+N weeks"`, `"-N weeks"`
- `"+N months"`, `"-N months"`
- `"+N years"`, `"-N years"`

### Case Insensitivity

Property names and select values are case-insensitive:

```json
{
  "name": "Task",          // Works!
  "Name": "Task",          // Works!
  "NAME": "Task",          // Works!
  "status": "in progress", // Works!
  "Status": "In Progress"  // Works!
}
```

### Examples

**Create Task:**
```bash
notion-cli page create -d <DB_ID> -S --properties '{
  "Name": "Bug Fix: Login Error",
  "Status": "In Progress",
  "Priority": 8,
  "Due Date": "+3 days",
  "Tags": ["urgent", "bug"],
  "Assignee": ["user-id-123"],
  "Description": "User cannot log in with special characters"
}'
```

**Update Task:**
```bash
notion-cli page update <PAGE_ID> -S --properties '{
  "Status": "Done",
  "Completed": true,
  "Completion Date": "today"
}'
```

**Clear Property:**
```json
{
  "Description": null
}
```

[📖 Full Simple Properties Documentation](../SIMPLE_PROPERTIES.md)

---

## Database Operations

### Schema Discovery

```bash
# Get full schema
notion-cli db schema <ID> --json

# Get schema with examples
notion-cli db schema <ID> --with-examples --json

# Get specific properties
notion-cli db schema <ID> --properties Name,Status --json
```

### Query with Filters

```bash
# Simple filter
notion-cli db query <ID> --filter '{
  "property": "Status",
  "select": {"equals": "Done"}
}' --json

# Complex AND filter
notion-cli db query <ID> --filter '{
  "and": [
    {"property": "Status", "select": {"equals": "In Progress"}},
    {"property": "Priority", "number": {"greater_than": 5}}
  ]
}' --json

# OR filter
notion-cli db query <ID> --filter '{
  "or": [
    {"property": "Tags", "multi_select": {"contains": "urgent"}},
    {"property": "Tags", "multi_select": {"contains": "bug"}}
  ]
}' --json
```

### List Cached Databases

```bash
# List all cached databases
notion-cli list --json

# Output includes:
# - Database IDs
# - Titles
# - Cache age
# - Last modified
```

### Cache Management

```bash
# Check cache status
notion-cli cache:info --json

# Sync workspace
notion-cli sync

# Sync with progress
notion-cli sync --json
```

---

## Troubleshooting

### Workflow for Resolving Issues

**1. Run Health Check**
```bash
notion-cli doctor
```

This will identify the specific issue (token, connectivity, permissions, etc.)

**2. Common Issues and Solutions**

**Issue: Token Not Configured**
```bash
# Solution: Run init wizard
notion-cli init

# Or set manually
export NOTION_TOKEN="secret_your_token_here"
```

**Issue: API Connection Failed**
```bash
# Check token validity
notion-cli whoami

# Re-run setup
notion-cli init
```

**Issue: Database Not Found**
```bash
# Sync workspace to update cache
notion-cli sync

# Or use data_source_id instead of database_id
# The CLI auto-converts between them
```

**Issue: Invalid Properties**
```bash
# Get schema to see valid properties
notion-cli db schema <ID> --with-examples --json

# Use simple properties for easier syntax
notion-cli page create -d <ID> -S --properties '{...}'
```

**Issue: Permission Denied**
```bash
# Check workspace access with doctor
notion-cli doctor

# Verify integration has access to the database
# Visit: https://www.notion.so/my-integrations
```

### Error Messages

The CLI provides **platform-specific error messages**:

**Windows (Command Prompt):**
```
Error: NOTION_TOKEN not found

Set your token:
  set NOTION_TOKEN=secret_your_token_here

Or run: notion-cli init
```

**Windows (PowerShell):**
```
Error: NOTION_TOKEN not found

Set your token:
  $env:NOTION_TOKEN="secret_your_token_here"

Or run: notion-cli init
```

**Unix/Mac:**
```
Error: NOTION_TOKEN not found

Set your token:
  export NOTION_TOKEN="secret_your_token_here"

Or run: notion-cli init
```

---

## Best Practices

### For AI Agents

**1. Always Run Health Check First**
```bash
# Verify setup before operations
notion-cli doctor
```

**2. Cache Schema Before Creating Pages**
```bash
# Get schema to understand structure
notion-cli db schema <ID> --with-examples --json
```

**3. Use Simple Properties Mode**
```bash
# Reduces errors by 70%
notion-cli page create -d <ID> -S --properties '{...}'
```

**4. Use JSON Output for Parsing**
```bash
# All commands support --json
notion-cli db query <ID> --json | jq '.data.results'
```

**5. Check Cache Freshness**
```bash
# Before bulk operations
notion-cli cache:info --json
```

**6. Sync Workspace Periodically**
```bash
# Recommended: every 24 hours
notion-cli sync
```

**7. Use Verbose Mode for Debugging**
```bash
# Shows cache hits, retries, latency
notion-cli db query <ID> --verbose --json
```

### Performance Optimization

**1. Leverage Caching**
```bash
# One-time sync for faster lookups
notion-cli sync

# Use database names instead of IDs
notion-cli db query "Tasks Database" --json
```

**2. Filter Early**
```bash
# Reduce data transfer with filters
notion-cli db query <ID> --filter '{...}' --json
```

**3. Use Compact JSON**
```bash
# Faster parsing for large responses
notion-cli db query <ID> --compact-json
```

### Error Handling

**1. Check Exit Codes**
```bash
notion-cli db retrieve <ID> --json
if [ $? -eq 0 ]; then
  echo "Success!"
else
  echo "Failed!"
fi
```

**Exit codes:**
- `0` = Success
- `1` = Notion API error
- `2` = CLI error (invalid flags, etc.)

**2. Parse Error Responses**
```json
{
  "success": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "Database not found",
    "details": "Check that the database ID is correct"
  }
}
```

### Security Best Practices

**1. Use Environment Variables**
```bash
# Never hardcode tokens
export NOTION_TOKEN="secret_token"
```

**2. Use Config Commands**
```bash
# Secure token storage
notion-cli config set-token
```

**3. Verify Permissions**
```bash
# Check integration access
notion-cli doctor
```

---

## Additional Resources

- [Simple Properties Guide](../SIMPLE_PROPERTIES.md)
- [AI Agent Cookbook](./ai-agent-cookbook.md)
- [Filter Guide](./filter-guide.md)
- [Verbose Logging Guide](../VERBOSE_LOGGING.md)
- [Output Formats Guide](../../OUTPUT_FORMATS.md)

---

## Summary

**Key Commands for AI Agents:**

```bash
# Setup & Health
notion-cli init           # First-time setup wizard
notion-cli doctor         # Comprehensive health check
notion-cli whoami         # Quick connection test

# Workspace Management
notion-cli sync           # Cache all databases
notion-cli list --json    # List cached databases

# Schema Discovery
notion-cli db schema <ID> --with-examples --json

# Create/Update Pages (Simple Properties)
notion-cli page create -d <ID> -S --properties '{...}'
notion-cli page update <ID> -S --properties '{...}'

# Query Databases
notion-cli db query <ID> --filter '{...}' --json

# Cache Management
notion-cli cache:info --json
```

**Recommended Workflow:**
1. `notion-cli init` - First-time setup
2. `notion-cli doctor` - Verify health
3. `notion-cli sync` - Cache workspace
4. `notion-cli db schema <ID> --with-examples --json` - Discover structure
5. Use `-S` flag for simple properties
6. Use `--json` for all output

---

**Built for AI agents, optimized for automation.**
