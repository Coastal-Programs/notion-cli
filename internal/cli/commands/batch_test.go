package commands

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/spf13/cobra"
)

func testBatchServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
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

func runBatchRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterBatchCommands(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

// ---------------------------------------------------------------------------
// retrieveResource pure-function tests (uses a fake cobra command)
// ---------------------------------------------------------------------------

func makeFakeCmd(t *testing.T, srvURL string) (*cobra.Command, func()) {
	t.Helper()
	origBase := os.Getenv("NOTION_CLI_BASE_URL")
	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Setenv("NOTION_CLI_BASE_URL", srvURL)
	_ = os.Setenv("NOTION_TOKEN", "secret_test_token")

	cmd := &cobra.Command{Use: "fake", SilenceErrors: true, SilenceUsage: true}
	addOutputFlags(cmd)
	cmd.SetContext(context.Background())
	if err := cmd.ParseFlags([]string{}); err != nil {
		t.Fatalf("ParseFlags: %v", err)
	}

	return cmd, func() {
		if origBase == "" {
			_ = os.Unsetenv("NOTION_CLI_BASE_URL")
		} else {
			_ = os.Setenv("NOTION_CLI_BASE_URL", origBase)
		}
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	}
}

func TestRetrieveResource_Page(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/pages/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"object":"page","id":"page-1"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cmd, cleanup := makeFakeCmd(t, srv.URL)
	defer cleanup()

	client, err := newClientForCommand(cmd)
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}

	result, err := retrieveResource(cmd, client, "page", "00000000000000000000000000000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["object"] != "page" {
		t.Errorf("object = %v, want page", result["object"])
	}
}

func TestRetrieveResource_Block(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/blocks/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"object":"block","id":"block-1"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cmd, cleanup := makeFakeCmd(t, srv.URL)
	defer cleanup()

	client, err := newClientForCommand(cmd)
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}

	result, err := retrieveResource(cmd, client, "block", "00000000000000000000000000000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["object"] != "block" {
		t.Errorf("object = %v, want block", result["object"])
	}
}

func TestRetrieveResource_Database(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/databases/") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`{"object":"database","id":"db-1"}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer srv.Close()

	cmd, cleanup := makeFakeCmd(t, srv.URL)
	defer cleanup()

	client, err := newClientForCommand(cmd)
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}

	result, err := retrieveResource(cmd, client, "db", "00000000000000000000000000000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["object"] != "database" {
		t.Errorf("object = %v, want database", result["object"])
	}
}

func TestRetrieveResource_UnknownType(t *testing.T) {
	cmd := &cobra.Command{Use: "fake", SilenceErrors: true}
	addOutputFlags(cmd)
	cmd.SetContext(context.Background())
	_ = cmd.ParseFlags([]string{})

	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Setenv("NOTION_TOKEN", "secret_test_token")
	t.Cleanup(func() {
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})

	client, err := newClientForCommand(cmd)
	if err != nil {
		t.Fatalf("newClient: %v", err)
	}

	_, resErr := retrieveResource(cmd, client, "unsupported_type", "some-id")
	if resErr == nil {
		t.Fatal("expected error for unknown resource type")
	}
	cliErr, ok := resErr.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("expected *NotionCLIError, got %T", resErr)
	}
	if cliErr.Code != clierrors.CodeInvalidRequest {
		t.Errorf("Code = %q, want %q", cliErr.Code, clierrors.CodeInvalidRequest)
	}
}

// ---------------------------------------------------------------------------
// Command-level tests
// ---------------------------------------------------------------------------

func TestBatchRetrieve_NoIDs(t *testing.T) {
	// No args, no --ids, stdin is a terminal → CodeMissingRequired.
	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Setenv("NOTION_TOKEN", "secret_test_token")
	t.Cleanup(func() {
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})

	_, _, err := runBatchRoot(t, "batch", "retrieve")
	if err == nil {
		t.Fatal("expected error when no IDs provided")
	}
}

func TestBatchRetrieve_WithArgs(t *testing.T) {
	const id1 = "00000000000000000000000000000001"
	const id2 = "00000000000000000000000000000002"

	_, cleanup := testBatchServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"object":"page","id":"page-1"}`))
	})
	defer cleanup()

	var err error
	out := captureStdout(t, func() {
		_, _, err = runBatchRoot(t, "batch", "retrieve", id1, id2, "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	data, _ := envelope["data"].(map[string]any)
	if rc, ok := data["result_count"].(float64); !ok || rc != 2 {
		t.Errorf("result_count = %v, want 2", data["result_count"])
	}
}

func TestBatchRetrieve_WithIdsFlag(t *testing.T) {
	const id1 = "00000000000000000000000000000001"
	const id2 = "00000000000000000000000000000002"

	_, cleanup := testBatchServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"object":"page","id":"page-1"}`))
	})
	defer cleanup()

	var err error
	out := captureStdout(t, func() {
		_, _, err = runBatchRoot(t, "batch", "retrieve", "--ids", id1+","+id2, "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	data, _ := envelope["data"].(map[string]any)
	if rc, ok := data["result_count"].(float64); !ok || rc != 2 {
		t.Errorf("result_count = %v, want 2", data["result_count"])
	}
}

func TestBatchRetrieve_PartialError(t *testing.T) {
	const id1 = "00000000000000000000000000000001"
	const id2 = "00000000000000000000000000000002"

	// The formatted UUID for id2 ends with ...00000002.
	_, cleanup := testBatchServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "00000002") {
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"code":"object_not_found","message":"not found"}`))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"object":"page","id":"` + id1 + `"}`))
	})
	defer cleanup()

	var err error
	out := captureStdout(t, func() {
		// batch retrieve succeeds even with partial errors.
		_, _, err = runBatchRoot(t, "batch", "retrieve", id1, id2, "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	data, _ := envelope["data"].(map[string]any)
	if rc, ok := data["result_count"].(float64); !ok || rc != 1 {
		t.Errorf("result_count = %v, want 1", data["result_count"])
	}
	if ec, ok := data["error_count"].(float64); !ok || ec != 1 {
		t.Errorf("error_count = %v, want 1", data["error_count"])
	}
}

func TestBatchRetrieve_TypeBlock(t *testing.T) {
	const id1 = "00000000000000000000000000000001"

	var hitPath string
	_, cleanup := testBatchServer(t, func(w http.ResponseWriter, r *http.Request) {
		hitPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"object":"block","id":"block-1"}`))
	})
	defer cleanup()

	_, _, err := runBatchRoot(t, "batch", "retrieve", "--type", "block", id1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(hitPath, "/blocks/") {
		t.Errorf("path = %q, want /blocks/...", hitPath)
	}
}
