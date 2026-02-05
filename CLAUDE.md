# Claude Code Development Guidelines

This document provides guidelines for Claude Code when working on the notion-cli project.

## 🚨 Critical Checklist: Before Completing Any Task

### 1. Code Changes
- [ ] All tests pass locally (`npm test`)
- [ ] Code coverage is maintained or improved
- [ ] Linting passes (`npm run lint`)
- [ ] Build completes successfully (`npm run build`)

### 2. Documentation
- [ ] CHANGELOG.md updated with changes
- [ ] Relevant docs updated (if applicable)
- [ ] README updated (if user-facing changes)

### 3. Git Workflow
- [ ] Commits follow conventional commit format
- [ ] Changes pushed to feature branch
- [ ] Pull request created (if needed)

### 4. 🎯 **RELEASE TAG (CRITICAL)** 🎯

**IMPORTANT: After merging significant changes, always create a release tag!**

#### When to Create a Release Tag:
- ✅ After adding new features
- ✅ After bug fixes
- ✅ After improving test coverage
- ✅ After documentation updates
- ✅ After any merged PR that should be released

#### How to Create a Release Tag:

```bash
# 1. Ensure you're on main branch with latest changes
git checkout main
git pull origin main

# 2. Check current version
cat package.json | grep '"version"'

# 3. Create and push the tag (use the version from package.json)
git tag -a v5.8.0 -m "Release v5.8.0: Brief description"
git push origin v5.8.0

# 4. Create GitHub Release
gh release create v5.8.0 \
  --title "v5.8.0 - Release Title" \
  --notes "$(cat <<'EOF'
## What's New

- Feature 1
- Feature 2

## Bug Fixes

- Fix 1
- Fix 2

## Improvements

- Improvement 1
- Improvement 2

EOF
)"
```

#### Quick Release Tag Command (Copy-Paste Ready):

```bash
# Get version and create release in one go
VERSION=$(cat package.json | grep '"version"' | head -1 | awk -F'"' '{print $4}')
git tag -a v$VERSION -m "Release v$VERSION"
git push origin v$VERSION
echo "✅ Created and pushed tag v$VERSION"
echo "📝 Now create GitHub release at: https://github.com/Coastal-Programs/notion-cli/releases/new?tag=v$VERSION"
```

### 5. Verify Release
- [ ] Tag appears at https://github.com/Coastal-Programs/notion-cli/tags
- [ ] GitHub Release created at https://github.com/Coastal-Programs/notion-cli/releases
- [ ] CI/CD pipeline triggered (if configured)
- [ ] npm package published (if automated)

---

## 📋 Standard Development Workflow

### Starting a New Task
1. Read relevant documentation in `/docs`
2. Review existing code patterns
3. Check CHANGELOG.md for recent changes
4. Create feature branch: `git checkout -b feature/description`

### During Development
1. Write tests first (TDD approach)
2. Implement functionality
3. Ensure 90%+ code coverage
4. Follow existing code style
5. Add inline documentation for complex logic

### Committing Changes
```bash
# Use conventional commits
git commit -m "feat: add new feature"
git commit -m "fix: resolve bug in table formatter"
git commit -m "test: add comprehensive tests for utility"
git commit -m "docs: update API documentation"
git commit -m "chore: update dependencies"
```

### Creating Pull Requests
```bash
# Push branch
git push origin feature/description

# Create PR with gh CLI
gh pr create \
  --title "feat: Brief description" \
  --body "## Summary

Details about the changes...

## Testing
- [ ] Tests pass
- [ ] Coverage maintained"
```

---

## 🧪 Testing Guidelines

### Coverage Requirements
- **Minimum**: 90% line coverage
- **Target**: 95%+ coverage for utilities
- **100% coverage**: Critical paths (auth, API calls, data formatting)

### Test Structure
```typescript
describe('feature-name', () => {
  describe('functionality group', () => {
    it('should do something specific', () => {
      // Arrange
      const data = setupTestData()

      // Act
      const result = functionUnderTest(data)

      // Assert
      expect(result).to.equal(expected)
    })
  })
})
```

### Running Tests
```bash
# Run all tests
npm test

# Run specific test file
npm test -- test/utils/table-formatter.test.ts

# Run with coverage
npm run test:coverage

# Run tests matching pattern
npm test -- --grep "table-formatter"
```

---

## 📝 Documentation Standards

### Code Documentation
- Use JSDoc for public APIs
- Add inline comments for complex logic
- Keep comments concise and meaningful

### User Documentation
- Update `/docs` for user-facing changes
- Include examples in documentation
- Keep README.md in sync with features

### Changelog
- Update CHANGELOG.md for every release
- Follow Keep a Changelog format
- Group changes: Added, Changed, Fixed, Deprecated, Removed

---

## 🔍 Code Review Checklist

Before requesting review:
- [ ] Code follows project conventions
- [ ] No console.log or debug code
- [ ] Error handling is comprehensive
- [ ] Tests cover edge cases
- [ ] Documentation is updated
- [ ] No breaking changes (or clearly documented)
- [ ] Performance impact considered

---

## 🚀 Release Process

### Version Numbering (Semantic Versioning)
- **Patch** (5.8.0 → 5.8.1): Bug fixes only
- **Minor** (5.8.0 → 5.9.0): New features, backwards compatible
- **Major** (5.8.0 → 6.0.0): Breaking changes

### Release Checklist
1. [ ] All tests pass in CI
2. [ ] CHANGELOG.md updated
3. [ ] Version bumped in package.json (if not already)
4. [ ] Changes merged to main
5. [ ] **Release tag created** (see above)
6. [ ] GitHub Release published
7. [ ] npm package published (automated)
8. [ ] Verify installation: `npm install -g @coastal-programs/notion-cli@latest`

---

## 🎯 Project-Specific Notes

### Key Files
- `src/utils/table-formatter.ts` - Table rendering utility (100% coverage required)
- `src/commands/` - CLI commands (test through integration tests)
- `test/` - Mocha + Chai test suite
- `docs/` - User documentation

### Important Commands
```bash
# Build project
npm run build

# Lint code
npm run lint

# Run all checks
npm run build && npm test && npm run lint

# Local package test
npm pack && npm install -g ./coastal-programs-notion-cli-*.tgz
```

### Code Patterns
- Use `formatTable()` for all table output
- Use `NotionCLIError` for error handling
- Use `this.log()` for command output (not console.log)
- Follow oclif v4 command structure

---

## 🤖 Claude Code Reminders

### Don't Forget!
1. ✅ Run tests after every change
2. ✅ Update CHANGELOG.md
3. ✅ **CREATE RELEASE TAG** after merging
4. ✅ Check CI status after pushing
5. ✅ Verify coverage didn't decrease

### Quick Commands for Claude
```bash
# Full quality check
npm run build && npm test && npm run lint

# Coverage check
npx nyc --reporter=text npm test -- --grep "your-feature"

# Create release (NEVER FORGET THIS!)
VERSION=$(cat package.json | grep '"version"' | head -1 | awk -F'"' '{print $4}')
git tag -a v$VERSION -m "Release v$VERSION" && git push origin v$VERSION
```

---

## 📚 Additional Resources

- [PUBLISHING.md](./PUBLISHING.md) - Detailed publishing guide
- [CONTRIBUTING.md](./CONTRIBUTING.md) - Contribution guidelines
- [CHANGELOG.md](./CHANGELOG.md) - Release history
- [GitHub Releases](https://github.com/Coastal-Programs/notion-cli/releases) - View releases
