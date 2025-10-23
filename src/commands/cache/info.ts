import { Command } from '@oclif/core'
import { AutomationFlags } from '../../base-flags'
import { loadCache, getCachePath } from '../../utils/workspace-cache'
import { cacheManager } from '../../cache'
import {
  NotionCLIError,
  wrapNotionError
} from '../../errors'
import * as os from 'os'
import * as path from 'path'

export default class CacheInfo extends Command {
  static description = 'Show cache statistics and configuration'

  static aliases = ['cache:stats', 'cache:status']

  static examples = [
    {
      description: 'Show cache info in JSON format',
      command: 'notion-cli cache:info --json',
    },
    {
      description: 'Show cache statistics',
      command: 'notion-cli cache:info',
    },
  ]

  static flags = {
    ...AutomationFlags,
  }

  async run() {
    const { flags } = await this.parse(CacheInfo)

    try {
      // Get workspace cache
      const workspaceCache = await loadCache()
      const cachePath = await getCachePath()

      // Get in-memory cache stats
      const inMemoryStats = cacheManager.getStats()
      const hitRate = cacheManager.getHitRate()

      // Calculate workspace cache age if available
      let workspaceInfo = null
      if (workspaceCache) {
        const lastSyncTime = new Date(workspaceCache.lastSync)
        const cacheAgeMs = Date.now() - lastSyncTime.getTime()
        const cacheAgeHours = cacheAgeMs / (1000 * 60 * 60)
        const isStale = cacheAgeHours > 24

        workspaceInfo = {
          databases_cached: workspaceCache.databases.length,
          last_sync: workspaceCache.lastSync,
          cache_age_ms: cacheAgeMs,
          cache_age_hours: parseFloat(cacheAgeHours.toFixed(2)),
          is_stale: isStale,
          stale_threshold_hours: 24,
          cache_version: workspaceCache.version,
          cache_location: cachePath,
        }
      }

      // Build comprehensive cache info
      const cacheInfo = {
        in_memory: {
          enabled: cacheManager.isEnabled(),
          stats: {
            size: inMemoryStats.size,
            hits: inMemoryStats.hits,
            misses: inMemoryStats.misses,
            sets: inMemoryStats.sets,
            evictions: inMemoryStats.evictions,
            hit_rate: parseFloat((hitRate * 100).toFixed(2)),
          },
          ttls_ms: {
            data_source: parseInt(process.env.NOTION_CLI_CACHE_DS_TTL || '600000', 10),
            page: parseInt(process.env.NOTION_CLI_CACHE_PAGE_TTL || '60000', 10),
            user: parseInt(process.env.NOTION_CLI_CACHE_USER_TTL || '3600000', 10),
            block: parseInt(process.env.NOTION_CLI_CACHE_BLOCK_TTL || '30000', 10),
          },
          max_size: parseInt(process.env.NOTION_CLI_CACHE_MAX_SIZE || '1000', 10),
        },
        workspace: workspaceInfo,
        recommendations: {
          sync_interval_hours: 24,
          next_sync: workspaceCache ?
            new Date(new Date(workspaceCache.lastSync).getTime() + 24 * 60 * 60 * 1000).toISOString() :
            null,
          action_needed: !workspaceCache ? 'Run "notion-cli sync" to initialize cache' :
            (workspaceInfo && workspaceInfo.is_stale) ? 'Cache is stale, run "notion-cli sync"' :
            'Cache is fresh',
        },
      }

      // JSON output
      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          data: cacheInfo,
          metadata: {
            timestamp: new Date().toISOString(),
            command: 'cache:info',
          },
        }, null, 2))
        process.exit(0)
      }

      // Human-readable output
      this.log('Cache Configuration')
      this.log('='.repeat(60))

      this.log('\nIn-Memory Cache:')
      this.log(`  Enabled: ${cacheInfo.in_memory.enabled ? 'Yes' : 'No'}`)
      this.log(`  Size: ${inMemoryStats.size} / ${cacheInfo.in_memory.max_size}`)
      this.log(`  Hits: ${inMemoryStats.hits}`)
      this.log(`  Misses: ${inMemoryStats.misses}`)
      this.log(`  Hit Rate: ${(hitRate * 100).toFixed(1)}%`)
      this.log(`  Evictions: ${inMemoryStats.evictions}`)

      this.log('\n  TTLs (milliseconds):')
      this.log(`    Data Sources: ${cacheInfo.in_memory.ttls_ms.data_source} (${(cacheInfo.in_memory.ttls_ms.data_source / 60000).toFixed(0)} min)`)
      this.log(`    Pages: ${cacheInfo.in_memory.ttls_ms.page} (${(cacheInfo.in_memory.ttls_ms.page / 1000).toFixed(0)} sec)`)
      this.log(`    Users: ${cacheInfo.in_memory.ttls_ms.user} (${(cacheInfo.in_memory.ttls_ms.user / 60000).toFixed(0)} min)`)
      this.log(`    Blocks: ${cacheInfo.in_memory.ttls_ms.block} (${(cacheInfo.in_memory.ttls_ms.block / 1000).toFixed(0)} sec)`)

      this.log('\nWorkspace Cache:')
      if (workspaceInfo) {
        this.log(`  Databases: ${workspaceInfo.databases_cached}`)
        this.log(`  Last Sync: ${new Date(workspaceInfo.last_sync).toLocaleString()}`)
        this.log(`  Age: ${workspaceInfo.cache_age_hours} hours`)
        this.log(`  Status: ${workspaceInfo.is_stale ? '⚠️  STALE' : '✓ Fresh'}`)
        this.log(`  Location: ${workspaceInfo.cache_location}`)
      } else {
        this.log(`  Status: Not initialized`)
        this.log(`  Action: Run "notion-cli sync"`)
      }

      this.log('\nRecommendations:')
      this.log(`  Sync Interval: Every ${cacheInfo.recommendations.sync_interval_hours} hours`)
      if (cacheInfo.recommendations.next_sync) {
        this.log(`  Next Sync: ${new Date(cacheInfo.recommendations.next_sync).toLocaleString()}`)
      }
      this.log(`  Action: ${cacheInfo.recommendations.action_needed}`)

      process.exit(0)
    } catch (error: any) {
      const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, {
            endpoint: 'cache.info'
          })

      if (flags.json) {
        this.log(JSON.stringify(cliError.toJSON(), null, 2))
      } else {
        this.error(cliError.toHumanString())
      }

      process.exit(1)
    }
  }
}
