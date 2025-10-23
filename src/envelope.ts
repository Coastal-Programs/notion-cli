/**
 * JSON Envelope Standardization System for Notion CLI
 *
 * Provides consistent machine-readable output across all commands with:
 * - Standard success/error envelopes
 * - Metadata tracking (command, timestamp, execution time)
 * - Exit code standardization (0=success, 1=API error, 2=CLI error)
 * - Proper stdout/stderr separation
 */

import { NotionCLIErrorCode, NotionCLIError } from './errors/index'

/**
 * Standard metadata included in all envelopes
 */
export interface EnvelopeMetadata {
  /** ISO 8601 timestamp when the command was executed */
  timestamp: string
  /** Full command name (e.g., "page retrieve", "db query") */
  command: string
  /** Execution time in milliseconds */
  execution_time_ms: number
  /** CLI version for debugging and compatibility */
  version: string
}

/**
 * Success envelope structure
 * Used when a command completes successfully
 */
export interface SuccessEnvelope<T = any> {
  success: true
  data: T
  metadata: EnvelopeMetadata
}

/**
 * Error details structure
 */
export interface ErrorDetails {
  /** Semantic error code (e.g., "DATABASE_NOT_FOUND", "RATE_LIMITED") */
  code: NotionCLIErrorCode | string
  /** Human-readable error message */
  message: string
  /** Additional context about the error */
  details?: any
  /** Actionable suggestions for the user */
  suggestions?: string[]
  /** Original Notion API error (if applicable) */
  notionError?: any
}

/**
 * Error envelope structure
 * Used when a command fails
 */
export interface ErrorEnvelope {
  success: false
  error: ErrorDetails
  metadata: Omit<EnvelopeMetadata, 'execution_time_ms'> & { execution_time_ms?: number }
}

/**
 * Union type for all envelope responses
 */
export type Envelope<T = any> = SuccessEnvelope<T> | ErrorEnvelope

/**
 * Exit codes for consistent process termination
 */
export enum ExitCode {
  /** Command completed successfully */
  SUCCESS = 0,
  /** API/Notion error (auth, not found, rate limit, network, etc.) */
  API_ERROR = 1,
  /** CLI/validation error (invalid args, syntax, config issues) */
  CLI_ERROR = 2,
}

/**
 * Output flags that determine envelope formatting
 */
export interface OutputFlags {
  json?: boolean
  'compact-json'?: boolean
  raw?: boolean
  markdown?: boolean
  pretty?: boolean
  csv?: boolean
}

/**
 * Maps error codes to appropriate exit codes
 */
function getExitCodeForError(errorCode: NotionCLIErrorCode | string): ExitCode {
  // CLI/validation errors
  const cliErrors = [
    NotionCLIErrorCode.VALIDATION_ERROR,
    'VALIDATION_ERROR',
    'CLI_ERROR',
    'CONFIG_ERROR',
    'INVALID_ARGUMENT',
  ]

  if (cliErrors.includes(errorCode as any)) {
    return ExitCode.CLI_ERROR
  }

  // All other errors are API-related
  return ExitCode.API_ERROR
}

/**
 * Suggestion generator based on error codes
 */
function generateSuggestions(errorCode: NotionCLIErrorCode | string): string[] {
  const suggestions: string[] = []

  switch (errorCode) {
    case NotionCLIErrorCode.UNAUTHORIZED:
      suggestions.push('Verify your NOTION_TOKEN is set correctly')
      suggestions.push('Check token at: https://www.notion.so/my-integrations')
      break
    case NotionCLIErrorCode.NOT_FOUND:
      suggestions.push('Verify the resource ID is correct')
      suggestions.push('Ensure your integration has access to the resource')
      suggestions.push('Try running: notion-cli sync')
      break
    case NotionCLIErrorCode.RATE_LIMITED:
      suggestions.push('Wait and retry - the CLI will auto-retry with backoff')
      suggestions.push('Reduce request frequency if this persists')
      break
    case NotionCLIErrorCode.VALIDATION_ERROR:
      suggestions.push('Check command syntax: notion-cli [command] --help')
      suggestions.push('Verify all required arguments are provided')
      break
    case 'CLI_ERROR':
    case 'CONFIG_ERROR':
      suggestions.push('Run: notion-cli config set-token')
      suggestions.push('Check your .env file configuration')
      break
  }

  return suggestions
}

/**
 * EnvelopeFormatter - Core utility for creating and outputting envelopes
 */
export class EnvelopeFormatter {
  private startTime: number
  private commandName: string
  private version: string

  /**
   * Initialize formatter with command metadata
   *
   * @param commandName - Full command name (e.g., "page retrieve")
   * @param version - CLI version from package.json
   */
  constructor(commandName: string, version: string) {
    this.startTime = Date.now()
    this.commandName = commandName
    this.version = version
  }

  /**
   * Create success envelope with data and metadata
   *
   * @param data - The actual response data
   * @param additionalMetadata - Optional additional metadata fields
   * @returns Success envelope ready for output
   */
  wrapSuccess<T>(data: T, additionalMetadata?: Record<string, any>): SuccessEnvelope<T> {
    const executionTime = Date.now() - this.startTime

    return {
      success: true,
      data,
      metadata: {
        timestamp: new Date().toISOString(),
        command: this.commandName,
        execution_time_ms: executionTime,
        version: this.version,
        ...additionalMetadata,
      },
    }
  }

  /**
   * Create error envelope from Error, NotionCLIError, or raw error object
   *
   * @param error - Error instance or error object
   * @param additionalContext - Optional additional error context
   * @returns Error envelope ready for output
   */
  wrapError(error: any, additionalContext?: Record<string, any>): ErrorEnvelope {
    const executionTime = Date.now() - this.startTime

    let errorDetails: ErrorDetails

    // Handle NotionCLIError
    if (error instanceof NotionCLIError) {
      errorDetails = {
        code: error.code,
        message: error.message,
        details: { ...error.context, ...additionalContext },
        suggestions: error.suggestions.map(s => s.description),
        notionError: error.context.originalError,
      }
    }
    // Handle standard Error
    else if (error instanceof Error) {
      errorDetails = {
        code: 'UNKNOWN',
        message: error.message,
        details: { stack: error.stack, ...additionalContext },
        suggestions: ['Check the error message for details'],
      }
    }
    // Handle raw error objects
    else {
      errorDetails = {
        code: error.code || 'UNKNOWN',
        message: error.message || 'An unknown error occurred',
        details: { ...error, ...additionalContext },
        suggestions: generateSuggestions(error.code || 'UNKNOWN'),
      }
    }

    return {
      success: false,
      error: errorDetails,
      metadata: {
        timestamp: new Date().toISOString(),
        command: this.commandName,
        execution_time_ms: executionTime,
        version: this.version,
      },
    }
  }

  /**
   * Output envelope to stdout with proper formatting
   * Handles flag-based format selection and stdout/stderr separation
   *
   * @param envelope - Success or error envelope
   * @param flags - Output format flags
   * @param logFn - Logging function (typically this.log from Command)
   */
  outputEnvelope(
    envelope: Envelope,
    flags: OutputFlags,
    logFn: (message: string) => void = console.log
  ): void {
    // Raw mode bypasses envelope - outputs data directly
    if (flags.raw && envelope.success) {
      logFn(JSON.stringify(envelope.data, null, 2))
      return
    }

    // Compact JSON - single line for piping
    if (flags['compact-json']) {
      logFn(JSON.stringify(envelope))
      return
    }

    // Default: Pretty JSON (--json flag or error state)
    logFn(JSON.stringify(envelope, null, 2))
  }

  /**
   * Get appropriate exit code for the envelope
   *
   * @param envelope - Success or error envelope
   * @returns Exit code (0, 1, or 2)
   */
  getExitCode(envelope: Envelope): ExitCode {
    if (envelope.success) {
      return ExitCode.SUCCESS
    }

    // Type narrowing: at this point, envelope is ErrorEnvelope
    return getExitCodeForError((envelope as ErrorEnvelope).error.code)
  }

  /**
   * Write diagnostic messages to stderr (won't pollute JSON on stdout)
   * Useful for retry messages, cache hits, debug info, etc.
   *
   * @param message - Diagnostic message
   * @param level - Message level (info, warn, error)
   */
  static writeDiagnostic(message: string, level: 'info' | 'warn' | 'error' = 'info'): void {
    const prefix = {
      info: '[INFO]',
      warn: '[WARN]',
      error: '[ERROR]',
    }[level]

    // Write to stderr to avoid polluting JSON output on stdout
    console.error(`${prefix} ${message}`)
  }

  /**
   * Helper to log retry attempts to stderr (doesn't pollute JSON output)
   *
   * @param attempt - Retry attempt number
   * @param maxRetries - Maximum retry attempts
   * @param delay - Delay before next retry in milliseconds
   */
  static logRetry(attempt: number, maxRetries: number, delay: number): void {
    EnvelopeFormatter.writeDiagnostic(
      `Retry attempt ${attempt}/${maxRetries} after ${delay}ms`,
      'warn'
    )
  }

  /**
   * Helper to log cache hits to stderr (for debugging)
   *
   * @param cacheKey - Cache key that was hit
   */
  static logCacheHit(cacheKey: string): void {
    if (process.env.DEBUG === 'true') {
      EnvelopeFormatter.writeDiagnostic(`Cache hit: ${cacheKey}`, 'info')
    }
  }
}

/**
 * Convenience function to create an envelope formatter
 *
 * @param commandName - Full command name
 * @param version - CLI version
 * @returns New EnvelopeFormatter instance
 */
export function createEnvelopeFormatter(commandName: string, version: string): EnvelopeFormatter {
  return new EnvelopeFormatter(commandName, version)
}

/**
 * Type guard to check if envelope is a success envelope
 */
export function isSuccessEnvelope<T>(envelope: Envelope<T>): envelope is SuccessEnvelope<T> {
  return envelope.success === true
}

/**
 * Type guard to check if envelope is an error envelope
 */
export function isErrorEnvelope(envelope: Envelope): envelope is ErrorEnvelope {
  return envelope.success === false
}
