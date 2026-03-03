# Release Workflow

CRITICAL: After merging any significant changes to main, always create a release tag.

## When to Tag

- New features
- Bug fixes
- Test coverage improvements
- Documentation updates
- Any merged PR that should be released

## Version Numbering (SemVer)

- Patch (6.2.0 -> 6.2.1): bug fixes only
- Minor (6.2.0 -> 6.3.0): new features, backwards compatible
- Major (6.2.0 -> 7.0.0): breaking changes

## Version Files (ALL must match)

When bumping version, update ALL 6 files:
1. `package.json` — `version` field AND all `optionalDependencies` versions
2. `npm/notion-cli-darwin-arm64/package.json`
3. `npm/notion-cli-darwin-x64/package.json`
4. `npm/notion-cli-linux-x64/package.json`
5. `npm/notion-cli-linux-arm64/package.json`
6. `npm/notion-cli-win32-x64/package.json`

## Create Release Tag

```bash
VERSION=$(grep '"version"' package.json | head -1 | sed 's/.*"\([0-9]*\.[0-9]*\.[0-9]*\)".*/\1/')
git tag -a v$VERSION -m "Release v$VERSION"
git push origin v$VERSION
```

## Create GitHub Release

```bash
gh release create v$VERSION \
  --title "v$VERSION - Release Title" \
  --notes "## What's New
- Feature 1

## Bug Fixes
- Fix 1" \
  build/notion-cli-darwin-arm64 \
  build/notion-cli-darwin-amd64 \
  build/notion-cli-linux-amd64 \
  build/notion-cli-linux-arm64 \
  build/notion-cli-windows-amd64.exe \
  --clobber
```

## Release Checklist

1. All tests pass: `make test`
2. Build succeeds: `make build`
3. Lint passes: `make lint`
4. CHANGELOG.md updated
5. Version bumped in ALL 6 package.json files
6. `make release` produces all 5 platform binaries
7. Binaries copied into `npm/*/bin/` directories
8. Platform packages published to npm FIRST
9. Main wrapper published to npm
10. Git commit, tag, push
11. GitHub Release created with binary attachments
12. Verify: `npm view @coastal-programs/notion-cli version`

## npm Publishing Rules

- **Token type**: MUST use Classic Automation token (not Granular)
- **Why**: Granular tokens cannot create new scoped packages (404 on PUT)
- **2FA bypass**: Token must have "Bypass 2FA" enabled for automation
- **Publish order**: Platform packages FIRST, then main wrapper LAST
- **Token management**: https://www.npmjs.com/settings/jakeschepis/tokens
- **GitHub secret**: `NPM_TOKEN` in repo settings

## Automated Publishing

GitHub Actions (`publish.yml`) publishes to npm automatically when a GitHub release is created. It handles:
- Cross-compilation via `make release`
- Copying binaries to platform packages
- Publishing platform packages then main wrapper
- Version existence check (skips if already published)

Manual fallback: Use `/release` command.
