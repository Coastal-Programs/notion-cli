# Security Policy

## Supported Versions

We actively support the following versions with security updates:

| Version | Supported          | Notes                                      |
| ------- | ------------------ | ------------------------------------------ |
| 6.x     | :white_check_mark: | Current release (Go rewrite), fully supported |
| 5.x     | :x:                | Legacy TypeScript version, no longer maintained |
| < 5.0   | :x:                | No longer supported                        |

## Supply Chain Security

### v6.x (Go Binary)

The v6.0.0 rewrite from TypeScript to Go dramatically reduced the attack surface:

- **2 runtime dependencies**: `github.com/spf13/cobra` and `github.com/spf13/pflag`
- **Single static binary**: ~8MB, no interpreter or runtime required
- **No npm production dependencies**: The npm package is a thin wrapper that downloads and runs the platform-specific Go binary
- **No external code execution**: The CLI does not evaluate user-provided code, load plugins, or shell out to external programs
- **Cross-compiled binaries**: Built from source for darwin/amd64, darwin/arm64, linux/amd64, linux/arm64, and windows/amd64

Compared to v5.x (573 npm dependencies), v6.x has a near-zero supply chain risk profile.

## Data Storage and File Permissions

### Configuration File

- **Path**: `~/.config/notion-cli/config.json`
- **Permissions**: `0o600` (owner read/write only)
- **Contents**: Notion API token and user preferences
- **Atomic writes**: Configuration is written to a temporary file and atomically renamed to prevent corruption

### Workspace Cache

- **Path**: `~/.notion-cli/databases.json`
- **Permissions**: `0o600` (owner read/write only)
- **Directory permissions**: `0o700` (owner access only)
- **Contents**: Cached database metadata (IDs, titles, aliases) -- no page content or sensitive data
- **Atomic writes**: Cache is written to a temporary file and atomically renamed to prevent corruption

### Token Handling

- **Token masking**: All CLI output masks tokens by default, displaying only the prefix and last 3 characters (e.g., `secret_***...***abc`)
- **Opt-in reveal**: The `--show-secret` flag is required to display the full token value
- **Environment variable**: `NOTION_TOKEN` is the primary token source; it takes precedence over the config file
- **No token logging**: Tokens are never written to log files or included in error reports

## Reporting a Vulnerability

**Please do not report security vulnerabilities through public GitHub issues.**

### Reporting Process

1. **Email:** Send details to `jake@coastalprograms.com`

2. **Include:**
   - Type of vulnerability
   - Full paths of source files related to the vulnerability
   - Location of affected code (tag/branch/commit)
   - Step-by-step instructions to reproduce
   - Proof-of-concept or exploit code (if possible)
   - Impact of the vulnerability
   - Suggested fix (if you have one)

3. **Response Time:**
   - Initial response: Within 48 hours
   - Status update: Within 7 days
   - Resolution timeline: Depends on severity

### What to Expect

1. **Acknowledgment:** We will confirm receipt of your report within 48 hours
2. **Investigation:** We will investigate and validate the vulnerability
3. **Updates:** You will receive updates on our progress every 7 days
4. **Resolution:** We will work on a fix and coordinate release timing
5. **Credit:** You will be credited in release notes (unless you prefer anonymity)

### Disclosure Policy

- We follow **coordinated disclosure**
- Please allow us reasonable time to address the issue before public disclosure
- We aim to release security fixes within 90 days of initial report
- Critical vulnerabilities may be expedited

## Security Best Practices for Users

### Token Management

**NEVER commit your Notion token to version control.**

```bash
# BAD - Don't do this
export NOTION_TOKEN=secret_abc123...
git add .env
git commit -m "Add config"

# GOOD - Use environment variables
export NOTION_TOKEN=secret_abc123...  # In shell session only
# Or add to ~/.bashrc, ~/.zshrc (never commit)

# GOOD - Use the built-in config command
notion-cli config set-token
# Stores token in ~/.config/notion-cli/config.json with 0o600 permissions
```

**Best Practices:**

1. **Use environment variables** for tokens, never hardcode them
2. **Add `.env` to `.gitignore`** if using env files
3. **Rotate tokens regularly** (every 90 days recommended)
4. **Use minimal permissions** -- only grant integration access to needed databases
5. **Revoke unused tokens** at https://www.notion.so/my-integrations

### Integration Permissions

Follow the **principle of least privilege**:

1. Create an integration at https://www.notion.so/my-integrations
2. Share ONLY the databases you need to access
3. Review permissions regularly
4. Do not share your entire workspace if you only need a few databases

### Verifying the Binary

After installing via npm, verify the binary is the expected one:

```bash
# Check version and build info
notion-cli --version

# Verify the binary path
which notion-cli

# Run health check
notion-cli doctor
```

### CI/CD Security

When using in CI/CD pipelines:

```yaml
# GOOD - Use encrypted secrets
env:
  NOTION_TOKEN: ${{ secrets.NOTION_TOKEN }}

# BAD - Never expose tokens in logs
- run: echo "Token: $NOTION_TOKEN"  # DON'T DO THIS
```

**Best Practices:**

1. Store tokens in encrypted CI/CD secrets
2. Never print tokens in logs
3. Use read-only tokens when possible
4. Audit CI/CD logs for accidental token exposure
5. Rotate tokens if exposed in logs

## Secure Development

### For Contributors

If you are contributing code:

1. **Never commit secrets** -- Use environment variables
2. **Validate all inputs** -- Sanitize user input via the resolver and error packages
3. **Use `context.Context`** for all API calls to support timeouts and cancellation
4. **Follow least privilege** -- Minimize API permissions in integration code
5. **Keep dependencies minimal** -- The project intentionally uses only 2 Go dependencies
6. **Write security tests** -- Test authentication, authorization, and input validation
7. **Use `internal/errors.NotionCLIError`** -- Never expose raw errors that could leak sensitive information

### Code Review Checklist

Security considerations for code reviewers:

- [ ] No hardcoded credentials or tokens
- [ ] Input validation for user-provided data (IDs, URLs, JSON)
- [ ] Error messages do not leak sensitive information
- [ ] File operations use restrictive permissions (0o600 for files, 0o700 for directories)
- [ ] No command injection vulnerabilities (no `os/exec` with user input)
- [ ] Token values are masked in all output paths
- [ ] Atomic file writes used for persistent data

## Vulnerability Disclosure Timeline

**Example Timeline for Critical Vulnerability:**

- **Day 0:** Vulnerability reported
- **Day 1:** Acknowledgment sent to reporter
- **Day 2-7:** Investigation and validation
- **Day 8-14:** Develop and test fix
- **Day 15:** Release security patch
- **Day 16:** Public disclosure (coordinated with reporter)

**Example Timeline for Low/Moderate Vulnerability:**

- **Day 0:** Vulnerability reported
- **Day 1-2:** Acknowledgment and initial assessment
- **Day 3-30:** Investigation and fix development
- **Day 31-60:** Testing and release preparation
- **Day 61:** Release with next scheduled version
- **Day 62:** Public disclosure

## Contact

**Security Issues:** jake@coastalprograms.com

**General Issues:** https://github.com/Coastal-Programs/notion-cli/issues

**Discussions:** https://github.com/Coastal-Programs/notion-cli/discussions

## Additional Resources

- [Notion Security Best Practices](https://developers.notion.com/docs/security)
- [Go Security Best Practices](https://go.dev/doc/security/best-practices)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CVE Database](https://cve.mitre.org/)

---

**Last Updated:** 2026-03-01
**Next Review:** 2026-06-01
