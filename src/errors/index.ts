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

// Export enhanced error system (primary exports)
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
