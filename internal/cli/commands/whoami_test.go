package commands

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// whoamiCaptureStderr redirects os.Stderr for the duration of fn() and
// returns whatever was written to it. Local helper for whoami envelope
// assertions; named with a scenario-specific prefix to avoid colliding
// with any shared helper.
func whoamiCaptureStderr(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stderr
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(): %v", err)
	}
	os.Stderr = w

	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	fn()

	_ = w.Close()
	os.Stderr = orig
	<-done
	_ = r.Close()
	return buf.String()
}

func testWhoamiServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)

	origToken := os.Getenv("NOTION_TOKEN")
	origBase := os.Getenv("NOTION_CLI_BASE_URL")
	_ = os.Setenv("NOTION_TOKEN", "secret_test_token")
	_ = os.Setenv("NOTION_CLI_BASE_URL", srv.URL)

	return srv, func() {
		srv.Close()
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
		if origBase == "" {
			_ = os.Unsetenv("NOTION_CLI_BASE_URL")
		} else {
			_ = os.Setenv("NOTION_CLI_BASE_URL", origBase)
		}
	}
}

func runWhoamiRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterWhoamiCommand(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

func TestWhoamiCmd_Success(t *testing.T) {
	_, cleanup := testWhoamiServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/me" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"object":"user","id":"bot-1","name":"My Bot","type":"bot"}`))
	})
	defer cleanup()

	var err error
	out := captureStdout(t, func() {
		_, _, err = runWhoamiRoot(t, "whoami", "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	if envelope["success"] != true {
		t.Errorf("expected success=true")
	}
	data, _ := envelope["data"].(map[string]any)
	if data["api_latency"] == nil {
		t.Error("api_latency should be present in whoami output")
	}
}

func TestWhoamiCmd_APIError(t *testing.T) {
	_, cleanup := testWhoamiServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"code":"unauthorized","message":"API token is invalid"}`))
	})
	defer cleanup()

	_, _, err := runWhoamiRoot(t, "whoami")
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

// whoamiDisableRetries sets NOTION_CLI_MAX_RETRIES=0 so retryable status codes
// (429, 503, etc.) do not slow the test down with the default 1s base delay.
// Returns a cleanup func that restores the prior value.
func whoamiDisableRetries(t *testing.T) func() {
	t.Helper()
	orig, had := os.LookupEnv("NOTION_CLI_MAX_RETRIES")
	_ = os.Setenv("NOTION_CLI_MAX_RETRIES", "0")
	return func() {
		if had {
			_ = os.Setenv("NOTION_CLI_MAX_RETRIES", orig)
		} else {
			_ = os.Unsetenv("NOTION_CLI_MAX_RETRIES")
		}
	}
}

// TestWhoamiCmd_404MapsToNotFoundEnvelope verifies that a 404 from the API
// with Notion's object_not_found code is translated into the CLI's NOT_FOUND
// envelope code and emitted to stderr as a JSON envelope with success=false.
func TestWhoamiCmd_404MapsToNotFoundEnvelope(t *testing.T) {
	defer whoamiDisableRetries(t)()
	_, cleanup := testWhoamiServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"object":"error","status":404,"code":"object_not_found","message":"Not found"}`))
	})
	defer cleanup()

	var err error
	stderr := whoamiCaptureStderr(t, func() {
		_, _, err = runWhoamiRoot(t, "whoami", "--output", "json")
	})
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(stderr, `"success": false`) {
		t.Errorf("stderr does not contain success=false envelope; got:\n%s", stderr)
	}
	if !strings.Contains(stderr, `"code": "NOT_FOUND"`) {
		t.Errorf("stderr does not contain code=NOT_FOUND; got:\n%s", stderr)
	}
}

// TestWhoamiCmd_429MapsToRateLimitedEnvelope verifies that a 429 from the API
// is mapped to the CLI's RATE_LIMITED envelope code on stderr. Retries are
// disabled via NOTION_CLI_MAX_RETRIES=0 so the 429 surfaces immediately
// instead of being retried with the default 1s+ backoff.
func TestWhoamiCmd_429MapsToRateLimitedEnvelope(t *testing.T) {
	defer whoamiDisableRetries(t)()
	_, cleanup := testWhoamiServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"object":"error","status":429,"code":"rate_limited","message":"slow down"}`))
	})
	defer cleanup()

	var err error
	stderr := whoamiCaptureStderr(t, func() {
		_, _, err = runWhoamiRoot(t, "whoami", "--output", "json")
	})
	if err == nil {
		t.Fatal("expected error for 429 response")
	}
	if !strings.Contains(stderr, `"success": false`) {
		t.Errorf("stderr does not contain success=false envelope; got:\n%s", stderr)
	}
	if !strings.Contains(stderr, `"code": "RATE_LIMITED"`) {
		t.Errorf("stderr does not contain code=RATE_LIMITED; got:\n%s", stderr)
	}
}

func TestWhoamiCmd_WithWorkspaceOwner(t *testing.T) {
	_, cleanup := testWhoamiServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{
			"object":"user",
			"id":"bot-1",
			"name":"My Bot",
			"type":"bot",
			"bot":{"owner":{"type":"workspace","workspace":true}}
		}`))
	})
	defer cleanup()

	var err error
	out := captureStdout(t, func() {
		_, _, err = runWhoamiRoot(t, "whoami", "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	data, _ := envelope["data"].(map[string]any)
	if data["workspace"] == nil {
		t.Error("workspace field should be present for workspace-level integration")
	}
}
