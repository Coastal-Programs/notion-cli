"use strict";
/**
 * Enhanced retry logic with exponential backoff and jitter
 * Handles rate limiting, network errors, and transient failures
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.CircuitBreaker = void 0;
exports.isRetryableError = isRetryableError;
exports.calculateDelay = calculateDelay;
exports.fetchWithRetry = fetchWithRetry;
exports.batchWithRetry = batchWithRetry;
/**
 * Default retry configuration
 */
const DEFAULT_CONFIG = {
    maxRetries: parseInt(process.env.NOTION_CLI_MAX_RETRIES || '3', 10),
    baseDelay: parseInt(process.env.NOTION_CLI_BASE_DELAY || '1000', 10), // 1 second
    maxDelay: parseInt(process.env.NOTION_CLI_MAX_DELAY || '30000', 10), // 30 seconds
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
 * Check if verbose logging is enabled
 */
function isVerboseEnabled() {
    return process.env.DEBUG === 'true' ||
        process.env.NOTION_CLI_DEBUG === 'true' ||
        process.env.NOTION_CLI_VERBOSE === 'true';
}
/**
 * Log structured retry event to stderr
 * Never pollutes stdout - safe for JSON output
 */
function logRetryEvent(event) {
    // Only log if verbose mode is enabled
    if (!isVerboseEnabled()) {
        return;
    }
    // Always write to stderr, never stdout
    console.error(JSON.stringify(event));
}
/**
 * Extract error reason from error object
 */
function getErrorReason(error) {
    if (error.code === 'rate_limited' || error.status === 429)
        return 'RATE_LIMITED';
    if (error.status === 503)
        return 'SERVICE_UNAVAILABLE';
    if (error.status === 502)
        return 'BAD_GATEWAY';
    if (error.status === 504)
        return 'GATEWAY_TIMEOUT';
    if (error.status === 500)
        return 'INTERNAL_SERVER_ERROR';
    if (error.status === 408)
        return 'REQUEST_TIMEOUT';
    if (error.code === 'ECONNRESET')
        return 'CONNECTION_RESET';
    if (error.code === 'ETIMEDOUT')
        return 'TIMEOUT';
    if (error.code === 'ENOTFOUND')
        return 'DNS_ERROR';
    if (error.code === 'EAI_AGAIN')
        return 'DNS_LOOKUP_FAILED';
    if (error.code === 'service_unavailable')
        return 'SERVICE_UNAVAILABLE';
    if (error.code === 'internal_server_error')
        return 'INTERNAL_SERVER_ERROR';
    if (error.code === 'conflict_error')
        return 'CONFLICT';
    return 'UNKNOWN';
}
/**
 * Extract URL/endpoint from error object
 */
function extractUrl(error, context) {
    var _a, _b;
    if (error.url)
        return error.url;
    if ((_a = error.request) === null || _a === void 0 ? void 0 : _a.url)
        return error.request.url;
    if ((_b = error.config) === null || _b === void 0 ? void 0 : _b.url)
        return error.config.url;
    return context;
}
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
            // Log attempt start (if verbose and not first attempt)
            if (attempt > 1 && isVerboseEnabled()) {
                logRetryEvent({
                    level: 'info',
                    event: 'retry_attempt',
                    attempt,
                    max_retries: config.maxRetries,
                    context,
                    timestamp: new Date().toISOString(),
                });
            }
            return await fn();
        }
        catch (error) {
            lastError = error;
            // Check if we should retry
            const shouldRetry = attempt <= config.maxRetries && isRetryableError(error, config);
            if (!shouldRetry) {
                // Log non-retryable error if verbose
                if (isVerboseEnabled() && attempt > 1) {
                    logRetryEvent({
                        level: 'error',
                        event: 'retry_exhausted',
                        attempt,
                        max_retries: config.maxRetries,
                        reason: getErrorReason(error),
                        context,
                        status_code: error.status,
                        error_code: error.code,
                        timestamp: new Date().toISOString(),
                    });
                }
                throw error;
            }
            // Calculate delay
            const retryAfter = ((_a = error.headers) === null || _a === void 0 ? void 0 : _a['retry-after']) || ((_b = error.headers) === null || _b === void 0 ? void 0 : _b['Retry-After']);
            const delay = calculateDelay(attempt, config, retryAfter);
            totalDelay += delay;
            // Log rate limit event specifically
            if (error.status === 429 || error.code === 'rate_limited') {
                logRetryEvent({
                    level: 'warn',
                    event: 'rate_limited',
                    attempt,
                    max_retries: config.maxRetries,
                    reason: 'RATE_LIMITED',
                    retry_after_ms: delay,
                    url: extractUrl(error, context),
                    context,
                    status_code: error.status,
                    timestamp: new Date().toISOString(),
                });
            }
            else {
                // Log general retry event
                logRetryEvent({
                    level: 'warn',
                    event: 'retry',
                    attempt,
                    max_retries: config.maxRetries,
                    reason: getErrorReason(error),
                    retry_after_ms: delay,
                    url: extractUrl(error, context),
                    context,
                    status_code: error.status,
                    error_code: error.code,
                    timestamp: new Date().toISOString(),
                });
            }
            // Create retry context
            const retryContext = {
                attempt,
                maxRetries: config.maxRetries,
                lastError: error,
                totalDelay,
            };
            // Call retry callback if provided (for custom logging/monitoring)
            if (onRetry) {
                onRetry(retryContext);
            }
            // Wait before retrying
            await sleep(delay);
        }
    }
    // Should never reach here, but TypeScript needs it
    throw lastError;
}
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
                // Log circuit breaker open event
                if (isVerboseEnabled()) {
                    logRetryEvent({
                        level: 'error',
                        event: 'retry_exhausted',
                        attempt: 0,
                        max_retries: 0,
                        reason: 'CIRCUIT_OPEN',
                        context: 'Circuit breaker is open',
                        timestamp: new Date().toISOString(),
                    });
                }
                throw new Error('Circuit breaker is open. Too many failures.');
            }
            this.state = 'half-open';
            // Log circuit breaker half-open event
            if (isVerboseEnabled()) {
                logRetryEvent({
                    level: 'info',
                    event: 'retry_attempt',
                    attempt: 1,
                    max_retries: this.successThreshold,
                    context: 'Circuit breaker entering half-open state',
                    timestamp: new Date().toISOString(),
                });
            }
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
                // Log circuit breaker closed event
                if (isVerboseEnabled()) {
                    logRetryEvent({
                        level: 'info',
                        event: 'retry_attempt',
                        attempt: this.successThreshold,
                        max_retries: this.successThreshold,
                        context: 'Circuit breaker closed - service recovered',
                        timestamp: new Date().toISOString(),
                    });
                }
            }
        }
    }
    onFailure() {
        this.failures++;
        this.successes = 0;
        if (this.failures >= this.failureThreshold) {
            this.state = 'open';
            this.nextAttempt = Date.now() + this.timeout;
            // Log circuit breaker open event
            if (isVerboseEnabled()) {
                logRetryEvent({
                    level: 'error',
                    event: 'retry_exhausted',
                    attempt: this.failures,
                    max_retries: this.failureThreshold,
                    reason: 'CIRCUIT_OPENED',
                    retry_after_ms: this.timeout,
                    context: `Circuit breaker opened after ${this.failures} failures`,
                    timestamp: new Date().toISOString(),
                });
            }
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
