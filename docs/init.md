`notion-cli init`
=================

> **Note:** The `init` command is planned for a future release (Phase 2). It is not available in the current Go rewrite (v6.0.0).

Interactive first-time setup wizard for Notion CLI

## Status

This command was available in v5.x (TypeScript) and will be re-implemented in a future Go release.

## Alternative Setup (v6.0.0)

In the current version, set up your token using one of these methods:

**Option 1: Environment variable (recommended)**

```bash
export NOTION_TOKEN="secret_your_token_here"
```

**Option 2: Config file**

```bash
# Set token via config command
notion-cli config set-token

# Or pipe it for security
echo "$NOTION_TOKEN" | notion-cli config set-token
```

**Option 3: Verify setup**

```bash
# Check connectivity
notion-cli whoami

# Run diagnostics
notion-cli doctor

# Sync workspace databases
notion-cli sync
```

