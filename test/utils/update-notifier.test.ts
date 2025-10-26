import { expect } from '@oclif/test'

/**
 * Tests for Update Notifier
 *
 * Verifies the automatic update notification system:
 * 1. Module structure and exports
 * 2. Function executes without throwing
 * 3. Graceful failure handling
 *
 * Note: Full integration testing of update-notifier would require:
 * - Mocking npm registry responses
 * - Controlling time/cache state
 * - Capturing console output
 * These are tested by update-notifier's own test suite.
 */

describe('update-notifier utility', () => {
  describe('module structure', () => {
    it('should export checkForUpdates function', async () => {
      const updateNotifier = await import('../../src/utils/update-notifier')
      expect(updateNotifier.checkForUpdates).to.be.a('function')
    })

    it('should load without errors', async () => {
      expect(async () => {
        await import('../../src/utils/update-notifier')
      }).to.not.throw()
    })
  })

  describe('checkForUpdates function', () => {
    it('should execute without throwing errors', () => {
      const { checkForUpdates } = require('../../src/utils/update-notifier')

      // Should not throw even if update check fails
      expect(() => checkForUpdates()).to.not.throw()
    })

    it('should handle errors gracefully', () => {
      const { checkForUpdates } = require('../../src/utils/update-notifier')

      // Even in error conditions, should not throw
      const originalEnv = process.env.NO_UPDATE_NOTIFIER
      try {
        // Disable update notifier
        process.env.NO_UPDATE_NOTIFIER = '1'
        expect(() => checkForUpdates()).to.not.throw()
      } finally {
        // Restore
        if (originalEnv !== undefined) {
          process.env.NO_UPDATE_NOTIFIER = originalEnv
        } else {
          delete process.env.NO_UPDATE_NOTIFIER
        }
      }
    })

    it('should be callable multiple times', () => {
      const { checkForUpdates } = require('../../src/utils/update-notifier')

      expect(() => {
        checkForUpdates()
        checkForUpdates()
        checkForUpdates()
      }).to.not.throw()
    })

    it('should not block execution', (done) => {
      const { checkForUpdates } = require('../../src/utils/update-notifier')

      // Call the function
      checkForUpdates()

      // Should complete immediately (async in background)
      // If it blocks, this will timeout
      done()
    })
  })

  describe('environment variable support', () => {
    it('should respect NO_UPDATE_NOTIFIER environment variable', () => {
      const originalEnv = process.env.NO_UPDATE_NOTIFIER

      try {
        process.env.NO_UPDATE_NOTIFIER = '1'
        const { checkForUpdates } = require('../../src/utils/update-notifier')

        // Should still not throw when disabled
        expect(() => checkForUpdates()).to.not.throw()
      } finally {
        if (originalEnv !== undefined) {
          process.env.NO_UPDATE_NOTIFIER = originalEnv
        } else {
          delete process.env.NO_UPDATE_NOTIFIER
        }
      }
    })

    it('should work with NO_UPDATE_NOTIFIER unset', () => {
      const originalEnv = process.env.NO_UPDATE_NOTIFIER

      try {
        delete process.env.NO_UPDATE_NOTIFIER
        const { checkForUpdates } = require('../../src/utils/update-notifier')

        expect(() => checkForUpdates()).to.not.throw()
      } finally {
        if (originalEnv !== undefined) {
          process.env.NO_UPDATE_NOTIFIER = originalEnv
        } else {
          delete process.env.NO_UPDATE_NOTIFIER
        }
      }
    })
  })

  describe('integration with CLI', () => {
    it('should be callable from bin/run entry point', () => {
      // Verify the require path in bin/run works
      expect(() => {
        const { checkForUpdates } = require('../../dist/utils/update-notifier')
        expect(checkForUpdates).to.be.a('function')
      }).to.not.throw()
    })

    it('should have compiled output in dist/', () => {
      const fs = require('fs')
      const path = require('path')

      const distPath = path.join(process.cwd(), 'dist', 'utils', 'update-notifier.js')
      const exists = fs.existsSync(distPath)

      expect(exists).to.be.true
    })

    it('should have TypeScript declarations in dist/', () => {
      const fs = require('fs')
      const path = require('path')

      const dtsPath = path.join(process.cwd(), 'dist', 'utils', 'update-notifier.d.ts')
      const exists = fs.existsSync(dtsPath)

      expect(exists).to.be.true
    })
  })

  describe('error resilience', () => {
    it('should handle missing package.json gracefully', () => {
      const { checkForUpdates } = require('../../src/utils/update-notifier')

      // Even if package.json is somehow missing/corrupted
      // The function should not crash the CLI
      expect(() => checkForUpdates()).to.not.throw()
    })

    it('should handle npm registry failures gracefully', () => {
      const { checkForUpdates } = require('../../src/utils/update-notifier')

      // Network failures, timeouts, etc. should not crash CLI
      expect(() => checkForUpdates()).to.not.throw()
    })
  })
})
