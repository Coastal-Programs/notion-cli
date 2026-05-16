`notion-cli init`
=================

> **Note:** The standalone `init` wizard is planned for a future release. Current Go builds use OAuth login and first-run setup instead.

Interactive first-time setup wizard for Notion CLI

## Status

This command was available in v5.x (TypeScript) and will be re-implemented in a future Go release.

Current Go builds use OAuth login and first-run setup instead of a standalone
`init` command. In an interactive shell, commands that need the Notion API can
start OAuth setup automatically when no credentials exist. Non-interactive
shells print setup instructions instead of opening a browser.

## Current Setup

**Option 1: OAuth login (recommended for interactive use)**

```bash
notion-cli auth login

# Official-style alias
notion-cli login
```

**Option 2: Environment variable (recommended for CI and automation)**

```bash
export NOTION_TOKEN="secret_your_token_here"

# Official CLI compatibility alias when NOTION_TOKEN is unset
export NOTION_API_TOKEN="secret_your_token_here"
```

**Option 3: Legacy config file**

```bash
# Set token via config command
notion-cli config set-token

# Or pipe it for security
echo "$NOTION_TOKEN" | notion-cli config set-token
```

**Verify setup**

```bash
# Check connectivity
notion-cli whoami

# Run diagnostics
notion-cli doctor

# Sync workspace databases
notion-cli sync
```
