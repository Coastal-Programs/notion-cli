# Security Policy

## Supported Versions

We actively support the following versions with security updates:

| Version | Supported          | Notes                          |
| ------- | ------------------ | ------------------------------ |
| 5.6.x   | :white_check_mark: | Current release, fully supported |
| 5.5.x   | :white_check_mark: | Previous release, critical fixes only |
| 5.4.x   | :x:                | Upgrade to 5.6.x recommended |
| < 5.4   | :x:                | No longer supported |

## Security Status

### Current Release (v5.6.0)

- **Production Dependencies:** 0 vulnerabilities
- **Development Dependencies:** 11 low/moderate vulnerabilities (non-critical)
- **Last Security Audit:** 2025-10-25

### Fixed Vulnerabilities

The following vulnerabilities were resolved in recent releases:

**v5.5.0 (2025-10-24):**
- CVE-2023-48618: katex XSS vulnerability (removed @tryfabric/martian)
- CVE-2024-28245: katex XSS vulnerability (removed @tryfabric/martian)
- CVE-2021-23906: yargs-parser prototype pollution
- CVE-2020-28469: glob-parent ReDoS vulnerability

**v5.6.0 (2025-10-25):**
- 14 development dependency vulnerabilities
- Zero critical or high-severity issues remaining in production

### Known Issues

**Development Dependencies (Non-Critical):**
- 2 moderate severity vulnerabilities in oclif v2 dependencies
- 9 low severity vulnerabilities in test infrastructure
- **Impact:** Development environment only, no production runtime impact
- **Status:** Tracked for resolution in future oclif v4 migration

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

1. **Acknowledgment:** We'll confirm receipt of your report within 48 hours
2. **Investigation:** We'll investigate and validate the vulnerability
3. **Updates:** You'll receive updates on our progress every 7 days
4. **Resolution:** We'll work on a fix and coordinate release timing
5. **Credit:** You'll be credited in release notes (unless you prefer anonymity)

### Disclosure Policy

- We follow **coordinated disclosure**
- Please allow us reasonable time to address the issue before public disclosure
- We aim to release security fixes within 90 days of initial report
- Critical vulnerabilities may be expedited

## Security Best Practices for Users

### Token Management

**NEVER commit your Notion token to version control!**

```bash
# ❌ BAD - Don't do this
export NOTION_TOKEN=secret_abc123...
git add .env
git commit -m "Add config"

# ✅ GOOD - Use environment variables
export NOTION_TOKEN=secret_abc123...  # In shell session only
# Or add to ~/.bashrc, ~/.zshrc (never commit)
```

**Best Practices:**

1. **Use environment variables** for tokens, never hardcode
2. **Add `.env` to `.gitignore`** if using env files
3. **Rotate tokens regularly** (every 90 days recommended)
4. **Use minimal permissions** - only grant integration access to needed databases
5. **Revoke unused tokens** at https://www.notion.so/my-integrations

### Integration Permissions

Follow the **principle of least privilege**:

```bash
# ✅ GOOD - Share only specific databases
1. Create integration at https://www.notion.so/my-integrations
2. Share ONLY the databases you need to access
3. Review permissions regularly

# ❌ BAD - Sharing entire workspace unnecessarily
Don't share your entire workspace if you only need a few databases
```

### Secure Usage

1. **Verify package integrity** before installation:
   ```bash
   npm install @coastal-programs/notion-cli --dry-run
   ```

2. **Use specific versions** in production:
   ```json
   {
     "dependencies": {
       "@coastal-programs/notion-cli": "5.6.0"
     }
   }
   ```

3. **Review audit reports** regularly:
   ```bash
   npm audit
   ```

4. **Keep updated** to latest version:
   ```bash
   npm update @coastal-programs/notion-cli
   ```

### CI/CD Security

When using in CI/CD pipelines:

```yaml
# ✅ GOOD - Use encrypted secrets
env:
  NOTION_TOKEN: ${{ secrets.NOTION_TOKEN }}

# ❌ BAD - Never expose tokens in logs
- run: echo "Token: $NOTION_TOKEN"  # DON'T DO THIS
```

**Best Practices:**

1. Store tokens in encrypted CI/CD secrets
2. Never print tokens in logs
3. Use read-only tokens when possible
4. Audit CI/CD logs for accidental token exposure
5. Rotate tokens if exposed in logs

## Security Audit History

### Recent Audits

| Date       | Tool      | Critical | High | Moderate | Low | Notes |
|------------|-----------|----------|------|----------|-----|-------|
| 2025-10-25 | npm audit | 0        | 0    | 2        | 9   | DevDeps only |
| 2025-10-24 | npm audit | 0        | 0    | 0        | 0   | All production vulns fixed |
| 2025-10-23 | npm audit | 1        | 3    | 18       | 4   | Pre-v5.5.0 baseline |

### Continuous Monitoring

We continuously monitor for security vulnerabilities using:

- **npm audit** - Automated dependency scanning
- **Dependabot** - Automated dependency updates
- **GitHub Security Advisories** - CVE monitoring
- **Manual code review** - Security-focused code reviews

## Secure Development

### For Contributors

If you're contributing code:

1. **Never commit secrets** - Use environment variables
2. **Validate all inputs** - Sanitize user input
3. **Use parameterized queries** - Prevent injection attacks
4. **Follow least privilege** - Minimize API permissions
5. **Keep dependencies updated** - Run `npm audit` regularly
6. **Write security tests** - Test authentication, authorization, input validation

### Code Review Checklist

Security considerations for code reviewers:

- [ ] No hardcoded credentials or tokens
- [ ] Input validation for user-provided data
- [ ] Error messages don't leak sensitive information
- [ ] Dependencies are up to date
- [ ] No SQL/command injection vulnerabilities
- [ ] Proper authentication/authorization checks
- [ ] Sensitive data not logged

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
- [npm Security Best Practices](https://docs.npmjs.com/security-best-practices)
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CVE Database](https://cve.mitre.org/)

---

**Last Updated:** 2025-10-25
**Next Review:** 2026-01-25
