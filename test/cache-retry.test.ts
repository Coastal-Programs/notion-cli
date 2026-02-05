/**
 * Unit tests for enhanced retry logic and caching layer
 */

import { expect } from 'chai'
import { CacheManager } from '../dist/cache.js'
import {
  calculateDelay,
  isRetryableError,
  fetchWithRetry,
  CircuitBreaker,
} from '../src/retry'

describe('Cache Manager', () => {
  let cache: CacheManager
  let originalDiskCacheEnabled: string | undefined

  before(() => {
    // Disable disk cache for unit tests
    originalDiskCacheEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED
    process.env.NOTION_CLI_DISK_CACHE_ENABLED = 'false'
  })

  after(() => {
    // Restore original disk cache setting
    if (originalDiskCacheEnabled === undefined) {
      delete process.env.NOTION_CLI_DISK_CACHE_ENABLED
    } else {
      process.env.NOTION_CLI_DISK_CACHE_ENABLED = originalDiskCacheEnabled
    }
  })

  beforeEach(() => {
    // Create a fresh cache instance for each test
    cache = new CacheManager({
      enabled: true,
      defaultTtl: 1000,
      maxSize: 10,
      ttlByType: {
        dataSource: 1000,
        database: 1000,
        user: 1000,
        page: 1000,
        block: 1000,
      },
    })
  })

  describe('Basic Operations', () => {
    it('should store and retrieve values', async () => {
      const data = { id: '123', name: 'test' }
      cache.set('dataSource', data, undefined, '123')

      const retrieved = await cache.get('dataSource', '123')
      expect(retrieved).to.deep.equal(data)
    })

    it('should return null for non-existent keys', async () => {
      const retrieved = await cache.get('dataSource', 'non-existent')
      expect(retrieved).to.be.null
    })

    it('should handle multiple identifiers', async () => {
      const data = { content: 'test' }
      cache.set('block', data, undefined, 'parent-id', 'block-id')

      const retrieved = await cache.get('block', 'parent-id', 'block-id')
      expect(retrieved).to.deep.equal(data)
    })
  })

  describe('TTL (Time-to-Live)', () => {
    it('should expire entries after TTL', async () => {
      const data = { id: '123' }
      cache.set('dataSource', data, 100, '123') // 100ms TTL

      // Should be available immediately
      let retrieved = await cache.get('dataSource', '123')
      expect(retrieved).to.not.be.null

      // Wait for TTL to expire
      await new Promise(resolve => setTimeout(resolve, 150))

      // Should be expired
      retrieved = await cache.get('dataSource', '123')
      expect(retrieved).to.be.null
    })

    it('should use custom TTL when provided', async () => {
      const data = { id: '123' }
      cache.set('dataSource', data, 50, '123') // Custom 50ms TTL

      await new Promise(resolve => setTimeout(resolve, 75))

      const retrieved = await cache.get('dataSource', '123')
      expect(retrieved).to.be.null
    })
  })

  describe('Cache Invalidation', () => {
    it('should invalidate specific entries', async () => {
      cache.set('dataSource', { id: '1' }, undefined, '1')
      cache.set('dataSource', { id: '2' }, undefined, '2')

      cache.invalidate('dataSource', '1')

      expect(await cache.get('dataSource', '1')).to.be.null
      expect(await cache.get('dataSource', '2')).to.not.be.null
    })

    it('should invalidate all entries of a type', async () => {
      cache.set('dataSource', { id: '1' }, undefined, '1')
      cache.set('dataSource', { id: '2' }, undefined, '2')
      cache.set('user', { id: '3' }, undefined, '3')

      cache.invalidate('dataSource')

      expect(await cache.get('dataSource', '1')).to.be.null
      expect(await cache.get('dataSource', '2')).to.be.null
      expect(await cache.get('user', '3')).to.not.be.null
    })

    it('should clear all entries', async () => {
      cache.set('dataSource', { id: '1' }, undefined, '1')
      cache.set('user', { id: '2' }, undefined, '2')

      cache.clear()

      expect(await cache.get('dataSource', '1')).to.be.null
      expect(await cache.get('user', '2')).to.be.null
      expect(cache.getStats().size).to.equal(0)
    })
  })

  describe('Cache Statistics', () => {
    it('should track hits and misses', async () => {
      cache.set('dataSource', { id: '1' }, undefined, '1')

      // Hit
      await cache.get('dataSource', '1')
      // Miss
      await cache.get('dataSource', '2')
      // Hit
      await cache.get('dataSource', '1')

      const stats = cache.getStats()
      expect(stats.hits).to.equal(2)
      expect(stats.misses).to.equal(1)
    })

    it('should calculate hit rate correctly', async () => {
      cache.set('dataSource', { id: '1' }, undefined, '1')

      await cache.get('dataSource', '1') // Hit
      await cache.get('dataSource', '1') // Hit
      await cache.get('dataSource', '2') // Miss

      const hitRate = cache.getHitRate()
      expect(hitRate).to.be.closeTo(0.667, 0.01) // 2/3
    })

    it('should track cache size', () => {
      expect(cache.getStats().size).to.equal(0)

      cache.set('dataSource', { id: '1' }, undefined, '1')
      expect(cache.getStats().size).to.equal(1)

      cache.set('dataSource', { id: '2' }, undefined, '2')
      expect(cache.getStats().size).to.equal(2)

      cache.invalidate('dataSource', '1')
      expect(cache.getStats().size).to.equal(1)
    })
  })

  describe('Size Limits', () => {
    it('should evict oldest entries when full', async () => {
      // Fill cache to capacity (10 entries)
      for (let i = 0; i < 10; i++) {
        cache.set('dataSource', { id: i }, undefined, String(i))
      }

      expect(cache.getStats().size).to.equal(10)

      // Add one more - should evict oldest
      cache.set('dataSource', { id: 10 }, undefined, '10')

      expect(cache.getStats().size).to.equal(10)
      expect(await cache.get('dataSource', '0')).to.be.null // Oldest evicted
      expect(await cache.get('dataSource', '10')).to.not.be.null // Newest exists
    })
  })

  describe('Configuration', () => {
    it('should respect enabled flag', async () => {
      const disabledCache = new CacheManager({ enabled: false })

      disabledCache.set('dataSource', { id: '1' }, undefined, '1')
      const retrieved = await disabledCache.get('dataSource', '1')

      expect(retrieved).to.be.null
    })

    it('should return configuration', () => {
      const config = cache.getConfig()

      expect(config.enabled).to.be.true
      expect(config.defaultTtl).to.equal(1000)
      expect(config.maxSize).to.equal(10)
    })

    it('should report enabled status', () => {
      expect(cache.isEnabled()).to.be.true

      const disabledCache = new CacheManager({ enabled: false })
      expect(disabledCache.isEnabled()).to.be.false
    })
  })
})

describe('Retry Logic', () => {
  describe('Error Categorization', () => {
    it('should identify retryable HTTP status codes', () => {
      expect(isRetryableError({ status: 429 })).to.be.true // Rate limit
      expect(isRetryableError({ status: 408 })).to.be.true // Timeout
      expect(isRetryableError({ status: 500 })).to.be.true // Server error
      expect(isRetryableError({ status: 502 })).to.be.true // Bad gateway
      expect(isRetryableError({ status: 503 })).to.be.true // Service unavailable
      expect(isRetryableError({ status: 504 })).to.be.true // Gateway timeout
    })

    it('should identify non-retryable HTTP status codes', () => {
      expect(isRetryableError({ status: 400 })).to.be.false // Bad request
      expect(isRetryableError({ status: 401 })).to.be.false // Unauthorized
      expect(isRetryableError({ status: 403 })).to.be.false // Forbidden
      expect(isRetryableError({ status: 404 })).to.be.false // Not found
    })

    it('should identify retryable network errors', () => {
      expect(isRetryableError({ code: 'ECONNRESET' })).to.be.true
      expect(isRetryableError({ code: 'ETIMEDOUT' })).to.be.true
      expect(isRetryableError({ code: 'ENOTFOUND' })).to.be.true
      expect(isRetryableError({ code: 'EAI_AGAIN' })).to.be.true
    })

    it('should identify retryable Notion API errors', () => {
      expect(isRetryableError({ code: 'rate_limited' })).to.be.true
      expect(isRetryableError({ code: 'service_unavailable' })).to.be.true
      expect(isRetryableError({ code: 'internal_server_error' })).to.be.true
      expect(isRetryableError({ code: 'conflict_error' })).to.be.true
    })
  })

  describe('Delay Calculation', () => {
    it('should calculate exponential backoff', () => {
      const config = {
        maxRetries: 5,
        baseDelay: 1000,
        maxDelay: 30000,
        exponentialBase: 2,
        jitterFactor: 0,
        retryableStatusCodes: [],
        retryableErrorCodes: [],
      }

      const delay1 = calculateDelay(1, config)
      const delay2 = calculateDelay(2, config)
      const delay3 = calculateDelay(3, config)

      expect(delay1).to.equal(1000) // 1000 * 2^0
      expect(delay2).to.equal(2000) // 1000 * 2^1
      expect(delay3).to.equal(4000) // 1000 * 2^2
    })

    it('should respect max delay cap', () => {
      const config = {
        maxRetries: 10,
        baseDelay: 1000,
        maxDelay: 5000,
        exponentialBase: 2,
        jitterFactor: 0,
        retryableStatusCodes: [],
        retryableErrorCodes: [],
      }

      const delay = calculateDelay(10, config) // Would be 512000 without cap
      expect(delay).to.equal(5000)
    })

    it('should respect Retry-After header', () => {
      const config = {
        maxRetries: 5,
        baseDelay: 1000,
        maxDelay: 30000,
        exponentialBase: 2,
        jitterFactor: 0,
        retryableStatusCodes: [],
        retryableErrorCodes: [],
      }

      const delay = calculateDelay(1, config, '5') // 5 seconds
      expect(delay).to.equal(5000)
    })

    it('should add jitter to delays', () => {
      const config = {
        maxRetries: 5,
        baseDelay: 1000,
        maxDelay: 30000,
        exponentialBase: 2,
        jitterFactor: 0.2, // 20% jitter
        retryableStatusCodes: [],
        retryableErrorCodes: [],
      }

      // Calculate multiple times and verify variance
      const delays = []
      for (let i = 0; i < 10; i++) {
        delays.push(calculateDelay(1, config))
      }

      // Should have some variance due to jitter
      const allSame = delays.every(d => d === delays[0])
      expect(allSame).to.be.false

      // All should be within expected range (1000 ± 200)
      delays.forEach(delay => {
        expect(delay).to.be.at.least(800)
        expect(delay).to.be.at.most(1200)
      })
    })
  })

  describe('Fetch with Retry', () => {
    it('should succeed on first attempt', async () => {
      let attempts = 0
      const result = await fetchWithRetry(async () => {
        attempts++
        return 'success'
      })

      expect(result).to.equal('success')
      expect(attempts).to.equal(1)
    })

    it('should retry on retryable errors', async () => {
      let attempts = 0
      const result = await fetchWithRetry(
        async () => {
          attempts++
          if (attempts < 3) {
            const error: any = new Error('Service unavailable')
            error.status = 503
            throw error
          }
          return 'success'
        },
        {
          config: {
            maxRetries: 3,
            baseDelay: 10,
            maxDelay: 100,
            exponentialBase: 2,
            jitterFactor: 0,
            retryableStatusCodes: [503],
            retryableErrorCodes: [],
          },
        }
      )

      expect(result).to.equal('success')
      expect(attempts).to.equal(3)
    })

    it('should not retry on non-retryable errors', async () => {
      let attempts = 0
      try {
        await fetchWithRetry(
          async () => {
            attempts++
            const error: any = new Error('Bad request')
            error.status = 400
            throw error
          },
          {
            config: {
              maxRetries: 3,
              baseDelay: 10,
              maxDelay: 100,
              exponentialBase: 2,
              jitterFactor: 0,
              retryableStatusCodes: [503],
              retryableErrorCodes: [],
            },
          }
        )
        expect.fail('Should have thrown')
      } catch (error: any) {
        expect(error.status).to.equal(400)
        expect(attempts).to.equal(1) // No retries
      }
    })

    it('should call onRetry callback', async () => {
      const retryContexts: any[] = []

      try {
        await fetchWithRetry(
          async () => {
            const error: any = new Error('Service unavailable')
            error.status = 503
            throw error
          },
          {
            config: {
              maxRetries: 2,
              baseDelay: 10,
              maxDelay: 100,
              exponentialBase: 2,
              jitterFactor: 0,
              retryableStatusCodes: [503],
              retryableErrorCodes: [],
            },
            onRetry: (context) => {
              retryContexts.push(context)
            },
          }
        )
      } catch {
        // Expected to fail
      }

      expect(retryContexts).to.have.lengthOf(2)
      expect(retryContexts[0].attempt).to.equal(1)
      expect(retryContexts[1].attempt).to.equal(2)
    })
  })

  describe('Circuit Breaker', () => {
    it('should start in closed state', () => {
      const breaker = new CircuitBreaker(3, 2, 1000)
      const state = breaker.getState()

      expect(state.state).to.equal('closed')
      expect(state.failures).to.equal(0)
    })

    it('should open after threshold failures', async () => {
      const breaker = new CircuitBreaker(3, 2, 1000)

      // Cause 3 failures
      for (let i = 0; i < 3; i++) {
        try {
          await breaker.execute(async () => {
            throw new Error('Failure')
          })
        } catch {
          // Expected
        }
      }

      const state = breaker.getState()
      expect(state.state).to.equal('open')
      expect(state.failures).to.equal(3)
    })

    it('should reject requests when open', async () => {
      const breaker = new CircuitBreaker(2, 2, 1000)

      // Cause failures to open circuit
      for (let i = 0; i < 2; i++) {
        try {
          await breaker.execute(async () => {
            throw new Error('Failure')
          })
        } catch {
          // Expected
        }
      }

      // Should reject immediately when open
      try {
        await breaker.execute(async () => {
          return 'success'
        })
        expect.fail('Should have thrown')
      } catch (error: any) {
        expect(error.message).to.include('Circuit breaker is open')
      }
    })

    it('should reset failures on success', async () => {
      const breaker = new CircuitBreaker(3, 2, 1000)

      // Cause 2 failures
      for (let i = 0; i < 2; i++) {
        try {
          await breaker.execute(async () => {
            throw new Error('Failure')
          })
        } catch {
          // Expected
        }
      }

      // Success should reset
      await breaker.execute(async () => {
        return 'success'
      })

      const state = breaker.getState()
      expect(state.failures).to.equal(0)
    })

    it('should allow manual reset', async () => {
      const breaker = new CircuitBreaker(2, 2, 1000)

      // Open the circuit
      for (let i = 0; i < 2; i++) {
        try {
          await breaker.execute(async () => {
            throw new Error('Failure')
          })
        } catch {
          // Expected
        }
      }

      expect(breaker.getState().state).to.equal('open')

      // Reset
      breaker.reset()

      const state = breaker.getState()
      expect(state.state).to.equal('closed')
      expect(state.failures).to.equal(0)
    })
  })
})
