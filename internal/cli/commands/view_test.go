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

// testViewServer creates a test HTTP server, sets NOTION_TOKEN /
// NOTION_CLI_BASE_URL, and returns a cleanup func.
func testViewServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
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

// runViewRoot returns a root command with view subcommands registered and
// stdout/stderr captured into a single buffer for assertions.
func runViewRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterViewCommands(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

// ---------------------------------------------------------------------------
// Registration / validation tests (no network)
// ---------------------------------------------------------------------------

func TestViewSubcommands_Registered(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterViewCommands(root)
	for _, sub := range []string{"create", "list", "retrieve", "update", "delete", "query", "results", "delete-query"} {
		c, _, err := root.Find([]string{"view", sub})
		if err != nil || c == nil || c.Name() != sub {
			t.Errorf("subcommand %q not found: %v", sub, err)
		}
	}
}

func TestViewList_RequiresDataSourceOrDatabase(t *testing.T) {
	_, _, err := runViewRoot(t, "view", "list")
	if err == nil {
		t.Fatal("expected error when neither --data-source nor --database is set")
	}
	if !strings.Contains(err.Error(), "data-source") && !strings.Contains(err.Error(), "database") {
		t.Errorf("expected error mentioning --data-source/--database, got: %v", err)
	}
}

func TestViewList_DataSourceAndDatabaseMutuallyExclusive(t *testing.T) {
	_, _, err := runViewRoot(t, "view", "list",
		"--data-source", "5c6a28216bb14a7eb6e1c50111515c3d",
		"--database", "5c6a28216bb14a7eb6e1c50111515c3d",
	)
	if err == nil {
		t.Fatal("expected mutually-exclusive error for --data-source/--database")
	}
}

func TestViewRetrieve_RequiresArg(t *testing.T) {
	_, _, err := runViewRoot(t, "view", "retrieve")
	if err == nil {
		t.Fatal("expected error when view_id arg is missing")
	}
}

func TestViewRetrieve_RejectsBadID(t *testing.T) {
	_, _, err := runViewRoot(t, "view", "retrieve", "not-a-real-id")
	if err == nil {
		t.Fatal("expected error for invalid view ID")
	}
}

func TestViewQuery_RequiresArg(t *testing.T) {
	_, _, err := runViewRoot(t, "view", "query")
	if err == nil {
		t.Fatal("expected error when view_id arg is missing")
	}
}

func TestViewResults_RequiresTwoArgs(t *testing.T) {
	_, _, err := runViewRoot(t, "view", "results", "5c6a28216bb14a7eb6e1c50111515c3d")
	if err == nil {
		t.Fatal("expected error when query_id arg is missing")
	}
}

func TestViewDeleteQuery_RequiresTwoArgs(t *testing.T) {
	_, _, err := runViewRoot(t, "view", "delete-query", "5c6a28216bb14a7eb6e1c50111515c3d")
	if err == nil {
		t.Fatal("expected error when query_id arg is missing")
	}
}

func TestViewCreate_RequiresFlags(t *testing.T) {
	_, _, err := runViewRoot(t, "view", "create")
	if err == nil {
		t.Fatal("expected error when required flags are missing")
	}
}

func TestViewCreate_RejectsInvalidType(t *testing.T) {
	_, _, err := runViewRoot(t, "view", "create",
		"--data-source", "5c6a28216bb14a7eb6e1c50111515c3d",
		"--name", "My View",
		"--type", "invalid",
	)
	if err == nil {
		t.Fatal("expected error for invalid view type")
	}
	if !strings.Contains(err.Error(), "Invalid view type") {
		t.Errorf("expected error mentioning invalid view type, got: %v", err)
	}
}

func TestViewUpdate_RequiresAtLeastOneFlag(t *testing.T) {
	_, _, err := runViewRoot(t, "view", "update", "5c6a28216bb14a7eb6e1c50111515c3d")
	if err == nil {
		t.Fatal("expected error when no update fields provided")
	}
	if !strings.Contains(err.Error(), "--name") && !strings.Contains(err.Error(), "--filter") {
		t.Errorf("expected error mentioning update flags, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// HTTP roundtrip tests
// ---------------------------------------------------------------------------

func TestViewCreate_SendsExpectedBody(t *testing.T) {
	var captured map[string]any
	var hits int32

	srv, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/views" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
			return
		}
		atomic.AddInt32(&hits, 1)
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "view", "id": "v1"})
	})
	defer cleanup()
	_ = srv

	_, _, err := runViewRoot(t, "view", "create",
		"--data-source", testIDRaw,
		"--name", "My Table",
		"--type", "table",
		"--json",
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if atomic.LoadInt32(&hits) != 1 {
		t.Fatalf("expected 1 POST /views, got %d", hits)
	}
	if captured["data_source_id"] != testID {
		t.Errorf("expected data_source_id=%s, got %v", testID, captured["data_source_id"])
	}
	if captured["name"] != "My Table" {
		t.Errorf("expected name=My Table, got %v", captured["name"])
	}
	if captured["type"] != "table" {
		t.Errorf("expected type=table, got %v", captured["type"])
	}
}

func TestViewCreate_WithFilter(t *testing.T) {
	var captured map[string]any

	srv, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "view", "id": "v1"})
	})
	defer cleanup()
	_ = srv

	_, _, err := runViewRoot(t, "view", "create",
		"--data-source", testIDRaw,
		"--name", "Filtered",
		"--type", "board",
		"--filter", `{"property":"Status","select":{"equals":"Done"}}`,
		"--json",
	)
	if err != nil {
		t.Fatalf("create with filter failed: %v", err)
	}
	if captured["filter"] == nil {
		t.Errorf("expected filter in request body, got nil")
	}
}

func TestViewUpdate_SendsExpectedBody(t *testing.T) {
	var captured map[string]any
	var hits int32

	srv, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/views/"+testID {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
			return
		}
		atomic.AddInt32(&hits, 1)
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "view", "id": testID})
	})
	defer cleanup()
	_ = srv

	_, _, err := runViewRoot(t, "view", "update", testIDRaw,
		"--name", "Renamed View",
		"--json",
	)
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if atomic.LoadInt32(&hits) != 1 {
		t.Fatalf("expected 1 PATCH /views/%s, got %d", testID, hits)
	}
	if captured["name"] != "Renamed View" {
		t.Errorf("expected name=Renamed View, got %v", captured["name"])
	}
}

func TestViewDelete_Roundtrip(t *testing.T) {
	var hits int32

	srv, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/views/"+testID {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
			return
		}
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "view", "id": testID})
	})
	defer cleanup()
	_ = srv

	_, _, err := runViewRoot(t, "view", "delete", testIDRaw, "--json")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if atomic.LoadInt32(&hits) != 1 {
		t.Fatalf("expected 1 DELETE /views/%s, got %d", testID, hits)
	}
}

func TestViewQuery_OneShot_ReturnsQueryID(t *testing.T) {
	var hits int32

	srv, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/views/"+testID+"/queries" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
			return
		}
		atomic.AddInt32(&hits, 1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"query_id": "qid1",
			"results":  []any{map[string]any{"id": "p1"}},
			"has_more": false,
		})
	})
	defer cleanup()
	_ = srv

	_, _, err := runViewRoot(t, "view", "query", testIDRaw, "--json")
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if atomic.LoadInt32(&hits) != 1 {
		t.Fatalf("expected 1 POST /views/%s/queries, got %d", testID, hits)
	}
}

func TestViewQuery_All_PaginatesAndDeletesQuery(t *testing.T) {
	var calls int32
	var deleteCalled int32

	srv, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/views/"+testID+"/queries":
			atomic.AddInt32(&calls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"query_id":    "qid2",
				"results":     []any{map[string]any{"id": "p1"}},
				"has_more":    true,
				"next_cursor": "cur2",
			})

		case r.Method == "GET" && r.URL.Path == "/views/"+testID+"/queries/qid2":
			atomic.AddInt32(&calls, 1)
			if r.URL.Query().Get("start_cursor") != "cur2" {
				t.Errorf("expected start_cursor=cur2, got %q", r.URL.Query().Get("start_cursor"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"query_id": "qid2",
				"results":  []any{map[string]any{"id": "p2"}},
				"has_more": false,
			})

		case r.Method == "DELETE" && r.URL.Path == "/views/"+testID+"/queries/qid2":
			atomic.AddInt32(&deleteCalled, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{"object": "view_query", "id": "qid2"})

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
		}
	})
	defer cleanup()
	_ = srv

	_, _, err := runViewRoot(t, "view", "query", testIDRaw, "--all", "--json")
	if err != nil {
		t.Fatalf("query --all failed: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("expected 2 query calls (1 create + 1 results), got %d", calls)
	}
	if atomic.LoadInt32(&deleteCalled) != 1 {
		t.Errorf("expected 1 DELETE query call, got %d", deleteCalled)
	}
}

// ---------------------------------------------------------------------------
// view results / delete-query / list / retrieve direct tests
// ---------------------------------------------------------------------------

const testView2ID = "22222222222222222222222222222222"
const testQueryID = "33333333333333333333333333333333"

func TestViewResults_Success(t *testing.T) {
	_, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results":  []any{map[string]any{"id": "p1"}},
			"has_more": false,
		})
	})
	defer cleanup()

	_, _, err := runViewRoot(t, "view", "results", testView2ID, testQueryID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewDeleteQuery_Success(t *testing.T) {
	_, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "view_query", "id": testQueryID})
	})
	defer cleanup()

	_, _, err := runViewRoot(t, "view", "delete-query", testView2ID, testQueryID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewList_Success(t *testing.T) {
	_, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results":  []any{map[string]any{"id": "v1"}},
			"has_more": false,
		})
	})
	defer cleanup()

	_, _, err := runViewRoot(t, "view", "list", "--data-source", testView2ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewRetrieve_Success(t *testing.T) {
	_, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": testView2ID, "object": "view"})
	})
	defer cleanup()

	_, _, err := runViewRoot(t, "view", "retrieve", testView2ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ---------------------------------------------------------------------------
// loadJSONFlag pure-function tests
// ---------------------------------------------------------------------------

func TestLoadJSONFlag_InlineJSON(t *testing.T) {
	var dest map[string]any
	if err := loadJSONFlag(`{"key":"value"}`, &dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dest["key"] != "value" {
		t.Errorf("key = %v, want value", dest["key"])
	}
}

func TestLoadJSONFlag_InvalidJSON(t *testing.T) {
	var dest map[string]any
	if err := loadJSONFlag(`not json`, &dest); err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLoadJSONFlag_FromFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "flag-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.WriteString(`{"from":"file"}`)
	_ = tmpFile.Close()

	var dest map[string]any
	if err := loadJSONFlag("@"+tmpFile.Name(), &dest); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if dest["from"] != "file" {
		t.Errorf("from = %v, want file", dest["from"])
	}
}

func TestLoadJSONFlag_FileNotFound(t *testing.T) {
	var dest map[string]any
	if err := loadJSONFlag("@/nonexistent/file.json", &dest); err == nil {
		t.Fatal("expected error for missing file")
	}
}

// ---------------------------------------------------------------------------
// view update / list extra coverage
// ---------------------------------------------------------------------------

func TestViewUpdate_WithFilter(t *testing.T) {
	_, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": testView2ID, "object": "view"})
	})
	defer cleanup()

	_, _, err := runViewRoot(t, "view", "update", testView2ID, "--filter", `{"property":"Status"}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewUpdate_WithSorts(t *testing.T) {
	_, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": testView2ID, "object": "view"})
	})
	defer cleanup()

	_, _, err := runViewRoot(t, "view", "update", testView2ID, "--sorts", `[{"property":"Name","direction":"ascending"}]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewList_WithDatabase(t *testing.T) {
	_, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"results":  []any{map[string]any{"id": "v1"}},
			"has_more": false,
		})
	})
	defer cleanup()

	_, _, err := runViewRoot(t, "view", "list", "--database", testView2ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewCreate_WithDatabase(t *testing.T) {
	_, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "v-new", "object": "view"})
	})
	defer cleanup()

	_, _, err := runViewRoot(t, "view", "create",
		"--data-source", testView2ID,
		"--name", "My View",
		"--type", "table",
		"--database", testView2ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestViewCreate_WithFilterAndSorts(t *testing.T) {
	_, cleanup := testViewServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": "v-new", "object": "view"})
	})
	defer cleanup()

	_, _, err := runViewRoot(t, "view", "create",
		"--data-source", testView2ID,
		"--name", "Filtered View",
		"--type", "table",
		"--filter", `{"property":"Status"}`,
		"--sorts", `[{"property":"Name","direction":"ascending"}]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
