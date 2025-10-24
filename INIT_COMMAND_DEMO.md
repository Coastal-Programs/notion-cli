# `notion-cli init` - Interactive Setup Demo

## Quick Start

```bash
notion-cli init
```

## Interactive Mode Demo

### Scenario 1: First-Time Setup (No Token)

```bash
$ notion-cli init

===========================================================
  Welcome to Notion CLI!
===========================================================

This wizard will help you set up your Notion CLI in 3 steps:
  1. Configure your Notion integration token
  2. Test the connection to Notion API
  3. Sync your workspace databases

Let's get started!

===========================================================
Step 1/3: Set your Notion token
===========================================================

You need a Notion integration token to use this CLI.
Get one at: https://www.notion.so/my-integrations

Enter your Notion integration token: secret_abc123...

Token set for this session.

Note: To persist this token, add it to your shell configuration:
  export NOTION_TOKEN="secret_abc123..."

Or use: notion-cli config set-token

Step 1 complete!

===========================================================
Step 2/3: Test connection
===========================================================

⣽ Connecting to Notion API... connected

Bot Name: My Integration
Bot ID: 12345678-1234-1234-1234-123456789012
Workspace: My Workspace
Connection latency: 187ms

Step 2 complete!

===========================================================
Step 3/3: Sync workspace
===========================================================

This will index all databases your integration can access.

⣾ Syncing databases... found 33

Synced 33 databases in 1.29s

Your integration has access to these databases:
  - Project Tracker
  - Tasks
  - Meeting Notes
  - Clients
  - Product Roadmap
  ... and 28 more

Step 3 complete!

===========================================================
  Setup Complete!
===========================================================

Your Notion CLI is ready to use!

Quick Start Commands:
  notion-cli list              - List all databases
  notion-cli db query <name>   - Query a database
  notion-cli whoami            - Check connection status
  notion-cli sync              - Refresh workspace cache

Documentation:
  https://github.com/Coastal-Programs/notion-cli

Need help? Run any command with --help flag

Happy building with Notion!
```

---

### Scenario 2: Already Configured (Reconfiguration)

```bash
$ notion-cli init

You already have a configured Notion token.
Running init again will update your configuration.

Do you want to reconfigure? (y/n): n

Setup cancelled. Your existing configuration is unchanged.
```

If user chooses "y", the wizard proceeds as normal.

---

### Scenario 3: JSON Mode (Automation)

```bash
$ export NOTION_TOKEN="secret_abc123..."
$ notion-cli init --json
```

**Output:**
```json
{
  "success": true,
  "message": "Notion CLI setup complete",
  "data": {
    "token": {
      "source": "user_input",
      "updated": true,
      "tokenLength": 50
    },
    "connection": {
      "success": true,
      "bot": {
        "id": "9292ffa8-7279-4a71-86ba-8cea0b6ac20a",
        "name": "Claude Code Access",
        "type": "bot"
      },
      "workspace": {
        "name": "Coastal Programs's ",
        "id": "502c1ab7-2d3e-433e-b82d-f61b4d623f29"
      },
      "latency_ms": 0
    },
    "sync": {
      "success": true,
      "databases_found": 33,
      "sync_time_ms": 1293,
      "cached": 33
    }
  },
  "next_steps": [
    "notion-cli list - List all databases",
    "notion-cli db query <name-or-id> - Query a database",
    "notion-cli whoami - Check connection status",
    "notion-cli sync - Refresh workspace cache"
  ],
  "metadata": {
    "timestamp": "2025-10-24T05:51:33.091Z",
    "command": "init"
  }
}
```

---

## Error Handling Examples

### Error 1: Missing Token (JSON Mode)

```bash
$ unset NOTION_TOKEN
$ notion-cli init --json
```

**Output:**
```json
{
  "success": false,
  "error": {
    "code": "TOKEN_MISSING",
    "message": "NOTION_TOKEN required in JSON mode",
    "suggestions": [
      {
        "description": "Set token in environment before running init",
        "command": "export NOTION_TOKEN=\"secret_your_token_here\""
      }
    ],
    "context": {},
    "timestamp": "2025-10-24T05:51:08.269Z"
  }
}
```

---

### Error 2: Invalid Token Format

```bash
$ notion-cli init
Enter your Notion integration token: invalid_token

❌ Invalid token format - Notion tokens must start with "secret_"
   Error Code: TOKEN_INVALID

💡 Possible causes and fixes:
   1. Get your integration token from Notion
      🔗 https://developers.notion.com/docs/create-a-notion-integration
   2. Tokens should look like: secret_abc123...
```

---

### Error 3: Connection Failure

```bash
$ notion-cli init
# ... token setup succeeds ...

===========================================================
Step 2/3: Test connection
===========================================================

⣾ Connecting to Notion API... failed

❌ Authentication failed - your NOTION_TOKEN is invalid or expired
   Error Code: TOKEN_INVALID

💡 Possible causes and fixes:
   1. Verify your integration still exists and is active
      🔗 https://www.notion.so/my-integrations
   2. Generate a new internal integration token
      🔗 https://developers.notion.com/docs/create-a-notion-integration
   3. Update your token using the config command
      $ notion-cli config set-token
   4. Check if the integration has been removed or revoked by workspace admin
```

---

## Use Cases

### 1. New User Onboarding
Perfect for first-time users who need guidance setting up their environment.

```bash
notion-cli init
```

### 2. CI/CD Pipeline Setup
Automated setup in deployment pipelines with JSON output for logging.

```bash
#!/bin/bash
export NOTION_TOKEN="${NOTION_API_TOKEN}"
OUTPUT=$(notion-cli init --json)
if echo "$OUTPUT" | jq -e '.success' > /dev/null; then
  echo "Setup successful"
  echo "$OUTPUT" | jq -r '.data.sync.databases_found' | xargs echo "Databases found:"
else
  echo "Setup failed"
  echo "$OUTPUT" | jq -r '.error.message'
  exit 1
fi
```

### 3. Workspace Switching
When switching between different Notion workspaces.

```bash
# Switch to new workspace
export NOTION_TOKEN="secret_new_workspace_token"
notion-cli init
```

### 4. Troubleshooting
When something isn't working, re-running init can help diagnose issues.

```bash
# Reconfigure and test connection
notion-cli init
```

---

## Command Features

### ✅ Smart Detection
- Detects existing configuration
- Validates token health
- Prompts before overwriting

### ✅ Progressive Steps
- Token setup with validation
- Connection test with latency
- Database sync with progress

### ✅ Dual Output Modes
- Interactive for humans
- JSON for automation

### ✅ Comprehensive Errors
- Clear error messages
- Actionable suggestions
- Helpful documentation links

### ✅ Educational
- Explains each step
- Shows what's happening
- Provides next steps

---

## Integration with Other Commands

After running `init`, users can immediately use:

```bash
# List all databases
notion-cli list

# Query a database
notion-cli db query "Tasks"

# Check connection
notion-cli whoami

# Refresh cache
notion-cli sync
```

---

## Performance

Based on actual test with 33 databases:

| Step | Duration | Notes |
|------|----------|-------|
| Token Setup | ~instant | User input bound |
| Connection Test | <100ms | Single API call |
| Database Sync | ~1.3s | Paginated search (33 DBs) |
| **Total** | **~1.4s** | Excluding user input time |

---

## Comparison: Before vs After

### Before (Manual Setup)
```bash
# User has to figure this out themselves
export NOTION_TOKEN="secret_abc123..."
notion-cli sync
notion-cli whoami
```

### After (Guided Setup)
```bash
# Everything in one command
notion-cli init

# Guided through:
# 1. Where to get token
# 2. How to set it up
# 3. Verification it works
# 4. What to do next
```

---

## Developer Notes

### Implementation Quality
- **TypeScript:** Strict type checking, no `any` types
- **Error Handling:** Uses NotionCLIError system
- **Code Reuse:** Leverages existing utilities
- **Testing:** Manually verified all paths

### Design Decisions
1. **No Command Spawning:** Calls Notion API directly for reliability
2. **Progressive Disclosure:** Shows help when needed, not overwhelming
3. **Escape Hatches:** Can skip reconfiguration, use existing token
4. **Educational:** Teaches users about the CLI as they set it up

### Future Enhancements
- Token permission testing
- Database selection wizard
- Multi-workspace profiles
- Auto-run on first install

---

## Conclusion

The `notion-cli init` command transforms the first-time setup experience from confusing to delightful. It guides users through configuration, validates their setup, and leaves them ready to use the CLI with confidence.

**Ready to try it?**
```bash
notion-cli init
```
