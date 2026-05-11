package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

const testPageID = "11111111111111111111111111111111"

func TestRegisterMarkdownCommands(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterMarkdownCommands(root)

	mdCmd, _, err := root.Find([]string{"markdown"})
	if err != nil {
		t.Fatalf("markdown command not found: %v", err)
	}

	wantSubs := map[string]bool{"get": false, "set": false}
	for _, c := range mdCmd.Commands() {
		wantSubs[c.Name()] = true
	}
	for name, found := range wantSubs {
		if !found {
			t.Errorf("subcommand %q not registered", name)
		}
	}
}

func TestMarkdownGet_RawStdout(t *testing.T) {
	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/markdown") {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":            "page_markdown",
			"id":                testPageID,
			"markdown":          "# Hello\n\nWorld",
			"truncated":         false,
			"unknown_block_ids": []any{},
		})
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterMarkdownCommands(root)

	out := &bytes.Buffer{}
	root.SetArgs([]string{"markdown", "get", testPageID})
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	if !strings.Contains(out.String(), "# Hello") {
		t.Errorf("stdout missing markdown: %q", out.String())
	}
}

func TestMarkdownGet_FileFlag(t *testing.T) {
	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":   "page_markdown",
			"id":       testPageID,
			"markdown": "# From file",
		})
	})
	defer cleanup()
	_ = srv

	dir := t.TempDir()
	outPath := filepath.Join(dir, "out.md")

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterMarkdownCommands(root)
	root.SetArgs([]string{"markdown", "get", testPageID, "--file", outPath})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	b, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read out: %v", err)
	}
	if string(b) != "# From file" {
		t.Errorf("file content = %q", string(b))
	}
}

func TestMarkdownGet_JSONEnvelope(t *testing.T) {
	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":            "page_markdown",
			"id":                testPageID,
			"markdown":          "## Section",
			"truncated":         false,
			"unknown_block_ids": []any{},
		})
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterMarkdownCommands(root)
	out := &bytes.Buffer{}
	root.SetArgs([]string{"markdown", "get", testPageID, "--json"})
	root.SetOut(out)
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	var env map[string]any
	if err := json.Unmarshal(out.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal: %v\noutput: %s", err, out.String())
	}
	data, _ := env["data"].(map[string]any)
	if data == nil {
		t.Fatalf("envelope missing data: %v", env)
	}
	if data["content"] != "## Section" {
		t.Errorf("data.content = %v", data["content"])
	}
}

func TestMarkdownSet_Replace_ContentFlag(t *testing.T) {
	var sawType, sawNewStr string
	var calls int32

	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s", r.Method)
		}
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		sawType, _ = body["type"].(string)
		if rc, ok := body["replace_content"].(map[string]any); ok {
			sawNewStr, _ = rc["new_str"].(string)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":   "page_markdown",
			"id":       testPageID,
			"markdown": "hello",
		})
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterMarkdownCommands(root)
	root.SetArgs([]string{"markdown", "set", testPageID, "--content", "hello", "--json"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if atomic.LoadInt32(&calls) != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
	if sawType != "replace_content" {
		t.Errorf("type = %q, want replace_content", sawType)
	}
	if sawNewStr != "hello" {
		t.Errorf("new_str = %q", sawNewStr)
	}
}

func TestMarkdownSet_Append_FileFlag(t *testing.T) {
	dir := t.TempDir()
	mdPath := filepath.Join(dir, "in.md")
	if err := os.WriteFile(mdPath, []byte("appended"), 0o644); err != nil {
		t.Fatal(err)
	}

	var sawType, sawNewStr string

	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		sawType, _ = body["type"].(string)
		if ic, ok := body["insert_content"].(map[string]any); ok {
			sawNewStr, _ = ic["new_str"].(string)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":   "page_markdown",
			"id":       testPageID,
			"markdown": "existing\nappended",
		})
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterMarkdownCommands(root)
	root.SetArgs([]string{"markdown", "set", testPageID, "--file", mdPath, "--append", "--json"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if sawType != "insert_content" {
		t.Errorf("type = %q, want insert_content", sawType)
	}
	if sawNewStr != "appended" {
		t.Errorf("new_str = %q", sawNewStr)
	}
}

func TestMarkdownSet_Stdin(t *testing.T) {
	var sawNewStr string
	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if rc, ok := body["replace_content"].(map[string]any); ok {
			sawNewStr, _ = rc["new_str"].(string)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":   "page_markdown",
			"id":       testPageID,
			"markdown": "from stdin",
		})
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterMarkdownCommands(root)
	root.SetArgs([]string{"markdown", "set", testPageID, "--json"})
	root.SetIn(strings.NewReader("from stdin"))
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}
	if sawNewStr != "from stdin" {
		t.Errorf("new_str = %q", sawNewStr)
	}
}

func TestMarkdownSet_RequiresContent(t *testing.T) {
	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("API should not be called for empty content")
		w.WriteHeader(500)
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterMarkdownCommands(root)
	root.SetArgs([]string{"markdown", "set", testPageID, "--json"})
	root.SetIn(strings.NewReader(""))
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err == nil {
		t.Fatal("expected error for missing content")
	}
}
