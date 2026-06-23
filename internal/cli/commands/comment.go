// Package commands implements the Notion CLI subcommands.
//
// This file implements the `comment` command group, which exposes the Notion
// Comments API (Notion-Version 2026-03-11):
//
//	POST   /v1/comments
//	GET    /v1/comments?block_id=...
//	GET    /v1/comments/{id}
//	PATCH  /v1/comments/{id}
//	DELETE /v1/comments/{id}
//
// Per the API contract, exactly one of `rich_text` or `markdown` must be
// provided when creating or updating a comment. The CLI surfaces `--text`
// (sugar for a single-run rich_text array) and `--rich-text <file>` (load
// a rich_text array from a JSON file).
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/internal/notion"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterCommentCommands registers all comment subcommands under root.
func RegisterCommentCommands(root *cobra.Command) {
	commentCmd := &cobra.Command{
		Use:   "comment",
		Short: "Comment operations",
		Long:  "Create, list, retrieve, update, and delete Notion comments.",
	}

	commentCmd.AddCommand(
		newCommentCreateCmd(),
		newCommentListCmd(),
		newCommentRetrieveCmd(),
		newCommentUpdateCmd(),
		newCommentDeleteCmd(),
	)

	root.AddCommand(commentCmd)
}

// loadRichText reads a JSON file containing a rich_text array and returns it
// as []any. The file must contain a JSON array of rich text objects.
func loadRichText(path string) ([]any, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: fmt.Sprintf("Failed to read rich-text file %q: %v", path, err),
			Suggestions: []string{
				"Verify the path exists and is readable",
				"Provide a JSON array of rich text objects, e.g. [{\"type\":\"text\",\"text\":{\"content\":\"hi\"}}]",
			},
		}
	}
	var arr []any
	if err := json.Unmarshal(raw, &arr); err != nil {
		return nil, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidJSON,
			Message: fmt.Sprintf("Failed to parse rich-text file %q as a JSON array: %v", path, err),
			Suggestions: []string{
				"Ensure the file contains a JSON array of rich text objects",
			},
		}
	}
	return arr, nil
}

// ---------------------------------------------------------------------------
// comment create
// ---------------------------------------------------------------------------

func newCommentCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create a comment",
		Long:    "Create a comment on a page, block, or existing discussion thread.",
		RunE:    runCommentCreate,
	}

	cmd.Flags().String("page", "", "Parent page ID (top-level page comment)")
	cmd.Flags().String("block", "", "Parent block ID (top-level block comment)")
	cmd.Flags().String("discussion", "", "Existing discussion ID to reply to")
	cmd.Flags().String("text", "", "Comment body as plain text (becomes a single rich_text run)")
	cmd.Flags().String("rich-text", "", "Path to a JSON file containing a rich_text array")
	cmd.Flags().String("display-name", "", "Custom display name for the comment author (integration-defined)")
	cmd.Flags().StringSlice("attach-file", nil, "File upload ID to attach (repeatable, max 3)")

	cmd.MarkFlagsMutuallyExclusive("page", "block", "discussion")
	cmd.MarkFlagsMutuallyExclusive("text", "rich-text")

	addOutputFlags(cmd)
	return cmd
}

func runCommentCreate(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	parentPage, _ := cmd.Flags().GetString("page")
	parentBlock, _ := cmd.Flags().GetString("block")
	discussion, _ := cmd.Flags().GetString("discussion")
	text, _ := cmd.Flags().GetString("text")
	richTextPath, _ := cmd.Flags().GetString("rich-text")
	displayName, _ := cmd.Flags().GetString("display-name")
	attachFiles, _ := cmd.Flags().GetStringSlice("attach-file")

	if parentPage == "" && parentBlock == "" && discussion == "" {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "Exactly one of --page, --block, or --discussion is required",
			Suggestions: []string{
				"Use --page <page_id> to add a top-level page comment",
				"Use --block <block_id> to add a top-level block comment",
				"Use --discussion <discussion_id> to reply to an existing thread",
			},
		})
	}

	if text == "" && richTextPath == "" {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "Exactly one of --text or --rich-text is required",
			Suggestions: []string{
				"Use --text <plain text> for a single rich_text run",
				"Use --rich-text <file.json> to provide a full rich_text array",
			},
		})
	}

	if len(attachFiles) > 3 {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:        clierrors.CodeInvalidRequest,
			Message:     fmt.Sprintf("Too many --attach-file values: %d (max 3)", len(attachFiles)),
			Suggestions: []string{"Notion comments support up to 3 attachments"},
		})
	}

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	body := map[string]any{}

	switch {
	case discussion != "":
		id, err := resolveID(discussion)
		if err != nil {
			return handleError(cmd, err)
		}
		body["discussion_id"] = id
	case parentPage != "":
		id, err := resolveID(parentPage)
		if err != nil {
			return handleError(cmd, err)
		}
		body["parent"] = map[string]any{"page_id": id}
	case parentBlock != "":
		id, err := resolveID(parentBlock)
		if err != nil {
			return handleError(cmd, err)
		}
		body["parent"] = map[string]any{"block_id": id}
	}

	if richTextPath != "" {
		rt, err := loadRichText(richTextPath)
		if err != nil {
			return handleError(cmd, err)
		}
		body["rich_text"] = rt
	} else {
		body["rich_text"] = []map[string]any{
			{"type": "text", "text": map[string]any{"content": text}},
		}
	}

	if displayName != "" {
		body["display_name"] = map[string]any{
			"type":   "custom",
			"custom": map[string]any{"name": displayName},
		}
	}

	if len(attachFiles) > 0 {
		atts := make([]map[string]any, 0, len(attachFiles))
		for _, id := range attachFiles {
			atts = append(atts, map[string]any{
				"type":           "file_upload",
				"file_upload_id": id,
			})
		}
		body["attachments"] = atts
	}

	result, err := client.CommentCreate(cmd.Context(), body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "comment create", start)
	return nil
}

// ---------------------------------------------------------------------------
// comment list
// ---------------------------------------------------------------------------

func newCommentListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List comments",
		Long:    "List open (unresolved) comments for a block (or page, since pages are blocks).",
		RunE:    runCommentList,
	}

	cmd.Flags().String("block", "", "Block (or page) ID whose comments to list")
	cmd.Flags().Int("page-size", 0, "Number of results per page (max 100)")
	cmd.Flags().String("start-cursor", "", "Pagination cursor")
	cmd.Flags().Bool("all", false, "Paginate until all comments are retrieved")

	_ = cmd.MarkFlagRequired("block")

	addOutputFlags(cmd)
	return cmd
}

func runCommentList(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	block, _ := cmd.Flags().GetString("block")
	pageAll, _ := cmd.Flags().GetBool("all")
	pageSize, _ := cmd.Flags().GetInt("page-size")
	cursor, _ := cmd.Flags().GetString("start-cursor")

	id, err := resolveID(block)
	if err != nil {
		return handleError(cmd, err)
	}

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	qp := notion.QueryParams{BlockID: id}
	if pageSize > 0 {
		qp.PageSize = pageSize
	}
	if cursor != "" {
		qp.StartCursor = cursor
	}

	if !pageAll {
		result, err := client.CommentList(cmd.Context(), qp)
		if err != nil {
			return handleError(cmd, err)
		}
		p := output.NewPrinter(outputFormat(cmd))
		p.PrintSuccess(result, "comment list", start)
		return nil
	}

	var allResults []any
	pageCount := 0
	for {
		pageCount++
		if pageCount > maxPaginationPages {
			fmt.Fprintf(os.Stderr, "Warning: reached maximum pagination limit (%d pages)\n", maxPaginationPages)
			break
		}

		result, err := client.CommentList(cmd.Context(), qp)
		if err != nil {
			return handleError(cmd, err)
		}

		results, _ := result["results"].([]any)
		allResults = append(allResults, results...)

		hasMore, _ := result["has_more"].(bool)
		nextCursor, _ := result["next_cursor"].(string)
		if !hasMore || nextCursor == "" {
			break
		}
		qp.StartCursor = nextCursor
	}

	data := map[string]any{
		"object":       "list",
		"results":      allResults,
		"has_more":     false,
		"next_cursor":  nil,
		"result_count": len(allResults),
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "comment list", start)
	return nil
}

// ---------------------------------------------------------------------------
// comment retrieve
// ---------------------------------------------------------------------------

func newCommentRetrieveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "retrieve <comment_id>",
		Aliases: []string{"r", "get"},
		Short:   "Retrieve a comment",
		Long:    "Retrieve a single Notion comment by ID.",
		Args:    cobra.ExactArgs(1),
		RunE:    runCommentRetrieve,
	}

	addOutputFlags(cmd)
	return cmd
}

func runCommentRetrieve(cmd *cobra.Command, args []string) error {
	start := time.Now()

	id, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.CommentRetrieve(cmd.Context(), id)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "comment retrieve", start)
	return nil
}

// ---------------------------------------------------------------------------
// comment update
// ---------------------------------------------------------------------------

func newCommentUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <comment_id>",
		Aliases: []string{"u"},
		Short:   "Update a comment",
		Long:    "Update an existing Notion comment's body. Exactly one of --text or --rich-text must be provided.",
		Args:    cobra.ExactArgs(1),
		RunE:    runCommentUpdate,
	}

	cmd.Flags().String("text", "", "Replacement body as plain text (single rich_text run)")
	cmd.Flags().String("rich-text", "", "Path to a JSON file containing a rich_text array")
	cmd.MarkFlagsMutuallyExclusive("text", "rich-text")

	addOutputFlags(cmd)
	return cmd
}

func runCommentUpdate(cmd *cobra.Command, args []string) error {
	start := time.Now()

	text, _ := cmd.Flags().GetString("text")
	richTextPath, _ := cmd.Flags().GetString("rich-text")

	if text == "" && richTextPath == "" {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "Exactly one of --text or --rich-text is required",
			Suggestions: []string{
				"Use --text <plain text> for a single rich_text run",
				"Use --rich-text <file.json> to provide a full rich_text array",
			},
		})
	}

	id, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	body := map[string]any{}
	if richTextPath != "" {
		rt, err := loadRichText(richTextPath)
		if err != nil {
			return handleError(cmd, err)
		}
		body["rich_text"] = rt
	} else {
		body["rich_text"] = []map[string]any{
			{"type": "text", "text": map[string]any{"content": text}},
		}
	}

	result, err := client.CommentUpdate(cmd.Context(), id, body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "comment update", start)
	return nil
}

// ---------------------------------------------------------------------------
// comment delete
// ---------------------------------------------------------------------------

func newCommentDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <comment_id>",
		Aliases: []string{"d", "rm"},
		Short:   "Delete a comment",
		Long:    "Delete a Notion comment by ID. A connection can only delete comments it created.",
		Args:    cobra.ExactArgs(1),
		RunE:    runCommentDelete,
	}

	addOutputFlags(cmd)
	return cmd
}

func runCommentDelete(cmd *cobra.Command, args []string) error {
	start := time.Now()

	id, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.CommentDelete(cmd.Context(), id)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "comment delete", start)
	return nil
}
