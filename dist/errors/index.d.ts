/**
 * Error Handling System - Clean Exports
 *
 * Central import point for all error-related functionality.
 * Use this for clean imports in command files:
 *
 * @example
 * ```typescript
 * import {
 *   NotionCLIError,
 *   NotionCLIErrorCode,
 *   NotionCLIErrorFactory,
 *   handleCliError,
 *   wrapNotionError
 * } from '../errors'
 * ```
 */
export { NotionCLIError, NotionCLIErrorCode, NotionCLIErrorFactory, wrapNotionError, handleCliError, ErrorSuggestion, ErrorContext, } from './enhanced-errors';
export { ErrorCode as LegacyErrorCode, NotionCLIError as LegacyNotionCLIError, wrapNotionError as legacyWrapNotionError, } from '../errors';
