package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestNotionCLIError_Error(t *testing.T) {
	t.Run("without wrapped error", func(t *testing.T) {
		e := &NotionCLIError{Code: CodeTokenMissing, Message: "no token"}
		got := e.Error()
		want := "[TOKEN_MISSING] no token"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})

	t.Run("with wrapped error", func(t *testing.T) {
		inner := fmt.Errorf("connection refused")
		e := &NotionCLIError{Code: CodeNetworkError, Message: "network fail", Err: inner}
		got := e.Error()
		want := "[NETWORK_ERROR] network fail: connection refused"
		if got != want {
			t.Errorf("Error() = %q, want %q", got, want)
		}
	})
}

func TestNotionCLIError_Unwrap(t *testing.T) {
	inner := fmt.Errorf("root cause")
	e := &NotionCLIError{Code: CodeInternalError, Message: "oops", Err: inner}

	if !errors.Is(e, inner) {
		t.Error("errors.Is should find wrapped error")
	}

	e2 := &NotionCLIError{Code: CodeInternalError, Message: "no wrap"}
	if e2.Unwrap() != nil {
		t.Error("Unwrap should return nil when no wrapped error")
	}
}

func TestNotionCLIError_ExitCode(t *testing.T) {
	t.Run("API error has exit code 1", func(t *testing.T) {
		e := &NotionCLIError{Code: CodeNotFound, HTTPStatus: 404}
		if e.ExitCode() != ExitAPIError {
			t.Errorf("ExitCode() = %d, want %d", e.ExitCode(), ExitAPIError)
		}
	})

	t.Run("CLI error has exit code 2", func(t *testing.T) {
		e := &NotionCLIError{Code: CodeInvalidIDFormat}
		if e.ExitCode() != ExitCLIError {
			t.Errorf("ExitCode() = %d, want %d", e.ExitCode(), ExitCLIError)
		}
	})
}

func TestTokenMissing(t *testing.T) {
	e := TokenMissing()
	if e.Code != CodeTokenMissing {
		t.Errorf("Code = %q, want %q", e.Code, CodeTokenMissing)
	}
	if len(e.Suggestions) == 0 {
		t.Error("expected suggestions")
	}
}

func TestTokenInvalid(t *testing.T) {
	e := TokenInvalid("bad prefix")
	if e.Code != CodeTokenInvalid {
		t.Errorf("Code = %q, want %q", e.Code, CodeTokenInvalid)
	}
	if e.HTTPStatus != 401 {
		t.Errorf("HTTPStatus = %d, want 401", e.HTTPStatus)
	}
	if e.Details != "bad prefix" {
		t.Errorf("Details = %v, want %q", e.Details, "bad prefix")
	}
}

func TestIntegrationNotShared(t *testing.T) {
	e := IntegrationNotShared("database")
	if e.Code != CodePermissionDenied {
		t.Errorf("Code = %q, want %q", e.Code, CodePermissionDenied)
	}
	if e.HTTPStatus != 403 {
		t.Errorf("HTTPStatus = %d, want 403", e.HTTPStatus)
	}
}

func TestResourceNotFound(t *testing.T) {
	tests := []struct {
		resourceType string
		wantCode     string
	}{
		{"database", CodeDatabaseNotFound},
		{"page", CodePageNotFound},
		{"block", CodeBlockNotFound},
		{"user", CodeUserNotFound},
		{"comment", CodeNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.resourceType, func(t *testing.T) {
			e := ResourceNotFound(tt.resourceType, "abc-123")
			if e.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", e.Code, tt.wantCode)
			}
			if e.HTTPStatus != 404 {
				t.Errorf("HTTPStatus = %d, want 404", e.HTTPStatus)
			}
		})
	}
}

func TestInvalidIDFormat(t *testing.T) {
	e := InvalidIDFormat("not-an-id")
	if e.Code != CodeInvalidIDFormat {
		t.Errorf("Code = %q, want %q", e.Code, CodeInvalidIDFormat)
	}
	if e.Details != "not-an-id" {
		t.Errorf("Details = %v, want %q", e.Details, "not-an-id")
	}
}

func TestDatabaseIdConfusion(t *testing.T) {
	e := DatabaseIdConfusion("abc-123")
	if e.Code != CodeDatabaseIDConfusion {
		t.Errorf("Code = %q, want %q", e.Code, CodeDatabaseIDConfusion)
	}
}

func TestWorkspaceNotSynced(t *testing.T) {
	e := WorkspaceNotSynced()
	if e.Code != CodeWorkspaceNotSynced {
		t.Errorf("Code = %q, want %q", e.Code, CodeWorkspaceNotSynced)
	}
}

func TestRateLimited(t *testing.T) {
	t.Run("with retry-after", func(t *testing.T) {
		e := RateLimited("2s")
		if e.Code != CodeRateLimited {
			t.Errorf("Code = %q, want %q", e.Code, CodeRateLimited)
		}
		if e.HTTPStatus != 429 {
			t.Errorf("HTTPStatus = %d, want 429", e.HTTPStatus)
		}
		if e.Details != "2s" {
			t.Errorf("Details = %v, want %q", e.Details, "2s")
		}
	})

	t.Run("without retry-after", func(t *testing.T) {
		e := RateLimited("")
		if e.Message != "Rate limited by Notion API" {
			t.Errorf("Message = %q, want plain message", e.Message)
		}
	})
}

func TestInvalidJSON(t *testing.T) {
	e := InvalidJSON("unexpected EOF")
	if e.Code != CodeInvalidJSON {
		t.Errorf("Code = %q, want %q", e.Code, CodeInvalidJSON)
	}
	if e.Details != "unexpected EOF" {
		t.Errorf("Details = %v, want %q", e.Details, "unexpected EOF")
	}
}

func TestInvalidProperty(t *testing.T) {
	e := InvalidProperty("Status", "not a valid select option")
	if e.Code != CodeInvalidProperty {
		t.Errorf("Code = %q, want %q", e.Code, CodeInvalidProperty)
	}
	details, ok := e.Details.(map[string]string)
	if !ok {
		t.Fatal("Details should be map[string]string")
	}
	if details["property"] != "Status" {
		t.Errorf("Details.property = %q, want %q", details["property"], "Status")
	}
}

func TestNetworkError(t *testing.T) {
	inner := fmt.Errorf("dial tcp: no route to host")
	e := NetworkError(inner)
	if e.Code != CodeNetworkError {
		t.Errorf("Code = %q, want %q", e.Code, CodeNetworkError)
	}
	if !errors.Is(e, inner) {
		t.Error("should wrap the original error")
	}
}

func TestTimeout(t *testing.T) {
	e := Timeout("30s")
	if e.Code != CodeTimeout {
		t.Errorf("Code = %q, want %q", e.Code, CodeTimeout)
	}
	if e.Details != "30s" {
		t.Errorf("Details = %v, want %q", e.Details, "30s")
	}
}

func TestFromNotionAPI(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		body       map[string]any
		wantCode   string
	}{
		{
			"400 validation error",
			400,
			map[string]any{"code": "validation_error", "message": "invalid filter"},
			CodeValidationError,
		},
		{
			"401 unauthorized",
			401,
			map[string]any{"code": "unauthorized", "message": "API token is invalid"},
			CodeUnauthorized,
		},
		{
			"403 restricted",
			403,
			map[string]any{"code": "restricted_resource", "message": "no access"},
			CodePermissionDenied,
		},
		{
			"404 not found",
			404,
			map[string]any{"code": "object_not_found", "message": "not found"},
			CodeNotFound,
		},
		{
			"409 conflict",
			409,
			map[string]any{"code": "conflict_error", "message": "conflict"},
			CodeConflict,
		},
		{
			"429 rate limited",
			429,
			map[string]any{"code": "rate_limited", "message": "slow down"},
			CodeRateLimited,
		},
		{
			"502 bad gateway",
			502,
			map[string]any{"code": "", "message": "bad gateway"},
			CodeBadGateway,
		},
		{
			"503 service unavailable",
			503,
			map[string]any{"code": "service_unavailable", "message": "down"},
			CodeServiceUnavailable,
		},
		{
			"500 internal with Notion code",
			500,
			map[string]any{"code": "internal_server_error", "message": "oops"},
			CodeInternalError,
		},
		{
			"500 unknown code",
			500,
			map[string]any{"code": "something_new", "message": "surprise"},
			CodeInternalError,
		},
		{
			"empty body",
			500,
			map[string]any{},
			CodeInternalError,
		},
		{
			"400 with invalid_json code",
			400,
			map[string]any{"code": "invalid_json", "message": "bad json"},
			CodeInvalidJSON,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := FromNotionAPI(tt.statusCode, tt.body)
			if e.Code != tt.wantCode {
				t.Errorf("Code = %q, want %q", e.Code, tt.wantCode)
			}
			if e.HTTPStatus != tt.statusCode {
				t.Errorf("HTTPStatus = %d, want %d", e.HTTPStatus, tt.statusCode)
			}
			if len(e.Suggestions) == 0 {
				t.Error("expected suggestions")
			}
		})
	}
}

func TestExitCodeConstants(t *testing.T) {
	if ExitSuccess != 0 {
		t.Errorf("ExitSuccess = %d, want 0", ExitSuccess)
	}
	if ExitAPIError != 1 {
		t.Errorf("ExitAPIError = %d, want 1", ExitAPIError)
	}
	if ExitCLIError != 2 {
		t.Errorf("ExitCLIError = %d, want 2", ExitCLIError)
	}
}

func TestErrorCodeConstants(t *testing.T) {
	// Verify a selection of constants are non-empty and distinct.
	codes := []string{
		CodeUnauthorized, CodeTokenMissing, CodeTokenInvalid,
		CodeNotFound, CodeDatabaseNotFound, CodePageNotFound,
		CodeBlockNotFound, CodeRateLimited, CodeNetworkError,
		CodeTimeout, CodeInternalError,
	}
	seen := make(map[string]bool)
	for _, c := range codes {
		if c == "" {
			t.Error("error code constant is empty")
		}
		if seen[c] {
			t.Errorf("duplicate error code constant: %s", c)
		}
		seen[c] = true
	}
}
