# Publishing to npm

## 🚀 Automated Publishing (Recommended)

The project now has **automated npm publishing** via GitHub Actions!

### One-Time Setup (3 minutes)

#### 1. Create npm Access Token
1. Go to https://www.npmjs.com/settings/YOUR_USERNAME/tokens
2. Click "Generate New Token" → "Granular Access Token"
3. Fill out the form:
   - **Token name**: `GitHub Actions - notion-cli CI/CD`
   - **Description**: `Automated publishing token for CI/CD`
   - **✅ Bypass two-factor authentication (2FA)**: **CHECK THIS BOX** ← Critical for automation!
   - **Allowed IP ranges**: Leave empty (GitHub Actions uses dynamic IPs)
   - **Expiration**: 90 days (maximum for write tokens)
4. **Packages and scopes**:
   - Change permissions dropdown to: **"Read and write"**
   - Select package: `@coastal-programs/notion-cli`
5. **Organizations**: Leave as "No access"
6. Click "Generate token"
7. **Copy the token immediately** (starts with `npm_...`) - you won't see it again!

**⚠️ Important**: The "Bypass 2FA" checkbox is essential! Without it, automated publishing will fail with OTP errors even if you have a valid token.

#### 2. Add Token to GitHub Secrets
1. Go to https://github.com/Coastal-Programs/notion-cli/settings/secrets/actions
2. Click "New repository secret"
3. Name: `NPM_TOKEN`
4. Value: Paste your npm token
5. Click "Add secret"

That's it! Publishing is now automated. ✅

### How to Publish a New Version

**Option 1: GitHub Release (Recommended)**
```bash
# 1. Bump version locally
npm version patch  # 5.6.0 → 5.6.1 (bug fixes)
npm version minor  # 5.6.0 → 5.7.0 (new features)
npm version major  # 5.6.0 → 6.0.0 (breaking changes)

# 2. Push version tag to GitHub
git push --follow-tags

# 3. Create GitHub Release
# Go to: https://github.com/Coastal-Programs/notion-cli/releases/new
# - Tag: Select the tag you just pushed (e.g., v5.6.1)
# - Title: v5.6.1 - Production Polish
# - Description: Copy from CHANGELOG
# - Click "Publish release"

# 4. GitHub Action automatically publishes to npm! 🎉
```

**Option 2: Manual Trigger**
1. Go to Actions tab: https://github.com/Coastal-Programs/notion-cli/actions/workflows/publish.yml
2. Click "Run workflow"
3. Click "Run workflow" button
4. Workflow builds, tests, and publishes to npm automatically

### What Happens Automatically

When you create a GitHub Release:
- ✅ Runs full test suite
- ✅ Builds the project
- ✅ Checks if version already exists on npm
- ✅ Publishes to npm with provenance (secure)
- ✅ Shows success message with package URL

---

## 📝 Manual Publishing (Fallback)

If you prefer manual control or automation fails:

### 1. Create npm Account
```bash
# Sign up at https://www.npmjs.com/signup
# Then login locally:
npm login
```

### 2. Verify Package Name is Available
```bash
npm search @coastal-programs/notion-cli
# Should return no results or confirm your package
```

## Before First Publish (One-Time)

### Update README Installation Section
Update the README to include npm installation as the primary method:

```bash
# From npm (recommended)
npm install -g @coastal-programs/notion-cli

# Or from source
npm install -g Coastal-Programs/notion-cli
```

This should be done BEFORE your first publish so the README is ready when the package goes live.

---

## Before Each Release

### 1. Update Version
```bash
# Choose one:
npm version patch  # 5.6.0 -> 5.6.1 (bug fixes)
npm version minor  # 5.6.0 -> 5.7.0 (new features)
npm version major  # 5.6.0 -> 6.0.0 (breaking changes)
```

### 2. Test Build
```bash
npm run build
npm test
npm run lint
```

### 3. Test Installation Locally
```bash
npm pack
# Creates: coastal-programs-notion-cli-5.6.0.tgz
# Test it: npm install -g ./coastal-programs-notion-cli-5.6.0.tgz
```

## Publishing

### Publish to npm
```bash
npm publish --access public
```

That's it! Your package is live at:
- **Install**: `npm install -g @coastal-programs/notion-cli`
- **Page**: https://www.npmjs.com/package/@coastal-programs/notion-cli

## After Publishing

### Create GitHub Release
1. Go to: https://github.com/Coastal-Programs/notion-cli/releases
2. Click "Draft a new release"
3. Tag: `v5.6.0` (match your version)
4. Title: `v5.6.0 - Your Release Name`
5. Description: Copy from CHANGELOG or What's New section
6. Publish!

## Pro Tips

- **Dry run first**: `npm publish --dry-run` to test
- **Check files**: `npm pack --dry-run` shows what will be published
- **Scoped packages**: Already using `@coastal-programs/` scope ✅
- **Update often**: Users love frequent, small updates
- **Semantic versioning**: Follow semver.org strictly

## Common Issues

**"Package already exists"**
- Version already published. Update version number.

**"403 Forbidden"**
- Not logged in: `npm login`
- No access: `npm owner add <username> @coastal-programs/notion-cli`

**"Payment required"**
- Use `--access public` flag
- Scoped packages default to private

## Automation (Future)

You can automate with GitHub Actions:
```yaml
# .github/workflows/publish.yml
on:
  release:
    types: [published]

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '22.x'
          registry-url: 'https://registry.npmjs.org'
      - run: npm ci
      - run: npm run build
      - run: npm publish --access public
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```
