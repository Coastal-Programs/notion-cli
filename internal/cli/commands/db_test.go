package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

// testDBServer creates a test HTTP server mimicking the Notion API for db
// endpoints, sets NOTION_TOKEN, and returns a cleanup func.
func testDBServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
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

func TestRegisterDBCommands(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterDBCommands(root)

	dbCmd, _, err := root.Find([]string{"db"})
	if err != nil {
		t.Fatalf("db command not found: %v", err)
	}

	wantSubs := map[string]bool{
		"query": false, "retrieve": false, "schema": false,
		"create": false, "update": false,
	}
	for _, c := range dbCmd.Commands() {
		wantSubs[c.Name()] = true
	}
	for name, found := range wantSubs {
		if !found {
			t.Errorf("subcommand %q not registered", name)
		}
	}
}

func TestDBQuery_HasDataSourceFlag(t *testing.T) {
	cmd := newDBQueryCmd()
	if cmd.Flags().Lookup("data-source") == nil {
		t.Error("expected --data-source flag on db query")
	}
}

// TestDBQuery_ResolvesPrimaryDataSource verifies that the new flow:
//  1. Calls DatabaseRetrieve to find data_sources[0].id.
//  2. Then calls /v1/data_sources/{id}/query (NOT the deprecated
//     /v1/databases/{id}/query endpoint).
func TestDBQuery_ResolvesPrimaryDataSource(t *testing.T) {
	const dbID = "11111111111111111111111111111111"
	const dsID = "22222222222222222222222222222222"

	var retrieveCalls, queryCalls int32
	var queriedPath string

	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "GET" && strings.Contains(r.URL.Path, "/databases/"):
			atomic.AddInt32(&retrieveCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "database",
				"id":     dbID,
				"data_sources": []any{
					map[string]any{"id": dsID, "name": "Default"},
				},
			})
		case r.Method == "POST" && strings.Contains(r.URL.Path, "/data_sources/"):
			atomic.AddInt32(&queryCalls, 1)
			queriedPath = r.URL.Path
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object":      "list",
				"results":     []any{},
				"has_more":    false,
				"next_cursor": nil,
			})
		case r.Method == "POST" && strings.HasSuffix(r.URL.Path, "/query") && strings.Contains(r.URL.Path, "/databases/"):
			t.Errorf("deprecated /v1/databases/{id}/query was hit: %s", r.URL.Path)
			w.WriteHeader(http.StatusGone)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusInternalServerError)
		}
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterDBCommands(root)
	root.SetArgs([]string{"db", "query", dbID, "--json"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("db query failed: %v", err)
	}

	if atomic.LoadInt32(&retrieveCalls) != 1 {
		t.Errorf("DatabaseRetrieve calls = %d, want 1", retrieveCalls)
	}
	if atomic.LoadInt32(&queryCalls) != 1 {
		t.Errorf("DataSourceQuery calls = %d, want 1", queryCalls)
	}
	if !strings.Contains(queriedPath, dsID) {
		t.Errorf("queried path %q does not contain data_source id %q", queriedPath, dsID)
	}
}

// TestDBQuery_DataSourceFlagSkipsResolution ensures that --data-source bypasses
// the DatabaseRetrieve round trip and goes straight to the data_source query.
func TestDBQuery_DataSourceFlagSkipsResolution(t *testing.T) {
	const dbID = "11111111111111111111111111111111"
	const dsID = "33333333333333333333333333333333"

	var retrieveCalls, queryCalls int32

	srv, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "GET" && strings.Contains(r.URL.Path, "/databases/"):
			atomic.AddInt32(&retrieveCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": dbID})
		case r.Method == "POST" && strings.Contains(r.URL.Path, "/data_sources/"):
			atomic.AddInt32(&queryCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "list", "results": []any{}, "has_more": false,
			})
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	defer cleanup()
	_ = srv

	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterDBCommands(root)
	root.SetArgs([]string{"db", "query", dbID, "--data-source", dsID, "--json"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})

	if err := root.Execute(); err != nil {
		t.Fatalf("db query failed: %v", err)
	}

	if got := atomic.LoadInt32(&retrieveCalls); got != 0 {
		t.Errorf("DatabaseRetrieve called %d times; expected 0 when --data-source supplied", got)
	}
	if got := atomic.LoadInt32(&queryCalls); got != 1 {
		t.Errorf("DataSourceQuery calls = %d, want 1", got)
	}
}

func TestDBQuery_RequiresArg(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterDBCommands(root)
	root.SetArgs([]string{"db", "query"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	if err := root.Execute(); err == nil {
		t.Error("expected error for missing database_id argument")
	}
}

// ---------------------------------------------------------------------------
// Pure-function tests (no server)
// ---------------------------------------------------------------------------

func TestFilterProperties_FiltersKeys(t *testing.T) {
	results := []any{
		map[string]any{
			"id": "p1",
			"properties": map[string]any{
				"Name":   map[string]any{"type": "title"},
				"Status": map[string]any{"type": "select"},
				"Due":    map[string]any{"type": "date"},
			},
		},
	}
	out := filterProperties(results, []string{"Name", "Status"})
	page := out[0].(map[string]any)
	props := page["properties"].(map[string]any)
	if _, ok := props["Due"]; ok {
		t.Error("Due should have been filtered out")
	}
	if _, ok := props["Name"]; !ok {
		t.Error("Name should remain")
	}
}

func TestFilterProperties_NonMapItem(t *testing.T) {
	// Non-map items should be skipped without panicking.
	results := []any{"not-a-map", nil}
	out := filterProperties(results, []string{"Name"})
	if len(out) != 2 {
		t.Errorf("got %d items, want 2 (non-map items pass through)", len(out))
	}
}

func TestExtractPlainText_Normal(t *testing.T) {
	v := []any{
		map[string]any{"plain_text": "Hello "},
		map[string]any{"plain_text": "World"},
	}
	if got := extractPlainText(v); got != "Hello World" {
		t.Errorf("got %q, want %q", got, "Hello World")
	}
}

func TestExtractPlainText_NonSlice(t *testing.T) {
	if got := extractPlainText("not a slice"); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestExtractPlainText_NonMapItem(t *testing.T) {
	v := []any{42, "str"}
	if got := extractPlainText(v); got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestFilterSchemaProperties_Basic(t *testing.T) {
	props := []map[string]any{
		{"name": "Name", "type": "title"},
		{"name": "Status", "type": "select"},
		{"name": "Due", "type": "date"},
	}
	out := filterSchemaProperties(props, []string{"name", "due"})
	if len(out) != 2 {
		t.Errorf("got %d props, want 2", len(out))
	}
}

func TestFilterSchemaProperties_Empty(t *testing.T) {
	props := []map[string]any{{"name": "Name", "type": "title"}}
	out := filterSchemaProperties(props, []string{})
	if len(out) != 0 {
		t.Errorf("got %d props, want 0", len(out))
	}
}

func TestExtractSchema_Basic(t *testing.T) {
	db := map[string]any{
		"id": "db-1",
		"title": []any{
			map[string]any{"plain_text": "My DB"},
		},
		"properties": map[string]any{
			"Name": map[string]any{"type": "title"},
			"Status": map[string]any{
				"type": "select",
				"select": map[string]any{
					"options": []any{
						map[string]any{"name": "Todo"},
						map[string]any{"name": "Done"},
					},
				},
			},
		},
	}
	schema := extractSchema(db)
	if schema["id"] != "db-1" {
		t.Errorf("id = %v", schema["id"])
	}
	if schema["title"] != "My DB" {
		t.Errorf("title = %v", schema["title"])
	}
	props, ok := schema["properties"].([]map[string]any)
	if !ok {
		t.Fatal("properties should be []map[string]any")
	}
	if len(props) != 2 {
		t.Errorf("got %d props, want 2", len(props))
	}
}

func TestExtractSchema_WithRelation(t *testing.T) {
	db := map[string]any{
		"id": "db-rel",
		"properties": map[string]any{
			"Tasks": map[string]any{
				"type": "relation",
				"relation": map[string]any{
					"database_id": "other-db",
					"type":        "single_property",
				},
			},
		},
	}
	schema := extractSchema(db)
	props := schema["properties"].([]map[string]any)
	if props[0]["relation_database_id"] != "other-db" {
		t.Errorf("relation_database_id = %v", props[0]["relation_database_id"])
	}
}

func TestExtractSchema_WithFormula(t *testing.T) {
	db := map[string]any{
		"id": "db-formula",
		"properties": map[string]any{
			"Calc": map[string]any{
				"type":    "formula",
				"formula": map[string]any{"expression": "prop(\"Name\")"},
			},
		},
	}
	schema := extractSchema(db)
	props := schema["properties"].([]map[string]any)
	if props[0]["expression"] != "prop(\"Name\")" {
		t.Errorf("expression = %v", props[0]["expression"])
	}
}

func TestExtractSchema_WithRollup(t *testing.T) {
	db := map[string]any{
		"id": "db-rollup",
		"properties": map[string]any{
			"Sum": map[string]any{
				"type": "rollup",
				"rollup": map[string]any{
					"rollup_property_name":   "Amount",
					"relation_property_name": "Tasks",
					"function":               "sum",
				},
			},
		},
	}
	schema := extractSchema(db)
	props := schema["properties"].([]map[string]any)
	if props[0]["rollup_function"] != "sum" {
		t.Errorf("rollup_function = %v", props[0]["rollup_function"])
	}
}

func TestExtractSchema_WithStatus(t *testing.T) {
	db := map[string]any{
		"id": "db-status",
		"properties": map[string]any{
			"State": map[string]any{
				"type": "status",
				"status": map[string]any{
					"options": []any{map[string]any{"name": "Not started"}},
					"groups":  []any{map[string]any{"name": "To-do"}},
				},
			},
		},
	}
	schema := extractSchema(db)
	props := schema["properties"].([]map[string]any)
	if groups, ok := props[0]["groups"].([]string); !ok || len(groups) == 0 {
		t.Errorf("groups = %v, expected non-empty", props[0]["groups"])
	}
}

func TestGenerateExample_AllTypes(t *testing.T) {
	types := []string{
		"title", "rich_text", "number", "checkbox", "select",
		"multi_select", "status", "date", "url", "email",
		"phone_number", "people", "files", "relation", "formula",
	}
	for _, propType := range types {
		t.Run(propType, func(t *testing.T) {
			prop := map[string]any{"type": propType, "name": propType}
			example := generateExample(prop)
			if example == nil {
				t.Errorf("generateExample(%q) returned nil", propType)
			}
		})
	}
}

func TestGenerateExample_SelectWithOptions(t *testing.T) {
	prop := map[string]any{"type": "select", "options": []string{"Active", "Done"}}
	example := generateExample(prop).(map[string]any)
	sel := example["select"].(map[string]any)
	if sel["name"] != "Active" {
		t.Errorf("select.name = %v, want Active", sel["name"])
	}
}

func TestGenerateExample_UnknownType(t *testing.T) {
	prop := map[string]any{"type": "created_by"}
	example := generateExample(prop).(map[string]any)
	if example["note"] == nil {
		t.Error("unknown type should produce note key")
	}
}

// ---------------------------------------------------------------------------
// db retrieve / schema / create / update command tests
// ---------------------------------------------------------------------------

func runDBRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterDBCommands(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

func TestDBRetrieve_Success(t *testing.T) {
	_, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasPrefix(r.URL.Path, "/databases/") {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "database", "id": "db-1"})
	})
	defer cleanup()

	_, _, err := runDBRoot(t, "db", "retrieve", "11111111111111111111111111111111")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDBRetrieve_RequiresArg(t *testing.T) {
	_, _, err := runDBRoot(t, "db", "retrieve")
	if err == nil {
		t.Fatal("expected error when no ID given")
	}
}

func TestDBSchema_Success(t *testing.T) {
	_, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "database",
			"id":     "db-1",
			"title":  []any{map[string]any{"plain_text": "My DB"}},
			"properties": map[string]any{
				"Name": map[string]any{"type": "title"},
			},
		})
	})
	defer cleanup()

	_, _, err := runDBRoot(t, "db", "schema", "11111111111111111111111111111111")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDBCreate_Success(t *testing.T) {
	_, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "database", "id": "new-db"})
	})
	defer cleanup()

	_, _, err := runDBRoot(t, "db", "create", "11111111111111111111111111111111", "--title", "Test DB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDBCreate_RequiresTitle(t *testing.T) {
	_, _, err := runDBRoot(t, "db", "create", "11111111111111111111111111111111")
	if err == nil {
		t.Fatal("expected error when --title not provided")
	}
}

func TestDBUpdate_Success(t *testing.T) {
	_, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "database", "id": "db-1"})
	})
	defer cleanup()

	_, _, err := runDBRoot(t, "db", "update", "11111111111111111111111111111111", "--title", "Renamed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// outputFormat pure-function tests
// ---------------------------------------------------------------------------

func TestOutputFormat_Flags(t *testing.T) {
	for _, tc := range []struct {
		flag string
		want string
	}{
		{"--compact-json", "compact-json"},
		{"--raw", "raw"},
		{"--csv", "csv"},
		{"--markdown", "markdown"},
		{"--pretty", "pretty"},
	} {
		t.Run(tc.flag, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			addOutputFlags(cmd)
			_ = cmd.ParseFlags([]string{tc.flag})
			got := string(outputFormat(cmd))
			if got != tc.want {
				t.Errorf("outputFormat(%s) = %q, want %q", tc.flag, got, tc.want)
			}
		})
	}
}

func TestDBQuery_PageSizeTooLarge(t *testing.T) {
	_, _, err := runDBRoot(t, "db", "query", "11111111111111111111111111111111", "--page-size", "200")
	if err == nil {
		t.Fatal("expected error for --page-size > 100")
	}
}

func TestDBQuery_InvalidSortDirection(t *testing.T) {
	const dbID = "11111111111111111111111111111111"
	const dsID = "22222222222222222222222222222222"
	_, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "database", "id": dbID,
				"data_sources": []any{map[string]any{"id": dsID}},
			})
		} else {
			_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "results": []any{}, "has_more": false})
		}
	})
	defer cleanup()

	_, _, err := runDBRoot(t, "db", "query", dbID, "--sort-property", "Name", "--sort-direction", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid --sort-direction")
	}
}

func TestDBQuery_WithFilter(t *testing.T) {
	const dbID = "11111111111111111111111111111111"
	const dsID = "22222222222222222222222222222222"
	var capturedBody map[string]any
	_, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == "GET" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "database", "id": dbID,
				"data_sources": []any{map[string]any{"id": dsID}},
			})
		} else {
			_ = json.NewDecoder(r.Body).Decode(&capturedBody)
			_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "results": []any{}, "has_more": false})
		}
	})
	defer cleanup()

	_, _, err := runDBRoot(t, "db", "query", dbID,
		"--filter", `{"property":"Status","select":{"equals":"Done"}}`,
		"--quiet")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedBody["filter"] == nil {
		t.Error("expected filter in request body")
	}
}

func TestDBSchema_WithPropertiesFlag(t *testing.T) {
	_, cleanup := testDBServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":     "database",
			"id":         "db-1",
			"properties": map[string]any{"Name": map[string]any{"type": "title"}, "Status": map[string]any{"type": "select", "select": map[string]any{}}},
		})
	})
	defer cleanup()

	_, _, err := runDBRoot(t, "db", "schema", "11111111111111111111111111111111", "--properties", "Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
