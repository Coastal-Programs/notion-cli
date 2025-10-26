"use strict";
/**
 * Token Validation Utility
 *
 * Provides consistent token validation across all commands that interact with the Notion API.
 * This ensures users get helpful, actionable error messages before attempting API calls.
 */
Object.defineProperty(exports, "__esModule", { value: true });
exports.validateNotionToken = exports.maskToken = void 0;
const errors_1 = require("../errors");
/**
 * Masks a Notion token for safe display in logs and console output
 *
 * Shows only the prefix and last 3 characters to prevent token leakage
 * in screen recordings, terminal sharing, or logs.
 *
 * @param token - The token to mask
 * @returns Masked token string (e.g., "secret_***...***abc")
 *
 * @example
 * ```typescript
 * const token = "secret_1234567890abcdef"
 * const masked = maskToken(token)
 * // Returns: "secret_***...***def"
 * ```
 */
function maskToken(token) {
    if (!token)
        return '';
    if (token.length <= 10) {
        // Token too short to safely mask, obscure completely
        return '***';
    }
    // Show prefix (secret_ or ntn_) and last 3 chars
    const prefix = token.startsWith('secret_') ? 'secret_' :
        token.startsWith('ntn_') ? 'ntn_' :
            token.slice(0, 7);
    const suffix = token.slice(-3);
    return `${prefix}***...***${suffix}`;
}
exports.maskToken = maskToken;
/**
 * Validates that NOTION_TOKEN environment variable is set
 *
 * @throws {NotionCLIError} If token is not set, throws with helpful suggestions
 *
 * @example
 * ```typescript
 * import { validateNotionToken } from '../utils/token-validator'
 *
 * // In your command's run() method:
 * async run() {
 *   const { flags } = await this.parse(MyCommand)
 *   validateNotionToken() // Throws if token not set
 *
 *   // Continue with API calls...
 * }
 * ```
 */
function validateNotionToken() {
    if (!process.env.NOTION_TOKEN) {
        throw errors_1.NotionCLIErrorFactory.tokenMissing();
    }
}
exports.validateNotionToken = validateNotionToken;
