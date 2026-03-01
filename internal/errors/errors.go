// Package errors provides structured error types and factory functions for the
// Notion CLI. All errors are represented as NotionCLIError, which includes an
// error code, human-readable message, optional details, and user-facing
// suggestions for resolution.
package errors

import "fmt"

// Exit codes returned by the CLI process.
const (
	ExitSuccess  = 0
	ExitAPIError = 1
	ExitCLIError = 2
)

// Error code constants. These are stable string identifiers used for
// programmatic error handling and JSON envelope responses.
const (
	CodeUnauthorized        = "UNAUTHORIZED"
	CodeTokenMissing        = "TOKEN_MISSING"
	CodeTokenInvalid        = "TOKEN_INVALID"
	CodeNotFound            = "NOT_FOUND"
	CodeDatabaseNotFound    = "DATABASE_NOT_FOUND"
	CodePageNotFound        = "PAGE_NOT_FOUND"
	CodeBlockNotFound       = "BLOCK_NOT_FOUND"
	CodeUserNotFound        = "USER_NOT_FOUND"
	CodeInvalidIDFormat     = "INVALID_ID_FORMAT"
	CodeDatabaseIDConfusion = "DATABASE_ID_CONFUSION"
	CodeRateLimited         = "RATE_LIMITED"
	CodeValidationError     = "VALIDATION_ERROR"
	CodeInvalidJSON         = "INVALID_JSON"
	CodeWorkspaceNotSynced  = "WORKSPACE_NOT_SYNCED"
	CodeNetworkError        = "NETWORK_ERROR"
	CodeTimeout             = "TIMEOUT"
	CodeInternalError       = "INTERNAL_ERROR"
	CodeConflict            = "CONFLICT"
	CodeServiceUnavailable  = "SERVICE_UNAVAILABLE"
	CodeInvalidRequest      = "INVALID_REQUEST"
	CodePropertyNotFound    = "PROPERTY_NOT_FOUND"
	CodePermissionDenied    = "PERMISSION_DENIED"
	CodeBadGateway          = "BAD_GATEWAY"
	CodeInvalidFilter       = "INVALID_FILTER"
	CodeInvalidSort         = "INVALID_SORT"
	CodeInvalidProperty     = "INVALID_PROPERTY"
	CodeMissingRequired     = "MISSING_REQUIRED"
	CodeInvalidEnum         = "INVALID_ENUM"
	CodeObjectNotFound      = "OBJECT_NOT_FOUND"
	CodeSizeLimitExceeded   = "SIZE_LIMIT_EXCEEDED"
)

// NotionCLIError is the canonical error type for the CLI. It carries structured
// information about what went wrong and how to fix it.
type NotionCLIError struct {
	Code        string   `json:"code"`
	Message     string   `json:"message"`
	Details     any      `json:"details,omitempty"`
	Suggestions []string `json:"suggestions,omitempty"`
	HTTPStatus  int      `json:"http_status,omitempty"`
	Err         error    `json:"-"`
}

// Error implements the error interface.
func (e *NotionCLIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// Unwrap returns the underlying error, supporting errors.Is/As chains.
func (e *NotionCLIError) Unwrap() error {
	return e.Err
}

// ExitCode returns the process exit code appropriate for this error.
func (e *NotionCLIError) ExitCode() int {
	if e.HTTPStatus > 0 {
		return ExitAPIError
	}
	return ExitCLIError
}

// ---------------------------------------------------------------------------
// Factory functions
// ---------------------------------------------------------------------------

// TokenMissing returns an error indicating no API token is configured.
func TokenMissing() *NotionCLIError {
	return &NotionCLIError{
		Code:    CodeTokenMissing,
		Message: "Notion API token is not configured",
		Suggestions: []string{
			"Set the NOTION_TOKEN environment variable",
			"Run 'notion-cli init' to configure your token",
			"Get a token at https://www.notion.so/my-integrations",
		},
	}
}

// TokenInvalid returns an error for a malformed or rejected API token.
func TokenInvalid(detail string) *NotionCLIError {
	return &NotionCLIError{
		Code:       CodeTokenInvalid,
		Message:    "Notion API token is invalid",
		Details:    detail,
		HTTPStatus: 401,
		Suggestions: []string{
			"Verify your token at https://www.notion.so/my-integrations",
			"Ensure the token starts with 'secret_' or 'ntn_'",
			"Run 'notion-cli init' to reconfigure",
		},
	}
}

// IntegrationNotShared returns an error when a resource is not shared with
// the integration.
func IntegrationNotShared(resource string) *NotionCLIError {
	return &NotionCLIError{
		Code:       CodePermissionDenied,
		Message:    fmt.Sprintf("Integration does not have access to this %s", resource),
		HTTPStatus: 403,
		Suggestions: []string{
			fmt.Sprintf("Share the %s with your integration in Notion", resource),
			"Open the page/database in Notion → ··· → Connections → Add your integration",
		},
	}
}

// ResourceNotFound returns an error for a missing Notion resource.
func ResourceNotFound(resourceType, id string) *NotionCLIError {
	code := CodeNotFound
	switch resourceType {
	case "database":
		code = CodeDatabaseNotFound
	case "page":
		code = CodePageNotFound
	case "block":
		code = CodeBlockNotFound
	case "user":
		code = CodeUserNotFound
	}
	return &NotionCLIError{
		Code:       code,
		Message:    fmt.Sprintf("%s not found: %s", resourceType, id),
		HTTPStatus: 404,
		Suggestions: []string{
			"Verify the ID is correct",
			"Ensure the resource is shared with your integration",
			"Check if the resource has been deleted or archived",
		},
	}
}

// InvalidIDFormat returns an error for a string that cannot be parsed as a
// Notion resource ID.
func InvalidIDFormat(id string) *NotionCLIError {
	return &NotionCLIError{
		Code:    CodeInvalidIDFormat,
		Message: fmt.Sprintf("Invalid ID format: %s", id),
		Details: id,
		Suggestions: []string{
			"Notion IDs are 32 hex characters (with or without hyphens)",
			"You can also paste a Notion URL and the ID will be extracted",
			"Example: 8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
		},
	}
}

// DatabaseIdConfusion returns an error when a database_id is used where a
// data_source_id is expected, or vice versa.
func DatabaseIdConfusion(id string) *NotionCLIError {
	return &NotionCLIError{
		Code:    CodeDatabaseIDConfusion,
		Message: fmt.Sprintf("Wrong database ID type: %s", id),
		Details: id,
		Suggestions: []string{
			"Notion databases have two IDs: database_id and data_source_id",
			"Use 'notion-cli db retrieve' with either ID - the CLI will resolve it",
			"Run 'notion-cli sync' then 'notion-cli list' to see correct IDs",
		},
	}
}

// WorkspaceNotSynced returns an error when an operation requires workspace
// data that has not been cached yet.
func WorkspaceNotSynced() *NotionCLIError {
	return &NotionCLIError{
		Code:    CodeWorkspaceNotSynced,
		Message: "Workspace has not been synced yet",
		Suggestions: []string{
			"Run 'notion-cli sync' to cache workspace databases",
			"This is required for name-based database resolution",
		},
	}
}

// RateLimited returns an error when the Notion API returns 429.
func RateLimited(retryAfter string) *NotionCLIError {
	msg := "Rate limited by Notion API"
	if retryAfter != "" {
		msg = fmt.Sprintf("Rate limited by Notion API (retry after %s)", retryAfter)
	}
	return &NotionCLIError{
		Code:       CodeRateLimited,
		Message:    msg,
		Details:    retryAfter,
		HTTPStatus: 429,
		Suggestions: []string{
			"The request will be retried automatically",
			"Reduce request frequency if this persists",
		},
	}
}

// InvalidJSON returns an error for malformed JSON input.
func InvalidJSON(detail string) *NotionCLIError {
	return &NotionCLIError{
		Code:    CodeInvalidJSON,
		Message: "Invalid JSON input",
		Details: detail,
		Suggestions: []string{
			"Verify your JSON is well-formed",
			"Use a JSON validator to check syntax",
			"Ensure strings are double-quoted",
		},
	}
}

// InvalidProperty returns an error for a property name/value problem.
func InvalidProperty(name, reason string) *NotionCLIError {
	return &NotionCLIError{
		Code:    CodeInvalidProperty,
		Message: fmt.Sprintf("Invalid property '%s': %s", name, reason),
		Details: map[string]string{"property": name, "reason": reason},
		Suggestions: []string{
			"Run 'notion-cli db schema <DATABASE_ID>' to see valid properties",
			"Property names are case-sensitive",
		},
	}
}

// NetworkError returns an error for connection-level failures.
func NetworkError(err error) *NotionCLIError {
	return &NotionCLIError{
		Code:    CodeNetworkError,
		Message: "Network error communicating with Notion API",
		Err:     err,
		Suggestions: []string{
			"Check your internet connection",
			"Verify https://api.notion.com is reachable",
			"Check if a proxy or firewall is blocking the request",
		},
	}
}

// Timeout returns an error when a request exceeds the allowed duration.
func Timeout(duration string) *NotionCLIError {
	return &NotionCLIError{
		Code:    CodeTimeout,
		Message: fmt.Sprintf("Request timed out after %s", duration),
		Details: duration,
		Suggestions: []string{
			"The Notion API may be experiencing high load",
			"Try again in a few moments",
			"For large operations, consider breaking them into smaller batches",
		},
	}
}

// FromNotionAPI parses a Notion API error response into a NotionCLIError.
// The body is expected to contain "code" and "message" keys as returned by
// the Notion API.
func FromNotionAPI(statusCode int, body map[string]any) *NotionCLIError {
	apiCode, _ := body["code"].(string)
	apiMessage, _ := body["message"].(string)

	if apiMessage == "" {
		apiMessage = "Unknown API error"
	}

	e := &NotionCLIError{
		Message:    apiMessage,
		Details:    body,
		HTTPStatus: statusCode,
	}

	switch statusCode {
	case 400:
		e.Code = CodeValidationError
		e.Suggestions = []string{
			"Check the request parameters",
			"Run with --verbose for more details",
		}
	case 401:
		e.Code = CodeUnauthorized
		e.Suggestions = []string{
			"Your API token may be invalid or expired",
			"Run 'notion-cli init' to reconfigure",
		}
	case 403:
		e.Code = CodePermissionDenied
		e.Suggestions = []string{
			"Share the resource with your integration in Notion",
			"Open the page/database → ··· → Connections → Add your integration",
		}
	case 404:
		e.Code = CodeNotFound
		e.Suggestions = []string{
			"Verify the resource ID is correct",
			"Ensure the resource is shared with your integration",
		}
	case 409:
		e.Code = CodeConflict
		e.Suggestions = []string{
			"The resource may have been modified concurrently",
			"Retry the operation",
		}
	case 429:
		e.Code = CodeRateLimited
		e.Suggestions = []string{
			"Request will be retried automatically",
			"Reduce request frequency if this persists",
		}
	case 502:
		e.Code = CodeBadGateway
		e.Suggestions = []string{
			"The Notion API may be temporarily unavailable",
			"Try again in a few moments",
		}
	case 503:
		e.Code = CodeServiceUnavailable
		e.Suggestions = []string{
			"The Notion API is temporarily unavailable",
			"Check https://status.notion.so for service status",
		}
	default:
		e.Code = CodeInternalError
		e.Suggestions = []string{
			"An unexpected error occurred",
			"Try again or contact support if this persists",
		}
	}

	// Override code with Notion's own code if it maps to something specific.
	switch apiCode {
	case "unauthorized":
		e.Code = CodeUnauthorized
	case "restricted_resource":
		e.Code = CodePermissionDenied
	case "object_not_found":
		e.Code = CodeNotFound
	case "rate_limited":
		e.Code = CodeRateLimited
	case "invalid_json":
		e.Code = CodeInvalidJSON
	case "validation_error":
		e.Code = CodeValidationError
	case "conflict_error":
		e.Code = CodeConflict
	case "internal_server_error":
		e.Code = CodeInternalError
	case "service_unavailable":
		e.Code = CodeServiceUnavailable
	}

	return e
}
