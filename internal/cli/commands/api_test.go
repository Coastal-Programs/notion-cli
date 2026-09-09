package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/spf13/cobra"
)

func runAPIRoot(t *testing.T, args ...string) (string, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterAPICommand(root)
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	var err error
	out := captureStdout(t, func() { err = root.Execute() })
	return out, err
}

// capturedRequest records what the test server received.
type capturedRequest struct {
	Method string
	Path   string
	Query  url.Values
	Header http.Header
	Body   map[string]any
}

func apiTestServer(t *testing.T, respond func(n int, w http.ResponseWriter)) (*[]capturedRequest, func()) {
	t.Helper()
	var mu sync.Mutex
	var got []capturedRequest
	_, cleanup := testWhoamiServer(t, func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		var body map[string]any
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &body)
		}
		mu.Lock()
		got = append(got, capturedRequest{r.Method, r.URL.Path, r.URL.Query(), r.Header.Clone(), body})
		n := len(got)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		respond(n, w)
	})
	return &got, cleanup
}

func TestAPICmd_GetWithQueryFields(t *testing.T) {
	got, cleanup := apiTestServer(t, func(_ int, w http.ResponseWriter) {
		_, _ = w.Write([]byte(`{"object":"user","id":"bot-1"}`))
	})
	defer cleanup()

	out, err := runAPIRoot(t, "api", "/v1/users/me", "-X", "GET", "-f", "page_size=5", "-H", "Notion-Beta: things", "--raw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*got) != 1 {
		t.Fatalf("requests = %d, want 1", len(*got))
	}
	r := (*got)[0]
	if r.Method != http.MethodGet || r.Path != "/users/me" {
		t.Errorf("got %s %s, want GET /users/me", r.Method, r.Path)
	}
	if r.Query.Get("page_size") != "5" {
		t.Errorf("query page_size = %q, want 5 (GET fields become query params)", r.Query.Get("page_size"))
	}
	if r.Header.Get("Notion-Beta") != "things" {
		t.Errorf("Notion-Beta header not forwarded: %v", r.Header)
	}
	if r.Header.Get("Authorization") == "" || r.Header.Get("Notion-Version") == "" {
		t.Errorf("default auth/version headers missing: %v", r.Header)
	}
	if !strings.Contains(out, `"id": "bot-1"`) {
		t.Errorf("raw output missing response body: %s", out)
	}
}

func TestAPICmd_PostBodyFromInputAndFields(t *testing.T) {
	got, cleanup := apiTestServer(t, func(_ int, w http.ResponseWriter) {
		_, _ = w.Write([]byte(`{"object":"page","id":"p1"}`))
	})
	defer cleanup()

	dir := t.TempDir()
	input := filepath.Join(dir, "body.json")
	if err := os.WriteFile(input, []byte(`{"parent":{"page_id":"old"},"keep":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	propsFile := filepath.Join(dir, "props.json")
	if err := os.WriteFile(propsFile, []byte(`{"Name":{"title":[]}}`), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := runAPIRoot(t, "api", "pages",
		"--input", input,
		"-f", "parent.page_id=new",
		"-F", "in_trash=true",
		"-F", "properties=@"+propsFile,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := (*got)[0]
	if r.Method != http.MethodPost {
		t.Errorf("method = %s, want POST (implied by body)", r.Method)
	}
	parent, _ := r.Body["parent"].(map[string]any)
	if parent["page_id"] != "new" {
		t.Errorf("dotted -f did not override nested input field: %v", r.Body)
	}
	if r.Body["keep"] != float64(1) {
		t.Errorf("--input fields were not preserved: %v", r.Body)
	}
	if r.Body["in_trash"] != true {
		t.Errorf("-F should parse JSON booleans, got %v", r.Body["in_trash"])
	}
	if props, _ := r.Body["properties"].(map[string]any); props["Name"] == nil {
		t.Errorf("-F @file value not read: %v", r.Body["properties"])
	}
	if r.Header.Get("Content-Type") != "application/json" {
		t.Errorf("Content-Type = %q", r.Header.Get("Content-Type"))
	}
}

func TestAPICmd_Paginate(t *testing.T) {
	got, cleanup := apiTestServer(t, func(n int, w http.ResponseWriter) {
		if n == 1 {
			_, _ = w.Write([]byte(`{"object":"list","results":[{"id":"a"}],"has_more":true,"next_cursor":"c2"}`))
			return
		}
		_, _ = w.Write([]byte(`{"object":"list","results":[{"id":"b"}],"has_more":false,"next_cursor":null}`))
	})
	defer cleanup()

	out, err := runAPIRoot(t, "api", "data_sources/ds1/query", "-X", "POST", "-F", "page_size=1", "--paginate", "--output", "json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*got) != 2 {
		t.Fatalf("requests = %d, want 2", len(*got))
	}
	if (*got)[0].Body["start_cursor"] != nil {
		t.Errorf("first request must not carry start_cursor")
	}
	if (*got)[1].Body["start_cursor"] != "c2" {
		t.Errorf("second POST body start_cursor = %v, want c2", (*got)[1].Body["start_cursor"])
	}
	env := parseEnvelope(t, out)
	data, _ := env["data"].(map[string]any)
	if data["result_count"] != float64(2) {
		t.Errorf("result_count = %v, want 2", data["result_count"])
	}
}

func TestAPICmd_PaginateGetUsesQueryCursor(t *testing.T) {
	got, cleanup := apiTestServer(t, func(n int, w http.ResponseWriter) {
		if n == 1 {
			_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":true,"next_cursor":"c2"}`))
			return
		}
		_, _ = w.Write([]byte(`{"object":"list","results":[],"has_more":false}`))
	})
	defer cleanup()

	if _, err := runAPIRoot(t, "api", "users", "--paginate"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if (*got)[1].Query.Get("start_cursor") != "c2" {
		t.Errorf("GET pagination must pass start_cursor as a query param: %v", (*got)[1].Query)
	}
}

// A bodyless POST (e.g. an unfiltered data source query) must still carry the
// cursor in the body; Notion ignores a start_cursor query param on POST, which
// would otherwise re-fetch page one until the pagination cap.
func TestAPICmd_PaginateBodylessPostUsesBodyCursor(t *testing.T) {
	got, cleanup := apiTestServer(t, func(n int, w http.ResponseWriter) {
		if n == 1 {
			_, _ = w.Write([]byte(`{"object":"list","results":[{"id":"a"}],"has_more":true,"next_cursor":"c2"}`))
			return
		}
		_, _ = w.Write([]byte(`{"object":"list","results":[{"id":"b"}],"has_more":false}`))
	})
	defer cleanup()

	if _, err := runAPIRoot(t, "api", "data_sources/ds1/query", "-X", "POST", "--paginate"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*got) != 2 {
		t.Fatalf("requests = %d, want 2", len(*got))
	}
	if (*got)[1].Body["start_cursor"] != "c2" {
		t.Errorf("second POST body start_cursor = %v, want c2", (*got)[1].Body)
	}
	if (*got)[1].Query.Get("start_cursor") != "" {
		t.Errorf("POST pagination must not use a query cursor: %v", (*got)[1].Query)
	}
}

// --paginate against a non-list endpoint prints the response instead of
// silently reducing it to an empty results list.
func TestAPICmd_PaginateNonListPrintsResponse(t *testing.T) {
	got, cleanup := apiTestServer(t, func(_ int, w http.ResponseWriter) {
		_, _ = w.Write([]byte(`{"object":"user","id":"bot-1"}`))
	})
	defer cleanup()

	out, err := runAPIRoot(t, "api", "users/me", "--paginate", "--raw")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*got) != 1 {
		t.Errorf("requests = %d, want 1", len(*got))
	}
	if !strings.Contains(out, `"id": "bot-1"`) || strings.Contains(out, `"result_count"`) {
		t.Errorf("non-list response should be printed as-is, got: %s", out)
	}
}

func TestAPICmd_RejectsAbsoluteURL(t *testing.T) {
	_, cleanup := apiTestServer(t, func(_ int, w http.ResponseWriter) { _, _ = w.Write([]byte(`{}`)) })
	defer cleanup()

	for _, ep := range []string{"https://evil.example/v1/users", "//evil.example/users", ""} {
		_, err := runAPIRoot(t, "api", ep)
		if err == nil {
			t.Errorf("endpoint %q: expected rejection, token must only go to the API host", ep)
		}
	}
}

func TestAPICmd_APIErrorIsStructured(t *testing.T) {
	_, cleanup := apiTestServer(t, func(_ int, w http.ResponseWriter) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"object":"error","status":404,"code":"object_not_found","message":"nope"}`))
	})
	defer cleanup()

	_, err := runAPIRoot(t, "api", "pages/missing")
	if err == nil || !strings.Contains(err.Error(), "NOT_FOUND") {
		t.Errorf("expected structured NOT_FOUND error, got %v", err)
	}
}

func TestParseAPIFields_BadInput(t *testing.T) {
	if _, err := parseAPIFields([]string{"novalue"}, nil, strings.NewReader("")); err == nil {
		t.Error("-f without '=' should fail")
	}
	if _, err := parseAPIFields(nil, []string{"x=not json"}, strings.NewReader("")); err == nil {
		t.Error("-F with non-JSON value should fail")
	}
	fields, err := parseAPIFields(nil, []string{"x=@-"}, strings.NewReader(`{"a":1}`))
	if err != nil || fields["x"].(map[string]any)["a"] != float64(1) {
		t.Errorf("-F @- should read JSON from stdin, got %v %v", fields, err)
	}
}

func TestSetNestedField_ConflictingType(t *testing.T) {
	m := map[string]any{"parent": "a-string"}
	if err := setNestedField(m, "parent.page_id", "x"); err == nil {
		t.Error("nesting under a non-object should fail rather than silently overwrite")
	}
}
