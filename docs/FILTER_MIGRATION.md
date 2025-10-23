# Filter Flag Migration Guide

## Summary of Changes

The `db query` command has been simplified to provide a clearer, more intuitive filtering interface for AI agents and automation scripts.

## New Filter Flags

### 1. `--filter` (Primary Method)
- **Description**: JSON filter object following Notion's filter API format
- **Shorthand**: `-f`
- **Use case**: Programmatic filtering by AI agents
- **Example**:
  ```bash
  notion-cli db query <ID> --filter '{"property": "Status", "select": {"equals": "Done"}}' --json
  ```

### 2. `--search` (Human Convenience)
- **Description**: Simple text search across common properties (Name, Title, Description)
- **Shorthand**: `-s`
- **Use case**: Quick human searches
- **Example**:
  ```bash
  notion-cli db query <ID> --search "urgent" --json
  ```

### 3. `--file-filter` (Complex Queries)
- **Description**: Load filter from JSON file
- **Shorthand**: `-F`
- **Use case**: Reusable complex filters
- **Example**:
  ```bash
  notion-cli db query <ID> --file-filter ./filter.json --json
  ```

## Deprecated Flags

The following flags are deprecated but still functional with warnings:

- `--rawFilter` (use `--filter` instead)
- `--fileFilter` (use `--file-filter` instead)

These will be removed in v6.0.0.

## Migration Examples

### Old (Deprecated)
```bash
notion-cli db query <ID> --rawFilter '{"property": "Status", "select": {"equals": "Done"}}' --json
notion-cli db query <ID> --fileFilter ./filter.json --json
```

### New (Recommended)
```bash
notion-cli db query <ID> --filter '{"property": "Status", "select": {"equals": "Done"}}' --json
notion-cli db query <ID> --file-filter ./filter.json --json
```

## Backward Compatibility

All existing scripts using `--rawFilter` or `--fileFilter` will continue to work but will show deprecation warnings to stderr:

```
⚠️  Warning: --rawFilter is deprecated and will be removed in v6.0.0
   Use --filter instead: notion-cli db query DS_ID --filter '...'
```

These warnings do not pollute stdout JSON output, so automated scripts will continue to function correctly.

## Documentation

- **Full Filter Guide**: [docs/FILTER_GUIDE.md](./docs/FILTER_GUIDE.md)
- **Filter Examples**: [examples/filters/](./examples/filters/)
- **Updated README**: [README.md](./README.md#database-query-filtering)

## Files Changed

1. **src/commands/db/query.ts**
   - Added `--filter`, `--file-filter`, `--search` flags
   - Deprecated `--rawFilter`, `--fileFilter` with warnings
   - Improved error messages
   - Enhanced examples

2. **docs/FILTER_GUIDE.md** (NEW)
   - Comprehensive filter documentation
   - Examples for all property types
   - Common use cases
   - Troubleshooting guide

3. **README.md**
   - Added "Database Query Filtering" section
   - Quick reference examples
   - Links to full documentation

4. **examples/filters/** (NEW)
   - active-tasks.json
   - high-priority-tasks.json
   - overdue-items.json
   - needs-review.json
   - README.md

## Testing Checklist

- [x] Build succeeds without errors
- [x] `--filter` flag accepts JSON and parses correctly
- [x] `--search` flag generates correct OR filter
- [x] `--file-filter` flag loads JSON from file
- [x] Deprecated flags show warnings to stderr
- [x] Deprecated flags still function correctly
- [x] Flags are mutually exclusive
- [x] Help text shows correct descriptions
- [x] Examples compile and display correctly

## Benefits

1. **Clearer naming**: `--filter` is more intuitive than `--rawFilter`
2. **Better DX**: Short flags `-f`, `-F`, `-s` for common operations
3. **AI-friendly**: Primary method (`--filter`) is optimized for programmatic use
4. **Human-friendly**: `--search` flag for quick text searches
5. **Reusability**: `--file-filter` encourages filter organization
6. **Documentation**: Comprehensive guide with examples
7. **Backward compatible**: No breaking changes for existing users

## Next Steps

1. Monitor usage of deprecated flags
2. Collect feedback on new interface
3. Plan removal of deprecated flags in v6.0.0
4. Consider adding filter templates/presets
