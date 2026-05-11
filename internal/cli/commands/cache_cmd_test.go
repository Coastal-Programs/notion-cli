package commands

import (
	"bytes"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func runCacheRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterCacheCommands(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

func TestCacheCmdInfo_ReturnsOutput(t *testing.T) {
	var err error
	out := captureStdout(t, func() {
		_, _, err = runCacheRoot(t, "cache", "info", "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	if envelope["success"] != true {
		t.Errorf("expected success=true")
	}

	data, ok := envelope["data"].(map[string]any)
	if !ok {
		t.Fatalf("envelope.data not a map")
	}

	for _, key := range []string{"cache_enabled", "cache_max_size", "disk_cache_enabled"} {
		if _, found := data[key]; !found {
			t.Errorf("cache info missing key %q", key)
		}
	}
}

func TestCacheCmdInfo_WorkspaceCacheNotSynced(t *testing.T) {
	// Use a temp HOME with no .notion-cli directory.
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	var err error
	out := captureStdout(t, func() {
		_, _, err = runCacheRoot(t, "cache", "info", "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	data, _ := envelope["data"].(map[string]any)
	wc, ok := data["workspace_cache"].(map[string]any)
	if !ok {
		t.Fatalf("workspace_cache not a map: %T", data["workspace_cache"])
	}
	if wc["status"] != "not synced" {
		t.Errorf("workspace_cache.status = %v, want \"not synced\"", wc["status"])
	}
}
