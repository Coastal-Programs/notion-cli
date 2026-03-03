---
argument-hint: [version-bump: patch|minor|major]
description: Full release workflow - bump version, test, build, publish to npm + GitHub
allowed-tools: Bash, Read, Write, Edit, Glob, Grep
---

# Release: notion-cli v$ARGUMENTS

Execute the complete release workflow for notion-cli. This handles version bumping, testing, cross-compilation, npm publishing, and GitHub Release creation.

## CRITICAL LESSONS (from past failures)

### npm Token Requirements
- **Classic Automation token** is REQUIRED for publishing. Granular tokens CANNOT create new packages (404 on PUT).
- The token must have **"Bypass 2FA"** enabled — without this, automated publishing fails with OTP errors.
- Token is stored as `NPM_TOKEN` in GitHub repo secrets AND can be used locally via `NPM_TOKEN=npm_... npm publish`.
- If token is expired or missing, guide the user to: https://www.npmjs.com/settings/jakeschepis/tokens → "Generate New Token" → **Classic** → **Automation** type.

### Platform Packages
- There are 5 platform-specific npm packages under `@coastal-programs/` scope. They MUST be published BEFORE the main wrapper package.
- If a platform package has NEVER been published before, you NEED a Classic Automation token (granular tokens return 404 for new packages).
- Package publish order: platform packages first, then main wrapper last.

### GitHub Release
- Use `--clobber` flag with `gh release upload` to handle re-uploads.
- The CI workflow (`publish.yml`) triggers on GitHub Release creation — if publishing locally, you may want to skip creating the release until after npm publish succeeds.

---

## Step 1: Pre-flight Checks

Run all checks FIRST. Do not proceed if any fail.

```bash
make test    # All Go tests must pass
make build   # Binary must compile
make lint    # No lint errors
```

Verify you're on the main branch and working tree is clean:
```bash
git status
git branch --show-current
```

---

## Step 2: Determine Version Bump

The user specified: `$ARGUMENTS`

- **patch** (6.2.0 → 6.2.1): bug fixes only
- **minor** (6.2.0 → 6.3.0): new features, backwards compatible
- **major** (6.2.0 → 7.0.0): breaking changes

Read the current version:
```bash
grep '"version"' package.json | head -1
```

---

## Step 3: Bump Version in ALL Package Files

**You must update ALL 6 package.json files to the same version.** Missing one causes install failures.

Files to update:
1. `package.json` (root — version AND optionalDependencies versions)
2. `npm/notion-cli-darwin-arm64/package.json`
3. `npm/notion-cli-darwin-x64/package.json`
4. `npm/notion-cli-linux-x64/package.json`
5. `npm/notion-cli-linux-arm64/package.json`
6. `npm/notion-cli-win32-x64/package.json`

Use the Edit tool to update each file. Double-check that the `optionalDependencies` in root `package.json` all reference the new version.

---

## Step 4: Update CHANGELOG.md

Add a new section at the top of the changelog under `## [Unreleased]`:

```markdown
## [X.Y.Z] - YYYY-MM-DD

### Added/Changed/Fixed
- Description of changes
```

---

## Step 5: Cross-Compile Binaries

```bash
make release
```

This produces 5 binaries in `build/`:
- `notion-cli-darwin-arm64`
- `notion-cli-darwin-amd64`
- `notion-cli-linux-amd64`
- `notion-cli-linux-arm64`
- `notion-cli-windows-amd64.exe`

Verify all 5 exist:
```bash
ls -la build/
```

---

## Step 6: Copy Binaries to Platform Packages

```bash
mkdir -p npm/notion-cli-darwin-arm64/bin
mkdir -p npm/notion-cli-darwin-x64/bin
mkdir -p npm/notion-cli-linux-x64/bin
mkdir -p npm/notion-cli-linux-arm64/bin
mkdir -p npm/notion-cli-win32-x64/bin

cp build/notion-cli-darwin-arm64 npm/notion-cli-darwin-arm64/bin/notion-cli
cp build/notion-cli-darwin-amd64 npm/notion-cli-darwin-x64/bin/notion-cli
cp build/notion-cli-linux-amd64 npm/notion-cli-linux-x64/bin/notion-cli
cp build/notion-cli-linux-arm64 npm/notion-cli-linux-arm64/bin/notion-cli
cp build/notion-cli-windows-amd64.exe npm/notion-cli-win32-x64/bin/notion-cli.exe
```

---

## Step 7: Publish to npm

### Check npm auth
```bash
npm whoami
```

If not logged in, run `npm login`. If using a token: `NPM_TOKEN=npm_... npm publish`.

### Check if version already exists
```bash
npm view @coastal-programs/notion-cli@X.Y.Z version 2>/dev/null && echo "EXISTS" || echo "NEW"
```

### Publish platform packages FIRST (order matters)
```bash
cd npm/notion-cli-darwin-arm64 && npm publish --access public && cd ../..
cd npm/notion-cli-darwin-x64 && npm publish --access public && cd ../..
cd npm/notion-cli-linux-x64 && npm publish --access public && cd ../..
cd npm/notion-cli-linux-arm64 && npm publish --access public && cd ../..
cd npm/notion-cli-win32-x64 && npm publish --access public && cd ../..
```

### Then publish main wrapper
```bash
npm publish --access public
```

### If npm publish fails with 403/OTP/auth errors:
1. Check token type — MUST be **Classic Automation** (not Granular)
2. Check token has **"Bypass 2FA"** enabled
3. Check token has write access to `@coastal-programs` scope
4. For brand new packages that have never been published: Granular tokens will NOT work (404 on PUT). Must use Classic.
5. Guide user to create new token: https://www.npmjs.com/settings/jakeschepis/tokens

---

## Step 8: Git Commit, Tag, and Push

```bash
git add package.json npm/*/package.json CHANGELOG.md
git commit -m "chore: release vX.Y.Z"
git tag -a vX.Y.Z -m "Release vX.Y.Z"
git push --follow-tags
```

---

## Step 9: Create GitHub Release

```bash
gh release create vX.Y.Z \
  --title "vX.Y.Z - Release Title" \
  --notes "$(cat <<'EOF'
## What's New
- Summary of changes

See [CHANGELOG.md](CHANGELOG.md) for full details.
EOF
)" \
  build/notion-cli-darwin-arm64 \
  build/notion-cli-darwin-amd64 \
  build/notion-cli-linux-amd64 \
  build/notion-cli-linux-arm64 \
  build/notion-cli-windows-amd64.exe \
  --clobber
```

---

## Step 10: Verify

```bash
# Check npm
npm view @coastal-programs/notion-cli version

# Check GitHub Release
gh release view vX.Y.Z

# Test install (optional)
npm install -g @coastal-programs/notion-cli@latest
notion-cli --version
```

---

## Troubleshooting Quick Reference

| Problem | Cause | Fix |
|---------|-------|-----|
| `npm publish` → 404 on PUT | Granular token + new package | Use Classic Automation token |
| `npm publish` → 403 Forbidden | Not logged in or no scope access | `npm login` or check token scope |
| `npm publish` → OTP required | Token doesn't bypass 2FA | Create new Classic Automation token with "Bypass 2FA" |
| `gh release upload` → asset exists | Binary already attached | Add `--clobber` flag |
| Version mismatch after install | Platform package.json not bumped | Update ALL 6 package.json files |
| CI publish fails | `NPM_TOKEN` secret expired/wrong | Update secret in GitHub repo settings |
