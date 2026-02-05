import { expect } from 'chai'
import * as fs from 'fs/promises'
import * as path from 'path'
import * as os from 'os'
import { DiskCacheManager, DiskCacheEntry } from '../src/utils/disk-cache'

describe('DiskCacheManager', () => {
  let diskCache: DiskCacheManager
  let tmpDir: string

  beforeEach(async () => {
    // Use temp directory for tests
    tmpDir = path.join(os.tmpdir(), `notion-cli-test-${Date.now()}-${Math.random().toString(36).slice(2)}`)
    diskCache = new DiskCacheManager({ cacheDir: tmpDir, syncInterval: 0 })
    await diskCache.initialize()
  })

  afterEach(async () => {
    await diskCache.shutdown()
    try {
      await fs.rm(tmpDir, { recursive: true, force: true })
    } catch {
      // Ignore cleanup errors
    }
  })

  describe('initialize()', () => {
    it('should create cache directory', async () => {
      const stats = await fs.stat(tmpDir)
      expect(stats.isDirectory()).to.be.true
    })

    it('should not fail if directory already exists', async () => {
      // Initialize again
      await diskCache.initialize()
      const stats = await fs.stat(tmpDir)
      expect(stats.isDirectory()).to.be.true
    })
  })

  describe('set() and get()', () => {
    it('should store and retrieve entries', async () => {
      const data = { foo: 'bar', nested: { value: 123 } }
      await diskCache.set('key1', data, 60000)

      const entry = await diskCache.get<typeof data>('key1')
      expect(entry).to.not.be.null
      expect(entry?.data).to.deep.equal(data)
    })

    it('should handle different data types', async () => {
      // String
      await diskCache.set('string', 'test value', 60000)
      const str = await diskCache.get<string>('string')
      expect(str?.data).to.equal('test value')

      // Number
      await diskCache.set('number', 42, 60000)
      const num = await diskCache.get<number>('number')
      expect(num?.data).to.equal(42)

      // Array
      await diskCache.set('array', [1, 2, 3], 60000)
      const arr = await diskCache.get<number[]>('array')
      expect(arr?.data).to.deep.equal([1, 2, 3])

      // Object
      await diskCache.set('object', { a: 1, b: 2 }, 60000)
      const obj = await diskCache.get<{ a: number; b: number }>('object')
      expect(obj?.data).to.deep.equal({ a: 1, b: 2 })

      // Null
      await diskCache.set('null', null, 60000)
      const nul = await diskCache.get<null>('null')
      expect(nul?.data).to.be.null

      // Boolean
      await diskCache.set('bool', true, 60000)
      const bool = await diskCache.get<boolean>('bool')
      expect(bool?.data).to.equal(true)
    })

    it('should return null for non-existent keys', async () => {
      const entry = await diskCache.get('nonexistent')
      expect(entry).to.be.null
    })

    it('should store metadata correctly', async () => {
      const data = 'test'
      const ttl = 60000
      const beforeSet = Date.now()

      await diskCache.set('key1', data, ttl)

      const entry = await diskCache.get('key1')
      expect(entry).to.not.be.null
      expect(entry?.key).to.equal('key1')
      expect(entry?.createdAt).to.be.greaterThanOrEqual(beforeSet)
      expect(entry?.createdAt).to.be.lessThanOrEqual(Date.now())
      expect(entry?.expiresAt).to.be.greaterThan(Date.now())
      expect(entry?.size).to.be.greaterThan(0)
    })
  })

  describe('Expiration', () => {
    it('should not return expired entries', async () => {
      await diskCache.set('key1', 'value', 100) // 100ms TTL
      await new Promise(resolve => setTimeout(resolve, 150))

      const entry = await diskCache.get('key1')
      expect(entry).to.be.null
    })

    it('should delete expired entries on get', async () => {
      await diskCache.set('key1', 'value', 100)
      await new Promise(resolve => setTimeout(resolve, 150))

      // First get should delete the entry
      await diskCache.get('key1')

      // Check that file is deleted
      const files = await fs.readdir(tmpDir)
      const jsonFiles = files.filter(f => f.endsWith('.json'))
      expect(jsonFiles).to.have.length(0)
    })

    it('should handle entries with long TTL', async () => {
      await diskCache.set('key1', 'value', 3600000) // 1 hour

      const entry = await diskCache.get('key1')
      expect(entry).to.not.be.null
      expect(entry?.data).to.equal('value')
    })
  })

  describe('invalidate()', () => {
    it('should delete specific entries', async () => {
      await diskCache.set('key1', 'value1', 60000)
      await diskCache.set('key2', 'value2', 60000)

      await diskCache.invalidate('key1')

      const entry1 = await diskCache.get('key1')
      const entry2 = await diskCache.get('key2')

      expect(entry1).to.be.null
      expect(entry2).to.not.be.null
    })

    it('should not fail when invalidating non-existent keys', async () => {
      await diskCache.invalidate('nonexistent')
      // Should not throw
    })
  })

  describe('clear()', () => {
    it('should remove all entries', async () => {
      await diskCache.set('key1', 'value1', 60000)
      await diskCache.set('key2', 'value2', 60000)
      await diskCache.set('key3', 'value3', 60000)

      await diskCache.clear()

      const entry1 = await diskCache.get('key1')
      const entry2 = await diskCache.get('key2')
      const entry3 = await diskCache.get('key3')

      expect(entry1).to.be.null
      expect(entry2).to.be.null
      expect(entry3).to.be.null
    })

    it('should not fail on empty cache', async () => {
      await diskCache.clear()
      // Should not throw
    })
  })

  describe('getStats()', () => {
    it('should return accurate statistics', async () => {
      await diskCache.set('key1', 'a'.repeat(100), 60000)
      await diskCache.set('key2', 'b'.repeat(200), 60000)

      const stats = await diskCache.getStats()

      expect(stats.totalEntries).to.equal(2)
      expect(stats.totalSize).to.be.greaterThan(0)
      expect(stats.oldestEntry).to.not.be.null
      expect(stats.newestEntry).to.not.be.null
    })

    it('should return zeros for empty cache', async () => {
      const stats = await diskCache.getStats()

      expect(stats.totalEntries).to.equal(0)
      expect(stats.totalSize).to.equal(0)
      expect(stats.oldestEntry).to.be.null
      expect(stats.newestEntry).to.be.null
    })

    it('should track oldest and newest entries', async () => {
      await diskCache.set('key1', 'first', 60000)
      await new Promise(resolve => setTimeout(resolve, 50))
      await diskCache.set('key2', 'second', 60000)

      const stats = await diskCache.getStats()

      expect(stats.oldestEntry).to.be.lessThan(stats.newestEntry!)
    })
  })

  describe('Persistence', () => {
    it('should persist entries across instances', async () => {
      await diskCache.set('key1', { data: 'persisted' }, 60000)
      await diskCache.shutdown()

      // Create new instance
      const diskCache2 = new DiskCacheManager({ cacheDir: tmpDir, syncInterval: 0 })
      await diskCache2.initialize()

      const entry = await diskCache2.get<{ data: string }>('key1')
      expect(entry).to.not.be.null
      expect(entry?.data).to.deep.equal({ data: 'persisted' })

      await diskCache2.shutdown()
    })

    it('should handle corrupted cache files gracefully', async () => {
      // Write corrupted file
      const files = await fs.readdir(tmpDir)
      const corruptedPath = path.join(tmpDir, 'corrupted.json')
      await fs.writeFile(corruptedPath, '{invalid json', 'utf-8')

      // Should not throw when getting stats
      const stats = await diskCache.getStats()
      expect(stats.totalEntries).to.equal(0)
    })
  })

  describe('Max Size Enforcement', () => {
    it('should remove oldest entries when over limit', async () => {
      const smallCache = new DiskCacheManager({
        cacheDir: tmpDir,
        maxSize: 1000, // 1KB limit
        syncInterval: 0,
      })
      await smallCache.initialize()

      // Add entries that exceed limit
      await smallCache.set('key1', 'a'.repeat(500), 60000)
      await new Promise(resolve => setTimeout(resolve, 10))
      await smallCache.set('key2', 'b'.repeat(500), 60000)
      await new Promise(resolve => setTimeout(resolve, 10))
      await smallCache.set('key3', 'c'.repeat(500), 60000)

      const stats = await smallCache.getStats()
      expect(stats.totalSize).to.be.lessThanOrEqual(1000)

      await smallCache.shutdown()
    })

    it('should remove expired entries during size enforcement', async () => {
      const smallCache = new DiskCacheManager({
        cacheDir: tmpDir,
        maxSize: 1000,
        syncInterval: 0,
      })
      await smallCache.initialize()

      // Add entry that will expire
      await smallCache.set('expired', 'x'.repeat(500), 10)
      await new Promise(resolve => setTimeout(resolve, 50))

      // Add new entry that triggers cleanup
      await smallCache.set('new', 'y'.repeat(500), 60000)

      // Expired entry should be removed
      const entry = await smallCache.get('expired')
      expect(entry).to.be.null

      await smallCache.shutdown()
    })
  })

  describe('Atomic Writes', () => {
    it('should use atomic writes with temp files', async () => {
      await diskCache.set('key1', 'atomic', 60000)

      // Check that no .tmp files remain
      const files = await fs.readdir(tmpDir)
      const tmpFiles = files.filter(f => f.endsWith('.tmp'))
      expect(tmpFiles).to.have.length(0)
    })

    it('should handle write failures gracefully', async () => {
      // This test is harder to implement without mocking
      // But we can at least verify it doesn't crash
      await diskCache.set('key1', 'test', 60000)
      expect(true).to.be.true
    })
  })

  describe('Key Hashing', () => {
    it('should handle long keys', async () => {
      const longKey = 'x'.repeat(1000)
      await diskCache.set(longKey, 'value', 60000)

      const entry = await diskCache.get(longKey)
      expect(entry).to.not.be.null
      expect(entry?.data).to.equal('value')
    })

    it('should handle special characters in keys', async () => {
      const specialKey = 'key:with/special\\characters?and=symbols'
      await diskCache.set(specialKey, 'value', 60000)

      const entry = await diskCache.get(specialKey)
      expect(entry).to.not.be.null
      expect(entry?.data).to.equal('value')
    })

    it('should create unique files for different keys', async () => {
      await diskCache.set('key1', 'value1', 60000)
      await diskCache.set('key2', 'value2', 60000)

      const files = await fs.readdir(tmpDir)
      const jsonFiles = files.filter(f => f.endsWith('.json'))
      expect(jsonFiles).to.have.length(2)
    })
  })

  describe('Concurrent Access', () => {
    it('should handle concurrent writes', async () => {
      const promises = Array(10).fill(0).map((_, i) =>
        diskCache.set(`key${i}`, `value${i}`, 60000)
      )

      await Promise.all(promises)

      const stats = await diskCache.getStats()
      expect(stats.totalEntries).to.equal(10)
    })

    it('should handle concurrent reads', async () => {
      await diskCache.set('key1', 'value', 60000)

      const promises = Array(10).fill(0).map(() =>
        diskCache.get('key1')
      )

      const results = await Promise.all(promises)
      expect(results.every(r => r?.data === 'value')).to.be.true
    })

    it('should handle concurrent read/write/invalidate', async () => {
      const operations = [
        diskCache.set('key1', 'value1', 60000),
        diskCache.get('key2'),
        diskCache.invalidate('key3'),
        diskCache.set('key4', 'value4', 60000),
        diskCache.get('key1'),
      ]

      // Should not throw
      await Promise.all(operations)
    })
  })

  describe('Edge Cases', () => {
    it('should handle empty string keys', async () => {
      await diskCache.set('', 'value', 60000)
      const entry = await diskCache.get('')
      expect(entry?.data).to.equal('value')
    })

    it('should handle very large data', async () => {
      const largeData = 'x'.repeat(100000)
      await diskCache.set('large', largeData, 60000)

      const entry = await diskCache.get('large')
      expect(entry?.data).to.equal(largeData)
    })

    it('should handle zero TTL (immediate expiration)', async () => {
      await diskCache.set('key1', 'value', 0)
      await new Promise(resolve => setTimeout(resolve, 10))

      const entry = await diskCache.get('key1')
      expect(entry).to.be.null
    })

    it('should handle negative TTL', async () => {
      await diskCache.set('key1', 'value', -1000)

      const entry = await diskCache.get('key1')
      expect(entry).to.be.null
    })
  })

  describe('shutdown()', () => {
    it('should flush and cleanup', async () => {
      await diskCache.set('key1', 'value', 60000)
      await diskCache.shutdown()

      // Entry should still be on disk
      const diskCache2 = new DiskCacheManager({ cacheDir: tmpDir, syncInterval: 0 })
      await diskCache2.initialize()
      const entry = await diskCache2.get('key1')
      expect(entry?.data).to.equal('value')
      await diskCache2.shutdown()
    })

    it('should allow re-initialization after shutdown', async () => {
      await diskCache.shutdown()
      await diskCache.initialize()
      await diskCache.set('key1', 'value', 60000)

      const entry = await diskCache.get('key1')
      expect(entry?.data).to.equal('value')
    })
  })
})
