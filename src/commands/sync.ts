import { Command, Flags, ux } from '@oclif/core'
import { client } from '../notion'
import { fetchWithRetry as enhancedFetchWithRetry } from '../retry'
import {
  saveCache,
  getCachePath,
  buildCacheEntry,
  WorkspaceCache,
} from '../utils/workspace-cache'
import { AutomationFlags } from '../base-flags'
import {
  NotionCLIError,
  wrapNotionError
} from '../errors'
import { validateNotionToken } from '../utils/token-validator'
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
    const startTime = Date.now()

    try {
      // Verify NOTION_TOKEN is set (throws if not)
      validateNotionToken()

      if (!flags.json) {
        ux.action.start('Syncing workspace databases')
      }

      // Fetch all databases from Notion API with progress updates
      const databases = await this.fetchAllDatabases(flags.json)

      // const _fetchTime = Date.now() - startTime

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
      const executionTime = Date.now() - startTime

      // Build comprehensive metadata
      const metadata = {
        sync_time: new Date().toISOString(),
        execution_time_ms: executionTime,
        databases_found: databases.length,
        cache_ttls: {
          in_memory: {
            data_source_ms: parseInt(process.env.NOTION_CLI_CACHE_DS_TTL || '600000', 10),
            page_ms: parseInt(process.env.NOTION_CLI_CACHE_PAGE_TTL || '60000', 10),
            user_ms: parseInt(process.env.NOTION_CLI_CACHE_USER_TTL || '3600000', 10),
            block_ms: parseInt(process.env.NOTION_CLI_CACHE_BLOCK_TTL || '30000', 10),
          },
          workspace: {
            persistence: 'until next sync',
            recommended_sync_interval_hours: 24,
          },
        },
        next_recommended_sync: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(),
        cache_location: cachePath,
      }

      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          data: {
            databases: cacheEntries.map(db => ({
              id: db.id,
              title: db.title,
              aliases: db.aliases,
              url: db.url,
            })),
            summary: {
              total: databases.length,
              cached_at: cache.lastSync,
              cache_version: cache.version,
            },
          },
          metadata,
        }, null, 2))
      } else {
        ux.action.stop()

        // Enhanced completion summary
        const elapsedSeconds = (executionTime / 1000).toFixed(2)
        this.log(`\n✓ Synced ${databases.length} database${databases.length === 1 ? '' : 's'} in ${elapsedSeconds}s`)
        this.log('')
        this.log(`📁 Cache: ${cachePath}`)
        this.log(`🕐 Last updated: ${new Date(cache.lastSync).toLocaleString()}`)
        this.log(`📊 Databases: ${databases.length} total`)
        this.log('')
        this.log(`Next sync recommended: ${new Date(metadata.next_recommended_sync).toLocaleString()}`)

        if (databases.length > 0) {
          this.log('\nIndexed databases:')
          cacheEntries.slice(0, 10).forEach(db => {
            const aliasesStr = db.aliases.slice(0, 3).join(', ')
            this.log(`  • ${db.title} (aliases: ${aliasesStr})`)
          })

          if (databases.length > 10) {
            this.log(`  ... and ${databases.length - 10} more`)
          }

          this.log('\nTry: notion-cli list')
        } else {
          this.log('\nNo databases found in workspace.')
          this.log('Make sure your integration has access to databases.')
        }
      }

      process.exit(0)
    } catch (error) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            endpoint: 'search',
            resourceType: 'database'
          })

      if (flags.json) {
        this.log(JSON.stringify(cliError.toJSON(), null, 2))
      } else {
        ux.action.stop('failed')
        this.error(cliError.toHumanString())
      }

      process.exit(1)
    }
  }

  /**
   * Fetch all databases from Notion API with pagination
   */
  private async fetchAllDatabases(isJsonMode: boolean): Promise<DataSourceObjectResponse[]> {
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

      // Show progress update (only in non-JSON mode)
      if (!isJsonMode && response.has_more) {
        // Update the spinner text to show current count
        ux.action.start(`Syncing workspace databases (found ${databases.length} so far)`)
      }

      if (!response.has_more || !response.next_cursor) {
        break
      }

      cursor = response.next_cursor
    }

    return databases
  }
}
