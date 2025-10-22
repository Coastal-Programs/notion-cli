/**
 * Enhanced retry logic with exponential backoff and jitter
 * Handles rate limiting, network errors, and transient failures
 */

export interface RetryConfig {
  maxRetries: number
  baseDelay: number
  maxDelay: number
  exponentialBase: number
  jitterFactor: number
  retryableStatusCodes: number[]
  retryableErrorCodes: string[]
}

export interface RetryContext {
  attempt: number
  maxRetries: number
  lastError: any
  totalDelay: number
}

export type RetryCallback = (context: RetryContext) => void

/**
 * Default retry configuration
 */
const DEFAULT_CONFIG: RetryConfig = {
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
}

/**
 * Categorize errors into retryable and non-retryable
 */
export function isRetryableError(error: any, config: RetryConfig = DEFAULT_CONFIG): boolean {
  // Network errors (no response)
  if (!error.status && (error.code === 'ECONNRESET' || error.code === 'ETIMEDOUT' ||
      error.code === 'ENOTFOUND' || error.code === 'EAI_AGAIN')) {
    return true
  }

  // HTTP status codes
  if (error.status && config.retryableStatusCodes.includes(error.status)) {
    return true
  }

  // Notion API error codes
  if (error.code && config.retryableErrorCodes.includes(error.code)) {
    return true
  }

  // Don't retry client errors (400-499, except 408 and 429)
  if (error.status >= 400 && error.status < 500 && error.status !== 408 && error.status !== 429) {
    return false
  }

  return false
}

/**
 * Calculate delay with exponential backoff and jitter
 */
export function calculateDelay(
  attempt: number,
  config: RetryConfig = DEFAULT_CONFIG,
  retryAfterHeader?: string
): number {
  // If we have a Retry-After header from rate limiting, use it
  if (retryAfterHeader) {
    const retryAfter = parseInt(retryAfterHeader, 10)
    if (!isNaN(retryAfter)) {
      return Math.min(retryAfter * 1000, config.maxDelay)
    }
  }

  // Calculate exponential backoff: baseDelay * (exponentialBase ^ attempt)
  const exponentialDelay = config.baseDelay * Math.pow(config.exponentialBase, attempt - 1)

  // Cap at maxDelay
  const cappedDelay = Math.min(exponentialDelay, config.maxDelay)

  // Add jitter: random value between -jitterFactor and +jitterFactor
  const jitter = cappedDelay * config.jitterFactor * (Math.random() * 2 - 1)
  const finalDelay = Math.max(0, cappedDelay + jitter)

  return Math.round(finalDelay)
}

/**
 * Sleep for specified milliseconds
 */
function sleep(ms: number): Promise<void> {
  return new Promise(resolve => setTimeout(resolve, ms))
}

/**
 * Enhanced retry wrapper with exponential backoff and jitter
 */
export async function fetchWithRetry<T>(
  fn: () => Promise<T>,
  options: {
    config?: Partial<RetryConfig>
    onRetry?: RetryCallback
    context?: string
  } = {}
): Promise<T> {
  const config: RetryConfig = { ...DEFAULT_CONFIG, ...options.config }
  const { onRetry, context } = options

  let lastError: any
  let totalDelay = 0

  for (let attempt = 1; attempt <= config.maxRetries + 1; attempt++) {
    try {
      return await fn()
    } catch (error: any) {
      lastError = error

      // Check if we should retry
      const shouldRetry = attempt <= config.maxRetries && isRetryableError(error, config)

      if (!shouldRetry) {
        throw error
      }

      // Calculate delay
      const retryAfter = error.headers?.['retry-after'] || error.headers?.['Retry-After']
      const delay = calculateDelay(attempt, config, retryAfter)
      totalDelay += delay

      // Create retry context
      const retryContext: RetryContext = {
        attempt,
        maxRetries: config.maxRetries,
        lastError: error,
        totalDelay,
      }

      // Call retry callback if provided
      if (onRetry) {
        onRetry(retryContext)
      } else {
        // Default logging
        const errorType = error.status === 429 ? 'Rate limited' :
                         error.status ? `HTTP ${error.status}` :
                         error.code || 'Network error'
        const contextStr = context ? ` [${context}]` : ''
        console.error(
          `${errorType}${contextStr}. Retrying in ${delay}ms... ` +
          `(attempt ${attempt}/${config.maxRetries}, total delay: ${totalDelay}ms)`
        )
      }

      // Wait before retrying
      await sleep(delay)
    }
  }

  // Should never reach here, but TypeScript needs it
  throw lastError
}

/**
 * Batch retry wrapper for multiple operations
 * Executes operations with retry logic and collects results
 */
export async function batchWithRetry<T>(
  operations: Array<() => Promise<T>>,
  options: {
    config?: Partial<RetryConfig>
    onRetry?: RetryCallback
    concurrency?: number
  } = {}
): Promise<Array<{ success: boolean; data?: T; error?: any }>> {
  const { concurrency = 5 } = options
  const results: Array<{ success: boolean; data?: T; error?: any }> = []

  // Process operations in batches
  for (let i = 0; i < operations.length; i += concurrency) {
    const batch = operations.slice(i, i + concurrency)
    const batchPromises = batch.map(async (op, index) => {
      try {
        const data = await fetchWithRetry(op, {
          ...options,
          context: `Operation ${i + index + 1}/${operations.length}`,
        })
        return { success: true, data }
      } catch (error) {
        return { success: false, error }
      }
    })

    const batchResults = await Promise.all(batchPromises)
    results.push(...batchResults)
  }

  return results
}

/**
 * Retry wrapper with circuit breaker pattern
 * Prevents cascading failures by stopping retries after too many failures
 */
export class CircuitBreaker {
  private failures = 0
  private successes = 0
  private state: 'closed' | 'open' | 'half-open' = 'closed'
  private nextAttempt = 0

  constructor(
    private readonly failureThreshold = 5,
    private readonly successThreshold = 2,
    private readonly timeout = 60000 // 1 minute
  ) {}

  async execute<T>(fn: () => Promise<T>, retryOptions?: Parameters<typeof fetchWithRetry>[1]): Promise<T> {
    if (this.state === 'open') {
      if (Date.now() < this.nextAttempt) {
        throw new Error('Circuit breaker is open. Too many failures.')
      }
      this.state = 'half-open'
    }

    try {
      const result = await fetchWithRetry(fn, retryOptions)
      this.onSuccess()
      return result
    } catch (error) {
      this.onFailure()
      throw error
    }
  }

  private onSuccess(): void {
    this.failures = 0
    if (this.state === 'half-open') {
      this.successes++
      if (this.successes >= this.successThreshold) {
        this.state = 'closed'
        this.successes = 0
      }
    }
  }

  private onFailure(): void {
    this.failures++
    this.successes = 0
    if (this.failures >= this.failureThreshold) {
      this.state = 'open'
      this.nextAttempt = Date.now() + this.timeout
    }
  }

  getState(): { state: string; failures: number; successes: number } {
    return {
      state: this.state,
      failures: this.failures,
      successes: this.successes,
    }
  }

  reset(): void {
    this.state = 'closed'
    this.failures = 0
    this.successes = 0
    this.nextAttempt = 0
  }
}
