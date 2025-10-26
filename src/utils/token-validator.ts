/**
 * Token Validation Utility
 *
 * Provides consistent token validation across all commands that interact with the Notion API.
 * This ensures users get helpful, actionable error messages before attempting API calls.
 */

import { NotionCLIErrorFactory } from '../errors'

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
export function maskToken(token: string): string {
  if (!token) return ''

  if (token.length <= 10) {
    // Token too short to safely mask, obscure completely
    // Threshold: 10 chars ensures at least 3 chars are masked after prefix+suffix
    return '***'
  }

  // Show prefix (secret_ or ntn_) and last 3 chars
  // For unknown prefixes: use max 4 chars to ensure at least 4 chars are masked
  const prefix = token.startsWith('secret_') ? 'secret_' :
                 token.startsWith('ntn_') ? 'ntn_' :
                 token.slice(0, Math.min(4, token.length - 7))
  const suffix = token.slice(-3)

  return `${prefix}***...***${suffix}`
}

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
export function validateNotionToken(): void {
  if (!process.env.NOTION_TOKEN) {
    throw NotionCLIErrorFactory.tokenMissing()
  }
}
