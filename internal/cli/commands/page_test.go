package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
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
