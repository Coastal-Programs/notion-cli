package commands

import (
	"errors"
	"testing"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/internal/notion"
	"github.com/spf13/cobra"
)

// A plan-limit 403 must surface as BLOCK_LIMIT_EXCEEDED, not as a generic
// "share the page" permission error. Shape from
// https://developers.notion.com/reference/workspace-block-limits.
func TestHandleError_BlockLimit403(t *testing.T) {
	cmd := &cobra.Command{Use: "x"}
	addOutputFlags(cmd)

	apiErr := &notion.APIError{
		Status:         403,
		Code:           "restricted_resource",
		Message:        "This workspace has used all of its free blocks.",
		AdditionalData: map[string]any{"block_limit": "block_creation"},
	}

	err := whoamiCaptureStderrErr(t, func() error { return handleError(cmd, apiErr) })

	var cliErr *clierrors.NotionCLIError
	if !errors.As(err, &cliErr) {
		t.Fatalf("expected *NotionCLIError, got %T", err)
	}
	if cliErr.Code != clierrors.CodeBlockLimitExceeded {
		t.Errorf("Code = %q, want %q", cliErr.Code, clierrors.CodeBlockLimitExceeded)
	}
	for _, s := range cliErr.Suggestions {
		if s == "Share the resource with your integration in Notion" {
			t.Errorf("block-limit error must not suggest sharing the page")
		}
	}

	// An ordinary 403 keeps the sharing guidance.
	plain := &notion.APIError{Status: 403, Code: "restricted_resource", Message: "no access"}
	err = whoamiCaptureStderrErr(t, func() error { return handleError(cmd, plain) })
	if !errors.As(err, &cliErr) || cliErr.Code != clierrors.CodePermissionDenied {
		t.Errorf("plain 403: got %v, want %s", err, clierrors.CodePermissionDenied)
	}
}

// whoamiCaptureStderrErr runs fn with stderr silenced and returns its error.
func whoamiCaptureStderrErr(t *testing.T, fn func() error) error {
	t.Helper()
	var err error
	_ = whoamiCaptureStderr(t, func() { err = fn() })
	return err
}
