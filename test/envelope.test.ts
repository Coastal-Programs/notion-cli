/**
 * Unit tests for EnvelopeFormatter
 *
 * Tests the core envelope system including:
 * - Success envelope creation
 * - Error envelope creation
 * - Metadata tracking
 * - Exit code determination
 * - Suggestion generation
 */

import { expect } from 'chai'
import { EnvelopeFormatter, ExitCode, isSuccessEnvelope, isErrorEnvelope } from '../src/envelope'
import { NotionCLIError, ErrorCode } from '../src/errors'

describe('EnvelopeFormatter', () => {
  describe('constructor', () => {
    it('should initialize with command name and version', () => {
      const formatter = new EnvelopeFormatter('test command', '1.0.0')
      expect(formatter).to.be.instanceOf(EnvelopeFormatter)
    })

    it('should record start time for execution tracking', (done) => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')

      setTimeout(() => {
        const envelope = formatter.wrapSuccess({})
        expect(envelope.metadata.execution_time_ms).to.be.gte(50)
        done()
      }, 100)
    })
  })

  describe('wrapSuccess', () => {
    it('should create success envelope with data', () => {
      const formatter = new EnvelopeFormatter('test command', '1.0.0')
      const testData = { id: 'abc-123', object: 'page' }
      const envelope = formatter.wrapSuccess(testData)

      expect(envelope.success).to.be.true
      expect(envelope.data).to.deep.equal(testData)
    })

    it('should include all required metadata fields', () => {
      const formatter = new EnvelopeFormatter('test command', '1.0.0')
      const envelope = formatter.wrapSuccess({})

      expect(envelope.metadata).to.have.property('timestamp')
      expect(envelope.metadata).to.have.property('command', 'test command')
      expect(envelope.metadata).to.have.property('execution_time_ms')
      expect(envelope.metadata).to.have.property('version', '1.0.0')
    })

    it('should track execution time accurately', (done) => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')

      setTimeout(() => {
        const envelope = formatter.wrapSuccess({})
        // Should be at least 50ms (accounting for setTimeout variance)
        expect(envelope.metadata.execution_time_ms).to.be.gte(50)
        // Should be less than 200ms (generous upper bound)
        expect(envelope.metadata.execution_time_ms).to.be.lte(200)
        done()
      }, 100)
    })

    it('should accept additional metadata', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const envelope = formatter.wrapSuccess({}, {
        page_size: 100,
        has_more: false,
        custom_field: 'custom_value',
      })

      expect(envelope.metadata).to.have.property('page_size', 100)
      expect(envelope.metadata).to.have.property('has_more', false)
      expect(envelope.metadata).to.have.property('custom_field', 'custom_value')
    })

    it('should handle null data', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const envelope = formatter.wrapSuccess(null)

      expect(envelope.success).to.be.true
      expect(envelope.data).to.be.null
    })

    it('should handle undefined data', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const envelope = formatter.wrapSuccess(undefined)

      expect(envelope.success).to.be.true
      expect(envelope.data).to.be.undefined
    })

    it('should handle large data objects', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const largeData = {
        results: Array(1000).fill({ id: 'test', properties: {} })
      }
      const envelope = formatter.wrapSuccess(largeData)

      expect(envelope.success).to.be.true
      expect(envelope.data.results).to.have.length(1000)
    })

    it('should preserve data type information', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const complexData = {
        string: 'test',
        number: 42,
        boolean: true,
        null: null,
        array: [1, 2, 3],
        object: { nested: 'value' },
      }
      const envelope = formatter.wrapSuccess(complexData)

      expect(envelope.data).to.deep.equal(complexData)
    })

    it('should create valid ISO 8601 timestamp', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const envelope = formatter.wrapSuccess({})

      // ISO 8601 format: YYYY-MM-DDTHH:mm:ss.sssZ
      const isoRegex = /^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{3}Z$/
      expect(envelope.metadata.timestamp).to.match(isoRegex)

      // Verify it's a valid date
      const date = new Date(envelope.metadata.timestamp)
      expect(date.toString()).to.not.equal('Invalid Date')
    })
  })

  describe('wrapError', () => {
    it('should wrap NotionCLIError correctly', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(
        ErrorCode.NOT_FOUND,
        'Resource not found',
        { resourceId: 'abc-123' }
      )
      const envelope = formatter.wrapError(error)

      expect(envelope.success).to.be.false
      expect(envelope.error.code).to.equal('NOT_FOUND')
      expect(envelope.error.message).to.equal('Resource not found')
      expect(envelope.error.details).to.have.property('resourceId', 'abc-123')
      expect(envelope.error.suggestions).to.be.an('array')
    })

    it('should wrap standard Error objects', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new Error('Something went wrong')
      const envelope = formatter.wrapError(error)

      expect(envelope.success).to.be.false
      expect(envelope.error.code).to.equal('UNKNOWN')
      expect(envelope.error.message).to.equal('Something went wrong')
      expect(envelope.error.details).to.have.property('stack')
    })

    it('should wrap raw error objects', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = {
        code: 'CUSTOM_ERROR',
        message: 'Custom error message',
        extra: 'data',
      }
      const envelope = formatter.wrapError(error)

      expect(envelope.success).to.be.false
      expect(envelope.error.code).to.equal('CUSTOM_ERROR')
      expect(envelope.error.message).to.equal('Custom error message')
      expect(envelope.error.details).to.have.property('extra', 'data')
    })

    it('should generate suggestions for UNAUTHORIZED', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.UNAUTHORIZED, 'Auth failed')
      const envelope = formatter.wrapError(error)

      expect(envelope.error.suggestions).to.be.an('array')
      expect(envelope.error.suggestions!.length).to.be.greaterThan(0)
      expect(envelope.error.suggestions!.join(' ')).to.include('NOTION_TOKEN')
    })

    it('should generate suggestions for NOT_FOUND', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.NOT_FOUND, 'Not found')
      const envelope = formatter.wrapError(error)

      const suggestionsText = envelope.error.suggestions!.join(' ')
      expect(suggestionsText).to.match(/resource ID/i)
      expect(suggestionsText).to.match(/notion-cli sync/i)
    })

    it('should generate suggestions for RATE_LIMITED', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.RATE_LIMITED, 'Rate limited')
      const envelope = formatter.wrapError(error)

      const suggestionsText = envelope.error.suggestions!.join(' ')
      expect(suggestionsText).to.match(/retry/i)
    })

    it('should generate suggestions for VALIDATION_ERROR', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.VALIDATION_ERROR, 'Validation failed')
      const envelope = formatter.wrapError(error)

      const suggestionsText = envelope.error.suggestions!.join(' ')
      expect(suggestionsText).to.match(/--help/i)
    })

    it('should handle errors without details', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.API_ERROR, 'API error')
      const envelope = formatter.wrapError(error)

      expect(envelope.success).to.be.false
      expect(envelope.error.details).to.exist
    })

    it('should include notionError if present', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const notionError = { status: 404, code: 'object_not_found' }
      const error = new NotionCLIError(
        ErrorCode.NOT_FOUND,
        'Not found',
        {},
        notionError
      )
      const envelope = formatter.wrapError(error)

      expect(envelope.error.notionError).to.deep.equal(notionError)
    })

    it('should add additional context when provided', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.API_ERROR, 'Error')
      const envelope = formatter.wrapError(error, {
        database_id: 'db-123',
        operation: 'query',
      })

      expect(envelope.error.details).to.have.property('database_id', 'db-123')
      expect(envelope.error.details).to.have.property('operation', 'query')
    })
  })

  describe('getExitCode', () => {
    it('should return 0 for success envelope', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const envelope = formatter.wrapSuccess({})

      expect(formatter.getExitCode(envelope)).to.equal(ExitCode.SUCCESS)
      expect(formatter.getExitCode(envelope)).to.equal(0)
    })

    it('should return 2 for VALIDATION_ERROR', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.VALIDATION_ERROR, 'Invalid')
      const envelope = formatter.wrapError(error)

      expect(formatter.getExitCode(envelope)).to.equal(ExitCode.CLI_ERROR)
      expect(formatter.getExitCode(envelope)).to.equal(2)
    })

    it('should return 1 for UNAUTHORIZED', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.UNAUTHORIZED, 'Unauthorized')
      const envelope = formatter.wrapError(error)

      expect(formatter.getExitCode(envelope)).to.equal(ExitCode.API_ERROR)
      expect(formatter.getExitCode(envelope)).to.equal(1)
    })

    it('should return 1 for NOT_FOUND', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.NOT_FOUND, 'Not found')
      const envelope = formatter.wrapError(error)

      expect(formatter.getExitCode(envelope)).to.equal(ExitCode.API_ERROR)
      expect(formatter.getExitCode(envelope)).to.equal(1)
    })

    it('should return 1 for RATE_LIMITED', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.RATE_LIMITED, 'Rate limited')
      const envelope = formatter.wrapError(error)

      expect(formatter.getExitCode(envelope)).to.equal(ExitCode.API_ERROR)
      expect(formatter.getExitCode(envelope)).to.equal(1)
    })

    it('should return 1 for API_ERROR', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.API_ERROR, 'API error')
      const envelope = formatter.wrapError(error)

      expect(formatter.getExitCode(envelope)).to.equal(ExitCode.API_ERROR)
      expect(formatter.getExitCode(envelope)).to.equal(1)
    })

    it('should return 1 for UNKNOWN', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.UNKNOWN, 'Unknown error')
      const envelope = formatter.wrapError(error)

      expect(formatter.getExitCode(envelope)).to.equal(ExitCode.API_ERROR)
      expect(formatter.getExitCode(envelope)).to.equal(1)
    })
  })

  describe('outputEnvelope', () => {
    it('should output pretty JSON by default', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const envelope = formatter.wrapSuccess({ test: 'data' })
      let output = ''

      formatter.outputEnvelope(envelope, {}, (msg) => { output = msg })

      const parsed = JSON.parse(output)
      expect(parsed).to.deep.equal(envelope)
      // Pretty JSON should have newlines
      expect(output).to.include('\n')
    })

    it('should output compact JSON with --compact-json', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const envelope = formatter.wrapSuccess({ test: 'data' })
      let output = ''

      formatter.outputEnvelope(envelope, { 'compact-json': true }, (msg) => { output = msg })

      const parsed = JSON.parse(output)
      expect(parsed).to.deep.equal(envelope)
      // Compact JSON should be single line (no newlines except maybe at end)
      expect(output.trim().split('\n').length).to.equal(1)
    })

    it('should output raw data with --raw flag', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const testData = { id: 'abc', object: 'page' }
      const envelope = formatter.wrapSuccess(testData)
      let output = ''

      formatter.outputEnvelope(envelope, { raw: true }, (msg) => { output = msg })

      const parsed = JSON.parse(output)
      // Raw mode: output data only, no envelope
      expect(parsed).to.deep.equal(testData)
      expect(parsed).to.not.have.property('success')
      expect(parsed).to.not.have.property('metadata')
    })

    it('should use provided log function', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const envelope = formatter.wrapSuccess({})
      let called = false
      let output = ''

      const customLog = (msg: string) => {
        called = true
        output = msg
      }

      formatter.outputEnvelope(envelope, {}, customLog)

      expect(called).to.be.true
      expect(output).to.be.a('string')
    })

    it('should not output raw data for error envelope with --raw', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(ErrorCode.NOT_FOUND, 'Not found')
      const envelope = formatter.wrapError(error)
      let output = ''

      // Raw flag should be ignored for error envelopes
      formatter.outputEnvelope(envelope, { raw: true }, (msg) => { output = msg })

      const parsed = JSON.parse(output)
      // Should still be full error envelope
      expect(parsed).to.have.property('success', false)
      expect(parsed).to.have.property('error')
      expect(parsed).to.have.property('metadata')
    })
  })

  describe('type guards', () => {
    describe('isSuccessEnvelope', () => {
      it('should return true for success envelope', () => {
        const formatter = new EnvelopeFormatter('test', '1.0.0')
        const envelope = formatter.wrapSuccess({})

        expect(isSuccessEnvelope(envelope)).to.be.true
      })

      it('should return false for error envelope', () => {
        const formatter = new EnvelopeFormatter('test', '1.0.0')
        const error = new NotionCLIError(ErrorCode.API_ERROR, 'Error')
        const envelope = formatter.wrapError(error)

        expect(isSuccessEnvelope(envelope)).to.be.false
      })
    })

    describe('isErrorEnvelope', () => {
      it('should return true for error envelope', () => {
        const formatter = new EnvelopeFormatter('test', '1.0.0')
        const error = new NotionCLIError(ErrorCode.API_ERROR, 'Error')
        const envelope = formatter.wrapError(error)

        expect(isErrorEnvelope(envelope)).to.be.true
      })

      it('should return false for success envelope', () => {
        const formatter = new EnvelopeFormatter('test', '1.0.0')
        const envelope = formatter.wrapSuccess({})

        expect(isErrorEnvelope(envelope)).to.be.false
      })
    })
  })

  describe('edge cases', () => {
    it('should handle very long execution times', (done) => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')

      setTimeout(() => {
        const envelope = formatter.wrapSuccess({})
        expect(envelope.metadata.execution_time_ms).to.be.gte(500)
        done()
      }, 500)
    })

    it('should handle rapid successive envelope creation', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')

      const envelope1 = formatter.wrapSuccess({ id: 1 })
      const envelope2 = formatter.wrapSuccess({ id: 2 })
      const envelope3 = formatter.wrapSuccess({ id: 3 })

      // All should have same base execution time (within a few ms)
      const times = [
        envelope1.metadata.execution_time_ms,
        envelope2.metadata.execution_time_ms,
        envelope3.metadata.execution_time_ms,
      ]

      // Times should be very close (within 10ms)
      const maxTime = Math.max(...times)
      const minTime = Math.min(...times)
      expect(maxTime - minTime).to.be.lte(10)
    })

    it('should handle special characters in error messages', () => {
      const formatter = new EnvelopeFormatter('test', '1.0.0')
      const error = new NotionCLIError(
        ErrorCode.API_ERROR,
        'Error with "quotes" and \'apostrophes\' and \n newlines'
      )
      const envelope = formatter.wrapError(error)

      expect(envelope.error.message).to.include('quotes')
      expect(envelope.error.message).to.include('apostrophes')
      expect(envelope.error.message).to.include('newlines')

      // Should be valid JSON
      expect(() => JSON.stringify(envelope)).to.not.throw()
    })
  })
})
