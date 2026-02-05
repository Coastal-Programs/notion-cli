/**
 * HTTP Agent Configuration
 *
 * Configures connection pooling and HTTP keep-alive to reduce connection overhead.
 * Enables connection reuse across multiple API requests for better performance.
 */
import { Agent } from 'undici';
/**
 * Undici Agent with keep-alive and connection pooling enabled
 * Undici is used instead of native https.Agent because Node.js fetch uses undici under the hood
 */
export declare const httpsAgent: Agent;
/**
 * Default request timeout in milliseconds
 * Note: timeout is set per-request, not on the agent
 */
export declare const REQUEST_TIMEOUT: number;
/**
 * Get current agent statistics
 * Note: undici Agent doesn't expose socket statistics like https.Agent
 */
export declare function getAgentStats(): {
    sockets: number;
    freeSockets: number;
    requests: number;
};
/**
 * Destroy all connections (cleanup)
 */
export declare function destroyAgents(): void;
/**
 * Get agent configuration
 */
export declare function getAgentConfig(): {
    connections: number;
    keepAliveTimeout: number;
    requestTimeout: number;
};
