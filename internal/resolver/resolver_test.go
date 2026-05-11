package resolver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	cliErrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/internal/notion"
	"github.com/Coastal-Programs/notion-cli/v6/internal/retry"
)

func TestExtractID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			"hyphenated UUID passthrough",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"raw 32-char hex",
			"8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"uppercase hex",
			"8C4D6E5FA1B23C4D5E6F7A8B9C0D1E2F",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"Notion URL with page title",
			"https://www.notion.so/My-Page-Title-8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"Notion URL with query params",
			"https://notion.so/workspace/8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f?v=abc123",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"Notion URL with hash",
			"https://notion.so/page-8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f#section",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"input with whitespace",
			"  8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f  ",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"notion.site URL",
			"https://mysite.notion.site/Page-8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"empty input",
			"",
			"",
			true,
		},
		{
			"whitespace only",
			"   ",
			"",
			true,
		},
		{
			"invalid input",
			"not-an-id",
			"",
			true,
		},
		{
			"too short hex",
			"8c4d6e5f",
			"",
			true,
		},
		{
			"Notion URL without ID",
			"https://notion.so/just-a-page-title",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ExtractID(%q) expected error, got %q", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("ExtractID(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"32 hex chars",
			"8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
		},
		{
			"already formatted",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
		},
		{
			"uppercase input",
			"8C4D6E5FA1B23C4D5E6F7A8B9C0D1E2F",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
		},
		{
			"too short returns as-is (lowercased)",
			"abc123",
			"abc123",
		},
		{
			"not hex returns as-is (lowercased)",
			"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatID(tt.input)
			if got != tt.want {
				t.Errorf("FormatID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f", true},
		{"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f", true},
		{"8C4D6E5FA1B23C4D5E6F7A8B9C0D1E2F", true},
		{"  8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f  ", true},
		{"abc123", false},
		{"not-a-valid-id", false},
		{"", false},
		{"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", false},
		{"8c4d6e5f-a1b2-3c4d-5e6f", false}, // too short UUID
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsValidID(tt.input)
			if got != tt.want {
				t.Errorf("IsValidID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripHyphens(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f", "8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f"},
		{"no-hyphens-at-all", "nohyphensatall"},
		{"already-clean", "alreadyclean"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := StripHyphens(tt.input)
			if got != tt.want {
				t.Errorf("StripHyphens(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractID_AllZeros(t *testing.T) {
	id, err := ExtractID("00000000000000000000000000000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "00000000-0000-0000-0000-000000000000"
	if id != want {
		t.Errorf("got %q, want %q", id, want)
	}
}

func TestExtractID_MixedCaseUUID(t *testing.T) {
	id, err := ExtractID("8C4D6E5F-A1B2-3C4D-5E6F-7A8B9C0D1E2F")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f"
	if id != want {
		t.Errorf("got %q, want %q", id, want)
	}
}

// newTestClient returns a notion.Client pointed at the given test server URL.
func newTestClient(srvURL string) *notion.Client {
	return notion.NewClient("test-token",
		notion.WithBaseURL(srvURL),
		notion.WithRetryConfig(retry.RetryConfig{MaxRetries: 0}),
	)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func TestPrimaryDataSourceID_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"object": "database",
			"id":     "db-abc",
			"data_sources": []any{
				map[string]any{"id": "ds-abc"},
			},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	id, err := PrimaryDataSourceID(context.Background(), c, "db-abc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "ds-abc" {
		t.Errorf("id = %q, want %q", id, "ds-abc")
	}
}

func TestPrimaryDataSourceID_APIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 404, map[string]any{"code": "object_not_found", "message": "not found"})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := PrimaryDataSourceID(context.Background(), c, "db-missing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestPrimaryDataSourceID_NoDataSources(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"object":       "database",
			"id":           "db-empty",
			"data_sources": []any{},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := PrimaryDataSourceID(context.Background(), c, "db-empty")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *cliErrors.NotionCLIError
	if !isNotionCLIError(err, &cliErr) || cliErr.Code != cliErrors.CodeDataSourceNotFound {
		t.Errorf("expected CodeDataSourceNotFound, got %v", err)
	}
}

func TestPrimaryDataSourceID_MalformedEntry(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"object":       "database",
			"id":           "db-bad",
			"data_sources": []any{42},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := PrimaryDataSourceID(context.Background(), c, "db-bad")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *cliErrors.NotionCLIError
	if !isNotionCLIError(err, &cliErr) || cliErr.Code != cliErrors.CodeDataSourceNotFound {
		t.Errorf("expected CodeDataSourceNotFound, got %v", err)
	}
}

func TestPrimaryDataSourceID_MissingID(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"object":       "database",
			"id":           "db-noid",
			"data_sources": []any{map[string]any{}},
		})
	}))
	defer srv.Close()

	c := newTestClient(srv.URL)
	_, err := PrimaryDataSourceID(context.Background(), c, "db-noid")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	var cliErr *cliErrors.NotionCLIError
	if !isNotionCLIError(err, &cliErr) || cliErr.Code != cliErrors.CodeDataSourceNotFound {
		t.Errorf("expected CodeDataSourceNotFound, got %v", err)
	}
}

// isNotionCLIError asserts err is a *cliErrors.NotionCLIError and assigns it.
func isNotionCLIError(err error, out **cliErrors.NotionCLIError) bool {
	if e, ok := err.(*cliErrors.NotionCLIError); ok {
		*out = e
		return true
	}
	return false
}

func TestExtractID_DataSourceQueryParam(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"dataSource param extracts data source ID",
			"https://notion.so/workspace/page-8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f?v=abc&dataSource=aabbccdd00112233aabbccdd00112233",
			"aabbccdd-0011-2233-aabb-ccdd00112233",
		},
		{
			"dataSource param with hyphenated UUID value",
			"https://notion.so/ws/db-8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f?dataSource=aabbccdd-0011-2233-aabb-ccdd00112233",
			"aabbccdd-0011-2233-aabb-ccdd00112233",
		},
		{
			"no dataSource param falls back to path ID",
			"https://notion.so/workspace/8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractID(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("ExtractID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
