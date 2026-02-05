/**
 * HTTP Agent Configuration
 *
 * Configures connection pooling and HTTP keep-alive to reduce connection overhead.
 * Enables connection reuse across multiple API requests for better performance.
 */

import { Agent } from 'undici'

/**
 * Undici Agent with keep-alive and connection pooling enabled
 * Undici is used instead of native https.Agent because Node.js fetch uses undici under the hood
 */
export const httpsAgent = new Agent({
  // Connection pooling
  connections: parseInt(process.env.NOTION_CLI_HTTP_MAX_SOCKETS || '50', 10),

  // Keep-alive settings
  keepAliveTimeout: parseInt(process.env.NOTION_CLI_HTTP_KEEP_ALIVE_MS || '60000', 10),
  keepAliveMaxTimeout: parseInt(process.env.NOTION_CLI_HTTP_KEEP_ALIVE_MS || '60000', 10),

  // Pipelining (HTTP/1.1 request pipelining, 0 = disabled)
  pipelining: 0,
})

/**
 * Default request timeout in milliseconds
 * Note: timeout is set per-request, not on the agent
 */
export const REQUEST_TIMEOUT = parseInt(process.env.NOTION_CLI_HTTP_TIMEOUT || '30000', 10)

/**
 * Get current agent statistics
 * Note: undici Agent doesn't expose socket statistics like https.Agent
 */
export function getAgentStats(): {
  sockets: number
  freeSockets: number
  requests: number
} {
  // undici's Agent doesn't expose internal socket statistics
  // Return placeholder values for now
  return {
    sockets: 0,
    freeSockets: 0,
    requests: 0,
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
  connections: number
  keepAliveTimeout: number
  requestTimeout: number
} {
  return {
    connections: parseInt(process.env.NOTION_CLI_HTTP_MAX_SOCKETS || '50', 10),
    keepAliveTimeout: parseInt(process.env.NOTION_CLI_HTTP_KEEP_ALIVE_MS || '60000', 10),
    requestTimeout: REQUEST_TIMEOUT,
  }
}
