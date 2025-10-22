/**
 * Enhanced retry logic with exponential backoff and jitter
 * Handles rate limiting, network errors, and transient failures
 */
export interface RetryConfig {
    maxRetries: number;
    baseDelay: number;
    maxDelay: number;
    exponentialBase: number;
    jitterFactor: number;
    retryableStatusCodes: number[];
    retryableErrorCodes: string[];
}
export interface RetryContext {
    attempt: number;
    maxRetries: number;
    lastError: any;
    totalDelay: number;
}
export type RetryCallback = (context: RetryContext) => void;
/**
 * Categorize errors into retryable and non-retryable
 */
export declare function isRetryableError(error: any, config?: RetryConfig): boolean;
/**
 * Calculate delay with exponential backoff and jitter
 */
export declare function calculateDelay(attempt: number, config?: RetryConfig, retryAfterHeader?: string): number;
/**
 * Enhanced retry wrapper with exponential backoff and jitter
 */
export declare function fetchWithRetry<T>(fn: () => Promise<T>, options?: {
    config?: Partial<RetryConfig>;
    onRetry?: RetryCallback;
    context?: string;
}): Promise<T>;
/**
 * Batch retry wrapper for multiple operations
 * Executes operations with retry logic and collects results
 */
export declare function batchWithRetry<T>(operations: Array<() => Promise<T>>, options?: {
    config?: Partial<RetryConfig>;
    onRetry?: RetryCallback;
    concurrency?: number;
}): Promise<Array<{
    success: boolean;
    data?: T;
    error?: any;
}>>;
/**
 * Retry wrapper with circuit breaker pattern
 * Prevents cascading failures by stopping retries after too many failures
 */
export declare class CircuitBreaker {
    private readonly failureThreshold;
    private readonly successThreshold;
    private readonly timeout;
    private failures;
    private successes;
    private state;
    private nextAttempt;
    constructor(failureThreshold?: number, successThreshold?: number, timeout?: number);
    execute<T>(fn: () => Promise<T>, retryOptions?: Parameters<typeof fetchWithRetry>[1]): Promise<T>;
    private onSuccess;
    private onFailure;
    getState(): {
        state: string;
        failures: number;
        successes: number;
    };
    reset(): void;
}
