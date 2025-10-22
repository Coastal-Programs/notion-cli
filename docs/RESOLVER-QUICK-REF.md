# Name Resolver - Quick Reference

## For Developers

### Using the Resolver in Commands

```typescript
import { resolveNotionId } from '../../utils/notion-resolver'

// In your command's run() method:
const databaseId = await resolveNotionId(args.database_id, 'database')
const pageId = await resolveNotionId(args.page_id, 'page')
```

### Supported Input Formats

```typescript
// All of these work:
await resolveNotionId('https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00', 'database')
await resolveNotionId('notion.so/1fb79d4c71bb8032b722c82305b63a00', 'database')
await resolveNotionId('1fb79d4c71bb8032b722c82305b63a00', 'database')
await resolveNotionId('1fb79d4c-71bb-8032-b722-c82305b63a00', 'database')
// Coming soon: await resolveNotionId('Tasks Database', 'database')
```

### Error Handling

The resolver throws `NotionCLIError` exceptions:

```typescript
try {
  const id = await resolveNotionId(userInput, 'database')
  // Use the ID...
} catch (error) {
  if (error instanceof NotionCLIError) {
    // Handle CLI error (already formatted for user)
    if (flags.json) {
      this.log(JSON.stringify(error.toJSON(), null, 2))
    } else {
      this.error(error.message)
    }
  }
}
```

### Type Parameter

Always specify the resource type for better error messages:

```typescript
// For databases/data sources
await resolveNotionId(input, 'database')

// For pages/blocks
await resolveNotionId(input, 'page')
```

## For Users

### Using the CLI

All commands now accept URLs, IDs, or database names:

```bash
# Using a full URL
notion-cli db retrieve https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00

# Using just the ID
notion-cli db retrieve 1fb79d4c71bb8032b722c82305b63a00

# Using the ID with dashes
notion-cli db retrieve 1fb79d4c-71bb-8032-b722-c82305b63a00

# Coming soon: Using the database name
# notion-cli db retrieve "Tasks Database"
```

### Commands Updated

All these commands now support URL/ID/name inputs:

**Database Commands**:
- `db retrieve <database_id>`
- `db update <database_id>`
- `db create <page_id>` (parent page)
- `db schema <data_source_id>`
- `db query <database_id>`

**Page Commands**:
- `page retrieve <page_id>`
- `page update <page_id>`
- `page create -p <parent_page_id>` or `-d <parent_data_source_id>`

**Block Commands**:
- `block append -b <block_id>`
- `block update <block_id>`

### Error Messages

If you provide a name but the cache isn't set up yet:

```
Database "Tasks" not found.

Name-based lookups are coming soon!

For now, please use:
  1. The full Notion URL
  2. The database ID directly

Example URL: https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00
Example ID: 1fb79d4c71bb8032b722c82305b63a00
```

## Implementation Details

### Resolution Stages

1. **URL Extraction**: If input contains "notion.so", extract the ID from the URL
2. **Direct ID Validation**: If input is 32 hex characters (with optional dashes), use it directly
3. **Cache Lookup** (coming soon): Search local cache for database name
4. **API Search** (coming soon): Query Notion API as fallback

### Architecture

```
User Input → resolveNotionId() → Clean Notion ID
                 ↓
    ┌────────────┴────────────┐
    │                         │
  URL?                      ID?
    ↓                         ↓
extractNotionId()      isValidNotionId()
    │                         │
    └────────────┬────────────┘
                 ↓
          Clean 32-char ID
```

### Future: Cache Integration

When the cache is implemented:

```
User Input → resolveNotionId()
                 ↓
    ┌────────────┴────────────┐
    │                         │
  URL/ID?                  Name?
    ↓                         ↓
Return ID            searchCache()
                          ↓
                    Cache Hit? → Return ID
                          ↓
                    Cache Miss → searchNotionApi()
                          ↓
                    API Hit? → Return ID
                          ↓
                    Not Found → Error
```

## Migration Guide

### Updating Existing Commands

Replace:
```typescript
import { extractNotionId } from '../../utils/notion-url-parser'

const id = extractNotionId(args.database_id)
```

With:
```typescript
import { resolveNotionId } from '../../utils/notion-resolver'

const id = await resolveNotionId(args.database_id, 'database')
```

**Important**: Note the `await` keyword - the function is now async!

### Updating Examples

Add URL examples to your command documentation:

```typescript
static examples = [
  {
    description: 'Retrieve a database via ID',
    command: 'notion-cli db retrieve DATABASE_ID',
  },
  {
    description: 'Retrieve a database via URL',
    command: 'notion-cli db retrieve https://notion.so/DATABASE_ID',
  },
  // ... more examples
]
```

## Troubleshooting

### "Invalid input: expected a database name, ID, or URL"

**Cause**: Input is null, undefined, or not a string

**Solution**: Ensure you're passing a valid string to the resolver

### "Invalid Notion URL: ..."

**Cause**: URL format is incorrect or doesn't contain a valid Notion ID

**Solution**: Check that the URL follows the pattern `notion.so/{id}`

### "Invalid database ID format: ..."

**Cause**: Input appears to be an ID but doesn't have 32 hexadecimal characters

**Solution**: Verify the ID is correct. Notion IDs are always 32 hex characters.

### "Database '...' not found"

**Cause**: Input looks like a name but cache isn't implemented yet

**Solution**: Use the full URL or ID instead, or wait for cache feature

## Best Practices

1. **Always use the resolver** instead of `extractNotionId()` in new code
2. **Specify the correct type** ('database' or 'page') for better error messages
3. **Handle errors gracefully** with try-catch blocks
4. **Add URL examples** to all command documentation
5. **Test with both URLs and IDs** to ensure both paths work

## Related Files

- `src/utils/notion-resolver.ts` - Main resolver implementation
- `src/utils/notion-url-parser.ts` - URL parsing utilities (used by resolver)
- `src/errors.ts` - Error types and handling
- `docs/CACHING-ARCHITECTURE.md` - Full architecture documentation
- `docs/RESOLVER-IMPLEMENTATION.md` - Implementation details

## Future Enhancements

Coming in the next phase:

- Cache-based name lookups
- Alias support
- Fuzzy matching
- Auto-sync on cache miss
- Interactive database selection

---

*Last updated: 2025-10-22*
