package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func testUserServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
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

func runUserRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterUserCommands(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

func TestRegisterUserCommands(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterUserCommands(root)

	for _, sub := range []string{"list", "retrieve", "bot"} {
		c, _, err := root.Find([]string{"user", sub})
		if err != nil || c == nil || c.Name() != sub {
			t.Errorf("subcommand user %q not found: %v", sub, err)
		}
	}
}

func TestUserList_Success(t *testing.T) {
	_, cleanup := testUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"results":[{"object":"user","id":"u1","name":"Alice","type":"person"}],"has_more":false}`))
	})
	defer cleanup()

	var err error
	out := captureStdout(t, func() {
		_, _, err = runUserRoot(t, "user", "list", "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	if envelope["success"] != true {
		t.Errorf("expected success=true")
	}
	data, _ := envelope["data"].(map[string]any)
	if rc, ok := data["result_count"].(float64); !ok || rc < 1 {
		t.Errorf("result_count = %v, want >= 1", data["result_count"])
	}
}

func TestUserList_Pagination(t *testing.T) {
	callCount := 0
	_, cleanup := testUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		if callCount == 1 {
			_, _ = w.Write([]byte(`{"results":[{"object":"user","id":"u1","name":"Alice","type":"person"}],"has_more":true,"next_cursor":"cursor1"}`))
		} else {
			_, _ = w.Write([]byte(`{"results":[{"object":"user","id":"u2","name":"Bob","type":"person"}],"has_more":false}`))
		}
	})
	defer cleanup()

	var err error
	out := captureStdout(t, func() {
		_, _, err = runUserRoot(t, "user", "list", "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	data, _ := envelope["data"].(map[string]any)
	if rc, ok := data["result_count"].(float64); !ok || rc != 2 {
		t.Errorf("result_count = %v, want 2 (two pages merged)", data["result_count"])
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 API calls for pagination, got %d", callCount)
	}
}

func TestUserRetrieve_RequiresArg(t *testing.T) {
	_, _, err := runUserRoot(t, "user", "retrieve")
	if err == nil {
		t.Fatal("expected error when no user ID arg is given")
	}
}

func TestUserRetrieve_Success(t *testing.T) {
	_, cleanup := testUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"object":"user","id":"u1","name":"Bot"}`))
	})
	defer cleanup()

	var err error
	out := captureStdout(t, func() {
		_, _, err = runUserRoot(t, "user", "retrieve", "00000000000000000000000000000001", "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	if envelope["success"] != true {
		t.Errorf("expected success=true")
	}
}

func TestUserBot_Success(t *testing.T) {
	_, cleanup := testUserServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/me" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"object":"user","type":"bot","id":"bot-1"}`))
	})
	defer cleanup()

	var err error
	out := captureStdout(t, func() {
		_, _, err = runUserRoot(t, "user", "bot", "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	if envelope["success"] != true {
		t.Errorf("expected success=true")
	}
}
