export declare enum ErrorCode {
    RATE_LIMITED = "RATE_LIMITED",
    NOT_FOUND = "NOT_FOUND",
    UNAUTHORIZED = "UNAUTHORIZED",
    VALIDATION_ERROR = "VALIDATION_ERROR",
    API_ERROR = "API_ERROR",
    UNKNOWN = "UNKNOWN"
}
export declare class NotionCLIError extends Error {
    code: ErrorCode;
    details?: any;
    notionError?: any;
    constructor(code: ErrorCode, message: string, details?: any, notionError?: any);
    toJSON(): {
        success: boolean;
        error: {
            code: ErrorCode;
            message: string;
            details: any;
            notionError: any;
        };
        timestamp: string;
    };
}
export declare const wrapNotionError: (error: any) => NotionCLIError;
