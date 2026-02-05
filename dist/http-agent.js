"use strict";
/**
 * HTTP Agent Configuration
 *
 * Configures connection pooling and HTTP keep-alive to reduce connection overhead.
 * Enables connection reuse across multiple API requests for better performance.
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.REQUEST_TIMEOUT = exports.httpsAgent = void 0;
exports.getAgentStats = getAgentStats;
exports.destroyAgents = destroyAgents;
exports.getAgentConfig = getAgentConfig;
const https = require("https");
/**
 * HTTPS Agent with keep-alive enabled
 */
exports.httpsAgent = new https.Agent({
    // Enable keep-alive connections
    keepAlive: process.env.NOTION_CLI_HTTP_KEEP_ALIVE !== 'false',
    // Keep-alive timeout (how long to keep idle connections open)
    keepAliveMsecs: parseInt(process.env.NOTION_CLI_HTTP_KEEP_ALIVE_MS || '60000', 10),
    // Maximum number of sockets to allow (concurrent connections)
    maxSockets: parseInt(process.env.NOTION_CLI_HTTP_MAX_SOCKETS || '50', 10),
    // Maximum number of sockets to leave open in a free state (connection pool size)
    maxFreeSockets: parseInt(process.env.NOTION_CLI_HTTP_MAX_FREE_SOCKETS || '10', 10),
});
/**
 * Default request timeout in milliseconds
 * Note: timeout is set per-request, not on the agent
 */
exports.REQUEST_TIMEOUT = parseInt(process.env.NOTION_CLI_HTTP_TIMEOUT || '30000', 10);
/**
 * Get current agent statistics
 */
function getAgentStats() {
    const agent = exports.httpsAgent;
    // Count active sockets
    const socketsCount = Object.keys(agent.sockets || {}).reduce((count, key) => {
        var _a;
        return count + (((_a = agent.sockets[key]) === null || _a === void 0 ? void 0 : _a.length) || 0);
    }, 0);
    // Count free sockets
    const freeSocketsCount = Object.keys(agent.freeSockets || {}).reduce((count, key) => {
        var _a;
        return count + (((_a = agent.freeSockets[key]) === null || _a === void 0 ? void 0 : _a.length) || 0);
    }, 0);
    // Count pending requests
    const requestsCount = Object.keys(agent.requests || {}).reduce((count, key) => {
        var _a;
        return count + (((_a = agent.requests[key]) === null || _a === void 0 ? void 0 : _a.length) || 0);
    }, 0);
    return {
        sockets: socketsCount,
        freeSockets: freeSocketsCount,
        requests: requestsCount,
    };
}
/**
 * Destroy all connections (cleanup)
 */
function destroyAgents() {
    exports.httpsAgent.destroy();
}
/**
 * Get agent configuration
 */
function getAgentConfig() {
    var _a, _b, _c, _d;
    const agent = exports.httpsAgent;
    return {
        keepAlive: (_a = agent.keepAlive) !== null && _a !== void 0 ? _a : false,
        keepAliveMsecs: (_b = agent.keepAliveMsecs) !== null && _b !== void 0 ? _b : 1000,
        maxSockets: (_c = agent.maxSockets) !== null && _c !== void 0 ? _c : Infinity,
        maxFreeSockets: (_d = agent.maxFreeSockets) !== null && _d !== void 0 ? _d : 256,
        requestTimeout: exports.REQUEST_TIMEOUT,
    };
}
