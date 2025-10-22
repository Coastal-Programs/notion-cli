# Testing Guide for Beta Testers

This guide explains how to test notion-cli on both Mac and Windows platforms before the official npm registry release.

---

## For Mac/Linux Testers

### Option 1: GitHub Install (Recommended)
```bash
npm install -g Coastal-Programs/notion-cli
```

### Option 2: Local Install
```bash
git clone https://github.com/Coastal-Programs/notion-cli
cd notion-cli
npm install -g .
```

Both methods work on Mac/Linux.

---

## For Windows Testers

**⚠️ Important:** GitHub installations have a known issue on Windows (broken symlinks). You must use the local install method.

### Local Install (Required for Windows)
```bash
git clone https://github.com/Coastal-Programs/notion-cli
cd notion-cli
npm install -g .
```

**Why this issue exists:** npm creates broken symlinks when installing from GitHub on Windows. This is documented in `docs/research-npm-windows-symlinks.md`. The issue will be resolved when we publish to npm registry.

---

## Setup Your API Token

### 1. Get Your Token
1. Go to https://developers.notion.com/
2. Click "Create new integration"
3. Give it a name (e.g., "notion-cli testing")
4. Copy your "Internal Integration Token" (starts with `ntn_`)

### 2. Share Your Workspace
1. Open any page in your Notion workspace
2. Click "Share" in the top right
3. Invite your integration
4. This gives the CLI access to your workspace

### 3. Set the Token

**Mac/Linux:**
```bash
export NOTION_TOKEN="ntn_your_token_here"

# Make it permanent (optional)
echo 'export NOTION_TOKEN="ntn_your_token_here"' >> ~/.bashrc
source ~/.bashrc
```

**Windows (Command Prompt):**
```cmd
set NOTION_TOKEN=ntn_your_token_here

REM Make it permanent (optional)
setx NOTION_TOKEN "ntn_your_token_here"
```

**Windows (PowerShell):**
```powershell
$env:NOTION_TOKEN="ntn_your_token_here"

# Make it permanent (optional)
[System.Environment]::SetEnvironmentVariable('NOTION_TOKEN', 'ntn_your_token_here', 'User')
```

---

## Verify Installation

### Test 1: Version Check
```bash
notion-cli --version
```

**Expected output:**
```
@coastal-programs/notion-cli/5.1.0 [platform]-[arch] node-v[version]
```

### Test 2: Help Command
```bash
notion-cli --help
```

**Expected:** Should display list of available commands

### Test 3: API Connection
```bash
notion-cli user retrieve bot --output json
```

**Expected output:**
```json
[
  {
    "id": "your-bot-id",
    "name": "Your Integration Name",
    "object": "user",
    "type": "bot",
    ...
  }
]
```

**If you see an error about NOTION_TOKEN:**
The new error message should tell you exactly how to fix it:
```
NOTION_TOKEN environment variable is not set.

To fix:
  export NOTION_TOKEN="your-token-here"  # Mac/Linux
  set NOTION_TOKEN=your-token-here       # Windows CMD
  $env:NOTION_TOKEN="your-token-here"    # Windows PowerShell

Get your token at: https://developers.notion.com/docs/create-a-notion-integration
```

---

## Test Commands

### Basic Commands to Test

**1. List Users:**
```bash
notion-cli user list --output json
```

**2. Search Pages:**
```bash
notion-cli search --query "test" --output json
```

**3. Retrieve a Page:**
```bash
# Replace PAGE_ID with actual page ID from your workspace
notion-cli page retrieve PAGE_ID --output json
```

**4. Query a Database/Data Source:**
```bash
# Replace DATA_SOURCE_ID with actual database/table ID
notion-cli db query DATA_SOURCE_ID --output json
```

**5. Test Different Output Formats:**
```bash
notion-cli user list --output table
notion-cli user list --output csv
notion-cli user list --output yaml
notion-cli user list --output json
```

---

## What to Report

### Issues to Report:

1. **Installation Problems:**
   - Did the installation method work?
   - Were the instructions clear?
   - How long did installation take?

2. **Error Messages:**
   - Were error messages helpful?
   - Did they tell you how to fix problems?
   - Any confusing error messages?

3. **Command Functionality:**
   - Which commands did you test?
   - Did they work as expected?
   - Any unexpected behavior?

4. **Documentation:**
   - Was the README clear?
   - What was confusing?
   - What could be improved?

5. **Platform-Specific Issues:**
   - Note your OS: Windows 10/11, Mac (Intel/M1/M2), Linux distro
   - Node.js version: `node --version`
   - npm version: `npm --version`

### Where to Report:

Create an issue at: https://github.com/Coastal-Programs/notion-cli/issues

**Issue Template:**
```markdown
## Platform
- OS: [Windows 11 / Mac M2 / Ubuntu 22.04]
- Node.js: [v22.17.0]
- npm: [v10.9.2]

## Installation Method
- [ ] GitHub install (Mac/Linux only)
- [ ] Local install (git clone)

## Issue Description
[Describe what happened]

## Expected Behavior
[What you expected to happen]

## Steps to Reproduce
1. [First step]
2. [Second step]
3. [...]

## Error Messages
```
[Paste any error messages here]
```

## Additional Context
[Any other relevant information]
```

---

## Known Issues

### Windows GitHub Installation
- **Issue:** `npm install -g Coastal-Programs/notion-cli` creates broken symlinks
- **Status:** Known issue, documented in research
- **Workaround:** Use local install method
- **Fix:** Will be resolved when published to npm registry

### Error Message Improvements
- **Status:** Recently improved in v5.1.0
- **What Changed:** Error messages now include platform-specific setup commands
- **Feedback Needed:** Are the new error messages helpful?

---

## Success Criteria

For the CLI to be considered "beta ready":

- ✅ Works on Windows (local install)
- ✅ Works on Mac (both install methods)
- ✅ Works on Linux (both install methods)
- ✅ Clear error messages
- ✅ All major commands functional
- ✅ Documentation is clear
- ✅ Installation takes < 5 minutes

---

## After Testing

Once beta testing is complete and issues are resolved, we'll publish to npm registry. This will:
- Fix Windows installation (no more local install required)
- Make it available via `npm install -g @coastal-programs/notion-cli`
- Enable version updates via `npm update -g @coastal-programs/notion-cli`

Thank you for helping test notion-cli!
