/**
 * HTTP Agent Configuration
 *
 * Configures connection pooling and HTTP keep-alive to reduce connection overhead.
 * Enables connection reuse across multiple API requests for better performance.
 */
import * as https from 'https';
/**
 * HTTPS Agent with keep-alive enabled
 */
export declare const httpsAgent: https.Agent;
/**
 * Default request timeout in milliseconds
 * Note: timeout is set per-request, not on the agent
 */
export declare const REQUEST_TIMEOUT: number;
/**
 * Get current agent statistics
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
    keepAlive: boolean;
    keepAliveMsecs: number;
    maxSockets: number;
    maxFreeSockets: number;
    requestTimeout: number;
};
