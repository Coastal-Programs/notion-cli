/**
 * Token Validation Utility
 *
 * Provides consistent token validation across all commands that interact with the Notion API.
 * This ensures users get helpful, actionable error messages before attempting API calls.
 */
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
export declare function maskToken(token: string): string;
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
export declare function validateNotionToken(): void;
