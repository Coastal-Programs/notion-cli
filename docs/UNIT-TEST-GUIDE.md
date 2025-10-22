# Unit Test Implementation Guide - Name Resolution

## Overview

This guide provides concrete, copy-paste-ready unit tests for the name resolution feature. Tests are written using Mocha and Chai, following the existing patterns in the codebase.

---

## Table of Contents

1. [Test Setup](#test-setup)
2. [Notion URL Parser Tests](#notion-url-parser-tests)
3. [Workspace Cache Tests](#workspace-cache-tests)
4. [Notion Resolver Tests](#notion-resolver-tests)
5. [Running Tests](#running-tests)

---

## Test Setup

### Prerequisites

```bash
# Install dev dependencies (should already be installed)
npm install --save-dev @types/chai @types/mocha chai mocha

# Make sure test directory exists
mkdir -p test/utils
```

### Test Configuration

The project already has `test/tsconfig.json` configured. Verify it includes:

```json
{
  "extends": "../tsconfig.json",
  "compilerOptions": {
    "types": ["mocha", "chai", "node"]
  }
}
```

---

## Notion URL Parser Tests

### File: `test/utils/notion-url-parser.test.ts`

```typescript
import { expect } from 'chai'
import { isNotionUrl, extractNotionId } from '../../src/utils/notion-url-parser'

describe('Notion URL Parser', () => {
  describe('isNotionUrl()', () => {
    it('should return true for valid Notion URLs', () => {
      const validUrls = [
        'https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00',
        'https://notion.so/1fb79d4c71bb8032b722c82305b63a00',
        'https://www.notion.so/Database-Name-1fb79d4c71bb8032b722c82305b63a00',
        'https://www.notion.so/workspace/1fb79d4c-71bb-8032-b722-c82305b63a00',
      ]

      validUrls.forEach(url => {
        expect(isNotionUrl(url)).to.be.true
      })
    })

    it('should return false for non-Notion URLs', () => {
      const invalidUrls = [
        'https://example.com',
        'https://google.com/search?q=notion',
        'not a url',
        '1fb79d4c71bb8032b722c82305b63a00', // Just an ID
      ]

      invalidUrls.forEach(url => {
        expect(isNotionUrl(url)).to.be.false
      })
    })

    it('should handle edge cases', () => {
      expect(isNotionUrl('')).to.be.false
      expect(isNotionUrl('http://notion.so/abc')).to.be.true // http (not https)
    })
  })

  describe('extractNotionId()', () => {
    it('should extract ID from URL without dashes', () => {
      const url = 'https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00'
      const id = extractNotionId(url)
      expect(id).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should extract ID from URL with dashes', () => {
      const url = 'https://www.notion.so/1fb79d4c-71bb-8032-b722-c82305b63a00'
      const id = extractNotionId(url)
      expect(id).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should extract ID from URL with page title', () => {
      const url = 'https://www.notion.so/My-Database-1fb79d4c71bb8032b722c82305b63a00'
      const id = extractNotionId(url)
      expect(id).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should normalize ID to remove dashes', () => {
      const id = extractNotionId('1fb79d4c-71bb-8032-b722-c82305b63a00')
      expect(id).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should handle direct ID input without dashes', () => {
      const id = extractNotionId('1fb79d4c71bb8032b722c82305b63a00')
      expect(id).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should throw error for invalid ID format', () => {
      expect(() => extractNotionId('invalid-id')).to.throw()
      expect(() => extractNotionId('1fb79d4c71bb8032')).to.throw() // Too short
      expect(() => extractNotionId('not-a-valid-id-at-all')).to.throw()
    })
  })
})
```

---

## Workspace Cache Tests

### File: `test/utils/workspace-cache.test.ts`

```typescript
import { expect } from 'chai'
import * as fs from 'fs/promises'
import * as path from 'path'
import * as os from 'os'
import {
  generateAliases,
  buildCacheEntry,
  saveCache,
  loadCache,
  getCacheDir,
  getCachePath,
  WorkspaceCache,
} from '../../src/utils/workspace-cache'

describe('Workspace Cache', () => {
  describe('generateAliases()', () => {
    it('should generate basic aliases for simple names', () => {
      const aliases = generateAliases('Tasks Database')

      expect(aliases).to.include('tasks database')
      expect(aliases).to.include('tasks')
      expect(aliases).to.include('task')
      expect(aliases).to.include('td')
    })

    it('should handle plural/singular conversion', () => {
      const aliases = generateAliases('Meeting Notes')

      expect(aliases).to.include('meeting notes')
      expect(aliases).to.include('meeting note') // Singular
    })

    it('should remove common suffixes', () => {
      const aliases = generateAliases('Customer Database')

      expect(aliases).to.include('customer')
      expect(aliases).to.include('customer db')
    })

    it('should create acronyms for multi-word titles', () => {
      const aliases = generateAliases('Customer Relationship Management')

      expect(aliases).to.include('crm')
    })

    it('should handle single-word titles', () => {
      const aliases = generateAliases('Tasks')

      expect(aliases).to.include('tasks')
      expect(aliases).to.include('task')
    })

    it('should handle titles with special characters', () => {
      const aliases = generateAliases('Q1 2025 - Revenue')

      expect(aliases).to.include('q1 2025 - revenue')
    })

    it('should normalize to lowercase', () => {
      const aliases = generateAliases('TASKS DATABASE')

      aliases.forEach(alias => {
        expect(alias).to.equal(alias.toLowerCase())
      })
    })

    it('should not create acronyms for single words', () => {
      const aliases = generateAliases('Tasks')

      // Should not include 't' as acronym
      expect(aliases).to.not.include('t')
    })
  })

  describe('buildCacheEntry()', () => {
    it('should build cache entry from full data source', () => {
      const dataSource: any = {
        id: '1fb79d4c71bb8032b722c82305b63a00',
        object: 'data_source',
        title: [
          {
            type: 'text',
            plain_text: 'Tasks Database',
          },
        ],
        url: 'https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00',
        properties: {
          Name: { type: 'title' },
          Status: { type: 'select' },
        },
        last_edited_time: '2025-10-22T10:00:00.000Z',
      }

      const entry = buildCacheEntry(dataSource)

      expect(entry.id).to.equal('1fb79d4c71bb8032b722c82305b63a00')
      expect(entry.title).to.equal('Tasks Database')
      expect(entry.titleNormalized).to.equal('tasks database')
      expect(entry.aliases).to.be.an('array')
      expect(entry.url).to.equal('https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00')
      expect(entry.lastEditedTime).to.equal('2025-10-22T10:00:00.000Z')
      expect(entry.properties).to.deep.equal({
        Name: { type: 'title' },
        Status: { type: 'select' },
      })
    })

    it('should handle data source with empty title', () => {
      const dataSource: any = {
        id: '1fb79d4c71bb8032b722c82305b63a00',
        object: 'data_source',
        title: [],
      }

      const entry = buildCacheEntry(dataSource)

      expect(entry.title).to.equal('Untitled')
      expect(entry.titleNormalized).to.equal('untitled')
    })

    it('should handle data source without optional fields', () => {
      const dataSource: any = {
        id: '1fb79d4c71bb8032b722c82305b63a00',
        object: 'data_source',
        title: [
          {
            type: 'text',
            plain_text: 'Tasks',
          },
        ],
      }

      const entry = buildCacheEntry(dataSource)

      expect(entry.id).to.equal('1fb79d4c71bb8032b722c82305b63a00')
      expect(entry.url).to.be.undefined
      expect(entry.lastEditedTime).to.be.undefined
      expect(entry.properties).to.deep.equal({})
    })
  })

  describe('Cache File Operations', () => {
    const TEST_CACHE_DIR = path.join(os.tmpdir(), '.notion-cli-test')
    const TEST_CACHE_FILE = path.join(TEST_CACHE_DIR, 'databases.json')

    before(async () => {
      // Override cache path for testing
      // Note: You may need to add a test-specific cache path function
      process.env.TEST_CACHE_DIR = TEST_CACHE_DIR
    })

    afterEach(async () => {
      // Clean up test cache
      try {
        await fs.rm(TEST_CACHE_DIR, { recursive: true, force: true })
      } catch (error) {
        // Ignore cleanup errors
      }
    })

    it('should save and load cache correctly', async () => {
      const cache: WorkspaceCache = {
        version: '1.0.0',
        lastSync: new Date().toISOString(),
        databases: [
          {
            id: '1fb79d4c71bb8032b722c82305b63a00',
            title: 'Tasks Database',
            titleNormalized: 'tasks database',
            aliases: ['tasks database', 'tasks', 'task', 'td'],
          },
        ],
      }

      // Save cache
      await saveCache(cache)

      // Load cache
      const loaded = await loadCache()

      expect(loaded).to.not.be.null
      expect(loaded?.version).to.equal('1.0.0')
      expect(loaded?.databases).to.have.lengthOf(1)
      expect(loaded?.databases[0].id).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should return null for missing cache', async () => {
      const loaded = await loadCache()
      expect(loaded).to.be.null
    })

    it('should handle corrupted cache gracefully', async () => {
      // Write invalid JSON
      await fs.mkdir(TEST_CACHE_DIR, { recursive: true })
      await fs.writeFile(TEST_CACHE_FILE, 'invalid json', 'utf-8')

      const loaded = await loadCache()
      expect(loaded).to.be.null
    })

    it('should create cache directory if missing', async () => {
      const cache: WorkspaceCache = {
        version: '1.0.0',
        lastSync: new Date().toISOString(),
        databases: [],
      }

      await saveCache(cache)

      const dirExists = await fs.access(TEST_CACHE_DIR).then(() => true).catch(() => false)
      expect(dirExists).to.be.true
    })
  })
})
```

---

## Notion Resolver Tests

### File: `test/utils/notion-resolver.test.ts`

```typescript
import { expect } from 'chai'
import { resolveNotionId } from '../../src/utils/notion-resolver'
import { NotionCLIError } from '../../src/errors'
import * as workspaceCache from '../../src/utils/workspace-cache'

describe('Notion Resolver', () => {
  describe('resolveNotionId() - URL input', () => {
    it('should resolve full Notion URL', async () => {
      const url = 'https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00'
      const id = await resolveNotionId(url)
      expect(id).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should resolve URL with dashes', async () => {
      const url = 'https://www.notion.so/1fb79d4c-71bb-8032-b722-c82305b63a00'
      const id = await resolveNotionId(url)
      expect(id).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should resolve URL with page title', async () => {
      const url = 'https://www.notion.so/Database-Name-1fb79d4c71bb8032b722c82305b63a00'
      const id = await resolveNotionId(url)
      expect(id).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should throw error for invalid Notion URL', async () => {
      const url = 'https://example.com/not-notion'

      try {
        await resolveNotionId(url)
        expect.fail('Should have thrown error')
      } catch (error: any) {
        expect(error).to.be.instanceOf(NotionCLIError)
        expect(error.message).to.include('Invalid Notion URL')
      }
    })
  })

  describe('resolveNotionId() - Direct ID input', () => {
    it('should accept ID without dashes', async () => {
      const id = '1fb79d4c71bb8032b722c82305b63a00'
      const result = await resolveNotionId(id)
      expect(result).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should accept ID with dashes', async () => {
      const id = '1fb79d4c-71bb-8032-b722-c82305b63a00'
      const result = await resolveNotionId(id)
      expect(result).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should normalize ID (remove dashes)', async () => {
      const id = '1fb79d4c-71bb-8032-b722-c82305b63a00'
      const result = await resolveNotionId(id)
      expect(result).to.not.include('-')
    })

    it('should reject ID with invalid characters', async () => {
      const id = '1fb79d4c71bb8032b722c82305b63aXX' // XX not valid hex

      try {
        await resolveNotionId(id)
        expect.fail('Should have thrown error')
      } catch (error: any) {
        expect(error).to.be.instanceOf(NotionCLIError)
      }
    })

    it('should reject ID with wrong length', async () => {
      const id = '1fb79d4c71bb8032' // Too short

      try {
        await resolveNotionId(id)
        expect.fail('Should have thrown error')
      } catch (error: any) {
        expect(error).to.be.instanceOf(NotionCLIError)
      }
    })
  })

  describe('resolveNotionId() - Error handling', () => {
    it('should reject null input', async () => {
      try {
        await resolveNotionId(null as any)
        expect.fail('Should have thrown error')
      } catch (error: any) {
        expect(error).to.be.instanceOf(NotionCLIError)
        expect(error.message).to.include('Invalid input')
      }
    })

    it('should reject empty string', async () => {
      try {
        await resolveNotionId('')
        expect.fail('Should have thrown error')
      } catch (error: any) {
        expect(error).to.be.instanceOf(NotionCLIError)
        expect(error.message).to.include('Invalid input')
      }
    })

    it('should reject undefined', async () => {
      try {
        await resolveNotionId(undefined as any)
        expect.fail('Should have thrown error')
      } catch (error: any) {
        expect(error).to.be.instanceOf(NotionCLIError)
        expect(error.message).to.include('Invalid input')
      }
    })

    it('should trim whitespace', async () => {
      const id = '  1fb79d4c71bb8032b722c82305b63a00  '
      const result = await resolveNotionId(id)
      expect(result).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should show helpful error for name-like input', async () => {
      try {
        await resolveNotionId('Tasks Database')
        expect.fail('Should have thrown error')
      } catch (error: any) {
        expect(error).to.be.instanceOf(NotionCLIError)
        expect(error.message).to.include('not found')
        expect(error.message).to.include('coming soon')
      }
    })
  })

  // TODO: Add these tests after cache lookup is implemented
  describe.skip('resolveNotionId() - Cache lookup (TODO)', () => {
    beforeEach(async () => {
      // Mock cache with test data
      const testCache = {
        version: '1.0.0',
        lastSync: new Date().toISOString(),
        databases: [
          {
            id: '1fb79d4c71bb8032b722c82305b63a00',
            title: 'Tasks Database',
            titleNormalized: 'tasks database',
            aliases: ['tasks database', 'tasks', 'task', 'td'],
          },
          {
            id: '2a8c3d5e71bb8042b833d94316c74b11',
            title: 'Meeting Notes',
            titleNormalized: 'meeting notes',
            aliases: ['meeting notes', 'meetings', 'notes', 'mn'],
          },
        ],
      }

      await workspaceCache.saveCache(testCache)
    })

    it('should resolve by exact title', async () => {
      const result = await resolveNotionId('Tasks Database')
      expect(result).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should be case-insensitive', async () => {
      const result = await resolveNotionId('tasks database')
      expect(result).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should resolve by alias', async () => {
      const result = await resolveNotionId('tasks')
      expect(result).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should resolve by acronym', async () => {
      const result = await resolveNotionId('td')
      expect(result).to.equal('1fb79d4c71bb8032b722c82305b63a00')
    })

    it('should throw error for not found', async () => {
      try {
        await resolveNotionId('Nonexistent Database')
        expect.fail('Should have thrown error')
      } catch (error: any) {
        expect(error).to.be.instanceOf(NotionCLIError)
        expect(error.message).to.include('not found')
      }
    })
  })
})
```

---

## Running Tests

### Run All Tests

```bash
npm test
```

### Run Specific Test File

```bash
npm test -- test/utils/notion-url-parser.test.ts
```

### Run Tests with Coverage

```bash
npm test -- --coverage
```

### Run Tests in Watch Mode (for development)

```bash
npm test -- --watch
```

### Run Tests with Debug Output

```bash
TEST_DEBUG=true npm test
```

---

## Test Patterns & Best Practices

### Pattern 1: Arrange-Act-Assert (AAA)

```typescript
it('should do something', async () => {
  // Arrange - Set up test data
  const input = 'test input'

  // Act - Call the function
  const result = await functionUnderTest(input)

  // Assert - Check the result
  expect(result).to.equal('expected output')
})
```

### Pattern 2: Testing Async Functions

```typescript
it('should handle async operations', async () => {
  const result = await asyncFunction()
  expect(result).to.not.be.null
})
```

### Pattern 3: Testing Errors

```typescript
it('should throw error for invalid input', async () => {
  try {
    await functionThatThrows()
    expect.fail('Should have thrown error')
  } catch (error: any) {
    expect(error).to.be.instanceOf(NotionCLIError)
    expect(error.message).to.include('expected error text')
  }
})
```

### Pattern 4: Using Hooks

```typescript
describe('Feature', () => {
  before(() => {
    // Runs once before all tests in this describe block
  })

  beforeEach(() => {
    // Runs before each test
  })

  afterEach(() => {
    // Runs after each test (cleanup)
  })

  after(() => {
    // Runs once after all tests
  })

  it('test 1', () => { /* ... */ })
  it('test 2', () => { /* ... */ })
})
```

### Pattern 5: Skipping Tests

```typescript
// Skip a single test
it.skip('test not ready yet', () => {
  // Will not run
})

// Skip entire suite
describe.skip('Feature not ready', () => {
  it('test 1', () => { /* ... */ })
  it('test 2', () => { /* ... */ })
})
```

---

## Integration with CI/CD

### GitHub Actions Example

Add to `.github/workflows/test.yml`:

```yaml
name: Tests

on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest

    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'

      - name: Install dependencies
        run: npm install

      - name: Run tests
        run: npm test

      - name: Check build
        run: npm run build
```

---

## Next Steps

1. **Copy test files** to appropriate locations
2. **Run tests** to verify they work
3. **Add more tests** as features are implemented
4. **Update tests** when bugs are found
5. **Monitor coverage** to ensure comprehensive testing

---

## Troubleshooting

### Issue: "Cannot find module"

**Solution:**
```bash
# Rebuild TypeScript
npm run build

# Re-run tests
npm test
```

### Issue: "Timeout exceeded"

**Solution:**
```typescript
// Increase timeout for slow tests
it('slow test', async function() {
  this.timeout(5000) // 5 seconds
  // ... test code
})
```

### Issue: "Port already in use" (for API mocks)

**Solution:**
```typescript
// Use random ports in tests
const port = Math.floor(Math.random() * 10000) + 10000
```

---

## Additional Resources

- [Mocha Documentation](https://mochajs.org/)
- [Chai Assertion Library](https://www.chaijs.com/)
- [Existing Test Examples](../test/commands/)
- [NAME-RESOLUTION-TESTS.md](./NAME-RESOLUTION-TESTS.md) - Integration test scenarios
