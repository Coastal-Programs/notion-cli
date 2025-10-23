/**
 * @deprecated This file is deprecated. Use the enhanced error system instead:
 * import { NotionCLIError, NotionCLIErrorCode, NotionCLIErrorFactory, wrapNotionError } from './errors/index'
 *
 * The new error system provides:
 * - More detailed error codes
 * - Context-rich error messages
 * - Actionable suggestions with commands
 * - Better AI-friendly formatting
 */

export enum ErrorCode {
  RATE_LIMITED = 'RATE_LIMITED',
  NOT_FOUND = 'NOT_FOUND',
  UNAUTHORIZED = 'UNAUTHORIZED',
  VALIDATION_ERROR = 'VALIDATION_ERROR',
  API_ERROR = 'API_ERROR',
  UNKNOWN = 'UNKNOWN',
}

export class NotionCLIError extends Error {
  constructor(
    public code: ErrorCode,
    message: string,
    public details?: any,
    public notionError?: any
  ) {
    super(message)
    this.name = 'NotionCLIError'
  }

  toJSON() {
    return {
      success: false,
      error: {
        code: this.code,
        message: this.message,
        details: this.details,
        notionError: this.notionError
      },
      timestamp: new Date().toISOString()
    }
  }
}

export const wrapNotionError = (error: any): NotionCLIError => {
  // Map HTTP status codes to error codes
  if (error.status === 429) {
    return new NotionCLIError(
      ErrorCode.RATE_LIMITED,
      'Rate limit exceeded',
      { retryAfter: error.headers?.['retry-after'] },
      error
    )
  }

  if (error.status === 401 || error.status === 403) {
    const isTokenMissing = !process.env.NOTION_TOKEN
    const message = isTokenMissing
      ? 'NOTION_TOKEN environment variable is not set.\n\nTo fix:\n  export NOTION_TOKEN="your-token-here"  # Mac/Linux\n  set NOTION_TOKEN=your-token-here       # Windows CMD\n  $env:NOTION_TOKEN="your-token-here"    # Windows PowerShell\n\nGet your token at: https://developers.notion.com/docs/create-a-notion-integration'
      : 'Authentication failed. Your NOTION_TOKEN may be invalid or expired.\n\nVerify your token at: https://www.notion.so/my-integrations'

    return new NotionCLIError(
      ErrorCode.UNAUTHORIZED,
      message,
      { tokenSet: !isTokenMissing },
      error
    )
  }

  if (error.status === 404) {
    return new NotionCLIError(
      ErrorCode.NOT_FOUND,
      'Resource not found',
      null,
      error
    )
  }

  if (error.status === 400) {
    return new NotionCLIError(
      ErrorCode.VALIDATION_ERROR,
      'Invalid request parameters',
      null,
      error
    )
  }

  // Default API error
  return new NotionCLIError(
    ErrorCode.API_ERROR,
    error.message || 'Notion API error',
    null,
    error
  )
}
