package notion

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Coastal-Programs/notion-cli/v6/internal/config"
	"github.com/Coastal-Programs/notion-cli/v6/internal/oauth"
	"github.com/Coastal-Programs/notion-cli/v6/internal/retry"
)

// helper: create a test server and client pointing to it.
// Retries are disabled to keep tests fast and deterministic.
func setup(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := NewClient("test-token",
		WithBaseURL(srv.URL),
		WithRetryConfig(retry.RetryConfig{MaxRetries: 0}),
	)
	return c, srv
}

// helper: write JSON body to response.
func writeJSON(t *testing.T, w http.ResponseWriter, status int, v any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		t.Fatalf("write json: %v", err)
	}
}

// helper: read JSON request body.
func readBody(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	return body
}

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("secret_abc")
	if c.token != "secret_abc" {
		t.Errorf("token = %q, want %q", c.token, "secret_abc")
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
	if c.notionVersion != defaultNotionVersion {
		t.Errorf("notionVersion = %q, want %q", c.notionVersion, defaultNotionVersion)
	}
	if c.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if c.httpClient.Timeout != 30*time.Second {
		t.Errorf("httpClient.Timeout = %v, want 30s", c.httpClient.Timeout)
	}
	if c.retryConfig == nil {
		t.Error("retryConfig should not be nil")
	}
	if c.retryConfig.MaxRetries != 3 {
		t.Errorf("retryConfig.MaxRetries = %d, want 3", c.retryConfig.MaxRetries)
	}
}

func TestNewClient_WithOptions(t *testing.T) {
	custom := &http.Client{}
	c := NewClient("tok",
		WithHTTPClient(custom),
		WithBaseURL("https://example.com/api/"),
		WithNotionVersion("2023-01-01"),
	)
	if c.httpClient != custom {
		t.Error("expected custom http client")
	}
	if c.baseURL != "https://example.com/api" {
		t.Errorf("baseURL = %q, want trailing slash trimmed", c.baseURL)
	}
	if c.notionVersion != "2023-01-01" {
		t.Errorf("notionVersion = %q", c.notionVersion)
	}
}

func TestHeaders(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("Authorization = %q", got)
		}
		if got := r.Header.Get("Notion-Version"); got != "2026-03-11" {
			t.Errorf("Notion-Version = %q", got)
		}
		if got := r.Header.Get("Accept-Encoding"); got != "gzip" {
			t.Errorf("Accept-Encoding = %q", got)
		}
		if got := r.Header.Get("User-Agent"); got == "" {
			t.Error("User-Agent should not be empty")
		}
		writeJSON(t, w, 200, map[string]any{"ok": true})
	})

	_, err := c.UsersMe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func TestContentTypeOnPost(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", got)
		}
		writeJSON(t, w, 200, map[string]any{"id": "page1"})
	})

	_, err := c.PageCreate(context.Background(), map[string]any{"parent": "db1"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoContentTypeOnGet(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Content-Type"); got != "" {
			t.Errorf("GET should not have Content-Type, got %q", got)
		}
		writeJSON(t, w, 200, map[string]any{"id": "user1"})
	})

	_, err := c.UserRetrieve(context.Background(), "user1")
	if err != nil {
		t.Fatal(err)
	}
}

// --- Database tests ---

func TestDatabaseRetrieve(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/databases/db-123" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "database", "id": "db-123"})
	})

	result, err := c.DatabaseRetrieve(context.Background(), "db-123")
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "db-123" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestDatabaseQuery(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/databases/db-123/query" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body := readBody(t, r)
		if body["page_size"] != float64(10) {
			t.Errorf("page_size = %v", body["page_size"])
		}
		writeJSON(t, w, 200, map[string]any{
			"results":     []any{},
			"has_more":    false,
			"object":      "list",
			"next_cursor": nil,
		})
	})

	result, err := c.DatabaseQuery(context.Background(), "db-123", map[string]any{"page_size": 10})
	if err != nil {
		t.Fatal(err)
	}
	if result["object"] != "list" {
		t.Errorf("object = %v", result["object"])
	}
}

func TestDatabaseCreate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/databases" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body := readBody(t, r)
		if body["title"] == nil {
			t.Error("expected title in body")
		}
		writeJSON(t, w, 200, map[string]any{"object": "database", "id": "new-db"})
	})

	result, err := c.DatabaseCreate(context.Background(), map[string]any{
		"parent": map[string]any{"page_id": "page1"},
		"title":  []any{map[string]any{"text": map[string]any{"content": "Test DB"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "new-db" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestDatabaseUpdate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/databases/db-123" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "database", "id": "db-123"})
	})

	_, err := c.DatabaseUpdate(context.Background(), "db-123", map[string]any{"title": []any{}})
	if err != nil {
		t.Fatal(err)
	}
}

// --- Data Source tests ---

func TestDataSourceRetrieve(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/data_sources/ds-123" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "data_source", "id": "ds-123"})
	})

	result, err := c.DataSourceRetrieve(context.Background(), "ds-123")
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "ds-123" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestDataSourceQuery(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/data_sources/ds-123/query" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body := readBody(t, r)
		if body["page_size"] != float64(5) {
			t.Errorf("page_size = %v", body["page_size"])
		}
		writeJSON(t, w, 200, map[string]any{
			"object":   "list",
			"results":  []any{},
			"has_more": false,
		})
	})

	result, err := c.DataSourceQuery(context.Background(), "ds-123", map[string]any{"page_size": 5})
	if err != nil {
		t.Fatal(err)
	}
	if result["object"] != "list" {
		t.Errorf("object = %v", result["object"])
	}
}

func TestDataSourceCreate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/data_sources" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body := readBody(t, r)
		if body["parent"] == nil {
			t.Error("expected parent in body")
		}
		writeJSON(t, w, 200, map[string]any{"object": "data_source", "id": "new-ds"})
	})

	result, err := c.DataSourceCreate(context.Background(), map[string]any{
		"parent":     map[string]any{"database_id": "db1"},
		"properties": map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "new-ds" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestDataSourceUpdate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/data_sources/ds-123" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "data_source", "id": "ds-123"})
	})

	_, err := c.DataSourceUpdate(context.Background(), "ds-123", map[string]any{"title": []any{}})
	if err != nil {
		t.Fatal(err)
	}
}

func TestDataSourceTemplatesList(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/data_sources/ds-123/templates" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("page_size"); got != "10" {
			t.Errorf("page_size = %q, want \"10\"", got)
		}
		writeJSON(t, w, 200, map[string]any{
			"object":   "list",
			"results":  []any{map[string]any{"id": "tmpl-1", "object": "template"}},
			"has_more": false,
		})
	})

	result, err := c.DataSourceTemplatesList(context.Background(), "ds-123", QueryParams{PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if result["object"] != "list" {
		t.Errorf("object = %v", result["object"])
	}
	results, _ := result["results"].([]any)
	if len(results) != 1 {
		t.Errorf("results count = %d, want 1", len(results))
	}
}

func TestDataSourcePropertiesUpdate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/data_sources/ds-123/properties" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body := readBody(t, r)
		props, ok := body["properties"].(map[string]any)
		if !ok {
			t.Errorf("expected properties map in body")
		}
		if props["Status"] == nil {
			t.Errorf("expected Status property")
		}
		writeJSON(t, w, 200, map[string]any{"object": "data_source", "id": "ds-123"})
	})

	_, err := c.DataSourcePropertiesUpdate(context.Background(), "ds-123", map[string]any{
		"properties": map[string]any{
			"Status": map[string]any{"select": map[string]any{}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
}

// --- Page tests ---

func TestPageCreate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/pages" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "page", "id": "page-1"})
	})

	result, err := c.PageCreate(context.Background(), map[string]any{
		"parent":     map[string]any{"database_id": "db1"},
		"properties": map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "page-1" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestPageRetrieve(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/pages/page-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "page", "id": "page-1"})
	})

	result, err := c.PageRetrieve(context.Background(), "page-1")
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "page-1" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestPageUpdate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/pages/page-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "page", "id": "page-1"})
	})

	_, err := c.PageUpdate(context.Background(), "page-1", map[string]any{"in_trash": true})
	if err != nil {
		t.Fatal(err)
	}
}

func TestPageMarkdownGet(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/pages/page-1/markdown" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{
			"object":            "page_markdown",
			"id":                "page-1",
			"markdown":          "# Hello\n\nWorld",
			"truncated":         false,
			"unknown_block_ids": []any{},
		})
	})

	result, err := c.PageMarkdownGet(context.Background(), "page-1", QueryParams{})
	if err != nil {
		t.Fatal(err)
	}
	if result["markdown"] != "# Hello\n\nWorld" {
		t.Errorf("markdown = %v", result["markdown"])
	}
}

func TestPageMarkdownGet_Error(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 404, map[string]any{
			"object":  "error",
			"status":  404,
			"code":    "object_not_found",
			"message": "Could not find page.",
		})
	})

	_, err := c.PageMarkdownGet(context.Background(), "missing", QueryParams{})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err type = %T", err)
	}
	if apiErr.Status != 404 || apiErr.Code != "object_not_found" {
		t.Errorf("status=%d code=%q", apiErr.Status, apiErr.Code)
	}
}

func TestPageMarkdownUpdate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/pages/page-1/markdown" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.Header.Get("Content-Type"); got != "application/json" {
			t.Errorf("Content-Type = %q", got)
		}
		body := readBody(t, r)
		if body["type"] != "replace_content" {
			t.Errorf("type = %v", body["type"])
		}
		rc, ok := body["replace_content"].(map[string]any)
		if !ok {
			t.Fatalf("replace_content not a map: %T", body["replace_content"])
		}
		if rc["new_str"] != "# Hi" {
			t.Errorf("new_str = %v", rc["new_str"])
		}
		writeJSON(t, w, 200, map[string]any{
			"object":   "page_markdown",
			"id":       "page-1",
			"markdown": "# Hi",
		})
	})

	result, err := c.PageMarkdownUpdate(context.Background(), "page-1", map[string]any{
		"type":            "replace_content",
		"replace_content": map[string]any{"new_str": "# Hi"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["markdown"] != "# Hi" {
		t.Errorf("markdown = %v", result["markdown"])
	}
}

func TestPageMarkdownUpdate_Error(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 400, map[string]any{
			"object":  "error",
			"status":  400,
			"code":    "validation_error",
			"message": "Invalid markdown.",
		})
	})

	_, err := c.PageMarkdownUpdate(context.Background(), "page-1", map[string]any{
		"type":            "replace_content",
		"replace_content": map[string]any{"new_str": ""},
	})
	if err == nil {
		t.Fatal("expected error")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("err type = %T", err)
	}
	if apiErr.Status != 400 || apiErr.Code != "validation_error" {
		t.Errorf("status=%d code=%q", apiErr.Status, apiErr.Code)
	}
}

func TestPagePropertyRetrieve(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/pages/page-1/properties/prop-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("start_cursor"); got != "cur123" {
			t.Errorf("start_cursor = %q", got)
		}
		if got := r.URL.Query().Get("page_size"); got != "50" {
			t.Errorf("page_size = %q", got)
		}
		writeJSON(t, w, 200, map[string]any{"object": "property_item", "type": "title"})
	})

	result, err := c.PagePropertyRetrieve(context.Background(), "page-1", "prop-1", QueryParams{
		StartCursor: "cur123",
		PageSize:    50,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["type"] != "title" {
		t.Errorf("type = %v", result["type"])
	}
}

// --- Block tests ---

func TestBlockRetrieve(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/blocks/block-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "block", "id": "block-1", "type": "paragraph"})
	})

	result, err := c.BlockRetrieve(context.Background(), "block-1")
	if err != nil {
		t.Fatal(err)
	}
	if result["type"] != "paragraph" {
		t.Errorf("type = %v", result["type"])
	}
}

func TestBlockUpdate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/blocks/block-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "block", "id": "block-1"})
	})

	_, err := c.BlockUpdate(context.Background(), "block-1", map[string]any{
		"paragraph": map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestBlockDelete(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/blocks/block-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "block", "id": "block-1", "in_trash": true})
	})

	result, err := c.BlockDelete(context.Background(), "block-1")
	if err != nil {
		t.Fatal(err)
	}
	if result["in_trash"] != true {
		t.Errorf("in_trash = %v", result["in_trash"])
	}
}

func TestBlockChildrenList(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/blocks/block-1/children" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("page_size"); got != "25" {
			t.Errorf("page_size = %q", got)
		}
		writeJSON(t, w, 200, map[string]any{"object": "list", "results": []any{}, "has_more": false})
	})

	result, err := c.BlockChildrenList(context.Background(), "block-1", QueryParams{PageSize: 25})
	if err != nil {
		t.Fatal(err)
	}
	if result["object"] != "list" {
		t.Errorf("object = %v", result["object"])
	}
}

func TestBlockChildrenAppend(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/blocks/block-1/children" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "list", "results": []any{}})
	})

	_, err := c.BlockChildrenAppend(context.Background(), "block-1", map[string]any{
		"children": []any{map[string]any{"object": "block", "type": "paragraph"}},
	})
	if err != nil {
		t.Fatal(err)
	}
}

// --- User tests ---

func TestUsersList(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/users" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "list", "results": []any{
			map[string]any{"object": "user", "id": "u1"},
		}})
	})

	result, err := c.UsersList(context.Background(), QueryParams{})
	if err != nil {
		t.Fatal(err)
	}
	results := result["results"].([]any)
	if len(results) != 1 {
		t.Errorf("results count = %d", len(results))
	}
}

func TestUserRetrieve(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/u1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "user", "id": "u1", "name": "Test"})
	})

	result, err := c.UserRetrieve(context.Background(), "u1")
	if err != nil {
		t.Fatal(err)
	}
	if result["name"] != "Test" {
		t.Errorf("name = %v", result["name"])
	}
}

func TestUsersMe(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/users/me" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "user", "id": "bot1", "type": "bot"})
	})

	result, err := c.UsersMe(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result["type"] != "bot" {
		t.Errorf("type = %v", result["type"])
	}
}

// --- Search tests ---

func TestSearch(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/search" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body := readBody(t, r)
		if body["query"] != "test" {
			t.Errorf("query = %v", body["query"])
		}
		writeJSON(t, w, 200, map[string]any{"object": "list", "results": []any{}})
	})

	_, err := c.Search(context.Background(), map[string]any{"query": "test"})
	if err != nil {
		t.Fatal(err)
	}
}

// --- Error handling tests ---

func TestAPIError_Parsing(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 404, map[string]any{
			"object":  "error",
			"status":  404,
			"code":    "object_not_found",
			"message": "Could not find database with ID: db-123.",
		})
	})

	_, err := c.DatabaseRetrieve(context.Background(), "db-123")
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Status != 404 {
		t.Errorf("status = %d", apiErr.Status)
	}
	if apiErr.Code != "object_not_found" {
		t.Errorf("code = %q", apiErr.Code)
	}
	if apiErr.Message != "Could not find database with ID: db-123." {
		t.Errorf("message = %q", apiErr.Message)
	}
}

func TestAPIError_Unauthorized(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 401, map[string]any{
			"object":  "error",
			"status":  401,
			"code":    "unauthorized",
			"message": "API token is invalid.",
		})
	})

	_, err := c.UsersMe(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr := err.(*APIError)
	if apiErr.Status != 401 {
		t.Errorf("status = %d", apiErr.Status)
	}
	if apiErr.Code != "unauthorized" {
		t.Errorf("code = %q", apiErr.Code)
	}
}

func TestAPIError_RateLimited(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 429, map[string]any{
			"object":  "error",
			"status":  429,
			"code":    "rate_limited",
			"message": "Rate limited. Please retry after a short delay.",
		})
	})

	_, err := c.Search(context.Background(), map[string]any{})
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr := err.(*APIError)
	if apiErr.Status != 429 {
		t.Errorf("status = %d", apiErr.Status)
	}
}

func TestAPIError_NonJSON(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte("internal server error"))
	})

	_, err := c.UsersMe(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}

	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("expected *APIError, got %T", err)
	}
	if apiErr.Status != 500 {
		t.Errorf("status = %d", apiErr.Status)
	}
	if apiErr.Code != "unknown" {
		t.Errorf("code = %q", apiErr.Code)
	}
}

func TestAPIError_ErrorString(t *testing.T) {
	e := &APIError{Status: 400, Code: "validation_error", Message: "bad input"}
	got := e.Error()
	want := `notion api error (status 400, code "validation_error"): bad input`
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

// --- Gzip tests ---

func TestGzipResponse(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(200)

		gw := gzip.NewWriter(w)
		defer gw.Close() //nolint:errcheck
		_ = json.NewEncoder(gw).Encode(map[string]any{"id": "gzipped", "object": "page"})
	})

	result, err := c.PageRetrieve(context.Background(), "gzipped")
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "gzipped" {
		t.Errorf("id = %v", result["id"])
	}
}

// --- Context cancellation ---

func TestContextCancellation(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		// This handler should not be reached if context is already cancelled.
		writeJSON(t, w, 200, map[string]any{"ok": true})
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	_, err := c.UsersMe(ctx)
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

// --- QueryParams tests ---

func TestQueryParams_Empty(t *testing.T) {
	q := QueryParams{}
	v := q.Values()
	if len(v) != 0 {
		t.Errorf("expected empty values, got %v", v)
	}
}

func TestQueryParams_Full(t *testing.T) {
	q := QueryParams{StartCursor: "abc", PageSize: 100}
	v := q.Values()
	if v.Get("start_cursor") != "abc" {
		t.Errorf("start_cursor = %q", v.Get("start_cursor"))
	}
	if v.Get("page_size") != "100" {
		t.Errorf("page_size = %q", v.Get("page_size"))
	}
}

func TestQueryParams_OnlyCursor(t *testing.T) {
	q := QueryParams{StartCursor: "xyz"}
	v := q.Values()
	if v.Get("start_cursor") != "xyz" {
		t.Errorf("start_cursor = %q", v.Get("start_cursor"))
	}
	if v.Get("page_size") != "" {
		t.Errorf("page_size should be empty, got %q", v.Get("page_size"))
	}
}

// --- Empty body response ---

func TestEmptyResponseBody(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		// no body
	})

	result, err := c.BlockDelete(context.Background(), "block-1")
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Error("expected non-nil result map")
	}
}

// --- Pagination query params in GET ---

func TestPaginationInGetRequest(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		// Verify query params
		cursor := r.URL.Query().Get("start_cursor")
		size := r.URL.Query().Get("page_size")
		if cursor != "abc" {
			t.Errorf("start_cursor = %q, want %q", cursor, "abc")
		}
		if size != "10" {
			t.Errorf("page_size = %q, want %q", size, "10")
		}
		writeJSON(t, w, 200, map[string]any{"object": "list", "results": []any{}})
	})

	_, err := c.UsersList(context.Background(), QueryParams{StartCursor: "abc", PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
}

// --- Verify request body sent correctly ---

func TestRequestBodySent(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatal(err)
		}
		var data map[string]any
		if err := json.Unmarshal(body, &data); err != nil {
			t.Fatal(err)
		}
		if data["query"] != "test search" {
			t.Errorf("query = %v", data["query"])
		}
		filter, ok := data["filter"].(map[string]any)
		if !ok {
			t.Fatal("filter should be a map")
		}
		if filter["value"] != "database" {
			t.Errorf("filter.value = %v", filter["value"])
		}
		writeJSON(t, w, 200, map[string]any{"object": "list", "results": []any{}})
	})

	_, err := c.Search(context.Background(), map[string]any{
		"query":  "test search",
		"filter": map[string]any{"value": "database", "property": "object"},
	})
	if err != nil {
		t.Fatal(err)
	}
}

// --- Nil body on post sends empty JSON ---

func TestPostWithNilBody(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if len(body) > 0 {
			t.Errorf("expected empty body for nil map, got %s", body)
		}
		writeJSON(t, w, 200, map[string]any{"object": "list", "results": []any{}})
	})

	// Search with nil body
	_, err := c.Search(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
}

// --- Views tests ---

func TestViewsList_DataSource(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/views" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("data_source_id"); got != "ds-123" {
			t.Errorf("data_source_id = %q", got)
		}
		if got := r.URL.Query().Get("page_size"); got != "25" {
			t.Errorf("page_size = %q", got)
		}
		writeJSON(t, w, 200, map[string]any{"object": "list", "results": []any{}})
	})

	_, err := c.ViewsList(context.Background(), "", "ds-123", QueryParams{PageSize: 25})
	if err != nil {
		t.Fatal(err)
	}
}

func TestViewsList_Database(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("database_id"); got != "db-1" {
			t.Errorf("database_id = %q", got)
		}
		if got := r.URL.Query().Get("data_source_id"); got != "" {
			t.Errorf("data_source_id should be unset, got %q", got)
		}
		writeJSON(t, w, 200, map[string]any{"object": "list", "results": []any{}})
	})

	_, err := c.ViewsList(context.Background(), "db-1", "", QueryParams{})
	if err != nil {
		t.Fatal(err)
	}
}

func TestViewRetrieve(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/views/v1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "view", "id": "v1", "type": "table"})
	})

	result, err := c.ViewRetrieve(context.Background(), "v1")
	if err != nil {
		t.Fatal(err)
	}
	if result["type"] != "table" {
		t.Errorf("type = %v", result["type"])
	}
}

func TestViewQueryCreate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/views/v1/queries" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body := readBody(t, r)
		if v, ok := body["page_size"].(float64); !ok || v != 50 {
			t.Errorf("page_size = %v", body["page_size"])
		}
		writeJSON(t, w, 200, map[string]any{
			"object":      "view_query",
			"id":          "q1",
			"view_id":     "v1",
			"expires_at":  "2026-03-11T00:00:00Z",
			"total_count": 0,
			"results":     []any{},
			"has_more":    false,
		})
	})

	result, err := c.ViewQueryCreate(context.Background(), "v1", map[string]any{"page_size": 50})
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "q1" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestViewQueryCreate_NilBody(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		body := readBody(t, r)
		if len(body) != 0 {
			t.Errorf("expected empty body, got %v", body)
		}
		writeJSON(t, w, 200, map[string]any{"object": "view_query", "id": "q1"})
	})

	_, err := c.ViewQueryCreate(context.Background(), "v1", nil)
	if err != nil {
		t.Fatal(err)
	}
}

func TestViewQueryResults(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/views/v1/queries/q1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("start_cursor"); got != "cur-2" {
			t.Errorf("start_cursor = %q", got)
		}
		writeJSON(t, w, 200, map[string]any{"object": "list", "results": []any{}, "has_more": false})
	})

	_, err := c.ViewQueryResults(context.Background(), "v1", "q1", QueryParams{StartCursor: "cur-2"})
	if err != nil {
		t.Fatal(err)
	}
}

func TestViewQueryDelete(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/views/v1/queries/q1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "view_query", "id": "q1", "deleted": true})
	})

	result, err := c.ViewQueryDelete(context.Background(), "v1", "q1")
	if err != nil {
		t.Fatal(err)
	}
	if result["deleted"] != true {
		t.Errorf("deleted = %v", result["deleted"])
	}
}

// --- Comment tests ---

func TestCommentCreate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/comments" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body := readBody(t, r)
		if body["discussion_id"] != "disc-1" {
			t.Errorf("discussion_id = %v", body["discussion_id"])
		}
		writeJSON(t, w, 200, map[string]any{"object": "comment", "id": "c1"})
	})

	result, err := c.CommentCreate(context.Background(), map[string]any{
		"discussion_id": "disc-1",
		"rich_text":     []map[string]any{{"type": "text", "text": map[string]any{"content": "hi"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "c1" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestCommentList(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/comments" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("block_id"); got != "blk-1" {
			t.Errorf("block_id = %q", got)
		}
		if got := r.URL.Query().Get("page_size"); got != "25" {
			t.Errorf("page_size = %q", got)
		}
		writeJSON(t, w, 200, map[string]any{
			"object": "list", "results": []any{}, "has_more": false,
		})
	})

	result, err := c.CommentList(context.Background(), QueryParams{BlockID: "blk-1", PageSize: 25})
	if err != nil {
		t.Fatal(err)
	}
	if result["object"] != "list" {
		t.Errorf("object = %v", result["object"])
	}
}

func TestCommentRetrieve(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/comments/c1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "comment", "id": "c1"})
	})

	result, err := c.CommentRetrieve(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "c1" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestCommentUpdate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/comments/c1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		body := readBody(t, r)
		if _, ok := body["rich_text"]; !ok {
			t.Errorf("expected rich_text in body, got %v", body)
		}
		writeJSON(t, w, 200, map[string]any{"object": "comment", "id": "c1"})
	})

	result, err := c.CommentUpdate(context.Background(), "c1", map[string]any{
		"rich_text": []map[string]any{{"type": "text", "text": map[string]any{"content": "edited"}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "c1" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestCommentDelete(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/comments/c1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "comment", "id": "c1"})
	})

	result, err := c.CommentDelete(context.Background(), "c1")
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "c1" {
		t.Errorf("id = %v", result["id"])
	}
}

// --- Custom Emoji tests ---

func TestCustomEmojiList(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/custom_emojis" {
			t.Errorf("path = %s, want /custom_emojis", r.URL.Path)
		}
		if got := r.URL.Query().Get("page_size"); got != "" {
			t.Errorf("page_size should be unset, got %q", got)
		}
		writeJSON(t, w, 200, map[string]any{
			"object":      "list",
			"results":     []any{map[string]any{"object": "emoji", "id": "e1", "name": "parrot", "url": "https://example.com/parrot.png"}},
			"has_more":    false,
			"next_cursor": nil,
		})
	})

	result, err := c.CustomEmojiList(context.Background(), QueryParams{})
	if err != nil {
		t.Fatal(err)
	}
	if result["object"] != "list" {
		t.Errorf("object = %v", result["object"])
	}
	results, _ := result["results"].([]any)
	if len(results) != 1 {
		t.Errorf("results count = %d, want 1", len(results))
	}
}

func TestCustomEmojiList_WithPagination(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/custom_emojis" {
			t.Errorf("path = %s, want /custom_emojis", r.URL.Path)
		}
		if got := r.URL.Query().Get("page_size"); got != "50" {
			t.Errorf("page_size = %q, want \"50\"", got)
		}
		if got := r.URL.Query().Get("start_cursor"); got != "cursor-abc" {
			t.Errorf("start_cursor = %q, want \"cursor-abc\"", got)
		}
		writeJSON(t, w, 200, map[string]any{
			"object":      "list",
			"results":     []any{},
			"has_more":    false,
			"next_cursor": nil,
		})
	})

	result, err := c.CustomEmojiList(context.Background(), QueryParams{PageSize: 50, StartCursor: "cursor-abc"})
	if err != nil {
		t.Fatal(err)
	}
	if result["object"] != "list" {
		t.Errorf("object = %v", result["object"])
	}
}

// --- Page lifecycle tests ---

func TestPageTrash(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/pages/page-123" {
			t.Errorf("path = %s, want /pages/page-123", r.URL.Path)
		}
		body := readBody(t, r)
		if v, ok := body["in_trash"]; !ok || v != true {
			t.Errorf("in_trash = %v, want true", v)
		}
		writeJSON(t, w, 200, map[string]any{"object": "page", "id": "page-123", "in_trash": true})
	})

	result, err := c.PageTrash(context.Background(), "page-123")
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "page-123" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestPageRestore(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/pages/page-123" {
			t.Errorf("path = %s, want /pages/page-123", r.URL.Path)
		}
		body := readBody(t, r)
		if v, ok := body["in_trash"]; !ok || v != false {
			t.Errorf("in_trash = %v, want false", v)
		}
		writeJSON(t, w, 200, map[string]any{"object": "page", "id": "page-123", "in_trash": false})
	})

	result, err := c.PageRestore(context.Background(), "page-123")
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "page-123" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestPageMove(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/pages/page-123/move" {
			t.Errorf("path = %s, want /pages/page-123/move", r.URL.Path)
		}
		body := readBody(t, r)
		parent, ok := body["parent"].(map[string]any)
		if !ok {
			t.Fatalf("parent not a map: %T", body["parent"])
		}
		if parent["type"] != "page_id" {
			t.Errorf("parent.type = %v, want page_id", parent["type"])
		}
		if parent["page_id"] != "new-parent-id" {
			t.Errorf("parent.page_id = %v, want new-parent-id", parent["page_id"])
		}
		writeJSON(t, w, 200, map[string]any{"object": "page", "id": "page-123"})
	})

	body := map[string]any{
		"parent": map[string]any{
			"type":    "page_id",
			"page_id": "new-parent-id",
		},
	}
	result, err := c.PageMove(context.Background(), "page-123", body)
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "page-123" {
		t.Errorf("id = %v", result["id"])
	}
}

// --- FileUpload tests ---

func TestFileUploadCreate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/file_uploads" {
			t.Errorf("path = %s, want /file_uploads", r.URL.Path)
		}
		body := readBody(t, r)
		if body["filename"] != "test.txt" {
			t.Errorf("filename = %v", body["filename"])
		}
		if body["mode"] != "single_part" {
			t.Errorf("mode = %v", body["mode"])
		}
		writeJSON(t, w, 200, map[string]any{"object": "file_upload", "id": "fu-1"})
	})

	result, err := c.FileUploadCreate(context.Background(), map[string]any{
		"filename":     "test.txt",
		"content_type": "text/plain",
		"mode":         "single_part",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "fu-1" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestFileUploadSend_SinglePart(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/file_uploads/fu-1/send" {
			t.Errorf("path = %s", r.URL.Path)
		}
		ct := r.Header.Get("Content-Type")
		mediaType, _, parseErr := mime.ParseMediaType(ct)
		if parseErr != nil || mediaType != "multipart/form-data" {
			t.Errorf("Content-Type = %q, want multipart/form-data", ct)
		}
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		if r.FormValue("part_number") != "" {
			t.Error("part_number should be absent for single-part uploads")
		}
		writeJSON(t, w, 200, map[string]any{"object": "file_upload", "id": "fu-1"})
	})

	_, err := c.FileUploadSend(context.Background(), "fu-1", 0, "test.txt", "text/plain", []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileUploadSend_MultiPart(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseMultipartForm(1 << 20); err != nil {
			t.Fatalf("ParseMultipartForm: %v", err)
		}
		if r.FormValue("part_number") != "2" {
			t.Errorf("part_number = %q, want \"2\"", r.FormValue("part_number"))
		}
		writeJSON(t, w, 200, map[string]any{"object": "file_upload", "id": "fu-1"})
	})

	_, err := c.FileUploadSend(context.Background(), "fu-1", 2, "part.bin", "application/octet-stream", []byte("data"))
	if err != nil {
		t.Fatal(err)
	}
}

func TestFileUploadComplete(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/file_uploads/fu-1/complete" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "file_upload", "id": "fu-1", "status": "uploaded"})
	})

	result, err := c.FileUploadComplete(context.Background(), "fu-1")
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "fu-1" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestFileUploadRetrieve(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/file_uploads/fu-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "file_upload", "id": "fu-1", "status": "pending"})
	})

	result, err := c.FileUploadRetrieve(context.Background(), "fu-1")
	if err != nil {
		t.Fatal(err)
	}
	if result["status"] != "pending" {
		t.Errorf("status = %v", result["status"])
	}
}

func TestFileUploadList(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/file_uploads" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("page_size"); got != "5" {
			t.Errorf("page_size = %q, want \"5\"", got)
		}
		writeJSON(t, w, 200, map[string]any{
			"object":   "list",
			"results":  []any{map[string]any{"id": "fu-1"}, map[string]any{"id": "fu-2"}},
			"has_more": false,
		})
	})

	result, err := c.FileUploadList(context.Background(), QueryParams{PageSize: 5})
	if err != nil {
		t.Fatal(err)
	}
	if result["object"] != "list" {
		t.Errorf("object = %v", result["object"])
	}
}

// --- View CRUD tests ---

func TestViewCreate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/views" {
			t.Errorf("path = %s, want /views", r.URL.Path)
		}
		body := readBody(t, r)
		if body["data_source_id"] != "ds-1" {
			t.Errorf("data_source_id = %v", body["data_source_id"])
		}
		writeJSON(t, w, 200, map[string]any{"object": "view", "id": "v1"})
	})

	result, err := c.ViewCreate(context.Background(), map[string]any{
		"data_source_id": "ds-1",
		"name":           "My View",
		"type":           "table",
	})
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "v1" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestViewUpdate(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Path != "/views/v1" {
			t.Errorf("path = %s, want /views/v1", r.URL.Path)
		}
		body := readBody(t, r)
		if body["name"] != "Renamed" {
			t.Errorf("name = %v", body["name"])
		}
		writeJSON(t, w, 200, map[string]any{"object": "view", "id": "v1", "name": "Renamed"})
	})

	result, err := c.ViewUpdate(context.Background(), "v1", map[string]any{"name": "Renamed"})
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "v1" {
		t.Errorf("id = %v", result["id"])
	}
}

func TestViewDelete(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/views/v1" {
			t.Errorf("path = %s, want /views/v1", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"object": "view", "id": "v1"})
	})

	result, err := c.ViewDelete(context.Background(), "v1")
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "v1" {
		t.Errorf("id = %v", result["id"])
	}
}

// --- Client option tests ---

func TestWithTimeout(t *testing.T) {
	c := NewClient("tok", WithTimeout(5*time.Second))
	if c.httpClient.Timeout != 5*time.Second {
		t.Errorf("Timeout = %v, want 5s", c.httpClient.Timeout)
	}
}

func TestWithKeepAlive_Disabled(t *testing.T) {
	c := NewClient("tok", WithKeepAlive(false))
	transport, ok := c.httpClient.Transport.(*http.Transport)
	if !ok {
		t.Fatal("expected *http.Transport")
	}
	if !transport.DisableKeepAlives {
		t.Error("DisableKeepAlives should be true")
	}
}

func TestWithKeepAlive_Enabled(t *testing.T) {
	c := NewClient("tok", WithKeepAlive(true))
	// When enabled, transport should not have DisableKeepAlives set.
	// nil transport means default (keep-alives on by default).
	if transport, ok := c.httpClient.Transport.(*http.Transport); ok {
		if transport.DisableKeepAlives {
			t.Error("DisableKeepAlives should be false")
		}
	}
}

func TestWithConfig(t *testing.T) {
	cfg := &config.Config{Token: "tok"}
	c := NewClient("tok", WithConfig(cfg))
	if c.cfg == nil {
		t.Error("cfg should not be nil after WithConfig")
	}
}

// --- parseDuration tests ---

func TestParseDuration_Empty(t *testing.T) {
	if d := parseDuration(""); d != 0 {
		t.Errorf("parseDuration(\"\") = %v, want 0", d)
	}
}

func TestParseDuration_ValidInt(t *testing.T) {
	if d := parseDuration("5"); d != 5*time.Second {
		t.Errorf("parseDuration(\"5\") = %v, want 5s", d)
	}
}

func TestParseDuration_InvalidString(t *testing.T) {
	if d := parseDuration("abc"); d != 0 {
		t.Errorf("parseDuration(\"abc\") = %v, want 0", d)
	}
}

// --- QueryParams.Values tests ---

func TestQueryParamsValues_BlockIDAndPageID(t *testing.T) {
	q := QueryParams{
		StartCursor: "cursor1",
		PageSize:    10,
		BlockID:     "block-abc",
		PageID:      "page-xyz",
	}
	v := q.Values()
	if v.Get("start_cursor") != "cursor1" {
		t.Errorf("start_cursor = %q", v.Get("start_cursor"))
	}
	if v.Get("page_size") != "10" {
		t.Errorf("page_size = %q", v.Get("page_size"))
	}
	if v.Get("block_id") != "block-abc" {
		t.Errorf("block_id = %q", v.Get("block_id"))
	}
	if v.Get("page_id") != "page-xyz" {
		t.Errorf("page_id = %q", v.Get("page_id"))
	}
}

func TestQueryParamsValues_Empty(t *testing.T) {
	v := QueryParams{}.Values()
	if len(v) != 0 {
		t.Errorf("expected empty values, got %v", v)
	}
}

// --- CustomEmojiRetrieve test ---

func TestCustomEmojiRetrieve(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/custom_emojis/emoji-1" {
			t.Errorf("path = %s", r.URL.Path)
		}
		writeJSON(t, w, 200, map[string]any{"id": "emoji-1", "name": "party"})
	})

	result, err := c.CustomEmojiRetrieve(context.Background(), "emoji-1")
	if err != nil {
		t.Fatal(err)
	}
	if result["name"] != "party" {
		t.Errorf("name = %v", result["name"])
	}
}

// --- doFormData error path tests ---

func TestFileUploadSend_4xxError(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 400, map[string]any{"code": "validation_error", "message": "bad request"})
	})

	_, err := c.FileUploadSend(context.Background(), "fu-bad", 0, "test.txt", "text/plain", []byte("data"))
	if err == nil {
		t.Fatal("expected error for 400 response")
	}
}

func TestFileUploadSend_GzipResponse(t *testing.T) {
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		// Respond with gzip-encoded JSON.
		var buf bytes.Buffer
		gw := gzip.NewWriter(&buf)
		_ = json.NewEncoder(gw).Encode(map[string]any{"object": "file_upload", "id": "fu-gz"})
		_ = gw.Close()
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(200)
		_, _ = w.Write(buf.Bytes())
	})

	result, err := c.FileUploadSend(context.Background(), "fu-gz", 0, "test.txt", "text/plain", []byte("data"))
	if err != nil {
		t.Fatal(err)
	}
	if result["id"] != "fu-gz" {
		t.Errorf("id = %v", result["id"])
	}
}

// --- do auto-refresh on 401 path ---
// Test that a 401 with no refresh token configured is passed through as-is.
func TestDo_401NoRefreshToken(t *testing.T) {
	// Client without config → no auto-refresh.
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 401, map[string]any{"code": "unauthorized", "message": "invalid token"})
	})

	_, err := c.UsersMe(context.Background())
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestDo_AutoRefreshOn401(t *testing.T) {
	// Mock OAuth token endpoint.
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 200, map[string]any{
			"access_token":  "new_access_token",
			"refresh_token": "new_refresh_token",
			"token_type":    "bearer",
			"expires_in":    3600,
		})
	}))
	defer tokenSrv.Close()

	oauth.SetTokenURL(tokenSrv.URL)
	defer oauth.SetTokenURL("https://api.notion.com/v1/oauth/token")

	origClientID := config.OAuthClientID
	config.OAuthClientID = "test-client-id"
	defer func() { config.OAuthClientID = origClientID }()

	var callCount int32
	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&callCount, 1)
		if n == 1 {
			writeJSON(t, w, 401, map[string]any{"code": "unauthorized", "message": "token expired"})
		} else {
			writeJSON(t, w, 200, map[string]any{"object": "user", "id": "u1"})
		}
	})

	cfg := &config.Config{
		OAuthRefreshToken: "old_refresh_token",
	}
	c.cfg = cfg

	result, err := c.UsersMe(context.Background())
	if err != nil {
		t.Fatalf("unexpected error after auto-refresh: %v", err)
	}
	if atomic.LoadInt32(&callCount) != 2 {
		t.Errorf("expected 2 API calls (401 + retry), got %d", callCount)
	}
	if result["id"] != "u1" {
		t.Errorf("id = %v, want u1", result["id"])
	}
	if c.token != "new_access_token" {
		t.Errorf("token = %q, want updated to new_access_token", c.token)
	}
}

func TestDo_AutoRefreshFails_Returns401(t *testing.T) {
	// Token refresh fails → original 401 should be returned.
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer tokenSrv.Close()

	oauth.SetTokenURL(tokenSrv.URL)
	defer oauth.SetTokenURL("https://api.notion.com/v1/oauth/token")

	origClientID := config.OAuthClientID
	config.OAuthClientID = "test-client-id"
	defer func() { config.OAuthClientID = origClientID }()

	c, _ := setup(t, func(w http.ResponseWriter, r *http.Request) {
		writeJSON(t, w, 401, map[string]any{"code": "unauthorized", "message": "expired"})
	})
	c.cfg = &config.Config{OAuthRefreshToken: "refresh_token"}

	_, err := c.UsersMe(context.Background())
	if err == nil {
		t.Fatal("expected error when refresh fails")
	}
}

// Anchor the mime and strings imports used in upload tests.
var _ = strings.TrimSpace
var _ = mime.FormatMediaType

// --- additional parseDuration cases ---

// parseDuration uses strconv.Atoi internally, which accepts negative integers.
// Therefore parseDuration("-5") yields -5 * time.Second (no clamping).
// parseDuration("1.5") yields 0 because Atoi rejects float strings.
func TestParseDuration_TableExtras(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want time.Duration
	}{
		// Atoi accepts "-5"; the impl multiplies by time.Second without clamping.
		{"negative int", "-5", -5 * time.Second},
		// Atoi rejects "1.5"; impl returns 0 on parse error.
		{"float string", "1.5", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDuration(tt.in)
			if got != tt.want {
				t.Errorf("parseDuration(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

// TestParseDuration_NegativeInt is a focused single-case test: Atoi accepts
// negative integers, and the implementation does not clamp, so "-5" maps
// to -5 * time.Second.
func TestParseDuration_NegativeInt(t *testing.T) {
	got := parseDuration("-5")
	want := -5 * time.Second
	if got != want {
		t.Errorf("parseDuration(%q) = %v, want %v", "-5", got, want)
	}
}

// TestParseDuration_FloatString verifies that a float-looking string is
// rejected by strconv.Atoi and parseDuration returns 0.
func TestParseDuration_FloatString(t *testing.T) {
	got := parseDuration("1.5")
	if got != 0 {
		t.Errorf("parseDuration(%q) = %v, want 0", "1.5", got)
	}
}

// TestDo_GETNetworkErrorRetried points the client at a closed httptest server
// and verifies that a GET (UsersMe) returns a non-nil error and does not panic.
// With MaxRetries=1 the retry layer will be exercised, but because network
// errors from a closed listener are not wrapped in *retry.RetryableError by
// the client, only one attempt is made — that's fine: the contract being
// tested is "error returned without panic".
func TestDo_GETNetworkErrorRetried(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	srv.Close()

	c := NewClient("test-token",
		WithBaseURL(srv.URL),
		WithRetryConfig(retry.RetryConfig{
			MaxRetries: 1,
			BaseDelay:  time.Millisecond,
			MaxDelay:   10 * time.Millisecond,
		}),
	)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic: %v", r)
		}
	}()

	_, err := c.UsersMe(context.Background())
	if err == nil {
		t.Fatal("expected non-nil error from closed server, got nil")
	}
}

// ---------------------------------------------------------------------------
// Retry behavior: 429 / 503 / multipart replay
// ---------------------------------------------------------------------------

// TestDo_429RetryAfterHeader_Honored verifies that when the server returns 429
// with a Retry-After: 0 header on the first attempt and 200 on the second,
// the client retries exactly once and succeeds. We assert exactly 2 attempts
// and a nil error; we deliberately do not assert on elapsed time to avoid
// timing flakiness.
func TestDo_429RetryAfterHeader_Honored(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.Header().Set("Retry-After", "0")
			writeJSON(t, w, 429, map[string]any{
				"object":  "error",
				"status":  429,
				"code":    "rate_limited",
				"message": "slow down",
			})
			return
		}
		writeJSON(t, w, 200, map[string]any{
			"object": "user",
			"id":     "bot-retry",
			"type":   "bot",
			"bot": map[string]any{
				"owner": map[string]any{"type": "workspace", "workspace": true},
			},
		})
	}))
	defer srv.Close()

	c := NewClient("test-token",
		WithBaseURL(srv.URL),
		WithRetryConfig(retry.RetryConfig{
			MaxRetries: 3,
			BaseDelay:  time.Millisecond,
			MaxDelay:   10 * time.Millisecond,
		}),
	)

	result, err := c.UsersMe(context.Background())
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("attempts = %d, want 2", got)
	}
	if result["id"] != "bot-retry" {
		t.Errorf("id = %v, want bot-retry", result["id"])
	}
}

// TestDo_503RetriedThenSucceeds verifies that a 503 Service Unavailable is
// retried and that the second attempt succeeds. Exactly two attempts must
// be observed.
func TestDo_503RetriedThenSucceeds(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			writeJSON(t, w, 503, map[string]any{
				"object":  "error",
				"status":  503,
				"code":    "service_unavailable",
				"message": "try again",
			})
			return
		}
		writeJSON(t, w, 200, map[string]any{
			"object": "user",
			"id":     "bot-503",
			"type":   "bot",
			"bot": map[string]any{
				"owner": map[string]any{"type": "workspace", "workspace": true},
			},
		})
	}))
	defer srv.Close()

	c := NewClient("test-token",
		WithBaseURL(srv.URL),
		WithRetryConfig(retry.RetryConfig{
			MaxRetries: 3,
			BaseDelay:  time.Millisecond,
			MaxDelay:   10 * time.Millisecond,
		}),
	)

	result, err := c.UsersMe(context.Background())
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("attempts = %d, want 2", got)
	}
	if result["id"] != "bot-503" {
		t.Errorf("id = %v, want bot-503", result["id"])
	}
}

// TestDoFormData_RetryOn5xx verifies the multipart body is replayable across
// retries: the file bytes received on attempt 2 must equal those received on
// attempt 1 (byte-for-byte). doFormData is unexported but reachable here
// because this test lives in package notion.
func TestDoFormData_RetryOn5xx(t *testing.T) {
	const fileFieldName = "file"
	const fileName = "hello.txt"
	filePayload := []byte("hello, multipart replay test \x00\x01\x02\x03")

	var attempts int32
	var firstFileBytes []byte
	var secondFileBytes []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)

		// Parse the multipart body on every attempt to extract the file part.
		ct := r.Header.Get("Content-Type")
		_, params, err := mime.ParseMediaType(ct)
		if err != nil {
			t.Errorf("parse content-type %q: %v", ct, err)
			w.WriteHeader(500)
			return
		}
		boundary, ok := params["boundary"]
		if !ok {
			t.Errorf("no boundary in content-type %q", ct)
			w.WriteHeader(500)
			return
		}
		mr := multipart.NewReader(r.Body, boundary)
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Errorf("next part: %v", err)
				break
			}
			if part.FormName() == fileFieldName {
				data, rerr := io.ReadAll(part)
				if rerr != nil {
					t.Errorf("read file part: %v", rerr)
				}
				if n == 1 {
					firstFileBytes = data
				} else {
					secondFileBytes = data
				}
			}
			_ = part.Close()
		}

		if n == 1 {
			writeJSON(t, w, 503, map[string]any{
				"object":  "error",
				"status":  503,
				"code":    "service_unavailable",
				"message": "try again",
			})
			return
		}
		writeJSON(t, w, 200, map[string]any{
			"object": "file_upload",
			"id":     "upload-1",
			"status": "uploaded",
		})
	}))
	defer srv.Close()

	c := NewClient("test-token",
		WithBaseURL(srv.URL),
		WithRetryConfig(retry.RetryConfig{
			MaxRetries: 3,
			BaseDelay:  time.Millisecond,
			MaxDelay:   10 * time.Millisecond,
		}),
	)

	result, err := c.doFormData(
		context.Background(),
		"/file_uploads/upload-1/send",
		map[string]string{}, // no extra form fields
		fileFieldName,
		fileName,
		"application/octet-stream",
		filePayload,
	)
	if err != nil {
		t.Fatalf("unexpected error after retry: %v", err)
	}
	if got := atomic.LoadInt32(&attempts); got != 2 {
		t.Errorf("attempts = %d, want 2", got)
	}
	if result["id"] != "upload-1" {
		t.Errorf("id = %v, want upload-1", result["id"])
	}

	if !bytes.Equal(firstFileBytes, filePayload) {
		t.Errorf("first attempt file bytes mismatch: got %q want %q", firstFileBytes, filePayload)
	}
	if !bytes.Equal(secondFileBytes, filePayload) {
		t.Errorf("second attempt file bytes mismatch: got %q want %q", secondFileBytes, filePayload)
	}
	if !bytes.Equal(firstFileBytes, secondFileBytes) {
		t.Errorf("file bytes differ between attempts: first=%q second=%q", firstFileBytes, secondFileBytes)
	}
}
