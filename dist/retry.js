"use strict";
/**
 * Enhanced retry logic with exponential backoff and jitter
 * Handles rate limiting, network errors, and transient failures
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.CircuitBreaker = exports.batchWithRetry = exports.fetchWithRetry = exports.calculateDelay = exports.isRetryableError = void 0;
/**
 * Default retry configuration
 */
const DEFAULT_CONFIG = {
    maxRetries: parseInt(process.env.NOTION_CLI_MAX_RETRIES || '3', 10),
    baseDelay: parseInt(process.env.NOTION_CLI_BASE_DELAY || '1000', 10),
    maxDelay: parseInt(process.env.NOTION_CLI_MAX_DELAY || '30000', 10),
    exponentialBase: parseFloat(process.env.NOTION_CLI_EXP_BASE || '2'),
    jitterFactor: parseFloat(process.env.NOTION_CLI_JITTER_FACTOR || '0.1'),
    // HTTP status codes that should trigger a retry
    retryableStatusCodes: [408, 429, 500, 502, 503, 504],
    // Notion API error codes that are retryable
    retryableErrorCodes: [
        'rate_limited',
        'service_unavailable',
        'internal_server_error',
        'conflict_error',
    ],
};
/**
 * Categorize errors into retryable and non-retryable
 */
function isRetryableError(error, config = DEFAULT_CONFIG) {
    // Network errors (no response)
    if (!error.status && (error.code === 'ECONNRESET' || error.code === 'ETIMEDOUT' ||
        error.code === 'ENOTFOUND' || error.code === 'EAI_AGAIN')) {
        return true;
    }
    // HTTP status codes
    if (error.status && config.retryableStatusCodes.includes(error.status)) {
        return true;
    }
    // Notion API error codes
    if (error.code && config.retryableErrorCodes.includes(error.code)) {
        return true;
    }
    // Don't retry client errors (400-499, except 408 and 429)
    if (error.status >= 400 && error.status < 500 && error.status !== 408 && error.status !== 429) {
        return false;
    }
    return false;
}
exports.isRetryableError = isRetryableError;
/**
 * Calculate delay with exponential backoff and jitter
 */
function calculateDelay(attempt, config = DEFAULT_CONFIG, retryAfterHeader) {
    // If we have a Retry-After header from rate limiting, use it
    if (retryAfterHeader) {
        const retryAfter = parseInt(retryAfterHeader, 10);
        if (!isNaN(retryAfter)) {
            return Math.min(retryAfter * 1000, config.maxDelay);
        }
    }
    // Calculate exponential backoff: baseDelay * (exponentialBase ^ attempt)
    const exponentialDelay = config.baseDelay * Math.pow(config.exponentialBase, attempt - 1);
    // Cap at maxDelay
    const cappedDelay = Math.min(exponentialDelay, config.maxDelay);
    // Add jitter: random value between -jitterFactor and +jitterFactor
    const jitter = cappedDelay * config.jitterFactor * (Math.random() * 2 - 1);
    const finalDelay = Math.max(0, cappedDelay + jitter);
    return Math.round(finalDelay);
}
exports.calculateDelay = calculateDelay;
/**
 * Sleep for specified milliseconds
 */
function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}
/**
 * Enhanced retry wrapper with exponential backoff and jitter
 */
async function fetchWithRetry(fn, options = {}) {
    var _a, _b;
    const config = { ...DEFAULT_CONFIG, ...options.config };
    const { onRetry, context } = options;
    let lastError;
    let totalDelay = 0;
    for (let attempt = 1; attempt <= config.maxRetries + 1; attempt++) {
        try {
            return await fn();
        }
        catch (error) {
            lastError = error;
            // Check if we should retry
            const shouldRetry = attempt <= config.maxRetries && isRetryableError(error, config);
            if (!shouldRetry) {
                throw error;
            }
            // Calculate delay
            const retryAfter = ((_a = error.headers) === null || _a === void 0 ? void 0 : _a['retry-after']) || ((_b = error.headers) === null || _b === void 0 ? void 0 : _b['Retry-After']);
            const delay = calculateDelay(attempt, config, retryAfter);
            totalDelay += delay;
            // Create retry context
            const retryContext = {
                attempt,
                maxRetries: config.maxRetries,
                lastError: error,
                totalDelay,
            };
            // Call retry callback if provided
            if (onRetry) {
                onRetry(retryContext);
            }
            else {
                // Default logging
                const errorType = error.status === 429 ? 'Rate limited' :
                    error.status ? `HTTP ${error.status}` :
                        error.code || 'Network error';
                const contextStr = context ? ` [${context}]` : '';
                console.error(`${errorType}${contextStr}. Retrying in ${delay}ms... ` +
                    `(attempt ${attempt}/${config.maxRetries}, total delay: ${totalDelay}ms)`);
            }
            // Wait before retrying
            await sleep(delay);
        }
    }
    // Should never reach here, but TypeScript needs it
    throw lastError;
}
exports.fetchWithRetry = fetchWithRetry;
/**
 * Batch retry wrapper for multiple operations
 * Executes operations with retry logic and collects results
 */
async function batchWithRetry(operations, options = {}) {
    const { concurrency = 5 } = options;
    const results = [];
    // Process operations in batches
    for (let i = 0; i < operations.length; i += concurrency) {
        const batch = operations.slice(i, i + concurrency);
        const batchPromises = batch.map(async (op, index) => {
            try {
                const data = await fetchWithRetry(op, {
                    ...options,
                    context: `Operation ${i + index + 1}/${operations.length}`,
                });
                return { success: true, data };
            }
            catch (error) {
                return { success: false, error };
            }
        });
        const batchResults = await Promise.all(batchPromises);
        results.push(...batchResults);
    }
    return results;
}
exports.batchWithRetry = batchWithRetry;
/**
 * Retry wrapper with circuit breaker pattern
 * Prevents cascading failures by stopping retries after too many failures
 */
class CircuitBreaker {
    constructor(failureThreshold = 5, successThreshold = 2, timeout = 60000 // 1 minute
    ) {
        this.failureThreshold = failureThreshold;
        this.successThreshold = successThreshold;
        this.timeout = timeout;
        this.failures = 0;
        this.successes = 0;
        this.state = 'closed';
        this.nextAttempt = 0;
    }
    async execute(fn, retryOptions) {
        if (this.state === 'open') {
            if (Date.now() < this.nextAttempt) {
                throw new Error('Circuit breaker is open. Too many failures.');
            }
            this.state = 'half-open';
        }
        try {
            const result = await fetchWithRetry(fn, retryOptions);
            this.onSuccess();
            return result;
        }
        catch (error) {
            this.onFailure();
            throw error;
        }
    }
    onSuccess() {
        this.failures = 0;
        if (this.state === 'half-open') {
            this.successes++;
            if (this.successes >= this.successThreshold) {
                this.state = 'closed';
                this.successes = 0;
            }
        }
    }
    onFailure() {
        this.failures++;
        this.successes = 0;
        if (this.failures >= this.failureThreshold) {
            this.state = 'open';
            this.nextAttempt = Date.now() + this.timeout;
        }
    }
    getState() {
        return {
            state: this.state,
            failures: this.failures,
            successes: this.successes,
        };
    }
    reset() {
        this.state = 'closed';
        this.failures = 0;
        this.successes = 0;
        this.nextAttempt = 0;
    }
}
exports.CircuitBreaker = CircuitBreaker;
