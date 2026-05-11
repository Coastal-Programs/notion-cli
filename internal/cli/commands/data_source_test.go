package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

func TestRegisterDataSourceCommands(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterDataSourceCommands(root)

	dsCmd, _, err := root.Find([]string{"data-source"})
	if err != nil {
		t.Fatalf("data-source command not found: %v", err)
	}

	want := map[string]bool{
		"retrieve": false, "create": false, "update": false, "query": false,
		"templates": false, "properties": false,
	}
	for _, c := range dsCmd.Commands() {
		want[c.Name()] = true
	}
	for name, found := range want {
		if !found {
			t.Errorf("subcommand %q not registered", name)
		}
	}
}

func TestDataSourceRetrieve_HappyPath(t *testing.T) {
	const dsID = "11111111111111111111111111111111"
	var got int32

	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/data_sources/") {
			atomic.AddInt32(&got, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "data_source", "id": dsID,
			})
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterDataSourceCommands(root)
	root.SetArgs([]string{"data-source", "retrieve", dsID, "--json"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("retrieve failed: %v", err)
	}
	if atomic.LoadInt32(&got) != 1 {
		t.Errorf("DataSourceRetrieve calls = %d, want 1", got)
	}
}

func TestDataSourceCreate_HappyPath(t *testing.T) {
	const parentDB = "22222222222222222222222222222222"
	var receivedBody map[string]any

	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/data_sources") {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &receivedBody)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "data_source", "id": "ds-new",
			})
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterDataSourceCommands(root)
	root.SetArgs([]string{
		"data-source", "create",
		"--parent-database", parentDB,
		"--properties", `{"Name":{"title":{}}}`,
		"--title", "My DS",
		"--json",
	})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("create failed: %v", err)
	}

	parent, _ := receivedBody["parent"].(map[string]any)
	gotID, _ := parent["database_id"].(string)
	if parent["type"] != "database_id" || strings.ReplaceAll(gotID, "-", "") != parentDB {
		t.Errorf("unexpected parent: %#v", parent)
	}
	if _, ok := receivedBody["properties"].(map[string]any); !ok {
		t.Errorf("expected properties map in body, got: %#v", receivedBody["properties"])
	}
	if _, ok := receivedBody["title"].([]any); !ok {
		t.Errorf("expected title array in body, got: %#v", receivedBody["title"])
	}
}

func TestDataSourceUpdate_HappyPath(t *testing.T) {
	const dsID = "33333333333333333333333333333333"
	var receivedBody map[string]any

	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "PATCH" && strings.Contains(r.URL.Path, "/data_sources/") {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &receivedBody)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "data_source", "id": dsID,
			})
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterDataSourceCommands(root)
	root.SetArgs([]string{
		"data-source", "update", dsID,
		"--title", "Renamed",
		"--in-trash", "true",
		"--json",
	})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	if v, _ := receivedBody["in_trash"].(bool); !v {
		t.Errorf("expected in_trash=true in body, got: %#v", receivedBody["in_trash"])
	}
	if _, ok := receivedBody["title"].([]any); !ok {
		t.Errorf("expected title array in body, got: %#v", receivedBody["title"])
	}
}

func TestDataSourceQuery_HappyPath(t *testing.T) {
	const dsID = "44444444444444444444444444444444"
	var queriedPath string
	var receivedBody map[string]any

	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "POST" && strings.Contains(r.URL.Path, "/data_sources/") && strings.HasSuffix(r.URL.Path, "/query") {
			queriedPath = r.URL.Path
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &receivedBody)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "list", "results": []any{}, "has_more": false,
			})
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterDataSourceCommands(root)
	root.SetArgs([]string{
		"data-source", "query", dsID,
		"--page-size", "5",
		"--json",
	})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("query failed: %v", err)
	}

	if !strings.Contains(strings.ReplaceAll(queriedPath, "-", ""), dsID) {
		t.Errorf("queried path %q does not contain data source id %q", queriedPath, dsID)
	}
	if v, _ := receivedBody["page_size"].(float64); v != 5 {
		t.Errorf("expected page_size=5, got: %#v", receivedBody["page_size"])
	}
}

func TestDataSourceTemplates_HappyPath(t *testing.T) {
	const dsID = "55555555555555555555555555555555"
	var calledPath string

	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/data_sources/") && strings.HasSuffix(r.URL.Path, "/templates") {
			calledPath = r.URL.Path
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object":   "list",
				"results":  []any{map[string]any{"id": "tmpl-1", "object": "template"}},
				"has_more": false,
			})
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterDataSourceCommands(root)
	root.SetArgs([]string{"data-source", "templates", dsID, "--page-size", "10", "--json"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("templates failed: %v", err)
	}
	if !strings.Contains(strings.ReplaceAll(calledPath, "-", ""), dsID) {
		t.Errorf("called path %q does not contain data source id %q", calledPath, dsID)
	}
}

func TestDataSourcePropertiesUpdate_HappyPath(t *testing.T) {
	const dsID = "66666666666666666666666666666666"
	var receivedBody map[string]any

	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "PATCH" && strings.Contains(r.URL.Path, "/data_sources/") && strings.HasSuffix(r.URL.Path, "/properties") {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, &receivedBody)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "data_source", "id": dsID,
			})
			return
		}
		t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterDataSourceCommands(root)
	root.SetArgs([]string{
		"data-source", "properties", "update", dsID,
		"--schema", `{"Status":{"select":{}}}`,
		"--json",
	})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("properties update failed: %v", err)
	}

	props, ok := receivedBody["properties"].(map[string]any)
	if !ok {
		t.Errorf("expected properties map in request body, got: %#v", receivedBody)
	}
	if props["Status"] == nil {
		t.Errorf("expected Status property, got: %#v", props)
	}
}
