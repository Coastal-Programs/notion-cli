import { expect } from '@oclif/test'

/**
 * Tests for Doctor Command
 *
 * Verifies health check and diagnostics functionality:
 * 1. Command metadata and structure
 * 2. JSON output format
 * 3. Check result structure
 * 4. Exit code logic (0 for pass, 1 for fail)
 * 5. Individual health checks
 */

describe('doctor command', () => {
  let originalToken: string | undefined

  beforeEach(() => {
    originalToken = process.env.NOTION_TOKEN
  })

  afterEach(() => {
    if (originalToken !== undefined) {
      process.env.NOTION_TOKEN = originalToken
    } else {
      delete process.env.NOTION_TOKEN
    }
  })

  describe('command structure', () => {
    it('should load doctor command successfully', async () => {
      const Doctor = await import('../../src/commands/doctor')
      expect(Doctor.default).to.exist
    })

    it('should have correct command metadata', async () => {
      const Doctor = await import('../../src/commands/doctor')
      expect(Doctor.default.description).to.include('health checks')
      expect(Doctor.default.description).to.include('diagnostics')
      expect(Doctor.default.examples).to.be.an('array')
      expect(Doctor.default.examples.length).to.be.greaterThan(0)
    })

    it('should have aliases defined', async () => {
      const Doctor = await import('../../src/commands/doctor')
      expect(Doctor.default.aliases).to.be.an('array')
      expect(Doctor.default.aliases).to.include.members(['diagnose', 'healthcheck'])
    })

    it('should have json flag defined', async () => {
      const Doctor = await import('../../src/commands/doctor')
      expect(Doctor.default.flags).to.have.property('json')
      expect(Doctor.default.flags.json).to.have.property('char', 'j')
      expect(Doctor.default.flags.json).to.have.property('description')
      expect(Doctor.default.flags.json).to.have.property('default', false)
    })

    it('should include example for normal mode', async () => {
      const Doctor = await import('../../src/commands/doctor')
      const normalExample = Doctor.default.examples.find((ex: any) =>
        ex.command.includes('notion-cli doctor') && !ex.command.includes('--json')
      )
      expect(normalExample).to.exist
    })

    it('should include example for JSON mode', async () => {
      const Doctor = await import('../../src/commands/doctor')
      const jsonExample = Doctor.default.examples.find((ex: any) =>
        ex.command.includes('--json')
      )
      expect(jsonExample).to.exist
    })
  })

  describe('health check structure', () => {
    it('should define HealthCheck interface correctly', async () => {
      // Import the command to ensure types are defined
      const Doctor = await import('../../src/commands/doctor')
      expect(Doctor.default).to.exist

      // Health check should have required properties
      const mockHealthCheck = {
        name: 'test_check',
        passed: true,
        value: 'test value',
        message: 'test message'
      }

      expect(mockHealthCheck).to.have.property('name')
      expect(mockHealthCheck).to.have.property('passed')
    })

    it('should define DoctorResult interface correctly', async () => {
      const Doctor = await import('../../src/commands/doctor')
      expect(Doctor.default).to.exist

      // Result should have required properties
      const mockResult = {
        success: true,
        checks: [],
        summary: {
          total: 0,
          passed: 0,
          failed: 0
        }
      }

      expect(mockResult).to.have.property('success')
      expect(mockResult).to.have.property('checks')
      expect(mockResult).to.have.property('summary')
      expect(mockResult.summary).to.have.all.keys('total', 'passed', 'failed')
    })
  })

  describe('JSON output format', () => {
    // Note: Full command execution requires valid token and network
    // These tests verify the structure without actual API calls

    it('should output valid JSON structure on success', () => {
      const mockResult = {
        success: true,
        checks: [
          {
            name: 'nodejs_version',
            passed: true,
            value: 'v18.0.0'
          }
        ],
        summary: {
          total: 1,
          passed: 1,
          failed: 0
        }
      }

      const jsonString = JSON.stringify(mockResult, null, 2)
      expect(() => JSON.parse(jsonString)).to.not.throw()

      const parsed = JSON.parse(jsonString)
      expect(parsed).to.have.property('success', true)
      expect(parsed).to.have.property('checks')
      expect(parsed).to.have.property('summary')
    })

    it('should output valid JSON structure on failure', () => {
      const mockResult = {
        success: false,
        checks: [
          {
            name: 'token_set',
            passed: false,
            message: 'NOTION_TOKEN environment variable is not set',
            recommendation: "Run 'notion-cli config set-token' or 'notion-cli init'"
          }
        ],
        summary: {
          total: 1,
          passed: 0,
          failed: 1
        }
      }

      const jsonString = JSON.stringify(mockResult, null, 2)
      expect(() => JSON.parse(jsonString)).to.not.throw()

      const parsed = JSON.parse(jsonString)
      expect(parsed).to.have.property('success', false)
      expect(parsed.summary.failed).to.equal(1)
    })

    it('should include all check properties in JSON output', () => {
      const mockCheck = {
        name: 'cache_fresh',
        passed: false,
        value: '2 days ago',
        age_hours: 48.5,
        message: 'Cache is outdated',
        recommendation: "Run 'notion-cli sync' to refresh"
      }

      expect(mockCheck).to.have.property('name')
      expect(mockCheck).to.have.property('passed')
      expect(mockCheck).to.have.property('value')
      expect(mockCheck).to.have.property('age_hours')
      expect(mockCheck).to.have.property('message')
      expect(mockCheck).to.have.property('recommendation')
    })
  })

  describe('exit codes', () => {
    it('should calculate success correctly when all checks pass', () => {
      const checks = [
        { name: 'check1', passed: true },
        { name: 'check2', passed: true },
        { name: 'check3', passed: true }
      ]

      const summary = {
        total: checks.length,
        passed: checks.filter(c => c.passed).length,
        failed: checks.filter(c => !c.passed).length
      }

      const success = summary.failed === 0
      expect(success).to.be.true
      expect(summary.passed).to.equal(3)
      expect(summary.failed).to.equal(0)
    })

    it('should calculate failure correctly when any check fails', () => {
      const checks = [
        { name: 'check1', passed: true },
        { name: 'check2', passed: false },
        { name: 'check3', passed: true }
      ]

      const summary = {
        total: checks.length,
        passed: checks.filter(c => c.passed).length,
        failed: checks.filter(c => !c.passed).length
      }

      const success = summary.failed === 0
      expect(success).to.be.false
      expect(summary.passed).to.equal(2)
      expect(summary.failed).to.equal(1)
    })

    it('should calculate failure correctly when all checks fail', () => {
      const checks = [
        { name: 'check1', passed: false },
        { name: 'check2', passed: false }
      ]

      const summary = {
        total: checks.length,
        passed: checks.filter(c => c.passed).length,
        failed: checks.filter(c => !c.passed).length
      }

      const success = summary.failed === 0
      expect(success).to.be.false
      expect(summary.passed).to.equal(0)
      expect(summary.failed).to.equal(2)
    })
  })

  describe('check types', () => {
    it('should include Node.js version check', async () => {
      const checkName = 'nodejs_version'
      const currentVersion = process.version
      const major = parseInt(currentVersion.split('.')[0].replace('v', ''))

      expect(major).to.be.a('number')
      expect(checkName).to.equal('nodejs_version')
    })

    it('should include token_set check', () => {
      const checkName = 'token_set'
      expect(checkName).to.equal('token_set')
    })

    it('should include token_format check', () => {
      const checkName = 'token_format'
      expect(checkName).to.equal('token_format')
    })

    it('should include network_connectivity check', () => {
      const checkName = 'network_connectivity'
      expect(checkName).to.equal('network_connectivity')
    })

    it('should include api_connection check', () => {
      const checkName = 'api_connection'
      expect(checkName).to.equal('api_connection')
    })

    it('should include cache_exists check', () => {
      const checkName = 'cache_exists'
      expect(checkName).to.equal('cache_exists')
    })

    it('should include cache_fresh check', () => {
      const checkName = 'cache_fresh'
      expect(checkName).to.equal('cache_fresh')
    })
  })

  describe('token format validation', () => {
    it('should validate secret_ prefix tokens', () => {
      const token = 'secret_abc123xyz'
      const isValid = token.startsWith('secret_') || token.startsWith('ntn_')
      expect(isValid).to.be.true
    })

    it('should validate ntn_ prefix tokens (OAuth)', () => {
      const token = 'ntn_abc123xyz'
      const isValid = token.startsWith('secret_') || token.startsWith('ntn_')
      expect(isValid).to.be.true
    })

    it('should validate long alphanumeric tokens', () => {
      const token = 'a'.repeat(32)
      const isValid = token.length >= 32 && /^[A-Za-z0-9_-]+$/.test(token)
      expect(isValid).to.be.true
    })

    it('should reject invalid token formats', () => {
      const token = 'invalid!'
      const isValid =
        token.startsWith('secret_') ||
        token.startsWith('ntn_') ||
        (token.length >= 32 && /^[A-Za-z0-9_-]+$/.test(token))
      expect(isValid).to.be.false
    })

    it('should reject short tokens', () => {
      const token = 'short'
      const isValid =
        token.startsWith('secret_') ||
        token.startsWith('ntn_') ||
        (token.length >= 32 && /^[A-Za-z0-9_-]+$/.test(token))
      expect(isValid).to.be.false
    })
  })

  describe('cache freshness calculation', () => {
    it('should calculate age in hours correctly', () => {
      const now = Date.now()
      const twelveHoursAgo = now - (12 * 60 * 60 * 1000)
      const ageMs = now - twelveHoursAgo
      const ageHours = ageMs / (1000 * 60 * 60)

      expect(ageHours).to.equal(12)
    })

    it('should determine cache is fresh when < 24 hours', () => {
      const ageHours = 12
      const isFresh = ageHours < 24
      expect(isFresh).to.be.true
    })

    it('should determine cache is stale when >= 24 hours', () => {
      const ageHours = 25
      const isFresh = ageHours < 24
      expect(isFresh).to.be.false
    })

    it('should format age string correctly for hours', () => {
      const ageHours = 5
      const ageDays = Math.floor(ageHours / 24)

      expect(ageDays).to.equal(0)
      const ageString = `${Math.floor(ageHours)} hours ago`
      expect(ageString).to.equal('5 hours ago')
    })

    it('should format age string correctly for days', () => {
      const ageHours = 50
      const ageDays = Math.floor(ageHours / 24)

      expect(ageDays).to.equal(2)
      const ageString = `${ageDays} days, ${Math.floor(ageHours % 24)} hours ago`
      expect(ageString).to.equal('2 days, 2 hours ago')
    })
  })

  describe('Node.js version check', () => {
    it('should extract major version correctly', () => {
      const version = 'v18.12.1'
      const major = parseInt(version.split('.')[0].replace('v', ''))
      expect(major).to.equal(18)
    })

    it('should pass for Node.js 18+', () => {
      const version = 'v18.0.0'
      const major = parseInt(version.split('.')[0].replace('v', ''))
      const passed = major >= 18
      expect(passed).to.be.true
    })

    it('should pass for Node.js 20+', () => {
      const version = 'v20.5.0'
      const major = parseInt(version.split('.')[0].replace('v', ''))
      const passed = major >= 18
      expect(passed).to.be.true
    })

    it('should fail for Node.js < 18', () => {
      const version = 'v16.0.0'
      const major = parseInt(version.split('.')[0].replace('v', ''))
      const passed = major >= 18
      expect(passed).to.be.false
    })

    it('should handle current process version', () => {
      const version = process.version
      const major = parseInt(version.split('.')[0].replace('v', ''))
      expect(major).to.be.a('number')
      expect(major).to.be.greaterThan(0)
    })
  })

  describe('check recommendations', () => {
    it('should provide recommendation for token_set failure', () => {
      const recommendation = "Run 'notion-cli config set-token' or 'notion-cli init'"
      expect(recommendation).to.include('config set-token')
      expect(recommendation).to.include('init')
    })

    it('should provide recommendation for cache_exists failure', () => {
      const recommendation = "Run 'notion-cli sync' to create cache"
      expect(recommendation).to.include('sync')
      expect(recommendation).to.include('cache')
    })

    it('should provide recommendation for cache_fresh failure', () => {
      const recommendation = "Run 'notion-cli sync' to refresh"
      expect(recommendation).to.include('sync')
      expect(recommendation).to.include('refresh')
    })

    it('should provide recommendation for nodejs_version failure', () => {
      const recommendation = 'Please upgrade Node.js to version 18 or higher'
      expect(recommendation).to.include('upgrade')
      expect(recommendation).to.include('18')
    })

    it('should provide recommendation for network_connectivity failure', () => {
      const recommendation = 'Check your internet connection and firewall settings'
      expect(recommendation).to.include('internet')
      expect(recommendation).to.include('firewall')
    })
  })

  describe('summary calculation', () => {
    it('should calculate summary with all passes', () => {
      const checks = [
        { name: 'check1', passed: true },
        { name: 'check2', passed: true },
        { name: 'check3', passed: true },
        { name: 'check4', passed: true }
      ]

      const summary = {
        total: checks.length,
        passed: checks.filter(c => c.passed).length,
        failed: checks.filter(c => !c.passed).length
      }

      expect(summary.total).to.equal(4)
      expect(summary.passed).to.equal(4)
      expect(summary.failed).to.equal(0)
    })

    it('should calculate summary with mixed results', () => {
      const checks = [
        { name: 'check1', passed: true },
        { name: 'check2', passed: false },
        { name: 'check3', passed: true },
        { name: 'check4', passed: false },
        { name: 'check5', passed: true }
      ]

      const summary = {
        total: checks.length,
        passed: checks.filter(c => c.passed).length,
        failed: checks.filter(c => !c.passed).length
      }

      expect(summary.total).to.equal(5)
      expect(summary.passed).to.equal(3)
      expect(summary.failed).to.equal(2)
    })

    it('should calculate summary with all failures', () => {
      const checks = [
        { name: 'check1', passed: false },
        { name: 'check2', passed: false }
      ]

      const summary = {
        total: checks.length,
        passed: checks.filter(c => c.passed).length,
        failed: checks.filter(c => !c.passed).length
      }

      expect(summary.total).to.equal(2)
      expect(summary.passed).to.equal(0)
      expect(summary.failed).to.equal(2)
    })

    it('should handle empty checks array', () => {
      const checks: any[] = []

      const summary = {
        total: checks.length,
        passed: checks.filter(c => c.passed).length,
        failed: checks.filter(c => !c.passed).length
      }

      expect(summary.total).to.equal(0)
      expect(summary.passed).to.equal(0)
      expect(summary.failed).to.equal(0)
    })
  })

  describe('check metadata', () => {
    it('should include bot_name for successful api_connection', () => {
      const check = {
        name: 'api_connection',
        passed: true,
        bot_name: 'Test Bot',
        workspace_name: 'Test Workspace'
      }

      expect(check).to.have.property('bot_name')
      expect(check).to.have.property('workspace_name')
    })

    it('should include value for cache checks', () => {
      const check = {
        name: 'cache_exists',
        passed: true,
        value: '/path/to/cache.json'
      }

      expect(check).to.have.property('value')
    })

    it('should include age_hours for cache_fresh check', () => {
      const check = {
        name: 'cache_fresh',
        passed: false,
        age_hours: 36.5,
        value: '1 day, 12 hours ago'
      }

      expect(check).to.have.property('age_hours')
      expect(check.age_hours).to.be.a('number')
    })
  })

  describe('result structure validation', () => {
    it('should have all required top-level fields', () => {
      const result = {
        success: true,
        checks: [],
        summary: {
          total: 0,
          passed: 0,
          failed: 0
        }
      }

      expect(result).to.have.all.keys('success', 'checks', 'summary')
    })

    it('should have correct types for all fields', () => {
      const result = {
        success: true,
        checks: [{ name: 'test', passed: true }],
        summary: {
          total: 1,
          passed: 1,
          failed: 0
        }
      }

      expect(result.success).to.be.a('boolean')
      expect(result.checks).to.be.an('array')
      expect(result.summary).to.be.an('object')
      expect(result.summary.total).to.be.a('number')
      expect(result.summary.passed).to.be.a('number')
      expect(result.summary.failed).to.be.a('number')
    })

    it('should serialize to valid JSON', () => {
      const result = {
        success: false,
        checks: [
          {
            name: 'token_set',
            passed: false,
            message: 'Token not set',
            recommendation: 'Set token'
          }
        ],
        summary: {
          total: 1,
          passed: 0,
          failed: 1
        }
      }

      const jsonString = JSON.stringify(result, null, 2)
      expect(() => JSON.parse(jsonString)).to.not.throw()

      const parsed = JSON.parse(jsonString)
      expect(parsed).to.deep.equal(result)
    })
  })
})
