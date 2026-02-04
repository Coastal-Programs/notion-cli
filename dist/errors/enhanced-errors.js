"use strict";
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
Object.defineProperty(exports, "__esModule", { value: true });
exports.NotionCLIErrorFactory = exports.NotionCLIError = exports.NotionCLIErrorCode = void 0;
exports.wrapNotionError = wrapNotionError;
exports.handleCliError = handleCliError;
/**
 * Comprehensive error codes covering all common scenarios
 */
var NotionCLIErrorCode;
(function (NotionCLIErrorCode) {
    // Authentication & Authorization
    NotionCLIErrorCode["UNAUTHORIZED"] = "UNAUTHORIZED";
    NotionCLIErrorCode["TOKEN_MISSING"] = "TOKEN_MISSING";
    NotionCLIErrorCode["TOKEN_INVALID"] = "TOKEN_INVALID";
    NotionCLIErrorCode["TOKEN_EXPIRED"] = "TOKEN_EXPIRED";
    NotionCLIErrorCode["PERMISSION_DENIED"] = "PERMISSION_DENIED";
    NotionCLIErrorCode["INTEGRATION_NOT_SHARED"] = "INTEGRATION_NOT_SHARED";
    // Resource Errors
    NotionCLIErrorCode["NOT_FOUND"] = "NOT_FOUND";
    NotionCLIErrorCode["OBJECT_NOT_FOUND"] = "OBJECT_NOT_FOUND";
    NotionCLIErrorCode["DATABASE_NOT_FOUND"] = "DATABASE_NOT_FOUND";
    NotionCLIErrorCode["PAGE_NOT_FOUND"] = "PAGE_NOT_FOUND";
    NotionCLIErrorCode["BLOCK_NOT_FOUND"] = "BLOCK_NOT_FOUND";
    // ID Format & Validation
    NotionCLIErrorCode["INVALID_ID_FORMAT"] = "INVALID_ID_FORMAT";
    NotionCLIErrorCode["INVALID_DATABASE_ID"] = "INVALID_DATABASE_ID";
    NotionCLIErrorCode["INVALID_PAGE_ID"] = "INVALID_PAGE_ID";
    NotionCLIErrorCode["INVALID_BLOCK_ID"] = "INVALID_BLOCK_ID";
    NotionCLIErrorCode["INVALID_URL"] = "INVALID_URL";
    // Common Confusions
    NotionCLIErrorCode["DATABASE_ID_CONFUSION"] = "DATABASE_ID_CONFUSION";
    NotionCLIErrorCode["WORKSPACE_VS_DATABASE"] = "WORKSPACE_VS_DATABASE";
    // API & Network
    NotionCLIErrorCode["RATE_LIMITED"] = "RATE_LIMITED";
    NotionCLIErrorCode["API_ERROR"] = "API_ERROR";
    NotionCLIErrorCode["NETWORK_ERROR"] = "NETWORK_ERROR";
    NotionCLIErrorCode["TIMEOUT"] = "TIMEOUT";
    NotionCLIErrorCode["SERVICE_UNAVAILABLE"] = "SERVICE_UNAVAILABLE";
    // Validation Errors
    NotionCLIErrorCode["VALIDATION_ERROR"] = "VALIDATION_ERROR";
    NotionCLIErrorCode["INVALID_PROPERTY"] = "INVALID_PROPERTY";
    NotionCLIErrorCode["INVALID_FILTER"] = "INVALID_FILTER";
    NotionCLIErrorCode["INVALID_JSON"] = "INVALID_JSON";
    NotionCLIErrorCode["MISSING_REQUIRED_FIELD"] = "MISSING_REQUIRED_FIELD";
    // Cache & State
    NotionCLIErrorCode["CACHE_ERROR"] = "CACHE_ERROR";
    NotionCLIErrorCode["WORKSPACE_NOT_SYNCED"] = "WORKSPACE_NOT_SYNCED";
    // General
    NotionCLIErrorCode["UNKNOWN"] = "UNKNOWN";
    NotionCLIErrorCode["INTERNAL_ERROR"] = "INTERNAL_ERROR";
})(NotionCLIErrorCode || (exports.NotionCLIErrorCode = NotionCLIErrorCode = {}));
/**
 * Enhanced CLI Error with AI-friendly formatting
 */
class NotionCLIError extends Error {
    constructor(code, userMessage, suggestions = [], context = {}) {
        super(userMessage);
        this.name = 'NotionCLIError';
        this.code = code;
        this.userMessage = userMessage;
        this.suggestions = suggestions;
        this.context = context;
        this.timestamp = new Date().toISOString();
        // Maintain proper stack trace
        if (Error.captureStackTrace) {
            Error.captureStackTrace(this, NotionCLIError);
        }
    }
    /**
     * Format error for human-readable console output
     */
    toHumanString() {
        const parts = [];
        // Error header with code
        parts.push(`\n❌ ${this.userMessage}`);
        parts.push(`   Error Code: ${this.code}`);
        // Add context if available
        if (this.context.attemptedId) {
            parts.push(`   Attempted ID: ${this.context.attemptedId}`);
        }
        if (this.context.resourceType) {
            parts.push(`   Resource Type: ${this.context.resourceType}`);
        }
        // Add suggestions
        if (this.suggestions.length > 0) {
            parts.push('\n💡 Possible causes and fixes:');
            this.suggestions.forEach((suggestion, index) => {
                parts.push(`   ${index + 1}. ${suggestion.description}`);
                if (suggestion.command) {
                    parts.push(`      $ ${suggestion.command}`);
                }
                if (suggestion.link) {
                    parts.push(`      🔗 ${suggestion.link}`);
                }
            });
        }
        // Add debug context in debug mode
        if (process.env.DEBUG && this.context.originalError) {
            parts.push('\n🐛 Debug Information:');
            parts.push(`   ${JSON.stringify(this.context.originalError, null, 2)}`);
        }
        parts.push(''); // Empty line at end
        return parts.join('\n');
    }
    /**
     * Format error for JSON output (automation-friendly)
     */
    toJSON() {
        return {
            success: false,
            error: {
                code: this.code,
                message: this.userMessage,
                suggestions: this.suggestions,
                context: this.context,
                timestamp: this.timestamp,
            }
        };
    }
    /**
     * Format error for compact JSON (single-line)
     */
    toCompactJSON() {
        return JSON.stringify(this.toJSON());
    }
}
exports.NotionCLIError = NotionCLIError;
/**
 * Error factory functions for common scenarios
 */
class NotionCLIErrorFactory {
    /**
     * Token is missing or not set
     */
    static tokenMissing() {
        return new NotionCLIError(NotionCLIErrorCode.TOKEN_MISSING, 'NOTION_TOKEN environment variable is not set', [
            {
                description: 'Set your Notion integration token using the config command',
                command: 'notion-cli config set-token'
            },
            {
                description: 'Or export it manually (Mac/Linux)',
                command: 'export NOTION_TOKEN="secret_your_token_here"'
            },
            {
                description: 'Or set it manually (Windows PowerShell)',
                command: '$env:NOTION_TOKEN="secret_your_token_here"'
            },
            {
                description: 'Get your integration token from Notion',
                link: 'https://developers.notion.com/docs/create-a-notion-integration'
            }
        ], { metadata: { tokenSet: false } });
    }
    /**
     * Token is invalid or expired
     */
    static tokenInvalid() {
        return new NotionCLIError(NotionCLIErrorCode.TOKEN_INVALID, 'Authentication failed - your NOTION_TOKEN is invalid or expired', [
            {
                description: 'Verify your integration still exists and is active',
                link: 'https://www.notion.so/my-integrations'
            },
            {
                description: 'Generate a new internal integration token',
                link: 'https://developers.notion.com/docs/create-a-notion-integration'
            },
            {
                description: 'Update your token using the config command',
                command: 'notion-cli config set-token'
            },
            {
                description: 'Check if the integration has been removed or revoked by workspace admin',
            }
        ], { metadata: { tokenSet: true } });
    }
    /**
     * Integration not shared with resource
     */
    static integrationNotShared(resourceType, resourceId) {
        return new NotionCLIError(NotionCLIErrorCode.INTEGRATION_NOT_SHARED, `Your integration doesn't have access to this ${resourceType}`, [
            {
                description: `Open the ${resourceType} in Notion and click the "..." menu`,
            },
            {
                description: 'Select "Add connections" or "Connect to"',
            },
            {
                description: 'Choose your integration from the list',
            },
            {
                description: 'If you don\'t see your integration, verify it exists',
                link: 'https://www.notion.so/my-integrations'
            },
            {
                description: 'Learn more about sharing with integrations',
                link: 'https://developers.notion.com/docs/create-a-notion-integration#give-your-integration-page-permissions'
            }
        ], {
            resourceType,
            attemptedId: resourceId,
            statusCode: 403
        });
    }
    /**
     * Database/Page/Block not found
     */
    static resourceNotFound(resourceType, identifier) {
        const isId = /^[a-f0-9]{32}$/i.test(identifier.replace(/-/g, ''));
        return new NotionCLIError(NotionCLIErrorCode.OBJECT_NOT_FOUND, `${resourceType.charAt(0).toUpperCase() + resourceType.slice(1)} not found: ${identifier}`, isId ? [
            {
                description: 'The ID may be incorrect - verify it in Notion',
            },
            {
                description: 'The integration may not have access - share the resource with your integration',
            },
            {
                description: 'The resource may have been deleted or archived',
            },
            {
                description: 'Try using the full Notion URL instead of just the ID',
                command: `notion-cli ${resourceType === 'database' ? 'db' : resourceType} retrieve https://notion.so/your-url-here`
            }
        ] : [
            {
                description: 'Run sync to refresh your workspace database index',
                command: 'notion-cli sync'
            },
            {
                description: 'List all available databases to find the correct name',
                command: 'notion-cli list'
            },
            {
                description: 'Try using the database ID or URL instead of name',
                command: `notion-cli ${resourceType === 'database' ? 'db' : resourceType} retrieve <ID_OR_URL>`
            }
        ], {
            resourceType,
            attemptedId: identifier,
            userInput: identifier,
            statusCode: 404
        });
    }
    /**
     * Invalid ID format
     */
    static invalidIdFormat(input, resourceType) {
        return new NotionCLIError(NotionCLIErrorCode.INVALID_ID_FORMAT, `Invalid ${resourceType || 'resource'} ID format: ${input}`, [
            {
                description: 'Notion IDs are 32 hexadecimal characters (with or without dashes)',
            },
            {
                description: 'Valid format: 1fb79d4c71bb8032b722c82305b63a00',
            },
            {
                description: 'Valid format: 1fb79d4c-71bb-8032-b722-c82305b63a00',
            },
            {
                description: 'Try using the full Notion URL instead',
                command: `notion-cli ${resourceType === 'database' ? 'db' : resourceType || 'page'} retrieve https://notion.so/your-url-here`
            },
            {
                description: 'Or find the resource by name after syncing',
                command: 'notion-cli sync && notion-cli list'
            }
        ], {
            resourceType,
            userInput: input,
            attemptedId: input
        });
    }
    /**
     * Common confusion: using database_id when data_source_id is needed
     */
    static databaseIdConfusion(attemptedId) {
        return new NotionCLIError(NotionCLIErrorCode.DATABASE_ID_CONFUSION, 'Notion API v5 uses "data_source_id" for databases, not "database_id"', [
            {
                description: 'This CLI automatically handles the conversion - you can use either',
            },
            {
                description: 'If you copied this from Notion API docs, the ID itself is still valid',
            },
            {
                description: 'Verify the database exists and is shared with your integration',
                command: 'notion-cli list'
            },
            {
                description: 'Try retrieving the database to check access',
                command: `notion-cli db retrieve ${attemptedId}`
            }
        ], {
            resourceType: 'database',
            attemptedId,
            metadata: { apiVersion: '5.2.1' }
        });
    }
    /**
     * Workspace not synced (cache miss for name resolution)
     */
    static workspaceNotSynced(databaseName) {
        return new NotionCLIError(NotionCLIErrorCode.WORKSPACE_NOT_SYNCED, `Database "${databaseName}" not found in workspace cache`, [
            {
                description: 'Run sync to index all accessible databases in your workspace',
                command: 'notion-cli sync'
            },
            {
                description: 'After syncing, list all databases to verify it was found',
                command: 'notion-cli list'
            },
            {
                description: 'If sync doesn\'t find it, the integration may not have access',
            },
            {
                description: 'Try using the database ID or URL directly instead of name',
                command: 'notion-cli db retrieve <DATABASE_ID_OR_URL>'
            }
        ], {
            resourceType: 'database',
            userInput: databaseName,
            metadata: { cacheState: 'not_synced' }
        });
    }
    /**
     * Rate limited by Notion API
     */
    static rateLimited(retryAfter) {
        return new NotionCLIError(NotionCLIErrorCode.RATE_LIMITED, 'Rate limited by Notion API - too many requests', [
            {
                description: retryAfter
                    ? `Wait ${retryAfter} seconds before retrying`
                    : 'Wait a few seconds before retrying',
            },
            {
                description: 'The CLI will automatically retry with exponential backoff',
            },
            {
                description: 'Consider using --page-size flag to reduce API calls',
                command: 'notion-cli db query <ID> --page-size 100'
            },
            {
                description: 'Learn about Notion API rate limits',
                link: 'https://developers.notion.com/reference/request-limits'
            }
        ], {
            statusCode: 429,
            metadata: { retryAfter }
        });
    }
    /**
     * Invalid JSON in filter or property value
     */
    static invalidJson(jsonString, parseError) {
        return new NotionCLIError(NotionCLIErrorCode.INVALID_JSON, 'Failed to parse JSON input', [
            {
                description: 'Check for common JSON syntax errors: missing quotes, trailing commas, unclosed brackets',
            },
            {
                description: 'Use a JSON validator to check your syntax',
                link: 'https://jsonlint.com/'
            },
            {
                description: 'For filters, you can use a filter file instead of inline JSON',
                command: 'notion-cli db query <ID> --file-filter ./filter.json'
            },
            {
                description: 'See filter examples in the documentation',
                link: 'https://developers.notion.com/reference/post-database-query-filter'
            }
        ], {
            userInput: jsonString,
            originalError: parseError,
            metadata: { parseError: parseError.message }
        });
    }
    /**
     * Invalid property name or type
     */
    static invalidProperty(propertyName, databaseId) {
        return new NotionCLIError(NotionCLIErrorCode.INVALID_PROPERTY, `Property "${propertyName}" not found or invalid`, [
            {
                description: 'Get the database schema to see all available properties',
                command: databaseId
                    ? `notion-cli db schema ${databaseId}`
                    : 'notion-cli db schema <DATABASE_ID>'
            },
            {
                description: 'Property names are case-sensitive - check exact spelling',
            },
            {
                description: 'Some property types don\'t support all operations',
            },
            {
                description: 'View the full database structure',
                command: databaseId
                    ? `notion-cli db retrieve ${databaseId} --raw`
                    : 'notion-cli db retrieve <DATABASE_ID> --raw'
            }
        ], {
            resourceType: 'database',
            attemptedId: databaseId,
            userInput: propertyName,
            metadata: { propertyName }
        });
    }
    /**
     * Network or connection error
     */
    static networkError(originalError) {
        return new NotionCLIError(NotionCLIErrorCode.NETWORK_ERROR, 'Network error - unable to reach Notion API', [
            {
                description: 'Check your internet connection',
            },
            {
                description: 'Verify Notion API status',
                link: 'https://status.notion.so/'
            },
            {
                description: 'The CLI will automatically retry transient network errors',
            },
            {
                description: 'If behind a proxy, ensure it\'s configured correctly',
            }
        ], {
            statusCode: 0,
            originalError,
            metadata: { errorCode: originalError.code }
        });
    }
}
exports.NotionCLIErrorFactory = NotionCLIErrorFactory;
/**
 * Map Notion API errors to CLI errors with context
 */
function wrapNotionError(error, context = {}) {
    var _a, _b, _c;
    // Already a NotionCLIError
    if (error instanceof NotionCLIError) {
        return error;
    }
    // Handle Notion API errors
    if (error.code) {
        // const _notionError = error as APIResponseError
        switch (error.code) {
            case 'unauthorized':
            case 'restricted_resource':
                return NotionCLIErrorFactory.tokenInvalid();
            case 'object_not_found':
                // Only pass valid resource types to resourceNotFound
                if (context.resourceType && ['database', 'page', 'block'].includes(context.resourceType)) {
                    return NotionCLIErrorFactory.resourceNotFound(context.resourceType, context.attemptedId || context.userInput || 'unknown');
                }
                return NotionCLIErrorFactory.resourceNotFound('database', context.attemptedId || context.userInput || 'unknown');
            case 'validation_error':
                if ((_a = error.message) === null || _a === void 0 ? void 0 : _a.includes('invalid json')) {
                    return NotionCLIErrorFactory.invalidJson(context.userInput || '', error);
                }
                return new NotionCLIError(NotionCLIErrorCode.VALIDATION_ERROR, error.message || 'Validation error', [
                    {
                        description: 'Check the API documentation for correct parameter format',
                        link: 'https://developers.notion.com/reference/intro'
                    }
                ], { ...context, originalError: error });
            case 'rate_limited': {
                const retryAfter = parseInt(((_b = error.headers) === null || _b === void 0 ? void 0 : _b['retry-after']) || '60', 10);
                return NotionCLIErrorFactory.rateLimited(retryAfter);
            }
            case 'conflict_error':
                return new NotionCLIError(NotionCLIErrorCode.API_ERROR, 'Conflict error - the resource is being modified by another request', [
                    {
                        description: 'Wait a moment and try again',
                    },
                    {
                        description: 'The CLI will automatically retry this operation',
                    }
                ], { ...context, originalError: error });
            case 'service_unavailable':
                return new NotionCLIError(NotionCLIErrorCode.SERVICE_UNAVAILABLE, 'Notion API is temporarily unavailable', [
                    {
                        description: 'Check Notion API status',
                        link: 'https://status.notion.so/'
                    },
                    {
                        description: 'The CLI will automatically retry this operation',
                    }
                ], { ...context, statusCode: 503, originalError: error });
        }
    }
    // Handle HTTP status codes
    if (error.status) {
        switch (error.status) {
            case 401:
            case 403: {
                const isTokenMissing = !process.env.NOTION_TOKEN;
                return isTokenMissing
                    ? NotionCLIErrorFactory.tokenMissing()
                    : NotionCLIErrorFactory.tokenInvalid();
            }
            case 404:
                // Only pass valid resource types to resourceNotFound
                if (context.resourceType && ['database', 'page', 'block'].includes(context.resourceType)) {
                    return NotionCLIErrorFactory.resourceNotFound(context.resourceType, context.attemptedId || context.userInput || 'unknown');
                }
                return new NotionCLIError(NotionCLIErrorCode.NOT_FOUND, 'Resource not found', [], { ...context, statusCode: 404, originalError: error });
            case 429: {
                const retryAfter = parseInt(((_c = error.headers) === null || _c === void 0 ? void 0 : _c['retry-after']) || '60', 10);
                return NotionCLIErrorFactory.rateLimited(retryAfter);
            }
            case 500:
            case 502:
            case 503:
            case 504:
                return new NotionCLIError(NotionCLIErrorCode.SERVICE_UNAVAILABLE, 'Notion API is experiencing issues', [
                    {
                        description: 'This is a temporary server error',
                    },
                    {
                        description: 'The CLI will automatically retry',
                    },
                    {
                        description: 'Check Notion API status',
                        link: 'https://status.notion.so/'
                    }
                ], { ...context, statusCode: error.status, originalError: error });
        }
    }
    // Handle network errors
    if (error.code && ['ECONNRESET', 'ETIMEDOUT', 'ENOTFOUND', 'EAI_AGAIN'].includes(error.code)) {
        return NotionCLIErrorFactory.networkError(error);
    }
    // Generic error
    return new NotionCLIError(NotionCLIErrorCode.UNKNOWN, error.message || 'An unexpected error occurred', [
        {
            description: 'If this error persists, please report it',
            link: 'https://github.com/Coastal-Programs/notion-cli/issues'
        }
    ], { ...context, originalError: error });
}
/**
 * Handle CLI errors with proper formatting based on output mode
 */
function handleCliError(error, outputJson = false, context = {}) {
    const cliError = error instanceof NotionCLIError
        ? error
        : wrapNotionError(error, context);
    if (outputJson) {
        console.log(JSON.stringify(cliError.toJSON(), null, 2));
    }
    else {
        console.error(cliError.toHumanString());
    }
    process.exit(1);
}
