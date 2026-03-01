package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/internal/errors"
	"github.com/Coastal-Programs/notion-cli/internal/notion"
	"github.com/Coastal-Programs/notion-cli/pkg/output"
	"github.com/spf13/cobra"
)

// validColors lists the block color values accepted by the Notion API.
var validColors = map[string]bool{
	"default": true, "gray": true, "brown": true, "orange": true,
	"yellow": true, "green": true, "blue": true, "purple": true,
	"pink": true, "red": true,
	"gray_background": true, "brown_background": true, "orange_background": true,
	"yellow_background": true, "green_background": true, "blue_background": true,
	"purple_background": true, "pink_background": true, "red_background": true,
}

// RegisterBlockCommands registers all block subcommands under root.
func RegisterBlockCommands(root *cobra.Command) {
	blockCmd := &cobra.Command{
		Use:     "block",
		Aliases: []string{"b"},
		Short:   "Block operations",
		Long:    "Append, retrieve, update, and delete Notion blocks.",
	}

	blockCmd.AddCommand(
		newBlockAppendCmd(),
		newBlockRetrieveCmd(),
		newBlockChildrenCmd(),
		newBlockUpdateCmd(),
		newBlockDeleteCmd(),
	)

	root.AddCommand(blockCmd)
}

// ---------------------------------------------------------------------------
// block append
// ---------------------------------------------------------------------------

func newBlockAppendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "append",
		Aliases: []string{"a"},
		Short:   "Append block children",
		Long:    "Append child blocks to a parent block. Use --children for raw JSON or shorthand flags for common block types.",
		RunE:    runBlockAppend,
	}

	cmd.Flags().StringP("block-id", "b", "", "Parent block ID (required)")
	cmd.MarkFlagRequired("block-id")

	// Raw children JSON.
	cmd.Flags().StringP("children", "c", "", "Block children as JSON array")

	// Shorthand block type flags.
	cmd.Flags().String("text", "", "Append a paragraph block with text")
	cmd.Flags().String("heading-1", "", "Append a heading_1 block")
	cmd.Flags().String("heading-2", "", "Append a heading_2 block")
	cmd.Flags().String("heading-3", "", "Append a heading_3 block")
	cmd.Flags().String("bullet", "", "Append a bulleted_list_item block")
	cmd.Flags().String("numbered", "", "Append a numbered_list_item block")
	cmd.Flags().String("todo", "", "Append a to_do block")
	cmd.Flags().String("toggle", "", "Append a toggle block")
	cmd.Flags().String("code", "", "Append a code block")
	cmd.Flags().String("language", "plain text", "Language for code block")
	cmd.Flags().String("quote", "", "Append a quote block")
	cmd.Flags().String("callout", "", "Append a callout block")
	cmd.Flags().StringP("after", "a", "", "Insert after this block ID")

	addOutputFlags(cmd)
	return cmd
}

func runBlockAppend(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	rawID, _ := cmd.Flags().GetString("block-id")
	blockID, err := resolveID(rawID)
	if err != nil {
		return handleError(cmd, err)
	}

	children, err := buildChildren(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	if len(children) == 0 {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "No block content specified",
			Suggestions: []string{
				"Use --children with a JSON array of blocks",
				"Or use shorthand: --text, --heading-1, --bullet, --todo, etc.",
			},
		})
	}

	body := map[string]any{
		"children": children,
	}

	if after, _ := cmd.Flags().GetString("after"); after != "" {
		afterID, err := resolveID(after)
		if err != nil {
			return handleError(cmd, err)
		}
		body["after"] = afterID
	}

	result, err := client.BlockChildrenAppend(cmd.Context(), blockID, body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "block append", start)
	return nil
}

// buildChildren constructs the children array from either --children JSON or
// shorthand flags.
func buildChildren(cmd *cobra.Command) ([]any, error) {
	if raw, _ := cmd.Flags().GetString("children"); raw != "" {
		var children []any
		if err := json.Unmarshal([]byte(raw), &children); err != nil {
			return nil, clierrors.InvalidJSON(fmt.Sprintf("--children: %s", err))
		}
		return children, nil
	}

	var children []any

	type shorthand struct {
		flag      string
		blockType string
	}

	shorthands := []shorthand{
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

	for _, sh := range shorthands {
		if text, _ := cmd.Flags().GetString(sh.flag); text != "" {
			block := makeTextBlock(sh.blockType, text)
			children = append(children, block)
		}
	}

	if code, _ := cmd.Flags().GetString("code"); code != "" {
		lang, _ := cmd.Flags().GetString("language")
		children = append(children, makeCodeBlock(code, lang))
	}

	return children, nil
}

// makeTextBlock creates a Notion block object with rich text content.
func makeTextBlock(blockType, text string) map[string]any {
	return map[string]any{
		"object": "block",
		"type":   blockType,
		blockType: map[string]any{
			"rich_text": []map[string]any{
				{"type": "text", "text": map[string]any{"content": text}},
			},
		},
	}
}

// makeCodeBlock creates a Notion code block.
func makeCodeBlock(code, language string) map[string]any {
	return map[string]any{
		"object": "block",
		"type":   "code",
		"code": map[string]any{
			"rich_text": []map[string]any{
				{"type": "text", "text": map[string]any{"content": code}},
			},
			"language": language,
		},
	}
}

// ---------------------------------------------------------------------------
// block retrieve
// ---------------------------------------------------------------------------

func newBlockRetrieveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "retrieve <block_id>",
		Aliases: []string{"r", "get"},
		Short:   "Retrieve a block",
		Long:    "Retrieve a single Notion block by ID.",
		Args:    cobra.ExactArgs(1),
		RunE:    runBlockRetrieve,
	}

	addOutputFlags(cmd)
	return cmd
}

func runBlockRetrieve(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	blockID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.BlockRetrieve(cmd.Context(), blockID)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "block retrieve", start)
	return nil
}

// ---------------------------------------------------------------------------
// block children (retrieve children)
// ---------------------------------------------------------------------------

func newBlockChildrenCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "children <block_id>",
		Aliases: []string{"c", "list"},
		Short:   "List block children",
		Long:    "List children of a block with pagination support.",
		Args:    cobra.ExactArgs(1),
		RunE:    runBlockChildren,
	}

	cmd.Flags().Int("page-size", 100, "Number of results per page")
	cmd.Flags().Bool("page-all", false, "Fetch all children (auto-paginate)")
	cmd.Flags().BoolP("show-databases", "d", false, "Filter to child_database blocks and enrich with data_source_id")
	addOutputFlags(cmd)

	return cmd
}

func runBlockChildren(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	blockID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	pageAll, _ := cmd.Flags().GetBool("page-all")
	showDBs, _ := cmd.Flags().GetBool("show-databases")
	pageSize, _ := cmd.Flags().GetInt("page-size")

	var allResults []any
	qp := notion.QueryParams{}
	if pageSize > 0 {
		qp.PageSize = pageSize
	}

	pageCount := 0
	for {
		pageCount++
		if pageCount > maxPaginationPages {
			fmt.Fprintf(os.Stderr, "Warning: reached maximum pagination limit (%d pages)\n", maxPaginationPages)
			break
		}

		result, err := client.BlockChildrenList(cmd.Context(), blockID, qp)
		if err != nil {
			return handleError(cmd, err)
		}

		results, _ := result["results"].([]any)
		allResults = append(allResults, results...)

		hasMore, _ := result["has_more"].(bool)
		nextCursor, _ := result["next_cursor"].(string)

		if !pageAll || !hasMore || nextCursor == "" {
			break
		}
		qp.StartCursor = nextCursor
	}

	// Filter to child_database blocks if --show-databases.
	if showDBs {
		allResults = filterChildDatabases(allResults)
	}

	data := map[string]any{
		"results":      allResults,
		"result_count": len(allResults),
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "block children", start)
	return nil
}

// filterChildDatabases keeps only child_database blocks.
func filterChildDatabases(results []any) []any {
	var filtered []any
	for _, r := range results {
		block, ok := r.(map[string]any)
		if !ok {
			continue
		}
		blockType, _ := block["type"].(string)
		if blockType == "child_database" {
			filtered = append(filtered, block)
		}
	}
	return filtered
}

// ---------------------------------------------------------------------------
// block update
// ---------------------------------------------------------------------------

func newBlockUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <block_id>",
		Aliases: []string{"u"},
		Short:   "Update a block",
		Long:    "Update a Notion block's content, type, or archived status.",
		Args:    cobra.ExactArgs(1),
		RunE:    runBlockUpdate,
	}

	cmd.Flags().BoolP("archived", "a", false, "Archive the block")
	cmd.Flags().StringP("content", "c", "", "Block content as JSON")

	// Shorthand text flags.
	cmd.Flags().String("text", "", "Update block as paragraph with text")
	cmd.Flags().String("heading-1", "", "Update block as heading_1")
	cmd.Flags().String("heading-2", "", "Update block as heading_2")
	cmd.Flags().String("heading-3", "", "Update block as heading_3")
	cmd.Flags().String("bullet", "", "Update block as bulleted_list_item")
	cmd.Flags().String("numbered", "", "Update block as numbered_list_item")
	cmd.Flags().String("todo", "", "Update block as to_do")
	cmd.Flags().String("toggle", "", "Update block as toggle")
	cmd.Flags().String("code", "", "Update block as code")
	cmd.Flags().String("language", "plain text", "Language for code block")
	cmd.Flags().String("quote", "", "Update block as quote")
	cmd.Flags().String("callout", "", "Update block as callout")
	cmd.Flags().String("color", "", "Block color (default, gray, brown, orange, yellow, green, blue, purple, pink, red)")

	addOutputFlags(cmd)
	return cmd
}

func runBlockUpdate(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	blockID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	body, err := buildUpdateBody(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	if len(body) == 0 {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "No update specified",
			Suggestions: []string{
				"Use --content with JSON to update block content",
				"Or use shorthand: --text, --heading-1, --bullet, etc.",
				"Or use --archived to archive/unarchive the block",
			},
		})
	}

	result, err := client.BlockUpdate(cmd.Context(), blockID, body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "block update", start)
	return nil
}

// buildUpdateBody constructs the PATCH body from flags.
func buildUpdateBody(cmd *cobra.Command) (map[string]any, error) {
	body := map[string]any{}

	// --archived flag.
	if cmd.Flags().Changed("archived") {
		archived, _ := cmd.Flags().GetBool("archived")
		body["archived"] = archived
	}

	// --content raw JSON.
	if raw, _ := cmd.Flags().GetString("content"); raw != "" {
		var content map[string]any
		if err := json.Unmarshal([]byte(raw), &content); err != nil {
			return nil, clierrors.InvalidJSON(fmt.Sprintf("--content: %s", err))
		}
		for k, v := range content {
			body[k] = v
		}
		return body, nil
	}

	// Shorthand text flags.
	type shorthand struct {
		flag      string
		blockType string
	}

	shorthands := []shorthand{
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

	color, _ := cmd.Flags().GetString("color")
	if color != "" && !validColors[color] {
		return nil, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidEnum,
			Message: fmt.Sprintf("Invalid color: %s", color),
			Suggestions: []string{
				"Valid colors: default, gray, brown, orange, yellow, green, blue, purple, pink, red",
				"Background colors: gray_background, brown_background, etc.",
			},
		}
	}

	for _, sh := range shorthands {
		if text, _ := cmd.Flags().GetString(sh.flag); text != "" {
			blockContent := map[string]any{
				"rich_text": []map[string]any{
					{"type": "text", "text": map[string]any{"content": text}},
				},
			}
			if color != "" {
				blockContent["color"] = color
			}
			body[sh.blockType] = blockContent
			return body, nil
		}
	}

	if code, _ := cmd.Flags().GetString("code"); code != "" {
		lang, _ := cmd.Flags().GetString("language")
		blockContent := map[string]any{
			"rich_text": []map[string]any{
				{"type": "text", "text": map[string]any{"content": code}},
			},
			"language": lang,
		}
		if color != "" {
			blockContent["color"] = color
		}
		body["code"] = blockContent
		return body, nil
	}

	return body, nil
}

// ---------------------------------------------------------------------------
// block delete
// ---------------------------------------------------------------------------

func newBlockDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <block_id>",
		Aliases: []string{"d", "rm"},
		Short:   "Delete a block",
		Long:    "Delete (archive) a Notion block by ID.",
		Args:    cobra.ExactArgs(1),
		RunE:    runBlockDelete,
	}

	addOutputFlags(cmd)
	return cmd
}

func runBlockDelete(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	blockID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.BlockDelete(cmd.Context(), blockID)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "block delete", start)
	return nil
}
