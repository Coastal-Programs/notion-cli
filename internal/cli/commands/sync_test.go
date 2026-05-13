package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// ---------------------------------------------------------------------------
// generateAliases pure-function tests
// ---------------------------------------------------------------------------

func TestGenerateAliases_Empty(t *testing.T) {
	aliases := generateAliases("")
	if aliases != nil {
		t.Errorf("generateAliases(\"\") = %v, want nil", aliases)
	}
}

func TestGenerateAliases_SingleWord(t *testing.T) {
	aliases := generateAliases("Tasks")
	if len(aliases) != 1 {
		t.Fatalf("got %d aliases, want 1", len(aliases))
	}
	if aliases[0] != "tasks" {
		t.Errorf("aliases[0] = %q, want %q", aliases[0], "tasks")
	}
}

func TestGenerateAliases_MultiWord(t *testing.T) {
	aliases := generateAliases("Project Tasks")
	if len(aliases) != 2 {
		t.Fatalf("got %d aliases, want 2 (lowercase + acronym)", len(aliases))
	}
	if aliases[0] != "project tasks" {
		t.Errorf("aliases[0] = %q, want %q", aliases[0], "project tasks")
	}
	if aliases[1] != "pt" {
		t.Errorf("aliases[1] = %q, want %q", aliases[1], "pt")
	}
}

func TestGenerateAliases_ThreeWords(t *testing.T) {
	aliases := generateAliases("My Action Items")
	if len(aliases) != 2 {
		t.Fatalf("got %d aliases, want 2", len(aliases))
	}
	if aliases[0] != "my action items" {
		t.Errorf("aliases[0] = %q, want %q", aliases[0], "my action items")
	}
	if aliases[1] != "mai" {
		t.Errorf("aliases[1] = %q, want %q", aliases[1], "mai")
	}
}

// ---------------------------------------------------------------------------
// sync command integration tests
// ---------------------------------------------------------------------------

func runSyncRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterSyncCommand(root)
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

func TestSync_FreshCacheSkips(t *testing.T) {
	// Write a fresh (non-stale) databases.json so sync is skipped.
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Setenv("NOTION_TOKEN", "secret_test_token")
	t.Cleanup(func() {
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})

	cacheDir := tmpDir + "/.notion-cli"
	_ = os.MkdirAll(cacheDir, 0o700)
	cacheData := map[string]any{
		"databases": []map[string]any{{"id": "db-1", "title": "Test", "last_edited": time.Now().Format(time.RFC3339)}},
		"last_sync": time.Now().Format(time.RFC3339),
	}
	raw, _ := json.Marshal(cacheData)
	_ = os.WriteFile(cacheDir+"/databases.json", raw, 0o600)

	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"results":[],"has_more":false}`))
	}))
	defer srv.Close()
	origBase := os.Getenv("NOTION_CLI_BASE_URL")
	_ = os.Setenv("NOTION_CLI_BASE_URL", srv.URL)
	t.Cleanup(func() {
		if origBase == "" {
			_ = os.Unsetenv("NOTION_CLI_BASE_URL")
		} else {
			_ = os.Setenv("NOTION_CLI_BASE_URL", origBase)
		}
	})

	_, _, err := runSyncRoot(t, "sync")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Fresh cache should not call the API.
	if callCount != 0 {
		t.Errorf("expected 0 API calls for fresh cache, got %d", callCount)
	}
}

func TestSync_Force(t *testing.T) {
	// --force triggers a real sync even with a fresh cache.
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Setenv("NOTION_TOKEN", "secret_test_token")
	t.Cleanup(func() {
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})

	var capturedBody map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results": []any{
				map[string]any{
					"object": "data_source",
					"id":     "db-1",
					"parent": map[string]any{
						"type":        "database_id",
						"database_id": "parent-db-1",
					},
					"title":            []any{map[string]any{"plain_text": "My DB"}},
					"last_edited_time": time.Now().Format(time.RFC3339),
				},
			},
			"has_more": false,
		})
	}))
	defer srv.Close()
	origBase := os.Getenv("NOTION_CLI_BASE_URL")
	_ = os.Setenv("NOTION_CLI_BASE_URL", srv.URL)
	t.Cleanup(func() {
		if origBase == "" {
			_ = os.Unsetenv("NOTION_CLI_BASE_URL")
		} else {
			_ = os.Setenv("NOTION_CLI_BASE_URL", origBase)
		}
	})

	_, _, err := runSyncRoot(t, "sync", "--force")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := capturedBody["filter"].(map[string]any)
	if !ok {
		t.Fatal("filter field missing from sync request")
	}
	if filter["value"] != "data_source" {
		t.Errorf("filter.value = %v, want data_source", filter["value"])
	}

	raw, err := os.ReadFile(tmpDir + "/.notion-cli/databases.json")
	if err != nil {
		t.Fatalf("cache file not written: %v", err)
	}
	var saved map[string]any
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatalf("cache file invalid JSON: %v", err)
	}
	if got := len(saved["databases"].([]any)); got != 1 {
		t.Fatalf("cached databases = %d, want 1", got)
	}
	if got := len(saved["data_sources"].([]any)); got != 1 {
		t.Fatalf("cached data sources = %d, want 1", got)
	}
}
