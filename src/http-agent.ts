/**
 * HTTP Agent Configuration
 *
 * Configures connection pooling and HTTP keep-alive to reduce connection overhead.
 * Enables connection reuse across multiple API requests for better performance.
 */

import * as https from 'https'

/**
 * HTTPS Agent with keep-alive enabled
 */
export const httpsAgent = new https.Agent({
  // Enable keep-alive connections
  keepAlive: process.env.NOTION_CLI_HTTP_KEEP_ALIVE !== 'false',

  // Keep-alive timeout (how long to keep idle connections open)
  keepAliveMsecs: parseInt(process.env.NOTION_CLI_HTTP_KEEP_ALIVE_MS || '60000', 10),

  // Maximum number of sockets to allow (concurrent connections)
  maxSockets: parseInt(process.env.NOTION_CLI_HTTP_MAX_SOCKETS || '50', 10),

  // Maximum number of sockets to leave open in a free state (connection pool size)
  maxFreeSockets: parseInt(process.env.NOTION_CLI_HTTP_MAX_FREE_SOCKETS || '10', 10),
})

/**
 * Default request timeout in milliseconds
 * Note: timeout is set per-request, not on the agent
 */
export const REQUEST_TIMEOUT = parseInt(process.env.NOTION_CLI_HTTP_TIMEOUT || '30000', 10)

/**
 * Get current agent statistics
 */
export function getAgentStats(): {
  sockets: number
  freeSockets: number
  requests: number
} {
  const agent = httpsAgent as any

  // Count active sockets
  const socketsCount = Object.keys(agent.sockets || {}).reduce((count, key) => {
    return count + (agent.sockets[key]?.length || 0)
  }, 0)

  // Count free sockets
  const freeSocketsCount = Object.keys(agent.freeSockets || {}).reduce((count, key) => {
    return count + (agent.freeSockets[key]?.length || 0)
  }, 0)

  // Count pending requests
  const requestsCount = Object.keys(agent.requests || {}).reduce((count, key) => {
    return count + (agent.requests[key]?.length || 0)
  }, 0)

  return {
    sockets: socketsCount,
    freeSockets: freeSocketsCount,
    requests: requestsCount,
  }
}

/**
 * Destroy all connections (cleanup)
 */
export function destroyAgents(): void {
  httpsAgent.destroy()
}

/**
 * Get agent configuration
 */
export function getAgentConfig(): {
  keepAlive: boolean
  keepAliveMsecs: number
  maxSockets: number
  maxFreeSockets: number
  requestTimeout: number
} {
  const agent = httpsAgent as any
  return {
    keepAlive: agent.keepAlive ?? false,
    keepAliveMsecs: agent.keepAliveMsecs ?? 1000,
    maxSockets: agent.maxSockets ?? Infinity,
    maxFreeSockets: agent.maxFreeSockets ?? 256,
    requestTimeout: REQUEST_TIMEOUT,
  }
}
