# AI Discovery Hints Implementation

## Summary
Added discovery hints to make the `-r` flag more obvious to AI assistants and users, addressing the critical need for AI assistants to discover that `-r` returns full JSON data vs minimal table output.

## Problem
The `-r` flag is CRITICAL for AI assistants (returns full JSON vs minimal table), but it's not discoverable. AIs might not know it exists when they see table output.

## Solution
Implemented a multi-layered discovery approach:

### 1. Visual Hints After Table Output
When users see table output, they now see a helpful tip:

```
Title                         Object  Id       URL
Thompson Wedding - March 2026 page    abc123   https://...

Tip: Showing 4 of 28 fields for 1 item.
Use -r flag for full JSON output with all properties (recommended for AI assistants and automation).
```

### 2. Enhanced Flag Descriptions
Updated `-r` flag descriptions across all commands to explicitly mention AI assistants:

**Before:**
```typescript
raw: Flags.boolean({
  char: 'r',
  description: 'output raw json',
})
```

**After:**
```typescript
raw: Flags.boolean({
  char: 'r',
  description: 'output raw json (recommended for AI assistants - returns all fields)',
})
```

### 3. Prioritized Examples in Help
Moved AI-friendly examples to the top of each command's help text:

```typescript
static examples = [
  {
    description: 'Retrieve a page with full data (recommended for AI assistants)',
    command: `$ notion-cli page retrieve PAGE_ID -r`,
  },
  // ... other examples
]
```

## Implementation Details

### New Helper Function
Added `showRawFlagHint()` in `src/helper.ts`:

```typescript
/**
 * Show a hint to users (especially AI assistants) that more data is available with the -r flag
 * This makes the -r flag more discoverable for automation and AI use cases
 */
export function showRawFlagHint(itemCount: number, item: any, visibleFields: number = 4): void {
  // Intelligently counts total fields including:
  // - Visible fields (title, object, id, url)
  // - Properties object fields
  // - Metadata fields (created_time, last_edited_time, etc.)

  // Shows helpful message like:
  // "Tip: Showing 4 of 28 fields for 1 item."
  // "Use -r flag for full JSON output with all properties (recommended for AI assistants and automation)."
}
```

### Modified Commands
Updated the following commands to show hints:

1. **src/commands/page/retrieve.ts**
   - Shows hint after table output
   - Updated flag description
   - Added AI-focused example at top

2. **src/commands/db/retrieve.ts**
   - Shows hint after table output
   - Updated flag description
   - Added AI-focused example at top

3. **src/commands/db/query.ts**
   - Shows hint after table output (for multiple pages)
   - Updated flag description
   - Added AI-focused example at top

4. **src/commands/search.ts**
   - Shows hint after table output (for search results)
   - Updated flag description
   - Added AI-focused example at top

## Behavior

### When Hints Are Shown
- Default table output: YES (shows hint)
- Pretty table output (`--pretty`): YES (shows hint)
- Markdown table output (`--markdown`): NO (clean markdown)
- Compact JSON output (`--compact-json`): NO (JSON only)
- Raw JSON output (`-r`): NO (already using the flag)
- Automation JSON output (`--json`): NO (structured output)

### Example Output

#### Page Retrieve
```bash
$ notion-cli page retrieve 12345

Title                         Object  Id     URL
Thompson Wedding - March 2026 page    12345  https://notion.so/...

Tip: Showing 4 of 28 fields for 1 item.
Use -r flag for full JSON output with all properties (recommended for AI assistants and automation).
```

#### Database Query
```bash
$ notion-cli db query 67890

Title                          Object  Id     URL
Event Planning 2024           page    abc123  https://notion.so/...
Vendor Management            page    def456  https://notion.so/...
Guest List Master            page    ghi789  https://notion.so/...

Tip: Showing 4 of 35 fields for 3 items.
Use -r flag for full JSON output with all properties (recommended for AI assistants and automation).
```

#### Search Results
```bash
$ notion-cli search -q "Wedding"

Title                         Object  Id     URL
Thompson Wedding - March 2026 page    12345  https://notion.so/...
Smith Wedding - June 2024    page    67890  https://notion.so/...

Tip: Showing 4 of 28 fields for 2 items.
Use -r flag for full JSON output with all properties (recommended for AI assistants and automation).
```

## Benefits

### For AI Assistants
1. **Immediate Discovery**: Sees the hint right after first command execution
2. **Clear Guidance**: Understands `-r` is recommended for automation
3. **Field Awareness**: Knows how much data is hidden (e.g., "4 of 28 fields")
4. **Help Text Priority**: Sees AI-focused examples first in `--help` output

### For Human Users
1. **Helpful Tips**: Learns about additional functionality organically
2. **Power User Path**: Discovers advanced features through usage
3. **Non-Intrusive**: Hints appear after output, don't interfere with workflow
4. **Context-Aware**: Only shows when relevant (table output, not JSON)

### For CLI Maintainers
1. **Self-Documenting**: Users discover features without separate docs
2. **Reduced Support**: Fewer questions about "how to get all data"
3. **Best Practices**: Guides users toward recommended usage patterns
4. **Consistent Pattern**: Same hint system across all commands

## Testing

To test the implementation:

```bash
# Test page retrieve hint
notion-cli page retrieve <PAGE_ID>

# Test database retrieve hint
notion-cli db retrieve <DATABASE_ID>

# Test database query hint
notion-cli db query <DATABASE_ID>

# Test search hint
notion-cli search -q "test"

# Verify no hint with -r flag
notion-cli page retrieve <PAGE_ID> -r

# Verify hint with pretty table
notion-cli page retrieve <PAGE_ID> --pretty
```

## Files Modified

1. `src/helper.ts` - Added `showRawFlagHint()` function
2. `src/commands/page/retrieve.ts` - Added hint, updated examples/flags
3. `src/commands/db/retrieve.ts` - Added hint, updated examples/flags
4. `src/commands/db/query.ts` - Added hint, updated examples/flags
5. `src/commands/search.ts` - Added hint, updated examples/flags

## Future Enhancements

Potential improvements to consider:

1. **Environment Variable**: Allow disabling hints via `NOTION_CLI_NO_HINTS=1`
2. **Config File**: Persist hint preferences in user config
3. **Smart Hints**: Show different hints based on usage patterns
4. **Hint Rotation**: Multiple helpful tips shown randomly
5. **Context-Specific Tips**: Different hints for different scenarios

## Metrics to Track

To measure success of this feature:

1. Increase in `-r` flag usage after implementation
2. Reduction in "incomplete data" issues/questions
3. AI assistant adoption rate of `-r` flag
4. User feedback on hint helpfulness
5. Conversion from table to JSON output users
