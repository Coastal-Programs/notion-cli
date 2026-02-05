# Envelope Testing Strategy

## Overview

This document outlines the comprehensive testing strategy for the JSON envelope standardization system. It covers unit tests, integration tests, end-to-end tests, and manual testing procedures.

**Target Coverage:** 90%+ for envelope-related code
**Test Framework:** Mocha + Chai + @oclif/test
**CI/CD Integration:** GitHub Actions

## Test Levels

### 1. Unit Tests

**Purpose:** Test individual components in isolation
**Location:** `test/unit/`
**Framework:** Mocha + Chai

#### 1.1 EnvelopeFormatter Tests

**File:** `test/unit/envelope.test.ts`

**Test Cases:**

```typescript
describe('EnvelopeFormatter', () => {
  describe('constructor', () => {
    it('should initialize with command name and version')
    it('should record start time for execution tracking')
  })

  describe('wrapSuccess', () => {
    it('should create success envelope with data')
    it('should include all required metadata fields')
    it('should track execution time accurately')
    it('should accept additional metadata')
    it('should handle null/undefined data')
    it('should handle large data objects')
    it('should preserve data type information')
  })

  describe('wrapError', () => {
    it('should wrap NotionCLIError correctly')
    it('should wrap standard Error objects')
    it('should wrap raw error objects')
    it('should generate suggestions for known error codes')
    it('should handle errors without details')
    it('should include notionError if present')
    it('should add additional context when provided')
  })

  describe('outputEnvelope', () => {
    it('should output pretty JSON by default')
    it('should output compact JSON with --compact-json')
    it('should output raw data with --raw')
    it('should use provided log function')
    it('should handle undefined log function gracefully')
  })

  describe('getExitCode', () => {
    it('should return 0 for success envelope')
    it('should return 1 for API errors')
    it('should return 2 for validation errors')
    it('should return 2 for CLI errors')
    it('should return 1 for unknown errors')
  })

  describe('writeDiagnostic', () => {
    it('should write to stderr')
    it('should prefix message with level')
    it('should support info level')
    it('should support warn level')
    it('should support error level')
  })

  describe('logRetry', () => {
    it('should format retry message correctly')
    it('should write to stderr')
  })

  describe('logCacheHit', () => {
    it('should write to stderr when DEBUG=true')
    it('should not write when DEBUG=false')
  })
})
```

#### 1.2 BaseCommand Tests

**File:** `test/unit/base-command.test.ts`

**Test Cases:**

```typescript
describe('BaseCommand', () => {
  describe('init', () => {
    it('should create envelope formatter')
    it('should extract command name from id')
    it('should use config version')
  })

  describe('checkEnvelopeUsage', () => {
    it('should return true for --json flag')
    it('should return true for --compact-json flag')
    it('should return false for no JSON flags')
  })

  describe('outputSuccess', () => {
    it('should create and output success envelope')
    it('should respect --json flag')
    it('should respect --compact-json flag')
    it('should call process.exit(0)')
    it('should include additional metadata')
  })

  describe('outputError', () => {
    it('should create and output error envelope')
    it('should wrap non-NotionCLIError errors')
    it('should call process.exit with correct code')
    it('should output JSON in JSON mode')
    it('should use oclif error in non-JSON mode')
    it('should include additional context')
  })

  describe('getExitCodeForError', () => {
    it('should return 2 for VALIDATION_ERROR')
    it('should return 1 for other errors')
  })
})
```

#### 1.3 Error Code Mapping Tests

**File:** `test/unit/error-codes.test.ts`

**Test Cases:**

```typescript
describe('Error Code Mapping', () => {
  describe('getExitCodeForError', () => {
    it('should map VALIDATION_ERROR to exit code 2')
    it('should map CLI_ERROR to exit code 2')
    it('should map CONFIG_ERROR to exit code 2')
    it('should map INVALID_ARGUMENT to exit code 2')
    it('should map UNAUTHORIZED to exit code 1')
    it('should map NOT_FOUND to exit code 1')
    it('should map RATE_LIMITED to exit code 1')
    it('should map API_ERROR to exit code 1')
    it('should map UNKNOWN to exit code 1')
  })

  describe('generateSuggestions', () => {
    it('should suggest token check for UNAUTHORIZED')
    it('should suggest resource verification for NOT_FOUND')
    it('should suggest retry for RATE_LIMITED')
    it('should suggest syntax check for VALIDATION_ERROR')
    it('should suggest config for CLI_ERROR')
    it('should return empty array for unknown codes')
  })
})
```

### 2. Integration Tests

**Purpose:** Test commands with envelope integration
**Location:** `test/integration/`
**Framework:** @oclif/test + Mocha + Chai

#### 2.1 Page Commands

**File:** `test/integration/page-commands.test.ts`

**Test Cases:**

```typescript
describe('page retrieve', () => {
  describe('with --json flag', () => {
    it('should return success envelope with page data')
    it('should include all metadata fields')
    it('should exit with code 0')
  })

  describe('with --compact-json flag', () => {
    it('should return single-line envelope')
    it('should be valid JSON')
    it('should include all envelope fields')
  })

  describe('with --raw flag', () => {
    it('should return raw page data without envelope')
    it('should not include success field')
    it('should not include metadata field')
  })

  describe('with invalid page ID', () => {
    it('should return error envelope')
    it('should include error code')
    it('should include suggestions')
    it('should exit with code 1')
  })

  describe('with missing NOTION_TOKEN', () => {
    it('should return UNAUTHORIZED error')
    it('should suggest token configuration')
    it('should exit with code 1')
  })
})

describe('page create', () => {
  describe('with --json flag', () => {
    it('should return success envelope with created page')
    it('should include operation metadata')
  })

  describe('with missing required argument', () => {
    it('should return VALIDATION_ERROR')
    it('should exit with code 2')
  })
})
```

#### 2.2 Database Commands

**File:** `test/integration/db-commands.test.ts`

**Test Cases:**

```typescript
describe('db query', () => {
  describe('with --json flag', () => {
    it('should return success envelope with results')
    it('should include pagination metadata')
    it('should include total_results')
    it('should include page_size')
  })

  describe('with --pageAll flag', () => {
    it('should return all results in envelope')
    it('should include total count in metadata')
  })

  describe('with invalid filter JSON', () => {
    it('should return VALIDATION_ERROR')
    it('should include parse error details')
    it('should exit with code 2')
  })
})

describe('db retrieve', () => {
  describe('with --json flag', () => {
    it('should return success envelope with database')
    it('should include schema information')
  })
})
```

#### 2.3 Search Commands

**File:** `test/integration/search-commands.test.ts`

**Test Cases:**

```typescript
describe('search', () => {
  describe('with --json flag', () => {
    it('should return success envelope with results')
    it('should include query metadata')
    it('should include filter metadata')
    it('should include has_more flag')
  })

  describe('with no results', () => {
    it('should return empty results array')
    it('should have total_results: 0')
  })
})
```

### 3. End-to-End Tests

**Purpose:** Test complete workflows with real Notion API
**Location:** `test/e2e/`
**Prerequisites:** Valid NOTION_TOKEN, test workspace

#### 3.1 Complete Workflow Tests

**File:** `test/e2e/page-lifecycle.test.ts`

**Test Cases:**

```typescript
describe('Page Lifecycle E2E', () => {
  it('should create, retrieve, update, and delete page with envelopes', async () => {
    // 1. Create page
    const createResult = await runCommand(['page', 'create', DB_ID, '--title', 'Test', '--json'])
    const createEnvelope = JSON.parse(createResult.stdout)
    expect(createEnvelope.success).to.be.true
    expect(createEnvelope.metadata.operation).to.equal('create')
    const pageId = createEnvelope.data.id

    // 2. Retrieve page
    const retrieveResult = await runCommand(['page', 'retrieve', pageId, '--json'])
    const retrieveEnvelope = JSON.parse(retrieveResult.stdout)
    expect(retrieveEnvelope.success).to.be.true
    expect(retrieveEnvelope.data.id).to.equal(pageId)

    // 3. Update page
    const updateResult = await runCommand(['page', 'update', pageId, '--archived', 'true', '--json'])
    const updateEnvelope = JSON.parse(updateResult.stdout)
    expect(updateEnvelope.success).to.be.true
    expect(updateEnvelope.data.archived).to.be.true

    // 4. Verify envelope metadata consistency
    expect(createEnvelope.metadata.version).to.equal(retrieveEnvelope.metadata.version)
    expect(retrieveEnvelope.metadata.version).to.equal(updateEnvelope.metadata.version)
  })
})
```

#### 3.2 Error Handling E2E

**File:** `test/e2e/error-handling.test.ts`

**Test Cases:**

```typescript
describe('Error Handling E2E', () => {
  it('should handle rate limiting with proper envelope', async () => {
    // Make many rapid requests to trigger rate limit
    // Verify error envelope with RATE_LIMITED code
  })

  it('should handle network errors gracefully', async () => {
    // Simulate network failure
    // Verify error envelope and suggestions
  })

  it('should handle authorization errors', async () => {
    // Use invalid token
    // Verify UNAUTHORIZED error with suggestions
  })
})
```

### 4. Snapshot Tests

**Purpose:** Ensure envelope structure stability
**Location:** `test/snapshots/`

**Test Cases:**

```typescript
describe('Envelope Snapshots', () => {
  it('should match success envelope snapshot', () => {
    const envelope = formatter.wrapSuccess({ id: 'test', object: 'page' })
    // Remove dynamic fields
    envelope.metadata.timestamp = 'TIMESTAMP'
    envelope.metadata.execution_time_ms = 0
    expect(envelope).to.matchSnapshot()
  })

  it('should match error envelope snapshot', () => {
    const error = new NotionCLIError('NOT_FOUND', 'Not found')
    const envelope = formatter.wrapError(error)
    envelope.metadata.timestamp = 'TIMESTAMP'
    expect(envelope).to.matchSnapshot()
  })
})
```

## Test Data

### Mock Data

**File:** `test/fixtures/mock-data.ts`

```typescript
export const mockPage: PageObjectResponse = {
  object: 'page',
  id: 'abc-123',
  created_time: '2025-01-01T00:00:00.000Z',
  last_edited_time: '2025-01-01T00:00:00.000Z',
  properties: {
    title: {
      type: 'title',
      title: [{ text: { content: 'Test Page' } }]
    }
  },
  url: 'https://notion.so/abc-123',
  // ... rest of page object
}

export const mockDatabase: DatabaseObjectResponse = {
  object: 'database',
  id: 'db-123',
  title: [{ text: { content: 'Test DB' } }],
  properties: {},
  // ... rest of database object
}

export const mockSuccessEnvelope: SuccessEnvelope<PageObjectResponse> = {
  success: true,
  data: mockPage,
  metadata: {
    timestamp: '2025-01-01T00:00:00.000Z',
    command: 'page retrieve',
    execution_time_ms: 123,
    version: '5.4.0',
  }
}

export const mockErrorEnvelope: ErrorEnvelope = {
  success: false,
  error: {
    code: 'NOT_FOUND',
    message: 'Resource not found',
    details: { resourceId: 'abc-123' },
    suggestions: [
      'Verify the resource ID is correct',
      'Try running: notion-cli sync'
    ]
  },
  metadata: {
    timestamp: '2025-01-01T00:00:00.000Z',
    command: 'page retrieve',
    execution_time_ms: 56,
    version: '5.4.0',
  }
}
```

### Test Utilities

**File:** `test/helpers/test-utils.ts`

```typescript
import { spawn } from 'child_process'

/**
 * Run CLI command and capture stdout, stderr, and exit code
 */
export async function runCommand(args: string[]): Promise<{
  stdout: string
  stderr: string
  exitCode: number
}> {
  return new Promise((resolve) => {
    const child = spawn('node', ['./bin/run', ...args])
    let stdout = ''
    let stderr = ''

    child.stdout.on('data', (data) => { stdout += data.toString() })
    child.stderr.on('data', (data) => { stderr += data.toString() })

    child.on('close', (exitCode) => {
      resolve({ stdout, stderr, exitCode: exitCode || 0 })
    })
  })
}

/**
 * Parse envelope from command output
 */
export function parseEnvelope<T = any>(stdout: string): SuccessEnvelope<T> | ErrorEnvelope {
  return JSON.parse(stdout)
}

/**
 * Assert envelope is success envelope
 */
export function assertSuccessEnvelope<T = any>(
  envelope: any
): asserts envelope is SuccessEnvelope<T> {
  expect(envelope).to.have.property('success', true)
  expect(envelope).to.have.property('data')
  expect(envelope).to.have.property('metadata')
  expect(envelope.metadata).to.have.all.keys(
    'timestamp',
    'command',
    'execution_time_ms',
    'version'
  )
}

/**
 * Assert envelope is error envelope
 */
export function assertErrorEnvelope(
  envelope: any
): asserts envelope is ErrorEnvelope {
  expect(envelope).to.have.property('success', false)
  expect(envelope).to.have.property('error')
  expect(envelope.error).to.have.all.keys(
    'code',
    'message',
    'details',
    'suggestions',
    'notionError'
  )
  expect(envelope).to.have.property('metadata')
}

/**
 * Mock Notion API response
 */
export function mockNotionAPI(mockResponse: any) {
  // Stub Notion client methods
  // Return mock response for testing
}
```

## Coverage Requirements

### Minimum Coverage Thresholds

| Component | Line Coverage | Branch Coverage | Function Coverage |
|-----------|---------------|-----------------|-------------------|
| envelope.ts | 95% | 90% | 100% |
| base-command.ts | 90% | 85% | 100% |
| Commands (migrated) | 80% | 75% | 90% |

### Coverage Reporting

```bash
# Generate coverage report
npm run test:coverage

# View HTML report
open coverage/index.html

# Check coverage thresholds
npm run test:coverage -- --check-coverage
```

## CI/CD Integration

### GitHub Actions Workflow

**File:** `.github/workflows/test.yml`

```yaml
name: Test

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest

    strategy:
      matrix:
        node-version: [18.x, 20.x]

    steps:
      - uses: actions/checkout@v3

      - name: Use Node.js ${{ matrix.node-version }}
        uses: actions/setup-node@v3
        with:
          node-version: ${{ matrix.node-version }}

      - name: Install dependencies
        run: npm ci

      - name: Run linter
        run: npm run lint

      - name: Run unit tests
        run: npm run test:unit

      - name: Run integration tests
        run: npm run test:integration
        env:
          NOTION_TOKEN: ${{ secrets.NOTION_TEST_TOKEN }}

      - name: Generate coverage
        run: npm run test:coverage

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage/lcov.info
          flags: unittests
          name: codecov-umbrella
          fail_ci_if_error: false
        env:
          CODECOV_TOKEN: ${{ secrets.CODECOV_TOKEN }}
```

## Manual Testing Checklist

### Pre-Release Testing

- [ ] **Success Envelope Validation**
  - [ ] Run `notion-cli page retrieve <id> --json`
  - [ ] Verify `success: true`
  - [ ] Verify all metadata fields present
  - [ ] Verify execution_time_ms is reasonable
  - [ ] Verify timestamp is ISO 8601 format
  - [ ] Verify version matches package.json

- [ ] **Error Envelope Validation**
  - [ ] Run command with invalid ID `--json`
  - [ ] Verify `success: false`
  - [ ] Verify error code is semantic
  - [ ] Verify suggestions are present and relevant
  - [ ] Verify exit code is 1 (API error)

- [ ] **Validation Error Testing**
  - [ ] Run command with missing required arg
  - [ ] Verify VALIDATION_ERROR code
  - [ ] Verify exit code is 2 (CLI error)
  - [ ] Verify suggestions mention command help

- [ ] **Output Format Testing**
  - [ ] Test `--json` (pretty, indented)
  - [ ] Test `--compact-json` (single line)
  - [ ] Test `--raw` (no envelope)
  - [ ] Verify mutually exclusive flags error

- [ ] **Stdout/Stderr Separation**
  - [ ] Run `command --json > out.json 2> err.log`
  - [ ] Verify out.json contains only JSON
  - [ ] Verify err.log contains diagnostics
  - [ ] Verify no pollution of stdout

- [ ] **Exit Code Testing**
  - [ ] Success: `echo $?` should be 0
  - [ ] API error: `echo $?` should be 1
  - [ ] CLI error: `echo $?` should be 2

- [ ] **Backward Compatibility**
  - [ ] Test existing scripts with `--raw`
  - [ ] Verify output is unchanged from pre-envelope
  - [ ] Test piping to jq with both formats

### Performance Testing

- [ ] **Envelope Overhead**
  - [ ] Measure execution time with/without envelope
  - [ ] Verify overhead is <10ms for typical responses
  - [ ] Test with large responses (1000+ results)
  - [ ] Verify memory usage is reasonable

- [ ] **Concurrent Requests**
  - [ ] Run multiple commands in parallel
  - [ ] Verify envelopes don't interfere
  - [ ] Verify execution times are independent

## Regression Testing

### After Each Migration

- [ ] Run full test suite: `npm test`
- [ ] Test migrated command with all output formats
- [ ] Verify exit codes for success and error cases
- [ ] Check coverage hasn't decreased
- [ ] Test automation scripts still work

### Before Release

- [ ] Full integration test suite passes
- [ ] E2E tests with real Notion API pass
- [ ] Manual testing checklist complete
- [ ] Coverage thresholds met
- [ ] No regressions in existing functionality

## Test Maintenance

### When Adding New Commands

1. Create integration test file
2. Test all output formats
3. Test error cases
4. Add to CI/CD pipeline
5. Update coverage thresholds if needed

### When Modifying Envelope Structure

1. Update envelope.ts
2. Update all related tests
3. Update snapshots
4. Update documentation
5. Verify backward compatibility

### When Adding Error Codes

1. Add to ErrorCode enum
2. Add suggestion generation logic
3. Add exit code mapping
4. Add unit tests
5. Add integration test

## Resources

- **Test Files:** `test/` directory
- **Mock Data:** `test/fixtures/`
- **Test Utilities:** `test/helpers/`
- **CI Configuration:** `.github/workflows/`
- **Coverage Reports:** `coverage/` (generated)

## Next Steps

1. Implement unit tests for EnvelopeFormatter
2. Implement unit tests for BaseCommand
3. Create integration tests for core commands
4. Set up CI/CD pipeline
5. Migrate commands incrementally with testing
6. Monitor coverage and maintain thresholds
