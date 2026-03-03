---
paths:
  - "test/**/*.ts"
  - "src/**/*.ts"
---

# Testing Rules

## Coverage Requirements

- Minimum: 90% line coverage for all modules
- Target: 95%+ for utilities (`src/utils/`)
- Required 100%: auth, API calls, data formatting, table-formatter

## Framework

- Mocha + Chai (expect style) + Sinon (stubs/spies) + Nock (HTTP mocking)
- Config: `.mocharc.json` with 5s timeout, `forbid-only: true`
- Coverage: NYC with lcov + text reporters

## Test Structure (AAA Pattern)

```typescript
describe('module-name', () => {
  describe('function or behavior group', () => {
    it('should do specific thing', () => {
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

## Test Isolation

- Disable real network: `nock.disableNetConnect()` in setup.ts
- Clear cache between tests: `cacheManager.clear()`
- Save/restore env vars in beforeEach/afterEach
- Stub `process.exit` to prevent test runner termination
- Clean nock interceptors: `nock.cleanAll()` in afterEach

## Running Tests

```bash
npm test                                          # All tests
npm test -- test/utils/table-formatter.test.ts    # Specific file
npm test -- --grep "table-formatter"              # Pattern match
npx nyc --reporter=text npm test                  # With coverage
```

## Key Conventions

- Test files mirror src/ structure: `src/utils/foo.ts` -> `test/utils/foo.test.ts`
- Use `@oclif/test` helper for command integration tests
- Mock Notion API with nock against `https://api.notion.com`
- Use `test/helpers/notion-stubs.ts` for mock data factories
- Use compiled JS coverage numbers (dist/*.js), not TS numbers
