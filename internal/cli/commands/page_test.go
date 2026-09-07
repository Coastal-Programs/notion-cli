package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/spf13/cobra"
)

func testPageServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
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

func runPageRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterPageCommands(root)
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

// testPageID is declared in markdown_test.go

func newIconCoverFlagSet(t *testing.T) *cobra.Command {
	t.Helper()
	c := &cobra.Command{Use: "test"}
	c.Flags().String("icon-emoji", "", "")
	c.Flags().String("icon-url", "", "")
	c.Flags().String("cover-url", "", "")
	return c
}

func TestBuildIconCover_Emoji(t *testing.T) {
	c := newIconCoverFlagSet(t)
	_ = c.Flags().Set("icon-emoji", "💰")

	icon, cover, hasIcon, hasCover, err := buildIconCover(c, false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !hasIcon || hasCover {
		t.Fatalf("hasIcon=%v hasCover=%v want true,false", hasIcon, hasCover)
	}
	m, ok := icon.(map[string]any)
	if !ok {
		t.Fatalf("icon not a map: %T", icon)
	}
	if m["type"] != "emoji" || m["emoji"] != "💰" {
		t.Errorf("icon = %v", m)
	}
	if cover != nil {
		t.Errorf("cover should be nil, got %v", cover)
	}
}

func TestBuildIconCover_IconURLAndCover(t *testing.T) {
	c := newIconCoverFlagSet(t)
	_ = c.Flags().Set("icon-url", "https://example.com/i.png")
	_ = c.Flags().Set("cover-url", "https://example.com/c.jpg")

	icon, cover, hasIcon, hasCover, err := buildIconCover(c, false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !hasIcon || !hasCover {
		t.Fatalf("hasIcon=%v hasCover=%v want both true", hasIcon, hasCover)
	}
	im := icon.(map[string]any)
	if im["type"] != "external" {
		t.Errorf("icon type = %v", im["type"])
	}
	if im["external"].(map[string]any)["url"] != "https://example.com/i.png" {
		t.Errorf("icon url wrong: %v", im["external"])
	}
	cm := cover.(map[string]any)
	if cm["type"] != "external" {
		t.Errorf("cover type = %v", cm["type"])
	}
	if cm["external"].(map[string]any)["url"] != "https://example.com/c.jpg" {
		t.Errorf("cover url wrong: %v", cm["external"])
	}
}

func TestBuildIconCover_InvalidURL(t *testing.T) {
	c := newIconCoverFlagSet(t)
	_ = c.Flags().Set("icon-url", "not-a-url")

	_, _, _, _, err := buildIconCover(c, false)
	if err == nil {
		t.Fatal("expected validation error")
	}
	cliErr, ok := err.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("err type = %T", err)
	}
	if cliErr.Code != clierrors.CodeInvalidRequest {
		t.Errorf("code = %q", cliErr.Code)
	}
}

func TestBuildIconCover_CoverWrongScheme(t *testing.T) {
	c := newIconCoverFlagSet(t)
	_ = c.Flags().Set("cover-url", "ftp://example.com/c.jpg")

	_, _, _, _, err := buildIconCover(c, false)
	if err == nil {
		t.Fatal("expected scheme error")
	}
}

func TestBuildIconCover_ClearWithNone(t *testing.T) {
	c := newIconCoverFlagSet(t)
	_ = c.Flags().Set("icon-emoji", "none")
	_ = c.Flags().Set("cover-url", "none")

	icon, cover, hasIcon, hasCover, err := buildIconCover(c, true)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !hasIcon || !hasCover {
		t.Fatalf("hasIcon/hasCover both should be true")
	}
	if icon != nil {
		t.Errorf("icon should be nil for clear, got %v", icon)
	}
	if cover != nil {
		t.Errorf("cover should be nil for clear, got %v", cover)
	}
}

func TestBuildIconCover_NoneNotAllowedOnCreate(t *testing.T) {
	// On create (allowClear=false), "none" should be treated as a literal
	// emoji — current behaviour. Just confirm no panic and field set.
	c := newIconCoverFlagSet(t)
	_ = c.Flags().Set("icon-emoji", "none")

	icon, _, hasIcon, _, err := buildIconCover(c, false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !hasIcon {
		t.Fatal("expected hasIcon true")
	}
	m := icon.(map[string]any)
	if m["emoji"] != "none" {
		t.Errorf("expected literal 'none', got %v", m["emoji"])
	}
}

func TestPageCreate_IconEmojiAndURLMutuallyExclusive(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterPageCommands(root)

	root.SetArgs([]string{"page", "create", "-d", "abc", "--icon-emoji", "💰", "--icon-url", "https://x/y.png"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	err := root.Execute()
	if err == nil {
		t.Fatal("expected mutually-exclusive error")
	}
	combined := strings.ToLower(err.Error() + " " + out.String())
	if !strings.Contains(combined, "icon-emoji") || !strings.Contains(combined, "icon-url") {
		t.Errorf("expected error mentioning icon-emoji and icon-url, got %v / %s", err, out.String())
	}
}

// --- page create --template tests ---

func newTemplateFlagSet(t *testing.T, template, timezone string) *cobra.Command {
	t.Helper()
	c := &cobra.Command{Use: "test"}
	c.Flags().String("template", "", "")
	c.Flags().String("template-timezone", "", "")
	_ = c.Flags().Set("template", template)
	_ = c.Flags().Set("template-timezone", timezone)
	return c
}

func TestBuildTemplateParam(t *testing.T) {
	const dashed = "11111111-2222-3333-4444-555555555555"
	const bare = "11111111222233334444555555555555"

	tests := []struct {
		name     string
		template string
		timezone string
		wantSet  bool
		want     map[string]any
		wantErr  bool
	}{
		{name: "empty omits key", template: "", wantSet: false},
		{name: "none omits key", template: "none", wantSet: false},
		{name: "default", template: "default", wantSet: true, want: map[string]any{"type": "default"}},
		{
			name: "default with timezone", template: "default", timezone: "America/New_York", wantSet: true,
			want: map[string]any{"type": "default", "timezone": "America/New_York"},
		},
		// resolveID normalises every accepted form to a dashed UUID.
		{
			name: "bare uuid", template: bare, wantSet: true,
			want: map[string]any{"type": "template_id", "template_id": dashed},
		},
		{
			name: "dashed uuid", template: dashed, wantSet: true,
			want: map[string]any{"type": "template_id", "template_id": dashed},
		},
		{
			name: "notion url", template: "https://www.notion.so/My-Template-" + bare, wantSet: true,
			want: map[string]any{"type": "template_id", "template_id": dashed},
		},
		{name: "timezone without template errors", template: "", timezone: "America/New_York", wantErr: true},
		{name: "invalid template id errors", template: "not-an-id", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, set, err := buildTemplateParam(newTemplateFlagSet(t, tc.template, tc.timezone))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if set != tc.wantSet {
				t.Fatalf("set = %v, want %v", set, tc.wantSet)
			}
			if !tc.wantSet {
				if got != nil {
					t.Errorf("expected nil param, got %v", got)
				}
				return
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Errorf("param = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestPageCreate_TemplateAndFilePathMutuallyExclusive(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterPageCommands(root)

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"page", "create", "-d", testPageID, "--template", "default", "--file-path", "notes.md"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected mutually-exclusive error")
	}
	combined := strings.ToLower(err.Error() + " " + out.String())
	if !strings.Contains(combined, "template") || !strings.Contains(combined, "file-path") {
		t.Errorf("expected error mentioning template and file-path, got %v / %s", err, out.String())
	}
}

func TestPageCreate_TemplateRequiresDataSourceParent(t *testing.T) {
	_, cleanup := testPageServer(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("no request should be sent")
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer cleanup()

	_, _, err := runPageRoot(t, "page", "create", "-p", testPageID, "--template", "default")
	if err == nil {
		t.Fatal("expected error when --template is used with -p")
	}
	cliErr, ok := err.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("expected NotionCLIError, got %T: %v", err, err)
	}
	if cliErr.Code != clierrors.CodeInvalidRequest {
		t.Errorf("code = %q, want %q", cliErr.Code, clierrors.CodeInvalidRequest)
	}
}

func TestPageCreate_TemplateSendsDataSourceParent(t *testing.T) {
	const dsID = "22222222222222222222222222222222"
	const dsIDDashed = "22222222-2222-2222-2222-222222222222"
	var got map[string]any

	_, cleanup := testPageServer(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "page", "id": testPageID})
	})
	defer cleanup()

	if _, _, err := runPageRoot(t, "page", "create", "-d", dsID, "--template", "default"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parent, ok := got["parent"].(map[string]any)
	if !ok {
		t.Fatalf("parent missing from body: %v", got)
	}
	if parent["type"] != "data_source_id" || parent["data_source_id"] != dsIDDashed {
		t.Errorf("parent = %v, want data_source_id %s", parent, dsIDDashed)
	}
	if !reflect.DeepEqual(got["template"], map[string]any{"type": "default"}) {
		t.Errorf("template = %v, want {type: default}", got["template"])
	}
}

func TestPageCreate_WaitRequiresTemplate(t *testing.T) {
	_, cleanup := testPageServer(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("no request should be sent")
		w.WriteHeader(http.StatusInternalServerError)
	})
	defer cleanup()

	_, _, err := runPageRoot(t, "page", "create", "-d", testPageID, "--wait")
	if err == nil {
		t.Fatal("expected error when --wait is used without --template")
	}
	cliErr, ok := err.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("expected NotionCLIError, got %T: %v", err, err)
	}
	if cliErr.Code != clierrors.CodeInvalidRequest {
		t.Errorf("code = %q, want %q", cliErr.Code, clierrors.CodeInvalidRequest)
	}
}

func TestPageCreate_WaitPollsUntilPopulated(t *testing.T) {
	var mu sync.Mutex
	var childCalls int

	_, cleanup := testPageServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !strings.HasSuffix(r.URL.Path, "/children") {
			_ = json.NewEncoder(w).Encode(map[string]any{"object": "page", "id": testPageID})
			return
		}
		mu.Lock()
		childCalls++
		first := childCalls == 1
		mu.Unlock()

		results := []any{map[string]any{"object": "block", "type": "paragraph"}}
		if first {
			results = []any{}
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "results": results})
	})
	defer cleanup()

	// The printer writes to os.Stdout, not the cobra buffer, so capture it.
	var err error
	stdout := captureStdout(t, func() {
		_, _, err = runPageRoot(t, "page", "create", "-d", "22222222222222222222222222222222",
			"--template", "default", "--wait", "--wait-timeout", "10s", "--json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Polling stops as soon as children are non-empty: one empty response,
	// then one populated response — never the full timeout.
	mu.Lock()
	defer mu.Unlock()
	if childCalls != 2 {
		t.Errorf("children polled %d times, want exactly 2", childCalls)
	}
	// Waiting must still print the created page.
	if !strings.Contains(stdout, testPageID) {
		t.Errorf("expected page id in stdout, got %s", stdout)
	}
}

func TestPageUpdate_FlagsRegistered(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterPageCommands(root)

	updateCmd, _, err := root.Find([]string{"page", "update"})
	if err != nil {
		t.Fatalf("page update not found: %v", err)
	}
	for _, name := range []string{"icon-emoji", "icon-url", "cover-url"} {
		if updateCmd.Flag(name) == nil {
			t.Errorf("flag %q not registered on page update", name)
		}
	}
}

// --- page trash tests ---

func TestPageTrash_YesFlagRegistered(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterPageCommands(root)

	trashCmd, _, err := root.Find([]string{"page", "trash"})
	if err != nil {
		t.Fatalf("page trash not found: %v", err)
	}
	if trashCmd.Flag("yes") == nil {
		t.Error("flag \"yes\" not registered on page trash")
	}
}

func TestPageTrash_RequiresYesInNonTTY(t *testing.T) {
	// Override isTerminal to simulate a non-interactive environment.
	original := isTerminal
	isTerminal = func() bool { return false }
	t.Cleanup(func() { isTerminal = original })

	root := &cobra.Command{Use: "notion-cli"}
	RegisterPageCommands(root)

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"page", "trash", "page-abc"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when --yes is missing in non-TTY")
	}
	cliErr, ok := err.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("expected NotionCLIError, got %T: %v", err, err)
	}
	if cliErr.Code != clierrors.CodeMissingRequired {
		t.Errorf("code = %q, want %q", cliErr.Code, clierrors.CodeMissingRequired)
	}
}

// --- page move tests ---

func TestPageMove_FlagsRegistered(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterPageCommands(root)

	moveCmd, _, err := root.Find([]string{"page", "move"})
	if err != nil {
		t.Fatalf("page move not found: %v", err)
	}
	for _, name := range []string{"parent", "data-source", "workspace"} {
		if moveCmd.Flag(name) == nil {
			t.Errorf("flag %q not registered on page move", name)
		}
	}
}

func TestPageMove_RequiresAtLeastOneTarget(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterPageCommands(root)

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"page", "move", "page-abc"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected error when no target flag is provided")
	}
	cliErr, ok := err.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("expected NotionCLIError, got %T: %v", err, err)
	}
	if cliErr.Code != clierrors.CodeMissingRequired {
		t.Errorf("code = %q, want %q", cliErr.Code, clierrors.CodeMissingRequired)
	}
}

func TestPageMove_MutuallyExclusiveFlags(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterPageCommands(root)

	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"page", "move", "page-abc", "--parent", "p1", "--data-source", "ds1"})

	err := root.Execute()
	if err == nil {
		t.Fatal("expected mutually-exclusive error")
	}
	combined := strings.ToLower(err.Error() + " " + out.String())
	if !strings.Contains(combined, "parent") || !strings.Contains(combined, "data-source") {
		t.Errorf("expected error mentioning parent and data-source, got: %v / %s", err, out.String())
	}
}

// ---------------------------------------------------------------------------
// Page command integration tests (with test server)
// ---------------------------------------------------------------------------

func TestPageRetrieve_Success(t *testing.T) {
	_, cleanup := testPageServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "page", "id": testPageID})
	})
	defer cleanup()

	_, _, err := runPageRoot(t, "page", "retrieve", testPageID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPageRetrieve_RequiresArg(t *testing.T) {
	_, _, err := runPageRoot(t, "page", "retrieve")
	if err == nil {
		t.Fatal("expected error when no page_id given")
	}
}

func TestPageUpdate_Success(t *testing.T) {
	_, cleanup := testPageServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "page", "id": testPageID})
	})
	defer cleanup()

	_, _, err := runPageRoot(t, "page", "update", testPageID, "--archived")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPageTrash_Success(t *testing.T) {
	_, cleanup := testPageServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "page", "id": testPageID, "in_trash": true})
	})
	defer cleanup()

	_, _, err := runPageRoot(t, "page", "trash", testPageID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPageRestore_Success(t *testing.T) {
	_, cleanup := testPageServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "page", "id": testPageID, "in_trash": false})
	})
	defer cleanup()

	_, _, err := runPageRoot(t, "page", "restore", testPageID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPagePropertyItem_Success(t *testing.T) {
	_, cleanup := testPageServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "property_item", "type": "title"})
	})
	defer cleanup()

	_, _, err := runPageRoot(t, "page", "property-item", testPageID, "Name")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPageCreate_NoParent(t *testing.T) {
	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Setenv("NOTION_TOKEN", "secret_test_token")
	t.Cleanup(func() {
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})
	// Both --parent-page-id and --parent-data-source-id missing → error.
	_, _, err := runPageRoot(t, "page", "create")
	if err == nil {
		t.Fatal("expected error when no parent given")
	}
}

func TestPageCreate_WithParentPage(t *testing.T) {
	_, cleanup := testPageServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "page", "id": testPageID})
	})
	defer cleanup()

	_, _, err := runPageRoot(t, "page", "create", "--parent-page-id", testPageID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPageTrash_WithYesFlag(t *testing.T) {
	_, cleanup := testPageServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "page", "id": testPageID, "in_trash": true})
	})
	defer cleanup()

	_, _, err := runPageRoot(t, "page", "trash", testPageID, "--yes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPageMove_WithParent(t *testing.T) {
	const parentID = "22222222222222222222222222222222"
	_, cleanup := testPageServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "page", "id": testPageID})
	})
	defer cleanup()

	_, _, err := runPageRoot(t, "page", "move", testPageID, "--parent", parentID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPageMove_WithDataSource(t *testing.T) {
	const dsID = "22222222222222222222222222222222"
	_, cleanup := testPageServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "page", "id": testPageID})
	})
	defer cleanup()

	_, _, err := runPageRoot(t, "page", "move", testPageID, "--data-source", dsID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
