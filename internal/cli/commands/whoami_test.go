package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

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
