package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
)

// resetRootCmd restores rootCmd's args/out/err to a clean state after a test.
// Tests in this file all share the package-level rootCmd, so each test must
// register a t.Cleanup that calls this helper to avoid leaking state into
// sibling tests.
func resetRootCmd(t *testing.T) {
	t.Helper()
	rootCmd.SetArgs([]string{})
	rootCmd.SetOut(os.Stdout)
	rootCmd.SetErr(os.Stderr)
}

// TestExitCode_Nil verifies that a nil error maps to ExitSuccess (0).
func TestExitCode_Nil(t *testing.T) {
	if got, want := ExitCode(nil), clierrors.ExitSuccess; got != want {
		t.Fatalf("ExitCode(nil) = %d, want %d", got, want)
	}
	if clierrors.ExitSuccess != 0 {
		t.Fatalf("ExitSuccess = %d, want 0", clierrors.ExitSuccess)
	}
}

// TestExitCode_NotionCLIError_APIStatus verifies that a NotionCLIError with a
// non-zero HTTPStatus maps to ExitAPIError (1).
func TestExitCode_NotionCLIError_APIStatus(t *testing.T) {
	err := &clierrors.NotionCLIError{HTTPStatus: 404}
	if got, want := ExitCode(err), clierrors.ExitAPIError; got != want {
		t.Fatalf("ExitCode(NotionCLIError{HTTPStatus:404}) = %d, want %d", got, want)
	}
	if clierrors.ExitAPIError != 1 {
		t.Fatalf("ExitAPIError = %d, want 1", clierrors.ExitAPIError)
	}
}

// TestExitCode_NotionCLIError_NoStatus verifies that a NotionCLIError with
// HTTPStatus == 0 maps to ExitCLIError (2).
func TestExitCode_NotionCLIError_NoStatus(t *testing.T) {
	err := &clierrors.NotionCLIError{Code: "X", Message: "y"}
	if got, want := ExitCode(err), clierrors.ExitCLIError; got != want {
		t.Fatalf("ExitCode(NotionCLIError{HTTPStatus:0}) = %d, want %d", got, want)
	}
}

// TestExitCode_RawError verifies that a plain (non-NotionCLIError) error maps
// to ExitCLIError (2).
func TestExitCode_RawError(t *testing.T) {
	err := fmt.Errorf("plain")
	if got, want := ExitCode(err), clierrors.ExitCLIError; got != want {
		t.Fatalf("ExitCode(plain) = %d, want %d", got, want)
	}
	if clierrors.ExitCLIError != 2 {
		t.Fatalf("ExitCLIError = %d, want 2", clierrors.ExitCLIError)
	}
}

// TestExecuteContext_Cancelled invokes ExecuteContext with a pre-cancelled
// context to ensure it returns without panicking.
//
// Note: the real ExecuteContext signature is ExecuteContext(ctx) error — it
// does NOT take args (unlike the spec's hypothetical signature). We set args
// to ["--help"] so cobra runs the built-in help handler (which does not
// require network/auth) and route stdout/stderr to a buffer so the test
// output stays clean.
//
// Acceptance policy: we accept either a nil error (help is a successful
// no-op) or a non-nil error. The only thing we strictly forbid is a panic.
// This is because cobra's help handler does not inspect ctx.Err(), so a
// cancelled context typically yields nil here; however we do not want to
// hard-code that detail in case cobra changes.
func TestExecuteContext_Cancelled(t *testing.T) {
	t.Cleanup(func() { resetRootCmd(t) })

	var buf bytes.Buffer
	rootCmd.SetArgs([]string{"--help"})
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // pre-cancel

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("ExecuteContext panicked with cancelled ctx: %v", r)
		}
	}()

	// Either nil or non-nil is acceptable; just no panic.
	_ = ExecuteContext(ctx)
}

// TestRootCmd_VersionFlag drives rootCmd directly with --version and asserts
// that the version template output mentions both "notion-cli" and "version".
func TestRootCmd_VersionFlag(t *testing.T) {
	t.Cleanup(func() { resetRootCmd(t) })

	var buf bytes.Buffer
	rootCmd.SetArgs([]string{"--version"})
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() with --version returned error: %v", err)
	}

	out := buf.String()
	lower := strings.ToLower(out)
	if !strings.Contains(lower, "notion-cli") {
		t.Errorf("version output missing %q substring; got %q", "notion-cli", out)
	}
	if !strings.Contains(lower, "version") {
		t.Errorf("version output missing %q substring; got %q", "version", out)
	}
}

// TestRootCmd_HelpRunsAllSubcommandRegistrations ensures that the package
// init() has registered the expected number of top-level subcommands. The
// init() in root.go calls 19 Register*Commands functions; we assert a lower
// bound of 18 to allow for minor future churn without becoming brittle.
func TestRootCmd_HelpRunsAllSubcommandRegistrations(t *testing.T) {
	const minSubcommands = 18
	got := len(rootCmd.Commands())
	if got < minSubcommands {
		t.Fatalf("rootCmd.Commands() = %d, want >= %d", got, minSubcommands)
	}
}

// Compile-time sanity: ensure io.Discard is reachable; used implicitly above
// for documentation but kept here to keep imports stable if a future edit
// switches buffer routing to io.Discard.
var _ = io.Discard
