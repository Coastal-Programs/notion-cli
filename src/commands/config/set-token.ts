import { Command, Args, Flags } from '@oclif/core'
import * as fs from 'fs/promises'
import * as path from 'path'
import * as os from 'os'
import { AutomationFlags } from '../../base-flags'

export default class ConfigSetToken extends Command {
  static description = 'Set NOTION_TOKEN in your shell configuration file'

  static aliases: string[] = ['config:token']

  static examples = [
    {
      description: 'Set Notion token interactively',
      command: 'notion-cli config set-token',
    },
    {
      description: 'Set Notion token directly',
      command: 'notion-cli config set-token secret_abc123...',
    },
    {
      description: 'Set token with JSON output',
      command: 'notion-cli config set-token secret_abc123... --json',
    },
  ]

  static args = {
    token: Args.string({
      description: 'Notion integration token (starts with secret_)',
      required: false,
    }),
  }

  static flags = {
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { args, flags } = await this.parse(ConfigSetToken)

    try {
      // Get token from args or prompt
      let token = args.token

      if (!token) {
        if (flags.json) {
          this.log(JSON.stringify({
            success: false,
            error: 'Token required in JSON mode',
            message: 'Please provide the token as an argument when using --json flag',
          }, null, 2))
          process.exit(1)
          return
        }

        // Interactive prompt
        const readline = require('readline')
        const rl = readline.createInterface({
          input: process.stdin,
          output: process.stdout,
        })

        token = await new Promise<string>((resolve) => {
          rl.question('Enter your Notion integration token: ', (answer: string) => {
            rl.close()
            resolve(answer.trim())
          })
        })
      }

      // Validate token format
      if (!token || !token.startsWith('secret_')) {
        if (flags.json) {
          this.log(JSON.stringify({
            success: false,
            error: 'Invalid token format',
            message: 'Notion tokens must start with "secret_"',
          }, null, 2))
        } else {
          this.error('Invalid token format. Notion tokens must start with "secret_"')
        }
        process.exit(1)
        return
      }

      // Detect shell and rc file
      const shell = this.detectShell()
      const rcFile = this.getRcFilePath(shell)

      // Read existing rc file
      let rcContent = ''
      try {
        rcContent = await fs.readFile(rcFile, 'utf-8')
      } catch (error: any) {
        if (error.code !== 'ENOENT') {
          throw error
        }
        // File doesn't exist, will create it
      }

      // Check if NOTION_TOKEN already exists
      const tokenLineRegex = /^export\s+NOTION_TOKEN=.*/gm
      const newTokenLine = `export NOTION_TOKEN="${token}"`

      let updatedContent: string
      if (tokenLineRegex.test(rcContent)) {
        // Replace existing token
        updatedContent = rcContent.replace(tokenLineRegex, newTokenLine)
      } else {
        // Add new token
        updatedContent = rcContent.trim() + '\n\n# Notion CLI Token\n' + newTokenLine + '\n'
      }

      // Write updated rc file
      await fs.writeFile(rcFile, updatedContent, 'utf-8')

      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          message: 'Token saved successfully',
          rcFile,
          shell,
          nextSteps: [
            `Reload your shell: source ${rcFile}`,
            'Run: notion-cli sync',
          ],
        }, null, 2))
      } else {
        this.log(`\n✓ Token saved to ${rcFile}`)
        this.log('\nNext steps:')
        this.log(`  1. Reload your shell: source ${rcFile}`)
        this.log(`  2. Or restart your terminal`)
        this.log(`  3. Run: notion-cli sync`)
        this.log('\nWould you like to sync your workspace now? (y/n)')

        const readline = require('readline')
        const rl = readline.createInterface({
          input: process.stdin,
          output: process.stdout,
        })

        const answer = await new Promise<string>((resolve) => {
          rl.question('> ', (answer: string) => {
            rl.close()
            resolve(answer.trim().toLowerCase())
          })
        })

        if (answer === 'y' || answer === 'yes') {
          // Set token in current process
          process.env.NOTION_TOKEN = token

          // Run sync command
          this.log('\nRunning sync...\n')
          const Sync = require('../sync').default
          await Sync.run([])
        } else {
          this.log('\nSkipping sync. You can run it manually with: notion-cli sync')
        }
      }

      process.exit(0)
    } catch (error: any) {
      if (flags.json) {
        this.log(JSON.stringify({
          success: false,
          error: error.message,
        }, null, 2))
      } else {
        this.error(error.message)
      }

      process.exit(1)
    }
  }

  /**
   * Detect the current shell
   */
  private detectShell(): string {
    const shell = process.env.SHELL || ''

    if (shell.includes('zsh')) return 'zsh'
    if (shell.includes('bash')) return 'bash'
    if (shell.includes('fish')) return 'fish'

    // Default to bash on Unix, powershell on Windows
    return process.platform === 'win32' ? 'powershell' : 'bash'
  }

  /**
   * Get the rc file path for the detected shell
   */
  private getRcFilePath(shell: string): string {
    const home = os.homedir()

    switch (shell) {
      case 'zsh':
        return path.join(home, '.zshrc')
      case 'bash':
        return path.join(home, '.bashrc')
      case 'fish':
        return path.join(home, '.config', 'fish', 'config.fish')
      case 'powershell':
        return path.join(home, 'Documents', 'PowerShell', 'Microsoft.PowerShell_profile.ps1')
      default:
        return path.join(home, '.bashrc')
    }
  }
}
