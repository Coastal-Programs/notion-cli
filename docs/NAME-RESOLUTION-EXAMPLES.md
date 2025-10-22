# Name Resolution Examples

## Overview

This guide demonstrates real-world workflows using the name resolution feature in notion-cli v5.3.0. Instead of remembering long database IDs or copying URLs, you can now reference databases by their natural names.

---

## Table of Contents

1. [First-Time Setup](#first-time-setup)
2. [Daily Workflows](#daily-workflows)
3. [Advanced Use Cases](#advanced-use-cases)
4. [Troubleshooting](#troubleshooting)
5. [Best Practices](#best-practices)

---

## First-Time Setup

### Step 1: Install notion-cli

**Mac/Linux:**
```bash
npm install -g Coastal-Programs/notion-cli
```

**Windows:**
```bash
git clone https://github.com/Coastal-Programs/notion-cli
cd notion-cli
npm install -g .
```

### Step 2: Configure Your API Token

The easy way (recommended):
```bash
notion-cli config set-token
# Follow the prompts - it will save to your shell config automatically
```

Manual way:
```bash
# Mac/Linux (bash/zsh)
export NOTION_TOKEN="secret_your_token_here"
echo 'export NOTION_TOKEN="secret_your_token_here"' >> ~/.bashrc
source ~/.bashrc

# Windows PowerShell
$env:NOTION_TOKEN="secret_your_token_here"
[System.Environment]::SetEnvironmentVariable('NOTION_TOKEN', 'secret_your_token_here', 'User')
```

### Step 3: Sync Your Workspace

This is the magic step that enables name-based lookups:

```bash
notion-cli sync
```

**Example output:**
```
Syncing workspace databases... done
✓ Found 15 databases
✓ Generating search aliases... done
✓ Cache saved to ~/.notion-cli/databases.json

Indexed databases:
  • Tasks Database (aliases: tasks database, tasks, task, td)
  • Meeting Notes (aliases: meeting notes, meetings, notes, mn)
  • Customer CRM (aliases: customer crm, customers, customer, crm, cc)
  • Product Roadmap (aliases: product roadmap, product, roadmap, pr)
  • Bug Tracker (aliases: bug tracker, bugs, bug, bt)
  ... and 10 more
```

### Step 4: Browse Your Cached Databases

```bash
notion-cli list
```

**Example output:**
```
Cached Databases (15 total)
Last synced: 10/22/2025, 2:30:00 PM

Title                    ID                                Aliases (first 3)
Tasks Database           1fb79d4c71bb8032b722c82305b63a00  tasks, task, td
Meeting Notes            2a8c3d5e71bb8042b833d94316c74b11  meetings, notes, mn
Customer CRM             3b9d4e6f82bc8053c944e05427d85c22  customers, customer, crm
Product Roadmap          4c0e5f7g93cd9164d055f16538e96d33  product, roadmap, pr
Bug Tracker              5d1f6g8h04de0275e166g27649f07e44  bugs, bug, bt

Tip: Run "notion-cli sync" to refresh the cache.
```

---

## Daily Workflows

### Scenario 1: Quick Task Management

**Old way (remembering IDs):**
```bash
# Had to remember or look up the ID every time
notion-cli db query 1fb79d4c71bb8032b722c82305b63a00 --json
```

**New way (using names):**
```bash
# Just use the database name!
notion-cli db query "Tasks Database" --json

# Or use an alias
notion-cli db query "tasks" --json

# Even partial matches work
notion-cli db query "task" --json
```

### Scenario 2: Meeting Notes Workflow

Create a new meeting note page:

```bash
# Old way
notion-cli page create -d 2a8c3d5e71bb8042b833d94316c74b11 --json

# New way - much more readable!
notion-cli page create -d "Meeting Notes" --json

# Even shorter with alias
notion-cli page create -d "meetings" --json
```

### Scenario 3: Customer Data Query

Get all customers with a specific status:

```bash
# Query using database name
notion-cli db query "Customer CRM" \
  --filter '{"property":"Status","select":{"equals":"Active"}}' \
  --json | jq '.data.results[] | {name: .properties.Name.title[0].plain_text, email: .properties.Email.email}'

# Or use the short alias
notion-cli db query "crm" \
  --filter '{"property":"Status","select":{"equals":"Active"}}' \
  --json
```

### Scenario 4: Bulk Operations Across Databases

Script to query multiple databases by name:

```bash
#!/bin/bash
# query-all-active-items.sh

DATABASES=("Tasks Database" "Bug Tracker" "Feature Requests")

for db in "${DATABASES[@]}"; do
  echo "=== $db ==="
  notion-cli db query "$db" \
    --filter '{"property":"Status","select":{"equals":"Active"}}' \
    --json | jq '.data.results | length'
  echo ""
done
```

Output:
```
=== Tasks Database ===
12 active items

=== Bug Tracker ===
8 active items

=== Feature Requests ===
23 active items
```

---

## Advanced Use Cases

### Use Case 1: AI Agent Automation

**Scenario:** An AI assistant needs to add tasks to Notion based on email analysis.

```python
# ai_email_to_notion.py
import subprocess
import json

def add_task_to_notion(task_title, priority, due_date):
    """Add a task to the Tasks database using name resolution."""

    # No need to hardcode database IDs!
    database = "Tasks Database"

    # Create page properties
    properties = {
        "Name": {"title": [{"text": {"content": task_title}}]},
        "Priority": {"select": {"name": priority}},
        "Due Date": {"date": {"start": due_date}}
    }

    # Call notion-cli using database name
    cmd = [
        "notion-cli", "page", "create",
        "-d", database,
        "--properties", json.dumps(properties),
        "--json"
    ]

    result = subprocess.run(cmd, capture_output=True, text=True)
    response = json.loads(result.stdout)

    if response['success']:
        print(f"✓ Created task: {task_title}")
        return response['data']['id']
    else:
        print(f"✗ Failed: {response['error']['message']}")
        return None

# Example usage
add_task_to_notion(
    task_title="Review Q1 budget proposal",
    priority="High",
    due_date="2025-11-01"
)
```

### Use Case 2: Cross-Database Reporting

**Scenario:** Generate a weekly report from multiple databases.

```bash
#!/bin/bash
# weekly-report.sh

REPORT_DATE=$(date +%Y-%m-%d)
REPORT_FILE="weekly-report-${REPORT_DATE}.md"

echo "# Weekly Report - ${REPORT_DATE}" > "$REPORT_FILE"
echo "" >> "$REPORT_FILE"

# Tasks completed this week
echo "## Tasks Completed" >> "$REPORT_FILE"
notion-cli db query "Tasks" \
  --filter '{"property":"Status","select":{"equals":"Done"}}' \
  --json | jq -r '.data.results[] | "- " + .properties.Name.title[0].plain_text' \
  >> "$REPORT_FILE"

echo "" >> "$REPORT_FILE"

# Bugs fixed this week
echo "## Bugs Fixed" >> "$REPORT_FILE"
notion-cli db query "Bug Tracker" \
  --filter '{"property":"Status","select":{"equals":"Fixed"}}' \
  --json | jq -r '.data.results[] | "- " + .properties.Title.title[0].plain_text' \
  >> "$REPORT_FILE"

echo "" >> "$REPORT_FILE"

# Meetings held this week
echo "## Meetings" >> "$REPORT_FILE"
notion-cli db query "Meeting Notes" \
  --json | jq -r '.data.results[] | "- " + .properties.Name.title[0].plain_text' \
  >> "$REPORT_FILE"

echo "✓ Report saved to ${REPORT_FILE}"
cat "$REPORT_FILE"
```

### Use Case 3: Schema-Aware Data Import

**Scenario:** Import CSV data into Notion, discovering schema automatically.

```python
# csv_to_notion.py
import subprocess
import json
import csv

def get_database_schema(db_name):
    """Get schema for a database using name resolution."""
    cmd = ["notion-cli", "db", "schema", db_name, "--json"]
    result = subprocess.run(cmd, capture_output=True, text=True)
    response = json.loads(result.stdout)

    if response['success']:
        return response['data']['properties']
    else:
        raise Exception(f"Failed to get schema: {response['error']['message']}")

def import_csv_to_notion(csv_file, database_name):
    """Import CSV data to Notion database."""

    # Discover schema using database name
    schema = get_database_schema(database_name)
    print(f"✓ Schema discovered for '{database_name}'")
    print(f"  Properties: {', '.join([p['name'] for p in schema])}")

    # Read CSV
    with open(csv_file, 'r') as f:
        reader = csv.DictReader(f)

        for row in reader:
            # Create page using database name (no ID needed!)
            cmd = [
                "notion-cli", "page", "create",
                "-d", database_name,
                "--json"
            ]

            # Add properties based on schema
            properties = {}
            for prop in schema:
                prop_name = prop['name']
                if prop_name in row:
                    properties[prop_name] = format_property(
                        row[prop_name],
                        prop['type']
                    )

            cmd.extend(["--properties", json.dumps(properties)])

            result = subprocess.run(cmd, capture_output=True, text=True)
            response = json.loads(result.stdout)

            if response['success']:
                print(f"✓ Imported: {row.get('Name', 'Unnamed')}")
            else:
                print(f"✗ Failed: {response['error']['message']}")

def format_property(value, prop_type):
    """Format value based on property type."""
    if prop_type == 'title':
        return {"title": [{"text": {"content": value}}]}
    elif prop_type == 'rich_text':
        return {"rich_text": [{"text": {"content": value}}]}
    elif prop_type == 'number':
        return {"number": float(value)}
    elif prop_type == 'select':
        return {"select": {"name": value}}
    # Add more types as needed

    return value

# Example usage
import_csv_to_notion('customers.csv', 'Customer CRM')
```

### Use Case 4: Continuous Sync for Real-Time Workflows

**Scenario:** Keep cache fresh for time-sensitive applications.

```bash
#!/bin/bash
# auto-sync-daemon.sh

SYNC_INTERVAL=3600  # 1 hour

while true; do
  echo "[$(date)] Syncing workspace..."

  notion-cli sync --json > /tmp/notion-sync.log

  if [ $? -eq 0 ]; then
    count=$(cat /tmp/notion-sync.log | jq '.count')
    echo "[$(date)] ✓ Synced $count databases"
  else
    echo "[$(date)] ✗ Sync failed"
    cat /tmp/notion-sync.log
  fi

  sleep $SYNC_INTERVAL
done
```

---

## Troubleshooting

### Problem: "Database not found" error

**Symptom:**
```bash
$ notion-cli db retrieve "Tasks"
Error: Database "Tasks" not found.
```

**Solutions:**

1. **Check if cache exists:**
   ```bash
   ls -la ~/.notion-cli/databases.json
   ```

   If missing, run:
   ```bash
   notion-cli sync
   ```

2. **Verify database name:**
   ```bash
   notion-cli list | grep -i task
   ```

   Make sure you're using the exact title or a valid alias.

3. **Try different variations:**
   ```bash
   # Try with full title
   notion-cli db retrieve "Tasks Database"

   # Try with alias
   notion-cli db retrieve "tasks"

   # Try partial match
   notion-cli db retrieve "task"
   ```

4. **Check cache freshness:**
   ```bash
   notion-cli list --json | jq '.lastSync'
   ```

   If outdated, force resync:
   ```bash
   notion-cli sync --force
   ```

### Problem: Multiple databases match the query

**Symptom:**
```bash
$ notion-cli db retrieve "task"
Error: Multiple databases match "task". Please be more specific.
```

**Solution:**

List all matching databases:
```bash
notion-cli list --json | jq '.databases[] | select(.title | contains("task"))'
```

Use more specific name:
```bash
# Instead of "task", use full title
notion-cli db retrieve "Tasks Database"

# Or use unique alias
notion-cli db retrieve "td"
```

### Problem: Cache is stale

**Symptom:**
New databases don't show up in `list` or name resolution fails.

**Solution:**

Force refresh the cache:
```bash
notion-cli sync --force
```

Set up automatic sync (cron job):
```bash
# Add to crontab (sync every hour)
crontab -e

# Add this line:
0 * * * * notion-cli sync --json > /dev/null 2>&1
```

### Problem: Permission denied on cache file

**Symptom:**
```bash
$ notion-cli sync
Error: Failed to save cache: EACCES: permission denied
```

**Solution:**

Fix permissions:
```bash
# Check current permissions
ls -la ~/.notion-cli/

# Fix if needed
chmod 755 ~/.notion-cli
chmod 644 ~/.notion-cli/databases.json

# Or recreate directory
rm -rf ~/.notion-cli
notion-cli sync
```

---

## Best Practices

### 1. Consistent Naming Conventions

**Good database names (work well with aliases):**
- "Tasks Database" → aliases: tasks, task, td
- "Meeting Notes" → aliases: meetings, notes, mn
- "Customer CRM" → aliases: customers, crm, cc

**Avoid:**
- Generic names: "Database 1", "New Database"
- All caps: "TASKS" (harder to type)
- Special characters: "Task$_Database" (breaks alias generation)

### 2. Regular Cache Refresh

**Recommendation:** Sync at least once daily for active projects.

**Manual sync:**
```bash
# First thing in the morning
notion-cli sync
```

**Automated sync (recommended):**
```bash
# Add to shell startup (~/.bashrc or ~/.zshrc)
echo 'notion-cli sync --json > /dev/null 2>&1 &' >> ~/.bashrc
```

**Cron job:**
```bash
# Every 6 hours
0 */6 * * * notion-cli sync --json > /dev/null 2>&1
```

### 3. Use Aliases in Scripts

**Good practice:**
```bash
# Use short, memorable aliases
notion-cli db query "tasks" --json
notion-cli db query "crm" --json
```

**Not recommended:**
```bash
# Don't hardcode IDs in scripts
notion-cli db query "1fb79d4c71bb8032b722c82305b63a00" --json

# Don't use partial matches in production (ambiguous)
notion-cli db query "task" --json  # Might match multiple DBs
```

### 4. Document Your Database Names

Keep a team reference doc:

```markdown
# Team Notion Databases

## Development
- `notion-cli db query "Bug Tracker"` - All bugs
- `notion-cli db query "Feature Requests"` - Feature backlog
- `notion-cli db query "Sprint Board"` - Current sprint

## Operations
- `notion-cli db query "Incident Log"` - Production incidents
- `notion-cli db query "Deploy Schedule"` - Release calendar

## Business
- `notion-cli db query "Customer CRM"` - Customer database
- `notion-cli db query "Sales Pipeline"` - Deal tracking
```

### 5. Error Handling in Scripts

Always check for success:

```bash
#!/bin/bash
# Good error handling

result=$(notion-cli db query "Tasks" --json)
success=$(echo "$result" | jq -r '.success')

if [ "$success" = "true" ]; then
  echo "$result" | jq '.data.results'
else
  error=$(echo "$result" | jq -r '.error.message')
  echo "Error: $error" >&2

  # Check if cache issue
  if echo "$error" | grep -q "not found"; then
    echo "Tip: Run 'notion-cli sync' to refresh cache" >&2
  fi

  exit 1
fi
```

### 6. Version Control for Cache

**Don't commit cache files:**
```bash
# .gitignore
.notion-cli/
databases.json
```

**Instead, document setup:**
```bash
# setup.sh
#!/bin/bash
echo "Setting up Notion CLI..."
notion-cli config set-token
notion-cli sync
echo "✓ Setup complete!"
```

---

## Migration Guide

### From ID-Based to Name-Based

**Before (old workflow):**
```bash
# Step 1: Find database ID
notion-cli search --query "Tasks" --json | jq '.data.results[0].id'
# Copy ID: 1fb79d4c71bb8032b722c82305b63a00

# Step 2: Save ID somewhere (env var, config file, etc.)
export TASKS_DB_ID="1fb79d4c71bb8032b722c82305b63a00"

# Step 3: Use in commands
notion-cli db query "$TASKS_DB_ID" --json
```

**After (new workflow):**
```bash
# Step 1: One-time sync
notion-cli sync

# Step 2: Use database name directly
notion-cli db query "Tasks Database" --json

# Or use alias
notion-cli db query "tasks" --json
```

**Migration script for existing codebase:**
```bash
#!/bin/bash
# migrate-to-name-resolution.sh

# Find all hardcoded database IDs
echo "Searching for hardcoded database IDs..."
grep -r "1fb79d4c71bb8032b722c82305b63a00" . --exclude-dir=node_modules

# For each ID, find the database name
echo "Fetching database names..."
notion-cli list --json | jq -r '.databases[] | "\(.id) -> \(.title)"'

echo ""
echo "Replace IDs with database names in your scripts:"
echo "  Before: notion-cli db query \"1fb79d4c71bb8032b722c82305b63a00\""
echo "  After:  notion-cli db query \"Tasks Database\""
```

---

## Quick Reference

### Common Commands

```bash
# Setup
notion-cli config set-token          # Configure API token
notion-cli sync                      # Sync workspace databases
notion-cli sync --force              # Force resync

# List & Browse
notion-cli list                      # List cached databases
notion-cli list --json               # JSON format
notion-cli list --markdown           # Markdown table

# Name Resolution
notion-cli db retrieve "Database Name"       # Exact title
notion-cli db retrieve "alias"               # Use alias
notion-cli db retrieve "partial"             # Partial match
notion-cli db retrieve <ID>                  # Still works!
notion-cli db retrieve <URL>                 # Still works!

# Query
notion-cli db query "Database Name" --json
notion-cli db query "alias" --filter '{...}' --json

# Schema Discovery
notion-cli db schema "Database Name" --json

# Create Pages
notion-cli page create -d "Database Name" --json
```

### Keyboard Shortcuts

Use shell aliases for even faster access:

```bash
# Add to ~/.bashrc or ~/.zshrc

# Sync
alias nsync='notion-cli sync'

# List
alias nls='notion-cli list'

# Query common databases
alias ntasks='notion-cli db query "Tasks" --json'
alias nmeetings='notion-cli db query "Meetings" --json'
alias ncrm='notion-cli db query "CRM" --json'

# Schema
alias nschema='notion-cli db schema'

# Create page
alias npage='notion-cli page create -d'
```

Usage:
```bash
nsync           # Sync workspace
nls             # List databases
ntasks          # Query tasks
nschema "CRM"   # Get CRM schema
npage "Tasks"   # Create page in Tasks
```

---

## Additional Resources

- [Testing Guide](./NAME-RESOLUTION-TESTS.md) - Comprehensive testing documentation
- [README](../README.md) - Full CLI documentation
- [AI Agent Cookbook](./AI-AGENT-COOKBOOK.md) - Automation recipes
- [Caching Architecture](./CACHING-ARCHITECTURE.md) - Cache system details

---

## Feedback & Support

Found an issue or have suggestions?
- GitHub Issues: https://github.com/Coastal-Programs/notion-cli/issues
- Provide feedback on name resolution accuracy
- Share your workflow examples!
