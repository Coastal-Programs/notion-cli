import { Command, Flags, ux } from '@oclif/core'
import { loadCache, getCachePath } from '../utils/workspace-cache'
import { outputMarkdownTable, outputPrettyTable, outputCompactJson } from '../helper'
import { AutomationFlags, OutputFormatFlags } from '../base-flags'
import { NotionCLIError } from '../errors'

export default class List extends Command {
  static description = 'List all cached databases from your workspace'

  static aliases: string[] = ['db:list', 'ls']

  static examples = [
    {
      description: 'List all cached databases',
      command: 'notion-cli list',
    },
    {
      description: 'List databases in markdown format',
      command: 'notion-cli list --markdown',
    },
    {
      description: 'List databases in JSON format',
      command: 'notion-cli list --json',
    },
    {
      description: 'List databases in pretty table format',
      command: 'notion-cli list --pretty',
    },
  ]

  static flags = {
    ...ux.table.flags(),
    ...AutomationFlags,
    ...OutputFormatFlags,
  }

  public async run(): Promise<void> {
    const { flags } = await this.parse(List)

    try {
      // Load cache
      const cache = await loadCache()

      if (!cache) {
        const cachePath = await getCachePath()

        if (flags.json) {
          this.log(JSON.stringify({
            success: false,
            error: 'Cache not found',
            message: `No cache file found at ${cachePath}`,
            suggestion: 'Run "notion-cli sync" to build the cache',
          }, null, 2))
        } else {
          this.log(`No cache found at ${cachePath}`)
          this.log('\nRun "notion-cli sync" to build the cache.')
        }

        process.exit(1)
        return
      }

      const databases = cache.databases

      if (databases.length === 0) {
        if (flags.json) {
          this.log(JSON.stringify({
            success: true,
            count: 0,
            databases: [],
            message: 'No databases found in cache',
          }, null, 2))
        } else {
          this.log('No databases found in cache.')
          this.log('Your integration may not have access to any databases.')
        }

        process.exit(0)
        return
      }

      // Define columns for table output
      const columns = {
        title: {
          header: 'Title',
          get: (row: any) => row.title,
        },
        id: {
          header: 'ID',
          get: (row: any) => row.id,
        },
        aliases: {
          header: 'Aliases (first 3)',
          get: (row: any) => row.aliases.slice(0, 3).join(', '),
        },
        url: {
          header: 'URL',
          get: (row: any) => row.url || '',
        },
      }

      // Handle compact JSON output
      if (flags['compact-json']) {
        outputCompactJson(databases)
        process.exit(0)
        return
      }

      // Handle markdown table output
      if (flags.markdown) {
        outputMarkdownTable(databases, columns)
        process.exit(0)
        return
      }

      // Handle pretty table output
      if (flags.pretty) {
        outputPrettyTable(databases, columns)
        process.exit(0)
        return
      }

      // Handle JSON output for automation
      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          count: databases.length,
          lastSync: cache.lastSync,
          databases: databases.map(db => ({
            id: db.id,
            title: db.title,
            aliases: db.aliases,
            url: db.url,
            lastEditedTime: db.lastEditedTime,
          })),
          timestamp: new Date().toISOString(),
        }, null, 2))
        process.exit(0)
        return
      }

      // Handle table output (default)
      this.log(`\nCached Databases (${databases.length} total)`)
      this.log(`Last synced: ${new Date(cache.lastSync).toLocaleString()}\n`)

      const options = {
        printLine: this.log.bind(this),
        ...flags,
      }
      ux.table(databases, columns, options)

      this.log(`\nTip: Run "notion-cli sync" to refresh the cache.`)
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
}
