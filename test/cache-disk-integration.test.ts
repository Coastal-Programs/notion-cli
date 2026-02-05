/**
 * Integration tests for CacheManager with DiskCacheManager
 * Tests the disk cache integration added in v5.9.0
 */

import { expect } from 'chai'
import * as fs from 'fs/promises'
import * as path from 'path'
import * as os from 'os'
import { CacheManager } from '../dist/cache.js'
import { diskCacheManager } from '../dist/utils/disk-cache.js'

describe('CacheManager Integration with DiskCacheManager', () => {
  let cache: CacheManager
  const originalDiskCacheEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED
  const originalDebug = process.env.DEBUG

  beforeEach(async () => {
    // Enable disk cache for these tests
    process.env.NOTION_CLI_DISK_CACHE_ENABLED = 'true'
    process.env.DEBUG = 'false'

    // Ensure global disk cache is initialized
    await diskCacheManager.initialize()

    // Create CacheManager instance
    cache = new CacheManager({
      enabled: true,
      defaultTtl: 60000,
      maxSize: 10,
      ttlByType: {
        dataSource: 60000,
        database: 60000,
        user: 60000,
        page: 60000,
        block: 60000,
      },
    })

    // Clear any existing cache
    await diskCacheManager.clear()
  })

  afterEach(async () => {
    // Clear cache
    await diskCacheManager.clear()

    // Restore original env vars
    if (originalDiskCacheEnabled !== undefined) {
      process.env.NOTION_CLI_DISK_CACHE_ENABLED = originalDiskCacheEnabled
    } else {
      delete process.env.NOTION_CLI_DISK_CACHE_ENABLED
    }

    if (originalDebug !== undefined) {
      process.env.DEBUG = originalDebug
    } else {
      delete process.env.DEBUG
    }
  })

  describe('Memory-to-Disk Write on set()', () => {
    it('should write to disk cache when setting values', async () => {
      const data = { id: '123', name: 'test' }
      cache.set('dataSource', data, undefined, '123')

      // Wait for async disk write to complete
      await new Promise(resolve => setTimeout(resolve, 150))

      // Verify it's in disk cache
      const diskEntry = await diskCacheManager.get('dataSource:123')
      expect(diskEntry).to.not.be.null
      expect(diskEntry?.data).to.have.property('data')
      expect((diskEntry?.data as any).data).to.deep.equal(data)
    })

    it('should not write to disk when NOTION_CLI_DISK_CACHE_ENABLED=false', async () => {
      process.env.NOTION_CLI_DISK_CACHE_ENABLED = 'false'

      const data = { id: '456', name: 'no-disk' }
      cache.set('dataSource', data, undefined, '456')

      // Wait to ensure no async write happens
      await new Promise(resolve => setTimeout(resolve, 150))

      // Verify it's NOT in disk cache
      const diskEntry = await diskCacheManager.get('dataSource:456')
      expect(diskEntry).to.be.null
    })

    it('should handle disk write failures gracefully', async () => {
      // Create a mock that will fail
      const originalSet = diskCacheManager.set.bind(diskCacheManager)
      diskCacheManager.set = async () => {
        throw new Error('Disk write failed')
      }

      // Should not throw - errors are silently ignored
      const data = { id: '789', name: 'fail-test' }
      expect(() => cache.set('dataSource', data, undefined, '789')).to.not.throw()

      // Restore original
      diskCacheManager.set = originalSet
    })
  })

  describe('Disk-to-Memory Promotion on get()', () => {
    it('should promote valid disk cache entries to memory', async () => {
      // Write directly to disk cache
      const cacheEntry = {
        data: { id: 'abc', name: 'from-disk' },
        timestamp: Date.now(),
        ttl: 60000,
      }
      await diskCacheManager.set('dataSource:abc', cacheEntry, 60000)

      // Clear memory cache to ensure we're testing disk promotion
      cache.clear()

      // First get should now retrieve from disk (with await - bug fix applied)
      const result = await cache.get('dataSource', 'abc')
      expect(result).to.not.be.null
      expect(result).to.deep.equal(cacheEntry.data)

      // Second get should hit memory after promotion
      const result2 = await cache.get('dataSource', 'abc')
      expect(result2).to.not.be.null
      expect(result2).to.deep.equal(cacheEntry.data)
    })

    it('should not promote expired disk entries', async () => {
      // Write expired entry to disk
      const expiredEntry = {
        data: { id: 'expired', name: 'old' },
        timestamp: Date.now() - 100000, // Very old
        ttl: 1000, // Short TTL
      }
      await diskCacheManager.set('dataSource:expired', expiredEntry, 1000)

      // Clear memory cache
      cache.clear()

      // Should not promote expired entry (validates TTL before promoting)
      const result = await cache.get('dataSource', 'expired')
      expect(result).to.be.null
    })

    it('should delete expired disk entries when validation fails', async () => {
      // Write expired entry to disk
      const expiredEntry = {
        data: { id: 'cleanup', name: 'old' },
        timestamp: Date.now() - 100000,
        ttl: 1000,
      }
      await diskCacheManager.set('dataSource:cleanup', expiredEntry, 1000)

      // Trigger promotion attempt (will validate and delete expired entry)
      await cache.get('dataSource', 'cleanup')

      // Give invalidation a moment to complete (fire-and-forget)
      await new Promise(resolve => setTimeout(resolve, 100))

      // Verify disk entry was deleted
      const diskEntry = await diskCacheManager.get('dataSource:cleanup')
      expect(diskEntry).to.be.null
    })

    it('should handle disk read failures gracefully', async () => {
      // Mock disk cache to fail
      const originalGet = diskCacheManager.get.bind(diskCacheManager)
      try {
        diskCacheManager.get = async () => {
          throw new Error('Disk read failed')
        }

        // Should not throw - errors are silently ignored
        const result = await cache.get('dataSource', 'fail-read')
        expect(result).to.be.null
      } finally {
        // Restore original
        diskCacheManager.get = originalGet
      }
    })

    it('should not check disk when NOTION_CLI_DISK_CACHE_ENABLED=false', async () => {
      process.env.NOTION_CLI_DISK_CACHE_ENABLED = 'false'

      // Write to disk
      const cacheEntry = {
        data: { id: 'no-check', name: 'test' },
        timestamp: Date.now(),
        ttl: 60000,
      }
      await diskCacheManager.set('dataSource:no-check', cacheEntry, 60000)

      // Clear memory
      cache.clear()
      await new Promise(resolve => setTimeout(resolve, 50))

      // Should not check disk (returns null immediately)
      const result = await cache.get('dataSource', 'no-check')
      expect(result).to.be.null

      // Wait and verify it's still not in memory
      await new Promise(resolve => setTimeout(resolve, 150))
      const result2 = await cache.get('dataSource', 'no-check')
      expect(result2).to.be.null
    })

    it('should log disk cache hit in DEBUG mode', async () => {
      const originalDebugValue = process.env.DEBUG
      process.env.DEBUG = 'true'

      // Capture console.error calls
      const originalError = console.error
      const errorLogs: string[] = []
      console.error = (msg: string) => {
        errorLogs.push(msg)
      }

      try {
        // Write to disk
        const cacheEntry = {
          data: { id: 'debug-test', name: 'test' },
          timestamp: Date.now(),
          ttl: 60000,
        }
        await diskCacheManager.set('dataSource:debug-test', cacheEntry, 60000)

        // Clear memory to force disk lookup
        cache.clear()

        // Wait longer for disk operations to settle
        await new Promise(resolve => setTimeout(resolve, 150))

        // Trigger disk promotion
        await cache.get('dataSource', 'debug-test')

        // Wait for async disk promotion logging to complete
        await new Promise(resolve => setTimeout(resolve, 300))

        // Verify debug log (disk cache hit happens asynchronously)
        const diskHitLog = errorLogs.find(log => {
          try {
            const parsed = JSON.parse(log)
            return parsed.event === 'disk_cache_hit' && parsed.namespace === 'dataSource'
          } catch {
            return false
          }
        })
        expect(diskHitLog).to.not.be.undefined
      } finally {
        // Restore console.error
        console.error = originalError
        // Restore DEBUG env var
        if (originalDebugValue !== undefined) {
          process.env.DEBUG = originalDebugValue
        } else {
          delete process.env.DEBUG
        }
      }
    })
  })

  describe('Disk Invalidation', () => {
    it('should invalidate specific entries from disk', async () => {
      // Set entries
      cache.set('dataSource', { id: '1' }, undefined, '1')
      cache.set('dataSource', { id: '2' }, undefined, '2')
      await new Promise(resolve => setTimeout(resolve, 150))

      // Invalidate one entry
      cache.invalidate('dataSource', '1')
      await new Promise(resolve => setTimeout(resolve, 150))

      // Verify disk state
      const entry1 = await diskCacheManager.get('dataSource:1')
      const entry2 = await diskCacheManager.get('dataSource:2')
      expect(entry1).to.be.null
      expect(entry2).to.not.be.null
    })

    it('should invalidate all entries of a type from disk', async () => {
      // Set multiple entries
      cache.set('dataSource', { id: '1' }, undefined, '1')
      cache.set('dataSource', { id: '2' }, undefined, '2')
      cache.set('user', { id: '3' }, undefined, '3')
      await new Promise(resolve => setTimeout(resolve, 150))

      // Invalidate all dataSource entries
      cache.invalidate('dataSource')
      await new Promise(resolve => setTimeout(resolve, 150))

      // Verify disk state
      const ds1 = await diskCacheManager.get('dataSource:1')
      const ds2 = await diskCacheManager.get('dataSource:2')
      const user3 = await diskCacheManager.get('user:3')
      expect(ds1).to.be.null
      expect(ds2).to.be.null
      expect(user3).to.not.be.null
    })

    it('should not invalidate disk when NOTION_CLI_DISK_CACHE_ENABLED=false', async () => {
      // Set entry with disk enabled
      cache.set('dataSource', { id: 'persist' }, undefined, 'persist')
      await new Promise(resolve => setTimeout(resolve, 150))

      // Disable disk cache
      process.env.NOTION_CLI_DISK_CACHE_ENABLED = 'false'

      // Invalidate
      cache.invalidate('dataSource', 'persist')
      await new Promise(resolve => setTimeout(resolve, 150))

      // Verify disk entry still exists
      const entry = await diskCacheManager.get('dataSource:persist')
      expect(entry).to.not.be.null
    })
  })

  describe('Disk Clear', () => {
    it('should clear all disk cache entries', async () => {
      // Set multiple entries
      cache.set('dataSource', { id: '1' }, undefined, '1')
      cache.set('user', { id: '2' }, undefined, '2')
      cache.set('page', { id: '3' }, undefined, '3')
      await new Promise(resolve => setTimeout(resolve, 150))

      // Clear all
      cache.clear()
      await new Promise(resolve => setTimeout(resolve, 150))

      // Verify disk is empty
      const stats = await diskCacheManager.getStats()
      expect(stats.totalEntries).to.equal(0)
    })

    it('should not clear disk when NOTION_CLI_DISK_CACHE_ENABLED=false', async () => {
      // Set entries with disk enabled
      cache.set('dataSource', { id: 'keep' }, undefined, 'keep')
      await new Promise(resolve => setTimeout(resolve, 150))

      // Disable disk cache
      process.env.NOTION_CLI_DISK_CACHE_ENABLED = 'false'

      // Clear
      cache.clear()
      await new Promise(resolve => setTimeout(resolve, 150))

      // Verify disk entry still exists
      const stats = await diskCacheManager.getStats()
      expect(stats.totalEntries).to.be.greaterThan(0)
    })

    it('should handle disk clear failures gracefully', async () => {
      // Mock disk cache to fail
      const originalClear = diskCacheManager.clear.bind(diskCacheManager)
      diskCacheManager.clear = async () => {
        throw new Error('Disk clear failed')
      }

      // Should not throw - errors are silently ignored
      expect(() => cache.clear()).to.not.throw()

      // Restore original
      diskCacheManager.clear = originalClear
    })
  })

  describe('Verbose Logging with NOTION_CLI_VERBOSE', () => {
    it('should log cache events when NOTION_CLI_VERBOSE=true', async () => {
      process.env.NOTION_CLI_VERBOSE = 'true'

      // Capture console.error
      const originalError = console.error
      const errorLogs: string[] = []
      console.error = (msg: string) => {
        errorLogs.push(msg)
      }

      // Perform cache operations
      cache.set('dataSource', { id: '1' }, undefined, 'verbose1')
      await cache.get('dataSource', 'verbose1')
      await cache.get('dataSource', 'nonexistent')
      cache.invalidate('dataSource', 'verbose1')

      // Verify logs were generated
      expect(errorLogs.length).to.be.greaterThan(0)

      // Verify log structure
      const parsedLogs = errorLogs.map(log => {
        try {
          return JSON.parse(log)
        } catch {
          return null
        }
      }).filter(Boolean)

      expect(parsedLogs.some(log => log.event === 'cache_set')).to.be.true
      expect(parsedLogs.some(log => log.event === 'cache_hit')).to.be.true
      expect(parsedLogs.some(log => log.event === 'cache_miss')).to.be.true
      expect(parsedLogs.some(log => log.event === 'cache_invalidate')).to.be.true

      // Restore
      console.error = originalError
      delete process.env.NOTION_CLI_VERBOSE
    })

    it('should log cache events when NOTION_CLI_DEBUG=true', async () => {
      process.env.NOTION_CLI_DEBUG = 'true'

      const originalError = console.error
      const errorLogs: string[] = []
      console.error = (msg: string) => {
        errorLogs.push(msg)
      }

      cache.set('dataSource', { id: '2' }, undefined, 'debug2')
      await cache.get('dataSource', 'debug2')

      expect(errorLogs.length).to.be.greaterThan(0)

      console.error = originalError
      delete process.env.NOTION_CLI_DEBUG
    })

    it('should log eviction events when NOTION_CLI_VERBOSE=true', async () => {
      process.env.NOTION_CLI_VERBOSE = 'true'

      const originalError = console.error
      const errorLogs: string[] = []
      console.error = (msg: string) => {
        errorLogs.push(msg)
      }

      // Create expired entry
      cache.set('dataSource', { id: 'exp' }, 10, 'expire')
      await new Promise(resolve => setTimeout(resolve, 50))

      // Trigger eviction by accessing expired entry
      await cache.get('dataSource', 'expire')

      // Check for eviction log
      const parsedLogs = errorLogs.map(log => {
        try {
          return JSON.parse(log)
        } catch {
          return null
        }
      }).filter(Boolean)

      expect(parsedLogs.some(log => log.event === 'cache_evict')).to.be.true

      console.error = originalError
      delete process.env.NOTION_CLI_VERBOSE
    })

    it('should log LRU eviction when cache is full', async () => {
      process.env.NOTION_CLI_VERBOSE = 'true'

      const originalError = console.error
      const errorLogs: string[] = []
      console.error = (msg: string) => {
        errorLogs.push(msg)
      }

      // Fill cache to capacity (maxSize is 10)
      for (let i = 0; i < 10; i++) {
        cache.set('dataSource', { id: i }, undefined, String(i))
      }

      // Add one more to trigger LRU eviction
      cache.set('dataSource', { id: 11 }, undefined, '11')

      // Check for LRU eviction log
      const parsedLogs = errorLogs.map(log => {
        try {
          return JSON.parse(log)
        } catch {
          return null
        }
      }).filter(Boolean)

      const lruEviction = parsedLogs.find(log => log.event === 'cache_evict' && log.namespace === 'lru')
      expect(lruEviction).to.not.be.undefined

      console.error = originalError
      delete process.env.NOTION_CLI_VERBOSE
    })

    it('should log when clearing cache with entries', async () => {
      process.env.NOTION_CLI_VERBOSE = 'true'

      const originalError = console.error
      const errorLogs: string[] = []
      console.error = (msg: string) => {
        errorLogs.push(msg)
      }

      // Add some entries
      cache.set('dataSource', { id: '1' }, undefined, '1')
      cache.set('dataSource', { id: '2' }, undefined, '2')

      // Clear
      cache.clear()

      const parsedLogs = errorLogs.map(log => {
        try {
          return JSON.parse(log)
        } catch {
          return null
        }
      }).filter(Boolean)

      const clearLog = parsedLogs.find(log =>
        log.event === 'cache_invalidate' &&
        log.namespace === 'all' &&
        log.level === 'info'
      )
      expect(clearLog).to.not.be.undefined

      console.error = originalError
      delete process.env.NOTION_CLI_VERBOSE
    })
  })

  describe('Edge Cases and Additional Coverage', () => {
    it('should handle object identifiers in key generation', async () => {
      const objId = { type: 'database', id: '123' }
      cache.set('query', { results: [] }, undefined, objId)

      const result = await cache.get('query', objId)
      expect(result).to.not.be.null
      expect(result).to.deep.equal({ results: [] })
    })

    it('should handle numeric identifiers', async () => {
      cache.set('dataSource', { id: 'numeric' }, undefined, 123)

      const result = await cache.get('dataSource', 123)
      expect(result).to.not.be.null
    })

    it('should invalidate all entries of a type even when some already evicted', async () => {
      cache.set('dataSource', { id: '1' }, 10, '1') // Will expire quickly
      cache.set('dataSource', { id: '2' }, 60000, '2')

      // Invalidate all - should work even with mixed valid/invalid entries
      cache.invalidate('dataSource')

      expect(await cache.get('dataSource', '1')).to.be.null
      expect(await cache.get('dataSource', '2')).to.be.null
    })

    it('should handle custom TTL from ttlByType config', async () => {
      const customCache = new CacheManager({
        enabled: true,
        defaultTtl: 5000,
        maxSize: 10,
        ttlByType: {
          dataSource: 100, // Very short TTL
          database: 60000,
          user: 60000,
          page: 60000,
          block: 60000,
        },
      })

      customCache.set('dataSource', { id: 'test' }, undefined, 'ds1')

      // Check that it exists initially
      let result = await customCache.get('dataSource', 'ds1')
      expect(result).to.not.be.null

      // Wait for expiration
      return new Promise<void>((resolve) => {
        setTimeout(async () => {
          const expired = await customCache.get('dataSource', 'ds1')
          expect(expired).to.be.null
          resolve()
        }, 150)
      })
    })

    it('should properly handle getStats', async () => {
      cache.clear()

      cache.set('dataSource', { id: '1' }, undefined, '1')
      cache.set('dataSource', { id: '2' }, undefined, '2')

      await cache.get('dataSource', '1') // Hit
      await cache.get('dataSource', 'nonexistent') // Miss

      const stats = cache.getStats()
      expect(stats.size).to.equal(2)
      expect(stats.sets).to.be.greaterThan(0)
      expect(stats.hits).to.be.greaterThan(0)
      expect(stats.misses).to.be.greaterThan(0)
    })

    it('should calculate hit rate', async () => {
      cache.clear()
      await new Promise(resolve => setTimeout(resolve, 50))

      cache.set('dataSource', { id: '1' }, undefined, '1')

      await cache.get('dataSource', '1') // Hit
      await cache.get('dataSource', '1') // Hit
      await cache.get('dataSource', '2') // Miss

      const hitRate = cache.getHitRate()
      expect(hitRate).to.be.closeTo(0.667, 0.01) // 2 hits / 3 total
    })

    it('should return 0 hit rate with no accesses', () => {
      const emptyCache = new CacheManager()
      expect(emptyCache.getHitRate()).to.equal(0)
    })

    it('should check if cache is enabled', () => {
      expect(cache.isEnabled()).to.be.true

      const disabledCache = new CacheManager({ enabled: false })
      expect(disabledCache.isEnabled()).to.be.false
    })

    it('should return cache config', () => {
      const config = cache.getConfig()
      expect(config).to.have.property('enabled')
      expect(config).to.have.property('defaultTtl')
      expect(config).to.have.property('maxSize')
      expect(config).to.have.property('ttlByType')
    })

    it('should handle invalid entry removal during get', async () => {
      // Set with very short TTL
      cache.set('dataSource', { id: 'shortlived' }, 10, 'short')

      // Wait for expiration
      await new Promise(resolve => setTimeout(resolve, 50))

      // Get should remove the invalid entry
      const result = await cache.get('dataSource', 'short')
      expect(result).to.be.null

      // Stats should show an eviction
      const stats = cache.getStats()
      expect(stats.evictions).to.be.greaterThan(0)
    })
  })
})
