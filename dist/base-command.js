"use strict";
/**
 * Base Command with Envelope Support
 *
 * Extends oclif Command with automatic envelope wrapping for consistent JSON output.
 * All commands should extend this class to get automatic envelope support.
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.EnvelopeFlags = exports.BaseCommand = void 0;
const core_1 = require("@oclif/core");
const envelope_1 = require("./envelope");
const index_1 = require("./errors/index");
const disk_cache_1 = require("./utils/disk-cache");
const http_agent_1 = require("./http-agent");
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
class BaseCommand extends core_1.Command {
    constructor() {
        super(...arguments);
        this.shouldUseEnvelope = false;
    }
    /**
     * Initialize command and create envelope formatter
     */
    async init() {
        var _a;
        await super.init();
        // Initialize disk cache (load from disk)
        const diskCacheEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED !== 'false';
        if (diskCacheEnabled) {
            try {
                await disk_cache_1.diskCacheManager.initialize();
            }
            catch (error) {
                // Silently ignore disk cache initialization errors
                if (process.env.DEBUG) {
                    console.error('Failed to initialize disk cache:', error);
                }
            }
        }
        // Get command name from ID (e.g., "page:retrieve" -> "page retrieve")
        const commandName = ((_a = this.id) === null || _a === void 0 ? void 0 : _a.replace(/:/g, ' ')) || 'unknown';
        // Get version from config
        const version = this.config.version;
        // Initialize envelope formatter
        this.envelope = new envelope_1.EnvelopeFormatter(commandName, version);
    }
    /**
     * Cleanup hook - flushes disk cache and destroys HTTP agents before exit
     */
    async finally(error) {
        // Destroy HTTP agents to close all connections
        try {
            (0, http_agent_1.destroyAgents)();
        }
        catch (agentError) {
            // Silently ignore agent cleanup errors
            if (process.env.DEBUG) {
                console.error('Failed to destroy HTTP agents:', agentError);
            }
        }
        // Flush disk cache before exit
        const diskCacheEnabled = process.env.NOTION_CLI_DISK_CACHE_ENABLED !== 'false';
        if (diskCacheEnabled) {
            try {
                await disk_cache_1.diskCacheManager.shutdown();
            }
            catch (shutdownError) {
                // Silently ignore shutdown errors
                if (process.env.DEBUG) {
                    console.error('Failed to shutdown disk cache:', shutdownError);
                }
            }
        }
        await super.finally(error);
    }
    /**
     * Determine if envelope should be used based on flags
     */
    checkEnvelopeUsage(flags) {
        return !!(flags.json || flags['compact-json']);
    }
    /**
     * Output success response with automatic envelope wrapping
     *
     * @param data - Response data
     * @param flags - Command flags
     * @param additionalMetadata - Optional metadata to include
     */
    outputSuccess(data, flags, additionalMetadata) {
        // Check if we should use envelope
        this.shouldUseEnvelope = this.checkEnvelopeUsage(flags);
        if (this.shouldUseEnvelope) {
            const envelope = this.envelope.wrapSuccess(data, additionalMetadata);
            this.envelope.outputEnvelope(envelope, flags, this.log.bind(this));
            process.exit(this.envelope.getExitCode(envelope));
        }
        else {
            // Non-envelope output (table, markdown, etc.) - handled by caller
            // This path should not normally be reached as caller handles non-JSON output
            throw new Error('outputSuccess should only be called for JSON output');
        }
    }
    /**
     * Output error response with automatic envelope wrapping
     *
     * @param error - Error object
     * @param flags - Command flags
     * @param additionalContext - Optional error context
     */
    outputError(error, flags, additionalContext) {
        // Wrap raw errors in NotionCLIError
        const cliError = error instanceof index_1.NotionCLIError ? error : (0, index_1.wrapNotionError)(error);
        // Check if we should use envelope
        this.shouldUseEnvelope = this.checkEnvelopeUsage(flags);
        if (this.shouldUseEnvelope) {
            const envelope = this.envelope.wrapError(cliError, additionalContext);
            this.envelope.outputEnvelope(envelope, flags, this.log.bind(this));
            process.exit(this.envelope.getExitCode(envelope));
        }
        else {
            // Non-JSON mode - use oclif's error handling
            this.error(cliError.message, { exit: this.getExitCodeForError(cliError) });
        }
    }
    /**
     * Get appropriate exit code for error
     */
    getExitCodeForError(error) {
        // CLI validation errors
        if (error.code === 'VALIDATION_ERROR') {
            return envelope_1.ExitCode.CLI_ERROR;
        }
        // API errors (default)
        return envelope_1.ExitCode.API_ERROR;
    }
    /**
     * Catch handler that ensures proper envelope error output
     */
    async catch(error) {
        // If command has already handled the error via outputError, just propagate
        if (error.exitCode !== undefined) {
            throw error;
        }
        // Otherwise, wrap and handle the error
        const cliError = (0, index_1.wrapNotionError)(error);
        this.error(cliError.message, { exit: this.getExitCodeForError(cliError) });
    }
}
exports.BaseCommand = BaseCommand;
/**
 * Standard flags that all envelope-enabled commands should include
 */
exports.EnvelopeFlags = {
    json: core_1.Flags.boolean({
        char: 'j',
        description: 'Output as JSON envelope (recommended for automation)',
        default: false,
    }),
    'compact-json': core_1.Flags.boolean({
        char: 'c',
        description: 'Output as compact JSON envelope (single-line, ideal for piping)',
        default: false,
        exclusive: ['markdown', 'pretty'],
    }),
    raw: core_1.Flags.boolean({
        char: 'r',
        description: 'Output raw API response without envelope (legacy mode)',
        default: false,
    }),
};
