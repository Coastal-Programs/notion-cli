package commands

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

func runListRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterListCommand(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

func TestListCmd_NoCache(t *testing.T) {
	// Point HOME to a temp dir with no databases.json → WorkspaceNotSynced.
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	_, _, err := runListRoot(t, "list")
	if err == nil {
		t.Fatal("expected error when no cache exists")
	}
}

func TestListCmd_Success(t *testing.T) {
	// Write a pre-populated databases.json into a temp HOME.
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	cacheDir := filepath.Join(tmpDir, ".notion-cli")
	if err := os.MkdirAll(cacheDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	cacheData := map[string]any{
		"databases": []map[string]any{
			{
				"id":          "db-11111111111111111111111111111111",
				"title":       "My Test DB",
				"last_edited": time.Now().Format(time.RFC3339),
			},
		},
		"last_sync": time.Now().Format(time.RFC3339),
	}
	raw, _ := json.Marshal(cacheData)
	if err := os.WriteFile(filepath.Join(cacheDir, "databases.json"), raw, 0o600); err != nil {
		t.Fatalf("write cache: %v", err)
	}

	var err error
	out := captureStdout(t, func() {
		_, _, err = runListRoot(t, "list", "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	if envelope["success"] != true {
		t.Errorf("expected success=true")
	}
	data, _ := envelope["data"].(map[string]any)
	dbs, _ := data["databases"].([]any)
	if len(dbs) == 0 {
		t.Error("expected at least one database in list output")
	}
}
