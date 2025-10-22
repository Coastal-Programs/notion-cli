"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.wrapNotionError = exports.NotionCLIError = exports.ErrorCode = void 0;
var ErrorCode;
(function (ErrorCode) {
    ErrorCode["RATE_LIMITED"] = "RATE_LIMITED";
    ErrorCode["NOT_FOUND"] = "NOT_FOUND";
    ErrorCode["UNAUTHORIZED"] = "UNAUTHORIZED";
    ErrorCode["VALIDATION_ERROR"] = "VALIDATION_ERROR";
    ErrorCode["API_ERROR"] = "API_ERROR";
    ErrorCode["UNKNOWN"] = "UNKNOWN";
})(ErrorCode = exports.ErrorCode || (exports.ErrorCode = {}));
class NotionCLIError extends Error {
    constructor(code, message, details, notionError) {
        super(message);
        this.code = code;
        this.details = details;
        this.notionError = notionError;
        this.name = 'NotionCLIError';
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
        };
    }
}
exports.NotionCLIError = NotionCLIError;
const wrapNotionError = (error) => {
    var _a;
    // Map HTTP status codes to error codes
    if (error.status === 429) {
        return new NotionCLIError(ErrorCode.RATE_LIMITED, 'Rate limit exceeded', { retryAfter: (_a = error.headers) === null || _a === void 0 ? void 0 : _a['retry-after'] }, error);
    }
    if (error.status === 401 || error.status === 403) {
        return new NotionCLIError(ErrorCode.UNAUTHORIZED, 'Authentication failed. Check your NOTION_TOKEN.', null, error);
    }
    if (error.status === 404) {
        return new NotionCLIError(ErrorCode.NOT_FOUND, 'Resource not found', null, error);
    }
    if (error.status === 400) {
        return new NotionCLIError(ErrorCode.VALIDATION_ERROR, 'Invalid request parameters', null, error);
    }
    // Default API error
    return new NotionCLIError(ErrorCode.API_ERROR, error.message || 'Notion API error', null, error);
};
exports.wrapNotionError = wrapNotionError;
