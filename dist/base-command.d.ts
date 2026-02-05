/**
 * Base Command with Envelope Support
 *
 * Extends oclif Command with automatic envelope wrapping for consistent JSON output.
 * All commands should extend this class to get automatic envelope support.
 */
import { Command, Interfaces } from '@oclif/core';
import { EnvelopeFormatter, OutputFlags } from './envelope';
import { NotionCLIError } from './errors/index';
/**
 * Base command configuration
 */
export type CommandConfig = Interfaces.Config;
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
export declare abstract class BaseCommand extends Command {
    protected envelope: EnvelopeFormatter;
    protected shouldUseEnvelope: boolean;
    /**
     * Initialize command and create envelope formatter
     */
    init(): Promise<void>;
    /**
     * Cleanup hook - flushes disk cache and destroys HTTP agents before exit
     */
    finally(error?: Error): Promise<void>;
    /**
     * Determine if envelope should be used based on flags
     */
    protected checkEnvelopeUsage(flags: Record<string, unknown>): boolean;
    /**
     * Output success response with automatic envelope wrapping
     *
     * @param data - Response data
     * @param flags - Command flags
     * @param additionalMetadata - Optional metadata to include
     */
    protected outputSuccess<T>(data: T, flags: OutputFlags & Record<string, unknown>, additionalMetadata?: Record<string, unknown>): never;
    /**
     * Output error response with automatic envelope wrapping
     *
     * @param error - Error object
     * @param flags - Command flags
     * @param additionalContext - Optional error context
     */
    protected outputError(error: Error | NotionCLIError, flags: OutputFlags & Record<string, unknown>, additionalContext?: Record<string, unknown>): never;
    /**
     * Get appropriate exit code for error
     */
    private getExitCodeForError;
    /**
     * Catch handler that ensures proper envelope error output
     */
    catch(error: Error & {
        exitCode?: number;
    }): Promise<void>;
}
/**
 * Standard flags that all envelope-enabled commands should include
 */
export declare const EnvelopeFlags: {
    json: Interfaces.BooleanFlag<boolean>;
    'compact-json': Interfaces.BooleanFlag<boolean>;
    raw: Interfaces.BooleanFlag<boolean>;
};
