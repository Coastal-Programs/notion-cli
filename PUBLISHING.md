# Publishing notion-cli

notion-cli is a Go binary distributed through three channels:

1. **npm** (primary) -- platform-specific binary packages via `npm install -g @coastal-programs/notion-cli`
2. **GitHub Releases** -- direct binary downloads for all platforms
3. **Go module** -- `go install github.com/Coastal-Programs/notion-cli/cmd/notion-cli@latest`

---

## How the npm Distribution Works

The npm distribution follows the esbuild pattern: a thin wrapper package plus platform-specific optional dependency packages that contain the actual Go binary.

### Package Layout

```
@coastal-programs/notion-cli              (main wrapper)
  bin/notion-cli.js                       JS shim that finds and exec's the Go binary
  install.js                              postinstall fallback: downloads binary from GitHub Releases
  package.json                            declares optionalDependencies for all platforms

@coastal-programs/notion-cli-darwin-arm64  (macOS Apple Silicon)
@coastal-programs/notion-cli-darwin-x64    (macOS Intel)
@coastal-programs/notion-cli-linux-x64     (Linux x86_64)
@coastal-programs/notion-cli-linux-arm64   (Linux ARM64)
@coastal-programs/notion-cli-win32-x64     (Windows x86_64)
```

Each platform package has `os` and `cpu` fields in its `package.json` so npm only installs the one matching the user's system.

### Binary Resolution Order

When a user runs `notion-cli`, the JS shim in `bin/notion-cli.js` looks for the binary in this order:

1. **Platform optional dependency** -- the `@coastal-programs/notion-cli-<platform>` package installed by npm
2. **Postinstall cache** -- `node_modules/.cache/notion-cli/bin/notion-cli`, downloaded by `install.js` as a fallback when the optional dependency fails
3. **Home directory** -- `~/.notion-cli/bin/notion-cli`, for manual installations

If the binary is not found in any of these locations, the shim prints an error and exits.

---

## Automated Publishing via GitHub Actions

The project has automated npm publishing via the workflow at `.github/workflows/publish.yml`.

### One-Time Setup (3 minutes)

#### 1. Create npm Access Token

1. Go to https://www.npmjs.com/settings/YOUR_USERNAME/tokens
2. Click "Generate New Token" then "Granular Access Token"
3. Fill out the form:
   - **Token name**: `GitHub Actions - notion-cli CI/CD`
   - **Description**: `Automated publishing token for CI/CD`
   - **Bypass two-factor authentication (2FA)**: **CHECK THIS BOX** -- critical for automation
   - **Allowed IP ranges**: Leave empty (GitHub Actions uses dynamic IPs)
   - **Expiration**: 90 days (maximum for write tokens)
4. **Packages and scopes**:
   - Change permissions dropdown to: **"Read and write"**
   - Select packages: `@coastal-programs/notion-cli` and all five platform packages
5. **Organizations**: Leave as "No access"
6. Click "Generate token"
7. **Copy the token immediately** (starts with `npm_...`) -- you will not see it again

**Important**: The "Bypass 2FA" checkbox is essential. Without it, automated publishing will fail with OTP errors even if you have a valid token.

#### 2. Add Token to GitHub Secrets

1. Go to https://github.com/Coastal-Programs/notion-cli/settings/secrets/actions
2. Click "New repository secret"
3. Name: `NPM_TOKEN`
4. Value: Paste your npm token
5. Click "Add secret"

That is the entire one-time setup.

---

## Publishing a New Version

### Step-by-Step Release Process

```bash
# 1. Make sure everything passes
make test
make build
make lint

# 2. Update the version in package.json (and all platform package.json files)
#    Choose one:
npm version patch  # 6.0.0 -> 6.0.1 (bug fixes)
npm version minor  # 6.0.0 -> 6.1.0 (new features)
npm version major  # 6.0.0 -> 7.0.0 (breaking changes)

# 3. Update the version in each platform package.json to match:
#    npm/notion-cli-darwin-arm64/package.json
#    npm/notion-cli-darwin-x64/package.json
#    npm/notion-cli-linux-x64/package.json
#    npm/notion-cli-linux-arm64/package.json
#    npm/notion-cli-win32-x64/package.json
#
#    Also update the optionalDependencies versions in the root package.json.

# 4. Update CHANGELOG.md with the new version and changes

# 5. Cross-compile Go binaries for all platforms
make release
# Produces:
#   build/notion-cli-darwin-arm64
#   build/notion-cli-darwin-amd64
#   build/notion-cli-linux-amd64
#   build/notion-cli-linux-arm64
#   build/notion-cli-windows-amd64.exe

# 6. Commit, tag, and push
git add -A
git commit -m "chore: release v6.0.1"
git tag -a v6.0.1 -m "Release v6.0.1"
git push --follow-tags

# 7. Create GitHub Release with binary attachments
gh release create v6.0.1 \
  --title "v6.0.1 - Release Title" \
  --notes "$(cat <<'EOF'
## What's New
- Description of changes

See CHANGELOG.md for full details.
EOF
)" \
  build/notion-cli-darwin-arm64 \
  build/notion-cli-darwin-amd64 \
  build/notion-cli-linux-amd64 \
  build/notion-cli-linux-arm64 \
  build/notion-cli-windows-amd64.exe

# 8. GitHub Actions automatically publishes to npm
```

### What Happens Automatically

When you create a GitHub Release, the publish workflow:

- Checks out the code
- Runs `make test` (Go test suite)
- Runs `make build` (Go binary build)
- Checks if the version already exists on npm
- Publishes the main wrapper package to npm with provenance
- Publishes each platform package with its respective binary

### Manual Workflow Trigger

You can also trigger the publish workflow manually:

1. Go to Actions tab: https://github.com/Coastal-Programs/notion-cli/actions/workflows/publish.yml
2. Click "Run workflow"
3. Click "Run workflow" button

---

## Publishing Platform Packages

Each platform package must be published separately with the correct binary inside. The process for each:

```bash
# Copy the correct binary into each platform package
cp build/notion-cli-darwin-arm64 npm/notion-cli-darwin-arm64/bin/notion-cli
cp build/notion-cli-darwin-amd64 npm/notion-cli-darwin-x64/bin/notion-cli
cp build/notion-cli-linux-amd64  npm/notion-cli-linux-x64/bin/notion-cli
cp build/notion-cli-linux-arm64  npm/notion-cli-linux-arm64/bin/notion-cli
cp build/notion-cli-windows-amd64.exe npm/notion-cli-win32-x64/bin/notion-cli.exe

# Publish each platform package
cd npm/notion-cli-darwin-arm64 && npm publish --access public && cd ../..
cd npm/notion-cli-darwin-x64   && npm publish --access public && cd ../..
cd npm/notion-cli-linux-x64    && npm publish --access public && cd ../..
cd npm/notion-cli-linux-arm64  && npm publish --access public && cd ../..
cd npm/notion-cli-win32-x64    && npm publish --access public && cd ../..

# Then publish the main wrapper package
npm publish --access public
```

**Note**: The GitHub Actions workflow should handle all of this automatically. Use this manual process only as a fallback.

---

## Direct Binary Distribution (GitHub Releases)

Users who do not use npm can download binaries directly from GitHub Releases:

```
https://github.com/Coastal-Programs/notion-cli/releases/download/v6.0.0/notion-cli-darwin-arm64
https://github.com/Coastal-Programs/notion-cli/releases/download/v6.0.0/notion-cli-darwin-amd64
https://github.com/Coastal-Programs/notion-cli/releases/download/v6.0.0/notion-cli-linux-amd64
https://github.com/Coastal-Programs/notion-cli/releases/download/v6.0.0/notion-cli-linux-arm64
https://github.com/Coastal-Programs/notion-cli/releases/download/v6.0.0/notion-cli-windows-amd64.exe
```

The `install.js` postinstall script uses these URLs as a fallback when the platform optional dependency is not available.

Users can also install manually:

```bash
# Download for your platform (example: macOS Apple Silicon)
curl -L -o notion-cli \
  https://github.com/Coastal-Programs/notion-cli/releases/download/v6.0.0/notion-cli-darwin-arm64
chmod +x notion-cli
sudo mv notion-cli /usr/local/bin/
```

---

## Go Module Publishing

Go module versions are published automatically when you push a git tag that matches Go's module versioning rules. No additional steps are needed beyond creating the tag.

```bash
# Users can install directly via Go:
go install github.com/Coastal-Programs/notion-cli/cmd/notion-cli@latest

# Or a specific version:
go install github.com/Coastal-Programs/notion-cli/cmd/notion-cli@v6.0.0
```

Go modules are served by the Go module proxy (proxy.golang.org) which automatically caches tagged versions from the git repository. After pushing a tag, the module becomes available within a few minutes.

---

## Manual npm Publishing (Fallback)

If automation fails, publish manually:

### Prerequisites

```bash
# Login to npm
npm login

# Verify access to the scoped packages
npm access ls-packages @coastal-programs
```

### Publish

```bash
# 1. Build all platform binaries
make release

# 2. Copy binaries into platform packages (see "Publishing Platform Packages" above)

# 3. Dry run to verify
npm publish --dry-run

# 4. Publish platform packages first, then the main wrapper
# (see "Publishing Platform Packages" above for the full sequence)
```

### Verify

```bash
# Check the published version
npm view @coastal-programs/notion-cli version

# Test installation
npm install -g @coastal-programs/notion-cli
notion-cli --version
```

---

## Version Checklist

Before every release, verify:

- [ ] All tests pass: `make test`
- [ ] Build succeeds: `make build`
- [ ] Lint passes: `make lint`
- [ ] CHANGELOG.md updated with new version section
- [ ] Version bumped in root `package.json`
- [ ] Version bumped in all five `npm/*/package.json` files
- [ ] `optionalDependencies` versions in root `package.json` match the new version
- [ ] `make release` produces all five platform binaries
- [ ] Git tag created and pushed
- [ ] GitHub Release created with binary attachments
- [ ] npm packages published (automated or manual)
- [ ] Verified: `npm install -g @coastal-programs/notion-cli@latest && notion-cli --version`

---

## Common Issues

**"Package already exists"**
- That version is already published on npm. Bump the version number.

**"403 Forbidden"**
- Not logged in: `npm login`
- No access to scope: `npm owner add <username> @coastal-programs/notion-cli`

**"Payment required"**
- Use `--access public` flag. Scoped packages default to private.

**Binary not found after npm install**
- The platform optional dependency may have failed to install. Run `node install.js` manually to trigger the GitHub Release fallback download.
- Alternatively, build from source: `make build` and copy the binary to your PATH.

**Go install fails**
- Ensure Go is installed and `GOPATH/bin` is in your PATH.
- Try with explicit version: `go install github.com/Coastal-Programs/notion-cli/cmd/notion-cli@v6.0.0`

**Platform package missing binary**
- Each platform `npm/*/` directory needs a `bin/` folder with the compiled binary before publishing. Run `make release` and copy binaries into place (see "Publishing Platform Packages").
