package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

// testCommentServer creates a test HTTP server, sets NOTION_TOKEN /
// NOTION_CLI_BASE_URL, and returns a cleanup func.
func testCommentServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
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

// runCommentRoot returns a root command with comment subcommands registered
// and a captured stdout/stderr buffer.
func runCommentRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterCommentCommands(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

const (
	testIDRaw = "5c6a28216bb14a7eb6e1c50111515c3d"
	testID    = "5c6a2821-6bb1-4a7e-b6e1-c50111515c3d"
)

// ---------------------------------------------------------------------------
// Validation tests (no network)
// ---------------------------------------------------------------------------

func TestCommentSubcommands_Registered(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterCommentCommands(root)
	for _, sub := range []string{"create", "list", "retrieve", "update", "delete"} {
		c, _, err := root.Find([]string{"comment", sub})
		if err != nil || c == nil || c.Name() != sub {
			t.Errorf("subcommand %q not found: %v", sub, err)
		}
	}
}

func TestCommentCreate_RequiresParent(t *testing.T) {
	_, _, err := runCommentRoot(t, "comment", "create", "--text", "hi")
	if err == nil {
		t.Fatal("expected error when no parent flag is set")
	}
	if !strings.Contains(err.Error(), "--page") &&
		!strings.Contains(err.Error(), "--block") &&
		!strings.Contains(err.Error(), "--discussion") {
		t.Errorf("expected error mentioning parent flags, got: %v", err)
	}
}

func TestCommentCreate_RequiresBody(t *testing.T) {
	_, _, err := runCommentRoot(t, "comment", "create", "--page", testID)
	if err == nil {
		t.Fatal("expected error when neither --text nor --rich-text is set")
	}
	if !strings.Contains(err.Error(), "--text") && !strings.Contains(err.Error(), "--rich-text") {
		t.Errorf("expected error mentioning --text/--rich-text, got: %v", err)
	}
}

func TestCommentCreate_ParentFlagsMutuallyExclusive(t *testing.T) {
	_, _, err := runCommentRoot(t, "comment", "create",
		"--page", testID, "--block", testID, "--text", "hi",
	)
	if err == nil {
		t.Fatal("expected mutually-exclusive error")
	}
}

func TestCommentCreate_TextAndRichTextMutuallyExclusive(t *testing.T) {
	_, _, err := runCommentRoot(t, "comment", "create",
		"--page", testID, "--text", "hi", "--rich-text", "rt.json",
	)
	if err == nil {
		t.Fatal("expected mutually-exclusive error for --text/--rich-text")
	}
}

func TestCommentList_RequiresBlock(t *testing.T) {
	_, _, err := runCommentRoot(t, "comment", "list")
	if err == nil {
		t.Fatal("expected error when --block is missing")
	}
}

func TestCommentUpdate_RequiresBody(t *testing.T) {
	_, _, err := runCommentRoot(t, "comment", "update", testID)
	if err == nil {
		t.Fatal("expected error when neither --text nor --rich-text is set")
	}
}

// ---------------------------------------------------------------------------
// HTTP roundtrip tests
// ---------------------------------------------------------------------------

func TestCommentCreate_PageParent_SendsExpectedBody(t *testing.T) {
	var captured map[string]any
	var hits int32

	srv, cleanup := testCommentServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/comments" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
			return
		}
		atomic.AddInt32(&hits, 1)
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "comment", "id": "c1"})
	})
	defer cleanup()
	_ = srv

	_, _, err := runCommentRoot(t, "comment", "create",
		"--page", testID,
		"--text", "hello",
		"--display-name", "Bot",
		"--attach-file", "upload-1",
		"--json",
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if atomic.LoadInt32(&hits) != 1 {
		t.Fatalf("expected 1 POST /comments, got %d", hits)
	}

	parent, _ := captured["parent"].(map[string]any)
	if parent == nil || parent["page_id"] != testID {
		t.Errorf("expected parent.page_id=%s, got %v", testID, captured["parent"])
	}
	rt, _ := captured["rich_text"].([]any)
	if len(rt) != 1 {
		t.Fatalf("expected 1 rich_text run, got %v", captured["rich_text"])
	}
	dn, _ := captured["display_name"].(map[string]any)
	if dn == nil || dn["type"] != "custom" {
		t.Errorf("expected custom display_name, got %v", captured["display_name"])
	}
	atts, _ := captured["attachments"].([]any)
	if len(atts) != 1 {
		t.Fatalf("expected 1 attachment, got %v", captured["attachments"])
	}
	att, _ := atts[0].(map[string]any)
	if att["file_upload_id"] != "upload-1" {
		t.Errorf("attachment file_upload_id wrong: %v", att)
	}
}

func TestCommentCreate_DiscussionParent(t *testing.T) {
	var captured map[string]any
	srv, cleanup := testCommentServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "comment", "id": "c1"})
	})
	defer cleanup()
	_ = srv

	_, _, err := runCommentRoot(t, "comment", "create",
		"--discussion", testID, "--text", "reply", "--json",
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if captured["discussion_id"] != testID {
		t.Errorf("expected discussion_id=%s, got %v", testID, captured["discussion_id"])
	}
	if _, has := captured["parent"]; has {
		t.Errorf("did not expect parent field, got %v", captured["parent"])
	}
}

func TestCommentCreate_RichTextFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rt.json")
	rtJSON := `[{"type":"text","text":{"content":"hello from file"}}]`
	if err := os.WriteFile(path, []byte(rtJSON), 0o600); err != nil {
		t.Fatal(err)
	}

	var captured map[string]any
	srv, cleanup := testCommentServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "comment", "id": "c1"})
	})
	defer cleanup()
	_ = srv

	_, _, err := runCommentRoot(t, "comment", "create",
		"--block", testID, "--rich-text", path, "--json",
	)
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	rt, _ := captured["rich_text"].([]any)
	if len(rt) != 1 {
		t.Fatalf("expected 1 run, got %v", captured["rich_text"])
	}
	first, _ := rt[0].(map[string]any)
	tx, _ := first["text"].(map[string]any)
	if tx["content"] != "hello from file" {
		t.Errorf("rich_text content not loaded from file: %v", first)
	}
}

func TestCommentList_QueriesBlockID(t *testing.T) {
	var seenBlockID string
	srv, cleanup := testCommentServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/comments" {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		seenBlockID = r.URL.Query().Get("block_id")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list", "results": []any{}, "has_more": false,
		})
	})
	defer cleanup()
	_ = srv

	_, _, err := runCommentRoot(t, "comment", "list", "--block", testID, "--json")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if seenBlockID != testID {
		t.Errorf("expected block_id=%s, got %s", testID, seenBlockID)
	}
}

func TestCommentList_AllPaginates(t *testing.T) {
	var calls int32
	srv, cleanup := testCommentServer(t, func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		w.Header().Set("Content-Type", "application/json")
		switch n {
		case 1:
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "list", "results": []any{map[string]any{"id": "c1"}},
				"has_more": true, "next_cursor": "cur2",
			})
		case 2:
			if r.URL.Query().Get("start_cursor") != "cur2" {
				t.Errorf("expected start_cursor=cur2, got %q", r.URL.Query().Get("start_cursor"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object": "list", "results": []any{map[string]any{"id": "c2"}},
				"has_more": false,
			})
		default:
			t.Errorf("unexpected extra call %d", n)
		}
	})
	defer cleanup()
	_ = srv

	_, _, err := runCommentRoot(t, "comment", "list", "--block", testID, "--all", "--json")
	if err != nil {
		t.Fatalf("list --all failed: %v", err)
	}
	if atomic.LoadInt32(&calls) != 2 {
		t.Errorf("expected 2 paginated calls, got %d", calls)
	}
}

func TestCommentRetrieve_Roundtrip(t *testing.T) {
	srv, cleanup := testCommentServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/comments/"+testID {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "comment", "id": testID})
	})
	defer cleanup()
	_ = srv

	_, _, err := runCommentRoot(t, "comment", "retrieve", testID, "--json")
	if err != nil {
		t.Fatalf("retrieve failed: %v", err)
	}
}

func TestCommentUpdate_Roundtrip(t *testing.T) {
	var captured map[string]any
	srv, cleanup := testCommentServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" || r.URL.Path != "/comments/"+testID {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&captured)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "comment", "id": testID})
	})
	defer cleanup()
	_ = srv

	_, _, err := runCommentRoot(t, "comment", "update", testID, "--text", "edited", "--json")
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	rt, _ := captured["rich_text"].([]any)
	if len(rt) != 1 {
		t.Errorf("expected rich_text in update body, got %v", captured)
	}
}

func TestCommentDelete_Roundtrip(t *testing.T) {
	var hit int32
	srv, cleanup := testCommentServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" || r.URL.Path != "/comments/"+testID {
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
		atomic.AddInt32(&hit, 1)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "comment", "id": testID})
	})
	defer cleanup()
	_ = srv

	_, _, err := runCommentRoot(t, "comment", "delete", testID, "--json")
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	if atomic.LoadInt32(&hit) != 1 {
		t.Errorf("expected 1 DELETE, got %d", hit)
	}
}

// ---------------------------------------------------------------------------
// loadRichText pure-function tests
// ---------------------------------------------------------------------------

func TestLoadRichText_Success(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rt-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.WriteString(`[{"type":"text","text":{"content":"hi"}}]`)
	_ = tmpFile.Close()

	arr, err := loadRichText(tmpFile.Name())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(arr) != 1 {
		t.Errorf("got %d items, want 1", len(arr))
	}
}

func TestLoadRichText_FileNotFound(t *testing.T) {
	_, err := loadRichText("/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadRichText_InvalidJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "rt-bad-*.json")
	if err != nil {
		t.Fatalf("CreateTemp: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()
	_, _ = tmpFile.WriteString(`not json`)
	_ = tmpFile.Close()

	_, err = loadRichText(tmpFile.Name())
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}
