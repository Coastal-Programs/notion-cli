package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/spf13/cobra"
)

// testFilesServer creates a test HTTP server, sets NOTION_TOKEN /
// NOTION_CLI_BASE_URL, and returns the server plus a cleanup func.
func testFilesServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
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

// runFilesRoot returns a root command with files subcommands registered
// and a captured stdout buffer.
func runFilesRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterFilesCommands(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

// ---------------------------------------------------------------------------
// 1. Subcommands registered
// ---------------------------------------------------------------------------

func TestFilesSubcommands_Registered(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterFilesCommands(root)
	for _, sub := range []string{"upload", "retrieve", "list"} {
		c, _, err := root.Find([]string{"files", sub})
		if err != nil || c == nil || c.Name() != sub {
			t.Errorf("subcommand %q not found: %v", sub, err)
		}
	}
}

// ---------------------------------------------------------------------------
// 2. Single-part upload flow
// ---------------------------------------------------------------------------

func TestFilesUpload_SinglePart_Flow(t *testing.T) {
	const uploadID = "upload-single-001"

	var createBody map[string]any
	var sendMethod, sendPath string
	var sendFilename string
	var callOrder []string

	_, cleanup := testFilesServer(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == "POST" && r.URL.Path == "/file_uploads":
			callOrder = append(callOrder, "create")
			_ = json.NewDecoder(r.Body).Decode(&createBody)
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": uploadID, "object": "file_upload"})

		case r.Method == "POST" && r.URL.Path == "/file_uploads/"+uploadID+"/send":
			callOrder = append(callOrder, "send")
			sendMethod = r.Method
			sendPath = r.URL.Path

			mt, params, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
			if !strings.HasPrefix(mt, "multipart/") {
				t.Errorf("expected multipart content-type, got %q", mt)
			}
			mr := multipart.NewReader(r.Body, params["boundary"])
			for {
				part, err := mr.NextPart()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("multipart read error: %v", err)
				}
				if part.FormName() == "file" {
					sendFilename = part.FileName()
				}
				_ = part.Close()
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"id": uploadID, "object": "file_upload", "status": "uploaded"})

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
		}
	})
	defer cleanup()

	// Create a small temp file (<20 MB).
	dir := t.TempDir()
	filePath := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(filePath, []byte("hello world"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Override TTY detection to suppress progress output.
	orig := isStderrTTY
	isStderrTTY = func() bool { return false }
	defer func() { isStderrTTY = orig }()

	_, _, err := runFilesRoot(t, "files", "upload", filePath, "--json")
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	if len(callOrder) != 2 || callOrder[0] != "create" || callOrder[1] != "send" {
		t.Errorf("expected [create send], got %v", callOrder)
	}
	if createBody["mode"] != "single_part" {
		t.Errorf("expected mode=single_part, got %v", createBody["mode"])
	}
	if createBody["filename"] != "hello.txt" {
		t.Errorf("expected filename=hello.txt, got %v", createBody["filename"])
	}
	if sendMethod != "POST" || sendPath != "/file_uploads/"+uploadID+"/send" {
		t.Errorf("unexpected send request: %s %s", sendMethod, sendPath)
	}
	if sendFilename != "hello.txt" {
		t.Errorf("expected file part filename=hello.txt, got %q", sendFilename)
	}
}

// ---------------------------------------------------------------------------
// 3. Multi-part upload flow
// ---------------------------------------------------------------------------

func TestFilesUpload_MultiPart_Flow(t *testing.T) {
	const uploadID = "upload-multi-001"

	var createBody map[string]any
	var sendCalls int32
	var completeCalls int32

	_, cleanup := testFilesServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/file_uploads":
			_ = json.NewDecoder(r.Body).Decode(&createBody)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": uploadID, "object": "file_upload"})

		case r.Method == "POST" && r.URL.Path == "/file_uploads/"+uploadID+"/send":
			// Drain the body.
			_, _ = io.Copy(io.Discard, r.Body)
			atomic.AddInt32(&sendCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": uploadID})

		case r.Method == "POST" && r.URL.Path == "/file_uploads/"+uploadID+"/complete":
			atomic.AddInt32(&completeCalls, 1)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": uploadID, "status": "uploaded"})

		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
		}
	})
	defer cleanup()

	// Override the single-part threshold so a small test file triggers multi-part.
	// With threshold=4 bytes and chunk-size=5MB, any file >4 bytes uses multi-part.
	// We write a 16-byte file and set threshold to 4 bytes, giving ceil(16/4)=4 parts.
	// However, parseChunkSize enforces 5MB minimum, so we work around it differently:
	// write a 4 * 5MB-chunk file worth of data by setting threshold below file size.
	//
	// Simplest approach: use a 12-byte file, override threshold to 4 bytes so numParts=3,
	// and pass --chunk-size 5MB (only used for the API body; actual read loop uses chunkSize).
	// But chunkSize is parsed from the flag, so the read loop reads in 5MB chunks regardless.
	// For correctness we need the file to be large enough to produce multiple 5MB chunks.
	//
	// Instead, we create a temporary helper: override singlePartMaxBytes to force multi-part,
	// and create a file whose size / chunkSize = 3 parts. chunkSize must stay ≥5MB per the
	// flag validator, so we need a file of at least 10MB+1 byte (two full 5MB chunks + remainder).
	// We write 11 MB and use --chunk-size 5MB → ceil(11/5)=3 parts.

	origSingle := singlePartMaxBytes
	singlePartMaxBytes = 1 // force multi-part for any non-empty file
	defer func() { singlePartMaxBytes = origSingle }()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "large.bin")
	// 11 MB → ceil(11 MB / 5 MB) = 3 parts (two full 5MB + one 1MB remainder)
	data := make([]byte, 11*1024*1024)
	if err := os.WriteFile(filePath, data, 0o600); err != nil {
		t.Fatal(err)
	}

	orig := isStderrTTY
	isStderrTTY = func() bool { return false }
	defer func() { isStderrTTY = orig }()

	_, _, err := runFilesRoot(t, "files", "upload", filePath, "--chunk-size", "5MB", "--json")
	if err != nil {
		t.Fatalf("multipart upload failed: %v", err)
	}

	if createBody["mode"] != "multi_part" {
		t.Errorf("expected mode=multi_part, got %v", createBody["mode"])
	}

	// 11 MB / 5 MB = ceil = 3 parts.
	expectedParts := float64(3)
	if np, ok := createBody["number_of_parts"].(float64); !ok || np != expectedParts {
		t.Errorf("expected number_of_parts=%v, got %v", expectedParts, createBody["number_of_parts"])
	}

	if got := atomic.LoadInt32(&sendCalls); got != 3 {
		t.Errorf("expected 3 send calls, got %d", got)
	}
	if got := atomic.LoadInt32(&completeCalls); got != 1 {
		t.Errorf("expected 1 complete call, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// 4. Retry on 429
// ---------------------------------------------------------------------------

func TestFilesUpload_Retry429(t *testing.T) {
	const uploadID = "upload-retry-001"

	var sendAttempts int32

	_, cleanup := testFilesServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/file_uploads":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": uploadID})

		case r.Method == "POST" && r.URL.Path == "/file_uploads/"+uploadID+"/send":
			n := atomic.AddInt32(&sendAttempts, 1)
			_, _ = io.Copy(io.Discard, r.Body)
			if n < 3 {
				w.WriteHeader(429)
				_ = json.NewEncoder(w).Encode(map[string]any{"status": 429, "code": "rate_limited", "message": "slow down"})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"id": uploadID, "status": "uploaded"})

		default:
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(500)
		}
	})
	defer cleanup()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "retry.txt")
	if err := os.WriteFile(filePath, []byte("retry test data"), 0o600); err != nil {
		t.Fatal(err)
	}

	orig := isStderrTTY
	isStderrTTY = func() bool { return false }
	defer func() { isStderrTTY = orig }()

	_, _, err := runFilesRoot(t, "files", "upload", filePath, "--json")
	if err != nil {
		t.Fatalf("upload with retries failed: %v", err)
	}

	if got := atomic.LoadInt32(&sendAttempts); got < 3 {
		t.Errorf("expected at least 3 send attempts (2 retries), got %d", got)
	}
}

// ---------------------------------------------------------------------------
// 5. Progress silent when not a TTY
// ---------------------------------------------------------------------------

func TestFilesUpload_ProgressSilentNonTTY(t *testing.T) {
	const uploadID = "upload-progress-001"

	_, cleanup := testFilesServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/file_uploads":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": uploadID})
		case r.Method == "POST" && r.URL.Path == "/file_uploads/"+uploadID+"/send":
			_, _ = io.Copy(io.Discard, r.Body)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": uploadID})
		default:
			w.WriteHeader(500)
		}
	})
	defer cleanup()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "silent.txt")
	if err := os.WriteFile(filePath, []byte("data"), 0o600); err != nil {
		t.Fatal(err)
	}

	// Override isStderrTTY to false and capture stderr separately.
	orig := isStderrTTY
	isStderrTTY = func() bool { return false }
	defer func() { isStderrTTY = orig }()

	// Redirect stderr to a pipe to verify nothing is written.
	oldStderr := os.Stderr
	pr, pw, _ := os.Pipe()
	os.Stderr = pw

	_, _, err := runFilesRoot(t, "files", "upload", filePath)
	_ = pw.Close()
	os.Stderr = oldStderr

	var stderrBuf bytes.Buffer
	_, _ = io.Copy(&stderrBuf, pr)

	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	// Progress writes to os.Stderr directly; when isStderrTTY returns false
	// we expect nothing.
	if stderrBuf.Len() > 0 {
		t.Errorf("expected no stderr output (non-TTY), got: %q", stderrBuf.String())
	}
}

// ---------------------------------------------------------------------------
// 6. Output is bare ID without --json
// ---------------------------------------------------------------------------

func TestFilesUpload_OutputID_NoEnvelope(t *testing.T) {
	const uploadID = "upload-bare-001"

	_, cleanup := testFilesServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == "POST" && r.URL.Path == "/file_uploads":
			_ = json.NewEncoder(w).Encode(map[string]any{"id": uploadID})
		case r.Method == "POST" && r.URL.Path == "/file_uploads/"+uploadID+"/send":
			_, _ = io.Copy(io.Discard, r.Body)
			_ = json.NewEncoder(w).Encode(map[string]any{"id": uploadID, "status": "uploaded"})
		default:
			w.WriteHeader(500)
		}
	})
	defer cleanup()

	dir := t.TempDir()
	filePath := filepath.Join(dir, "bare.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o600); err != nil {
		t.Fatal(err)
	}

	orig := isStderrTTY
	isStderrTTY = func() bool { return false }
	defer func() { isStderrTTY = orig }()

	_, buf, err := runFilesRoot(t, "files", "upload", filePath)
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}

	got := strings.TrimRight(buf.String(), "\n")
	if got != uploadID {
		t.Errorf("expected bare ID %q, got %q", uploadID, got)
	}
}

// ---------------------------------------------------------------------------
// 7. Retrieve hits correct endpoint
// ---------------------------------------------------------------------------

func TestFilesRetrieve_HitsCorrectEndpoint(t *testing.T) {
	const uploadID = "upload-retrieve-001"

	var hitPath string
	_, cleanup := testFilesServer(t, func(w http.ResponseWriter, r *http.Request) {
		hitPath = r.URL.Path
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"id": uploadID, "object": "file_upload"})
	})
	defer cleanup()

	_, _, err := runFilesRoot(t, "files", "retrieve", uploadID, "--json")
	if err != nil {
		t.Fatalf("retrieve failed: %v", err)
	}
	if hitPath != "/file_uploads/"+uploadID {
		t.Errorf("expected GET /file_uploads/%s, got %s", uploadID, hitPath)
	}
}

// ---------------------------------------------------------------------------
// 8. List respects --page-size
// ---------------------------------------------------------------------------

func TestFilesList_PageSize(t *testing.T) {
	var seenPageSize string

	_, cleanup := testFilesServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" || r.URL.Path != "/file_uploads" {
			t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
		}
		seenPageSize = r.URL.Query().Get("page_size")
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":   "list",
			"results":  []any{},
			"has_more": false,
		})
	})
	defer cleanup()

	_, _, err := runFilesRoot(t, "files", "list", "--page-size", "25", "--json")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if seenPageSize != "25" {
		t.Errorf("expected page_size=25, got %q", seenPageSize)
	}
}

func TestFilesList_All(t *testing.T) {
	callCount := 0
	_, cleanup := testFilesServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object":      "list",
				"results":     []any{map[string]any{"id": "fu-1"}},
				"has_more":    true,
				"next_cursor": "cursor1",
			})
		} else {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"object":   "list",
				"results":  []any{map[string]any{"id": "fu-2"}},
				"has_more": false,
			})
		}
	})
	defer cleanup()

	_, _, err := runFilesRoot(t, "files", "list", "--all")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 API calls for --all, got %d", callCount)
	}
}

// ---------------------------------------------------------------------------
// parseChunkSize pure-function tests
// ---------------------------------------------------------------------------

func TestParseChunkSize_MB(t *testing.T) {
	n, err := parseChunkSize("10MB")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 10*1024*1024 {
		t.Errorf("n = %d, want %d", n, 10*1024*1024)
	}
}

func TestParseChunkSize_OutOfRange(t *testing.T) {
	_, err := parseChunkSize("100MB")
	if err == nil {
		t.Fatal("expected error for out-of-range size")
	}
}

func TestParseChunkSize_UnknownUnit(t *testing.T) {
	_, err := parseChunkSize("100GB")
	if err == nil {
		t.Fatal("expected error for unknown unit")
	}
}

func TestParseChunkSize_InvalidNumber(t *testing.T) {
	_, err := parseChunkSize("abcMB")
	if err == nil {
		t.Fatal("expected error for invalid number")
	}
}
