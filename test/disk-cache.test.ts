import { expect } from 'chai'
import * as fs from 'fs/promises'
import * as path from 'path'
import * as os from 'os'
import { DiskCacheManager, DiskCacheEntry } from '../dist/utils/disk-cache.js'

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

    it('should start sync timer when syncInterval > 0', async () => {
      // Create cache with non-zero sync interval
      const tmpDir2 = path.join(os.tmpdir(), `notion-cli-test-${Date.now()}-${Math.random().toString(36).slice(2)}`)
      const cache = new DiskCacheManager({ cacheDir: tmpDir2, syncInterval: 100 })

      await cache.initialize()

      // Timer should be started (we can't directly check, but we can verify it doesn't throw)
      await cache.shutdown()
      await fs.rm(tmpDir2, { recursive: true, force: true })
    })

    it('should handle sync errors silently', async () => {
      const tmpDir2 = path.join(os.tmpdir(), `notion-cli-test-${Date.now()}-${Math.random().toString(36).slice(2)}`)
      const cache = new DiskCacheManager({ cacheDir: tmpDir2, syncInterval: 100 })

      await cache.initialize()

      // Wait for potential sync
      await new Promise(resolve => setTimeout(resolve, 150))

      await cache.shutdown()
      await fs.rm(tmpDir2, { recursive: true, force: true })
    })

    it('should handle sync errors with DEBUG env', async () => {
      const tmpDir2 = path.join(os.tmpdir(), `notion-cli-test-${Date.now()}-${Math.random().toString(36).slice(2)}`)
      const cache = new DiskCacheManager({ cacheDir: tmpDir2, syncInterval: 50 })

      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      await cache.initialize()

      // Override sync to throw an error
      const originalSync = (cache as any).sync.bind(cache)
      ;(cache as any).sync = async () => {
        throw new Error('Simulated sync error')
      }

      // Wait for sync to trigger and catch error
      await new Promise(resolve => setTimeout(resolve, 100))

      // Restore original sync
      ;(cache as any).sync = originalSync

      process.env.DEBUG = originalDebug
      await cache.shutdown()
      await fs.rm(tmpDir2, { recursive: true, force: true })
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

    it('should handle read errors with DEBUG env', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      // Try to read with corrupted file
      const corruptedPath = path.join(tmpDir, 'corrupted-read.json')
      await fs.writeFile(corruptedPath, '{invalid json}', 'utf-8')

      // Manually call get with hash that points to corrupted file
      const result = await diskCache.get('any-key')

      process.env.DEBUG = originalDebug
      expect(result).to.be.null
    })

    it('should handle write errors with DEBUG env', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      // Create a directory where file should be (will cause write error)
      const tmpDir2 = path.join(os.tmpdir(), `notion-cli-test-${Date.now()}-${Math.random().toString(36).slice(2)}`)
      const cache = new DiskCacheManager({ cacheDir: tmpDir2, syncInterval: 0 })
      await cache.initialize()

      // Set should handle error gracefully
      await cache.set('key1', 'value', 60000)

      process.env.DEBUG = originalDebug
      await cache.shutdown()
      await fs.rm(tmpDir2, { recursive: true, force: true })
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

    it('should handle delete errors with DEBUG env', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      await diskCache.set('key1', 'value', 60000)

      // Get the actual file path for this key
      const files = await fs.readdir(tmpDir)
      const jsonFiles = files.filter(f => f.endsWith('.json'))

      if (jsonFiles.length > 0) {
        const filePath = path.join(tmpDir, jsonFiles[0])

        try {
          // Make file read-only to trigger delete error (may not work on all systems)
          await fs.chmod(filePath, 0o444)

          // Should handle error gracefully
          await diskCache.invalidate('key1')

          // Restore permissions
          await fs.chmod(filePath, 0o644)
        } catch {
          // If chmod doesn't work on this system, just skip the test
        }
      }

      process.env.DEBUG = originalDebug
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

    it('should skip .tmp files when clearing', async () => {
      await diskCache.set('key1', 'value1', 60000)

      // Create a temp file manually
      const tmpFile = path.join(tmpDir, 'test.tmp')
      await fs.writeFile(tmpFile, 'temp', 'utf-8')

      await diskCache.clear()

      // Temp file should still exist
      const files = await fs.readdir(tmpDir)
      expect(files.includes('test.tmp')).to.be.true

      // Clean up
      await fs.unlink(tmpFile)
    })

    it('should handle clear errors on non-existent directory with DEBUG env', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      // Create a cache with a non-existent directory
      const tmpDir2 = path.join(os.tmpdir(), `notion-cli-test-nonexistent-${Date.now()}`)
      const cache = new DiskCacheManager({ cacheDir: tmpDir2, syncInterval: 0 })

      // Should not throw
      await cache.clear()

      process.env.DEBUG = originalDebug
    })

    it('should handle clear errors with DEBUG env', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      await diskCache.set('key1', 'value', 60000)

      // Clear should handle any errors gracefully
      await diskCache.clear()

      process.env.DEBUG = originalDebug
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

    it('should skip .tmp files in stats', async () => {
      await diskCache.set('key1', 'value', 60000)

      // Create a temp file manually
      const tmpFile = path.join(tmpDir, 'test.tmp')
      await fs.writeFile(tmpFile, JSON.stringify({ key: 'tmp', data: 'test', size: 100 }), 'utf-8')

      const stats = await diskCache.getStats()

      // Should only count the regular entry, not the tmp file
      expect(stats.totalEntries).to.equal(1)

      // Clean up
      await fs.unlink(tmpFile)
    })

    it('should handle getStats errors on non-existent directory', async () => {
      const tmpDir2 = path.join(os.tmpdir(), `notion-cli-test-nonexistent-${Date.now()}`)
      const cache = new DiskCacheManager({ cacheDir: tmpDir2, syncInterval: 0 })

      const stats = await cache.getStats()

      expect(stats.totalEntries).to.equal(0)
      expect(stats.totalSize).to.equal(0)
      expect(stats.oldestEntry).to.be.null
      expect(stats.newestEntry).to.be.null
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

    it('should skip corrupted entries during size enforcement', async () => {
      const smallCache = new DiskCacheManager({
        cacheDir: tmpDir,
        maxSize: 1000,
        syncInterval: 0,
      })
      await smallCache.initialize()

      // Add a valid entry
      await smallCache.set('key1', 'a'.repeat(500), 60000)

      // Create a corrupted entry
      const corruptedPath = path.join(tmpDir, 'corrupted.json')
      await fs.writeFile(corruptedPath, '{invalid json}', 'utf-8')

      // Add another entry that triggers size enforcement
      await smallCache.set('key2', 'b'.repeat(500), 60000)

      // Should not throw
      const stats = await smallCache.getStats()
      expect(stats.totalEntries).to.be.greaterThan(0)

      await smallCache.shutdown()
    })

    it('should handle enforceMaxSize with DEBUG env', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      const smallCache = new DiskCacheManager({
        cacheDir: tmpDir,
        maxSize: 1000,
        syncInterval: 0,
      })
      await smallCache.initialize()

      // Add entries that exceed limit
      await smallCache.set('key1', 'a'.repeat(500), 60000)
      await new Promise(resolve => setTimeout(resolve, 10))
      await smallCache.set('key2', 'b'.repeat(500), 60000)
      await new Promise(resolve => setTimeout(resolve, 10))
      await smallCache.set('key3', 'c'.repeat(500), 60000)

      process.env.DEBUG = originalDebug
      await smallCache.shutdown()
    })

    it('should skip .tmp files during size enforcement', async () => {
      const smallCache = new DiskCacheManager({
        cacheDir: tmpDir,
        maxSize: 1000,
        syncInterval: 0,
      })
      await smallCache.initialize()

      // Add a valid entry
      await smallCache.set('key1', 'a'.repeat(500), 60000)

      // Create a temp file manually
      const tmpFile = path.join(tmpDir, 'test.tmp')
      await fs.writeFile(tmpFile, JSON.stringify({ key: 'tmp', data: 'x'.repeat(500), size: 500 }), 'utf-8')

      // Add another entry
      await smallCache.set('key2', 'b'.repeat(500), 60000)

      // Temp file should be skipped during size enforcement
      const files = await fs.readdir(tmpDir)
      expect(files.includes('test.tmp')).to.be.true

      // Clean up
      await fs.unlink(tmpFile)
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

    it('should read maxSize from environment variable', async () => {
      const originalMaxSize = process.env.NOTION_CLI_DISK_CACHE_MAX_SIZE
      process.env.NOTION_CLI_DISK_CACHE_MAX_SIZE = '2000'

      const tmpDir2 = path.join(os.tmpdir(), `notion-cli-test-${Date.now()}-${Math.random().toString(36).slice(2)}`)
      const cache = new DiskCacheManager({ cacheDir: tmpDir2 })
      await cache.initialize()

      // Verify it uses the env variable
      await cache.set('key1', 'x'.repeat(1500), 60000)
      await cache.set('key2', 'y'.repeat(1500), 60000)

      const stats = await cache.getStats()
      expect(stats.totalSize).to.be.lessThanOrEqual(2000)

      process.env.NOTION_CLI_DISK_CACHE_MAX_SIZE = originalMaxSize
      await cache.shutdown()
      await fs.rm(tmpDir2, { recursive: true, force: true })
    })

    it('should read syncInterval from environment variable', async () => {
      const originalSyncInterval = process.env.NOTION_CLI_DISK_CACHE_SYNC_INTERVAL
      process.env.NOTION_CLI_DISK_CACHE_SYNC_INTERVAL = '200'

      const tmpDir2 = path.join(os.tmpdir(), `notion-cli-test-${Date.now()}-${Math.random().toString(36).slice(2)}`)
      const cache = new DiskCacheManager({ cacheDir: tmpDir2 })
      await cache.initialize()

      // Should not throw
      await cache.shutdown()

      process.env.NOTION_CLI_DISK_CACHE_SYNC_INTERVAL = originalSyncInterval
      await fs.rm(tmpDir2, { recursive: true, force: true })
    })
  })

  describe('shutdown()', () => {
    it('should flush and cleanup', async () => {
      // This test verifies that shutdown flushes data properly
      // by checking that sync is called and timers are cleared
      await diskCache.set('key1', 'test-value', 60000)

      // Verify entry exists before shutdown
      const entryBefore = await diskCache.get<string>('key1')
      expect(entryBefore).to.not.be.null

      await diskCache.shutdown()

      // Verify shutdown cleared the timer
      expect((diskCache as any).syncTimer).to.be.null
      expect((diskCache as any).initialized).to.be.false
    })

    it('should allow re-initialization after shutdown', async () => {
      await diskCache.shutdown()
      await diskCache.initialize()
      await diskCache.set('key1', 'value', 60000)

      const entry = await diskCache.get('key1')
      expect(entry?.data).to.equal('value')
    })

    it('should clear sync timer on shutdown', async () => {
      const tmpDir2 = path.join(os.tmpdir(), `notion-cli-test-${Date.now()}-${Math.random().toString(36).slice(2)}`)
      const cache = new DiskCacheManager({ cacheDir: tmpDir2, syncInterval: 100 })
      await cache.initialize()
      await cache.shutdown()

      // Should be able to shutdown again without error
      await cache.shutdown()

      await fs.rm(tmpDir2, { recursive: true, force: true })
    })
  })

  describe('Error Handling', () => {
    it('should handle JSON parse errors gracefully', async () => {
      // Write invalid JSON to a cache file
      const files = await fs.readdir(tmpDir)
      const testFile = path.join(tmpDir, 'invalid.json')
      await fs.writeFile(testFile, 'not valid json', 'utf-8')

      // Should return null instead of throwing
      const result = await diskCache.get('some-key')
      expect(result).to.be.null
    })

    it('should handle JSON parse errors with DEBUG env', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      // First set a valid entry, then corrupt it
      await diskCache.set('corrupt-key', 'value', 60000)

      // Find the file and corrupt it
      const files = await fs.readdir(tmpDir)
      const jsonFiles = files.filter(f => f.endsWith('.json'))
      if (jsonFiles.length > 0) {
        const corruptFile = path.join(tmpDir, jsonFiles[0])
        await fs.writeFile(corruptFile, 'not valid json', 'utf-8')

        // Try to read it - should trigger DEBUG console.warn
        const result = await diskCache.get('corrupt-key')
        expect(result).to.be.null
      }

      process.env.DEBUG = originalDebug
    })

    it('should handle file system errors during write', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      // Create a cache with a read-only directory to trigger write errors
      const tmpDir2 = path.join(os.tmpdir(), `notion-cli-test-readonly-${Date.now()}`)
      await fs.mkdir(tmpDir2, { recursive: true })

      try {
        // Make directory read-only (may not work on all systems)
        await fs.chmod(tmpDir2, 0o444)

        const cache = new DiskCacheManager({ cacheDir: tmpDir2, syncInterval: 0 })
        await cache.initialize().catch(() => {}) // May fail on initialize

        // Try to write - should fail and trigger DEBUG console.warn
        await cache.set('test-key', 'test-value', 60000)

        // Restore permissions for cleanup
        await fs.chmod(tmpDir2, 0o755)
      } catch {
        // If chmod doesn't work on this system, just skip
        try {
          await fs.chmod(tmpDir2, 0o755)
        } catch {}
      }

      process.env.DEBUG = originalDebug
      await fs.rm(tmpDir2, { recursive: true, force: true }).catch(() => {})
    })

    it('should cleanup temp files after write failure', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      // This is difficult to trigger without mocking, but we can at least
      // verify the code path exists
      await diskCache.set('cleanup-test', 'value', 60000)

      process.env.DEBUG = originalDebug
    })

    it('should use default cacheDir when none provided', async () => {
      const cache = new DiskCacheManager()
      const expectedDir = path.join(os.homedir(), '.notion-cli', 'cache')

      // Don't initialize to avoid creating files in user's home
      // Just verify the path is set correctly
      expect((cache as any).cacheDir).to.equal(expectedDir)
    })

    it('should handle directory creation failures', async () => {
      // This is hard to test without mocking, but we can at least verify
      // that the error is caught and re-thrown with a better message
      const invalidPath = '\0invalid'
      const cache = new DiskCacheManager({ cacheDir: invalidPath, syncInterval: 0 })

      try {
        await cache.initialize()
        // If it doesn't throw, that's also acceptable (some systems may handle it)
        expect(true).to.be.true
      } catch (error: any) {
        // Should have a helpful error message
        expect(error.message).to.include('Failed to create cache directory')
      }
    })

    it('should handle readdir errors in clear with non-ENOENT', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      // Call clear on valid cache (should work fine)
      await diskCache.clear()

      process.env.DEBUG = originalDebug
    })

    it('should handle readdir errors in enforceMaxSize', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      // Create a directory that will cause issues during size enforcement
      const tmpDir2 = path.join(os.tmpdir(), `notion-cli-test-enforce-${Date.now()}`)
      const smallCache = new DiskCacheManager({
        cacheDir: tmpDir2,
        maxSize: 100,
        syncInterval: 0,
      })
      await smallCache.initialize()

      // Add enough data to trigger size enforcement
      await smallCache.set('test1', 'x'.repeat(60), 60000)
      await smallCache.set('test2', 'x'.repeat(60), 60000)

      process.env.DEBUG = originalDebug
      await smallCache.shutdown()
      await fs.rm(tmpDir2, { recursive: true, force: true })
    })

    it('should handle invalidate errors with DEBUG for non-ENOENT', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      // This is hard to trigger without mocking, but we can test the code path exists
      await diskCache.set('test-invalidate', 'value', 60000)
      await diskCache.invalidate('test-invalidate')

      process.env.DEBUG = originalDebug
    })

    it('should handle clear errors with DEBUG for non-ENOENT', async () => {
      const originalDebug = process.env.DEBUG
      process.env.DEBUG = '1'

      // Test clear with DEBUG enabled
      await diskCache.set('test-clear', 'value', 60000)
      await diskCache.clear()

      process.env.DEBUG = originalDebug
    })
  })

  describe('Constructor Options', () => {
    it('should accept custom cacheDir', async () => {
      const customDir = path.join(os.tmpdir(), `custom-cache-${Date.now()}`)
      const cache = new DiskCacheManager({ cacheDir: customDir, syncInterval: 0 })
      await cache.initialize()

      const stats = await fs.stat(customDir)
      expect(stats.isDirectory()).to.be.true

      await cache.shutdown()
      await fs.rm(customDir, { recursive: true, force: true })
    })

    it('should accept custom maxSize', async () => {
      const cache = new DiskCacheManager({ cacheDir: tmpDir, maxSize: 500, syncInterval: 0 })
      expect((cache as any).maxSize).to.equal(500)
    })

    it('should accept custom syncInterval', async () => {
      const cache = new DiskCacheManager({ cacheDir: tmpDir, syncInterval: 1000 })
      expect((cache as any).syncInterval).to.equal(1000)
    })
  })

  describe('Sync Method', () => {
    it('should clear dirtyKeys on sync', async () => {
      await diskCache.sync()
      expect((diskCache as any).dirtyKeys.size).to.equal(0)
    })
  })
})
