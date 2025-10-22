# Research: AI-Agent Friendly CLI Design & Installation

**Research Date:** 2025-10-22
**Subject:** Best practices for designing CLIs that work seamlessly for both human users and AI agents

---

## Executive Summary

AI agents are increasingly becoming primary users of command-line tools. This research explores how to design CLIs that are both human-friendly and automation-ready, with emphasis on non-interactive installation, robust configuration management, and credential handling that works programmatically.

**Key Finding:** The most successful AI-friendly CLIs follow these principles:
1. **Zero-interaction capability** - All features accessible without prompts
2. **Predictable configuration precedence** - Command line > Environment > Config file > Defaults
3. **Structured output** - JSON/YAML output modes for parsing
4. **Graceful degradation** - Interactive prompts that can be bypassed
5. **Clear error handling** - Machine-readable error codes and messages

---

## 1. Examples of AI-Agent Friendly CLIs

### 1.1 Model Context Protocol (MCP)

**What it is:** An open-source standard for connecting AI assistants to data sources and tools, created by Anthropic.

**Key Design Patterns:**
- **Client-Server Architecture:** MCP hosts (like Claude Desktop) connect to multiple MCP servers via a standardized protocol
- **Dynamic Tool Discovery:** AI agents automatically detect available MCP servers and their capabilities without custom coding
- **CLI Integration:** For local subprocesses, MCPServerStdio spawns processes and manages pipes automatically
- **Configuration Format:**
  ```json
  {
    "mcpServers": {
      "example": {
        "command": "node",
        "args": ["/path/to/server/index.js"],
        "env": {
          "API_KEY": "secret_xxx"
        }
      }
    }
  }
  ```

**Why it works for AI agents:**
- Standardized interface eliminates custom integration code
- Tools expose capabilities through discoverable schemas
- No manual configuration needed for tool discovery
- Automatic connection management and error handling

**Implementation Status:** Adopted by Block, Apollo, Zed, Replit, Codeium, and Sourcegraph. Available in TypeScript/Node.js, Python, and .NET.

### 1.2 Claude Code CLI

**What it is:** Anthropic's official CLI for Claude AI, designed as both MCP server and client.

**Configuration Approach:**
- **Hierarchical Context:** Global `~/.claude/CLAUDE.md` and project-specific `.claude/` directories
- **Slash Commands:** Markdown templates in `.claude/commands/` become reusable workflows
- **Environment Variables:** `ANTHROPIC_MODEL` for permanent defaults, `--model` flag for session overrides
- **Settings File:** `~/.claude/claude.json` for persistent configuration
- **MCP Integration:** Connects via `.mcp.json` files, shareable across teams

**Key Features:**
```bash
# Model configuration (3 ways)
/model                    # Interactive change
--model claude-3-opus     # One-time override
ANTHROPIC_MODEL=opus      # Permanent default in .zshrc

# Debug and control
ANTHROPIC_LOG=debug              # Debug logging
CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1  # Disable telemetry
BASH_DEFAULT_TIMEOUT_MS=120000   # Timeout settings
```

**Why it works:**
- Multiple configuration methods for different contexts
- Project-level commands can be version-controlled
- Clear precedence: flags > env vars > config files
- Programmatic tool access via MCP protocol

### 1.3 Heroku CLI

**Built with:** oclif framework (open-sourced by Heroku)

**Design Philosophy:**
- "CLI-first devtools" - reaching developers where they already are
- Show value with minimum commands: install → signup → create
- Configuration lives in developer-friendly, repeatable environment
- Direct code examples in emails and documentation

**Onboarding Excellence:**
- Almost perfect first integration experience
- Emails contain CLI commands directly (not just web UI instructions)
- Consistent input/output across all commands
- Follows CLI Style Guide principles

**Example of simple workflow:**
```bash
# Install via package manager
brew tap heroku/brew && brew install heroku

# Login
heroku login

# Create app (single command)
heroku create my-app
```

### 1.4 GitHub CLI (`gh`)

**Authentication Flow:**
- OAuth2 Device Authorization Flow
- Automatically opens browser for auth
- Falls back to manual code entry for remote/headless environments
- Streamlines developer onboarding significantly

**Key Feature - Simplified Onboarding:**
```bash
# Instead of complex SSH setup with 2FA:
gh auth login && gh repo clone xyz/project

# Everything works immediately
```

**Why it works:**
- Reduces friction in team onboarding
- Browser-based flow familiar to users
- Fallback for non-interactive environments
- Single command handles auth complexity

### 1.5 AWS CLI

**Configuration Command:**
```bash
$ aws configure
AWS Access Key ID [None]: AKIAIOSFODNN7EXAMPLE
AWS Secret Access Key [None]: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
Default region name [None]: us-west-2
Default output format [None]: json
```

**Advanced Features:**
- `aws configure sso` for SSO wizard
- `aws configure wizard` for custom interactive setup
- Browser-based OAuth for SSO authentication
- `--profile` flag for multiple account management

**Non-Interactive Usage:**
```bash
# Environment variables
export AWS_ACCESS_KEY_ID=xxx
export AWS_SECRET_ACCESS_KEY=xxx
export AWS_DEFAULT_REGION=us-west-2

# Or config files
~/.aws/credentials
~/.aws/config
```

### 1.6 Popular AI Agent Frameworks (CLI-Compatible)

| Framework | Language | Key CLI Features | Monthly Downloads |
|-----------|----------|------------------|-------------------|
| **AutoGen** | Python | Multi-agent orchestration, conversation-based coordination | - |
| **CrewAI** | Python | Role-playing agents, minimal setup | 1M+ |
| **LangGraph** | Python | Stateful agents, maintains context | 4.2M |
| **n8n** | TypeScript/Node.js | Workflow automation, visual + CLI | Active |

**Common Patterns:**
- Task decomposition with specialized sub-agents
- Living spec files with requirements and conventions
- Four-step workflow: assign → plan → iterate → execute
- Built-in policy enforcement and audit trails

---

## 2. Non-Interactive Installation & Configuration

### 2.1 Core Principles

**Silent/Non-Interactive Installation:**
- Supply installer with response file containing all configuration parameters
- No user interaction required during installation process
- Essential for CI/CD pipelines and automated deployments
- Enables consistent deployments across environments

### 2.2 Common Implementation Methods

#### Response Files
Pre-configured files that answer all installer prompts:

```bash
# Example response file structure
INSTALLATION_PATH=/opt/mycli
ENABLE_TELEMETRY=false
API_ENDPOINT=https://api.example.com
```

#### Command-Line Flags
```bash
# Common flags
--passive              # Non-interactive mode
--accept-eulas         # Accept license agreements
--prevent-reboot       # Don't restart during install
--silent               # Completely silent
--config-file=/path    # Use specific config file
```

#### API Key Authentication
```bash
# For headless/CI environments
mycli --api-key=xxx init --non-interactive
```

#### Environment-Based Setup
```bash
# All config via environment
export MYCLI_API_KEY=xxx
export MYCLI_WORKSPACE=default
mycli init --no-interaction
```

### 2.3 First-Run Detection Patterns

**Configuration File Detection:**
```typescript
// Check if config exists
const configPath = path.join(os.homedir(), '.mycli', 'config.json')
const isFirstRun = !fs.existsSync(configPath)

if (isFirstRun) {
  // Run setup wizard OR use defaults
  if (process.stdout.isTTY && !process.env.CI) {
    // Interactive mode
    await runSetupWizard()
  } else {
    // Non-interactive mode - use defaults/env vars
    await initializeDefaults()
  }
}
```

**.NET CLI Pattern:**
```bash
# Skip first-time experience
export DOTNET_SKIP_FIRST_TIME_EXPERIENCE=true
```

**Tool Manifest Detection:**
```bash
# .NET tools look for dotnet-tools.json
# If not found, treat as first run
```

### 2.4 Package Manager Post-Install Scripts

**npm Lifecycle Hooks:**
- `preinstall` - Before package installation
- `install` - During package installation
- `postinstall` - After package installation

**Limitations:**
```javascript
{
  "scripts": {
    // Only runs on `npm install`, not `npm i <package>`
    "postinstall": "node scripts/setup.js"
  }
}
```

**Best Practices:**
- Don't rely solely on postinstall for critical setup
- Detect first run on first command execution instead
- Use postinstall only for non-essential initialization
- Consider security implications (can be disabled with `--ignore-scripts`)

**Debugging:**
```bash
npm install --loglevel=verbose    # See all script execution
npm install --foreground-scripts  # Show script output
```

### 2.5 oclif Framework Best Practices

**Configuration Options:**
- Configure in `package.json` under `oclif` section
- Or use rc files: `.oclifrc`, `.oclifrc.json`, `.oclifrc.js`, etc.
- Support for both CommonJS and ESM

**Development Workflow:**
```bash
# Use dev scripts for auto-transpile
npm run dev

# No need to rebuild between changes
# TypeScript transpiled at runtime
```

**Convention Over Configuration:**
- Commands auto-discovered in `./commands/` folder
- Subcommands = subfolders
- Plugin system for extensibility
- Everything swappable if needed

---

## 3. Automated API Key & Credential Management

### 3.1 Configuration Precedence (Industry Standard)

**Standard Order (Lowest to Highest Priority):**

1. **Default Values** (hardcoded)
2. **Global Config File** (`/etc/mycli/config.json`)
3. **User Config File** (`~/.mycli/config.json`)
4. **Local Config File** (`./mycli.config.json`)
5. **Environment Variables** (`MYCLI_API_KEY`)
6. **Command-Line Arguments** (`--api-key=xxx`)

**POSIX Standard:** "configuration → environment → command line"

**Example Implementation:**
```typescript
function getConfig(key: string): string {
  // 1. Check command-line args
  if (flags[key]) return flags[key]

  // 2. Check environment variables
  const envKey = `MYCLI_${key.toUpperCase()}`
  if (process.env[envKey]) return process.env[envKey]

  // 3. Check local config
  const localConfig = loadConfig('./.mycli.json')
  if (localConfig[key]) return localConfig[key]

  // 4. Check user config
  const userConfig = loadConfig('~/.mycli/config.json')
  if (userConfig[key]) return userConfig[key]

  // 5. Check global config
  const globalConfig = loadConfig('/etc/mycli/config.json')
  if (globalConfig[key]) return globalConfig[key]

  // 6. Use default
  return DEFAULTS[key]
}
```

### 3.2 Environment Variable Detection

**Best Practices:**

**Naming Conventions:**
```bash
# Prefix with tool name
export NOTION_TOKEN=secret_xxx
export NOTION_CLI_MAX_RETRIES=5

# Use consistent separators
MYCLI_API_KEY          # Underscore-separated
MYCLI_CACHE_ENABLED    # All uppercase
```

**Validation at Startup:**
```typescript
// Required variables
const requiredEnvVars = ['NOTION_TOKEN']

for (const varName of requiredEnvVars) {
  if (!process.env[varName]) {
    console.error(`Error: ${varName} environment variable is required`)
    console.error(`Set it with: export ${varName}=your_value`)
    process.exit(1)
  }
}
```

**Using dotenv with Validation:**
```typescript
import * as envalid from 'envalid'

const env = envalid.cleanEnv(process.env, {
  NOTION_TOKEN: envalid.str(),
  MAX_RETRIES: envalid.num({ default: 3 }),
  CACHE_ENABLED: envalid.bool({ default: true })
})

// Throws error if validation fails
// App won't start with invalid config
```

**12-Factor App Principle:**
- Store config in environment variables
- Strict separation of config from code
- Easy to change between deploys without code changes
- Language and OS-agnostic standard

### 3.3 Interactive Prompts with Defaults

**Design Pattern:**
```typescript
import * as prompts from 'prompts'

async function setupWizard() {
  // Check if in non-interactive environment
  if (!process.stdout.isTTY || process.env.CI) {
    return useDefaults()
  }

  const response = await prompts([
    {
      type: 'text',
      name: 'apiKey',
      message: 'Enter your API key:',
      initial: process.env.NOTION_TOKEN || '',  // Pre-fill from env
      validate: value => value.length > 0 ? true : 'API key is required'
    },
    {
      type: 'number',
      name: 'maxRetries',
      message: 'Max retry attempts:',
      initial: 3,
      min: 1,
      max: 10
    },
    {
      type: 'confirm',
      name: 'enableCache',
      message: 'Enable caching?',
      initial: true
    }
  ])

  return response
}
```

**Key Features:**
- `initial` parameter provides default value
- Defaults can come from environment variables
- Validation ensures data quality
- Can be piped answers: `echo "y" | mycli setup`

**Automated Prompt Handling:**
```bash
# Using expect for automation
expect << EOF
spawn mycli setup
expect "Enter API key:"
send "secret_xxx\r"
expect "Max retries:"
send "5\r"
expect eof
EOF
```

### 3.4 Configuration File Auto-Generation

**Pattern: Generate on First Run**

```typescript
import * as fs from 'fs'
import * as path from 'path'
import * as os from 'os'

interface Config {
  apiKey?: string
  maxRetries: number
  cacheEnabled: boolean
  workspace: string
}

const CONFIG_DIR = path.join(os.homedir(), '.mycli')
const CONFIG_FILE = path.join(CONFIG_DIR, 'config.json')

async function initializeConfig(): Promise<Config> {
  // Create config directory if needed
  if (!fs.existsSync(CONFIG_DIR)) {
    fs.mkdirSync(CONFIG_DIR, { recursive: true })
  }

  // Check if config exists
  if (fs.existsSync(CONFIG_FILE)) {
    return JSON.parse(fs.readFileSync(CONFIG_FILE, 'utf-8'))
  }

  // Generate new config
  const config: Config = {
    apiKey: process.env.NOTION_TOKEN,
    maxRetries: 3,
    cacheEnabled: true,
    workspace: 'default'
  }

  // Write config file
  fs.writeFileSync(CONFIG_FILE, JSON.stringify(config, null, 2))
  console.log(`Created config file: ${CONFIG_FILE}`)

  return config
}
```

**Template-Based Generation:**
```typescript
async function generateConfigTemplate() {
  const template = `{
  // API Configuration
  "apiKey": "secret_xxx",  // Or set NOTION_TOKEN env var

  // Performance Settings
  "maxRetries": 3,
  "cacheEnabled": true,
  "cacheTtl": 600000,  // 10 minutes

  // Workspace
  "workspace": "default"
}
`

  const templatePath = path.join(CONFIG_DIR, 'config.template.json')
  fs.writeFileSync(templatePath, template)

  console.log(`Config template created at: ${templatePath}`)
  console.log(`Copy to config.json and fill in your values`)
}
```

**XDG Base Directory Specification:**

For cross-platform compatibility:

```typescript
function getConfigPath(): string {
  // Linux/Unix: Follow XDG spec
  if (process.platform !== 'win32') {
    const xdgConfig = process.env.XDG_CONFIG_HOME ||
                      path.join(os.homedir(), '.config')
    return path.join(xdgConfig, 'mycli', 'config.json')
  }

  // macOS: Use standard location
  if (process.platform === 'darwin') {
    return path.join(os.homedir(), 'Library', 'Preferences', 'mycli', 'config.json')
  }

  // Windows: Use APPDATA
  const appData = process.env.APPDATA ||
                  path.join(os.homedir(), 'AppData', 'Roaming')
  return path.join(appData, 'mycli', 'config.json')
}
```

### 3.5 Credential Storage Best Practices

**Security Principle: Never Hardcode, Never Plaintext**

#### Environment Variables (Basic Security)

**Pros:**
- Simple to implement
- Widely supported
- Easy to change per environment
- No files to manage

**Cons:**
- Accessible to anyone with server access
- Visible in process lists
- Not encrypted at rest
- Can leak in logs/error messages

**Best Practice:**
```bash
# Good: Per-environment
export NOTION_TOKEN_DEV=secret_dev_xxx
export NOTION_TOKEN_PROD=secret_prod_xxx

# Good: Per-developer
# Each dev uses their own key, not shared prod key

# Bad: Hardcoded in scripts
API_KEY="secret_xxx"  # Never do this
```

#### Operating System Keychains (Better Security)

**Cross-Platform Solutions:**

**1. Git Credential Manager (Microsoft)**
- Windows: DPAPI encrypted files
- macOS: Keychain
- Linux: Secret Service API
- Used by GitHub, Azure, GitLab

**2. Python pycreds**
```python
from pycreds import Keyring

# Store credential
keyring = Keyring()
keyring.set('mycli', 'api_key', 'secret_xxx')

# Retrieve credential
api_key = keyring.get('mycli', 'api_key')

# CLI interface
pycreds set mycli api_key secret_xxx
pycreds get mycli api_key
```

**3. Node.js keytar**
```typescript
import * as keytar from 'keytar'

// Store credential
await keytar.setPassword('mycli', 'api_key', 'secret_xxx')

// Retrieve credential
const apiKey = await keytar.getPassword('mycli', 'api_key')

// Delete credential
await keytar.deletePassword('mycli', 'api_key')
```

**4. MCP Secrets Plugin**
- Leverages OS-native secure storage
- macOS: Keychain
- Windows: Credential Locker
- Linux: Keyring backends
- CLI for direct secret management

#### Azure Key Vault / HashiCorp Vault (Enterprise)

**For Production Systems:**

```typescript
import { SecretClient } from '@azure/keyvault-secrets'
import { DefaultAzureCredential } from '@azure/identity'

const vaultUrl = 'https://my-vault.vault.azure.net'
const credential = new DefaultAzureCredential()
const client = new SecretClient(vaultUrl, credential)

// Retrieve secret
const secret = await client.getSecret('notion-api-key')
const apiKey = secret.value
```

**Benefits:**
- Centralized secret management
- Audit logs of access
- Automatic rotation
- Fine-grained access control
- Compliance-ready (HIPAA, SOX)

#### .env File with .gitignore (Development Only)

**Structure:**
```bash
# .env
NOTION_TOKEN=secret_xxx
MAX_RETRIES=5
CACHE_ENABLED=true
```

```bash
# .gitignore
.env
.env.local
*.key
*.pem
credentials.json
```

**Loading with Validation:**
```typescript
import * as dotenv from 'dotenv'
import * as envalid from 'envalid'

// Load .env file
dotenv.config()

// Validate required variables
const env = envalid.cleanEnv(process.env, {
  NOTION_TOKEN: envalid.str({
    desc: 'Notion API token',
    example: 'secret_xxx',
  }),
  MAX_RETRIES: envalid.num({ default: 3 }),
})
```

**Warning:** .env files are NOT secure for production. They're plain text and can be accidentally committed. Use only for local development.

#### Summary: Which Method When?

| Environment | Recommended Method | Fallback |
|-------------|-------------------|----------|
| **Local Development** | .env file + .gitignore | Environment variables |
| **CI/CD Pipeline** | Repository secrets / Encrypted secrets | Environment variables |
| **Production Server** | Key Vault / Secrets Manager | OS Keychain |
| **Desktop Application** | OS Keychain (keytar/pycreds) | Encrypted config file |
| **CLI Tool (Personal Use)** | OS Keychain | ~/.config with restricted permissions |
| **CLI Tool (Automation)** | Environment variables | Config file with secrets |

---

## 4. Setup Wizard Design for Humans & AI Agents

### 4.1 Dual-Mode Design Pattern

**Core Principle:** Detect the execution environment and adapt behavior accordingly.

```typescript
interface SetupOptions {
  interactive?: boolean
  apiKey?: string
  workspace?: string
  useDefaults?: boolean
}

async function setup(options: SetupOptions = {}) {
  // Detect environment
  const isInteractive = options.interactive ??
                       (process.stdout.isTTY && !process.env.CI)

  if (isInteractive && !options.useDefaults) {
    return await interactiveSetup(options)
  } else {
    return await nonInteractiveSetup(options)
  }
}
```

**Environment Detection:**
```typescript
function getExecutionContext() {
  return {
    isTTY: process.stdout.isTTY,           // Terminal available?
    isCI: !!process.env.CI,                // CI environment?
    hasDisplay: !!process.env.DISPLAY,     // GUI available?
    isDocker: fs.existsSync('/.dockerenv'), // Running in Docker?
    isPipe: !process.stdin.isTTY,          // Input piped?
  }
}
```

### 4.2 Interactive Mode (Human Users)

**Rich Terminal UI:**
```typescript
import * as prompts from 'prompts'
import chalk from 'chalk'
import ora from 'ora'

async function interactiveSetup(options: SetupOptions) {
  console.log(chalk.blue.bold('\n🚀 Welcome to MyCLI Setup\n'))

  // Multi-step wizard
  const answers = await prompts([
    {
      type: 'text',
      name: 'apiKey',
      message: 'Enter your API key:',
      initial: options.apiKey || process.env.MYCLI_API_KEY,
      validate: (value) => {
        if (!value) return 'API key is required'
        if (!value.startsWith('secret_')) {
          return 'API key should start with "secret_"'
        }
        return true
      }
    },
    {
      type: 'select',
      name: 'workspace',
      message: 'Select workspace:',
      choices: [
        { title: 'Default', value: 'default' },
        { title: 'Production', value: 'prod' },
        { title: 'Development', value: 'dev' },
      ],
      initial: 0
    },
    {
      type: 'number',
      name: 'maxRetries',
      message: 'Max retry attempts:',
      initial: 3,
      min: 1,
      max: 10
    },
    {
      type: 'confirm',
      name: 'enableCache',
      message: 'Enable caching for better performance?',
      initial: true
    },
    {
      type: 'confirm',
      name: 'enableTelemetry',
      message: 'Send anonymous usage statistics?',
      initial: false
    }
  ])

  // Cancelled by user
  if (!answers.apiKey) {
    console.log(chalk.yellow('\nSetup cancelled'))
    process.exit(0)
  }

  // Verify API key works
  const spinner = ora('Verifying API key...').start()

  try {
    await verifyApiKey(answers.apiKey)
    spinner.succeed('API key verified')
  } catch (error) {
    spinner.fail('API key verification failed')
    console.error(chalk.red(error.message))
    process.exit(1)
  }

  // Save configuration
  await saveConfig({
    apiKey: answers.apiKey,
    workspace: answers.workspace,
    maxRetries: answers.maxRetries,
    cacheEnabled: answers.enableCache,
    telemetryEnabled: answers.enableTelemetry,
  })

  console.log(chalk.green('\n✓ Setup complete!\n'))
  console.log(`Configuration saved to: ${getConfigPath()}`)
  console.log(`\nTry running: ${chalk.cyan('mycli --help')}`)
}
```

### 4.3 Non-Interactive Mode (AI Agents & Automation)

**Silent Configuration:**
```typescript
async function nonInteractiveSetup(options: SetupOptions) {
  // Collect config from all sources
  const config = {
    apiKey: options.apiKey ||
            process.env.MYCLI_API_KEY ||
            undefined,
    workspace: options.workspace ||
               process.env.MYCLI_WORKSPACE ||
               'default',
    maxRetries: Number(process.env.MYCLI_MAX_RETRIES) || 3,
    cacheEnabled: process.env.MYCLI_CACHE_ENABLED !== 'false',
    telemetryEnabled: false,  // Default to false for automation
  }

  // Validate required fields
  if (!config.apiKey) {
    throw new Error(
      'API key is required. Set MYCLI_API_KEY environment variable ' +
      'or use --api-key flag'
    )
  }

  // Verify API key (with retry for transient failures)
  await verifyApiKeyWithRetry(config.apiKey)

  // Save configuration
  await saveConfig(config)

  // Silent success (or JSON output if --json flag)
  if (options.json) {
    console.log(JSON.stringify({
      success: true,
      configPath: getConfigPath(),
      workspace: config.workspace,
    }))
  }

  return config
}
```

**Flag-Based Configuration:**
```bash
# Complete non-interactive setup
mycli setup \
  --api-key=secret_xxx \
  --workspace=prod \
  --max-retries=5 \
  --cache=true \
  --no-telemetry \
  --json

# Or via environment variables
export MYCLI_API_KEY=secret_xxx
export MYCLI_WORKSPACE=prod
mycli setup --non-interactive --json
```

### 4.4 Hybrid Approach: Appwrite CLI Example

**Smart Defaults with Override Capability:**

From Appwrite documentation:
- Use `--non-interactive` flag to skip all prompts
- Required values must be provided via flags or config file
- Optional values use sensible defaults
- Can mix interactive and non-interactive steps

```bash
# Fully non-interactive
appwrite deploy function \
  --functionId=123 \
  --non-interactive

# Interactive for missing values only
appwrite deploy function
# Prompts for functionId if not provided
```

### 4.5 Configuration String Pattern

**Check Point Security Gateway Example:**

```bash
# Single-line configuration for automation
config_system --config-string "hostname=myhost&domainname=example.com&timezone=America/New_York&install_security_gw=true&ftw_sic_key=secretkey"

# Or from file
config_system --config-file /path/to/config.txt

# Or interactive
config_system  # Prompts for each value

# Validate without applying
config_system --config-file /path/to/config.txt --dry-run

# List available parameters
config_system --list-params
```

**Benefits:**
- Single command for complete setup
- Easy to template and generate
- Validation before application
- Discovery of available options

### 4.6 First-Run Wizard Implementation

**Automatic First-Run Detection:**
```typescript
export class SetupCommand extends Command {
  async run() {
    const config = await loadConfig()

    if (!config.apiKey) {
      console.log(chalk.yellow('⚠ First time setup required\n'))

      // Interactive if possible
      if (process.stdout.isTTY && !process.env.CI) {
        return await interactiveSetup()
      } else {
        // Non-interactive error with instructions
        console.error(chalk.red('Error: API key not configured\n'))
        console.error('For interactive setup:')
        console.error('  mycli setup\n')
        console.error('For automated setup:')
        console.error('  export MYCLI_API_KEY=secret_xxx')
        console.error('  mycli setup --non-interactive\n')
        process.exit(1)
      }
    }

    // Already configured
    console.log(chalk.green('✓ Already configured'))
    console.log(`API Key: ${config.apiKey.substring(0, 10)}...`)
    console.log(`Workspace: ${config.workspace}`)
  }
}
```

**Re-run Setup:**
```bash
# Force re-run setup
mycli setup --force

# Reset to defaults
mycli setup --reset

# Update specific values
mycli setup --workspace=prod
```

---

## 5. Excellent CLI Onboarding Examples

### 5.1 Heroku CLI - "Almost Perfect"

**What Makes It Great:**

1. **Developer-Centric Emails:**
   - Contains actual CLI commands, not just web UI instructions
   - Can copy-paste directly from email to terminal
   - Shows expected output

2. **Minimal Command Count:**
   - Install (package manager)
   - Login (automatic browser auth)
   - Create (one command to deploy)

3. **Consistent UX:**
   - All commands follow same patterns
   - Predictable flag names
   - Clear error messages with remediation steps

4. **Built on oclif:**
   - Extensible plugin architecture
   - Auto-discovered commands
   - TypeScript with auto-transpilation

**Example Flow:**
```bash
# Step 1: Install (via package manager)
brew tap heroku/brew && brew install heroku

# Step 2: Login (opens browser)
heroku login

# Step 3: Create and deploy
git init
git add .
git commit -m "Initial commit"
heroku create
git push heroku main
# App is live!
```

### 5.2 AWS CLI - Comprehensive Configuration

**Interactive Configuration:**
```bash
$ aws configure
AWS Access Key ID [None]: AKIAIOSFODNN7EXAMPLE
AWS Secret Access Key [None]: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
Default region name [None]: us-west-2
Default output format [None]: json
```

**Advanced Setup:**
```bash
# SSO setup with browser
$ aws configure sso
SSO session name: my-dev
SSO start URL: https://my-sso.awsapps.com/start
SSO region: us-east-1
# Opens browser automatically for authentication
```

**Multiple Profiles:**
```bash
# Configure named profiles
aws configure --profile production
aws configure --profile development

# Use specific profile
aws s3 ls --profile production

# Set default profile
export AWS_PROFILE=production
```

**What Makes It Great:**
- Progressive disclosure (simple first, advanced when needed)
- Multiple authentication methods (keys, SSO, IAM roles)
- Clear separation of environments via profiles
- Automatic browser integration for OAuth flows

### 5.3 GitHub CLI - Simplified Developer Onboarding

**Authentication Excellence:**
```bash
$ gh auth login

? What account do you want to log into? GitHub.com
? What is your preferred protocol for Git operations? SSH
? Generate a new SSH key to add to your GitHub account? Yes
? Enter a passphrase for your new SSH key (Optional) ****
? Title for your SSH key: GitHub CLI
? How would you like to authenticate GitHub CLI? Login with a web browser

! First copy your one-time code: ABCD-1234
Press Enter to open github.com in your browser...
✓ Authentication complete.
- gh config set -h github.com git_protocol ssh
✓ Logged in as username
```

**What Makes It Great:**
- Handles SSH key generation automatically
- One-time code for security
- Browser-based auth (familiar to users)
- Immediate verification
- Streamlines what used to be complex 2FA + SSH setup

**After Setup:**
```bash
# Everything just works
gh repo clone username/repo
gh issue create
gh pr create
```

### 5.4 Google Cloud CLI - Flexible Authentication

**Initial Setup:**
```bash
$ gcloud auth login
# Opens browser for authentication

$ gcloud config set project my-project
$ gcloud config set compute/region us-central1
```

**Console-Only Mode (for remote/SSH):**
```bash
$ gcloud auth login --console-only
Go to the following link in your browser:
    https://accounts.google.com/o/oauth2/auth?...

Enter verification code: 4/xxxxxxxxxxxxx
```

**What Makes It Great:**
- Browser-based primary flow
- Console-only fallback for remote systems
- Clear instructions for each step
- Handles all OAuth complexity internally

### 5.5 Stripe CLI - Developer-First Design

**Quick Start:**
```bash
# Install
brew install stripe/stripe-cli/stripe

# Login
stripe login
# Opens browser for authentication

# Start local webhook testing
stripe listen --forward-to localhost:4242/webhook
```

**What Makes It Great:**
- Minimal steps to value
- Local webhook testing built-in
- Real-time event streaming
- Sample event trigger commands
- Test mode by default (safe experimentation)

### 5.6 .NET CLI - Smart First-Run Detection

**Environment Variable Approach:**
```bash
# Skip first-time experience in CI/automation
export DOTNET_SKIP_FIRST_TIME_EXPERIENCE=true

# Normal first run shows:
# "Welcome to .NET! Learn more at: https://aka.ms/dotnet-get-started"
# Then proceeds with command
```

**Tool Manifest Detection:**
```bash
# Looks for dotnet-tools.json
# If not found, treats as first run

# Initialize tools manifest
dotnet new tool-manifest

# Install tools
dotnet tool install --local <package>
```

**What Makes It Great:**
- Automatic first-run detection
- Can be disabled for automation
- Educates new users without blocking
- Manifest-based tool management

---

## 6. Implementation Recommendations for Notion CLI

### 6.1 Current State Analysis

**Strengths:**
- Already optimized for automation (non-interactive by design)
- JSON output mode for AI agents (`--json` flag)
- Environment variable configuration (`NOTION_TOKEN`)
- Exit codes for automation (0 = success, 1 = error)
- Clear error messages in JSON format
- oclif framework (proven, extensible)

**Gaps:**
- No setup wizard or first-run experience
- No configuration file support (only env vars)
- No credential management (relies on environment only)
- No way to detect misconfiguration until first command fails
- No guided onboarding for new users

### 6.2 Recommended Enhancements

#### Phase 1: Configuration Management

**Add Configuration File Support:**
```typescript
// src/config.ts
import * as fs from 'fs'
import * as path from 'path'
import * as os from 'os'

interface NotionCliConfig {
  apiKey?: string
  defaultWorkspace?: string
  cacheEnabled?: boolean
  cacheTtl?: number
  maxRetries?: number
  logLevel?: 'debug' | 'info' | 'warn' | 'error'
  outputFormat?: 'json' | 'table' | 'csv' | 'yaml'
}

function getConfigPath(): string {
  // Follow XDG spec on Linux, standard paths on other platforms
  if (process.platform !== 'win32') {
    const xdgConfig = process.env.XDG_CONFIG_HOME ||
                      path.join(os.homedir(), '.config')
    return path.join(xdgConfig, 'notion-cli', 'config.json')
  }

  return path.join(os.homedir(), '.notion-cli', 'config.json')
}

export function loadConfig(): NotionCliConfig {
  const configPath = getConfigPath()

  if (!fs.existsSync(configPath)) {
    return {}
  }

  try {
    return JSON.parse(fs.readFileSync(configPath, 'utf-8'))
  } catch (error) {
    console.warn(`Warning: Failed to load config from ${configPath}`)
    return {}
  }
}

export function getApiKey(): string {
  // Precedence: CLI flag > env var > config file

  // 1. Environment variable (current method)
  if (process.env.NOTION_TOKEN) {
    return process.env.NOTION_TOKEN
  }

  // 2. Config file
  const config = loadConfig()
  if (config.apiKey) {
    return config.apiKey
  }

  // 3. Not found - provide helpful error
  throw new Error(
    'Notion API token not found. Set it with:\n' +
    '  export NOTION_TOKEN=secret_xxx\n' +
    'Or run: notion-cli setup'
  )
}
```

**Configuration Precedence:**
1. Command-line flags (highest priority)
2. Environment variables
3. User config file (`~/.notion-cli/config.json`)
4. Defaults (hardcoded)

#### Phase 2: Setup Command

**Add `notion-cli setup` command:**
```typescript
// src/commands/setup.ts
import { Command, Flags } from '@oclif/core'
import * as prompts from 'prompts'
import chalk from 'chalk'
import ora from 'ora'

export default class Setup extends Command {
  static description = 'Configure Notion CLI'

  static flags = {
    'api-key': Flags.string({
      description: 'Notion API token',
      env: 'NOTION_TOKEN'
    }),
    'non-interactive': Flags.boolean({
      description: 'Skip interactive prompts',
      default: false
    }),
    'force': Flags.boolean({
      description: 'Overwrite existing configuration',
      default: false
    }),
    'json': Flags.boolean({
      description: 'Output result as JSON',
      char: 'j',
      default: false
    })
  }

  async run() {
    const { flags } = await this.parse(Setup)

    // Check for existing config
    const existingConfig = loadConfig()
    if (existingConfig.apiKey && !flags.force) {
      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          alreadyConfigured: true,
          configPath: getConfigPath()
        }))
      } else {
        this.log(chalk.green('✓ Notion CLI is already configured'))
        this.log(`\nConfig file: ${getConfigPath()}`)
        this.log(`API Key: ${existingConfig.apiKey.substring(0, 15)}...`)
        this.log(`\nTo reconfigure, run: ${chalk.cyan('notion-cli setup --force')}`)
      }
      return
    }

    // Interactive or non-interactive?
    if (flags['non-interactive'] || !process.stdout.isTTY || process.env.CI) {
      return await this.nonInteractiveSetup(flags)
    } else {
      return await this.interactiveSetup(flags)
    }
  }

  async interactiveSetup(flags: any) {
    this.log(chalk.blue.bold('\n🚀 Notion CLI Setup\n'))

    const answers = await prompts([
      {
        type: 'text',
        name: 'apiKey',
        message: 'Enter your Notion API token:',
        initial: flags['api-key'] || process.env.NOTION_TOKEN,
        validate: (value) => {
          if (!value) return 'API token is required'
          if (!value.startsWith('secret_')) {
            return 'API token should start with "secret_"'
          }
          return true
        }
      },
      {
        type: 'confirm',
        name: 'enableCache',
        message: 'Enable caching for better performance?',
        initial: true
      },
      {
        type: 'number',
        name: 'maxRetries',
        message: 'Max retry attempts for failed requests:',
        initial: 3,
        min: 1,
        max: 10
      },
      {
        type: 'select',
        name: 'outputFormat',
        message: 'Default output format:',
        choices: [
          { title: 'JSON (recommended for automation)', value: 'json' },
          { title: 'Table (human-readable)', value: 'table' },
          { title: 'CSV', value: 'csv' },
          { title: 'YAML', value: 'yaml' }
        ],
        initial: 0
      }
    ])

    if (!answers.apiKey) {
      this.log(chalk.yellow('\nSetup cancelled'))
      return
    }

    // Verify API token
    const spinner = ora('Verifying API token...').start()

    try {
      await this.verifyApiToken(answers.apiKey)
      spinner.succeed('API token verified')
    } catch (error) {
      spinner.fail('API token verification failed')
      this.error(error.message, { exit: 1 })
    }

    // Save configuration
    const config = {
      apiKey: answers.apiKey,
      cacheEnabled: answers.enableCache,
      maxRetries: answers.maxRetries,
      outputFormat: answers.outputFormat,
      version: this.config.version
    }

    saveConfig(config)

    this.log(chalk.green('\n✓ Setup complete!\n'))
    this.log(`Configuration saved to: ${getConfigPath()}`)
    this.log(`\nGet started with: ${chalk.cyan('notion-cli --help')}`)
    this.log(`\nExample command: ${chalk.cyan('notion-cli page retrieve <PAGE_ID>')}`)
  }

  async nonInteractiveSetup(flags: any) {
    const apiKey = flags['api-key'] || process.env.NOTION_TOKEN

    if (!apiKey) {
      if (flags.json) {
        this.log(JSON.stringify({
          success: false,
          error: 'API token required',
          hint: 'Set NOTION_TOKEN environment variable or use --api-key flag'
        }))
      } else {
        this.error(
          'API token required. Set NOTION_TOKEN environment variable ' +
          'or use --api-key flag',
          { exit: 1 }
        )
      }
      return
    }

    // Verify token
    try {
      await this.verifyApiToken(apiKey)
    } catch (error) {
      if (flags.json) {
        this.log(JSON.stringify({
          success: false,
          error: 'Invalid API token',
          details: error.message
        }))
      }
      this.error(error.message, { exit: 1 })
    }

    // Save config
    const config = {
      apiKey,
      cacheEnabled: true,
      maxRetries: 3,
      outputFormat: 'json',
      version: this.config.version
    }

    saveConfig(config)

    if (flags.json) {
      this.log(JSON.stringify({
        success: true,
        configPath: getConfigPath(),
        version: this.config.version
      }))
    }
  }

  async verifyApiToken(token: string): Promise<void> {
    // Use Notion client to verify token works
    const { Client } = require('@notionhq/client')
    const notion = new Client({ auth: token })

    try {
      // Try to get bot user as verification
      await notion.users.me()
    } catch (error) {
      throw new Error(`API token verification failed: ${error.message}`)
    }
  }
}
```

#### Phase 3: First-Run Detection

**Auto-detect first run and offer setup:**
```typescript
// src/hooks/init/check-config.ts
import { Hook } from '@oclif/core'
import chalk from 'chalk'

const hook: Hook<'init'> = async function (opts) {
  // Skip for help and version commands
  if (opts.id === 'help' || opts.id === 'version') {
    return
  }

  // Skip for setup command (would be circular)
  if (opts.id === 'setup') {
    return
  }

  try {
    // Try to get API key
    getApiKey()
  } catch (error) {
    // First run - no API key configured

    if (!process.stdout.isTTY || process.env.CI) {
      // Non-interactive environment
      this.error(
        '\nNotion API token not configured.\n\n' +
        'For automated setup:\n' +
        '  export NOTION_TOKEN=secret_xxx\n\n' +
        'Or run setup command:\n' +
        '  notion-cli setup --non-interactive --api-key=secret_xxx',
        { exit: 1 }
      )
    } else {
      // Interactive environment - offer to run setup
      this.log(chalk.yellow('\n⚠ Notion API token not configured\n'))

      const { default: prompts } = await import('prompts')

      const { runSetup } = await prompts({
        type: 'confirm',
        name: 'runSetup',
        message: 'Would you like to run the setup wizard now?',
        initial: true
      })

      if (runSetup) {
        this.log('')  // Blank line
        await this.config.runCommand('setup')
        process.exit(0)
      } else {
        this.log('\nYou can run setup later with: ' +
                chalk.cyan('notion-cli setup'))
        process.exit(1)
      }
    }
  }
}

export default hook
```

#### Phase 4: Enhanced Documentation

**Add to README.md:**

````markdown
## Installation & Setup

### 1. Install

```bash
npm install -g @coastal-programs/notion-cli
```

### 2. Setup

**Interactive Setup (Recommended for first-time users):**
```bash
notion-cli setup
```

This will:
- Prompt for your Notion API token
- Verify the token works
- Save configuration to `~/.notion-cli/config.json`
- Set recommended defaults

**Quick Setup (For automation/AI agents):**
```bash
export NOTION_TOKEN=secret_xxx
notion-cli setup --non-interactive
```

**Or skip setup and use environment variables:**
```bash
export NOTION_TOKEN=secret_xxx
notion-cli page retrieve <PAGE_ID>
```

### Configuration

Configuration is loaded in this order (highest priority first):

1. **Command-line flags** - `--api-key=xxx`
2. **Environment variables** - `NOTION_TOKEN=xxx`
3. **Config file** - `~/.notion-cli/config.json`
4. **Defaults**

**Config File Location:**
- Linux/macOS: `~/.config/notion-cli/config.json`
- Windows: `%APPDATA%\notion-cli\config.json`

**Example config.json:**
```json
{
  "apiKey": "secret_xxx",
  "cacheEnabled": true,
  "maxRetries": 5,
  "outputFormat": "json"
}
```

### Getting Your API Token

1. Go to https://developers.notion.com/
2. Click "Create new integration"
3. Give it a name and select capabilities
4. Click "Submit" to create the integration
5. Copy the "Internal Integration Token" (starts with `secret_`)
6. Share your pages/databases with the integration

For detailed setup instructions, see: https://developers.notion.com/docs/create-a-notion-integration
````

---

## 7. Key Takeaways & Best Practices

### For AI-Agent Friendly CLI Design:

1. **Support Both Modes:**
   - Interactive mode for humans
   - Non-interactive mode for automation
   - Automatic detection of environment (TTY, CI, etc.)

2. **Configuration Precedence:**
   - Command-line flags (highest)
   - Environment variables
   - Config files
   - Defaults (lowest)

3. **Output Formats:**
   - JSON mode for machine parsing (`--json`)
   - Human-readable table/text for interactive use
   - Consistent structure across commands

4. **Error Handling:**
   - Exit codes: 0 for success, non-zero for errors
   - Structured error output in JSON mode
   - Helpful error messages with remediation steps
   - Distinguish retryable vs non-retryable errors

5. **First-Run Experience:**
   - Detect unconfigured state
   - Offer setup wizard in interactive mode
   - Provide clear instructions for non-interactive setup
   - Don't block on first run - allow env vars as alternative

6. **Credential Management:**
   - Support multiple methods (env vars, config files, keychains)
   - Never log or echo credentials
   - Use OS keychain for persistent storage (desktop apps)
   - Use secret managers for production (servers)
   - Keep .env files out of version control

7. **Documentation:**
   - Quick start with minimal commands
   - Show both interactive and non-interactive examples
   - Document environment variables
   - Include troubleshooting section
   - Provide example config files

### Anti-Patterns to Avoid:

- ❌ **Requiring interactive prompts** - Always have non-interactive alternative
- ❌ **Silent failures** - Always return proper exit codes
- ❌ **Inconsistent flags** - Use same patterns across commands
- ❌ **Hardcoded credentials** - Never commit secrets to code
- ❌ **Unclear error messages** - Tell users what went wrong AND how to fix it
- ❌ **No configuration discovery** - Support standard config locations
- ❌ **Brittle auth flows** - Provide fallbacks for different environments
- ❌ **Assuming GUI available** - Support headless/SSH scenarios
- ❌ **Blocking on telemetry** - Make it optional, off by default
- ❌ **Complex installation** - Aim for single package manager command

---

## 8. References & Further Reading

### Official Documentation:
- **Model Context Protocol:** https://www.anthropic.com/news/model-context-protocol
- **Claude Code CLI:** https://docs.claude.com/en/docs/claude-code/cli-reference
- **oclif Framework:** https://oclif.io/docs/introduction
- **AWS CLI Setup:** https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-quickstart.html
- **GitHub CLI:** https://cli.github.com/manual/
- **XDG Base Directory Spec:** https://specifications.freedesktop.org/basedir-spec/latest/

### Articles & Blog Posts:
- **12 Factor CLI Apps:** https://medium.com/@jdxcode/12-factor-cli-apps-dd3c227a0e46
- **Heroku CLI Framework (oclif):** https://blog.heroku.com/open-cli-framework
- **Developer Experience Review - Heroku:** https://betta.io/blog/2018/02/27/developer-experience-review-heroku/
- **CLI-First Devtools:** https://blog.console.dev/cli-first-devtools/

### Security Resources:
- **API Key Safety Best Practices:** https://help.openai.com/en/articles/5112595-best-practices-for-api-key-safety
- **Environment Variables Security:** https://www.netlify.com/blog/a-guide-to-storing-api-keys-securely-with-environment-variables/
- **Git Credential Manager:** https://github.com/git-ecosystem/git-credential-manager

### Tools & Libraries:
- **prompts (Node.js):** https://github.com/terkelg/prompts
- **Commander.js:** https://github.com/tj/commander.js
- **yargs:** https://github.com/yargs/yargs
- **chalk:** https://github.com/chalk/chalk
- **ora:** https://github.com/sindresorhus/ora
- **dotenv:** https://github.com/motdotla/dotenv
- **envalid:** https://github.com/af/envalid
- **keytar (Node.js keychain):** https://github.com/atom/node-keytar
- **pycreds (Python keychain):** https://pypi.org/project/pycreds/

---

## Appendix A: Example Configuration Files

### notion-cli.config.json
```json
{
  "$schema": "https://example.com/notion-cli-config.schema.json",
  "apiKey": "secret_xxx",
  "defaultWorkspace": "personal",
  "cache": {
    "enabled": true,
    "ttl": {
      "dataSource": 600000,
      "user": 3600000,
      "page": 60000,
      "block": 30000
    },
    "maxSize": 1000
  },
  "retry": {
    "maxRetries": 5,
    "baseDelay": 1000,
    "maxDelay": 30000
  },
  "output": {
    "defaultFormat": "json",
    "colorize": true,
    "timestamps": true
  },
  "telemetry": {
    "enabled": false
  }
}
```

### .env.example
```bash
# Required
NOTION_TOKEN=secret_xxx

# Optional - Output
NOTION_CLI_OUTPUT_FORMAT=json  # json|table|csv|yaml
NOTION_CLI_NO_COLOR=false

# Optional - Cache Configuration
NOTION_CLI_CACHE_ENABLED=true
NOTION_CLI_CACHE_MAX_SIZE=1000
NOTION_CLI_CACHE_DS_TTL=600000       # 10 minutes
NOTION_CLI_CACHE_USER_TTL=3600000    # 1 hour
NOTION_CLI_CACHE_PAGE_TTL=60000      # 1 minute
NOTION_CLI_CACHE_BLOCK_TTL=30000     # 30 seconds

# Optional - Retry Configuration
NOTION_CLI_MAX_RETRIES=3
NOTION_CLI_BASE_DELAY=1000
NOTION_CLI_MAX_DELAY=30000

# Optional - Debug
DEBUG=true
NOTION_CLI_LOG_LEVEL=info  # debug|info|warn|error
```

---

## Appendix B: Setup Command Flowchart

```
Start
  ├─> Is `setup` command invoked?
  │     ├─> Yes ─> Continue to setup
  │     └─> No ─> Check config on command execution
  │               ├─> Config exists? ─> Run command
  │               └─> No config ─> Show first-run message
  │
  ├─> Check environment
  │     ├─> Is TTY && !CI?
  │     │     ├─> Yes ─> Interactive mode
  │     │     └─> No ─> Non-interactive mode
  │     │
  │     ├─> Interactive Mode:
  │     │     ├─> Show welcome message
  │     │     ├─> Prompt for API key
  │     │     ├─> Prompt for preferences
  │     │     ├─> Verify API key
  │     │     ├─> Save config
  │     │     └─> Show success message
  │     │
  │     └─> Non-Interactive Mode:
  │           ├─> Get API key from flag or env
  │           ├─> API key present?
  │           │     ├─> Yes ─> Verify
  │           │     └─> No ─> Error with instructions
  │           ├─> Save config
  │           └─> Output JSON (if --json flag)
  │
  └─> End
```

---

**Document Version:** 1.0
**Last Updated:** 2025-10-22
**Author:** Research compiled from industry best practices, framework documentation, and real-world CLI implementations
