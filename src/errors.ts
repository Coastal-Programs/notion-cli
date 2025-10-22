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
    return new NotionCLIError(
      ErrorCode.UNAUTHORIZED,
      'Authentication failed. Check your NOTION_TOKEN.',
      null,
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
