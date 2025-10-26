# Contributing to notion-cli

Thank you for your interest in contributing to notion-cli! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Code Style Guidelines](#code-style-guidelines)
- [Testing Requirements](#testing-requirements)
- [Pull Request Process](#pull-request-process)
- [Commit Message Format](#commit-message-format)
- [Project Structure](#project-structure)
- [Reporting Issues](#reporting-issues)

## Code of Conduct

This project follows a simple code of conduct: be respectful, constructive, and collaborative. We welcome contributions from everyone.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/notion-cli.git
   cd notion-cli
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/Coastal-Programs/notion-cli.git
   ```
4. **Create a branch** for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Development Setup

### Prerequisites

- Node.js >= 18.0.0
- npm >= 8.0.0
- Git

### Installation

```bash
# Install dependencies
npm install

# Build the project
npm run build

# Link for local testing
npm link
```

### Environment Setup

Create a `.env` file or set environment variables:

```bash
export NOTION_TOKEN="secret_your_token_here"
```

Get your token from: https://www.notion.so/my-integrations

## Code Style Guidelines

### TypeScript

- Use TypeScript for all new code
- Enable strict type checking
- Avoid `any` types when possible
- Add return types to all functions
- Use interfaces for object shapes

**Example:**

```typescript
interface PageCreateOptions {
  databaseId: string;
  properties: Record<string, any>;
}

async function createPage(options: PageCreateOptions): Promise<Page> {
  // Implementation
}
```

### ESLint

This project uses **ESLint v9** with flat config. All code must pass linting:

```bash
# Run linter
npm run lint

# Auto-fix issues
npm run lint -- --fix
```

**Key ESLint Rules:**

- No unused variables
- Consistent indentation (2 spaces)
- Single quotes for strings
- Semicolons required
- No console.log in production code (use debug logger)

### Code Formatting

We use **Prettier** for consistent formatting:

- 2 spaces for indentation
- Single quotes
- Semicolons required
- Trailing commas where valid
- Max line length: 100 characters

### Naming Conventions

- **Files:** kebab-case (`db-query.ts`)
- **Classes:** PascalCase (`DbQuery`)
- **Functions:** camelCase (`retrieveDatabase`)
- **Constants:** SCREAMING_SNAKE_CASE (`DEFAULT_TIMEOUT`)
- **Interfaces:** PascalCase with descriptive names (`NotionAPIResponse`)

### Documentation

All public APIs must have JSDoc comments:

```typescript
/**
 * Retrieve a database by ID
 *
 * @param databaseId - The ID of the database to retrieve
 * @param options - Optional retrieval options
 * @returns Promise resolving to database object
 * @throws {NotionCLIError} If database not found or API error occurs
 *
 * @example
 * ```typescript
 * const db = await retrieveDatabase('abc123');
 * console.log(db.title);
 * ```
 */
async function retrieveDatabase(
  databaseId: string,
  options?: RetrievalOptions
): Promise<Database> {
  // Implementation
}
```

## Testing Requirements

### Running Tests

```bash
# Run all tests
npm test

# Run specific test file
npm test -- test/commands/db/query.test.ts

# Run tests with verbose output
npm test -- --reporter spec
```

### Test Coverage

- All new features must include tests
- Aim for 80%+ code coverage
- Test both success and error cases
- Use mocks for external API calls

### Test Structure

We use **Mocha** and **Chai** for testing:

```typescript
import { expect } from 'chai'
import { describe, it } from 'mocha'

describe('db query command', () => {
  it('should query database successfully', async () => {
    // Arrange
    const databaseId = 'test-id'

    // Act
    const result = await queryDatabase(databaseId)

    // Assert
    expect(result).to.have.property('results')
  })

  it('should handle errors gracefully', async () => {
    // Test error case
  })
})
```

### Test Guidelines

1. **Mock external dependencies** - Don't make real API calls in tests
2. **Use descriptive test names** - Clearly state what is being tested
3. **Follow AAA pattern** - Arrange, Act, Assert
4. **Test edge cases** - Empty inputs, null values, large datasets
5. **Keep tests isolated** - No dependencies between tests

## Pull Request Process

### Before Submitting

1. **Update from upstream**:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Run all checks**:
   ```bash
   npm run build
   npm test
   npm run lint
   ```

3. **Update documentation** if needed:
   - Update README.md for new features
   - Add CHANGELOG.md entry
   - Update JSDoc comments

### Submitting

1. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

2. **Create Pull Request** on GitHub with:
   - Clear title describing the change
   - Detailed description of what changed and why
   - Reference any related issues (`Fixes #123`)
   - Screenshots for UI changes

3. **Fill out PR template** completely

### PR Review Process

- Maintainers will review within 1-2 weeks
- Address review feedback promptly
- Keep PRs focused on a single feature/fix
- Be open to suggestions and changes

### PR Checklist

- [ ] Code follows style guidelines
- [ ] All tests pass
- [ ] New tests added for new features
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Commit messages follow format
- [ ] No merge conflicts
- [ ] Build succeeds

## Commit Message Format

We follow **Conventional Commits** specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Types

- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

### Examples

```
feat(db): add schema discovery command

Implement new 'db schema' command to extract database schemas
in AI-parseable format. Supports JSON, YAML, and markdown output.

Closes #42
```

```
fix(cache): resolve race condition in cache invalidation

Fixed issue where concurrent writes could corrupt cache state.
Added mutex lock for write operations.

Fixes #56
```

```
docs: update README with simple properties examples

Added comprehensive examples for simple properties mode,
including all 13 supported property types.
```

### Commit Guidelines

- Use present tense ("add feature" not "added feature")
- Use imperative mood ("move cursor to..." not "moves cursor to...")
- Keep subject line under 72 characters
- Reference issues and PRs in footer
- Explain "what" and "why", not "how"

## Project Structure

```
notion-cli/
├── src/
│   ├── commands/          # CLI command implementations
│   │   ├── db/            # Database commands
│   │   ├── page/          # Page commands
│   │   ├── block/         # Block commands
│   │   ├── user/          # User commands
│   │   └── ...
│   ├── utils/             # Utility functions
│   │   ├── schema-extractor.ts
│   │   ├── property-expander.ts
│   │   ├── workspace-cache.ts
│   │   └── ...
│   ├── base-command.ts    # Base command class
│   ├── base-flags.ts      # Reusable CLI flags
│   ├── envelope.ts        # JSON response formatting
│   ├── notion.ts          # Notion API client wrapper
│   ├── cache.ts           # In-memory caching
│   └── errors.ts          # Error handling
├── test/                  # Test files
│   ├── commands/          # Command tests
│   └── utils/             # Utility tests
├── docs/                  # Documentation
├── dist/                  # Compiled output (gitignored)
└── package.json           # Project config
```

## Reporting Issues

### Bug Reports

Include:
- Clear description of the issue
- Steps to reproduce
- Expected vs actual behavior
- Environment (Node version, OS, npm version)
- Error messages and stack traces
- Minimal reproduction example

### Feature Requests

Include:
- Clear description of the feature
- Use case and motivation
- Example usage
- Potential implementation approach

### Security Vulnerabilities

**Do not open public issues for security vulnerabilities.**

See [SECURITY.md](SECURITY.md) for reporting instructions.

## Development Tips

### Debugging

Enable debug logging:

```bash
export DEBUG=notion-cli:*
notion-cli db query <id> --verbose
```

### Testing Local Changes

```bash
# Link local version
npm link

# Test commands
notion-cli --version
notion-cli db query <id>

# Unlink when done
npm unlink -g @coastal-programs/notion-cli
```

### Working with oclif

This project uses **oclif v2** framework:

- Commands extend `Command` class
- Flags defined with `@oclif/core` Flags
- Use `this.log()` for output, not `console.log()`
- Exit with `process.exit(0)` for success, `process.exit(1)` for errors

### Common Tasks

```bash
# Add a new command
npm run generate:command

# Rebuild on file changes
npm run build -- --watch

# Clear cache during development
rm -rf ~/.notion-cli/

# Run single test file
npm test -- test/commands/db/retrieve.test.ts
```

## Questions?

- Check existing issues and PRs first
- Open a discussion on GitHub Discussions
- Review documentation in `/docs` folder

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to notion-cli! 🚀
