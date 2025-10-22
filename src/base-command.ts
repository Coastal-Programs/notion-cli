/**
 * Base Command with Envelope Support
 *
 * Extends oclif Command with automatic envelope wrapping for consistent JSON output.
 * All commands should extend this class to get automatic envelope support.
 */

import { Command, Flags, Interfaces } from '@oclif/core'
import { EnvelopeFormatter, ExitCode, OutputFlags } from './envelope'
import { wrapNotionError, NotionCLIError } from './errors'

/**
 * Base command configuration
 */
export type CommandConfig = Interfaces.Config

/**
 * BaseCommand - Extends oclif Command with envelope support
 *
 * Features:
 * - Automatic envelope wrapping for JSON output
 * - Consistent error handling
 * - Execution time tracking
 * - Version metadata injection
 * - Stdout/stderr separation
 */
export abstract class BaseCommand extends Command {
  protected envelope!: EnvelopeFormatter
  protected shouldUseEnvelope = false

  /**
   * Initialize command and create envelope formatter
   */
  async init(): Promise<void> {
    await super.init()

    // Get command name from ID (e.g., "page:retrieve" -> "page retrieve")
    const commandName = this.id?.replace(/:/g, ' ') || 'unknown'

    // Get version from config
    const version = this.config.version

    // Initialize envelope formatter
    this.envelope = new EnvelopeFormatter(commandName, version)
  }

  /**
   * Determine if envelope should be used based on flags
   */
  protected checkEnvelopeUsage(flags: any): boolean {
    return !!(flags.json || flags['compact-json'])
  }

  /**
   * Output success response with automatic envelope wrapping
   *
   * @param data - Response data
   * @param flags - Command flags
   * @param additionalMetadata - Optional metadata to include
   */
  protected outputSuccess<T>(
    data: T,
    flags: OutputFlags & any,
    additionalMetadata?: Record<string, any>
  ): never {
    // Check if we should use envelope
    this.shouldUseEnvelope = this.checkEnvelopeUsage(flags)

    if (this.shouldUseEnvelope) {
      const envelope = this.envelope.wrapSuccess(data, additionalMetadata)
      this.envelope.outputEnvelope(envelope, flags, this.log.bind(this))
      process.exit(this.envelope.getExitCode(envelope))
    } else {
      // Non-envelope output (table, markdown, etc.) - handled by caller
      // This path should not normally be reached as caller handles non-JSON output
      throw new Error('outputSuccess should only be called for JSON output')
    }
  }

  /**
   * Output error response with automatic envelope wrapping
   *
   * @param error - Error object
   * @param flags - Command flags
   * @param additionalContext - Optional error context
   */
  protected outputError(
    error: any,
    flags: OutputFlags & any,
    additionalContext?: Record<string, any>
  ): never {
    // Wrap raw errors in NotionCLIError
    const cliError = error instanceof NotionCLIError ? error : wrapNotionError(error)

    // Check if we should use envelope
    this.shouldUseEnvelope = this.checkEnvelopeUsage(flags)

    if (this.shouldUseEnvelope) {
      const envelope = this.envelope.wrapError(cliError, additionalContext)
      this.envelope.outputEnvelope(envelope, flags, this.log.bind(this))
      process.exit(this.envelope.getExitCode(envelope))
    } else {
      // Non-JSON mode - use oclif's error handling
      this.error(cliError.message, { exit: this.getExitCodeForError(cliError) })
    }
  }

  /**
   * Get appropriate exit code for error
   */
  private getExitCodeForError(error: NotionCLIError): number {
    // CLI validation errors
    if (error.code === 'VALIDATION_ERROR') {
      return ExitCode.CLI_ERROR
    }

    // API errors (default)
    return ExitCode.API_ERROR
  }

  /**
   * Catch handler that ensures proper envelope error output
   */
  async catch(error: Error & { exitCode?: number }): Promise<any> {
    // If command has already handled the error via outputError, just propagate
    if (error.exitCode !== undefined) {
      throw error
    }

    // Otherwise, wrap and handle the error
    const cliError = wrapNotionError(error)
    this.error(cliError.message, { exit: this.getExitCodeForError(cliError) })
  }
}

/**
 * Standard flags that all envelope-enabled commands should include
 */
export const EnvelopeFlags = {
  json: Flags.boolean({
    char: 'j',
    description: 'Output as JSON envelope (recommended for automation)',
    default: false,
  }),
  'compact-json': Flags.boolean({
    char: 'c',
    description: 'Output as compact JSON envelope (single-line, ideal for piping)',
    default: false,
    exclusive: ['markdown', 'pretty'],
  }),
  raw: Flags.boolean({
    char: 'r',
    description: 'Output raw API response without envelope (legacy mode)',
    default: false,
  }),
}
