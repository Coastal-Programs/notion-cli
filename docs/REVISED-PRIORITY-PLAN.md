# Revised Priority Plan: AI-Assistant Friendly Improvements

## Based on Real Feedback from Claude on MacBook

### **Critical Issues to Fix (Do These First):**

---

## Priority 1: Fix GitHub Installation ⭐⭐⭐⭐⭐

**Problem:** `npm install -g Coastal-Programs/notion-cli` installs but dependencies fail
**Impact:** AI assistants can't do one-command install
**Solution:** Publish to npm registry

**Why this fixes it:**
- npm registry handles dependencies automatically
- GitHub installs are unreliable (both Windows symlinks AND Mac dependency resolution)
- Industry standard for CLI distribution

**Action:**
```bash
npm login
npm publish --access public
```

**After publishing:**
```bash
# This will just work on all platforms:
npm install -g @coastal-programs/notion-cli
```

**Timeline:** 5 minutes to publish, solves 80% of installation friction

---

## Priority 2: URL Parser (Accept Full Notion URLs) ⭐⭐⭐⭐⭐

**Problem:** Users copy full URLs like `https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00?v=...`
**Current:** CLI only accepts the ID part `1fb79d4c71bb8032b722c82305b63a00`
**Impact:** Extra mental overhead for users and AI assistants

**Solution:** Add URL parser that extracts IDs from full URLs

**Implementation:**
```typescript
// src/utils/notion-url-parser.ts

export function extractNotionId(input: string): string {
  // If it's already just an ID, return it
  if (!input.includes('notion.so') && !input.includes('http')) {
    return input.replace(/-/g, '') // Remove dashes if present
  }

  // Parse full URL
  // https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00?v=...
  const match = input.match(/notion\.so\/([a-f0-9-]+)/)
  if (match) {
    return match[1].replace(/-/g, '')
  }

  throw new Error(`Invalid Notion URL or ID: ${input}`)
}
```

**Usage in all commands:**
```typescript
// Before:
async run() {
  const { args } = await this.parse(PageRetrieve)
  const pageId = args.page_id  // User must provide clean ID
}

// After:
async run() {
  const { args } = await this.parse(PageRetrieve)
  const pageId = extractNotionId(args.page_id)  // Accepts URL or ID
}
```

**Impact:** Users can copy-paste URLs directly from Notion

---

## Priority 3: Config Command (Token Management) ⭐⭐⭐⭐

**Problem:** No programmatic way to set token
**Current:** Manual environment variable setup
**What Claude wanted:** `notion-cli config set-token <token>`

**Implementation:**
```typescript
// src/commands/config/set-token.ts

import {Command, Flags} from '@oclif/core'
import * as fs from 'fs'
import * as path from 'path'
import * as os from 'os'

export default class ConfigSetToken extends Command {
  static description = 'Configure Notion API token'

  static args = {
    token: {
      description: 'Your Notion Integration Token (starts with ntn_)',
      required: true
    }
  }

  async run() {
    const {args} = await this.parse(ConfigSetToken)

    // Validate token format
    if (!args.token.startsWith('ntn_')) {
      this.error('Invalid token format. Token should start with "ntn_"')
    }

    // Determine shell
    const shell = process.env.SHELL || '/bin/bash'
    const isZsh = shell.includes('zsh')
    const rcFile = isZsh ? '.zshrc' : '.bashrc'
    const rcPath = path.join(os.homedir(), rcFile)

    // Check if token already set
    if (fs.existsSync(rcPath)) {
      const content = fs.readFileSync(rcPath, 'utf-8')
      if (content.includes('NOTION_TOKEN')) {
        this.log('⚠️  NOTION_TOKEN already exists in', rcFile)
        this.log('Updating existing token...')

        // Replace existing token
        const newContent = content.replace(
          /export NOTION_TOKEN="[^"]*"/g,
          `export NOTION_TOKEN="${args.token}"`
        )
        fs.writeFileSync(rcPath, newContent)
      } else {
        // Append new token
        fs.appendFileSync(rcPath, `\n\n# Notion CLI\nexport NOTION_TOKEN="${args.token}"\n`)
      }
    }

    this.log('✅ Token configured successfully!')
    this.log(`\nRun this command to activate it:\n  source ~/${rcFile}`)
    this.log('\nOr restart your terminal.')
    this.log('\nVerify with:\n  notion-cli user retrieve bot --output json')
  }
}
```

**Commands to add:**
- `notion-cli config set-token <token>` - Set token
- `notion-cli config get-token` - View current token (masked)
- `notion-cli config clear-token` - Remove token

---

## Priority 4: Doctor Command (Installation Verification) ⭐⭐⭐⭐

**What Claude wanted:** `notion-cli doctor`

**Implementation:**
```typescript
// src/commands/doctor.ts

import {Command} from '@oclif/core'
import * as path from 'path'
import * as fs from 'fs'
import {client} from '../notion'

export default class Doctor extends Command {
  static description = 'Verify notion-cli installation and configuration'

  async run() {
    this.log('🔍 Checking notion-cli installation...\n')

    // Check 1: CLI installed
    this.log('✅ notion-cli is installed')
    this.log(`   Version: ${this.config.version}`)
    this.log(`   Location: ${this.config.root}\n`)

    // Check 2: NOTION_TOKEN set
    if (!process.env.NOTION_TOKEN) {
      this.log('❌ NOTION_TOKEN not set')
      this.log('   Fix: notion-cli config set-token YOUR_TOKEN')
      this.log('   Or: export NOTION_TOKEN="ntn_..."')
      return this.exit(1)
    }

    this.log('✅ NOTION_TOKEN is set')
    const token = process.env.NOTION_TOKEN
    this.log(`   Token: ${token.substring(0, 8)}...${token.substring(token.length - 4)}\n`)

    // Check 3: Token valid (try API call)
    this.log('🔄 Testing connection to Notion API...')
    try {
      const user = await client.users.me({})
      this.log('✅ Connection successful!')
      this.log(`   Bot name: ${user.name || 'Unknown'}`)
      this.log(`   Bot ID: ${user.id}\n`)
    } catch (error: any) {
      this.log('❌ Connection failed')
      this.log(`   Error: ${error.message}`)
      this.log('\n   Possible causes:')
      this.log('   - Invalid token')
      this.log('   - Token expired')
      this.log('   - No internet connection')
      this.log('\n   Get a new token at: https://developers.notion.com/')
      return this.exit(1)
    }

    // Check 4: Cache directory writable
    const cacheDir = path.join(this.config.cacheDir, 'notion-cli')
    try {
      if (!fs.existsSync(cacheDir)) {
        fs.mkdirSync(cacheDir, {recursive: true})
      }
      this.log('✅ Cache directory writable')
      this.log(`   Location: ${cacheDir}\n`)
    } catch (error) {
      this.log('⚠️  Cache directory not writable (non-critical)')
      this.log(`   Location: ${cacheDir}\n`)
    }

    this.log('🎉 All checks passed! notion-cli is ready to use.\n')
    this.log('Try it out:')
    this.log('  notion-cli search --query "test" --output json')
    this.log('  notion-cli user list --output json')
  }
}
```

---

## Priority 5: Better Error Messages for AIs ⭐⭐⭐

**Current error:**
```
Error: NOTION_TOKEN not set
```

**Better error:**
```
❌ NOTION_TOKEN environment variable is not set

For AI assistants:
  notion-cli config set-token "ntn_YOUR_TOKEN"

For manual setup:
  export NOTION_TOKEN="ntn_YOUR_TOKEN"  # Mac/Linux
  set NOTION_TOKEN=ntn_YOUR_TOKEN       # Windows CMD

Get your token at: https://developers.notion.com/
```

**Already implemented!** Just need to ensure consistency across all commands.

---

## Implementation Order:

### **Week 1: Critical Fixes**

**Day 1: Publish to npm Registry** (2 hours)
- Solves 80% of installation issues
- One command install on all platforms
- Test on Windows and Mac

**Day 2: URL Parser** (3 hours)
- Create `src/utils/notion-url-parser.ts`
- Update all commands that take IDs
- Add tests
- Update documentation

**Day 3: Config Command** (4 hours)
- Create `src/commands/config/set-token.ts`
- Create `src/commands/config/get-token.ts`
- Create `src/commands/config/clear-token.ts`
- Test on Mac (zsh and bash)

**Day 4: Doctor Command** (3 hours)
- Create `src/commands/doctor.ts`
- Comprehensive checks
- Helpful error messages
- Test all failure modes

**Day 5: Documentation & Testing** (2 hours)
- Update README with new commands
- Add AI Assistant Quick Start
- Test complete workflow
- Update cookbook examples

### **Week 2: Enhancements** (Already planned)
- Batch operations
- Enhanced examples
- CI/CD

---

## Success Metrics (After Week 1):

### AI Assistant Experience:
- [ ] Single command install: `npm install -g @coastal-programs/notion-cli`
- [ ] Works on Windows, Mac, Linux without workarounds
- [ ] Token setup: `notion-cli config set-token <token>`
- [ ] Installation verify: `notion-cli doctor`
- [ ] Accept full URLs: Paste directly from Notion
- [ ] Clear error messages with next steps

### User Experience:
- [ ] Copy URL from Notion → Paste into CLI → Works
- [ ] Installation takes < 2 minutes total
- [ ] Can verify everything works with `doctor` command
- [ ] No manual .bashrc/.zshrc editing required

---

## Comparison: Before vs After

### Before (Current):
```bash
# On Windows: Broken symlinks
# On Mac: Dependencies fail

# Workaround:
git clone https://github.com/Coastal-Programs/notion-cli
cd notion-cli
npm install
npm install -g .

# Manual token setup:
echo 'export NOTION_TOKEN="ntn_xxx"' >> ~/.zshrc
source ~/.zshrc

# Use with clean IDs only:
notion-cli page retrieve 1fb79d4c71bb8032b722c82305b63a00
```

### After (Target):
```bash
# One command install (all platforms):
npm install -g @coastal-programs/notion-cli

# One command config:
notion-cli config set-token "ntn_xxx"
source ~/.zshrc

# Verify everything:
notion-cli doctor

# Use with full URLs:
notion-cli page retrieve https://www.notion.so/1fb79d4c71bb8032b722c82305b63a00?v=...
```

**From 7 commands + workarounds → 4 commands that just work**

---

## Risk Assessment:

### Low Risk:
- ✅ Publishing to npm registry (standard practice)
- ✅ URL parser (additive, doesn't break existing usage)
- ✅ Doctor command (new command, no conflicts)

### Medium Risk:
- ⚠️ Config command (need to handle different shells correctly)
- ⚠️ Testing on multiple platforms

### Mitigation:
- Test on Mac (zsh, bash)
- Test on Windows (cmd, powershell)
- Make shell editing optional (fallback to instructions)
- Add `--dry-run` flag to show what would be changed

---

## Agent Orchestration for Week 1:

**Parallel Track 1:**
- Agent: DevOps (npm publishing)
- Agent: Documentation (update README)

**Parallel Track 2:**
- Agent: Backend (URL parser utility)
- Agent: Frontend (update all commands to use parser)

**Parallel Track 3:**
- Agent: Frontend (config commands)
- Agent: Frontend (doctor command)

**Parallel Track 4:**
- Agent: Test Writer (comprehensive tests)
- Agent: Documentation (AI quick start guide)

---

## Next Steps:

1. **Publish to npm registry** (solves 80% of Claude's issues)
2. **Implement URL parser** (solves copy-paste workflow)
3. **Add config/doctor commands** (makes setup programmatic)
4. **Update documentation** (clear AI assistant instructions)

**Ready to start?** Should we begin with npm publishing?
