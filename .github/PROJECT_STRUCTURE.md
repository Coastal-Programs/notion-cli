# Project Structure Guide

## Root Directory - User-Facing Documentation Only

The root directory contains **only** essential user-facing documentation:

### **Core Documentation (6 files)**

1. **README.md** - Primary project documentation
   - Quick start for AI agents
   - Feature showcase
   - Installation instructions
   - Links to all other docs

2. **CHANGELOG.md** - Release history
   - Standard location per Keep a Changelog
   - Complete version history
   - Breaking changes and migrations

3. **AI_AGENT_QUICK_REFERENCE.md** - Quick reference card
   - Linked from README
   - Fast lookup for AI agents
   - Property types cheat sheet
   - Common patterns

4. **ENHANCEMENTS.md** - Enhanced features deep dive
   - Linked from README
   - Caching architecture details
   - Retry logic explanation
   - Performance characteristics

5. **OUTPUT_FORMATS.md** - Output formats guide
   - Linked from README
   - All format options explained
   - Examples for each format

6. **CLAUDE.md** - Claude Code instructions
   - Standard .claude project file
   - Development guidelines
   - Architecture overview
   - Common patterns

---

## docs/ Directory - Detailed Documentation

All detailed, internal, and process documentation lives here:

### **User Guides**
- `SIMPLE_PROPERTIES.md` - Simple properties feature guide
- `FILTER_GUIDE.md` - Database query filter syntax
- `VERBOSE_LOGGING.md` - Debug logging guide
- `AI-AGENT-COOKBOOK.md` - Practical automation recipes

### **Technical Documentation**
- `ENVELOPE_*.md` (7 files) - JSON envelope system docs
- `CACHING-*.md` - Caching architecture docs
- `ERROR-HANDLING-*.md` - Error handling system
- `notion-api-*.md` - Notion API reference docs

### **Implementation Reports**
- `DOCUMENTATION-SYNC-2025-10-23.md` - Documentation alignment report
- `IMPLEMENTATION_SUMMARY.md` - Simple properties implementation
- `BUILD-VERIFICATION-REPORT.md` - Build verification notes
- `FILTER_MIGRATION.md` - Filter syntax migration guide

### **Research & Planning**
- `research-*.md` - Research documents
- `IMPLEMENTATION-PLAN.md` - Feature planning
- `TESTING-*.md` - Testing guides and summaries

---

## Rationale

### Why This Structure?

**Root = What Users Need First**
- New users see only essential docs
- No confusion from internal reports
- Quick access to most-used references
- Standard locations (README, CHANGELOG, CLAUDE.md)

**docs/ = Everything Else**
- Detailed technical docs
- Implementation reports
- Research documents
- Internal process notes

### Benefits

1. **Clean Root Directory**
   - 6 essential docs vs 9+ cluttered files
   - Easy to find what you need
   - Professional appearance

2. **Discoverable Documentation**
   - README links to everything important
   - Logical organization by topic
   - Clear naming conventions

3. **Separation of Concerns**
   - User-facing vs internal docs
   - Current vs historical reports
   - Guides vs implementation notes

---

## File Organization Rules

### ✅ Keep in Root If:
- Linked directly from README
- Part of standard project structure (README, CHANGELOG, LICENSE)
- Quick reference for frequent use
- Claude Code instructions (CLAUDE.md)

### ❌ Move to docs/ If:
- Implementation report or summary
- Internal process documentation
- Historical/archived information
- Deep technical dive
- Research or planning documents
- Migration guides (unless actively being used)

---

## Examples

### Good Root Structure ✅
```
notion-cli/
├── README.md                          # Primary docs
├── CHANGELOG.md                       # Release history
├── AI_AGENT_QUICK_REFERENCE.md       # Quick ref (linked from README)
├── ENHANCEMENTS.md                    # Features (linked from README)
├── OUTPUT_FORMATS.md                  # Guide (linked from README)
├── CLAUDE.md                          # Claude Code instructions
├── package.json
└── docs/
    ├── SIMPLE_PROPERTIES.md           # Detailed guide
    ├── IMPLEMENTATION_SUMMARY.md      # Internal report
    └── DOCUMENTATION-SYNC-*.md        # Process docs
```

### Bad Root Structure ❌
```
notion-cli/
├── README.md
├── CHANGELOG.md
├── IMPLEMENTATION_SUMMARY.md          # ❌ Should be in docs/
├── DOCUMENTATION-SYNC-*.md            # ❌ Should be in docs/
├── BUILD-VERIFICATION-REPORT.md       # ❌ Should be in docs/
├── FILTER_MIGRATION.md                # ❌ Should be in docs/
├── TESTING-SUMMARY.md                 # ❌ Should be in docs/
└── [Too many files!]
```

---

## Maintenance

### When Adding New Documentation

**Ask yourself:**
1. Is this linked from README? → Root
2. Is this a quick reference? → Root
3. Is this an implementation report? → docs/
4. Is this internal/process docs? → docs/
5. Is this detailed technical docs? → docs/

### Periodic Review

Every few releases, review:
- Are there new reports in root that should move to docs/?
- Are there old migration guides that can be archived?
- Are all README links still working?
- Is the structure still clean and logical?

---

## Current State (v5.4.0)

### Root Directory (6 files) ✅
- README.md
- CHANGELOG.md
- AI_AGENT_QUICK_REFERENCE.md
- ENHANCEMENTS.md
- OUTPUT_FORMATS.md
- CLAUDE.md

### docs/ Directory (45+ files)
All detailed documentation, implementation reports, research, and internal docs.

**Status:** Clean, organized, and maintainable! 🎉
