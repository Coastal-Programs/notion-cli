# Installation Research & Implementation Plan

## Executive Summary
This document compiles research on cross-platform npm installation issues, specifically focusing on Windows symlink problems when installing from GitHub, and designing a streamlined installation process optimized for AI agents.

---

## Current Status

### Known Issues
1. **Windows Symlink Problem**: `npm install -g github:Coastal-Programs/notion-cli` creates broken symlinks on Windows
   - Symlink points to temporary cache directory
   - Results in "Cannot find module" error
   - Local folder installation works as workaround

2. **Complex Installation Flow**: Multiple manual steps required
   - Install package
   - Set NOTION_TOKEN environment variable
   - Not AI-agent friendly

### Errors Encountered

#### Error 1: Broken Symlink on Windows
```
Error: Cannot find module 'C:\Users\jakes\AppData\Roaming\npm\node_modules\@coastal-programs\notion-cli\bin\run'
code: 'MODULE_NOT_FOUND'
```
- **Cause**: npm creates symlink to temp cache folder when installing from GitHub
- **Workaround**: Install from local folder
- **Status**: Needs permanent solution

#### Error 2: Yarn Dependency Conflict
```
npm error spawn C:\WINDOWS\system32\cmd.exe ENOENT
npm error path node_modules\yarn
```
- **Cause**: @oclif/plugin-plugins brought in yarn as transitive dependency
- **Solution**: Removed @oclif/plugin-plugins dependency
- **Status**: RESOLVED

#### Error 3: Missing dist/ Folder
- **Cause**: dist/ was in .gitignore
- **Solution**: Committed dist/ to repository
- **Status**: RESOLVED

---

## Research Questions

### 1. Windows vs Mac npm Installation Behavior
- [ ] How does npm handle GitHub installations differently on Windows vs Unix systems?
- [ ] Why do symlinks work on Mac/Linux but fail on Windows?
- [ ] What are the underlying file system differences?
- [ ] Are there npm configuration options to fix this?

### 2. npm GitHub Installation Best Practices
- [ ] What's the recommended way to distribute CLI tools?
- [ ] Should we publish to npm registry instead of GitHub installs?
- [ ] How do other popular CLIs handle this?
- [ ] What role does the "bin" field in package.json play?

### 3. Cross-Platform CLI Development
- [ ] What are best practices for Windows/Mac/Linux compatibility?
- [ ] How should the "files" field in package.json be configured?
- [ ] Should we commit compiled code (dist/) or compile during install?
- [ ] What build/lifecycle scripts are appropriate?

### 4. AI-Agent-Friendly Installation Design
- [ ] What makes a CLI "AI-agent friendly"?
- [ ] How can we automate API key configuration?
- [ ] What's the ideal installation flow?
- [ ] How do other AI-focused CLIs handle setup?

---

## Research Findings

### [To be populated by research agents]

---

## Proposed Solutions

### [To be developed after research]

---

## Implementation Plan

### Phase 1: Research (CURRENT)
- [ ] Web search for npm Windows symlink issues
- [ ] Analyze popular CLI tools (oclif, vercel, netlify-cli)
- [ ] Research npm registry publishing process
- [ ] Document AI-agent CLI best practices

### Phase 2: Solution Design
- [ ] Design cross-platform installation approach
- [ ] Create automated API key setup flow
- [ ] Plan testing strategy
- [ ] Define success criteria

### Phase 3: Implementation
- [ ] Implement chosen solution
- [ ] Add automated setup wizard
- [ ] Create cross-platform tests
- [ ] Update documentation

### Phase 4: Validation
- [ ] Test on Windows, Mac, Linux
- [ ] Test with AI agents
- [ ] Verify streamlined installation
- [ ] Performance testing

---

## Success Criteria
1. Single command installation works on all platforms: `npm install -g @coastal-programs/notion-cli`
2. API key setup is automated (prompt user or detect .env)
3. Zero manual configuration required
4. Installation takes < 30 seconds
5. Works identically for AI agents and humans

---

## Timeline
- Research Phase: [Current]
- Solution Design: [Pending]
- Implementation: [Pending]
- Validation: [Pending]
