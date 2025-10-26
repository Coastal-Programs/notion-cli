import { expect } from 'chai'
import { validateNotionToken, maskToken } from '../../src/utils/token-validator'
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

  describe('maskToken', () => {
    it('should mask standard Notion tokens with secret_ prefix', () => {
      const token = 'secret_1234567890abcdefghijklmnopqrstuvwxyz'
      const masked = maskToken(token)

      expect(masked).to.equal('secret_***...***xyz')
    })

    it('should mask OAuth tokens with ntn_ prefix', () => {
      const token = 'ntn_1234567890abcdefghijklmnopqrstuvwxyz'
      const masked = maskToken(token)

      expect(masked).to.equal('ntn_***...***xyz')
    })

    it('should preserve last 3 characters', () => {
      const token = 'secret_abcdefghijklmnopqrstuvwxyz123'
      const masked = maskToken(token)

      expect(masked).to.include('123')
      expect(masked.endsWith('123')).to.be.true
    })

    it('should handle tokens with unknown prefixes', () => {
      const token = 'custom_1234567890abcdefghijklmnopqrstuvwxyz'
      const masked = maskToken(token)

      // Should use max 4 chars for unknown prefixes (security: ensures at least 4 chars masked)
      expect(masked).to.equal('cust***...***xyz')
    })

    it('should mask at least 4 characters for short unknown-prefix tokens', () => {
      const token = 'mysecret123' // 11 chars, no standard prefix
      const masked = maskToken(token)

      // Should mask at least 4 chars: myse[cret]123
      expect(masked).to.equal('myse***...***123')
      expect(masked).to.not.include('mysecret')
    })

    it('should completely obscure short tokens', () => {
      const token = 'short'
      const masked = maskToken(token)

      expect(masked).to.equal('***')
      expect(masked).to.not.include('short')
    })

    it('should handle empty strings', () => {
      const token = ''
      const masked = maskToken(token)

      expect(masked).to.equal('')
    })

    it('should never expose the full token', () => {
      const token = 'secret_very_sensitive_token_value_12345'
      const masked = maskToken(token)

      expect(masked).to.not.include('sensitive')
      expect(masked).to.not.include('token_value')
      expect(masked.length).to.be.lessThan(token.length)
    })

    it('should be consistent for the same token', () => {
      const token = 'secret_consistent_token_value_abc'
      const masked1 = maskToken(token)
      const masked2 = maskToken(token)

      expect(masked1).to.equal(masked2)
    })

    it('should handle tokens at the minimum safe length boundary', () => {
      const token = 'secret_abc' // Exactly 10 chars
      const masked = maskToken(token)

      expect(masked).to.equal('***')
    })

    it('should handle tokens just above the minimum safe length', () => {
      const token = 'secret_abcd' // 11 chars
      const masked = maskToken(token)

      expect(masked).to.include('secret_')
      expect(masked).to.include('bcd')
      expect(masked).to.include('***...***')
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
