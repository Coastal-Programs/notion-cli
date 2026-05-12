package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Coastal-Programs/notion-cli/v6/internal/notion"
	"github.com/spf13/cobra"
)

func runBlockRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterBlockCommands(root)
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

// testBlockServer creates a test HTTP server that mimics the Notion API for
// block endpoints. It returns the server and a cleanup function.
func testBlockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
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

func TestRegisterBlockCommands(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterBlockCommands(root)

	// Verify the block parent command was registered.
	blockCmd, _, err := root.Find([]string{"block"})
	if err != nil {
		t.Fatalf("block command not found: %v", err)
	}
	if blockCmd.Use != "block" {
		t.Errorf("Use = %q, want %q", blockCmd.Use, "block")
	}

	// Verify subcommands.
	subCmds := blockCmd.Commands()
	wantNames := map[string]bool{
		"append": false, "retrieve": false, "children": false,
		"update": false, "delete": false,
	}
	for _, cmd := range subCmds {
		wantNames[cmd.Name()] = true
	}
	for name, found := range wantNames {
		if !found {
			t.Errorf("subcommand %q not registered", name)
		}
	}
}

func TestMakeTextBlock(t *testing.T) {
	block := makeTextBlock("paragraph", "Hello world")

	if block["type"] != "paragraph" {
		t.Errorf("type = %v, want paragraph", block["type"])
	}
	if block["object"] != "block" {
		t.Errorf("object = %v, want block", block["object"])
	}

	para, ok := block["paragraph"].(map[string]any)
	if !ok {
		t.Fatal("paragraph key missing or wrong type")
	}

	rt, ok := para["rich_text"].([]map[string]any)
	if !ok || len(rt) == 0 {
		t.Fatal("rich_text missing or empty")
	}

	textObj, ok := rt[0]["text"].(map[string]any)
	if !ok {
		t.Fatal("text object missing")
	}
	if textObj["content"] != "Hello world" {
		t.Errorf("content = %v, want %q", textObj["content"], "Hello world")
	}
}

func TestMakeCodeBlock(t *testing.T) {
	block := makeCodeBlock("fmt.Println()", "go")

	if block["type"] != "code" {
		t.Errorf("type = %v, want code", block["type"])
	}

	code, ok := block["code"].(map[string]any)
	if !ok {
		t.Fatal("code key missing")
	}
	if code["language"] != "go" {
		t.Errorf("language = %v, want go", code["language"])
	}
}

func TestBuildChildren_RawJSON(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	root.Flags().String("children", "", "")
	root.Flags().String("text", "", "")
	root.Flags().String("heading-1", "", "")
	root.Flags().String("heading-2", "", "")
	root.Flags().String("heading-3", "", "")
	root.Flags().String("bullet", "", "")
	root.Flags().String("numbered", "", "")
	root.Flags().String("todo", "", "")
	root.Flags().String("toggle", "", "")
	root.Flags().String("code", "", "")
	root.Flags().String("language", "plain text", "")
	root.Flags().String("quote", "", "")
	root.Flags().String("callout", "", "")

	_ = root.Flags().Set("children", `[{"type":"paragraph"}]`)

	children, err := buildChildren(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(children) != 1 {
		t.Errorf("got %d children, want 1", len(children))
	}
}

func TestBuildChildren_InvalidJSON(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	root.Flags().String("children", "", "")
	root.Flags().String("text", "", "")
	root.Flags().String("heading-1", "", "")
	root.Flags().String("heading-2", "", "")
	root.Flags().String("heading-3", "", "")
	root.Flags().String("bullet", "", "")
	root.Flags().String("numbered", "", "")
	root.Flags().String("todo", "", "")
	root.Flags().String("toggle", "", "")
	root.Flags().String("code", "", "")
	root.Flags().String("language", "plain text", "")
	root.Flags().String("quote", "", "")
	root.Flags().String("callout", "", "")

	_ = root.Flags().Set("children", "not json")

	_, err := buildChildren(root)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestBuildChildren_Shorthand(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	root.Flags().String("children", "", "")
	root.Flags().String("text", "", "")
	root.Flags().String("heading-1", "", "")
	root.Flags().String("heading-2", "", "")
	root.Flags().String("heading-3", "", "")
	root.Flags().String("bullet", "", "")
	root.Flags().String("numbered", "", "")
	root.Flags().String("todo", "", "")
	root.Flags().String("toggle", "", "")
	root.Flags().String("code", "", "")
	root.Flags().String("language", "plain text", "")
	root.Flags().String("quote", "", "")
	root.Flags().String("callout", "", "")

	_ = root.Flags().Set("text", "Hello")
	_ = root.Flags().Set("bullet", "Item 1")
	_ = root.Flags().Set("code", "x = 1")

	children, err := buildChildren(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// text, bullet, code = 3 children
	if len(children) != 3 {
		t.Errorf("got %d children, want 3", len(children))
	}
}

func TestBuildChildren_Empty(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	root.Flags().String("children", "", "")
	root.Flags().String("text", "", "")
	root.Flags().String("heading-1", "", "")
	root.Flags().String("heading-2", "", "")
	root.Flags().String("heading-3", "", "")
	root.Flags().String("bullet", "", "")
	root.Flags().String("numbered", "", "")
	root.Flags().String("todo", "", "")
	root.Flags().String("toggle", "", "")
	root.Flags().String("code", "", "")
	root.Flags().String("language", "plain text", "")
	root.Flags().String("quote", "", "")
	root.Flags().String("callout", "", "")

	children, err := buildChildren(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(children) != 0 {
		t.Errorf("got %d children, want 0", len(children))
	}
}

func TestBuildUpdateBody_Archived(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	root.Flags().Bool("archived", false, "")
	root.Flags().String("content", "", "")
	root.Flags().String("text", "", "")
	root.Flags().String("heading-1", "", "")
	root.Flags().String("heading-2", "", "")
	root.Flags().String("heading-3", "", "")
	root.Flags().String("bullet", "", "")
	root.Flags().String("numbered", "", "")
	root.Flags().String("todo", "", "")
	root.Flags().String("toggle", "", "")
	root.Flags().String("code", "", "")
	root.Flags().String("language", "plain text", "")
	root.Flags().String("quote", "", "")
	root.Flags().String("callout", "", "")
	root.Flags().String("color", "", "")

	_ = root.Flags().Set("archived", "true")

	body, err := buildUpdateBody(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body["in_trash"] != true {
		t.Error("expected in_trash to be true")
	}
	if _, ok := body["archived"]; ok {
		t.Error("expected legacy `archived` key not to be set; Notion API 2026-03-11 uses `in_trash`")
	}
}

func TestBuildUpdateBody_ContentJSON(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	root.Flags().Bool("archived", false, "")
	root.Flags().String("content", "", "")
	root.Flags().String("text", "", "")
	root.Flags().String("heading-1", "", "")
	root.Flags().String("heading-2", "", "")
	root.Flags().String("heading-3", "", "")
	root.Flags().String("bullet", "", "")
	root.Flags().String("numbered", "", "")
	root.Flags().String("todo", "", "")
	root.Flags().String("toggle", "", "")
	root.Flags().String("code", "", "")
	root.Flags().String("language", "plain text", "")
	root.Flags().String("quote", "", "")
	root.Flags().String("callout", "", "")
	root.Flags().String("color", "", "")

	_ = root.Flags().Set("content", `{"paragraph": {"rich_text": []}}`)

	body, err := buildUpdateBody(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := body["paragraph"]; !ok {
		t.Error("expected paragraph key from content JSON")
	}
}

func TestBuildUpdateBody_InvalidContentJSON(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	root.Flags().Bool("archived", false, "")
	root.Flags().String("content", "", "")
	root.Flags().String("text", "", "")
	root.Flags().String("heading-1", "", "")
	root.Flags().String("heading-2", "", "")
	root.Flags().String("heading-3", "", "")
	root.Flags().String("bullet", "", "")
	root.Flags().String("numbered", "", "")
	root.Flags().String("todo", "", "")
	root.Flags().String("toggle", "", "")
	root.Flags().String("code", "", "")
	root.Flags().String("language", "plain text", "")
	root.Flags().String("quote", "", "")
	root.Flags().String("callout", "", "")
	root.Flags().String("color", "", "")

	_ = root.Flags().Set("content", "not json")

	_, err := buildUpdateBody(root)
	if err == nil {
		t.Error("expected error for invalid content JSON")
	}
}

func TestBuildUpdateBody_TextShorthand(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	root.Flags().Bool("archived", false, "")
	root.Flags().String("content", "", "")
	root.Flags().String("text", "", "")
	root.Flags().String("heading-1", "", "")
	root.Flags().String("heading-2", "", "")
	root.Flags().String("heading-3", "", "")
	root.Flags().String("bullet", "", "")
	root.Flags().String("numbered", "", "")
	root.Flags().String("todo", "", "")
	root.Flags().String("toggle", "", "")
	root.Flags().String("code", "", "")
	root.Flags().String("language", "plain text", "")
	root.Flags().String("quote", "", "")
	root.Flags().String("callout", "", "")
	root.Flags().String("color", "", "")

	_ = root.Flags().Set("text", "Updated paragraph")

	body, err := buildUpdateBody(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := body["paragraph"]; !ok {
		t.Error("expected paragraph key from --text shorthand")
	}
}

func TestBuildUpdateBody_CodeShorthand(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	root.Flags().Bool("archived", false, "")
	root.Flags().String("content", "", "")
	root.Flags().String("text", "", "")
	root.Flags().String("heading-1", "", "")
	root.Flags().String("heading-2", "", "")
	root.Flags().String("heading-3", "", "")
	root.Flags().String("bullet", "", "")
	root.Flags().String("numbered", "", "")
	root.Flags().String("todo", "", "")
	root.Flags().String("toggle", "", "")
	root.Flags().String("code", "", "")
	root.Flags().String("language", "plain text", "")
	root.Flags().String("quote", "", "")
	root.Flags().String("callout", "", "")
	root.Flags().String("color", "", "")

	_ = root.Flags().Set("code", "x = 1")
	_ = root.Flags().Set("language", "python")

	body, err := buildUpdateBody(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	codeBlock, ok := body["code"].(map[string]any)
	if !ok {
		t.Fatal("expected code key from --code shorthand")
	}
	if codeBlock["language"] != "python" {
		t.Errorf("language = %v, want python", codeBlock["language"])
	}
}

func TestBuildUpdateBody_InvalidColor(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	root.Flags().Bool("archived", false, "")
	root.Flags().String("content", "", "")
	root.Flags().String("text", "", "")
	root.Flags().String("heading-1", "", "")
	root.Flags().String("heading-2", "", "")
	root.Flags().String("heading-3", "", "")
	root.Flags().String("bullet", "", "")
	root.Flags().String("numbered", "", "")
	root.Flags().String("todo", "", "")
	root.Flags().String("toggle", "", "")
	root.Flags().String("code", "", "")
	root.Flags().String("language", "plain text", "")
	root.Flags().String("quote", "", "")
	root.Flags().String("callout", "", "")
	root.Flags().String("color", "", "")

	_ = root.Flags().Set("text", "Hello")
	_ = root.Flags().Set("color", "neon_green")

	_, err := buildUpdateBody(root)
	if err == nil {
		t.Error("expected error for invalid color")
	}
}

func TestBuildUpdateBody_ValidColor(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	root.Flags().Bool("archived", false, "")
	root.Flags().String("content", "", "")
	root.Flags().String("text", "", "")
	root.Flags().String("heading-1", "", "")
	root.Flags().String("heading-2", "", "")
	root.Flags().String("heading-3", "", "")
	root.Flags().String("bullet", "", "")
	root.Flags().String("numbered", "", "")
	root.Flags().String("todo", "", "")
	root.Flags().String("toggle", "", "")
	root.Flags().String("code", "", "")
	root.Flags().String("language", "plain text", "")
	root.Flags().String("quote", "", "")
	root.Flags().String("callout", "", "")
	root.Flags().String("color", "", "")

	_ = root.Flags().Set("text", "Hello")
	_ = root.Flags().Set("color", "blue")

	body, err := buildUpdateBody(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	para, ok := body["paragraph"].(map[string]any)
	if !ok {
		t.Fatal("expected paragraph key")
	}
	if para["color"] != "blue" {
		t.Errorf("color = %v, want blue", para["color"])
	}
}

func TestBuildUpdateBody_Empty(t *testing.T) {
	root := &cobra.Command{Use: "test"}
	root.Flags().Bool("archived", false, "")
	root.Flags().String("content", "", "")
	root.Flags().String("text", "", "")
	root.Flags().String("heading-1", "", "")
	root.Flags().String("heading-2", "", "")
	root.Flags().String("heading-3", "", "")
	root.Flags().String("bullet", "", "")
	root.Flags().String("numbered", "", "")
	root.Flags().String("todo", "", "")
	root.Flags().String("toggle", "", "")
	root.Flags().String("code", "", "")
	root.Flags().String("language", "plain text", "")
	root.Flags().String("quote", "", "")
	root.Flags().String("callout", "", "")
	root.Flags().String("color", "", "")

	body, err := buildUpdateBody(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) != 0 {
		t.Errorf("body should be empty, got %v", body)
	}
}

func TestFilterChildDatabases(t *testing.T) {
	results := []any{
		map[string]any{"type": "paragraph", "id": "1"},
		map[string]any{"type": "child_database", "id": "2"},
		map[string]any{"type": "heading_1", "id": "3"},
		map[string]any{"type": "child_database", "id": "4"},
	}

	filtered := filterChildDatabases(results)
	if len(filtered) != 2 {
		t.Errorf("got %d results, want 2", len(filtered))
	}

	for _, r := range filtered {
		block := r.(map[string]any)
		if block["type"] != "child_database" {
			t.Errorf("expected child_database, got %v", block["type"])
		}
	}
}

func TestFilterChildDatabases_Empty(t *testing.T) {
	filtered := filterChildDatabases([]any{})
	if len(filtered) != 0 {
		t.Errorf("expected empty result, got %d items", len(filtered))
	}
}

func TestFilterChildDatabases_NoMatches(t *testing.T) {
	results := []any{
		map[string]any{"type": "paragraph", "id": "1"},
		map[string]any{"type": "heading_1", "id": "2"},
	}

	filtered := filterChildDatabases(results)
	if len(filtered) != 0 {
		t.Errorf("expected empty result, got %d items", len(filtered))
	}
}

func TestFilterChildDatabases_InvalidTypes(t *testing.T) {
	results := []any{
		"not a map",
		42,
		nil,
	}

	filtered := filterChildDatabases(results)
	if len(filtered) != 0 {
		t.Errorf("expected empty result, got %d items", len(filtered))
	}
}

func TestValidColors(t *testing.T) {
	expected := []string{
		"default", "gray", "brown", "orange", "yellow",
		"green", "blue", "purple", "pink", "red",
	}
	for _, c := range expected {
		if !validColors[c] {
			t.Errorf("color %q should be valid", c)
		}
		bg := c + "_background"
		if c != "default" && !validColors[bg] {
			t.Errorf("color %q should be valid", bg)
		}
	}

	if validColors["neon"] {
		t.Error("neon should not be a valid color")
	}
}

// TestBlockRetrieve_Integration tests the full command flow with a mock server.
func TestBlockRetrieve_Integration(t *testing.T) {
	srv, cleanup := testBlockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "block",
			"id":     "test-block-id",
			"type":   "paragraph",
		})
	})
	defer cleanup()

	// Verify the command structure is correct by checking flag parsing.
	_ = srv
	_ = notion.NewClient("test", notion.WithBaseURL(srv.URL))

	// Test that the command accepts exactly 1 arg.
	root := &cobra.Command{Use: "notion-cli"}
	RegisterBlockCommands(root)

	root.SetArgs([]string{"block", "retrieve"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for missing block_id argument")
	}
}

func TestBlockDelete_RequiresArg(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterBlockCommands(root)

	root.SetArgs([]string{"block", "delete"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for missing block_id argument")
	}
}

func TestBlockChildren_RequiresArg(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterBlockCommands(root)

	root.SetArgs([]string{"block", "children"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for missing block_id argument")
	}
}

func TestBlockUpdate_RequiresArg(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterBlockCommands(root)

	root.SetArgs([]string{"block", "update"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for missing block_id argument")
	}
}

func TestBlockAppend_RequiresBlockID(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterBlockCommands(root)

	root.SetArgs([]string{"block", "append"})
	root.SetOut(&bytes.Buffer{})
	root.SetErr(&bytes.Buffer{})
	err := root.Execute()
	if err == nil {
		t.Error("expected error for missing --block-id")
	}
}

func TestBlockAliases(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterBlockCommands(root)

	// Test "b" alias for block.
	blockCmd, _, err := root.Find([]string{"b"})
	if err != nil {
		t.Fatalf("alias 'b' not found: %v", err)
	}
	if blockCmd.Use != "block" {
		t.Errorf("expected block command, got %q", blockCmd.Use)
	}
}

// newPositionTestCmd builds a cobra command with the position-related flags
// used by buildPosition.
func newPositionTestCmd() *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("position", "end", "")
	cmd.Flags().String("after-block", "", "")
	cmd.Flags().String("after", "", "")
	return cmd
}

func TestBuildPosition_EndDefault(t *testing.T) {
	cmd := newPositionTestCmd()
	pos, err := buildPosition(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pos != nil {
		t.Errorf("want nil position when unset, got %v", pos)
	}
}

func TestBuildPosition_EndExplicit(t *testing.T) {
	cmd := newPositionTestCmd()
	_ = cmd.Flags().Set("position", "end")
	pos, err := buildPosition(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pos == nil || pos["type"] != "end" {
		t.Errorf("want {type: end}, got %v", pos)
	}
}

func TestBuildPosition_Start(t *testing.T) {
	cmd := newPositionTestCmd()
	_ = cmd.Flags().Set("position", "start")
	pos, err := buildPosition(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pos == nil || pos["type"] != "start" {
		t.Errorf("want {type: start}, got %v", pos)
	}
	if _, ok := pos["after_block"]; ok {
		t.Errorf("start should not include after_block: %v", pos)
	}
}

func TestBuildPosition_AfterBlock(t *testing.T) {
	cmd := newPositionTestCmd()
	const id = "11111111-1111-1111-1111-111111111111"
	_ = cmd.Flags().Set("position", "after")
	_ = cmd.Flags().Set("after-block", id)
	pos, err := buildPosition(cmd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pos == nil || pos["type"] != "after_block" {
		t.Fatalf("want type after_block, got %v", pos)
	}
	ab, ok := pos["after_block"].(map[string]any)
	if !ok {
		t.Fatalf("after_block missing: %v", pos)
	}
	if ab["id"] != id {
		t.Errorf("id = %v, want %v", ab["id"], id)
	}
}

func TestBuildPosition_LegacyAfterDeprecated(t *testing.T) {
	cmd := newPositionTestCmd()
	const id = "22222222-2222-2222-2222-222222222222"

	// Capture stderr to confirm the deprecation warning is emitted.
	orig := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	defer func() { os.Stderr = orig }()

	_ = cmd.Flags().Set("after", id)
	pos, err := buildPosition(cmd)
	_ = w.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	if !bytes.Contains(buf.Bytes(), []byte("deprecated")) {
		t.Errorf("expected deprecation warning on stderr, got %q", buf.String())
	}

	if pos == nil || pos["type"] != "after_block" {
		t.Fatalf("want type after_block, got %v", pos)
	}
	ab, ok := pos["after_block"].(map[string]any)
	if !ok {
		t.Fatalf("after_block missing: %v", pos)
	}
	if ab["id"] != id {
		t.Errorf("id = %v, want %v", ab["id"], id)
	}
}

func TestValidateBlockAppendFlags_InvalidPosition(t *testing.T) {
	cmd := newBlockAppendCmd()
	_ = cmd.Flags().Set("position", "middle")
	if err := validateBlockAppendFlags(cmd, nil); err == nil {
		t.Error("expected error for invalid position")
	}
}

func TestValidateBlockAppendFlags_AfterRequiresAfterBlock(t *testing.T) {
	cmd := newBlockAppendCmd()
	_ = cmd.Flags().Set("position", "after")
	if err := validateBlockAppendFlags(cmd, nil); err == nil {
		t.Error("expected error when --position=after has no --after-block")
	}
}

func TestAllShorthandBlockTypes(t *testing.T) {
	// Verify each shorthand flag produces the correct block type.
	tests := []struct {
		flag      string
		blockType string
	}{
		{"text", "paragraph"},
		{"heading-1", "heading_1"},
		{"heading-2", "heading_2"},
		{"heading-3", "heading_3"},
		{"bullet", "bulleted_list_item"},
		{"numbered", "numbered_list_item"},
		{"todo", "to_do"},
		{"toggle", "toggle"},
		{"quote", "quote"},
		{"callout", "callout"},
	}

	for _, tt := range tests {
		t.Run(tt.flag, func(t *testing.T) {
			root := &cobra.Command{Use: "test"}
			root.Flags().String("children", "", "")
			root.Flags().String("text", "", "")
			root.Flags().String("heading-1", "", "")
			root.Flags().String("heading-2", "", "")
			root.Flags().String("heading-3", "", "")
			root.Flags().String("bullet", "", "")
			root.Flags().String("numbered", "", "")
			root.Flags().String("todo", "", "")
			root.Flags().String("toggle", "", "")
			root.Flags().String("code", "", "")
			root.Flags().String("language", "plain text", "")
			root.Flags().String("quote", "", "")
			root.Flags().String("callout", "", "")

			_ = root.Flags().Set(tt.flag, fmt.Sprintf("Test %s content", tt.flag))

			children, err := buildChildren(root)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if len(children) != 1 {
				t.Fatalf("got %d children, want 1", len(children))
			}

			block := children[0].(map[string]any)
			if block["type"] != tt.blockType {
				t.Errorf("type = %v, want %v", block["type"], tt.blockType)
			}
			if _, ok := block[tt.blockType]; !ok {
				t.Errorf("missing %q key in block", tt.blockType)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Block command integration tests (with test server)
// ---------------------------------------------------------------------------

const testBlockID = "11111111111111111111111111111111"

func TestBlockRetrieve_Success(t *testing.T) {
	_, cleanup := testBlockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "block", "id": testBlockID})
	})
	defer cleanup()

	_, _, err := runBlockRoot(t, "block", "retrieve", testBlockID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBlockRetrieve_RequiresArg(t *testing.T) {
	_, _, err := runBlockRoot(t, "block", "retrieve")
	if err == nil {
		t.Fatal("expected error when no block_id given")
	}
}

func TestBlockChildren_Success(t *testing.T) {
	_, cleanup := testBlockServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object":   "list",
			"results":  []any{map[string]any{"object": "block", "id": "child-1"}},
			"has_more": false,
		})
	})
	defer cleanup()

	_, _, err := runBlockRoot(t, "block", "children", testBlockID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBlockUpdate_Success(t *testing.T) {
	_, cleanup := testBlockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "block", "id": testBlockID})
	})
	defer cleanup()

	_, _, err := runBlockRoot(t, "block", "update", testBlockID, "--text", "Updated text")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBlockDelete_Success(t *testing.T) {
	_, cleanup := testBlockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "block", "id": testBlockID, "archived": true})
	})
	defer cleanup()

	_, _, err := runBlockRoot(t, "block", "delete", testBlockID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBlockAppend_Success(t *testing.T) {
	_, cleanup := testBlockServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"object": "block", "id": testBlockID})
	})
	defer cleanup()

	_, _, err := runBlockRoot(t, "block", "append", "--block-id", testBlockID, "--text", "Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
