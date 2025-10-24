/**
 * Token Validation Utility
 *
 * Provides consistent token validation across all commands that interact with the Notion API.
 * This ensures users get helpful, actionable error messages before attempting API calls.
 */

import { NotionCLIErrorFactory } from '../errors'

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
