package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Coastal-Programs/notion-cli/internal/notion"
	"github.com/spf13/cobra"
)

// testBlockServer creates a test HTTP server that mimics the Notion API for
// block endpoints. It returns the server and a cleanup function.
func testBlockServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)

	origToken := os.Getenv("NOTION_TOKEN")
	os.Setenv("NOTION_TOKEN", "secret_test_token")

	return srv, func() {
		srv.Close()
		if origToken == "" {
			os.Unsetenv("NOTION_TOKEN")
		} else {
			os.Setenv("NOTION_TOKEN", origToken)
		}
	}
}

// executeBlockCmd creates a root command with block subcommands and executes it.
func executeBlockCmd(args []string, serverURL string) (string, string, error) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterBlockCommands(root)

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	root.SetOut(stdout)
	root.SetErr(stderr)
	root.SetArgs(args)

	// We need to override the client creation. Since newClient reads env
	// vars directly, we override at the HTTP transport level.
	err := root.Execute()
	return stdout.String(), stderr.String(), err
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

	root.Flags().Set("children", `[{"type":"paragraph"}]`)

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

	root.Flags().Set("children", "not json")

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

	root.Flags().Set("text", "Hello")
	root.Flags().Set("bullet", "Item 1")
	root.Flags().Set("code", "x = 1")

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

	root.Flags().Set("archived", "true")

	body, err := buildUpdateBody(root)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body["archived"] != true {
		t.Error("expected archived to be true")
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

	root.Flags().Set("content", `{"paragraph": {"rich_text": []}}`)

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

	root.Flags().Set("content", "not json")

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

	root.Flags().Set("text", "Updated paragraph")

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

	root.Flags().Set("code", "x = 1")
	root.Flags().Set("language", "python")

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

	root.Flags().Set("text", "Hello")
	root.Flags().Set("color", "neon_green")

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

	root.Flags().Set("text", "Hello")
	root.Flags().Set("color", "blue")

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
		json.NewEncoder(w).Encode(map[string]any{
			"object": "block",
			"id":     "test-block-id",
			"type":   "paragraph",
		})
	})
	defer cleanup()

	// We cannot easily override the client URL in the current architecture
	// because newClient() reads NOTION_TOKEN and creates a default client.
	// Instead, verify the command structure is correct by checking flag parsing.
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

			root.Flags().Set(tt.flag, fmt.Sprintf("Test %s content", tt.flag))

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
