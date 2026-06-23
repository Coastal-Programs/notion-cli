`notion-cli auth`
================

Authentication commands — log in, log out, check status, and refresh OAuth tokens.

On a fresh interactive install, commands that need the Notion API automatically
start first-time OAuth setup when no workspace credentials or legacy token
exist. The selected Notion workspace is stored as the default workspace. Scripts
and commands with an explicit `--auth-workspace` / `NOTION_WORKSPACE` do not
open a browser automatically; run `notion-cli auth login` first in those cases.

* [`notion-cli auth login`](#notion-cli-auth-login)
* [`notion-cli auth logout`](#notion-cli-auth-logout)
* [`notion-cli auth status`](#notion-cli-auth-status)
* [`notion-cli auth refresh`](#notion-cli-auth-refresh)
* [`notion-cli auth list`](#notion-cli-auth-list)
* [`notion-cli auth default`](#notion-cli-auth-default)

---

## `notion-cli auth login`

Authenticate with Notion via the OAuth 2.0 authorization flow.

```
USAGE
  $ notion-cli auth login [--manual] [--slug <value>] [-j]

FLAGS
  -j, --json     Output result as JSON
      --manual   Skip the local callback server and paste the redirected URL by hand
      --slug     Override the local workspace credential slug

DESCRIPTION
  Opens your browser to the Notion authorization page and exchanges the
  returned code for an OAuth access token. The workspace slug is derived from
  the returned Notion workspace name unless --slug is provided. Tokens are saved
  to the OS keychain; non-secret workspace metadata is saved to
  ~/.config/notion-cli/credentials.json.

  OAuth client credentials must be available either from the release binary,
  from a local build with NOTION_OAUTH_CLIENT_ID / NOTION_OAUTH_SECRET embedded,
  or from those runtime environment variables during local development. If they
  are missing or placeholder values, the CLI fails before opening the browser.

  Use --manual when running over SSH, in a container, or behind a firewall
  where the CLI's localhost callback server cannot be reached from the browser.
  In manual mode the authorization URL is printed; you open it yourself, then
  paste the full redirected URL (or bare code) back into the terminal.

  SSH sessions are automatically detected and force --manual mode.

EXAMPLES
  Log in via browser (default)

    $ notion-cli auth login

  Log in and force a local slug

    $ notion-cli auth login --slug personal

  Log in without a browser (SSH / container)

    $ notion-cli auth login --manual
```

---

## `notion-cli auth logout`

Remove stored workspace credentials.

```
USAGE
  $ notion-cli auth logout [workspace] [--local-only] [-j]

FLAGS
  -j, --json          Output result as JSON
      --local-only    Clear local config only; do not call the Notion revoke API

DESCRIPTION
  By default, an OAuth access token is revoked on the Notion API before being
  removed from local credential storage. If a workspace argument is provided,
  that stored workspace credential is removed. Without an argument, the active
  workspace is used.

  Use --local-only to skip the API call entirely and only delete local credentials.

EXAMPLES
  Log out and revoke the token remotely

    $ notion-cli auth logout

  Log out a specific stored workspace

    $ notion-cli auth logout haven

  Clear local credentials without calling the Notion API

    $ notion-cli auth logout --local-only
```

---

## `notion-cli auth status`

Display the current authentication method and details.

```
USAGE
  $ notion-cli [--auth-workspace <slug>] auth status [--remote] [-j]

FLAGS
  -j, --json     Output as JSON
      --remote   Also call the Notion token introspect API to verify active status

DESCRIPTION
  Prints the active authentication method (oauth, env, token, or none) along
  with masked token, local workspace slug, workspace info, and — if an expiry
  is stored — the token expiry timestamp.

  With --remote, the CLI additionally calls the Notion token introspect
  endpoint and merges active status, scope, and issued_at into the output.
  Only works when authenticated via OAuth and OAuth client credentials are
  embedded in the build.

OUTPUT FIELDS (oauth method)
  auth_method      "oauth"
  workspace        Local workspace credential slug
  workspace_id     Notion workspace UUID
  workspace_name   Human-readable workspace name
  bot_id           Integration bot user ID
  token            Masked access token (first/last 4 chars)
  expires_at       RFC3339 expiry timestamp (when stored)

EXTRA FIELDS (with --remote)
  active           true/false from the introspect API
  scope            OAuth scope string (if returned by Notion)
  issued_at        RFC3339 token issue time (if returned by Notion)
  remote_error     Error message if the introspect call failed

EXAMPLES
  Check local auth state

    $ notion-cli auth status

  Check a specific stored workspace

    $ notion-cli --auth-workspace haven auth status

  Check auth state and verify token with the Notion API

    $ notion-cli auth status --remote

  Output as JSON for scripting

    $ notion-cli auth status --json
```

---

## `notion-cli auth refresh`

Exchange the stored refresh token for a new OAuth access token.

```
USAGE
  $ notion-cli [--auth-workspace <slug>] auth refresh [-j]

FLAGS
  -j, --json   Output result as JSON

DESCRIPTION
  Calls the Notion token endpoint with grant_type=refresh_token using the
  refresh token saved by 'auth login'. On success the new access token (and
  updated refresh token, if returned) are persisted to the selected workspace
  credential.

  This command requires that:
    1. You previously ran 'notion-cli auth login' (refresh token is stored).
    2. OAuth client credentials are available in this build or environment.

  The CLI also performs this refresh automatically and transparently when any
  API call returns a 401 response and a refresh token is available — you do
  not normally need to run this command manually.

OUTPUT FIELDS
  auth_method      "oauth"
  workspace        Local workspace credential slug
  workspace_id     Notion workspace UUID (if returned by Notion)
  workspace_name   Human-readable workspace name (if returned by Notion)
  expires_at       RFC3339 expiry of the new token (when Notion returns expires_in)

EXAMPLES
  Manually refresh the stored OAuth token

    $ notion-cli auth refresh

  Refresh and inspect the result as JSON

    $ notion-cli auth refresh --json
```

---

## `notion-cli auth list`

List stored workspace credentials. Secrets are never printed.

```
USAGE
  $ notion-cli auth list [-j]

FLAGS
  -j, --json   Output result as JSON

EXAMPLES
  List stored workspaces

    $ notion-cli auth list
```

---

## `notion-cli auth default`

Show or set the default workspace credential.

```
USAGE
  $ notion-cli auth default [workspace] [-j]

FLAGS
  -j, --json   Output result as JSON

DESCRIPTION
  In an interactive terminal, running without an argument opens a workspace
  selector. Use Up/Down or j/k to move, Space to select, Enter to save, and q
  to cancel. In non-interactive shells or when an output flag is set, running
  without an argument prints the current default workspace.

  With an argument, sets the default stored workspace. If named workspace
  credentials exist, one of them must remain the default. Use --auth-workspace
  default on individual commands when you explicitly need the legacy config.json
  credential.

EXAMPLES
  Select the default workspace interactively

    $ notion-cli auth default

  Show the current default as JSON

    $ notion-cli auth default --json

  Set the default workspace

    $ notion-cli auth default haven
```
