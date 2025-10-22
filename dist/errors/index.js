"use strict";
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
Object.defineProperty(exports, "__esModule", { value: true });
exports.legacyWrapNotionError = exports.LegacyNotionCLIError = exports.LegacyErrorCode = exports.handleCliError = exports.wrapNotionError = exports.NotionCLIErrorFactory = exports.NotionCLIErrorCode = exports.NotionCLIError = void 0;
// Export enhanced error system
var enhanced_errors_1 = require("./enhanced-errors");
// Error Class
Object.defineProperty(exports, "NotionCLIError", { enumerable: true, get: function () { return enhanced_errors_1.NotionCLIError; } });
// Error Codes Enum
Object.defineProperty(exports, "NotionCLIErrorCode", { enumerable: true, get: function () { return enhanced_errors_1.NotionCLIErrorCode; } });
// Factory Functions
Object.defineProperty(exports, "NotionCLIErrorFactory", { enumerable: true, get: function () { return enhanced_errors_1.NotionCLIErrorFactory; } });
// Utility Functions
Object.defineProperty(exports, "wrapNotionError", { enumerable: true, get: function () { return enhanced_errors_1.wrapNotionError; } });
Object.defineProperty(exports, "handleCliError", { enumerable: true, get: function () { return enhanced_errors_1.handleCliError; } });
// Re-export legacy error system for backward compatibility
// TODO: Remove after migration is complete
var errors_1 = require("../errors");
Object.defineProperty(exports, "LegacyErrorCode", { enumerable: true, get: function () { return errors_1.ErrorCode; } });
Object.defineProperty(exports, "LegacyNotionCLIError", { enumerable: true, get: function () { return errors_1.NotionCLIError; } });
Object.defineProperty(exports, "legacyWrapNotionError", { enumerable: true, get: function () { return errors_1.wrapNotionError; } });
