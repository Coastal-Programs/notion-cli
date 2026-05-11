`notion-cli auth`
================

Authentication commands — log in, log out, check status, and refresh OAuth tokens.

* [`notion-cli auth login`](#notion-cli-auth-login)
* [`notion-cli auth logout`](#notion-cli-auth-logout)
* [`notion-cli auth status`](#notion-cli-auth-status)
* [`notion-cli auth refresh`](#notion-cli-auth-refresh)

---

## `notion-cli auth login`

Authenticate with Notion via the OAuth 2.0 authorization flow.

```
USAGE
  $ notion-cli auth login [--manual] [-j]

FLAGS
  -j, --json     Output result as JSON
      --manual   Skip the local callback server and paste the redirected URL by hand

DESCRIPTION
  Opens your browser to the Notion authorization page and exchanges the
  returned code for an OAuth access token. The token is saved to the local
  config file (~/.config/notion-cli/config.json).

  Use --manual when running over SSH, in a container, or behind a firewall
  where the CLI's localhost callback server cannot be reached from the browser.
  In manual mode the authorization URL is printed; you open it yourself, then
  paste the full redirected URL (or bare code) back into the terminal.

  SSH sessions are automatically detected and force --manual mode.

EXAMPLES
  Log in via browser (default)

    $ notion-cli auth login

  Log in without a browser (SSH / container)

    $ notion-cli auth login --manual
```

---

## `notion-cli auth logout`

Remove stored OAuth tokens from the config file.

```
USAGE
  $ notion-cli auth logout [--local-only] [-j]

FLAGS
  -j, --json          Output result as JSON
      --local-only    Clear local config only; do not call the Notion revoke API

DESCRIPTION
  By default, the access token is revoked on the Notion API before being removed
  from the local config. If the revoke call fails (e.g. the token has already
  expired), a warning is printed but the local config is still cleared.

  Use --local-only to skip the API call entirely and only delete local credentials.

EXAMPLES
  Log out and revoke the token remotely

    $ notion-cli auth logout

  Clear local credentials without calling the Notion API

    $ notion-cli auth logout --local-only
```

---

## `notion-cli auth status`

Display the current authentication method and details.

```
USAGE
  $ notion-cli auth status [--remote] [-j]

FLAGS
  -j, --json     Output as JSON
      --remote   Also call the Notion token introspect API to verify active status

DESCRIPTION
  Prints the active authentication method (oauth, env, token, or none) along
  with masked token, workspace info, and — if an expiry is stored — the
  token expiry timestamp.

  With --remote, the CLI additionally calls the Notion token introspect
  endpoint and merges active status, scope, and issued_at into the output.
  Only works when authenticated via OAuth and OAuth client credentials are
  embedded in the build.

OUTPUT FIELDS (oauth method)
  auth_method      "oauth"
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
  $ notion-cli auth refresh [-j]

FLAGS
  -j, --json   Output result as JSON

DESCRIPTION
  Calls the Notion token endpoint with grant_type=refresh_token using the
  refresh token saved by 'auth login'. On success the new access token (and
  updated refresh token, if returned) are persisted to the local config file.

  This command requires that:
    1. You previously ran 'notion-cli auth login' (refresh token is stored).
    2. OAuth client credentials are embedded in this build of notion-cli.

  The CLI also performs this refresh automatically and transparently when any
  API call returns a 401 response and a refresh token is available — you do
  not normally need to run this command manually.

OUTPUT FIELDS
  auth_method      "oauth"
  workspace_id     Notion workspace UUID (if returned by Notion)
  workspace_name   Human-readable workspace name (if returned by Notion)
  expires_at       RFC3339 expiry of the new token (when Notion returns expires_in)

EXAMPLES
  Manually refresh the stored OAuth token

    $ notion-cli auth refresh

  Refresh and inspect the result as JSON

    $ notion-cli auth refresh --json
```
