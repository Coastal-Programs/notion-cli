# Publishing notion-cli

notion-cli is a Go binary distributed through three channels:

1. **npm** (primary) -- platform-specific binary packages via `npm install -g @coastal-programs/notion-cli`
2. **GitHub Releases** -- direct binary downloads for all platforms
3. **Go module** -- `go install github.com/Coastal-Programs/notion-cli/v6/cmd/notion-cli@latest`

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

#### 2. Add Token to the `npm-publish` GitHub Environment

The `publish-npm` job in `.github/workflows/publish.yml` declares `environment: npm-publish`. That gate means GitHub Actions will **only** inject secrets that live inside the `npm-publish` environment when the job runs — repository-level secrets are not auto-inherited into environment-gated jobs. The job will also pause for manual approval if "Required reviewers" is configured on the environment.

Follow these steps exactly. Do **not** add `NPM_TOKEN` as a repository secret — a repo-level `NPM_TOKEN` will be invisible to the gated job and `npm publish` will fail with an authentication error.

1. Go to https://github.com/Coastal-Programs/notion-cli/settings/environments
2. Click **"New environment"**.
3. Name it exactly `npm-publish` (must match the `environment:` value in `publish.yml` character-for-character — no typos, no trailing spaces).
4. Click **"Configure environment"**.
5. Under **"Deployment protection rules"**, check **"Required reviewers"** and add the maintainer's GitHub username (and any co-maintainers who should be able to approve a release). Click **"Save protection rules"**.
   - This is intentional: every release will now pause and wait for a one-click manual approval in the Actions UI before any `npm publish` runs. The reviewer gets an email and a button in the workflow run page.
6. Scroll down to **"Environment secrets"** and click **"Add secret"**.
7. Name: `NPM_TOKEN`
8. Value: Paste the npm automation token you generated in step 1 (the one starting with `npm_...`).
9. Click **"Add secret"**.

That is the entire one-time setup.

#### Why the environment gate?

Without the `npm-publish` environment, anyone who can push a commit that modifies `.github/workflows/publish.yml` — or anyone who can trigger the workflow via a compromised dependency, a malicious PR from a fork that somehow runs against `main`, or a stolen contributor credential — could publish an arbitrary package to npm under the `@coastal-programs` scope using the stored token. Gating the `publish-npm` job on a protected environment with required reviewers means an actual human has to click "Approve and deploy" in the GitHub UI for every release, and that human will see the exact workflow run and tag they are approving. The token itself is also scoped to the environment, so a workflow file that does not target `environment: npm-publish` cannot read it at all. This is the same pattern recommended by GitHub's ["securing deployments"](https://docs.github.com/en/actions/deployment/targeting-different-environments/using-environments-for-deployment) guidance for any job that publishes to a public registry.

#### Repository secrets vs. environment secrets in this repo

To avoid confusion later: this repository uses **both** kinds of secrets, and they are not interchangeable.

| Secret | Scope | Used by | Why |
| --- | --- | --- | --- |
| `NPM_TOKEN` | **Environment secret** on `npm-publish` | The `publish-npm` job (has `environment: npm-publish`) | High-impact publish credential; gated behind manual approval. |
| `NOTION_OAUTH_CLIENT_ID` | **Repository secret** | The `build` job and the `publish-npm` job's build step (neither job's reads of these are environment-gated for compile-time embedding) | Build-time constants; no manual gate needed. |
| `NOTION_OAUTH_SECRET` | **Repository secret** | Same as above | Build-time constants; no manual gate needed. |

If you add `NPM_TOKEN` at the repository level it will silently not be read by the gated job. If you add `NOTION_OAUTH_CLIENT_ID` / `NOTION_OAUTH_SECRET` only at the environment level, the un-gated `build` job will fail its "Verify OAuth secrets" step. Put each secret where the table says.

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

# 8. GitHub Actions builds, then waits for your approval, then publishes to npm
```

### What Happens Automatically

When you create a GitHub Release, the publish workflow:

- Checks out the code
- Runs `make test` (Go test suite)
- Runs `make build` (Go binary build)
- **Pauses on the `publish-npm` job and waits for a required reviewer to approve the `npm-publish` environment deployment.** You will get an email; the workflow run page will show an "Approve and deploy" button. Nothing publishes until you click it.
- After approval: checks if the version already exists on npm
- Publishes each platform package with its respective binary
- Publishes the main wrapper package to npm with provenance

If the `publish-npm` job is stuck in a yellow "Waiting" state, that is the manual approval gate — not a bug. Open the run, click the job, and approve.

### Manual Workflow Trigger

You can also trigger the publish workflow manually:

1. Go to Actions tab: https://github.com/Coastal-Programs/notion-cli/actions/workflows/publish.yml
2. Click "Run workflow"
3. Click "Run workflow" button
4. The `publish-npm` job will still wait for environment approval — manual triggers do not bypass the gate.

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
go install github.com/Coastal-Programs/notion-cli/v6/cmd/notion-cli@latest

# Or a specific version:
go install github.com/Coastal-Programs/notion-cli/v6/cmd/notion-cli@v6.4.0
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
- Try with explicit version: `go install github.com/Coastal-Programs/notion-cli/v6/cmd/notion-cli@v6.4.0`

**Platform package missing binary**
- Each platform `npm/*/` directory needs a `bin/` folder with the compiled binary before publishing. Run `make release` and copy binaries into place (see "Publishing Platform Packages").
