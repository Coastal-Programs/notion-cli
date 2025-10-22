# Cross-Platform Node.js Package Development Research

**Date:** 2025-10-22
**Context:** Research for notion-cli cross-platform compatibility
**Target Platforms:** Windows, macOS, Linux

---

## Executive Summary

Cross-platform compatibility is critical for npm packages, as 24% of Node.js developers use Windows locally, 41% use Mac, while 85% of production environments run Linux. This research document compiles best practices, common pitfalls, and solutions for building robust cross-platform Node.js CLI tools.

**Key Takeaway:** Use Node.js APIs and cross-platform npm packages instead of shell commands, leverage the `path` module for all file operations, and test on all target platforms with CI/CD.

---

## Table of Contents

1. [Common Cross-Platform Issues](#1-common-cross-platform-issues)
2. [Testing Strategies](#2-testing-strategies-for-multi-platform-packages)
3. [File System Differences](#3-file-system-differences)
4. [Platform-Specific npm Behaviors](#4-platform-specific-npm-behaviors)
5. [Tools and Frameworks](#5-tools-and-frameworks-for-cross-platform-cli-development)
6. [Windows-Specific Issues](#6-windows-specific-issues)
7. [Best Practices from Node.js Documentation](#7-best-practices-from-nodejs-and-npm-documentation)
8. [Resources and References](#8-resources-and-references)

---

## 1. Common Cross-Platform Issues

### 1.1 Path Separators

**Problem:** Windows uses backslashes (`\`) while POSIX systems (Linux, macOS) use forward slashes (`/`).

**Solution:**
```javascript
// BAD - Hardcoded separators
const filePath = 'src' + '/' + 'index.js';  // Fails on Windows
const filePath2 = 'src\\index.js';          // Fails on Linux/Mac

// GOOD - Use path module
const path = require('path');
const filePath = path.join('src', 'index.js');  // Works everywhere
```

**Key APIs:**
- `path.join()` - Joins path segments using platform-specific separator
- `path.resolve()` - Resolves to absolute path
- `path.sep` - Platform-specific separator (`/` or `\`)
- `path.normalize()` - Normalizes paths for current platform

**Warning:** Never use path methods with URLs - Windows users will get URLs with backslashes!

### 1.2 Shell Command Differences

**Problem:** Different shells (cmd.exe on Windows, bash on Linux/Mac) have incompatible syntax.

**Examples of Incompatibilities:**
- Environment variables: `set VAR=value` (Windows) vs `export VAR=value` (Unix)
- Path separator in PATH: `;` (Windows) vs `:` (Unix)
- Multiple commands: `&&` works on both, but `;` only works on Unix
- Globbing: `*` expansion works differently across shells
- Quoting: Single quotes work differently in cmd.exe

**Solution:** Use Node.js APIs instead of shell commands.

```javascript
// BAD - Shell-dependent
exec('rm -rf dist');        // Unix only
exec('del /s /q dist');     // Windows only

// GOOD - Use Node.js or npm packages
const rimraf = require('rimraf');
rimraf.sync('dist');  // Works everywhere
```

### 1.3 Case Sensitivity

**Problem:** Windows and macOS filesystems are case-insensitive by default, but Linux is case-sensitive.

**Impact:**
```javascript
// Works on Windows/Mac, fails on Linux
require('./MyModule');  // File is actually myModule.js
```

**Solutions:**
1. **Always match case exactly** in require/import statements
2. **TypeScript:** Enable `"forceConsistentCasingInFileNames": true` in tsconfig.json
3. **ESLint:** Use `eslint-plugin-import` with "no-unresolved" rule
4. **Webpack:** Use case-sensitive path plugin
5. **Testing:** Run tests on Linux to catch case issues

**Note:** Don't assume filesystem behavior from `process.platform` - Mac can use case-sensitive HFSX volumes.

### 1.4 Line Endings (EOL)

**Problem:** Windows uses `\r\n` (CRLF), Unix uses `\n` (LF).

**Solution:**
```javascript
const os = require('os');

// GOOD - Platform-appropriate line endings
const content = lines.join(os.EOL);

// Also good for splitting
const lines = content.split(os.EOL);
```

**Git Configuration:**
```gitattributes
# .gitattributes file
* text=auto
*.js text eol=lf
*.json text eol=lf
*.md text eol=lf
```

### 1.5 Module Name Casing

**Problem:** Windows and macOS use case-insensitive file systems, so `require("foo")` and `require("FOO")` work there but fail on Linux production systems.

**Solution:** Always get module and package name casing exactly right. Use linting tools to enforce this.

---

## 2. Testing Strategies for Multi-Platform Packages

### 2.1 GitHub Actions Matrix Strategy

**Recommended Approach:** Use matrix builds to test across all platforms and Node versions.

```yaml
# .github/workflows/test.yml
name: Cross-Platform Tests

on: [push, pull_request]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, windows-latest, macos-latest]
        node-version: [18.x, 20.x, 22.x]

    runs-on: ${{ matrix.os }}

    steps:
      - uses: actions/checkout@v4

      - name: Setup Node.js ${{ matrix.node-version }}
        uses: actions/setup-node@v4
        with:
          node-version: ${{ matrix.node-version }}

      - name: Install dependencies
        run: npm ci

      - name: Run tests
        run: npm test

      - name: Run linter
        run: npm run lint
```

**Benefits:**
- Tests across all major operating systems in parallel
- Validates against multiple Node.js versions
- `npm ci` ensures deterministic installations
- Early detection of platform-specific issues

### 2.2 Local Testing Strategies

**For Windows Developers:**
1. Use WSL2 (Windows Subsystem for Linux) to test Linux compatibility
2. Test in both PowerShell and cmd.exe
3. Test with Windows Terminal and Git Bash

**For Mac Developers:**
1. Use Docker containers for Linux testing
2. Consider using Windows VM or Parallels
3. Test with both default HFS+ and case-sensitive HFSX volumes

**For Linux Developers:**
1. Use Wine for Windows testing (limited)
2. Use VirtualBox/VMware for Windows VM
3. Use Docker for different Linux distributions

### 2.3 Test Coverage Areas

**Must Test:**
- File path operations
- Environment variable handling
- Process spawning and child processes
- File system operations (read, write, delete, symlinks)
- Exit codes and error handling
- Terminal output and colors
- Binary execution via `bin` field

---

## 3. File System Differences

### 3.1 Symlinks and Permissions

#### Symlinks

**Windows:**
- Creating symlinks requires administrator privileges (or Developer Mode in Windows 10+)
- `fs.symlink()` often fails with EPERM error
- Directory symlinks cannot be created without "Run As Administrator"
- **Alternative:** Use junctions (directory hard links) - they don't require permissions
- **Alternative:** Use hard links via `fs.link()` - they don't require permissions

**Linux/Mac:**
- Regular users can create symlinks without special permissions
- Full support for symbolic links

**Cross-Platform Solution:**
```javascript
// Instead of using symlinks, copy files
const fs = require('fs-extra');

// More reliable cross-platform
await fs.copy(source, dest);

// Or use hard links if appropriate
fs.linkSync(existingPath, newPath);
```

**For npm packages:** npm automatically creates appropriate wrappers (symlinks on Unix, .cmd files on Windows) in `node_modules/.bin/` directory.

#### File Permissions

**Unix (Linux/Mac):**
- POSIX permission model (user/group/other, read/write/execute)
- `fs.chmod()`, `fs.chown()` work as expected
- File mode bits (0o755, 0o644, etc.)

**Windows:**
- Uses ACLs (Access Control Lists)
- Different file attributes: readonly, hidden, system
- **Important:** Node.js does NOT fully support Windows permissions
- `fs.chmod()` has limited functionality on Windows

**Cross-Platform Approach:**
- Use `fs.access()` to check read/write permissions
- Don't rely on Unix permission bits on Windows
- For executables, ensure proper shebang lines (npm handles the rest)

### 3.2 File System Type Support

**Considerations:**
- FAT32/exFAT do not support symlinks (affects USB drives, some Windows partitions)
- NTFS supports symlinks but requires permissions
- ext4, APFS, HFS+ all support symlinks natively
- Case sensitivity varies by filesystem (not just OS)

**Best Practice:** Always handle file system operations failures gracefully with proper error messages.

### 3.3 Path Length Limitations

**Windows:**
- Legacy limitation: 260 characters (MAX_PATH)
- Can be extended with registry setting or Windows 10 version 1607+
- Long path prefix: `\\?\` can extend to ~32,767 characters
- Node.js has some workarounds built-in

**Linux/Mac:**
- PATH_MAX typically 4096 characters
- Individual filename component limit: 255 bytes

**Solution:** Keep paths reasonably short, nest dependencies appropriately.

---

## 4. Platform-Specific npm Behaviors

### 4.1 npm Binary Wrappers

**How it Works:**

When you define a `bin` field in package.json:
```json
{
  "bin": {
    "notion-cli": "./bin/run"
  }
}
```

**On Unix (Linux/Mac):**
- npm creates symlink: `node_modules/.bin/notion-cli` → `../../bin/run`
- Symlink uses shebang line from target file
- Shebang example: `#!/usr/bin/env node`

**On Windows:**
- npm creates **two** files:
  - `node_modules/.bin/notion-cli` (bash script for Git Bash)
  - `node_modules/.bin/notion-cli.cmd` (batch file for cmd.exe)
- The .cmd wrapper parses the shebang line and invokes node explicitly
- **Critical:** Avoid .js extension in bin names (Windows file association issues)

**Best Practice for Shebang:**
```bash
#!/usr/bin/env node
# NOT #!/usr/local/bin/node (hardcoded path)
```

### 4.2 npm Scripts Execution

**Shell Used:**
- Windows: `cmd.exe` (or `ComSpec` environment variable)
- Unix: `/bin/sh`

**Implications:**
- Different quoting rules (double quotes on Windows, single/double on Unix)
- Different command separators (`&&` works on both, `;` only on Unix)
- Globbing handled differently
- No wildcard expansion in cmd.exe

**Solution:** Use cross-platform tools like `cross-env`, `npm-run-all`, and Node.js scripts.

### 4.3 Optional Dependencies and Platform Targeting

**package.json fields:**

```json
{
  "os": ["darwin", "linux", "!win32"],
  "cpu": ["x64", "arm64"],
  "optionalDependencies": {
    "fsevents": "^2.3.2"  // Mac-only file watcher
  }
}
```

**Limitations:**
- `optionalDependencies` designed for truly optional features
- No built-in way to specify "required on specific platform only"
- Can cause issues with package-lock.json on different platforms

**Best Practice:** Keep platform-specific dependencies to minimum, document requirements clearly.

---

## 5. Tools and Frameworks for Cross-Platform CLI Development

### 5.1 oclif Framework (Used by notion-cli)

**Overview:**
- Open-source CLI framework by Salesforce (used for Heroku CLI)
- Built on Node.js, runs on all major operating systems
- **Key Feature:** Convention over configuration

**Cross-Platform Benefits:**
- Automatic platform-specific installers (no Node.js required)
- Handles binary creation for Windows, Mac, Linux
- TypeScript support out of the box
- Plugin system for extensibility

**Commands:**
```bash
# Build standalone installers for all platforms
oclif pack tarballs
oclif pack win
oclif pack macos
oclif pack deb
```

**Current Usage in notion-cli:**
```json
{
  "oclif": {
    "bin": "notion-cli",
    "dirname": "notion-cli",
    "commands": "./dist/commands"
  }
}
```

**Version Support:** Node 18+ (follows Node.js LTS versions)

### 5.2 Essential Cross-Platform npm Packages

#### File System Operations

**rimraf** - Cross-platform `rm -rf`
```javascript
const rimraf = require('rimraf');
rimraf.sync('build/**/*.js');
```

**mkdirp** - Cross-platform `mkdir -p`
```javascript
const mkdirp = require('mkdirp');
await mkdirp('path/to/nested/dir');
```

**cpy** / **ncp** - Cross-platform file copying
```javascript
const cpy = require('cpy');
await cpy(['src/*.js'], 'dist');
```

**fs-extra** - Enhanced fs with cross-platform methods
```javascript
const fs = require('fs-extra');
await fs.copy('src', 'dest');
await fs.emptyDir('tmp');
```

#### Shell Commands

**shx** - ShellJS as CLI (cross-platform Unix commands)
```json
{
  "scripts": {
    "clean": "shx rm -rf dist",
    "copy": "shx cp -r src dist"
  }
}
```

**Important:** Wrap globs in quotes for shx: `shx rm -rf "dist/**/*.js"`

**cross-env** - Cross-platform environment variables
```json
{
  "scripts": {
    "build": "cross-env NODE_ENV=production webpack"
  }
}
```

**npm-run-all** - Run multiple npm scripts
```json
{
  "scripts": {
    "build": "npm-run-all clean compile test",
    "build:parallel": "npm-run-all --parallel build:*"
  }
}
```

#### Process Management

**cross-spawn** - Cross-platform child_process.spawn
```javascript
const spawn = require('cross-spawn');

// Works on Windows and Unix
const child = spawn('npm', ['install'], { stdio: 'inherit' });
```

**execa** - Better child_process (includes cross-spawn)
```javascript
const execa = require('execa');

const { stdout } = await execa('echo', ['hello']);
console.log(stdout);  // 'hello'
```

### 5.3 Current Tools in notion-cli

**From package.json:**
```json
{
  "devDependencies": {
    "shx": "^0.3.4",  // Cross-platform shell commands
    "oclif": "^3",     // CLI framework
    // ... others
  },
  "scripts": {
    "build": "shx rm -rf dist && tsc -b",
    "postpack": "shx rm -f oclif.manifest.json",
    "readme": "oclif readme --multi --no-aliases && shx sed -i \"s/^_See code:.*$//g\" docs/*.md > /dev/null"
  }
}
```

**Observations:**
- ✅ Already using `shx` for file operations
- ✅ Using TypeScript for cross-platform logic
- ⚠️ `sed` command in readme script may fail on Windows without Git Bash
- ⚠️ `/dev/null` is Unix-specific

**Recommended Improvements:**
```json
{
  "scripts": {
    "readme": "oclif readme --multi --no-aliases && node -e \"require('fs').writeFileSync('docs/temp.md', '')\""
  }
}
```

---

## 6. Windows-Specific Issues

### 6.1 cmd.exe vs Bash Differences

**cmd.exe Limitations:**
- No command chaining with `;` (use `&&` instead)
- No wildcard/glob expansion
- Different quoting rules (only double quotes work)
- Different environment variable syntax (`%VAR%` vs `$VAR`)
- Limited command set (no `cp`, `rm`, `grep`, etc.)

**PowerShell:**
- Better than cmd.exe but still different from bash
- Has aliases (`rm`, `cp`, etc.) but behavior differs
- Default in Windows 10+ but not always available in CI

**Git Bash:**
- Provides bash-like environment on Windows
- Good for developers but not available in all environments
- Not default for npm scripts

**Solution:** Use Node.js APIs or cross-platform npm packages, not shell commands.

### 6.2 Shebang Handling

**Problem:** Windows ignores shebang lines like `#!/usr/bin/env node`

**How npm Solves This:**

For a bin entry:
```json
{
  "bin": {
    "cli-tool": "./bin/cli.js"
  }
}
```

npm creates on Windows:
```batch
@rem cli-tool.cmd
@IF EXIST "%~dp0\node.exe" (
  "%~dp0\node.exe"  "%~dp0\cli.js" %*
) ELSE (
  node  "%~dp0\cli.js" %*
)
```

**Best Practice:**
1. Use standard shebang: `#!/usr/bin/env node`
2. Avoid .js extension in bin script names
3. Let npm handle wrapper creation

### 6.3 Character Encoding Issues

**Problem:** Windows console default code page is not UTF-8 (often CP437, CP850, Windows-1252)

**Node.js Behavior:**
- Node.js defaults to UTF-8
- Windows console may display incorrectly

**Solutions:**

1. **Set UTF-8 in scripts:**
```javascript
// For child_process spawning
const child = spawn('cmd', ['/c', 'chcp 65001>nul && your-command']);
```

2. **Set for entire process:**
```javascript
if (process.platform === 'win32') {
  // Set Windows console to UTF-8
  require('child_process').execSync('chcp 65001');
}
```

3. **Use iconv-lite for encoding conversion:**
```javascript
const iconv = require('iconv-lite');
const decoded = iconv.decode(buffer, 'win1252');
```

### 6.4 Path Handling Issues

**Absolute Paths:**
- Windows: `C:\Users\username\file.txt`
- Unix: `/home/username/file.txt`
- Different drive letters (C:, D:, etc.)

**UNC Paths:**
- Windows network paths: `\\server\share\file.txt`
- No equivalent on Unix (use mount points)

**Solutions:**
```javascript
const path = require('path');
const os = require('os');

// Home directory
const homedir = os.homedir();  // Cross-platform

// Temp directory
const tmpdir = os.tmpdir();    // Cross-platform

// Current working directory (cwd)
const cwd = process.cwd();     // Cross-platform
```

### 6.5 Argument Escaping in child_process

**Complex Issue:** Windows cmd.exe has different escaping rules than Unix shells

**Problems:**
- Spaces in arguments require quotes
- Quotes themselves need escaping
- Special characters (%, ^, &, |, <, >) need escaping

**Solution:** Use `cross-spawn` or `execa` packages
```javascript
// These handle escaping automatically
const spawn = require('cross-spawn');
spawn('echo', ['hello world'], { stdio: 'inherit' });
```

### 6.6 ENOENT Errors

**Problem:** Windows expects .exe files for spawning

**Example:**
```javascript
// Fails on Windows
spawn('npm', ['install']);

// Works on both (cross-spawn handles it)
const spawn = require('cross-spawn');
spawn('npm', ['install']);
```

**Why:** Windows looks for `npm.exe`, but npm is actually `npm.cmd`

**Solution:**
1. Use `cross-spawn` package
2. Or use shell option: `spawn('npm', ['install'], { shell: true })`

---

## 7. Best Practices from Node.js and npm Documentation

### 7.1 Official Node.js Recommendations

**From Node.js Documentation:**

1. **Use path module for ALL file paths**
   ```javascript
   // Good
   const fullPath = path.join(__dirname, 'data', 'file.txt');

   // Bad
   const fullPath = __dirname + '/data/file.txt';
   ```

2. **Use os module for system information**
   ```javascript
   const os = require('os');

   os.platform();    // 'win32', 'darwin', 'linux'
   os.homedir();     // User home directory
   os.tmpdir();      // Temp directory
   os.EOL;           // '\r\n' or '\n'
   os.arch();        // 'x64', 'arm64', etc.
   ```

3. **Don't assume filesystem behavior from process.platform**
   - Mac can use case-sensitive filesystems (HFSX)
   - Windows can have case-sensitive directories (Windows 10+)
   - Test actual behavior, not platform

4. **For cross-platform file operations**
   ```javascript
   const fs = require('fs').promises;

   // Check access
   await fs.access(path, fs.constants.R_OK | fs.constants.W_OK);

   // Use fs.promises for async/await
   await fs.readFile(path, 'utf8');
   ```

### 7.2 npm Package Best Practices

**From npm Documentation:**

1. **Specify engine requirements**
   ```json
   {
     "engines": {
       "node": ">=18.0.0",
       "npm": ">=9.0.0"
     }
   }
   ```

2. **List supported platforms**
   ```json
   {
     "os": ["darwin", "linux", "win32"],
     "cpu": ["x64", "arm64"]
   }
   ```

3. **Use files field to control package contents**
   ```json
   {
     "files": [
       "/bin",
       "/dist",
       "/npm-shrinkwrap.json",
       "/oclif.manifest.json"
     ]
   }
   ```

4. **Proper bin configuration**
   ```json
   {
     "bin": {
       "notion-cli": "./bin/run"
     }
   }
   ```

### 7.3 Cross-Platform Script Guidelines

**Writing npm Scripts:**

1. **Use Node.js for complex operations**
   ```json
   {
     "scripts": {
       "clean": "node scripts/clean.js",
       "build": "node scripts/build.js"
     }
   }
   ```

2. **Use cross-platform tools for simple operations**
   ```json
   {
     "scripts": {
       "clean": "shx rm -rf dist",
       "copy": "shx cp -r src dist",
       "build": "cross-env NODE_ENV=production tsc"
     }
   }
   ```

3. **Chain commands safely**
   ```json
   {
     "scripts": {
       "build": "npm run clean && npm run compile",
       "test:all": "npm-run-all test:unit test:integration"
     }
   }
   ```

4. **Quote glob patterns**
   ```json
   {
     "scripts": {
       "clean": "shx rm -rf \"dist/**/*.js\""
     }
   }
   ```

### 7.4 Error Handling

**Cross-Platform Error Codes:**

Node.js uses standardized POSIX error codes across platforms:
- `ENOENT` - No such file or directory
- `EACCES` - Permission denied
- `EEXIST` - File already exists
- `ENOTDIR` - Not a directory
- `EISDIR` - Is a directory

**These work consistently across Windows, Mac, and Linux.**

**Handling Errors:**
```javascript
try {
  await fs.access(filePath);
} catch (error) {
  if (error.code === 'ENOENT') {
    console.error('File not found:', filePath);
  } else if (error.code === 'EACCES') {
    console.error('Permission denied:', filePath);
  } else {
    throw error;
  }
}
```

### 7.5 Exit Codes

**Standard Exit Codes:**
- `0` - Success
- `1` - General error
- `2` - Misuse of shell command
- `126` - Command cannot execute
- `127` - Command not found
- `128 + N` - Fatal error signal N

**Best Practice:**
```javascript
// Good
process.exit(0);  // Success
process.exit(1);  // Error

// Better - let Node.js exit naturally
if (error) {
  console.error(error);
  process.exitCode = 1;
}
```

### 7.6 Environment Variables

**Cross-Platform Access:**
```javascript
// Good - works everywhere
const token = process.env.NOTION_TOKEN;
const nodeEnv = process.env.NODE_ENV;

// Setting in scripts
{
  "scripts": {
    "dev": "cross-env NODE_ENV=development node app.js"
  }
}
```

**Common Environment Variables:**
- `PATH` - Executable search path (`;` on Windows, `:` on Unix)
- `HOME` / `USERPROFILE` - User home (use `os.homedir()` instead)
- `TEMP` / `TMP` / `TMPDIR` - Temp directory (use `os.tmpdir()` instead)

---

## 8. Resources and References

### 8.1 Essential Guides

1. **ehmicky/cross-platform-node-guide**
   - GitHub: https://github.com/ehmicky/cross-platform-node-guide
   - Comprehensive guide covering all aspects of cross-platform Node.js
   - Topics: paths, shell, filesystem, encoding, permissions, terminal

2. **bcoe/awesome-cross-platform-nodejs**
   - GitHub: https://github.com/bcoe/awesome-cross-platform-nodejs
   - Curated list of cross-platform tools and libraries
   - Categories: Testing, file system, shell, process management

3. **Tips for Writing Portable Node.js Code**
   - Gist: https://gist.github.com/domenic/2790533
   - Classic reference by Domenic Denicola (TC39)
   - Core principles still relevant

### 8.2 Official Documentation

1. **Node.js Documentation**
   - Path module: https://nodejs.org/api/path.html
   - OS module: https://nodejs.org/api/os.html
   - Child Process: https://nodejs.org/api/child_process.html
   - File System: https://nodejs.org/api/fs.html
   - Working with Different Filesystems: https://nodejs.org/en/docs/guides/working-with-different-filesystems

2. **npm Documentation**
   - package.json: https://docs.npmjs.com/cli/v7/configuring-npm/package-json
   - npm scripts: https://docs.npmjs.com/cli/v7/using-npm/scripts

3. **oclif Documentation**
   - Main site: https://oclif.io/
   - GitHub: https://github.com/oclif/oclif
   - Building installers: https://fek.io/blog/how-to-build-oclif-installers-for-mac-os-windows-and-linux-with-github-actions/

### 8.3 Cross-Platform Tools

**File System:**
- rimraf: https://www.npmjs.com/package/rimraf
- mkdirp: https://www.npmjs.com/package/mkdirp
- fs-extra: https://www.npmjs.com/package/fs-extra
- cpy: https://www.npmjs.com/package/cpy

**Shell Commands:**
- shx: https://www.npmjs.com/package/shx
- cross-env: https://www.npmjs.com/package/cross-env
- npm-run-all: https://www.npmjs.com/package/npm-run-all

**Process Management:**
- cross-spawn: https://www.npmjs.com/package/cross-spawn
- execa: https://www.npmjs.com/package/execa

### 8.4 Testing Tools

**GitHub Actions:**
- Building and testing Node.js: https://docs.github.com/en/actions/use-cases-and-examples/building-and-testing/building-and-testing-nodejs

**Local Testing:**
- act: https://github.com/nektos/act (Run GitHub Actions locally)
- Docker Desktop: https://www.docker.com/products/docker-desktop

### 8.5 Articles and Tutorials

1. **Writing cross-platform Node.js** by George Ornbo
   - URL: https://shapeshed.com/writing-cross-platform-node/

2. **Running Cross-Platform scripts in NodeJs** by Saravanan M
   - URL: https://imsaravananm.medium.com/running-cross-platform-scripts-in-nodejs-2af9f06babf7

3. **Cross-platform Node.js** by Alan Norbauer
   - URL: https://alan.norbauer.com/articles/cross-platform-nodejs/

4. **Running cross-platform tasks via npm package scripts**
   - URL: https://2ality.com/2022/08/npm-package-scripts.html

---

## 9. Implementation Checklist for notion-cli

### 9.1 Current Status

**✅ Already Following:**
- Using `shx` for cross-platform shell commands
- Using `path` module in source code
- Using oclif framework (inherently cross-platform)
- TypeScript compilation to CommonJS
- Proper bin configuration
- GitHub Actions CI (needs expansion)

**⚠️ Needs Review:**
- npm scripts may have Unix-specific commands
- Missing Windows-specific testing
- No explicit cross-platform testing in CI
- Case sensitivity not enforced by linting

### 9.2 Recommended Actions

**Immediate (High Priority):**

1. **Expand GitHub Actions Matrix**
   ```yaml
   strategy:
     matrix:
       os: [ubuntu-latest, windows-latest, macos-latest]
       node-version: [18.x, 20.x, 22.x]
   ```

2. **Enable TypeScript Case Sensitivity Check**
   ```json
   // tsconfig.json
   {
     "compilerOptions": {
       "forceConsistentCasingInFileNames": true
     }
   }
   ```

3. **Review npm Scripts for Cross-Platform Issues**
   - Replace Unix-specific commands with Node.js or shx equivalents
   - Test all scripts on Windows

4. **Add ESLint Import Plugin**
   ```bash
   npm install --save-dev eslint-plugin-import
   ```
   ```json
   // .eslintrc.json
   {
     "plugins": ["import"],
     "rules": {
       "import/no-unresolved": "error"
     }
   }
   ```

**Medium Priority:**

5. **Add Cross-Platform Testing Documentation**
   - Document how to test on Windows
   - Document how to test on Linux (WSL, Docker)
   - Add to CONTRIBUTING.md

6. **Review File System Operations**
   - Audit all `fs` operations for cross-platform compatibility
   - Ensure proper error handling for ENOENT, EACCES
   - Document any platform-specific behavior

7. **Add Platform Detection Utilities**
   ```javascript
   // src/platform-utils.ts
   export const isWindows = process.platform === 'win32';
   export const isMac = process.platform === 'darwin';
   export const isLinux = process.platform === 'linux';
   ```

**Long-Term (Nice to Have):**

8. **Add Windows-Specific Tests**
   - Test with cmd.exe
   - Test with PowerShell
   - Test with Git Bash

9. **Create Platform-Specific Build Artifacts**
   - Use oclif pack for standalone installers
   - Test installers on each platform

10. **Add Integration Tests**
    - Test actual CLI commands on all platforms
    - Test file operations in temp directories
    - Test error scenarios

### 9.3 Testing Checklist

**Before Each Release:**

- [ ] Run tests on Windows
- [ ] Run tests on macOS
- [ ] Run tests on Linux
- [ ] Test in both Node 18 and 20+
- [ ] Test npm scripts on Windows
- [ ] Test npm scripts on Unix
- [ ] Verify bin execution on Windows (cmd.exe and PowerShell)
- [ ] Verify bin execution on Unix (bash, zsh)
- [ ] Check for case sensitivity issues
- [ ] Test with long file paths (Windows)
- [ ] Test with special characters in paths
- [ ] Verify line endings are consistent
- [ ] Test error messages on all platforms
- [ ] Verify exit codes

---

## 10. Common Pitfalls and How to Avoid Them

### 10.1 Path Construction Mistakes

❌ **Wrong:**
```javascript
const configPath = process.cwd() + '/config.json';
const srcPath = 'src/commands/' + filename;
```

✅ **Right:**
```javascript
const path = require('path');
const configPath = path.join(process.cwd(), 'config.json');
const srcPath = path.join('src', 'commands', filename);
```

### 10.2 Shell Command Mistakes

❌ **Wrong:**
```json
{
  "scripts": {
    "clean": "rm -rf dist",
    "copy": "cp -r src dist"
  }
}
```

✅ **Right:**
```json
{
  "scripts": {
    "clean": "shx rm -rf dist",
    "copy": "shx cp -r src dist"
  }
}
```

### 10.3 Environment Variable Mistakes

❌ **Wrong:**
```json
{
  "scripts": {
    "build": "NODE_ENV=production webpack"
  }
}
```

✅ **Right:**
```json
{
  "scripts": {
    "build": "cross-env NODE_ENV=production webpack"
  }
}
```

### 10.4 Case Sensitivity Mistakes

❌ **Wrong:**
```javascript
// File is actually myModule.js
import { MyClass } from './MyModule';
```

✅ **Right:**
```javascript
// Match exact case
import { MyClass } from './myModule';
```

### 10.5 Line Ending Mistakes

❌ **Wrong:**
```javascript
const content = lines.join('\n');
```

✅ **Right:**
```javascript
const os = require('os');
const content = lines.join(os.EOL);
```

### 10.6 Process Spawning Mistakes

❌ **Wrong:**
```javascript
const { spawn } = require('child_process');
spawn('npm', ['install']);  // Fails on Windows
```

✅ **Right:**
```javascript
const spawn = require('cross-spawn');
spawn('npm', ['install']);  // Works everywhere
```

---

## 11. Conclusion

Cross-platform compatibility is achievable by following established patterns:

1. **Use Node.js APIs** instead of shell commands
2. **Use the `path` module** for all file paths
3. **Use cross-platform npm packages** (shx, cross-env, rimraf, mkdirp)
4. **Test on all target platforms** using GitHub Actions matrix builds
5. **Be aware of case sensitivity** differences
6. **Handle file system differences** appropriately (permissions, symlinks)
7. **Use proper shebangs** and let npm create platform wrappers

The notion-cli project is already on the right track with oclif and shx. The main areas for improvement are:
- Expanding CI/CD testing to all platforms
- Enforcing case-sensitive imports
- Reviewing npm scripts for platform compatibility
- Adding comprehensive cross-platform tests

By following these guidelines, the notion-cli package can reliably work across Windows, macOS, and Linux, providing a consistent experience for all users.

---

**Document Version:** 1.0
**Last Updated:** 2025-10-22
**Maintained By:** Coastal Programs
**Next Review:** Before each major release
