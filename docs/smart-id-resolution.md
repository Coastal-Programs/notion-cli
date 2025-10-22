# Smart ID Resolution

## Overview

The Notion CLI now includes smart ID resolution that automatically handles the common confusion between `database_id` and `data_source_id` values.

## The Problem

When working with Notion's API v5, users often encounter two different ID types for databases:

1. **`data_source_id`** - The correct ID for database operations
   - Used with: `db retrieve`, `db query`, `db update`
   - Found in: Database objects returned by the API
   - Example: `2gc80e5d82cc9043c833d93416c74b11`

2. **`database_id`** - A reference ID found in page parents
   - Found in: `page.parent.database_id` field
   - Cannot be used directly for database operations
   - Example: `1fb79d4c71bb8032b722c82305b63a00`

### Why This Causes Confusion

When users retrieve a page and see `parent.database_id`, they naturally assume they can use that ID with database commands:

```bash
# Get a page
notion-cli page retrieve abc123 --raw

# Output shows:
{
  "parent": {
    "type": "database_id",
    "database_id": "1fb79d4c71bb8032b722c82305b63a00"
  }
}

# User tries to use that ID (would fail before smart resolution)
notion-cli db retrieve 1fb79d4c71bb8032b722c82305b63a00
# Error: object_not_found
```

## The Solution: Smart ID Resolution

The CLI now automatically detects when you provide a `database_id` and converts it to the correct `data_source_id`:

```bash
# Now this works automatically!
notion-cli db retrieve 1fb79d4c71bb8032b722c82305b63a00

# Output:
Info: Resolved database_id to data_source_id
  database_id:    1fb79d4c71bb8032b722c82305b63a00
  data_source_id: 2gc80e5d82cc9043c833d93416c74b11

Note: Use data_source_id for database operations.
      The database_id from parent.database_id won't work directly.

[Database information displayed...]
```

## How It Works

The smart resolution system follows this process:

1. **Try Direct Lookup**: First, try to use the provided ID as a `data_source_id`
2. **Detect Failure**: If the lookup fails with `object_not_found` error
3. **Search for Pages**: Search for pages that have this ID as their `parent.database_id`
4. **Extract data_source_id**: Get the `data_source_id` from the matching page's parent
5. **Use Correct ID**: Retry the operation with the correct `data_source_id`
6. **Inform User**: Show a helpful message explaining the conversion

## Benefits

### For Users
- **No More Confusion**: Use either ID type - it just works
- **Better Error Messages**: Clear explanation when conversion happens
- **Educational**: Learn the difference between ID types
- **Time Saving**: No need to manually find the correct ID

### For Developers
- **Reduced Support**: Fewer "ID not found" issues
- **Better UX**: Intuitive behavior matches user expectations
- **Backward Compatible**: Existing scripts continue to work
- **Transparent**: Clear logging shows what's happening

## Examples

### Basic Usage

```bash
# Any of these work now:
notion-cli db retrieve 1fb79d4c71bb8032b722c82305b63a00  # database_id (auto-converts)
notion-cli db retrieve 2gc80e5d82cc9043c833d93416c74b11  # data_source_id (direct)
notion-cli db retrieve "Tasks Database"                   # name (cache/search)
```

### Database Query

```bash
# Works with database_id
notion-cli db query 1fb79d4c71bb8032b722c82305b63a00 --filter status equals Done
```

### Database Update

```bash
# Works with database_id
notion-cli db update 1fb79d4c71bb8032b722c82305b63a00 --title "Updated Title"
```

### Workflow Example

```bash
# 1. Get a page ID from somewhere
PAGE_ID="abc123..."

# 2. Retrieve the page to see its parent database
notion-cli page retrieve $PAGE_ID --raw | jq -r '.parent.database_id'
# Output: 1fb79d4c71bb8032b722c82305b63a00

# 3. Use that database_id directly (auto-converts!)
notion-cli db query 1fb79d4c71bb8032b722c82305b63a00 --json > results.json

# 4. The system shows the conversion happened
# Info: Resolved database_id to data_source_id
#   database_id:    1fb79d4c71bb8032b722c82305b63a00
#   data_source_id: 2gc80e5d82cc9043c833d93416c74b11
```

## Implementation Details

### Resolution Algorithm

The smart resolution is implemented in `src/utils/notion-resolver.ts`:

```typescript
async function trySmartDatabaseResolution(databaseId: string): Promise<string> {
  try {
    // Try direct lookup with data_source_id
    await retrieveDataSource(databaseId)
    return databaseId  // Success - it's a valid data_source_id
  } catch (error) {
    if (isNotFoundError(error)) {
      // Try to resolve database_id → data_source_id
      const dataSourceId = await resolveDatabaseIdToDataSourceId(databaseId)
      if (dataSourceId) {
        console.log("Info: Resolved database_id to data_source_id")
        return dataSourceId
      }
    }
    throw error
  }
}
```

### Performance Considerations

- **Fast Path**: Valid `data_source_id` values work immediately (no extra API calls)
- **Fallback Path**: Invalid IDs trigger one additional API search (max 100 pages)
- **Caching**: Resolution results are cached to avoid repeated lookups
- **Timeout**: Search has reasonable timeout to avoid hanging

### Limitations

1. **Requires Pages**: The database must have at least one page to resolve
   - Empty databases cannot be resolved this way
   - Workaround: Add a test page to the database

2. **Search Scope**: Searches first 100 pages
   - Usually sufficient since we only need one match
   - If no match in first 100, resolution fails

3. **API Permissions**: Requires search permission
   - Integration must have access to search workspace
   - Standard integration permissions include this

## Troubleshooting

### "Database not found" after resolution attempt

**Cause**: The database has no pages, or they're not accessible to your integration.

**Solutions**:
1. Verify the integration has access to the database
2. Add at least one page to the database
3. Use the correct `data_source_id` directly (from database list)

### Resolution is slow

**Cause**: Searching through many pages to find a match.

**Solutions**:
1. Use `data_source_id` directly for better performance
2. Ensure your integration has proper permissions
3. Check network connectivity

### Wrong database returned

**Cause**: Multiple databases share the same `database_id` (unlikely).

**Solutions**:
1. Use `data_source_id` for precise targeting
2. Verify the database ID is correct
3. Check that you're using the right workspace

## Best Practices

### For Scripts and Automation

```bash
# For reliability, prefer data_source_id when possible
DATA_SOURCE_ID=$(notion-cli db list --json | jq -r '.[] | select(.title == "Tasks") | .id')
notion-cli db query "$DATA_SOURCE_ID"
```

### For Interactive Use

```bash
# Use whichever ID you have - smart resolution handles it
notion-cli db retrieve <ANY_ID>
```

### For Learning

When the conversion message appears, take note of both IDs:
- **Save the `data_source_id`** for faster future operations
- **Understand** when you're using `database_id` vs `data_source_id`
- **Update scripts** to use the correct ID type

## Related Features

- **Name-based lookup**: Still works! Use database names instead of IDs
- **URL parsing**: Accepts full Notion URLs for any ID type
- **Cache system**: Frequently-used IDs are cached for speed
- **Error messages**: Helpful suggestions when IDs can't be resolved

## Technical Notes

### API v5 Changes

In Notion API v5, databases became "data sources":
- Old API: Used `database_id` everywhere
- New API: Databases → Data Sources with `data_source_id`
- Legacy field: `parent.database_id` still exists for compatibility

### Type Definitions

```typescript
// Page parent can reference a database
interface PageParent {
  type: 'database_id'
  database_id: string      // Cannot use for database operations
  data_source_id?: string  // Use this for database operations
}

// Database operations require data_source_id
interface DataSourceQuery {
  data_source_id: string  // This is what you need!
}
```

## Feedback and Support

If smart ID resolution doesn't work for your use case:

1. **Enable Debug Mode**: `export DEBUG=notion-cli:*`
2. **Check the Logs**: Look for "Debug: Failed to resolve database_id"
3. **Verify IDs**: Use `--raw` flag to see actual ID values
4. **Report Issues**: Include debug output in issue reports

## Future Enhancements

Potential improvements to smart resolution:

- [ ] Cache successful resolutions across sessions
- [ ] Resolve IDs for empty databases via direct API lookup
- [ ] Support bulk ID resolution for better performance
- [ ] Add `--no-smart-resolve` flag to disable feature
- [ ] Extend to other resource types (blocks, pages)

## Changelog

### v5.4.0 (Current)
- Initial implementation of smart ID resolution
- Automatic database_id → data_source_id conversion
- Helpful user messaging
- Full test coverage

---

**Related Documentation:**
- [Database Commands](./db.md)
- [Page Commands](./page.md)
- [ID Resolution System](./notion-resolver.md)
