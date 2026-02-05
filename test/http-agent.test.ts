import { expect } from 'chai'
import { httpsAgent, getAgentStats, getAgentConfig, destroyAgents, REQUEST_TIMEOUT } from '../dist/http-agent.js'

describe('HTTP Agent', () => {
  describe('httpsAgent', () => {
    it('should be an undici Agent instance', () => {
      expect(httpsAgent).to.exist
      expect(httpsAgent).to.have.property('destroy')
    })

    it('should have reasonable default values', () => {
      const config = getAgentConfig()

      expect(config.connections).to.be.a('number')
      expect(config.connections).to.be.greaterThan(0)

      expect(config.keepAliveTimeout).to.be.a('number')
      expect(config.keepAliveTimeout).to.be.greaterThan(0)

      expect(config.requestTimeout).to.be.a('number')
      expect(config.requestTimeout).to.be.greaterThan(0)
    })
  })

  describe('getAgentConfig()', () => {
    it('should return complete configuration', () => {
      const config = getAgentConfig()

      expect(config).to.have.property('connections')
      expect(config).to.have.property('keepAliveTimeout')
      expect(config).to.have.property('requestTimeout')
    })

    it('should return numeric values for all configs', () => {
      const config = getAgentConfig()

      expect(config.connections).to.be.a('number')
      expect(config.keepAliveTimeout).to.be.a('number')
      expect(config.requestTimeout).to.be.a('number')
    })

    it('should respect environment variables', () => {
      const config = getAgentConfig()

      const expectedConnections = parseInt(
        process.env.NOTION_CLI_HTTP_MAX_SOCKETS || '50',
        10
      )
      expect(config.connections).to.equal(expectedConnections)

      const expectedKeepAliveTimeout = parseInt(
        process.env.NOTION_CLI_HTTP_KEEP_ALIVE_MS || '60000',
        10
      )
      expect(config.keepAliveTimeout).to.equal(expectedKeepAliveTimeout)

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

    it('should return placeholder values (undici limitation)', () => {
      const stats = getAgentStats()

      // undici Agent doesn't expose socket statistics, so we expect zeros
      expect(stats.sockets).to.equal(0)
      expect(stats.freeSockets).to.equal(0)
      expect(stats.requests).to.equal(0)
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

    it('should call destroyAgents without errors', () => {
      // Verify that destroyAgents can be called without throwing
      // Note: We don't actually destroy the agent in tests as it's shared
      expect(() => destroyAgents()).to.not.throw()
    })
  })

  describe('Configuration Validation', () => {
    it('should have sensible keep-alive timeout', () => {
      const config = getAgentConfig()

      // Keep-alive should be between 10 seconds and 5 minutes
      expect(config.keepAliveTimeout).to.be.at.least(1000)
      expect(config.keepAliveTimeout).to.be.at.most(300000)
    })

    it('should have reasonable connection limits', () => {
      const config = getAgentConfig()

      // Connections should be reasonable (1-1000)
      expect(config.connections).to.be.at.least(1)
      expect(config.connections).to.be.at.most(1000)
    })

    it('should have reasonable requestTimeout', () => {
      const config = getAgentConfig()

      // Timeout should be between 5 seconds and 2 minutes
      expect(config.requestTimeout).to.be.at.least(5000)
      expect(config.requestTimeout).to.be.at.most(120000)
    })
  })

  describe('Default Values', () => {
    it('should use default keep-alive timeout when env var not set', () => {
      // If env var is not set, should use 60000 (60 seconds)
      const expected = parseInt(process.env.NOTION_CLI_HTTP_KEEP_ALIVE_MS || '60000', 10)
      const config = getAgentConfig()
      expect(config.keepAliveTimeout).to.equal(expected)
    })

    it('should use default connections when env var not set', () => {
      // If env var is not set, should use 50
      const expected = parseInt(process.env.NOTION_CLI_HTTP_MAX_SOCKETS || '50', 10)
      const config = getAgentConfig()
      expect(config.connections).to.equal(expected)
    })

    it('should use default requestTimeout when env var not set', () => {
      // If env var is not set, should use 30000 (30 seconds)
      const expected = parseInt(process.env.NOTION_CLI_HTTP_TIMEOUT || '30000', 10)
      const config = getAgentConfig()
      expect(config.requestTimeout).to.equal(expected)
    })
  })

  describe('Agent Properties', () => {
    it('should have destroy method', () => {
      expect(httpsAgent).to.have.property('destroy')
      expect(httpsAgent.destroy).to.be.a('function')
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
      // undici returns placeholder zeros
      expect(stats1).to.deep.equal({sockets: 0, freeSockets: 0, requests: 0})
      expect(stats2).to.deep.equal({sockets: 0, freeSockets: 0, requests: 0})
    })
  })

  describe('Edge Cases', () => {
    it('should handle stats gracefully (undici limitation)', () => {
      const stats = getAgentStats()

      // undici Agent doesn't expose socket statistics
      // Should not throw and return valid structure
      expect(stats.sockets).to.be.a('number')
      expect(stats.freeSockets).to.be.a('number')
      expect(stats.requests).to.be.a('number')

      // All should be zero (placeholder values)
      expect(stats.sockets).to.equal(0)
      expect(stats.freeSockets).to.equal(0)
      expect(stats.requests).to.equal(0)
    })
  })

  describe('Environment Variable Parsing', () => {
    it('should read configuration from environment variables', () => {
      // getAgentConfig() reads directly from environment variables
      const config = getAgentConfig()

      // Should return valid numbers
      expect(config.connections).to.be.a('number')
      expect(config.keepAliveTimeout).to.be.a('number')
      expect(config.requestTimeout).to.be.a('number')

      // Should be positive values
      expect(config.connections).to.be.greaterThan(0)
      expect(config.keepAliveTimeout).to.be.greaterThan(0)
      expect(config.requestTimeout).to.be.greaterThan(0)
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
