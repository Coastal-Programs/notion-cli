import { Command } from '@oclif/core'
import { AutomationFlags } from '../base-flags'
import * as notion from '../notion'
import { cacheManager } from '../cache'
import {
  wrapNotionError
} from '../errors'
import { loadCache } from '../utils/workspace-cache'
import { validateNotionToken } from '../utils/token-validator'

export default class Whoami extends Command {
  static description = 'Verify API connectivity and show workspace context'

  static aliases = ['test', 'health', 'connectivity']

  static examples = [
    {
      description: 'Check connection and show bot info',
      command: '$ notion-cli whoami',
    },
    {
      description: 'Check connection and output as JSON',
      command: '$ notion-cli whoami --json',
    },
    {
      description: 'Bypass cache for fresh connectivity test',
      command: '$ notion-cli whoami --no-cache',
    },
  ]

  static flags = {
    ...AutomationFlags,
  }

  async run() {
    const { flags } = await this.parse(Whoami)
    const startTime = Date.now()

    try {
      // Verify NOTION_TOKEN is set (throws if not)
      validateNotionToken()

      // Get bot user info (with retry and caching)
      const user = await notion.botUser()

      // Get cache stats from in-memory cache
      const cacheStats = cacheManager.getStats()
      const cacheHitRate = cacheManager.getHitRate()

      // Load workspace cache (databases.json)
      const cache = await loadCache()

      // Calculate connection latency
      const latencyMs = Date.now() - startTime

      // Extract bot info safely
      let botInfo: any = null
      let workspaceInfo: any = null

      if (user.type === 'bot') {
        const botUser = user as any
        if (botUser.bot && typeof botUser.bot === 'object' && 'owner' in botUser.bot) {
          botInfo = {
            owner: botUser.bot.owner,
            workspace_name: botUser.bot.workspace_name,
            workspace_id: botUser.bot.workspace_id,
          }

          // Build workspace info if available
          if (botUser.bot.workspace_name) {
            workspaceInfo = {
              name: botUser.bot.workspace_name,
              id: botUser.bot.workspace_id,
            }
          }
        }
      }

      // Build response data
      const data = {
        bot: {
          id: user.id,
          name: user.name || 'Unnamed Bot',
          type: user.type,
          ...(botInfo && { bot_info: botInfo })
        },
        workspace: workspaceInfo,
        api_version: '2022-06-28',
        cli_version: this.config.version,
        cache_status: {
          enabled: !flags['no-cache'] && cacheManager.isEnabled(),
          in_memory: {
            size: cacheStats.size,
            hits: cacheStats.hits,
            misses: cacheStats.misses,
            hit_rate: cacheHitRate,
            evictions: cacheStats.evictions,
          },
          workspace: {
            databases_cached: cache?.databases?.length || 0,
            last_sync: cache?.lastSync || null,
            cache_version: cache?.version || null,
          }
        },
        connection: {
          status: 'connected',
          latency_ms: latencyMs
        }
      }

      // Output JSON envelope
      if (flags.json) {
        this.log(JSON.stringify({
          success: true,
          data,
          metadata: {
            timestamp: new Date().toISOString(),
            command: 'whoami',
            execution_time_ms: latencyMs
          }
        }, null, 2))
        process.exit(0)
      }

      // Human-readable output
      this.log('\nConnection Status')
      this.log('='.repeat(60))
      this.log(`Status:      Connected`)
      this.log(`Latency:     ${data.connection.latency_ms}ms`)

      this.log('\nBot Information')
      this.log('='.repeat(60))
      this.log(`Name:        ${data.bot.name}`)
      this.log(`ID:          ${data.bot.id}`)
      this.log(`Type:        ${data.bot.type}`)

      if (data.workspace) {
        this.log('\nWorkspace Information')
        this.log('='.repeat(60))
        this.log(`Name:        ${data.workspace.name || 'N/A'}`)
        if (data.workspace.id) {
          this.log(`ID:          ${data.workspace.id}`)
        }
      }

      this.log('\nAPI & CLI Version')
      this.log('='.repeat(60))
      this.log(`CLI:         ${data.cli_version}`)
      this.log(`API:         ${data.api_version}`)

      this.log('\nCache Status')
      this.log('='.repeat(60))
      this.log(`Enabled:     ${data.cache_status.enabled ? 'Yes' : 'No'}`)

      if (data.cache_status.enabled) {
        this.log('\nIn-Memory Cache:')
        this.log(`  Size:        ${data.cache_status.in_memory.size} entries`)
        this.log(`  Hits:        ${data.cache_status.in_memory.hits}`)
        this.log(`  Misses:      ${data.cache_status.in_memory.misses}`)
        this.log(`  Hit Rate:    ${(data.cache_status.in_memory.hit_rate * 100).toFixed(1)}%`)
        this.log(`  Evictions:   ${data.cache_status.in_memory.evictions}`)

        this.log('\nWorkspace Cache:')
        this.log(`  Databases:   ${data.cache_status.workspace.databases_cached}`)
        if (data.cache_status.workspace.last_sync) {
          const syncDate = new Date(data.cache_status.workspace.last_sync)
          this.log(`  Last Sync:   ${syncDate.toLocaleString()}`)
        } else {
          this.log(`  Last Sync:   Never (run 'notion-cli sync' to initialize)`)
        }
      }

      this.log('\n' + '='.repeat(60))
      this.log('\nConnection verified successfully!')

      // Provide helpful tips
      if (!cache || cache.databases.length === 0) {
        this.log('\nTip: Run "notion-cli sync" to cache workspace databases for faster lookups')
      }

      process.exit(0)
    } catch (error) {
      const cliError = wrapNotionError(error, {
        endpoint: 'users.botUser',
        resourceType: 'user'
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
