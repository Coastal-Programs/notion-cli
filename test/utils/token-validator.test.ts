import { expect } from 'chai'
import { validateNotionToken } from '../../src/utils/token-validator'
import { NotionCLIError, NotionCLIErrorCode } from '../../src/errors/enhanced-errors'

/**
 * Tests for Token Validator Utility
 *
 * Verifies that the token validator:
 * 1. Throws NotionCLIError when NOTION_TOKEN is not set
 * 2. Provides platform-specific error messages
 * 3. Succeeds when a valid token is set
 * 4. Handles different token scenarios correctly
 */

describe('token-validator', () => {
  let originalToken: string | undefined

  // Save original token before each test
  beforeEach(() => {
    originalToken = process.env.NOTION_TOKEN
  })

  // Restore original token after each test
  afterEach(() => {
    if (originalToken !== undefined) {
      process.env.NOTION_TOKEN = originalToken
    } else {
      delete process.env.NOTION_TOKEN
    }
  })

  describe('validateNotionToken', () => {
    it('should throw NotionCLIError when NOTION_TOKEN is not set', () => {
      delete process.env.NOTION_TOKEN

      try {
        validateNotionToken()
        expect.fail('Should have thrown NotionCLIError')
      } catch (error) {
        expect(error).to.be.instanceOf(NotionCLIError)
        expect((error as NotionCLIError).code).to.equal(NotionCLIErrorCode.TOKEN_MISSING)
      }
    })

    it('should throw NotionCLIError when NOTION_TOKEN is empty string', () => {
      process.env.NOTION_TOKEN = ''

      try {
        validateNotionToken()
        expect.fail('Should have thrown NotionCLIError')
      } catch (error) {
        expect(error).to.be.instanceOf(NotionCLIError)
        expect((error as NotionCLIError).code).to.equal(NotionCLIErrorCode.TOKEN_MISSING)
      }
    })

    it('should include Unix/Mac instructions in error suggestions', () => {
      delete process.env.NOTION_TOKEN

      try {
        validateNotionToken()
        expect.fail('Should have thrown NotionCLIError')
      } catch (error) {
        const cliError = error as NotionCLIError
        const hasUnixInstructions = cliError.suggestions.some(s =>
          s.command?.includes('export NOTION_TOKEN=')
        )
        expect(hasUnixInstructions).to.be.true
      }
    })

    it('should include Windows PowerShell instructions in error suggestions', () => {
      delete process.env.NOTION_TOKEN

      try {
        validateNotionToken()
        expect.fail('Should have thrown NotionCLIError')
      } catch (error) {
        const cliError = error as NotionCLIError
        const hasWindowsInstructions = cliError.suggestions.some(s =>
          s.command?.includes('$env:NOTION_TOKEN=')
        )
        expect(hasWindowsInstructions).to.be.true
      }
    })

    it('should include config set-token command in suggestions', () => {
      delete process.env.NOTION_TOKEN

      try {
        validateNotionToken()
        expect.fail('Should have thrown NotionCLIError')
      } catch (error) {
        const cliError = error as NotionCLIError
        const hasConfigCommand = cliError.suggestions.some(s =>
          s.command?.includes('notion-cli config set-token')
        )
        expect(hasConfigCommand).to.be.true
      }
    })

    it('should include link to Notion integration docs', () => {
      delete process.env.NOTION_TOKEN

      try {
        validateNotionToken()
        expect.fail('Should have thrown NotionCLIError')
      } catch (error) {
        const cliError = error as NotionCLIError
        const hasDocsLink = cliError.suggestions.some(s =>
          s.link?.includes('developers.notion.com')
        )
        expect(hasDocsLink).to.be.true
      }
    })

    it('should not throw when NOTION_TOKEN is set with valid format', () => {
      process.env.NOTION_TOKEN = 'secret_test_token_1234567890'

      expect(() => validateNotionToken()).to.not.throw()
    })

    it('should not throw when NOTION_TOKEN is set with OAuth format', () => {
      process.env.NOTION_TOKEN = 'ntn_test_oauth_token_1234567890'

      expect(() => validateNotionToken()).to.not.throw()
    })

    it('should not throw when NOTION_TOKEN is set with any non-empty value', () => {
      process.env.NOTION_TOKEN = 'any_token_value'

      expect(() => validateNotionToken()).to.not.throw()
    })

    it('should provide descriptive error message', () => {
      delete process.env.NOTION_TOKEN

      try {
        validateNotionToken()
        expect.fail('Should have thrown NotionCLIError')
      } catch (error) {
        const cliError = error as NotionCLIError
        expect(cliError.userMessage).to.include('NOTION_TOKEN')
        expect(cliError.userMessage).to.include('not set')
      }
    })

    it('should set metadata indicating token is not set', () => {
      delete process.env.NOTION_TOKEN

      try {
        validateNotionToken()
        expect.fail('Should have thrown NotionCLIError')
      } catch (error) {
        const cliError = error as NotionCLIError
        expect(cliError.context.metadata?.tokenSet).to.be.false
      }
    })
  })

  describe('error JSON format', () => {
    it('should produce valid JSON output for automation', () => {
      delete process.env.NOTION_TOKEN

      try {
        validateNotionToken()
        expect.fail('Should have thrown NotionCLIError')
      } catch (error) {
        const cliError = error as NotionCLIError
        const json = cliError.toJSON()

        expect(json).to.have.property('success', false)
        expect(json).to.have.property('error')
        expect(json.error).to.have.property('code', NotionCLIErrorCode.TOKEN_MISSING)
        expect(json.error).to.have.property('message')
        expect(json.error).to.have.property('suggestions')
        expect(json.error).to.have.property('timestamp')
      }
    })

    it('should produce compact JSON string', () => {
      delete process.env.NOTION_TOKEN

      try {
        validateNotionToken()
        expect.fail('Should have thrown NotionCLIError')
      } catch (error) {
        const cliError = error as NotionCLIError
        const compactJson = cliError.toCompactJSON()

        expect(compactJson).to.be.a('string')
        expect(() => JSON.parse(compactJson)).to.not.throw()
      }
    })
  })

  describe('error human-readable format', () => {
    it('should produce formatted human-readable output', () => {
      delete process.env.NOTION_TOKEN

      try {
        validateNotionToken()
        expect.fail('Should have thrown NotionCLIError')
      } catch (error) {
        const cliError = error as NotionCLIError
        const humanString = cliError.toHumanString()

        expect(humanString).to.include('❌')
        expect(humanString).to.include('Error Code')
        expect(humanString).to.include('💡')
        expect(humanString).to.include('NOTION_TOKEN')
      }
    })
  })
})
