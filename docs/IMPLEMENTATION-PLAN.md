# Implementation Plan: Mac & Windows Installation (2025 Best Practices)

## Executive Summary

**Problem:** Current GitHub installation (`npm install -g Coastal-Programs/notion-cli`) creates broken symlinks on Windows.

**Solution:** Publish to npm registry following 2025 best practices. This provides universal Mac/Windows compatibility, better performance, and standard user expectations.

**Timeline:** Can be implemented immediately - package is already properly structured.

---

## Current State Analysis

### ✅ What's Already Working
- Package structure is correct (`@coastal-programs/notion-cli`)
- `bin` field properly configured for cross-platform executable
- `files` field correctly specifies what to publish
- Oclif framework handles platform differences automatically
- TypeScript compiled to compatible JavaScript in `dist/`
- Local installation works perfectly on Windows

### ❌ Current Problem
- Installing from GitHub creates broken Windows symlinks
- Not following 2025 standard (npm registry is expected)
- Requires workaround (local folder installation)
- Not discoverable via `npm search`

---

## Solution: Publish to npm Registry

### Why This Solves Everything

1. **Cross-Platform Compatibility**
   - npm automatically creates `.cmd` wrappers on Windows
   - npm creates proper symlinks on Mac/Linux
   - No platform-specific code needed
   - Tested by millions of packages

2. **2025 Standard Practice**
   - npm registry is the expected distribution method
   - GitHub installations are for development only
   - Better performance and caching
   - Semantic versioning support

3. **AI Agent Friendly**
   - Standard installation command everyone knows
   - Predictable, reliable behavior
   - No special configuration needed

### Installation Commands

**Current (broken on Windows):**
```bash
npm install -g Coastal-Programs/notion-cli
```

**After Publishing (works everywhere):**
```bash
npm install -g @coastal-programs/notion-cli
```

---

## Implementation Steps

### Phase 1: Pre-Publishing Checklist ✅

All requirements are already met:

- [x] Package name scoped: `@coastal-programs/notion-cli`
- [x] Version number set: `5.1.0`
- [x] `bin` field configured for executable
- [x] `files` field specifies what to publish
- [x] `dist/` folder committed and up to date
- [x] README.md complete with documentation
- [x] LICENSE file present
- [x] No security vulnerabilities

### Phase 2: npm Account Setup

```bash
# 1. Create npm account (if not exists)
# Visit: https://www.npmjs.com/signup

# 2. Login from command line
npm login

# 3. Verify login
npm whoami
```

### Phase 3: Test Package Contents

```bash
# See exactly what will be published
npm pack --dry-run

# Create actual tarball for inspection
npm pack
# This creates: coastal-programs-notion-cli-5.1.0.tgz

# Test installation from tarball
npm install -g ./coastal-programs-notion-cli-5.1.0.tgz

# Verify it works
notion-cli --version
notion-cli user retrieve bot --output json
```

### Phase 4: Publish to npm

```bash
# Publish as public package
npm publish --access public

# Verify it's published
npm view @coastal-programs/notion-cli
```

### Phase 5: Test Installation on Both Platforms

**Windows Test:**
```cmd
:: Uninstall local version
npm uninstall -g @coastal-programs/notion-cli

:: Install from npm registry
npm install -g @coastal-programs/notion-cli

:: Test
notion-cli --version
set NOTION_TOKEN=your-token-here
notion-cli user retrieve bot --output json
```

**Mac Test:**
```bash
# Uninstall local version
npm uninstall -g @coastal-programs/notion-cli

# Install from npm registry
npm install -g @coastal-programs/notion-cli

# Test
notion-cli --version
export NOTION_TOKEN=your-token-here
notion-cli user retrieve bot --output json
```

### Phase 6: Update Documentation

Update README.md installation section:

```markdown
## Installation

### Global Installation (Recommended)
```bash
npm install -g @coastal-programs/notion-cli
```

### Verify Installation
```bash
notion-cli --version
```

### Setup
```bash
export NOTION_TOKEN="your-notion-integration-token"
# On Windows: set NOTION_TOKEN=your-notion-integration-token
```
```

---

## Automated Setup Wizard (Future Enhancement)

After successful npm registry publishing, we can add an optional setup command:

### Implementation Plan

**Add `setup` command:**
```typescript
// src/commands/setup.ts
import {Command} from '@oclif/core'
import * as fs from 'fs'
import * as path from 'path'
import * as os from 'os'

export default class Setup extends Command {
  static description = 'Interactive setup wizard for notion-cli'

  async run() {
    // Check if NOTION_TOKEN already set
    if (process.env.NOTION_TOKEN) {
      this.log('✓ NOTION_TOKEN already configured')
      return
    }

    // Prompt for token (or detect non-interactive)
    if (!process.stdout.isTTY) {
      this.log('Non-interactive environment detected')
      this.log('Set NOTION_TOKEN environment variable to configure')
      return
    }

    const {default: inquirer} = await import('inquirer')
    const answers = await inquirer.prompt([{
      type: 'password',
      name: 'token',
      message: 'Enter your Notion Integration Token:',
      validate: (input: string) => input.length > 0
    }])

    // Offer to save to shell profile
    const {saveToProfile} = await inquirer.prompt([{
      type: 'confirm',
      name: 'saveToProfile',
      message: 'Save to shell profile for persistence?',
      default: true
    }])

    if (saveToProfile) {
      const shell = process.env.SHELL || 'bash'
      const profile = shell.includes('zsh') ? '.zshrc' : '.bashrc'
      const profilePath = path.join(os.homedir(), profile)

      fs.appendFileSync(profilePath,
        `\n# Notion CLI\nexport NOTION_TOKEN="${answers.token}"\n`)

      this.log(`✓ Token saved to ${profile}`)
      this.log('Run: source ~/${profile} (or restart terminal)')
    } else {
      this.log('\nRun this command to set token for current session:')
      this.log(`export NOTION_TOKEN="${answers.token}"`)
    }
  }
}
```

**Add dependency:**
```json
{
  "dependencies": {
    "inquirer": "^9.2.12"
  }
}
```

---

## Testing Strategy

### Automated Testing (GitHub Actions)

```yaml
# .github/workflows/test.yml
name: Test

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        node: [18, 20, 22]

    runs-on: ${{ matrix.os }}

    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node }}

      - name: Install dependencies
        run: npm ci

      - name: Build
        run: npm run build

      - name: Install globally
        run: npm install -g .

      - name: Test version command
        run: notion-cli --version

      - name: Test help command
        run: notion-cli --help

      - name: Test with API (if token available)
        if: env.NOTION_TOKEN != ''
        run: notion-cli user retrieve bot --output json
        env:
          NOTION_TOKEN: ${{ secrets.NOTION_TOKEN }}
```

### Manual Testing Checklist

**Windows:**
- [ ] Clean install from npm registry works
- [ ] `notion-cli` command available in any directory
- [ ] `--version` shows correct version
- [ ] `--help` displays documentation
- [ ] API calls work with NOTION_TOKEN set
- [ ] Updates via `npm update -g @coastal-programs/notion-cli` work

**Mac:**
- [ ] Clean install from npm registry works
- [ ] `notion-cli` command available in any directory
- [ ] `--version` shows correct version
- [ ] `--help` displays documentation
- [ ] API calls work with NOTION_TOKEN set
- [ ] Updates via `npm update -g @coastal-programs/notion-cli` work

---

## Success Criteria

### Primary Goals (Must Have)
- [x] Package structure correct
- [ ] Published to npm registry
- [ ] Works on Windows without symlink issues
- [ ] Works on Mac/Linux
- [ ] Single command installation
- [ ] Discoverable via `npm search notion cli`

### Secondary Goals (Nice to Have)
- [ ] Automated setup wizard
- [ ] GitHub Actions CI/CD testing
- [ ] Detailed troubleshooting guide
- [ ] Version update notifications

---

## Risk Assessment

### Low Risk
- ✅ Package already properly structured
- ✅ Local testing confirms functionality
- ✅ Oclif framework handles platform differences
- ✅ Standard npm publishing process

### No Breaking Changes
- Version is currently unpublished
- No existing users to impact
- Can test thoroughly before announcement

---

## Timeline

### Immediate (Today)
1. Create npm account / login
2. Test with `npm pack`
3. Publish to registry
4. Test on Windows
5. Test on Mac (if available)
6. Update README with new installation instructions

### Short Term (This Week)
1. Add GitHub Actions CI/CD
2. Create detailed installation guide
3. Add troubleshooting section to README
4. Test with AI agents

### Medium Term (Next Sprint)
1. Implement setup wizard
2. Add shell completion scripts
3. Create installation video/guide
4. Publish announcement

---

## Rollback Plan

If issues arise after publishing:

```bash
# Unpublish version (within 72 hours)
npm unpublish @coastal-programs/notion-cli@5.1.0

# Or deprecate version
npm deprecate @coastal-programs/notion-cli@5.1.0 "Use version X.X.X instead"
```

---

## Additional Enhancements (Future)

### Security (2025 Standards)
- Set up Trusted Publishing with OIDC
- Enable 2FA on npm account
- Add socket.dev scanning
- Implement provenance attestations

### Distribution Alternatives
- Homebrew formula for Mac
- Chocolatey package for Windows
- Docker image for containerized usage
- Standalone binaries (pkg or ncc)

### Developer Experience
- Shell completion (bash, zsh, fish)
- Man pages
- Update notifications
- Configuration file support (~/.notionrc)

---

## Conclusion

**The package is ready to publish right now.** All required structure is in place. Publishing to npm registry will:

1. ✅ Fix Windows installation issues
2. ✅ Work perfectly on Mac
3. ✅ Follow 2025 best practices
4. ✅ Make it AI-agent friendly
5. ✅ Enable standard npm workflows

**Next Action:** Run `npm login` and `npm publish --access public`
