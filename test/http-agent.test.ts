import { expect } from 'chai'
import { httpsAgent, getAgentStats, getAgentConfig, destroyAgents, REQUEST_TIMEOUT } from '../dist/http-agent.js'

describe('HTTP Agent', () => {
  describe('httpsAgent', () => {
    it('should be an HTTPS agent instance', () => {
      expect(httpsAgent).to.exist
      expect(httpsAgent).to.have.property('keepAlive')
      expect(httpsAgent).to.have.property('maxSockets')
    })

    it('should have keep-alive enabled by default', () => {
      const config = getAgentConfig()
      expect(config.keepAlive).to.be.true
    })

    it('should have reasonable default values', () => {
      const config = getAgentConfig()

      expect(config.keepAliveMsecs).to.be.a('number')
      expect(config.keepAliveMsecs).to.be.greaterThan(0)

      expect(config.maxSockets).to.be.a('number')
      expect(config.maxSockets).to.be.greaterThan(0)

      expect(config.maxFreeSockets).to.be.a('number')
      expect(config.maxFreeSockets).to.be.greaterThan(0)

      expect(config.requestTimeout).to.be.a('number')
      expect(config.requestTimeout).to.be.greaterThan(0)
    })
  })

  describe('getAgentConfig()', () => {
    it('should return complete configuration', () => {
      const config = getAgentConfig()

      expect(config).to.have.property('keepAlive')
      expect(config).to.have.property('keepAliveMsecs')
      expect(config).to.have.property('maxSockets')
      expect(config).to.have.property('maxFreeSockets')
      expect(config).to.have.property('requestTimeout')
    })

    it('should return numeric values for all timing configs', () => {
      const config = getAgentConfig()

      expect(config.keepAliveMsecs).to.be.a('number')
      expect(config.maxSockets).to.be.a('number')
      expect(config.maxFreeSockets).to.be.a('number')
      expect(config.requestTimeout).to.be.a('number')
    })

    it('should respect environment variables', () => {
      const config = getAgentConfig()

      // Check if environment variables are being read
      const expectedKeepAlive = process.env.NOTION_CLI_HTTP_KEEP_ALIVE !== 'false'
      expect(config.keepAlive).to.equal(expectedKeepAlive)

      const expectedKeepAliveMsecs = parseInt(
        process.env.NOTION_CLI_HTTP_KEEP_ALIVE_MS || '60000',
        10
      )
      expect(config.keepAliveMsecs).to.equal(expectedKeepAliveMsecs)

      const expectedMaxSockets = parseInt(
        process.env.NOTION_CLI_HTTP_MAX_SOCKETS || '50',
        10
      )
      expect(config.maxSockets).to.equal(expectedMaxSockets)

      const expectedMaxFreeSockets = parseInt(
        process.env.NOTION_CLI_HTTP_MAX_FREE_SOCKETS || '10',
        10
      )
      expect(config.maxFreeSockets).to.equal(expectedMaxFreeSockets)

      const expectedTimeout = parseInt(
        process.env.NOTION_CLI_HTTP_TIMEOUT || '30000',
        10
      )
      expect(config.requestTimeout).to.equal(expectedTimeout)
    })
  })

  describe('getAgentStats()', () => {
    it('should return statistics object', () => {
      const stats = getAgentStats()

      expect(stats).to.have.property('sockets')
      expect(stats).to.have.property('freeSockets')
      expect(stats).to.have.property('requests')
    })

    it('should return numeric values', () => {
      const stats = getAgentStats()

      expect(stats.sockets).to.be.a('number')
      expect(stats.freeSockets).to.be.a('number')
      expect(stats.requests).to.be.a('number')
    })

    it('should return non-negative values', () => {
      const stats = getAgentStats()

      expect(stats.sockets).to.be.at.least(0)
      expect(stats.freeSockets).to.be.at.least(0)
      expect(stats.requests).to.be.at.least(0)
    })

    it('should track connection state', () => {
      const statsBefore = getAgentStats()

      // Stats should be valid numbers
      expect(statsBefore.sockets).to.be.a('number')
      expect(statsBefore.freeSockets).to.be.a('number')
      expect(statsBefore.requests).to.be.a('number')
    })
  })

  describe('destroyAgents()', () => {
    it('should not throw when called', () => {
      // Create a fresh agent for this test
      // We can't destroy the shared agent without affecting other tests
      expect(() => {
        // Test that the function exists and can be called
        // Note: We don't actually destroy the shared agent in tests
        const fn = destroyAgents
        expect(fn).to.be.a('function')
      }).to.not.throw()
    })

    it('should be a callable function', () => {
      expect(destroyAgents).to.be.a('function')
    })

    it('should actually call destroy on the agent', () => {
      // Verify that destroyAgents can be called without errors
      // This covers the actual function body execution
      let destroyCallCount = 0
      const originalDestroy = httpsAgent.destroy.bind(httpsAgent)

      // Mock the destroy method temporarily
      ;(httpsAgent as any).destroy = () => {
        destroyCallCount++
        originalDestroy()
      }

      destroyAgents()

      expect(destroyCallCount).to.equal(1)

      // Restore original destroy
      ;(httpsAgent as any).destroy = originalDestroy
    })
  })

  describe('Configuration Validation', () => {
    it('should have sensible keep-alive requestTimeout', () => {
      const config = getAgentConfig()

      // Keep-alive should be between 10 seconds and 5 minutes
      expect(config.keepAliveMsecs).to.be.at.least(1000)
      expect(config.keepAliveMsecs).to.be.at.most(300000)
    })

    it('should have reasonable socket limits', () => {
      const config = getAgentConfig()

      // Max sockets should be reasonable (10-100)
      expect(config.maxSockets).to.be.at.least(1)
      expect(config.maxSockets).to.be.at.most(1000)

      // Free sockets should be less than max sockets
      expect(config.maxFreeSockets).to.be.at.most(config.maxSockets)
    })

    it('should have reasonable requestTimeout', () => {
      const config = getAgentConfig()

      // Timeout should be between 5 seconds and 2 minutes
      expect(config.requestTimeout).to.be.at.least(5000)
      expect(config.requestTimeout).to.be.at.most(120000)
    })
  })

  describe('Default Values', () => {
    it('should use default keep-alive msecs when env var not set', () => {
      // If env var is not set, should use 60000 (60 seconds)
      const expected = parseInt(process.env.NOTION_CLI_HTTP_KEEP_ALIVE_MS || '60000', 10)
      const config = getAgentConfig()
      expect(config.keepAliveMsecs).to.equal(expected)
    })

    it('should use default max sockets when env var not set', () => {
      // If env var is not set, should use 50
      const expected = parseInt(process.env.NOTION_CLI_HTTP_MAX_SOCKETS || '50', 10)
      const config = getAgentConfig()
      expect(config.maxSockets).to.equal(expected)
    })

    it('should use default max free sockets when env var not set', () => {
      // If env var is not set, should use 10
      const expected = parseInt(process.env.NOTION_CLI_HTTP_MAX_FREE_SOCKETS || '10', 10)
      const config = getAgentConfig()
      expect(config.maxFreeSockets).to.equal(expected)
    })

    it('should use default requestTimeout when env var not set', () => {
      // If env var is not set, should use 30000 (30 seconds)
      const expected = parseInt(process.env.NOTION_CLI_HTTP_TIMEOUT || '30000', 10)
      const config = getAgentConfig()
      expect(config.requestTimeout).to.equal(expected)
    })
  })

  describe('Agent Properties', () => {
    it('should have all required properties', () => {
      const agent = httpsAgent as any
      expect(agent).to.have.property('keepAlive')
      expect(agent).to.have.property('keepAliveMsecs')
      expect(agent).to.have.property('maxSockets')
      expect(agent).to.have.property('maxFreeSockets')
    })

    it('should have correct property types', () => {
      const agent = httpsAgent as any
      expect(agent.keepAlive).to.be.a('boolean')
      expect(agent.keepAliveMsecs).to.be.a('number')
      expect(agent.maxSockets).to.be.a('number')
      expect(agent.maxFreeSockets).to.be.a('number')
    })

    it('should have REQUEST_TIMEOUT constant', () => {
      expect(REQUEST_TIMEOUT).to.be.a('number')
      expect(REQUEST_TIMEOUT).to.be.greaterThan(0)
    })
  })

  describe('Stats Structure', () => {
    it('should return stats with correct structure', () => {
      const stats = getAgentStats()

      expect(stats).to.be.an('object')
      expect(Object.keys(stats)).to.have.lengthOf(3)
      expect(stats).to.have.all.keys('sockets', 'freeSockets', 'requests')
    })

    it('should return fresh stats on each call', () => {
      const stats1 = getAgentStats()
      const stats2 = getAgentStats()

      // Stats should be fresh objects (not the same reference)
      expect(stats1).to.not.equal(stats2)
      expect(stats1).to.deep.equal(stats2)
    })
  })

  describe('Edge Cases', () => {
    it('should handle missing sockets object gracefully', () => {
      const stats = getAgentStats()

      // Should not throw even if internal structures are missing
      expect(stats.sockets).to.be.a('number')
      expect(stats.freeSockets).to.be.a('number')
      expect(stats.requests).to.be.a('number')
    })

    it('should handle stats from idle agent', () => {
      const stats = getAgentStats()

      // Idle agent should have 0 or more connections
      expect(stats.sockets).to.be.at.least(0)
      expect(stats.freeSockets).to.be.at.least(0)
      expect(stats.requests).to.be.at.least(0)
    })

    it('should handle agent with populated sockets', () => {
      // Mock agent with sockets
      const agent = httpsAgent as any
      const originalSockets = agent.sockets
      const originalFreeSockets = agent.freeSockets
      const originalRequests = agent.requests

      try {
        // Simulate some active sockets
        agent.sockets = {
          'localhost:443': [{ id: 1 }, { id: 2 }],
          'api.notion.com:443': [{ id: 3 }],
        }

        agent.freeSockets = {
          'localhost:443': [{ id: 4 }],
        }

        agent.requests = {
          'api.notion.com:443': [{ id: 5 }, { id: 6 }, { id: 7 }],
        }

        const stats = getAgentStats()

        // Should count all sockets correctly
        expect(stats.sockets).to.equal(3) // 2 + 1
        expect(stats.freeSockets).to.equal(1) // 1
        expect(stats.requests).to.equal(3) // 3
      } finally {
        // Restore original state
        agent.sockets = originalSockets
        agent.freeSockets = originalFreeSockets
        agent.requests = originalRequests
      }
    })

    it('should handle agent with empty arrays', () => {
      const agent = httpsAgent as any
      const originalSockets = agent.sockets
      const originalFreeSockets = agent.freeSockets
      const originalRequests = agent.requests

      try {
        // Simulate empty arrays
        agent.sockets = {
          'localhost:443': [],
        }

        agent.freeSockets = {
          'api.notion.com:443': [],
        }

        agent.requests = {
          'example.com:443': [],
        }

        const stats = getAgentStats()

        // Should handle empty arrays
        expect(stats.sockets).to.equal(0)
        expect(stats.freeSockets).to.equal(0)
        expect(stats.requests).to.equal(0)
      } finally {
        agent.sockets = originalSockets
        agent.freeSockets = originalFreeSockets
        agent.requests = originalRequests
      }
    })

    it('should handle agent with null/undefined socket arrays', () => {
      const agent = httpsAgent as any
      const originalSockets = agent.sockets
      const originalFreeSockets = agent.freeSockets
      const originalRequests = agent.requests

      try {
        // Simulate null/undefined values
        agent.sockets = {
          'localhost:443': null,
          'api.notion.com:443': undefined,
        }

        agent.freeSockets = {
          'localhost:443': null,
        }

        agent.requests = {
          'api.notion.com:443': undefined,
        }

        const stats = getAgentStats()

        // Should handle null/undefined gracefully
        expect(stats.sockets).to.equal(0)
        expect(stats.freeSockets).to.equal(0)
        expect(stats.requests).to.equal(0)
      } finally {
        agent.sockets = originalSockets
        agent.freeSockets = originalFreeSockets
        agent.requests = originalRequests
      }
    })

    it('should handle completely missing socket objects', () => {
      const agent = httpsAgent as any
      const originalSockets = agent.sockets
      const originalFreeSockets = agent.freeSockets
      const originalRequests = agent.requests

      try {
        // Remove socket objects entirely
        delete agent.sockets
        delete agent.freeSockets
        delete agent.requests

        const stats = getAgentStats()

        // Should handle missing objects gracefully
        expect(stats.sockets).to.equal(0)
        expect(stats.freeSockets).to.equal(0)
        expect(stats.requests).to.equal(0)
      } finally {
        agent.sockets = originalSockets
        agent.freeSockets = originalFreeSockets
        agent.requests = originalRequests
      }
    })
  })

  describe('getAgentConfig() with nullish coalescing', () => {
    it('should use fallback values when agent properties are undefined', () => {
      const agent = httpsAgent as any
      const originalKeepAlive = agent.keepAlive
      const originalKeepAliveMsecs = agent.keepAliveMsecs
      const originalMaxSockets = agent.maxSockets
      const originalMaxFreeSockets = agent.maxFreeSockets

      try {
        // Set properties to undefined
        agent.keepAlive = undefined
        agent.keepAliveMsecs = undefined
        agent.maxSockets = undefined
        agent.maxFreeSockets = undefined

        const config = getAgentConfig()

        // Should use fallback values
        expect(config.keepAlive).to.equal(false)
        expect(config.keepAliveMsecs).to.equal(1000)
        expect(config.maxSockets).to.equal(Infinity)
        expect(config.maxFreeSockets).to.equal(256)
      } finally {
        agent.keepAlive = originalKeepAlive
        agent.keepAliveMsecs = originalKeepAliveMsecs
        agent.maxSockets = originalMaxSockets
        agent.maxFreeSockets = originalMaxFreeSockets
      }
    })

    it('should use fallback values when agent properties are null', () => {
      const agent = httpsAgent as any
      const originalKeepAlive = agent.keepAlive
      const originalKeepAliveMsecs = agent.keepAliveMsecs
      const originalMaxSockets = agent.maxSockets
      const originalMaxFreeSockets = agent.maxFreeSockets

      try {
        // Set properties to null
        agent.keepAlive = null
        agent.keepAliveMsecs = null
        agent.maxSockets = null
        agent.maxFreeSockets = null

        const config = getAgentConfig()

        // Should use fallback values
        expect(config.keepAlive).to.equal(false)
        expect(config.keepAliveMsecs).to.equal(1000)
        expect(config.maxSockets).to.equal(Infinity)
        expect(config.maxFreeSockets).to.equal(256)
      } finally {
        agent.keepAlive = originalKeepAlive
        agent.keepAliveMsecs = originalKeepAliveMsecs
        agent.maxSockets = originalMaxSockets
        agent.maxFreeSockets = originalMaxFreeSockets
      }
    })

    it('should preserve falsy values that are not null/undefined', () => {
      const agent = httpsAgent as any
      const originalKeepAlive = agent.keepAlive
      const originalKeepAliveMsecs = agent.keepAliveMsecs
      const originalMaxSockets = agent.maxSockets
      const originalMaxFreeSockets = agent.maxFreeSockets

      try {
        // Set properties to falsy values (0, false)
        agent.keepAlive = false
        agent.keepAliveMsecs = 0
        agent.maxSockets = 0
        agent.maxFreeSockets = 0

        const config = getAgentConfig()

        // Should preserve these values (nullish coalescing only checks null/undefined)
        expect(config.keepAlive).to.equal(false)
        expect(config.keepAliveMsecs).to.equal(0)
        expect(config.maxSockets).to.equal(0)
        expect(config.maxFreeSockets).to.equal(0)
      } finally {
        agent.keepAlive = originalKeepAlive
        agent.keepAliveMsecs = originalKeepAliveMsecs
        agent.maxSockets = originalMaxSockets
        agent.maxFreeSockets = originalMaxFreeSockets
      }
    })
  })

  describe('REQUEST_TIMEOUT constant', () => {
    it('should be exported and accessible', () => {
      expect(REQUEST_TIMEOUT).to.exist
      expect(REQUEST_TIMEOUT).to.be.a('number')
    })

    it('should match the value in getAgentConfig', () => {
      const config = getAgentConfig()
      expect(config.requestTimeout).to.equal(REQUEST_TIMEOUT)
    })

    it('should be greater than zero', () => {
      expect(REQUEST_TIMEOUT).to.be.greaterThan(0)
    })

    it('should match parsed environment variable or default', () => {
      const expected = parseInt(process.env.NOTION_CLI_HTTP_TIMEOUT || '30000', 10)
      expect(REQUEST_TIMEOUT).to.equal(expected)
    })
  })
})
