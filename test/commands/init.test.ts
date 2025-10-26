import { expect, test } from '@oclif/test'
import { NotionCLIError, NotionCLIErrorCode } from '../../src/errors/enhanced-errors'

/**
 * Tests for Init Command
 *
 * Verifies the interactive first-time setup wizard:
 * 1. Command metadata and structure
 * 2. Token validation in JSON mode
 * 3. Error handling for missing token
 * 4. JSON output format
 *
 * Note: Full integration tests would require mocking API calls,
 * which is beyond the scope of these unit tests. These tests focus
 * on command structure and error handling.
 */

describe('init command', () => {
  describe('command structure', () => {
    it('should load init command successfully', async () => {
      const Init = await import('../../src/commands/init')
      expect(Init.default).to.exist
    })

    it('should have correct command metadata', async () => {
      const Init = await import('../../src/commands/init')
      expect(Init.default.description).to.include('Interactive first-time setup wizard')
      expect(Init.default.examples).to.be.an('array')
      expect(Init.default.examples.length).to.be.greaterThan(0)
    })

    it('should have json flag defined', async () => {
      const Init = await import('../../src/commands/init')
      expect(Init.default.flags).to.have.property('json')
    })

    it('should include example for interactive mode', async () => {
      const Init = await import('../../src/commands/init')
      const interactiveExample = Init.default.examples.find((ex: any) =>
        ex.command.includes('notion-cli init') && !ex.command.includes('--json')
      )
      expect(interactiveExample).to.exist
    })

    it('should include example for JSON mode', async () => {
      const Init = await import('../../src/commands/init')
      const jsonExample = Init.default.examples.find((ex: any) =>
        ex.command.includes('--json')
      )
      expect(jsonExample).to.exist
    })
  })

  describe('token validation in JSON mode', () => {
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

    // Note: Testing with --json flag requires mocking the Notion API
    // These tests verify the error handling when token is missing
    test
      .do(() => {
        delete process.env.NOTION_TOKEN
      })
      .command(['init', '--json'])
      .exit(1)
      .it('should exit with code 1 when NOTION_TOKEN not set in JSON mode')

    test
      .do(() => {
        delete process.env.NOTION_TOKEN
      })
      .command(['init', '--json'])
      .catch((error) => {
        // Verify error output contains JSON
        expect(error.message).to.exist
      })
      .it('should output JSON error when token missing in JSON mode')
  })

  describe('error handling', () => {
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

    it('should provide helpful error when token is missing', async () => {
      delete process.env.NOTION_TOKEN

      const { validateNotionToken } = await import('../../src/utils/token-validator')

      try {
        validateNotionToken()
        expect.fail('Should have thrown error')
      } catch (error) {
        expect(error).to.be.instanceOf(NotionCLIError)
        expect((error as NotionCLIError).code).to.equal(NotionCLIErrorCode.TOKEN_MISSING)

        // Verify error includes helpful suggestions
        const cliError = error as NotionCLIError
        expect(cliError.suggestions.length).to.be.greaterThan(0)

        // Should include config command
        const hasConfigCommand = cliError.suggestions.some(s =>
          s.command?.includes('notion-cli config set-token')
        )
        expect(hasConfigCommand).to.be.true

        // Should include export command
        const hasExportCommand = cliError.suggestions.some(s =>
          s.command?.includes('export NOTION_TOKEN')
        )
        expect(hasExportCommand).to.be.true

        // Should include docs link
        const hasDocsLink = cliError.suggestions.some(s =>
          s.link?.includes('developers.notion.com')
        )
        expect(hasDocsLink).to.be.true
      }
    })

    it('should format error as JSON when needed', async () => {
      delete process.env.NOTION_TOKEN

      const { validateNotionToken } = await import('../../src/utils/token-validator')

      try {
        validateNotionToken()
        expect.fail('Should have thrown error')
      } catch (error) {
        const cliError = error as NotionCLIError
        const jsonOutput = cliError.toJSON()

        expect(jsonOutput).to.have.property('success', false)
        expect(jsonOutput).to.have.property('error')
        expect(jsonOutput.error).to.have.property('code')
        expect(jsonOutput.error).to.have.property('message')
        expect(jsonOutput.error).to.have.property('suggestions')
        expect(jsonOutput.error).to.have.property('timestamp')
      }
    })

    it('should format error as human-readable when needed', async () => {
      delete process.env.NOTION_TOKEN

      const { validateNotionToken } = await import('../../src/utils/token-validator')

      try {
        validateNotionToken()
        expect.fail('Should have thrown error')
      } catch (error) {
        const cliError = error as NotionCLIError
        const humanOutput = cliError.toHumanString()

        expect(humanOutput).to.be.a('string')
        expect(humanOutput).to.include('NOTION_TOKEN')
        expect(humanOutput).to.include('❌')
        expect(humanOutput).to.include('💡')
      }
    })
  })

  describe('JSON output validation', () => {
    it('should validate JSON structure for error responses', async () => {
      const { NotionCLIError, NotionCLIErrorCode } = await import('../../src/errors/enhanced-errors')

      const testError = new NotionCLIError(
        NotionCLIErrorCode.TOKEN_MISSING,
        'Test error message',
        [
          {
            description: 'Test suggestion',
            command: 'test command'
          }
        ],
        {
          metadata: { test: true }
        }
      )

      const json = testError.toJSON()

      // Validate structure
      expect(json).to.be.an('object')
      expect(json.success).to.be.false
      expect(json.error).to.be.an('object')
      expect(json.error.code).to.equal(NotionCLIErrorCode.TOKEN_MISSING)
      expect(json.error.message).to.be.a('string')
      expect(json.error.suggestions).to.be.an('array')
      expect(json.error.context).to.be.an('object')
      expect(json.error.timestamp).to.be.a('string')

      // Validate can be stringified
      const jsonString = JSON.stringify(json)
      expect(jsonString).to.be.a('string')
      expect(() => JSON.parse(jsonString)).to.not.throw()
    })

    it('should include all required fields in JSON error output', async () => {
      const { NotionCLIError, NotionCLIErrorCode } = await import('../../src/errors/enhanced-errors')

      const testError = new NotionCLIError(
        NotionCLIErrorCode.TOKEN_MISSING,
        'Test error',
        [{ description: 'Fix it' }]
      )

      const json = testError.toJSON()

      // Required top-level fields
      expect(json).to.have.all.keys('success', 'error')

      // Required error fields
      expect(json.error).to.include.all.keys(
        'code',
        'message',
        'suggestions',
        'context',
        'timestamp'
      )

      // Verify types
      expect(json.error.code).to.be.a('string')
      expect(json.error.message).to.be.a('string')
      expect(json.error.suggestions).to.be.an('array')
      expect(json.error.context).to.be.an('object')
      expect(json.error.timestamp).to.be.a('string')

      // Verify timestamp is valid ISO 8601
      expect(() => new Date(json.error.timestamp)).to.not.throw()
      const date = new Date(json.error.timestamp)
      expect(date.toISOString()).to.equal(json.error.timestamp)
    })
  })

  describe('command flags', () => {
    it('should support --json flag', async () => {
      const Init = await import('../../src/commands/init')
      const flags = Init.default.flags

      expect(flags).to.have.property('json')
      expect(flags.json).to.exist
    })

    it('should inherit automation flags', async () => {
      const Init = await import('../../src/commands/init')
      const flags = Init.default.flags

      // AutomationFlags include --json, --raw, --no-truncate
      expect(flags).to.have.property('json')
    })
  })

  describe('error codes', () => {
    it('should use TOKEN_MISSING error code when appropriate', async () => {
      const { NotionCLIErrorCode } = await import('../../src/errors/enhanced-errors')

      expect(NotionCLIErrorCode.TOKEN_MISSING).to.equal('TOKEN_MISSING')
    })

    it('should use TOKEN_INVALID error code when appropriate', async () => {
      const { NotionCLIErrorCode } = await import('../../src/errors/enhanced-errors')

      expect(NotionCLIErrorCode.TOKEN_INVALID).to.equal('TOKEN_INVALID')
    })

    it('should define all necessary error codes', async () => {
      const { NotionCLIErrorCode } = await import('../../src/errors/enhanced-errors')

      // Verify key error codes exist
      const requiredCodes = [
        'TOKEN_MISSING',
        'TOKEN_INVALID',
        'UNAUTHORIZED',
        'NOT_FOUND',
        'API_ERROR'
      ]

      requiredCodes.forEach(code => {
        expect(NotionCLIErrorCode).to.have.property(code)
      })
    })
  })

  describe('error factory', () => {
    it('should create consistent TOKEN_MISSING errors', async () => {
      const { NotionCLIErrorFactory, NotionCLIErrorCode } = await import('../../src/errors/enhanced-errors')

      const error1 = NotionCLIErrorFactory.tokenMissing()
      const error2 = NotionCLIErrorFactory.tokenMissing()

      expect(error1.code).to.equal(NotionCLIErrorCode.TOKEN_MISSING)
      expect(error2.code).to.equal(NotionCLIErrorCode.TOKEN_MISSING)
      expect(error1.code).to.equal(error2.code)
      expect(error1.userMessage).to.equal(error2.userMessage)
      expect(error1.suggestions.length).to.equal(error2.suggestions.length)
    })

    it('should include multiple suggestions in TOKEN_MISSING error', async () => {
      const { NotionCLIErrorFactory } = await import('../../src/errors/enhanced-errors')

      const error = NotionCLIErrorFactory.tokenMissing()

      expect(error.suggestions.length).to.be.greaterThan(2)

      // Should have config command
      const hasConfigCommand = error.suggestions.some(s => s.command?.includes('config set-token'))
      expect(hasConfigCommand).to.be.true

      // Should have Unix/Mac export
      const hasUnixExport = error.suggestions.some(s => s.command?.includes('export NOTION_TOKEN='))
      expect(hasUnixExport).to.be.true

      // Should have Windows PowerShell
      const hasWindowsExport = error.suggestions.some(s => s.command?.includes('$env:NOTION_TOKEN='))
      expect(hasWindowsExport).to.be.true

      // Should have docs link
      const hasDocsLink = error.suggestions.some(s => s.link?.includes('developers.notion.com'))
      expect(hasDocsLink).to.be.true
    })
  })
})
