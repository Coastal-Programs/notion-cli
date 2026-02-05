import { expect } from 'chai'
import { BATCH_CONFIG } from '../src/notion'
import { batchWithRetry } from '../src/retry'

describe('Parallel Operations', () => {
  describe('BATCH_CONFIG', () => {
    it('should have default delete concurrency', () => {
      expect(BATCH_CONFIG.deleteConcurrency).to.be.a('number')
      expect(BATCH_CONFIG.deleteConcurrency).to.be.greaterThan(0)
    })

    it('should have default children concurrency', () => {
      expect(BATCH_CONFIG.childrenConcurrency).to.be.a('number')
      expect(BATCH_CONFIG.childrenConcurrency).to.be.greaterThan(0)
    })

    it('should respect environment variable for delete concurrency', () => {
      const expected = parseInt(process.env.NOTION_CLI_DELETE_CONCURRENCY || '5', 10)
      expect(BATCH_CONFIG.deleteConcurrency).to.equal(expected)
    })

    it('should respect environment variable for children concurrency', () => {
      const expected = parseInt(process.env.NOTION_CLI_CHILDREN_CONCURRENCY || '10', 10)
      expect(BATCH_CONFIG.childrenConcurrency).to.equal(expected)
    })
  })

  describe('batchWithRetry()', () => {
    it('should execute operations in parallel', async () => {
      const executionOrder: number[] = []
      const operations = [1, 2, 3, 4, 5].map(num => async () => {
        executionOrder.push(num)
        await new Promise(resolve => setTimeout(resolve, 50))
        return `result-${num}`
      })

      const startTime = Date.now()
      const results = await batchWithRetry(operations, { concurrency: 5 })
      const duration = Date.now() - startTime

      // Should complete in ~50ms (parallel) not ~250ms (sequential)
      expect(duration).to.be.lessThan(200)
      expect(results).to.have.length(5)
      expect(results.every(r => r.success)).to.be.true
    })

    it('should respect concurrency limit', async () => {
      let concurrent = 0
      let maxConcurrent = 0

      const operations = Array(10).fill(0).map(() => async () => {
        concurrent++
        maxConcurrent = Math.max(maxConcurrent, concurrent)
        await new Promise(resolve => setTimeout(resolve, 50))
        concurrent--
        return 'done'
      })

      await batchWithRetry(operations, { concurrency: 3 })

      expect(maxConcurrent).to.be.at.most(3)
    })

    it('should handle mixed success and failure', async () => {
      const operations = [
        async () => 'success-1',
        async () => { throw new Error('failure-1') },
        async () => 'success-2',
        async () => { throw new Error('failure-2') },
        async () => 'success-3',
      ]

      const results = await batchWithRetry(operations, { concurrency: 5 })

      expect(results).to.have.length(5)
      expect(results[0].success).to.be.true
      expect(results[0].data).to.equal('success-1')
      expect(results[1].success).to.be.false
      expect(results[1].error).to.be.instanceOf(Error)
      expect(results[2].success).to.be.true
      expect(results[3].success).to.be.false
      expect(results[4].success).to.be.true
    })

    it('should continue processing after failures', async () => {
      let successCount = 0
      const operations = [
        async () => { successCount++; return 'ok' },
        async () => { throw new Error('fail') },
        async () => { successCount++; return 'ok' },
        async () => { throw new Error('fail') },
        async () => { successCount++; return 'ok' },
      ]

      await batchWithRetry(operations, { concurrency: 5 })

      expect(successCount).to.equal(3)
    })

    it('should handle empty operations array', async () => {
      const results = await batchWithRetry([], { concurrency: 5 })
      expect(results).to.be.an('array')
      expect(results).to.have.length(0)
    })

    it('should handle single operation', async () => {
      const operations = [async () => 'single-result']
      const results = await batchWithRetry(operations, { concurrency: 5 })

      expect(results).to.have.length(1)
      expect(results[0].success).to.be.true
      expect(results[0].data).to.equal('single-result')
    })

    it('should process operations in batches when count exceeds concurrency', async () => {
      const batchOrder: number[] = []
      const operations = Array(15).fill(0).map((_, index) => async () => {
        batchOrder.push(index)
        await new Promise(resolve => setTimeout(resolve, 10))
        return index
      })

      const results = await batchWithRetry(operations, { concurrency: 5 })

      expect(results).to.have.length(15)
      expect(results.every(r => r.success)).to.be.true

      // First 5 should start before next 5
      const firstBatch = batchOrder.slice(0, 5)
      const secondBatch = batchOrder.slice(5, 10)
      expect(firstBatch.every(i => i < 5)).to.be.true
      expect(secondBatch.every(i => i >= 5 && i < 10)).to.be.true
    })

    it('should return results in order', async () => {
      const operations = [1, 2, 3, 4, 5].map(num => async () => {
        // Add random delay to simulate out-of-order completion
        await new Promise(resolve => setTimeout(resolve, Math.random() * 100))
        return num
      })

      const results = await batchWithRetry(operations, { concurrency: 5 })

      expect(results).to.have.length(5)
      expect(results[0].data).to.equal(1)
      expect(results[1].data).to.equal(2)
      expect(results[2].data).to.equal(3)
      expect(results[3].data).to.equal(4)
      expect(results[4].data).to.equal(5)
    })

    it('should handle operations that return different types', async () => {
      const operations = [
        async () => 'string',
        async () => 42,
        async () => ({ key: 'value' }),
        async () => [1, 2, 3],
        async () => true,
        async () => null,
      ]

      const results = await batchWithRetry(operations, { concurrency: 6 })

      expect(results).to.have.length(6)
      expect(results[0].data).to.equal('string')
      expect(results[1].data).to.equal(42)
      expect(results[2].data).to.deep.equal({ key: 'value' })
      expect(results[3].data).to.deep.equal([1, 2, 3])
      expect(results[4].data).to.equal(true)
      expect(results[5].data).to.be.null
    })
  })

  describe('Performance Characteristics', () => {
    it('should be significantly faster than sequential execution', async () => {
      const delay = 100
      const count = 5

      // Sequential timing
      const sequentialStart = Date.now()
      for (let i = 0; i < count; i++) {
        await new Promise(resolve => setTimeout(resolve, delay))
      }
      const sequentialDuration = Date.now() - sequentialStart

      // Parallel timing
      const operations = Array(count).fill(0).map(() => async () => {
        await new Promise(resolve => setTimeout(resolve, delay))
        return 'done'
      })

      const parallelStart = Date.now()
      await batchWithRetry(operations, { concurrency: count })
      const parallelDuration = Date.now() - parallelStart

      // Parallel should be at least 3x faster
      expect(parallelDuration).to.be.lessThan(sequentialDuration / 3)
    })

    it('should handle large batch sizes efficiently', async () => {
      const operations = Array(100).fill(0).map((_, index) => async () => {
        await new Promise(resolve => setTimeout(resolve, 10))
        return index
      })

      const startTime = Date.now()
      const results = await batchWithRetry(operations, { concurrency: 10 })
      const duration = Date.now() - startTime

      expect(results).to.have.length(100)
      expect(results.every(r => r.success)).to.be.true

      // Should complete in reasonable time (~100ms with concurrency 10)
      // 100 operations / 10 concurrent = 10 batches * 10ms = ~100ms
      expect(duration).to.be.lessThan(300)
    })
  })

  describe('Error Handling', () => {
    it('should capture error details in failed results', async () => {
      const errorMessage = 'Custom error message'
      const operations = [
        async () => { throw new Error(errorMessage) },
      ]

      const results = await batchWithRetry(operations, { concurrency: 1 })

      expect(results[0].success).to.be.false
      expect(results[0].error).to.be.instanceOf(Error)
      expect(results[0].error.message).to.equal(errorMessage)
    })

    it('should handle errors without stopping other operations', async () => {
      let completedCount = 0
      const operations = Array(10).fill(0).map((_, index) => async () => {
        if (index % 2 === 0) {
          throw new Error(`Error ${index}`)
        }
        completedCount++
        return `success-${index}`
      })

      const results = await batchWithRetry(operations, { concurrency: 5 })

      expect(completedCount).to.equal(5) // Half should succeed
      expect(results.filter(r => r.success)).to.have.length(5)
      expect(results.filter(r => !r.success)).to.have.length(5)
    })
  })

  describe('Edge Cases', () => {
    it('should handle concurrency of 1 (sequential execution)', async () => {
      const executionOrder: number[] = []
      const operations = [1, 2, 3].map(num => async () => {
        executionOrder.push(num)
        await new Promise(resolve => setTimeout(resolve, 10))
        return num
      })

      const results = await batchWithRetry(operations, { concurrency: 1 })

      expect(results).to.have.length(3)
      expect(executionOrder).to.deep.equal([1, 2, 3])
    })

    it('should handle concurrency greater than operation count', async () => {
      const operations = [1, 2, 3].map(num => async () => num)
      const results = await batchWithRetry(operations, { concurrency: 10 })

      expect(results).to.have.length(3)
      expect(results.every(r => r.success)).to.be.true
    })

    it('should handle operations that resolve immediately', async () => {
      const operations = Array(10).fill(0).map((_, i) => async () => i)
      const results = await batchWithRetry(operations, { concurrency: 5 })

      expect(results).to.have.length(10)
      expect(results.every(r => r.success)).to.be.true
    })

    it('should handle operations that take varying time', async () => {
      const delays = [100, 10, 50, 5, 75]
      const operations = delays.map(delay => async () => {
        await new Promise(resolve => setTimeout(resolve, delay))
        return delay
      })

      const results = await batchWithRetry(operations, { concurrency: 5 })

      expect(results).to.have.length(5)
      expect(results.every(r => r.success)).to.be.true
      // Results should maintain order despite different completion times
      expect(results.map(r => r.data)).to.deep.equal([100, 10, 50, 5, 75])
    })
  })
})
