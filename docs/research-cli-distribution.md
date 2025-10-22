# Research: CLI Distribution Best Practices

**Research Date:** 2025-10-22
**Project:** notion-cli
**Purpose:** Understand best practices for distributing Node.js CLI tools

---

## Table of Contents

1. [GitHub vs npm Registry Distribution](#github-vs-npm-registry-distribution)
2. [Popular CLI Distribution Analysis](#popular-cli-distribution-analysis)
3. [The bin Field in package.json](#the-bin-field-in-packagejson)
4. [The files Field Best Practices](#the-files-field-best-practices)
5. [Build Artifacts and Git](#build-artifacts-and-git)
6. [Lifecycle Scripts](#lifecycle-scripts)
7. [Recommendations for notion-cli](#recommendations-for-notion-cli)

---

## GitHub vs npm Registry Distribution

### npm Registry (Primary Recommendation)

**Pros:**
- **Universal Expectation**: Most JavaScript/Node.js developers expect to install CLI tools via `npm install -g` or `npx`
- **Performance**: Applications using npm dependencies are significantly faster than git dependencies. The npm registry is optimized for serving packages
- **Pre-optimized Packages**: Unneeded files are already removed in the registry. With git, you pull the entire repo then npm strips files
- **No Authentication**: Public packages don't require authentication setup
- **Better Discoverability**: More visible to developers searching for tools
- **Semantic Versioning**: Easy version management with semver
- **Improved Uptime**: The npm Registry has evolved specifically for reliability and performance
- **Standard Workflow**: Simple `npm publish` workflow that developers understand

**Cons:**
- **Risk of Publishing Secrets**: Need to be careful with `.npmignore` or `files` field to avoid exposing sensitive data
- **Public Registry**: Private packages require paid npm account
- **Cannot Be Modified Post-Publish**: Once published, a version cannot be changed (only unpublished within 72 hours)

**Best Practices:**
- Use `npm-shrinkwrap.json` as lockfile to ensure pinned dependencies propagate to end users
- Minimize package size for better `npx` performance (npx always fetches from registry)
- Use `.npmignore` or `files` field to control what gets published
- Run `npm pack` to inspect tarball contents before publishing

### GitHub Installation

**Pros:**
- **No npm Account Required**: Anyone can install directly from public GitHub repos
- **Install Specific Branches/Commits**: Can target specific refs: `npm install user/repo#branch`
- **Development Versions**: Users can install unreleased features
- **Private Repos**: Works with private repos (with authentication)
- **No Registry Downtime**: Not dependent on npm registry availability

**Cons:**
- **Slower Performance**: Must clone entire repository, then npm strips unneeded files
- **Build Step Issues**: If package requires building, must use `prepare` script (adds install time)
- **Less User-Friendly**: Requires longer install command with full GitHub URL
- **No Version Discovery**: Users cannot easily browse available versions
- **Authentication Complexity**: Private repos require token/SSH setup
- **Not Expected Pattern**: Most developers don't expect to install CLIs from git

**When to Use:**
- Internal tools within an organization with private repos
- Beta testing unreleased features
- Fork/development workflows
- When avoiding npm registry entirely

### GitHub Packages Registry

**Pros:**
- **Unified Authentication**: Same credentials and permissions as GitHub repos
- **Security Integration**: Leverages repository-level access controls
- **Multi-format Support**: Can host multiple package types in one registry
- **No Additional Service**: No need for third-party npm proxy services
- **GitHub Ecosystem**: Integrates with GitHub Actions, issues, discussions

**Cons:**
- **Authentication Always Required**: Even for public packages, requires `npm login` or `.npmrc` token config
- **Scoped Packages Only**: All packages must be scoped with org/owner name: `@org/package`
- **No Proxying to npm**: Cannot mix with npmjs.com packages easily - requires configuration
- **Setup Friction**: Every downstream developer needs package-manager-specific config
- **Less Discoverability**: Not as easily found by developers searching for tools
- **Smaller Ecosystem**: Most public packages are on npmjs.com

**When to Use:**
- Large organizations with many private packages needing unified access control
- When tight integration with GitHub repositories is critical
- Internal tooling where all developers already have GitHub access
- When you need multi-format package hosting

**Recommendation for Public CLIs:**
Use npm registry (npmjs.com) as the primary distribution method. GitHub installation can be documented as an alternative for development/testing, but should not be the primary method.

---

## Popular CLI Distribution Analysis

### Research Methodology

Analyzed package.json configuration from these popular CLI tools:
- **@oclif/cli** - CLI framework used by Salesforce and others
- **vercel** - Vercel deployment CLI
- **netlify-cli** - Netlify deployment CLI
- **@angular/cli** - Angular development CLI
- **create-react-app** - React project scaffolding tool

### Comparison Table

| CLI | bin Field | files Field | main | types | Distribution |
|-----|-----------|-------------|------|-------|--------------|
| **@oclif/cli** | `"oclif": "bin/run.js"` | `["oclif.manifest.json", "./bin", "./lib", "./templates"]` | `"lib/index.js"` | `"lib/index.d.ts"` | npm registry |
| **vercel** | `{"vc": "./dist/vc.js", "vercel": "./dist/vc.js"}` | `["dist"]` | Not specified | Not specified | npm registry |
| **netlify-cli** | `{"ntl": "./bin/run.js", "netlify": "./bin/run.js"}` | `["/bin", "/npm-shrinkwrap.json", "/scripts", "/functions-templates", "/dist"]` | Not specified | Not specified | npm registry |
| **@angular/cli** | `"ng": "bin/ng.js"` (assumed) | Not visible | Not visible | Not visible | npm registry |
| **create-react-app** | `"create-react-app": "./index.js"` | `["index.js", "createReactApp.js"]` | Not specified | Not specified | npm registry |

### Key Patterns Observed

**1. Multiple Command Names**
- Many CLIs provide multiple command names (aliases)
- Example: Vercel uses both `vercel` and `vc`
- Example: Netlify uses both `netlify` and `ntl`

**2. Build Output Structure**
- **oclif pattern**: Compiles to `lib/` directory, includes `bin/` with runner scripts
- **vercel/netlify pattern**: Compiles to `dist/` directory
- **create-react-app pattern**: No build step, ships source JavaScript

**3. Files Field Strategy**
- Most include only essential runtime files
- Common: compiled code directory (`dist/` or `lib/`)
- Some include templates, manifests, or helper scripts
- Exclude source TypeScript files from published package

**4. Distribution Strategy**

All analyzed CLIs use npm registry as primary distribution with these methods:

- **npm**: Standard `npm publish` workflow
- **Standalone Tarballs** (oclif): Built with `oclif pack tarballs` - includes Node.js binary so users don't need Node installed
- **Homebrew**: Some CLIs distribute via Homebrew for macOS
- **Snap**: Some Linux CLIs use Snap for Ubuntu 16+
- **Auto-update**: oclif-based CLIs can use `@oclif/plugin-update` for background updates

### Build Scripts Comparison

| CLI | Build Command | Pre-publish Hook |
|-----|---------------|------------------|
| **@oclif/cli** | `shx rm -rf lib && tsc` | `prepack: yarn build && bin/run.js manifest` |
| **vercel** | `node scripts/build.mjs` | Not visible |
| **netlify-cli** | `tsc --project tsconfig.build.json` | `prepublishOnly: node ./scripts/prepublishOnly.js` |
| **create-react-app** | None (ships source) | None |

**Pattern:** Most TypeScript-based CLIs use `prepack` or `prepublishOnly` to ensure the package is built before publishing.

---

## The bin Field in package.json

### What is the bin Field?

The `bin` field in `package.json` maps command names to executable script files. When a package is installed, npm creates the appropriate executable entries so users can run the command from the terminal.

### Syntax

**Single Command (package name = command name):**
```json
{
  "name": "my-tool",
  "bin": "./dist/cli.js"
}
```
This creates a `my-tool` command.

**Single Command (custom name):**
```json
{
  "name": "my-package",
  "bin": {
    "my-command": "./dist/cli.js"
  }
}
```
This creates a `my-command` command.

**Multiple Commands:**
```json
{
  "name": "my-package",
  "bin": {
    "command-1": "./dist/cli1.js",
    "command-2": "./dist/cli2.js"
  }
}
```

### How npm Handles bin Across Platforms

**Unix/Linux/macOS:**
- npm creates a symbolic link in the appropriate bin directory
- Local install: `node_modules/.bin/command-name`
- Global install: `/usr/local/bin/command-name` (or similar based on npm config)
- The shell executes the file using the shebang line

**Windows:**
- npm creates multiple files for compatibility:
  - `command-name.cmd` - Windows Command Prompt batch file
  - `command-name.ps1` - PowerShell script
  - `command-name` - Unix-style for Cygwin/WSL environments
- These wrapper files call Node.js to execute the target script

This cross-platform handling is automatic - you don't need to write platform-specific code.

### Shebang Requirement

**Required first line in your bin script:**
```javascript
#!/usr/bin/env node
```

**Why this specific shebang?**
- `/usr/bin/env` is reliably located at the same path on Unix systems
- `env` finds `node` in the user's PATH environment variable
- Allows Node.js to be installed anywhere (nvm, different versions, etc.)
- More portable than hardcoding Node.js path like `#!/usr/local/bin/node`

**Why it works on Windows:**
- Windows ignores shebang lines
- npm reads the shebang and creates `.cmd` wrappers that invoke node with your script
- This is why the shebang is required even for Windows-only tools

### Example bin Script

**File: `bin/run.js`**
```javascript
#!/usr/bin/env node

// Optional: Set up environment or handle errors
const oclif = require('@oclif/core');

oclif.run().then(require('@oclif/core/flush')).catch(require('@oclif/core/handle'));
```

Or simpler:
```javascript
#!/usr/bin/env node

require('../dist/index.js');
```

### Local Testing

**Test your bin script before publishing:**

```bash
# Create local symlink
npm link

# Your command is now available globally
my-command --help

# Remove symlink
npm unlink
```

### Installation Behavior

**Local Installation:**
```bash
npm install my-package
```
- Creates symlink at `node_modules/.bin/my-command`
- Run with `npx my-command` or `npm exec my-command`
- Or add to package.json scripts

**Global Installation:**
```bash
npm install -g my-package
```
- Creates executable in global bin directory (in PATH)
- Run directly: `my-command`

### Cross-Platform Best Practices

1. **Always include shebang** - Even if targeting only one platform
2. **Use `/usr/bin/env node`** - Most portable option
3. **Avoid platform-specific code** - Write JavaScript instead of shell scripts
4. **Test on multiple platforms** - Windows, macOS, Linux
5. **File paths**: Use `path.join()` and Node.js path utilities for cross-platform file handling
6. **Line endings**: Configure git to handle CRLF/LF properly (`.gitattributes`)

### Common Issues and Solutions

**Issue:** Command not found after installation
- **Solution:** Check that shebang line is present and is the first line
- **Solution:** Verify bin path points to correct file
- **Solution:** Check file permissions on Unix (should be executable)

**Issue:** Windows-specific errors
- **Solution:** Ensure you're not using shell-specific syntax (bash/sh)
- **Solution:** Use Node.js APIs instead of shell commands

**Issue:** Different behavior on different platforms
- **Solution:** Detect platform with `process.platform` if needed
- **Solution:** Use cross-platform libraries (e.g., `cross-spawn`, `shx`)

---

## The files Field Best Practices

### What is the files Field?

The `files` field is a **whitelist** array in `package.json` that specifies which files and directories should be included when your package is published to npm. If omitted, npm uses `.npmignore` rules (or `.gitignore` if no `.npmignore` exists).

### Behavior

**With `files` field:**
- Everything is excluded by default
- Only files/directories listed in the array are included
- Directories are walked recursively

**Without `files` field:**
- npm uses `.npmignore` rules (blacklist approach)
- If no `.npmignore`, uses `.gitignore` rules
- Less explicit control

**Always Included (regardless of `files` field):**
- `package.json`
- `README` (any variant)
- `LICENSE` / `LICENCE` (any variant)
- `CHANGELOG` (any variant)
- Files referenced by `main`, `bin`, `browser`, `types`

**Always Excluded:**
- `.git/`
- `node_modules/`
- `.npmrc`
- `package-lock.json` (use `npm-shrinkwrap.json` instead)
- npm-debug logs

### Common Patterns

**Compiled TypeScript Project:**
```json
{
  "files": [
    "/dist",
    "/bin"
  ]
}
```

**With Additional Resources:**
```json
{
  "files": [
    "/dist",
    "/bin",
    "/templates",
    "/npm-shrinkwrap.json"
  ]
}
```

**Minimal (source JavaScript):**
```json
{
  "files": [
    "index.js",
    "lib/"
  ]
}
```

### What Should Be Included?

**Include:**
- Compiled/transpiled code (dist, lib, build directories)
- Bin scripts
- TypeScript definition files (.d.ts)
- Required assets (templates, config files, etc.)
- npm-shrinkwrap.json (if used)
- Critical documentation referenced by code

**Exclude:**
- Source TypeScript files (.ts)
- Test files
- Development configuration (.eslintrc, .prettierrc, etc.)
- Build scripts
- Documentation (docs/ folder) - users can view on GitHub
- Examples and demos
- CI/CD configuration (.github/, .travis.yml, etc.)
- Environment files (.env)
- Editor config (.vscode/, .idea/)

### Key Insight: files vs .gitignore

The `files` field **ignores** `.gitignore` rules. This is a powerful feature:

```json
{
  "files": ["dist"]
}
```

Even if `dist/` is in `.gitignore` (keeping Git clean), npm will still include it when publishing because it's explicitly listed in `files`.

### Testing What Gets Published

**Before publishing, test what will be included:**

```bash
# Create tarball without publishing
npm pack

# This creates my-package-1.0.0.tgz
# Extract and inspect:
tar -xzf my-package-1.0.0.tgz
cd package
ls -la
```

**Or use npm's built-in tool:**
```bash
npm publish --dry-run
```

This shows what would be published without actually publishing.

### Best Practice: Use files Field

**Why prefer `files` over `.npmignore`?**

1. **Explicit Whitelist**: Clear about what's included
2. **Easier to Audit**: See at a glance what gets published
3. **Safer**: Less risk of accidentally including sensitive files
4. **Works with .gitignore**: Can keep build artifacts out of git but include in npm

**Recommended Approach:**
```json
{
  "files": [
    "/dist",
    "/bin",
    "npm-shrinkwrap.json"
  ]
}
```

And in `.gitignore`:
```
dist/
*.tgz
```

This keeps Git clean while ensuring npm packages include the necessary built artifacts.

### Common Mistakes

1. **Not Testing**: Forgetting to run `npm pack` before publishing
2. **Too Broad**: Including `src/` when `dist/` is sufficient
3. **Missing Assets**: Forgetting to include templates or other runtime assets
4. **Sensitive Data**: Accidentally including `.env` or credential files
5. **Large Files**: Including unnecessary large files that bloat the package

---

## Build Artifacts and Git

### The Core Question

**Should compiled JavaScript (dist/, lib/, build/) be committed to Git?**

**Short Answer:** No for the repository, Yes for npm.

### Recommended Strategy

**Git Repository:**
- Keep `dist/` in `.gitignore`
- Commit only source code (TypeScript, etc.)
- Keep the repository clean and focused on source
- Avoid merge conflicts in generated files
- Clearer diffs and code review

**npm Package:**
- Include `dist/` in published package
- Use `files` field to specify: `"files": ["dist"]`
- Run build before publishing

### Why This Approach?

**Benefits of excluding dist/ from Git:**
1. **Cleaner Repository**: Only source files tracked
2. **No Generated Code Conflicts**: Multiple developers don't create merge conflicts on built files
3. **Clear History**: Git log shows only intentional source changes
4. **Smaller Clone Size**: No redundant compiled code
5. **Clear Code Review**: Reviewers see only source changes

**How Users Get Compiled Code:**

**When installing from npm registry:**
- Registry contains the built package
- Users get pre-compiled code
- No build step needed on install

**When installing from GitHub (development):**
- Use the `prepare` script (covered in next section)
- npm automatically builds during installation

### Implementation

**In `.gitignore`:**
```
# Build output
dist/
lib/
build/
*.tgz
```

**In `package.json`:**
```json
{
  "files": [
    "/dist"
  ],
  "scripts": {
    "build": "tsc",
    "prepare": "npm run build"
  }
}
```

**Or use `.npmignore` approach:**

Create an empty `.npmignore` file (or one with different rules) to prevent npm from using `.gitignore` rules.

### Alternative: Committing dist/

Some projects do commit dist/ to Git. This is generally discouraged for libraries/CLIs but may be acceptable for:

1. **Simple Build-Free Projects**: If there's no build step
2. **Specific Requirements**: Team has specific reasons
3. **End-User Applications**: Deployable apps might commit built code

**If you do commit dist/:**
- Be prepared for merge conflicts
- Use pre-commit hooks to ensure it's always up to date
- Document clearly why this choice was made

### Real-World Examples

**Projects that DON'T commit dist/ (majority):**
- @oclif/cli (builds before publish with `prepack`)
- vercel CLI (builds before publish)
- netlify-cli (builds before publish)
- Most modern TypeScript projects

**Projects that DO commit dist/:**
- Some legacy projects
- Some browser libraries that users expect to include via `<script>` tags directly from GitHub
- Projects using bower (largely deprecated)

### Key Takeaway

**Standard Modern Practice:**
1. Add `dist/` to `.gitignore`
2. List `dist/` in `files` field in `package.json`
3. Use `prepare` or `prepack` script to build before publishing
4. Users installing from npm get compiled code
5. Users installing from GitHub get code built automatically via `prepare`

This keeps Git clean while ensuring all installation methods work correctly.

---

## Lifecycle Scripts

### Overview

npm provides lifecycle scripts that run automatically at specific times. Understanding these is critical for proper package distribution.

### Relevant Lifecycle Scripts for CLI Distribution

#### prepare

**When it runs:**
- Before the package is packed (publishing)
- Before the package is published
- On local `npm install` (without arguments)
- When installing from Git repositories

**Use case:**
- **Build step for Git installations**
- Ensures anyone installing from Git gets compiled code

**Example:**
```json
{
  "scripts": {
    "prepare": "npm run build"
  }
}
```

**Why it's important:**
When someone runs `npm install github:user/repo`, they get the source code. The `prepare` script builds it automatically.

#### prepublishOnly

**When it runs:**
- Before `npm publish` ONLY
- Does NOT run on `npm install`
- Does NOT run when installing from Git

**Use case:**
- **Validation before publishing**
- Run tests, linting, security checks
- Ensure package is in publishable state

**Example:**
```json
{
  "scripts": {
    "prepublishOnly": "npm run test && npm run lint"
  }
}
```

**Why it's important:**
Catches issues before publishing. Once published, a version cannot be changed (only unpublished within 72 hours).

#### prepack

**When it runs:**
- Before a tarball is packed
- Before `npm publish` (after `prepublishOnly`)
- Before `npm pack`
- When installing Git dependencies

**Use case:**
- **Build and prepare package for distribution**
- Generate manifests or other derived files
- Final preparation before packaging

**Example:**
```json
{
  "scripts": {
    "prepack": "npm run build && oclif manifest"
  }
}
```

**Why it's important:**
Ensures package is always built and manifests are up-to-date for all distribution methods.

#### postpack

**When it runs:**
- After a tarball is packed
- After `npm publish` completes
- After `npm pack`

**Use case:**
- **Cleanup temporary files**
- Remove files that were generated for packaging but shouldn't remain

**Example:**
```json
{
  "scripts": {
    "postpack": "rm -f oclif.manifest.json"
  }
}
```

#### preversion

**When it runs:**
- Before `npm version` command bumps version

**Use case:**
- **Pre-release validation**
- Run tests and linting before version bump

**Example:**
```json
{
  "scripts": {
    "preversion": "npm run lint"
  }
}
}
```

### Execution Order

When you run `npm publish`:

```
1. prepublishOnly
2. prepare
3. prepack
4. [tarball is created]
5. postpack
6. [publishes to registry]
7. publish
8. postpublish
```

When someone runs `npm install` from local directory:

```
1. preinstall
2. install
3. postinstall
4. prepare  (if package has prepare script)
```

When someone runs `npm install github:user/repo`:

```
1. [clones repository]
2. prepare  (builds the code)
3. preinstall
4. install
5. postinstall
```

### Recommended Pattern for TypeScript CLIs

```json
{
  "scripts": {
    "build": "shx rm -rf dist && tsc",
    "prepare": "npm run build",
    "prepublishOnly": "npm test && npm run lint",
    "prepack": "oclif manifest",
    "postpack": "shx rm -f oclif.manifest.json",
    "preversion": "npm run lint",
    "version": "oclif readme && git add README.md",
    "test": "mocha --forbid-only \"test/**/*.test.ts\"",
    "lint": "eslint . --ext .ts"
  }
}
```

**What this accomplishes:**

1. **`prepare`**: Builds the package for Git installations and before publishing
2. **`prepublishOnly`**: Validates code before publishing (tests + lint)
3. **`prepack`**: Generates oclif manifest before packaging
4. **`postpack`**: Cleans up generated manifest
5. **`preversion`**: Validates before version bump
6. **`version`**: Updates README and stages it for commit

### Common Pitfalls

**Don't use `prepublish`:**
- Deprecated since npm v4
- Split into `prepare` and `prepublishOnly`
- Confusing behavior led to its deprecation

**Don't build in `prepublishOnly`:**
- Git installations won't work (script doesn't run)
- Use `prepare` for building

**Don't have side effects in `prepare`:**
- Runs on every `npm install`
- Should be idempotent (safe to run multiple times)
- Keep it fast (impacts install time)

### Testing Your Scripts

**Test publishing workflow:**
```bash
npm pack --dry-run
```

**Test Git installation:**
```bash
# In another directory
npm install file:../path/to/your/package
```

**Test global installation:**
```bash
npm link
# Test your CLI
npm unlink
```

---

## Recommendations for notion-cli

Based on the research above and analysis of the current notion-cli configuration, here are specific recommendations:

### Current State Analysis

**From `package.json`:**
```json
{
  "name": "@coastal-programs/notion-cli",
  "version": "5.1.0",
  "bin": {
    "notion-cli": "./bin/run"
  },
  "main": "dist/index.js",
  "files": [
    "/bin",
    "/dist",
    "/npm-shrinkwrap.json",
    "/oclif.manifest.json"
  ],
  "scripts": {
    "build": "shx rm -rf dist && tsc -b",
    "postpack": "shx rm -f oclif.manifest.json"
  }
}
```

**From `.gitignore`:**
```
# Build output
lib/
*.tgz
*.js.map
```

**Note:** `dist/` is NOT in `.gitignore`

### Issues Identified

1. **Missing `prepare` script** - Git installations won't work (code won't be built)
2. **Missing `prepack` script** - oclif manifest might not be generated
3. **Missing `prepublishOnly` script** - No validation before publishing
4. **dist/ not in .gitignore** - Build artifacts might be committed to Git
5. **Missing validation scripts** - No pre-publish testing

### Recommended Changes

#### 1. Update `.gitignore`

**Add to `.gitignore`:**
```gitignore
# Build output
lib/
dist/
*.tgz
*.js.map

# oclif manifest (generated during pack)
oclif.manifest.json
```

**Why:**
- Keeps Git clean of build artifacts
- Prevents merge conflicts
- Follows best practices from research

#### 2. Update `package.json` scripts

**Add these scripts:**
```json
{
  "scripts": {
    "build": "shx rm -rf dist && tsc -b",
    "prepare": "npm run build",
    "prepublishOnly": "npm test && npm run lint",
    "prepack": "oclif manifest",
    "postpack": "shx rm -f oclif.manifest.json",
    "preversion": "npm run lint",
    "version": "npm run readme",
    "test": "mocha --forbid-only \"test/**/*.test.ts\"",
    "lint": "eslint . --ext .ts --config .eslintrc.json",
    "readme": "oclif readme --multi --no-aliases && shx sed -i \"s/^_See code:.*$//g\" docs/*.md > /dev/null"
  }
}
```

**What each does:**

- **`prepare`**: Builds TypeScript when installing from Git
- **`prepublishOnly`**: Validates before publishing (tests + lint)
- **`prepack`**: Generates oclif manifest before packaging
- **`postpack`**: Cleans up manifest file
- **`preversion`**: Validates before version bump
- **`version`**: Updates README and docs for version bump

#### 3. Verify `bin/run` has shebang

**Ensure `bin/run` starts with:**
```javascript
#!/usr/bin/env node
```

This is required for cross-platform compatibility.

#### 4. Create `.npmignore` (Optional)

If you need different rules for npm vs git:

**Create `.npmignore`:**
```
# Don't publish these
src/
test/
.github/
.vscode/
.eslintrc.json
.prettierrc
tsconfig.json
*.test.ts
```

**Or continue using `files` field (recommended):**
The current `files` field is good - it's explicit about what to include.

#### 5. Distribution Strategy

**Primary: npm Registry (npmjs.com)**

Publish to npm registry as `@coastal-programs/notion-cli`:

```bash
# Ensure you're logged in
npm login

# Publish (prepublishOnly runs automatically)
npm publish --access public
```

**Secondary: GitHub Installation**

Document in README that users can install from GitHub:

```bash
npm install -g github:Coastal-Programs/notion-cli
```

This will work with the `prepare` script to build automatically.

**Alternative: GitHub Packages**

If you want to also publish to GitHub Packages (probably not necessary for public CLI):

```bash
npm publish --registry=https://npm.pkg.github.com
```

Requires users to configure `.npmrc`. Generally not recommended unless you have specific needs.

### Testing Workflow

Before publishing:

```bash
# 1. Clean install to test
rm -rf node_modules dist
npm install

# 2. Build
npm run build

# 3. Test scripts
npm test
npm run lint

# 4. Test package contents
npm pack
tar -xzf coastal-programs-notion-cli-5.1.0.tgz
cd package
ls -la
# Verify dist/, bin/ are present, src/ is not
cd ..
rm -rf package *.tgz

# 5. Test local installation
npm link
notion-cli --help
npm unlink

# 6. Test Git installation (in another directory)
npm install -g github:Coastal-Programs/notion-cli
notion-cli --help
npm uninstall -g @coastal-programs/notion-cli

# 7. Ready to publish!
npm publish --access public
```

### Publishing Workflow

**Standard publish:**
```bash
# 1. Update version
npm version patch  # or minor, or major

# 2. Push changes and tags
git push && git push --tags

# 3. Publish to npm
npm publish --access public
```

This will:
1. Run `preversion` (lint)
2. Bump version
3. Run `version` (update README)
4. Commit and tag
5. Push to GitHub
6. Run `prepublishOnly` (test + lint)
7. Run `prepare` (build)
8. Run `prepack` (generate manifest)
9. Create tarball
10. Run `postpack` (cleanup manifest)
11. Publish to npm

### Documentation Updates

**Add to README.md:**

```markdown
## Installation

### npm (Recommended)

```bash
npm install -g @coastal-programs/notion-cli
```

### GitHub (Development)

```bash
npm install -g github:Coastal-Programs/notion-cli
```

### npx (No Installation)

```bash
npx @coastal-programs/notion-cli [command]
```

## For Developers

### Local Development

```bash
# Clone repository
git clone https://github.com/Coastal-Programs/notion-cli.git
cd notion-cli

# Install dependencies
npm install

# Build TypeScript
npm run build

# Link globally for testing
npm link

# Test
notion-cli --help

# Unlink when done
npm unlink
```

### Publishing

```bash
# 1. Ensure tests pass
npm test

# 2. Update version
npm version patch

# 3. Push
git push && git push --tags

# 4. Publish
npm publish --access public
```
```

### Summary of Changes Needed

**Immediate:**
1. Add `dist/` to `.gitignore`
2. Add `oclif.manifest.json` to `.gitignore`
3. Add `prepare`, `prepublishOnly`, and `prepack` scripts to `package.json`
4. Verify `bin/run` has shebang
5. Run `npm pack` to verify contents
6. Update README with installation instructions

**Optional:**
1. Add preversion and version scripts for automated README updates
2. Create comprehensive publishing documentation
3. Set up GitHub Actions for automated testing before publish

### Benefits of These Changes

1. **Git installations work** - `prepare` script builds code automatically
2. **Pre-publish validation** - Catches issues before publishing
3. **Clean Git history** - No build artifacts committed
4. **Professional distribution** - Follows patterns from popular CLIs
5. **Cross-platform compatibility** - Works on Windows, macOS, Linux
6. **Multiple install methods** - npm, npx, GitHub all work correctly

---

## Additional Resources

### Documentation
- [npm package.json documentation](https://docs.npmjs.com/cli/v7/configuring-npm/package-json/)
- [npm scripts lifecycle](https://docs.npmjs.com/cli/v6/using-npm/scripts/)
- [oclif documentation](https://oclif.io/)
- [Node.js CLI best practices](https://github.com/lirantal/nodejs-cli-apps-best-practices)

### Tools
- `npm pack` - Test what will be published
- `npm link` - Test CLI locally
- `npm publish --dry-run` - Simulate publishing
- `npm-check-updates` - Keep dependencies updated
- `npm audit` - Security check

### Key Principles

1. **Use npm registry as primary distribution** - It's what users expect
2. **Keep Git clean** - Don't commit build artifacts
3. **Use `files` field** - Explicitly whitelist what to publish
4. **Use lifecycle scripts** - Automate build and validation
5. **Test before publishing** - Use `npm pack` and local testing
6. **Cross-platform compatibility** - Use shebang, Node APIs, test on multiple platforms
7. **Semantic versioning** - Follow semver for version numbers
8. **Document installation** - Provide clear installation instructions in README

---

**Research compiled by:** Claude (Anthropic)
**Date:** 2025-10-22
**Based on:** Web research, documentation analysis, and real-world examples from popular CLI tools
