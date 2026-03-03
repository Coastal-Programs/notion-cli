---
paths:
  - "src/**/*.ts"
---

# Code Style Rules

## TypeScript Conventions

- Avoid `any` types; use proper typing with Notion SDK types
- Require return types on public functions
- Use optional chaining and nullish coalescing
- Files: kebab-case (`table-formatter.ts`)
- Classes: PascalCase (`DbQuery`)
- Functions: camelCase (`retrieveDatabase`)
- Constants: SCREAMING_SNAKE_CASE

## Formatting

- ESLint 9 flat config (eslint.config.mjs)
- 2-space indentation, single quotes, no semicolons (Prettier)
- Max line length: 100 characters
- Run `npm run lint` before committing

## Error Handling

- Use `NotionCLIError` with error codes from `NotionCLIErrorCode` enum
- Use `NotionCLIErrorFactory` static methods for common patterns:
  - `tokenMissing()`, `resourceNotFound()`, `invalidIdFormat()`
  - `rateLimited()`, `networkError()`, `invalidJson()`
- Always include suggestions array with fix commands or links
- Include context: resourceType, attemptedId, endpoint, statusCode

## API Client Patterns (src/notion.ts)

- All API calls go through `cachedFetch()` which handles caching + dedup + retry
- Cache TTLs are per-resource-type (blocks 30s, pages 1min, users 1hr, databases 10min)
- Use `BATCH_CONFIG` for concurrent operations (delete: 5, children: 10)
- Deduplication prevents duplicate in-flight requests automatically

## Documentation

- JSDoc with `@param`, `@returns`, `@example` for public APIs
- Inline comments only for non-obvious logic
- No debug code or console.log in production
