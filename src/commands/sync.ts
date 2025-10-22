import { Command, Flags, ux } from '@oclif/core'
import { client } from '../notion'
import { fetchWithRetry as enhancedFetchWithRetry } from '../retry'
import {
  loadCache,
  saveCache,
  getCachePath,
  buildCacheEntry,
  createEmptyCache,
  WorkspaceCache,
} from '../utils/workspace-cache'
import { getDataSourceTitle } from '../helper'
import { AutomationFlags } from '../base-flags'
import { NotionCLIError, wrapNotionError } from '../errors'
import { DataSourceObjectResponse } from '@notionhq/client/build/src/api-endpoints'

export default class Sync extends Command {
  static description = 'Sync workspace databases to local cache for fast lookups'

  static aliases: string[] = ['db:sync']

  static examples = [
    {
      description: 'Sync all workspace databases',
      command: 'notion-cli sync',
    },
    {
      description: 'Force resync even if cache exists',
      command: 'notion-cli sync --force',
    },
    {
      description: 'Sync and output as JSON',
      command: 'notion-cli sync --json',
    },
  ]

  static flags = {
    force: Flags.boolean({
      char: 'f',
      description: 'Force resync even if cache is fresh',
      default: false,
    }),
    ...AutomationFlags,
  }

  public async run(): Promise<void> {
    const { flags } = await this.parse(Sync)

    try {
      if (!flags.json) {
        ux.action.start('Syncing workspace databases')
      }

      // Fetch all databases from Notion API
      const databases = await this.fetchAllDatabases()

      if (!flags.json) {
        ux.action.stop(`Found ${databases.length} database${databases.length === 1 ? '' : 's'}`)
        ux.action.start('Generating search aliases')
      }

      // Build cache entries
      const cacheEntries = databases.map(db => buildCacheEntry(db))

      if (!flags.json) {
        ux.action.stop()
        ux.action.start('Saving cache')
      }

      // Save to cache
      const cache: WorkspaceCache = {
        version: '1.0.0',
        lastSync: new Date().toISOString(),
        databases: cacheEntries,
      }

      await saveCache(cache)

      const cachePath = await getCachePath()

      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          count: databases.length,
          cachePath,
          databases: cacheEntries.map(db => ({
            id: db.id,
            title: db.title,
            aliases: db.aliases,
            url: db.url,
          })),
          timestamp: new Date().toISOString(),
        }, null, 2))
      } else {
        ux.action.stop()
        this.log(`\n✓ Cache saved to ${cachePath}\n`)

        if (databases.length > 0) {
          this.log('Indexed databases:')
          cacheEntries.slice(0, 10).forEach(db => {
            const aliasesStr = db.aliases.slice(0, 3).join(', ')
            this.log(`  • ${db.title} (aliases: ${aliasesStr})`)
          })

          if (databases.length > 10) {
            this.log(`  ... and ${databases.length - 10} more`)
          }
        } else {
          this.log('No databases found in workspace.')
          this.log('Make sure your integration has access to databases.')
        }
      }

      process.exit(0)
    } catch (error) {
      const cliError = wrapNotionError(error)

      if (flags.json) {
        this.log(JSON.stringify(cliError.toJSON(), null, 2))
      } else {
        ux.action.stop('failed')
        this.error(cliError.message)
      }

      process.exit(1)
    }
  }

  /**
   * Fetch all databases from Notion API with pagination
   */
  private async fetchAllDatabases(): Promise<DataSourceObjectResponse[]> {
    const databases: DataSourceObjectResponse[] = []
    let cursor: string | undefined = undefined

    while (true) {
      const response = await enhancedFetchWithRetry(
        () => client.search({
          filter: {
            value: 'data_source',
            property: 'object',
          },
          start_cursor: cursor,
          page_size: 100, // Max allowed by API
        }),
        {
          context: 'sync:fetchAllDatabases',
          config: { maxRetries: 5 }, // Higher retries for sync
        }
      )

      databases.push(...response.results as DataSourceObjectResponse[])

      if (!response.has_more || !response.next_cursor) {
        break
      }

      cursor = response.next_cursor
    }

    return databases
  }
}
