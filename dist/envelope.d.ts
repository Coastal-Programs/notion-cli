/**
 * JSON Envelope Standardization System for Notion CLI
 *
 * Provides consistent machine-readable output across all commands with:
 * - Standard success/error envelopes
 * - Metadata tracking (command, timestamp, execution time)
 * - Exit code standardization (0=success, 1=API error, 2=CLI error)
 * - Proper stdout/stderr separation
 */
import { ErrorCode } from './errors';
/**
 * Standard metadata included in all envelopes
 */
export interface EnvelopeMetadata {
    /** ISO 8601 timestamp when the command was executed */
    timestamp: string;
    /** Full command name (e.g., "page retrieve", "db query") */
    command: string;
    /** Execution time in milliseconds */
    execution_time_ms: number;
    /** CLI version for debugging and compatibility */
    version: string;
}
/**
 * Success envelope structure
 * Used when a command completes successfully
 */
export interface SuccessEnvelope<T = any> {
    success: true;
    data: T;
    metadata: EnvelopeMetadata;
}
/**
 * Error details structure
 */
export interface ErrorDetails {
    /** Semantic error code (e.g., "DATABASE_NOT_FOUND", "RATE_LIMITED") */
    code: ErrorCode | string;
    /** Human-readable error message */
    message: string;
    /** Additional context about the error */
    details?: any;
    /** Actionable suggestions for the user */
    suggestions?: string[];
    /** Original Notion API error (if applicable) */
    notionError?: any;
}
/**
 * Error envelope structure
 * Used when a command fails
 */
export interface ErrorEnvelope {
    success: false;
    error: ErrorDetails;
    metadata: Omit<EnvelopeMetadata, 'execution_time_ms'> & {
        execution_time_ms?: number;
    };
}
/**
 * Union type for all envelope responses
 */
export type Envelope<T = any> = SuccessEnvelope<T> | ErrorEnvelope;
/**
 * Exit codes for consistent process termination
 */
export declare enum ExitCode {
    /** Command completed successfully */
    SUCCESS = 0,
    /** API/Notion error (auth, not found, rate limit, network, etc.) */
    API_ERROR = 1,
    /** CLI/validation error (invalid args, syntax, config issues) */
    CLI_ERROR = 2
}
/**
 * Output flags that determine envelope formatting
 */
export interface OutputFlags {
    json?: boolean;
    'compact-json'?: boolean;
    raw?: boolean;
    markdown?: boolean;
    pretty?: boolean;
    csv?: boolean;
}
/**
 * EnvelopeFormatter - Core utility for creating and outputting envelopes
 */
export declare class EnvelopeFormatter {
    private startTime;
    private commandName;
    private version;
    /**
     * Initialize formatter with command metadata
     *
     * @param commandName - Full command name (e.g., "page retrieve")
     * @param version - CLI version from package.json
     */
    constructor(commandName: string, version: string);
    /**
     * Create success envelope with data and metadata
     *
     * @param data - The actual response data
     * @param additionalMetadata - Optional additional metadata fields
     * @returns Success envelope ready for output
     */
    wrapSuccess<T>(data: T, additionalMetadata?: Record<string, any>): SuccessEnvelope<T>;
    /**
     * Create error envelope from Error, NotionCLIError, or raw error object
     *
     * @param error - Error instance or error object
     * @param additionalContext - Optional additional error context
     * @returns Error envelope ready for output
     */
    wrapError(error: any, additionalContext?: Record<string, any>): ErrorEnvelope;
    /**
     * Output envelope to stdout with proper formatting
     * Handles flag-based format selection and stdout/stderr separation
     *
     * @param envelope - Success or error envelope
     * @param flags - Output format flags
     * @param logFn - Logging function (typically this.log from Command)
     */
    outputEnvelope(envelope: Envelope, flags: OutputFlags, logFn?: (message: string) => void): void;
    /**
     * Get appropriate exit code for the envelope
     *
     * @param envelope - Success or error envelope
     * @returns Exit code (0, 1, or 2)
     */
    getExitCode(envelope: Envelope): ExitCode;
    /**
     * Write diagnostic messages to stderr (won't pollute JSON on stdout)
     * Useful for retry messages, cache hits, debug info, etc.
     *
     * @param message - Diagnostic message
     * @param level - Message level (info, warn, error)
     */
    static writeDiagnostic(message: string, level?: 'info' | 'warn' | 'error'): void;
    /**
     * Helper to log retry attempts to stderr (doesn't pollute JSON output)
     *
     * @param attempt - Retry attempt number
     * @param maxRetries - Maximum retry attempts
     * @param delay - Delay before next retry in milliseconds
     */
    static logRetry(attempt: number, maxRetries: number, delay: number): void;
    /**
     * Helper to log cache hits to stderr (for debugging)
     *
     * @param cacheKey - Cache key that was hit
     */
    static logCacheHit(cacheKey: string): void;
}
/**
 * Convenience function to create an envelope formatter
 *
 * @param commandName - Full command name
 * @param version - CLI version
 * @returns New EnvelopeFormatter instance
 */
export declare function createEnvelopeFormatter(commandName: string, version: string): EnvelopeFormatter;
/**
 * Type guard to check if envelope is a success envelope
 */
export declare function isSuccessEnvelope<T>(envelope: Envelope<T>): envelope is SuccessEnvelope<T>;
/**
 * Type guard to check if envelope is an error envelope
 */
export declare function isErrorEnvelope(envelope: Envelope): envelope is ErrorEnvelope;
