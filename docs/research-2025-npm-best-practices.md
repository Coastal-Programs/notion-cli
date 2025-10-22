# 2025 Best Practices for npm Package Distribution and CLI Tools

Research Date: October 22, 2025
Target Project: notion-cli (Oclif-based CLI tool)

## Executive Summary

In 2025, the npm ecosystem has matured significantly with enhanced security through trusted publishing, improved TypeScript/ESM support, and standardized cross-platform CLI distribution patterns. The npm registry remains the dominant standard for public CLI tool distribution, with GitHub Packages reserved primarily for private/internal use. Key trends include mandatory 2FA, OIDC-based authentication, dual ESM/CJS publishing challenges, and the rise of tools like tsup and socket.dev for security scanning.

---

## 1. npm Registry Publishing Workflow (2025)

### 1.1 Current npm Version Status

- **npm v10.x**: Default package manager bundled with Node.js (stable, widely deployed)
- **npm v11.x**: Released December 16, 2024, with v11.6.2 as of October 8, 2025
  - New "type" prompt in `npm init`
  - Automatic "latest" dist tag management on publish
  - Requires published version to exceed latest semver version (excluding pre-release tags)
  - Improved package.json entry sorting

### 1.2 Major Security Changes (2025)

**Critical Authentication Changes (Effective October 13, 2025):**

1. **Token Lifetime Limits:**
   - Classic tokens: 90-day maximum lifetime
   - Granular tokens: 7-day maximum lifetime
   - All classic tokens will be revoked in November 2025

2. **2FA Requirements:**
   - TOTP 2FA becomes mandatory for most publishing workflows
   - Legacy authentication methods being phased out

3. **Trusted Publishing (Generally Available July 2025):**
   - Uses OpenID Connect (OIDC) to verify package origin
   - Issues short-lived tokens automatically during CI/CD
   - Requires npm CLI v11.5.1 or later
   - Available for GitHub Actions and GitLab CI/CD
   - **Automatic provenance attestations** without `--provenance` flag
   - Implements OpenSSF trusted publishers standard
   - Every package includes cryptographic proof of source and build environment

**Publishing Options in 2025:**
- Local publishing with 2FA
- Granular tokens with 7-day lifetime
- **Trusted publishing (RECOMMENDED)** - OIDC-based, no long-lived tokens

**Limitations:**
- Provenance unavailable when publishing from private source repositories
- Opt-out: `NPM_CONFIG_PROVENANCE=false`

### 1.3 Recommended Publishing Workflow (2025)

```bash
# 1. Set up trusted publishing in npm registry (one-time setup)
# Configure your GitHub repo/workflow in npm dashboard

# 2. GitHub Actions workflow with trusted publishing
name: Publish to npm
on:
  release:
    types: [published]

jobs:
  publish:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write  # Required for OIDC
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          registry-url: 'https://registry.npmjs.org'
      - run: npm ci
      - run: npm publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}  # Only if not using trusted publishing
```

**Key Improvements:**
- No `--provenance` flag needed with trusted publishing
- Automatic SLSA attestations
- Enhanced supply chain security
- Integration with npm audit and socket.dev

---

## 2. Modern Cross-Platform CLI Distribution

### 2.1 npm Registry vs GitHub Packages

**Status in 2025:**

| Aspect | npm Registry | GitHub Packages |
|--------|-------------|-----------------|
| **Primary Use Case** | Public CLI tools & packages | Private/internal packages |
| **Authentication** | Multiple options (tokens, trusted publishing) | PAT only (classic) |
| **Default Visibility** | Public | Private |
| **Scope Handling** | Flexible | Locked to GitHub org (`@org/package`) |
| **Proxying** | N/A | Does NOT proxy to npm registry |
| **CLI Distribution** | **STANDARD (2025)** | Limited (internal use) |

**Critical GitHub Packages Limitation:**
```bash
# PROBLEM: Can't mix scoped packages from different registries
# If @myorg/public-pkg is on npmjs.com
# And @myorg/private-pkg is on GitHub Packages
# You CANNOT use both in the same project
# Because .npmrc routes ALL @myorg/* to one registry
```

**Verdict:** **npm registry is the 2025 standard for CLI distribution**

### 2.2 Direct Git Installation vs npm Registry

**Search Results:** No evidence of deprecation of direct git installations in 2025.

**Current Support:**
```json
{
  "dependencies": {
    "my-package": "github:username/repo",
    "another-package": "git+https://github.com/user/repo.git#v1.2.3"
  }
}
```

**npm's Position:**
- npm remains committed to supporting git dependencies
- Continuously improving git installation performance
- npm registry offers better performance (semantic versioning, CDN)

**Best Practice for CLI Tools:**
- **Development/Testing:** Git installation acceptable
- **Production/Public Distribution:** Publish to npm registry
  - Better performance
  - Semantic versioning support
  - Wider compatibility
  - Better user experience (`npm install -g package-name`)

**Current notion-cli Installation:**
```bash
# Current (git-based)
npm install -g Coastal-Programs/notion-cli

# Recommended (after npm publish)
npm install -g @coastal-programs/notion-cli
```

### 2.3 Cross-Platform Binary Distribution Methods

**Option 1: npm's Automatic Wrapper (RECOMMENDED for Node.js CLIs)**
```json
{
  "bin": {
    "notion-cli": "./bin/run"
  }
}
```
- npm automatically creates platform-specific wrappers
- `.cmd` files on Windows
- Shell scripts on Unix/Mac
- **No additional work required**

**Option 2: Platform-Specific Binaries (Native Code)**
```json
{
  "optionalDependencies": {
    "@myapp/darwin-x64": "^1.0.0",
    "@myapp/win32-x64": "^1.0.0",
    "@myapp/linux-x64": "^1.0.0"
  }
}
```
- Use `process.platform` to load correct binary
- Publish each platform as separate optional dependency
- Requires postinstall script as backup strategy
- Tools: pkg, nexe, node-pre-gyp

**Option 3: Universal Packages**
- ShellJS: Cross-platform Unix shell commands on Node.js API
- execa: Cross-platform child_process
- cross-spawn: Cross-platform spawn
- cross-env: Cross-platform environment variables
- cross-os: Platform-specific npm scripts

### 2.4 Shebang Handling (2025)

**Universal Node.js Shebang:**
```javascript
#!/usr/bin/env node
// Your CLI code here
```

**Windows Behavior:**
- Windows does NOT support shebangs natively
- npm creates `.cmd` wrapper files automatically
- Example: `ng.cmd` runs `node ng` under the hood
- **No special handling needed when using npm's bin field**

**Two-Line Shebang (Unix/Linux compatibility):**
```bash
#!/bin/sh
':' //; exec "$(command -v nodejs || command -v node)" "$0" "$@"
```
- Supports both `node` and `nodejs` executables
- Only needed for edge cases (not typical npm packages)

**Best Practice:** Use standard `#!/usr/bin/env node` and let npm handle platform differences

---

## 3. Windows vs Mac Installation Differences (2024-2025)

### 3.1 Core Behavioral Differences

| Aspect | Windows | Mac/Linux |
|--------|---------|-----------|
| **Package Manager** | MSI installer, winget, npm | Homebrew, npm, nvm |
| **Installation Speed** | Slower, occasional lock issues | Faster, more reliable |
| **Help System** | Opens browser (HTML docs) | Man pages in terminal |
| **Firewall** | Prompts for Node.js apps | Rare firewall warnings |
| **Shebang Support** | No native support | Native support |
| **npm Wrapper** | .cmd batch files | Shell scripts |
| **File Paths** | Backslashes (\\) | Forward slashes (/) |

### 3.2 npm Installation Issues

**Windows Challenges:**
- Postinstall scripts may hang or never complete
- File lock issues (code doesn't release locks)
- Slower installation overall
- Firewall warnings during first run

**Mitigations:**
- Use npm workspaces for monorepos (better on Windows)
- Prefer pure JavaScript over native modules
- Use cross-platform packages (shx, cross-env, rimraf)
- Test on Windows in CI/CD

### 3.3 Node Version Managers (2025)

**nvm (Mac/Linux) / nvm-windows (Windows):**
```bash
# Install specific version
nvm install 20.10.0
nvm use 20.10.0

# Install latest LTS
nvm install --lts
```

**Homebrew (Mac):**
```bash
# Automatically gets latest LTS Node.js + npm
brew install node
```

**Best Practice:** Specify minimum Node.js version in package.json
```json
{
  "engines": {
    "node": ">=18.0.0"  // notion-cli requirement (Notion API v5.2.1 compatibility)
  }
}
```

---

## 4. TypeScript and ESM Module Best Practices (2025)

### 4.1 The Current State: "Still a Mess"

According to Liran Tal's 2025 analysis, **"TypeScript in 2025 with ESM and CJS npm publishing is still a mess"**. However, significant progress has been made.

### 4.2 Node.js v22/v23 Game Changer (2025)

**Major Update:** Node.js v22 and v23 added native support for CommonJS modules to require ESM modules

```javascript
// Now works natively in Node.js v22+
const esmModule = require('./esm-module.mjs');
```

**Impact:**
- Reduces need for complex dual publishing
- Simplifies package.json exports configuration
- Still requires careful management for Node.js <v22 compatibility

### 4.3 ESM-Only Package Configuration (Recommended)

```json
{
  "type": "module",
  "exports": "./index.js",
  "engines": {
    "node": ">=18.0.0"
  }
}
```

**TypeScript Configuration (tsconfig.json):**
```json
{
  "compilerOptions": {
    "module": "node16",
    "moduleResolution": "node16",  // MUST be node16 or nodenext (NOT "node")
    "target": "ES2022",
    "outDir": "./dist",
    "declaration": true
  }
}
```

**Important:** Must use `.js` extension in imports even when importing `.ts` files:
```typescript
// In TypeScript source
import { helper } from './utils.js';  // Note: .js not .ts
```

### 4.4 Dual Publishing (ESM + CJS)

**When to Avoid:**
- Creates "Dual Package Hazard" risk
- Can cause extremely confusing bugs in consuming projects
- Different instances of your package loaded as ESM and CJS
- State synchronization issues

**When Required:**
- Supporting older Node.js versions (<v22)
- Wide ecosystem compatibility
- Popular libraries with legacy users

**Recommended Tool: tsup**
```bash
npm install -D tsup

# Build command
tsup src/index.ts --format cjs,esm --dts --clean
```

**package.json exports (Dual Publishing):**
```json
{
  "type": "module",
  "exports": {
    ".": {
      "types": "./dist/index.d.ts",      // ALWAYS FIRST
      "import": "./dist/index.mjs",
      "require": "./dist/index.cjs",
      "default": "./dist/index.mjs"      // Fallback for non-Node.js
    }
  },
  "main": "./dist/index.cjs",
  "module": "./dist/index.mjs",
  "types": "./dist/index.d.ts"
}
```

**Avoiding Dual Package Hazard:**
```json
{
  "exports": {
    ".": {
      "node": "./dist/index.cjs",     // Node.js always gets this
      "default": "./dist/index.mjs"   // Others get this
    }
  }
}
```

### 4.5 Build Tools Comparison (2025)

| Tool | Use Case | ESM Support | CJS Support | DTS Generation |
|------|----------|-------------|-------------|----------------|
| **tsup** | Simple dual publishing | Yes | Yes | Yes |
| **tshy** | Opinionated dual build | Yes | Yes | Yes |
| **tsc** | Direct TypeScript compilation | Yes | Yes | Yes |
| **esbuild** | Fast bundling | Yes | Yes | No (use tsup) |

**Recommended:** tsup for most projects, tsc for simple ESM-only

### 4.6 Testing ESM Packages (2025)

**Recommended:** Vitest
```bash
npm install -D vitest
```

**Why Vitest over Jest:**
- Native ESM and TypeScript support
- No complex configuration
- Better integration with modern tooling
- Faster test execution

**TypeScript Execution:**
- **tsx**: Recommended for running TypeScript files
- **ts-node**: Still used for tests (legacy)

### 4.7 Type Validation Tool

**arethetypeswrong:** Analyzes npm packages for TypeScript type issues
```bash
npx arethetypeswrong --pack .
```

Detects:
- ESM-related module resolution issues
- Incorrect exports configuration
- Missing or wrong type declarations

---

## 5. Modern CLI Framework Trends (2025)

### 5.1 Framework Comparison

| Framework | Best For | Ecosystem Size | Learning Curve | Plugin System |
|-----------|----------|----------------|----------------|---------------|
| **Oclif** | Enterprise CLIs | Large | Medium | Yes (robust) |
| **Commander.js** | Simple CLIs | Very Large | Low | No (DIY) |
| **Yargs** | Declarative syntax | Large | Low-Medium | Limited |
| **Ink** | React-based TUIs | Growing | Medium (React knowledge) | React ecosystem |

### 5.2 Oclif (Current choice for notion-cli)

**Strengths:**
- Enterprise-grade with plugin architecture
- Excellent for large-scale CLIs (Git-style subcommands)
- TypeScript-first
- Built-in help generation
- Auto-generated documentation
- Well-maintained by Salesforce

**Trade-offs:**
- Overkill for small CLIs
- More boilerplate than Commander
- Steeper learning curve

**When to Use:**
- Multiple commands with subcommands
- Plugin ecosystem needed
- Team wants enterprise support
- Generating documentation automatically

**2025 Status:** Still highly recommended for complex CLIs

### 5.3 Commander.js

**Strengths:**
- Lightweight and flexible
- Easy setup with decent defaults
- Large community
- Minimal dependencies

**When to Use:**
- Simple to medium CLIs
- Want full control over implementation
- Prefer lightweight solutions

### 5.4 Yargs

**Strengths:**
- Declarative syntax (elegant)
- Built-in argument parsing utilities
- Good for complex argument parsing

**When to Use:**
- Prefer declarative over imperative
- Complex argument validation needed

### 5.5 Ink (React for CLIs)

**Unique Approach:** Uses React components for terminal UIs

```jsx
import React from 'react';
import {render, Text} from 'ink';

const MyApp = () => <Text color="green">Hello World</Text>;

render(<MyApp />);
```

**When to Use:**
- Interactive terminal UIs
- Team familiar with React
- Rich terminal experiences (dashboards, progress bars)

**Not For:** Simple command-line tools (overkill)

### 5.6 Recommendation for notion-cli

**Current:** Oclif ✅ (Appropriate choice)

**Rationale:**
- Multiple commands (block, page, db, user, search)
- Subcommands (block:append, block:update, page:create, etc.)
- Auto-generated help and documentation
- Mature ecosystem
- TypeScript-native

**No change recommended** - Oclif is the right fit for notion-cli's complexity level.

---

## 6. Package.json Best Practices (2025)

### 6.1 Essential Fields for CLI Tools

```json
{
  "name": "@scope/package-name",
  "version": "1.0.0",
  "description": "Clear, concise description",
  "type": "module",  // ESM (or omit for CJS)

  "bin": {
    "cli-command": "./bin/run"
  },

  "exports": {
    ".": {
      "types": "./dist/index.d.ts",
      "import": "./dist/index.js",
      "require": "./dist/index.cjs"
    }
  },

  "files": [
    "/bin",
    "/dist",
    "/oclif.manifest.json"
  ],

  "engines": {
    "node": ">=18.0.0"
  },

  "keywords": [
    "cli",
    "notion",
    "automation",
    "ai-agent"
  ],

  "repository": {
    "type": "git",
    "url": "https://github.com/user/repo.git"
  },

  "bugs": "https://github.com/user/repo/issues",
  "homepage": "https://github.com/user/repo#readme",
  "license": "MIT"
}
```

### 6.2 Bin Field Best Practices

**Shebang Required:**
```javascript
#!/usr/bin/env node

// Your CLI entry point
```

**Cross-Platform Considerations:**
- npm automatically creates `.cmd` wrapper on Windows
- Use forward slashes in paths: `./bin/run` (works everywhere)
- No need for platform detection in package.json

### 6.3 Files Field (Minimize Package Size)

**Include:**
```json
{
  "files": [
    "/bin",
    "/dist",
    "/oclif.manifest.json",
    "README.md",
    "LICENSE"
  ]
}
```

**Auto-Included (don't specify):**
- package.json
- README (any extension)
- LICENSE / LICENCE

**Auto-Excluded:**
- node_modules/
- .git/
- Files matching .npmignore or .gitignore

### 6.4 Scripts for CLI Development

```json
{
  "scripts": {
    "build": "shx rm -rf dist && tsc -b",
    "prepack": "npm run build",
    "postpack": "shx rm -f oclif.manifest.json",
    "prepublishOnly": "npm run build",
    "test": "vitest",
    "lint": "eslint . --ext .ts",
    "format": "prettier --write \"src/**/*.ts\""
  }
}
```

**Cross-Platform Script Tools:**
- **shx**: Unix commands (rm, cp, mkdir) on all platforms
- **cross-env**: Environment variables
- **rimraf**: Delete files/directories

### 6.5 Semantic Versioning (2025)

**Caret (^) vs Tilde (~):**

```json
{
  "dependencies": {
    "stable-pkg": "^1.2.3",     // Accept 1.x.x (minor + patch updates)
    "critical-pkg": "~1.2.3",   // Accept 1.2.x (patch updates only)
    "exact-pkg": "1.2.3"        // Exact version (no updates)
  }
}
```

**Best Practices:**
- Use `^` for most dependencies (npm default)
- Use `~` for critical dependencies requiring stability
- Use exact versions for known breaking packages
- Run `npm audit` regularly
- Consider socket.dev for supply chain security

**Pre-1.0.0 Behavior:**
```json
{
  "dependencies": {
    "beta-pkg": "^0.2.3"  // Acts like ~0.2.3 (patch only)
  }
}
```

---

## 7. Security Best Practices (2025)

### 7.1 npm audit

**Regular Scanning:**
```bash
# Basic audit
npm audit

# Fix automatically (careful!)
npm audit fix

# Production only
npm audit --production

# Specific severity
npm audit --audit-level=high

# JSON output for CI/CD
npm audit --json
```

**CI/CD Integration:**
```yaml
- name: Security Audit
  run: |
    npm audit --audit-level=moderate
    if [ $? -ne 0 ]; then
      echo "Security vulnerabilities found"
      exit 1
    fi
```

### 7.2 Socket.dev (Enhanced Security)

**Why Socket.dev:**
- Catches MORE issues than npm audit
- Detects malicious code (not just known vulnerabilities)
- Supply chain risk analysis
- Quality, maintenance, license concerns

**Installation:**
```bash
npm install -g @socketregistry/cli

# Use instead of npm install
socket npm install

# Scan existing project
socket scan
```

**GitHub App:** Install Socket GitHub App for PR scanning

**Firewall:** Free tool to block malicious packages at install time

### 7.3 Additional Security Tools

**Snyk:**
```bash
npm install -g snyk
snyk auth
snyk test
```

**npm-audit-resolver:**
```bash
npm install -g npm-audit-resolver
npm-audit-resolver
```

### 7.4 Dependency Management Best Practices

1. **Keep Dependencies Updated:**
   ```bash
   npm outdated
   npm update
   ```

2. **Use Lock Files:** Commit `package-lock.json` (or `npm-shrinkwrap.json` for published packages)

3. **Review Dependency Licenses:** Ensure compatibility with your project

4. **Minimize Dependencies:** Each dependency is a security surface

5. **Use Scoped Packages:** Prevents typosquatting (`@scope/package`)

---

## 8. npm Workspaces and Monorepos (2025)

### 8.1 npm Workspaces Overview

**Introduced:** npm v7 (2020)
**Status (2025):** Mature, widely adopted

**Basic Configuration:**
```json
{
  "name": "my-project",
  "workspaces": [
    "packages/*"
  ]
}
```

**Structure:**
```
my-project/
├── package.json
├── packages/
│   ├── cli/
│   │   └── package.json
│   ├── shared/
│   │   └── package.json
│   └── api/
│       └── package.json
```

### 8.2 Workspace Commands

```bash
# Install all workspace dependencies
npm install

# Run script in specific workspace
npm run build --workspace=packages/cli

# Run script in all workspaces
npm run test --workspaces

# Add dependency to workspace
npm install lodash --workspace=packages/cli
```

### 8.3 Monorepo Tools Comparison (2025)

| Tool | Performance | Caching | Task Dependencies | Learning Curve |
|------|-------------|---------|-------------------|----------------|
| **npm workspaces** | Good | No | No | Low |
| **pnpm workspaces** | Excellent | Yes | Limited | Low |
| **Nx** | Excellent | Yes | Yes | High |
| **Turborepo** | Excellent | Yes | Yes | Medium |
| **Lerna** | Good | Limited | Limited | Medium |

**Recommendation:**
- **Simple monorepo:** npm workspaces
- **Medium complexity:** pnpm workspaces
- **Complex monorepo:** Nx or Turborepo

### 8.4 Best Practices

1. **Consistent Versions:** Keep dependencies aligned across workspaces
2. **Shared Configs:** Centralize ESLint, TypeScript, Prettier configs
3. **Root-Level Dev Dependencies:** Install shared dev tools at root
4. **Named Workspaces:** Use descriptive package names

### 8.5 When NOT to Use Workspaces

- Single package project (notion-cli current state)
- Different release cycles for subpackages
- Team unfamiliar with monorepo concepts

---

## 9. Publishing Checklist (2025)

### 9.1 Pre-Publish Checklist

**1. Code Quality:**
- [ ] All tests passing
- [ ] Linting clean
- [ ] TypeScript compilation successful
- [ ] No console.log statements

**2. Package Configuration:**
- [ ] Correct package name (scoped if possible)
- [ ] Appropriate version number (semver)
- [ ] Description filled in
- [ ] Keywords added
- [ ] Repository URL set
- [ ] License specified
- [ ] Engines field set (Node.js version)

**3. Files:**
- [ ] `files` field includes only necessary files
- [ ] README.md comprehensive
- [ ] LICENSE file present
- [ ] CHANGELOG.md updated

**4. Security:**
- [ ] `npm audit` passing
- [ ] No secrets in code
- [ ] Dependencies reviewed
- [ ] Socket.dev scan clean (optional)

**5. Exports:**
- [ ] Main entry point correct
- [ ] Types exported properly
- [ ] Bin commands work locally (`npm link`)

### 9.2 Publishing Process

**Dry Run:**
```bash
# See what will be published
npm pack
tar -tzf package-name-1.0.0.tgz

# Test installation locally
npm install ./package-name-1.0.0.tgz -g
```

**Version Bump:**
```bash
npm version patch  # 1.0.0 -> 1.0.1
npm version minor  # 1.0.0 -> 1.1.0
npm version major  # 1.0.0 -> 2.0.0
```

**Publish:**
```bash
# First time (public)
npm publish --access public

# Subsequent publishes
npm publish

# With provenance (if not using trusted publishing)
npm publish --provenance
```

### 9.3 Post-Publish Checklist

- [ ] Test installation: `npm install -g your-package`
- [ ] Test CLI commands work
- [ ] Check npm page: https://www.npmjs.com/package/your-package
- [ ] Test on different platforms (Windows, Mac, Linux)
- [ ] Update GitHub release with notes
- [ ] Announce to users (if applicable)

### 9.4 Automated Publishing (Recommended)

**GitHub Actions Example:**
```yaml
name: Publish Package

on:
  release:
    types: [published]

jobs:
  publish:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      id-token: write

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          registry-url: 'https://registry.npmjs.org'

      - run: npm ci
      - run: npm run build
      - run: npm test

      - run: npm publish --access public
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

---

## 10. Recommendations for notion-cli

### 10.1 Current State Analysis

**Strengths:**
- ✅ Scoped package name (`@coastal-programs/notion-cli`)
- ✅ Oclif framework (appropriate for complexity)
- ✅ TypeScript
- ✅ Clear documentation
- ✅ Bin field configured
- ✅ Engines field set (Node.js >=18.0.0)
- ✅ Cross-platform scripts (shx)

**Areas for Improvement:**

1. **Distribution:** Currently using GitHub installation
   ```bash
   # Current
   npm install -g Coastal-Programs/notion-cli

   # Recommended
   npm install -g @coastal-programs/notion-cli
   ```

2. **TypeScript/ESM:** Currently using TypeScript 4.9.5, CommonJS
   ```json
   // Current
   {
     "main": "dist/index.js"
   }

   // Consider (if ecosystem supports ESM)
   {
     "type": "module",
     "exports": "./dist/index.js"
   }
   ```

3. **Security:** Add security scanning to CI/CD
   ```bash
   npm audit
   npx socket scan  # Recommended addition
   ```

### 10.2 Recommended Migration Path

**Phase 1: Publish to npm Registry (HIGH PRIORITY)**

**Why:**
- Better user experience
- Standard distribution method (2025)
- Improved performance
- Semantic versioning support
- Wider discoverability

**Steps:**
1. Ensure tests passing
2. Review package.json (looks good)
3. Set up trusted publishing or 2FA
4. Publish to npm registry: `npm publish --access public`
5. Update README installation instructions
6. Update CI/CD to publish on release

**Expected Timeline:** 1-2 hours

**Phase 2: Enhanced Security (MEDIUM PRIORITY)**

**Steps:**
1. Add `npm audit` to CI/CD
2. Consider socket.dev GitHub App
3. Set up Dependabot for dependency updates
4. Review and update dependencies

**Expected Timeline:** 2-3 hours

**Phase 3: TypeScript/ESM Migration (LOW PRIORITY)**

**When:** If ecosystem dependencies support ESM fully

**Steps:**
1. Update tsconfig.json to use "module": "node16"
2. Add "type": "module" to package.json
3. Update imports to use .js extensions
4. Test on all platforms
5. Consider dual publishing if needed

**Expected Timeline:** 1-2 days

**Trade-offs:**
- Breaking change for users
- May complicate build process
- Benefits: Better tree-shaking, future-proof

### 10.3 Immediate Action Items

1. **Publish to npm registry** (DO THIS FIRST)
   - Current GitHub installation is non-standard for 2025
   - Users expect `npm install -g @coastal-programs/notion-cli`

2. **Set up trusted publishing**
   - More secure than long-lived tokens
   - Automatic provenance attestations
   - Industry best practice

3. **Add security scanning**
   - `npm audit` in CI/CD
   - Consider socket.dev

4. **Update documentation**
   - Installation instructions
   - Package manager badges
   - npm registry link

### 10.4 Long-Term Considerations

**Monorepo (Future):**
- If notion-cli grows to include plugins, shared libraries
- npm workspaces sufficient for most cases
- Not needed currently

**ESM Migration:**
- Watch for ecosystem ESM adoption
- Migrate when major dependencies are ESM-first
- Consider dual publishing during transition

**Alternative Distribution:**
- Standalone binaries (pkg, nexe) for non-Node.js users
- Docker container for isolation
- Platform-specific installers (Homebrew, Chocolatey)

---

## 11. Key Takeaways

### 11.1 What Changed in 2024-2025

**Security:**
- Mandatory 2FA for publishing (October 2025)
- Trusted publishing with OIDC (July 2025)
- Enhanced provenance attestations
- Short-lived tokens (90-day max)

**TypeScript/ESM:**
- Node.js v22/v23 native CJS-requires-ESM support
- Better tooling (tsup, tsx, Vitest)
- arethetypeswrong for validation
- Still complex for dual publishing

**npm:**
- npm v11 with improved publishing workflow
- Better dist tag management
- Type prompts in npm init

**Security Tools:**
- socket.dev for supply chain security
- Enhanced npm audit
- Better integration with CI/CD

### 11.2 Standards vs Trends

**Standards (Follow These):**
- npm registry for CLI distribution ✅
- Trusted publishing for security ✅
- Semantic versioning ✅
- Cross-platform scripts (shx, cross-env) ✅
- Node.js >=18 for modern APIs ✅

**Trends (Evaluate for Your Project):**
- ESM-only packages (if ecosystem supports)
- Monorepos with workspaces (if multiple packages)
- Vitest over Jest (for new projects)
- pnpm over npm (for performance)
- Socket.dev for security (beyond npm audit)

### 11.3 When to Break Best Practices

**Use Git Installation When:**
- Private package not worth publishing
- Rapid development/testing phase
- Very niche use case
- Organization policy requires GitHub Packages

**Use CommonJS When:**
- Supporting Node.js <v18
- Ecosystem dependencies not ESM-ready
- Build simplicity more important than module format

**Skip Dual Publishing When:**
- Pure ESM is sufficient (Node.js >=v22)
- Single use case (not a library)
- Maintenance burden too high

---

## 12. Resources and References

### 12.1 Official Documentation

- [npm CLI Changelog](https://docs.npmjs.com/cli/v10/using-npm/changelog/)
- [npm Trusted Publishing](https://docs.npmjs.com/trusted-publishers/)
- [Node.js Packages Documentation](https://nodejs.org/api/packages.html)
- [TypeScript Module Resolution](https://www.typescriptlang.org/docs/handbook/module-resolution.html)

### 12.2 Tools and Libraries

**Build Tools:**
- [tsup](https://tsup.egoist.dev/) - TypeScript bundler
- [tshy](https://github.com/isaacs/tshy) - Opinionated dual publishing

**Testing:**
- [Vitest](https://vitest.dev/) - Modern test framework
- [tsx](https://github.com/esbuild-kit/tsx) - TypeScript execution

**Security:**
- [socket.dev](https://socket.dev/) - Supply chain security
- [Snyk](https://snyk.io/) - Vulnerability scanning

**CLI Frameworks:**
- [Oclif](https://oclif.io/) - Enterprise CLI framework
- [Commander.js](https://github.com/tj/commander.js/) - Lightweight CLI
- [Yargs](https://yargs.js.org/) - Declarative argument parsing
- [Ink](https://github.com/vadimdemedes/ink) - React for CLIs

**Type Validation:**
- [arethetypeswrong](https://github.com/arethetypeswrong/arethetypeswrong.github.io) - Package type analysis

### 12.3 Articles and Guides

**Essential Reading:**
- [TypeScript in 2025 with ESM and CJS](https://lirantal.com/blog/typescript-in-2025-with-esm-and-cjs-npm-publishing) by Liran Tal
- [Tutorial: Publishing ESM-based npm packages with TypeScript](https://2ality.com/2025/02/typescript-esm-packages.html) by Dr. Axel Rauschmayer
- [Mastering npm & npx in 2025](https://jewelhuq.medium.com/mastering-npm-npx-in-2025-the-definitive-guide-to-node-js-86b2c8e2a39d)
- [Complete Monorepo Guide](https://jsdev.space/complete-monorepo-guide/) (pnpm + Workspace + Changesets)

**npm Security:**
- [npm Blog: The right tool for the job](https://blog.npmjs.org/post/154387331670/the-right-tool-for-the-job-why-not-to-use-version.html)
- [GitHub npm registry security changes](https://www.theregister.com/2025/09/23/github_npm_registry_security/)

**Cross-Platform:**
- [Awesome Cross-Platform Node.js](https://github.com/bcoe/awesome-cross-platform-nodejs)
- [Creating ESM-based shell scripts](https://2ality.com/2022/07/nodejs-esm-shell-scripts.html)

### 12.4 Community Resources

**GitHub:**
- [npm/cli releases](https://github.com/npm/cli/releases)
- [node-semver](https://github.com/npm/node-semver) - Semver parser

**npm Registry:**
- [npm package search](https://www.npmjs.com/)
- [npm trends](https://npmtrends.com/) - Compare package popularity

---

## Conclusion

The npm ecosystem in 2025 is more secure, standardized, and TypeScript-friendly than ever before. Key trends include:

1. **Security-first:** Trusted publishing, mandatory 2FA, provenance attestations
2. **ESM transition:** Ongoing, with improved tooling but still complex
3. **npm registry dominance:** Standard for public CLI distribution
4. **Cross-platform maturity:** Automated wrapper generation works reliably
5. **Enterprise tooling:** Oclif, Nx, Turborepo for complex projects

**For notion-cli specifically:** Publishing to npm registry should be the immediate priority. The current GitHub-based installation is functional but non-standard for 2025. Users expect and prefer `npm install -g @coastal-programs/notion-cli` over `npm install -g Coastal-Programs/notion-cli`.

The package is otherwise well-structured with appropriate technology choices (Oclif, TypeScript, scoped naming). Security enhancements (trusted publishing, socket.dev) and potential ESM migration can be considered as secondary improvements.

---

**Document Version:** 1.0
**Research Date:** October 22, 2025
**Author:** Claude (Anthropic)
**Target:** notion-cli by Jake Schepis / Coastal Programs
