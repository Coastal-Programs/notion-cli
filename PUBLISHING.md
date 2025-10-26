# Publishing to npm

## One-Time Setup (5 minutes)

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

### Update README
Replace this in README.md:
```bash
# From source (recommended until npm package is published)
npm install -g Coastal-Programs/notion-cli
```

With:
```bash
# From npm (recommended)
npm install -g @coastal-programs/notion-cli

# Or from source
npm install -g Coastal-Programs/notion-cli
```

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
