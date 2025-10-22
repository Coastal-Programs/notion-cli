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

// Export enhanced error system
export {
  // Error Class
  NotionCLIError,

  // Error Codes Enum
  NotionCLIErrorCode,

  // Factory Functions
  NotionCLIErrorFactory,

  // Utility Functions
  wrapNotionError,
  handleCliError,

  // Type Interfaces
  ErrorSuggestion,
  ErrorContext,
} from './enhanced-errors'

// Re-export legacy error system for backward compatibility
// TODO: Remove after migration is complete
export {
  ErrorCode as LegacyErrorCode,
  NotionCLIError as LegacyNotionCLIError,
  wrapNotionError as legacyWrapNotionError,
} from '../errors'
