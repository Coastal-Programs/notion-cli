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

// Note: Legacy error system is in src/errors.ts
// Commands should import from this file (src/errors/index.ts) to get enhanced errors
