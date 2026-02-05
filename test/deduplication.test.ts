import { expect } from 'chai'
import { DeduplicationManager, deduplicationManager } from '../src/deduplication'

describe('DeduplicationManager', () => {
  let dedup: DeduplicationManager

  beforeEach(() => {
    dedup = new DeduplicationManager()
  })

  afterEach(() => {
    dedup.clear()
  })

  describe('execute()', () => {
    it('should deduplicate concurrent requests with same key', async () => {
      let callCount = 0
      const fn = async () => {
        callCount++
        await new Promise(resolve => setTimeout(resolve, 100))
        return 'result'
      }

      // Execute three concurrent requests with same key
      const [r1, r2, r3] = await Promise.all([
        dedup.execute('key1', fn),
        dedup.execute('key1', fn),
        dedup.execute('key1', fn),
      ])

      expect(callCount).to.equal(1, 'Function should only be called once')
      expect(r1).to.equal('result')
      expect(r2).to.equal('result')
      expect(r3).to.equal('result')
      expect(r1).to.equal(r2, 'All results should be identical')
      expect(r2).to.equal(r3, 'All results should be identical')
    })

    it('should not deduplicate requests with different keys', async () => {
      const calls: string[] = []
      const fn = (key: string) => async () => {
        calls.push(key)
        await new Promise(resolve => setTimeout(resolve, 50))
        return `result-${key}`
      }

      // Execute concurrent requests with different keys
      const [r1, r2, r3] = await Promise.all([
        dedup.execute('key1', fn('key1')),
        dedup.execute('key2', fn('key2')),
        dedup.execute('key3', fn('key3')),
      ])

      expect(calls.length).to.equal(3, 'Function should be called three times')
      expect(r1).to.equal('result-key1')
      expect(r2).to.equal('result-key2')
      expect(r3).to.equal('result-key3')
    })

    it('should not deduplicate sequential requests with same key', async () => {
      let callCount = 0
      const fn = async () => {
        callCount++
        await new Promise(resolve => setTimeout(resolve, 50))
        return `result-${callCount}`
      }

      // Execute sequential requests
      const r1 = await dedup.execute('key1', fn)
      const r2 = await dedup.execute('key1', fn)
      const r3 = await dedup.execute('key1', fn)

      expect(callCount).to.equal(3, 'Function should be called three times')
      expect(r1).to.equal('result-1')
      expect(r2).to.equal('result-2')
      expect(r3).to.equal('result-3')
    })

    it('should propagate errors to all waiting callers', async () => {
      const error = new Error('Test error')
      const fn = async () => {
        await new Promise(resolve => setTimeout(resolve, 50))
        throw error
      }

      // Execute concurrent requests
      const promises = [
        dedup.execute('key1', fn),
        dedup.execute('key1', fn),
        dedup.execute('key1', fn),
      ]

      // All should reject with same error
      const results = await Promise.allSettled(promises)

      expect(results[0].status).to.equal('rejected')
      expect(results[1].status).to.equal('rejected')
      expect(results[2].status).to.equal('rejected')

      if (results[0].status === 'rejected' &&
          results[1].status === 'rejected' &&
          results[2].status === 'rejected') {
        expect(results[0].reason).to.equal(error)
        expect(results[1].reason).to.equal(error)
        expect(results[2].reason).to.equal(error)
      }
    })

    it('should clean up pending entry after promise resolves', async () => {
      const fn = async () => {
        await new Promise(resolve => setTimeout(resolve, 50))
        return 'result'
      }

      expect(dedup.getStats().pending).to.equal(0)

      const promise = dedup.execute('key1', fn)
      expect(dedup.getStats().pending).to.equal(1, 'Should have one pending request')

      await promise
      expect(dedup.getStats().pending).to.equal(0, 'Should clean up after resolution')
    })

    it('should clean up pending entry after promise rejects', async () => {
      const fn = async () => {
        await new Promise(resolve => setTimeout(resolve, 50))
        throw new Error('Test error')
      }

      expect(dedup.getStats().pending).to.equal(0)

      const promise = dedup.execute('key1', fn)
      expect(dedup.getStats().pending).to.equal(1, 'Should have one pending request')

      try {
        await promise
      } catch {
        // Expected error
      }

      expect(dedup.getStats().pending).to.equal(0, 'Should clean up after rejection')
    })

    it('should handle different types of return values', async () => {
      // String
      const r1 = await dedup.execute('key1', async () => 'string')
      expect(r1).to.equal('string')

      // Number
      const r2 = await dedup.execute('key2', async () => 42)
      expect(r2).to.equal(42)

      // Object
      const obj = { foo: 'bar' }
      const r3 = await dedup.execute('key3', async () => obj)
      expect(r3).to.deep.equal(obj)

      // Array
      const arr = [1, 2, 3]
      const r4 = await dedup.execute('key4', async () => arr)
      expect(r4).to.deep.equal(arr)

      // Null
      const r5 = await dedup.execute('key5', async () => null)
      expect(r5).to.be.null

      // Undefined
      const r6 = await dedup.execute('key6', async () => undefined)
      expect(r6).to.be.undefined
    })
  })

  describe('getStats()', () => {
    it('should track hits correctly', async () => {
      const fn = async () => {
        await new Promise(resolve => setTimeout(resolve, 100))
        return 'result'
      }

      expect(dedup.getStats().hits).to.equal(0)

      // First request is a miss
      const p1 = dedup.execute('key1', fn)
      expect(dedup.getStats().hits).to.equal(0)
      expect(dedup.getStats().misses).to.equal(1)

      // Concurrent requests are hits
      const p2 = dedup.execute('key1', fn)
      const p3 = dedup.execute('key1', fn)

      expect(dedup.getStats().hits).to.equal(2)
      expect(dedup.getStats().misses).to.equal(1)

      await Promise.all([p1, p2, p3])
    })

    it('should track misses correctly', async () => {
      const fn = async () => {
        await new Promise(resolve => setTimeout(resolve, 50))
        return 'result'
      }

      expect(dedup.getStats().misses).to.equal(0)

      await dedup.execute('key1', fn)
      expect(dedup.getStats().misses).to.equal(1)

      await dedup.execute('key2', fn)
      expect(dedup.getStats().misses).to.equal(2)

      await dedup.execute('key3', fn)
      expect(dedup.getStats().misses).to.equal(3)
    })

    it('should track pending requests correctly', async () => {
      const fn = async () => {
        await new Promise(resolve => setTimeout(resolve, 100))
        return 'result'
      }

      expect(dedup.getStats().pending).to.equal(0)

      const p1 = dedup.execute('key1', fn)
      expect(dedup.getStats().pending).to.equal(1)

      const p2 = dedup.execute('key2', fn)
      expect(dedup.getStats().pending).to.equal(2)

      const p3 = dedup.execute('key3', fn)
      expect(dedup.getStats().pending).to.equal(3)

      await Promise.all([p1, p2, p3])
      expect(dedup.getStats().pending).to.equal(0)
    })

    it('should return a copy of stats (not reference)', () => {
      const stats1 = dedup.getStats()
      stats1.hits = 999

      const stats2 = dedup.getStats()
      expect(stats2.hits).to.equal(0, 'Should not be affected by mutation')
    })
  })

  describe('clear()', () => {
    it('should reset statistics', async () => {
      const fn = async () => 'result'

      await Promise.all([
        dedup.execute('key1', fn),
        dedup.execute('key1', fn),
      ])

      expect(dedup.getStats().hits).to.be.greaterThan(0)
      expect(dedup.getStats().misses).to.be.greaterThan(0)

      dedup.clear()

      const stats = dedup.getStats()
      expect(stats.hits).to.equal(0)
      expect(stats.misses).to.equal(0)
      expect(stats.pending).to.equal(0)
    })

    it('should clear pending requests map', async () => {
      const fn = async () => {
        await new Promise(resolve => setTimeout(resolve, 100))
        return 'result'
      }

      dedup.execute('key1', fn)
      dedup.execute('key2', fn)
      expect(dedup.getStats().pending).to.equal(2)

      dedup.clear()
      expect(dedup.getStats().pending).to.equal(0)
    })
  })

  describe('cleanup()', () => {
    it('should not crash when called', () => {
      expect(() => dedup.cleanup()).to.not.throw()
    })

    it('should accept maxAge parameter', () => {
      expect(() => dedup.cleanup(60000)).to.not.throw()
    })
  })

  describe('Edge Cases', () => {
    it('should handle rapid sequential requests', async () => {
      let callCount = 0
      const fn = async () => {
        callCount++
        return `result-${callCount}`
      }

      // Execute requests rapidly in sequence
      const results: string[] = []
      for (let i = 0; i < 10; i++) {
        results.push(await dedup.execute(`key-${i}`, fn))
      }

      expect(callCount).to.equal(10)
      expect(results).to.deep.equal([
        'result-1', 'result-2', 'result-3', 'result-4', 'result-5',
        'result-6', 'result-7', 'result-8', 'result-9', 'result-10',
      ])
    })

    it('should handle mixed concurrent and sequential requests', async () => {
      let callCount = 0
      const fn = async () => {
        callCount++
        await new Promise(resolve => setTimeout(resolve, 50))
        return `result-${callCount}`
      }

      // First batch (concurrent)
      const [r1, r2] = await Promise.all([
        dedup.execute('key1', fn),
        dedup.execute('key1', fn),
      ])
      expect(r1).to.equal('result-1')
      expect(r2).to.equal('result-1')
      expect(callCount).to.equal(1)

      // Second batch (concurrent, different key)
      const [r3, r4] = await Promise.all([
        dedup.execute('key2', fn),
        dedup.execute('key2', fn),
      ])
      expect(r3).to.equal('result-2')
      expect(r4).to.equal('result-2')
      expect(callCount).to.equal(2)

      // Sequential (same key as first batch)
      const r5 = await dedup.execute('key1', fn)
      expect(r5).to.equal('result-3')
      expect(callCount).to.equal(3)
    })

    it('should handle empty key strings', async () => {
      const fn = async () => 'result'

      const [r1, r2] = await Promise.all([
        dedup.execute('', fn),
        dedup.execute('', fn),
      ])

      expect(r1).to.equal('result')
      expect(r2).to.equal('result')
    })

    it('should handle very long key strings', async () => {
      const longKey = 'a'.repeat(10000)
      const fn = async () => 'result'

      const [r1, r2] = await Promise.all([
        dedup.execute(longKey, fn),
        dedup.execute(longKey, fn),
      ])

      expect(r1).to.equal('result')
      expect(r2).to.equal('result')
    })
  })

  describe('Integration with cachedFetch', () => {
    beforeEach(() => {
      // Clear global deduplication manager before each test
      deduplicationManager.clear()
    })

    afterEach(() => {
      deduplicationManager.clear()
    })

    it('should work with global deduplicationManager instance', async () => {
      let callCount = 0
      const fn = async () => {
        callCount++
        await new Promise(resolve => setTimeout(resolve, 50))
        return 'result'
      }

      // Simulate concurrent calls through global manager
      const [r1, r2, r3] = await Promise.all([
        deduplicationManager.execute('test:key1', fn),
        deduplicationManager.execute('test:key1', fn),
        deduplicationManager.execute('test:key1', fn),
      ])

      expect(callCount).to.equal(1)
      expect(r1).to.equal('result')
      expect(r2).to.equal('result')
      expect(r3).to.equal('result')

      const stats = deduplicationManager.getStats()
      expect(stats.hits).to.equal(2)
      expect(stats.misses).to.equal(1)
    })

    it('should handle cache key serialization', async () => {
      let callCount = 0
      const fn = async () => {
        callCount++
        await new Promise(resolve => setTimeout(resolve, 50))
        return 'result'
      }

      // Simulate how cachedFetch generates dedup keys
      const cacheType = 'page'
      const cacheKey = { id: 'page-123' }
      const dedupKey = `${cacheType}:${JSON.stringify(cacheKey)}`

      const [r1, r2] = await Promise.all([
        deduplicationManager.execute(dedupKey, fn),
        deduplicationManager.execute(dedupKey, fn),
      ])

      expect(callCount).to.equal(1)
      expect(r1).to.equal(r2)
    })

    it('should deduplicate based on serialized cache keys', async () => {
      let callCount = 0
      const fn = async () => {
        callCount++
        await new Promise(resolve => setTimeout(resolve, 50))
        return 'result'
      }

      // Different object instances with same values should deduplicate
      const key1 = `page:${JSON.stringify({ id: 'page-123' })}`
      const key2 = `page:${JSON.stringify({ id: 'page-123' })}`

      const [r1, r2] = await Promise.all([
        deduplicationManager.execute(key1, fn),
        deduplicationManager.execute(key2, fn),
      ])

      expect(callCount).to.equal(1)
      expect(r1).to.equal(r2)
    })
  })
})
