/**
 * Global test setup
 * Loaded by mocha before running tests
 *
 * NOTE: fetch polyfill is in test/helpers/init.js which loads FIRST
 */

import * as nock from 'nock'

// Set test environment
process.env.NODE_ENV = 'test'

// Enable nock to intercept http/https requests
nock.disableNetConnect()
nock.enableNetConnect('127.0.0.1')

// Verify fetch polyfill is loaded
if (typeof global.fetch !== 'function') {
  throw new Error('FATAL: fetch polyfill not loaded! init.js must run before setup.ts')
}

// Debug: Log when nock can't match requests (only in debug mode)
if (process.env.TEST_DEBUG) {
  nock.emitter.on('no match', (req) => {
    if (req && req.method && req.href) {
      console.error('[NOCK] NO MATCH:', req.method, req.href)
    }
  })
  console.log('[NOCK] Enabled - network mocking active')
}

// Suppress console output during tests (optional)
// process.env.LOG_LEVEL = 'silent'
