package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func runAPIRoot(t *testing.T, input io.Reader, args ...string) (*bytes.Buffer, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	root.PersistentFlags().BoolP("verbose", "v", false, "")
	root.PersistentFlags().String("notion-version", "", "")
	RegisterAPICommands(root)
	var out, errBuf bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&errBuf)
	if input != nil {
		root.SetIn(input)
	}
	root.SetArgs(args)
	err := root.Execute()
	return &out, &errBuf, err
}

func withAPIServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Setenv("NOTION_TOKEN", "secret_test_token")
	t.Setenv("NOTION_API_TOKEN", "")
	t.Setenv("NOTION_CLI_BASE_URL", srv.URL)
	t.Setenv("NOTION_API_VERSION", "")
	return srv
}

func TestAPICommand_DefaultPostWithInlineBody(t *testing.T) {
	var gotBody map[string]any
	srv := withAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/v1/search" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer secret_test_token" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	defer srv.Close()

	out, _, err := runAPIRoot(t, nil, "api", "v1/search", "query=roadmap", "page_size:=10")
	if err != nil {
		t.Fatalf("api command failed: %v", err)
	}
	if !strings.Contains(out.String(), `"ok":true`) {
		t.Fatalf("raw output = %q", out.String())
	}
	if gotBody["query"] != "roadmap" || gotBody["page_size"] != float64(10) {
		t.Fatalf("body = %#v", gotBody)
	}
}

func TestAPICommand_DefaultGetWithoutBody(t *testing.T) {
	srv := withAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("method = %s, want GET", r.Method)
		}
		if r.URL.Path != "/v1/users/me" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "user"})
	})
	defer srv.Close()

	if _, _, err := runAPIRoot(t, nil, "api", "/v1/users/me"); err != nil {
		t.Fatalf("api command failed: %v", err)
	}
}

func TestAPICommand_MethodQueryHeaderAndVersion(t *testing.T) {
	srv := withAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("method = %s, want PATCH", r.Method)
		}
		if r.URL.Query()["filter_properties"] == nil || r.URL.Query().Get("page_size") != "10" {
			t.Fatalf("query = %#v", r.URL.Query())
		}
		if r.Header.Get("X-Trace-Id") != "cli-test" {
			t.Fatalf("X-Trace-Id = %q", r.Header.Get("X-Trace-Id"))
		}
		if r.Header.Get("Notion-Version") != "2026-03-11" {
			t.Fatalf("Notion-Version = %q", r.Header.Get("Notion-Version"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	defer srv.Close()
	t.Setenv("NOTION_API_VERSION", "2026-03-11")

	_, _, err := runAPIRoot(t, nil,
		"api", "v1/pages/abc",
		"-X", "PATCH",
		"archived:=true",
		"page_size==10",
		"filter_properties==Name",
		"X-Trace-Id:cli-test",
	)
	if err != nil {
		t.Fatalf("api command failed: %v", err)
	}
}

func TestAPICommand_DataAndInlineConflict(t *testing.T) {
	_, _, err := runAPIRoot(t, nil, "api", "v1/search", "--data", `{"query":"x"}`, "page_size:=1")
	if err == nil {
		t.Fatal("expected conflict error")
	}
}

func TestAPICommand_FileMultipart(t *testing.T) {
	tmp := t.TempDir() + "/chunk.txt"
	if err := os.WriteFile(tmp, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}
	srv := withAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method = %s, want POST", r.Method)
		}
		reader, err := r.MultipartReader()
		if err != nil {
			t.Fatalf("multipart reader: %v", err)
		}
		gotPartNumber := false
		gotFile := false
		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				t.Fatalf("next part: %v", err)
			}
			data, _ := io.ReadAll(part)
			switch part.FormName() {
			case "part_number":
				gotPartNumber = string(data) == "1"
			case "file":
				gotFile = part.FileName() == "chunk.txt" && string(data) == "hello"
			}
		}
		if !gotPartNumber || !gotFile {
			t.Fatalf("multipart fields missing: part=%v file=%v", gotPartNumber, gotFile)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	defer srv.Close()

	_, _, err := runAPIRoot(t, nil, "api", "v1/file_uploads/up/send", "--file", tmp, "part_number:=1")
	if err != nil {
		t.Fatalf("api command failed: %v", err)
	}
}

func TestAPICommand_LSSpecDocs(t *testing.T) {
	out, _, err := runAPIRoot(t, nil, "api", "ls", "--json")
	if err != nil {
		t.Fatalf("api ls failed: %v", err)
	}
	if !strings.Contains(out.String(), `"path": "v1/search"`) {
		t.Fatalf("api ls output missing search: %s", out.String())
	}

	out, _, err = runAPIRoot(t, nil, "api", "v1/search", "--spec", "-X", "POST")
	if err != nil {
		t.Fatalf("api spec failed: %v", err)
	}
	if !strings.Contains(out.String(), `"method": "POST"`) {
		t.Fatalf("spec output = %s", out.String())
	}

	out, _, err = runAPIRoot(t, nil, "api", "v1/search", "--docs", "-X", "POST")
	if err != nil {
		t.Fatalf("api docs failed: %v", err)
	}
	if !strings.Contains(out.String(), "# POST /v1/search") {
		t.Fatalf("docs output = %s", out.String())
	}
}

func TestAPICommand_EnvTokenAlias(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer ntn_test_token" {
			t.Fatalf("Authorization = %q", r.Header.Get("Authorization"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	}))
	defer srv.Close()
	t.Setenv("NOTION_TOKEN", "")
	t.Setenv("NOTION_API_TOKEN", "ntn_test_token")
	t.Setenv("NOTION_CLI_BASE_URL", srv.URL)

	if _, _, err := runAPIRoot(t, nil, "api", "v1/users/me"); err != nil {
		t.Fatalf("api command failed: %v", err)
	}
}

func TestAPIVerboseRedactsAuthorization(t *testing.T) {
	srv := withAPIServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true})
	})
	defer srv.Close()

	_, errBuf, err := runAPIRoot(t, nil, "api", "v1/users/me", "--verbose")
	if err != nil {
		t.Fatalf("api command failed: %v", err)
	}
	if strings.Contains(errBuf.String(), "secret_test_token") {
		t.Fatalf("verbose output leaked token: %s", errBuf.String())
	}
	if !strings.Contains(errBuf.String(), "<redacted>") {
		t.Fatalf("verbose output did not redact auth: %s", errBuf.String())
	}
}
