package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cobra"
)

// testSearchServer creates a test HTTP server and sets NOTION_TOKEN /
// NOTION_CLI_BASE_URL.
func testSearchServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
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

func runSearchRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterSearchCommand(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

// ---------------------------------------------------------------------------
// clientSideFilter pure-function tests
// ---------------------------------------------------------------------------

func makeItem(dbID, created, edited string) map[string]any {
	item := map[string]any{}
	if dbID != "" {
		item["parent"] = map[string]any{"database_id": dbID}
	}
	if created != "" {
		item["created_time"] = created
	}
	if edited != "" {
		item["last_edited_time"] = edited
	}
	return item
}

func TestClientSideFilter_DbFilter(t *testing.T) {
	items := []any{
		makeItem("db-a", "", ""),
		makeItem("db-b", "", ""),
		makeItem("db-a", "", ""),
	}
	got := clientSideFilter(items, "db-a", "", "", "", "")
	if len(got) != 2 {
		t.Errorf("got %d items, want 2", len(got))
	}
}

func TestClientSideFilter_CreatedAfter(t *testing.T) {
	old := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	recent := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	items := []any{
		makeItem("", old, ""),
		makeItem("", recent, ""),
	}
	got := clientSideFilter(items, "", "2024-01-01", "", "", "")
	if len(got) != 1 {
		t.Errorf("got %d items, want 1", len(got))
	}
}

func TestClientSideFilter_CreatedBefore(t *testing.T) {
	old := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	recent := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	items := []any{
		makeItem("", old, ""),
		makeItem("", recent, ""),
	}
	got := clientSideFilter(items, "", "", "2024-01-01", "", "")
	if len(got) != 1 {
		t.Errorf("got %d items, want 1", len(got))
	}
}

func TestClientSideFilter_EditedAfter(t *testing.T) {
	old := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	recent := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	items := []any{
		makeItem("", "", old),
		makeItem("", "", recent),
	}
	got := clientSideFilter(items, "", "", "", "2024-01-01", "")
	if len(got) != 1 {
		t.Errorf("got %d items, want 1", len(got))
	}
}

func TestClientSideFilter_EditedBefore(t *testing.T) {
	old := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	recent := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	items := []any{
		makeItem("", "", old),
		makeItem("", "", recent),
	}
	got := clientSideFilter(items, "", "", "", "", "2024-01-01")
	if len(got) != 1 {
		t.Errorf("got %d items, want 1", len(got))
	}
}

func TestClientSideFilter_MultipleFilters(t *testing.T) {
	created := time.Date(2024, 6, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	items := []any{
		makeItem("db-a", created, ""),          // passes both
		makeItem("db-b", created, ""),          // fails dbFilter
		makeItem("db-a", "2023-01-01T00:00:00Z", ""), // fails createdAfter
	}
	got := clientSideFilter(items, "db-a", "2024-01-01", "", "", "")
	if len(got) != 1 {
		t.Errorf("got %d items, want 1", len(got))
	}
}

func TestClientSideFilter_NonMapItem(t *testing.T) {
	items := []any{"not-a-map", 42, nil}
	got := clientSideFilter(items, "", "", "", "", "")
	// Non-map items are silently skipped.
	if len(got) != 0 {
		t.Errorf("got %d items, want 0 (non-map items skipped)", len(got))
	}
}

func TestClientSideFilter_EmptyFilters(t *testing.T) {
	items := []any{
		makeItem("db-a", "", ""),
		makeItem("db-b", "", ""),
	}
	got := clientSideFilter(items, "", "", "", "", "")
	if len(got) != 2 {
		t.Errorf("got %d items, want 2", len(got))
	}
}

func TestClientSideFilter_InvalidDateFormat(t *testing.T) {
	// If the item's created_time can't parse, the filter skips the date check.
	item := map[string]any{"created_time": "not-a-date"}
	items := []any{item}
	// Should not panic; item passes through when date parse fails.
	got := clientSideFilter(items, "", "2024-01-01", "", "", "")
	_ = got // result may or may not include the item; important: no crash
}

// ---------------------------------------------------------------------------
// Command-level tests
// ---------------------------------------------------------------------------

func TestSearchCmd_InvalidProperty(t *testing.T) {
	_, _, err := runSearchRoot(t, "search", "--property", "invalid")
	if err == nil {
		t.Fatal("expected error for invalid --property")
	}
}

func TestSearchCmd_PageSizeTooLarge(t *testing.T) {
	_, _, err := runSearchRoot(t, "search", "--page-size", "101")
	if err == nil {
		t.Fatal("expected error for --page-size > 100")
	}
}

func TestSearchCmd_InvalidDateFormat(t *testing.T) {
	_, _, err := runSearchRoot(t, "search", "--created-after", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date format")
	}
}

func TestSearchCmd_Success(t *testing.T) {
	_, cleanup := testSearchServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"results":[{"object":"page","id":"page-1"}],"has_more":false}`))
	})
	defer cleanup()

	var err error
	out := captureStdout(t, func() {
		_, _, err = runSearchRoot(t, "search", "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)
	if envelope["success"] != true {
		t.Errorf("expected success=true, got %v", envelope["success"])
	}
}

func TestSearchCmd_SortAscending(t *testing.T) {
	var capturedBody map[string]any
	_, cleanup := testSearchServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"results":[],"has_more":false}`))
	})
	defer cleanup()

	_, _, err := runSearchRoot(t, "search", "--sort-direction", "asc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sort, ok := capturedBody["sort"].(map[string]any)
	if !ok {
		t.Fatal("sort field missing from request body")
	}
	if sort["direction"] != "ascending" {
		t.Errorf("direction = %v, want ascending", sort["direction"])
	}
}

func TestSearchCmd_PropertyPageFilter(t *testing.T) {
	var capturedBody map[string]any
	_, cleanup := testSearchServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&capturedBody)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"results":[],"has_more":false}`))
	})
	defer cleanup()

	_, _, err := runSearchRoot(t, "search", "--property", "page")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	filter, ok := capturedBody["filter"].(map[string]any)
	if !ok {
		t.Fatal("filter field missing from request body")
	}
	if filter["value"] != "page" {
		t.Errorf("filter.value = %v, want page", filter["value"])
	}
	if filter["property"] != "object" {
		t.Errorf("filter.property = %v, want object", filter["property"])
	}
}

// Ensure strings is used.
var _ = strings.TrimSpace
