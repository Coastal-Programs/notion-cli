/**
 * Enhanced AI-Friendly Error Handling System
 *
 * Provides context-rich errors with actionable suggestions for:
 * - AI assistants debugging automation failures
 * - Human users troubleshooting CLI issues
 * - Automated systems logging meaningful errors
 *
 * Key Features:
 * - Error codes for programmatic handling
 * - Contextual suggestions with fix commands
 * - Support for both human and JSON output
 * - Notion API error mapping
 * - Common scenario detection
 */
/**
 * Comprehensive error codes covering all common scenarios
 */
export declare enum NotionCLIErrorCode {
    UNAUTHORIZED = "UNAUTHORIZED",
    TOKEN_MISSING = "TOKEN_MISSING",
    TOKEN_INVALID = "TOKEN_INVALID",
    TOKEN_EXPIRED = "TOKEN_EXPIRED",
    PERMISSION_DENIED = "PERMISSION_DENIED",
    INTEGRATION_NOT_SHARED = "INTEGRATION_NOT_SHARED",
    NOT_FOUND = "NOT_FOUND",
    OBJECT_NOT_FOUND = "OBJECT_NOT_FOUND",
    DATABASE_NOT_FOUND = "DATABASE_NOT_FOUND",
    PAGE_NOT_FOUND = "PAGE_NOT_FOUND",
    BLOCK_NOT_FOUND = "BLOCK_NOT_FOUND",
    INVALID_ID_FORMAT = "INVALID_ID_FORMAT",
    INVALID_DATABASE_ID = "INVALID_DATABASE_ID",
    INVALID_PAGE_ID = "INVALID_PAGE_ID",
    INVALID_BLOCK_ID = "INVALID_BLOCK_ID",
    INVALID_URL = "INVALID_URL",
    DATABASE_ID_CONFUSION = "DATABASE_ID_CONFUSION",
    WORKSPACE_VS_DATABASE = "WORKSPACE_VS_DATABASE",
    RATE_LIMITED = "RATE_LIMITED",
    API_ERROR = "API_ERROR",
    NETWORK_ERROR = "NETWORK_ERROR",
    TIMEOUT = "TIMEOUT",
    SERVICE_UNAVAILABLE = "SERVICE_UNAVAILABLE",
    VALIDATION_ERROR = "VALIDATION_ERROR",
    INVALID_PROPERTY = "INVALID_PROPERTY",
    INVALID_FILTER = "INVALID_FILTER",
    INVALID_JSON = "INVALID_JSON",
    MISSING_REQUIRED_FIELD = "MISSING_REQUIRED_FIELD",
    CACHE_ERROR = "CACHE_ERROR",
    WORKSPACE_NOT_SYNCED = "WORKSPACE_NOT_SYNCED",
    UNKNOWN = "UNKNOWN",
    INTERNAL_ERROR = "INTERNAL_ERROR"
}
/**
 * Suggested fix with command example
 */
export interface ErrorSuggestion {
    description: string;
    command?: string;
    link?: string;
}
/**
 * Contextual error information for better debugging
 */
export interface ErrorContext {
    /** The resource type being accessed */
    resourceType?: 'database' | 'page' | 'block' | 'user' | 'workspace';
    /** The ID that was attempted */
    attemptedId?: string;
    /** The input that led to the error */
    userInput?: string;
    /** The API endpoint that failed */
    endpoint?: string;
    /** HTTP status code if applicable */
    statusCode?: number;
    /** Original error from Notion API or other source */
    originalError?: any;
    /** Additional debug information */
    metadata?: Record<string, any>;
}
/**
 * Enhanced CLI Error with AI-friendly formatting
 */
export declare class NotionCLIError extends Error {
    readonly code: NotionCLIErrorCode;
    readonly userMessage: string;
    readonly suggestions: ErrorSuggestion[];
    readonly context: ErrorContext;
    readonly timestamp: string;
    constructor(code: NotionCLIErrorCode, userMessage: string, suggestions?: ErrorSuggestion[], context?: ErrorContext);
    /**
     * Format error for human-readable console output
     */
    toHumanString(): string;
    /**
     * Format error for JSON output (automation-friendly)
     */
    toJSON(): {
        success: boolean;
        error: {
            code: NotionCLIErrorCode;
            message: string;
            suggestions: ErrorSuggestion[];
            context: ErrorContext;
            timestamp: string;
        };
    };
    /**
     * Format error for compact JSON (single-line)
     */
    toCompactJSON(): string;
}
/**
 * Error factory functions for common scenarios
 */
export declare class NotionCLIErrorFactory {
    /**
     * Token is missing or not set
     */
    static tokenMissing(): NotionCLIError;
    /**
     * Token is invalid or expired
     */
    static tokenInvalid(): NotionCLIError;
    /**
     * Integration not shared with resource
     */
    static integrationNotShared(resourceType: 'database' | 'page', resourceId?: string): NotionCLIError;
    /**
     * Database/Page/Block not found
     */
    static resourceNotFound(resourceType: 'database' | 'page' | 'block', identifier: string): NotionCLIError;
    /**
     * Invalid ID format
     */
    static invalidIdFormat(input: string, resourceType?: 'database' | 'page' | 'block'): NotionCLIError;
    /**
     * Common confusion: using database_id when data_source_id is needed
     */
    static databaseIdConfusion(attemptedId: string): NotionCLIError;
    /**
     * Workspace not synced (cache miss for name resolution)
     */
    static workspaceNotSynced(databaseName: string): NotionCLIError;
    /**
     * Rate limited by Notion API
     */
    static rateLimited(retryAfter?: number): NotionCLIError;
    /**
     * Invalid JSON in filter or property value
     */
    static invalidJson(jsonString: string, parseError: Error): NotionCLIError;
    /**
     * Invalid property name or type
     */
    static invalidProperty(propertyName: string, databaseId?: string): NotionCLIError;
    /**
     * Network or connection error
     */
    static networkError(originalError: Error): NotionCLIError;
}
/**
 * Map Notion API errors to CLI errors with context
 */
export declare function wrapNotionError(error: any, context?: ErrorContext): NotionCLIError;
/**
 * Handle CLI errors with proper formatting based on output mode
 */
export declare function handleCliError(error: any, outputJson?: boolean, context?: ErrorContext): never;
