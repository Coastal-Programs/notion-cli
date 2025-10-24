# Output Format Options

The Notion CLI now supports multiple output formats to suit different use cases. This makes it easier to integrate with other tools, scripts, and workflows.

## Available Output Formats

### 1. Default Table (oclif table)
The standard table output with optional CSV export.

```bash
# Default table
notion-cli db query DATABASE_ID

# CSV output
notion-cli db query DATABASE_ID --csv
```

### 2. Markdown Table (`--markdown` or `-m`)
GitHub-flavored markdown tables, perfect for documentation and README files.

```bash
notion-cli db query DATABASE_ID --markdown
notion-cli page retrieve PAGE_ID --markdown
notion-cli search -q "My Page" --markdown
```

**Output Example:**
```
| title | object | id | url |
| --- | --- | --- | --- |
| My Page | page | abc123 | https://notion.so/... |
| Another Page | page | def456 | https://notion.so/... |
```

**Features:**
- Escapes pipe characters in content
- Handles null/undefined values gracefully
- Removes newlines to keep table formatting intact

### 3. Compact JSON (`--compact-json` or `-c`)
Single-line JSON output ideal for piping to other tools like `jq`, logging, or streaming.

```bash
notion-cli db query DATABASE_ID --compact-json
notion-cli db retrieve DATABASE_ID --compact-json
```

**Output Example:**
```json
[{"object":"page","id":"abc123","title":"My Page"},{"object":"page","id":"def456","title":"Another"}]
```

**Use Cases:**
- Piping to `jq` for JSON processing
- Logging to files (one entry per line)
- Integration with log aggregators
- Streaming data to other services

**Example with jq:**
```bash
notion-cli db query DATABASE_ID --compact-json | jq '.[] | select(.object == "page")'
```

### 4. Pretty Table (`--pretty` or `-P`)
Enhanced table with Unicode box-drawing characters for better visual clarity.

```bash
notion-cli db query DATABASE_ID --pretty
notion-cli search -q "My Page" --pretty
```

**Output Example:**
```
┌─────────────┬────────┬─────────┬─────────────────────┐
│ title       │ object │ id      │ url                 │
├─────────────┼────────┼─────────┼─────────────────────┤
│ My Page     │ page   │ abc123  │ https://notion.so/  │
│ Another Pge │ page   │ def456  │ https://notion.so/  │
└─────────────┴────────┴─────────┴─────────────────────┘
```

**Features:**
- Beautiful box-drawing borders
- Auto-sized columns based on content
- Enhanced readability for terminal output

### 5. Raw JSON (`--raw` or `-r`)
Pretty-printed JSON (2-space indentation) - the legacy format.

```bash
notion-cli db query DATABASE_ID --raw
notion-cli page retrieve PAGE_ID --raw
```

**Output Example:**
```json
[
  {
    "object": "page",
    "id": "abc123",
    "title": "My Page"
  }
]
```

### 6. Automation JSON (`--json` or `-j`)
Structured JSON with success/error metadata, timestamps, and consistent schema.

```bash
notion-cli db query DATABASE_ID --json
```

**Output Example:**
```json
{
  "success": true,
  "data": [...],
  "count": 2,
  "timestamp": "2025-10-22T10:30:00.000Z"
}
```

## Command Support

All major commands now support the new output formats:

### Database Commands
```bash
# Query database
notion-cli db query DATABASE_ID --markdown
notion-cli db query DATABASE_ID --compact-json
notion-cli db query DATABASE_ID --pretty

# Retrieve database
notion-cli db retrieve DATABASE_ID --markdown
notion-cli db retrieve DATABASE_ID --compact-json
notion-cli db retrieve DATABASE_ID --pretty
```

### Page Commands
```bash
# Retrieve page metadata
notion-cli page retrieve PAGE_ID --compact-json
notion-cli page retrieve PAGE_ID --pretty

# Note: -m/--markdown outputs page CONTENT, not metadata table
notion-cli page retrieve PAGE_ID -m  # Uses NotionToMarkdown
```

### Search Command
```bash
notion-cli search -q "keyword" --markdown
notion-cli search -q "keyword" --compact-json
notion-cli search -q "keyword" --pretty
```

## Mutually Exclusive Flags

The new output format flags are mutually exclusive. You cannot combine them:

```bash
# ❌ This will fail
notion-cli db query ID --markdown --compact-json

# ✅ Use one at a time
notion-cli db query ID --markdown
notion-cli db query ID --compact-json
```

## Integration Examples

### Example 1: Export to Markdown Documentation
```bash
#!/bin/bash
echo "# Database Contents" > README.md
echo "" >> README.md
notion-cli db query $DATABASE_ID --markdown >> README.md
```

### Example 2: Process with jq
```bash
# Get all page IDs
notion-cli db query $DATABASE_ID --compact-json | jq -r '.[].id'

# Filter and transform
notion-cli search -q "Project" --compact-json | jq '[.[] | {title, url}]'
```

### Example 3: Log to File
```bash
# One JSON object per line for easy parsing
notion-cli db query $DATABASE_ID --compact-json >> database.log
```

### Example 4: Pretty Display
```bash
# For human-readable terminal output
notion-cli search -q "Meeting" --pretty | less
```

## Implementation Details

### TypeScript Types
All output functions are properly typed in `C:\Users\jakes\Developer\GitHub\notion-cli\src\helper.ts`:

```typescript
export const outputCompactJson = (res: any) => void
export const outputMarkdownTable = (data: any[], columns: Record<string, any>) => void
export const outputPrettyTable = (data: any[], columns: Record<string, any>) => void
```

### Flag Definitions
Flags are defined in `C:\Users\jakes\Developer\GitHub\notion-cli\src\base-flags.ts`:

```typescript
export const OutputFormatFlags = {
  markdown: Flags.boolean({ char: 'm', exclusive: ['compact-json', 'pretty'] }),
  'compact-json': Flags.boolean({ char: 'c', exclusive: ['markdown', 'pretty'] }),
  pretty: Flags.boolean({ char: 'P', exclusive: ['markdown', 'compact-json'] }),
}
```

## Backward Compatibility

All existing flags and behavior are preserved:
- `--raw` still works for legacy JSON output
- `--csv` still works for CSV export
- `--json` still works for automation JSON
- Default table output unchanged when no flags are provided

## Performance Considerations

- **Compact JSON**: Fastest, minimal string operations
- **Markdown Table**: Fast, simple string concatenation
- **Pretty Table**: Slightly slower due to width calculations
- **Raw/Automation JSON**: Uses built-in JSON.stringify

All formats are suitable for production use with typical Notion dataset sizes.
