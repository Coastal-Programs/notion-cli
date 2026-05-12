// Package commands implements the Notion CLI subcommands.
//
// This file implements the `custom-emoji` command group, which exposes the
// Notion Custom Emojis API (Notion-Version 2026-03-11):
//
//	GET /v1/custom_emojis
//	GET /v1/custom_emojis/{id}
//
// See https://developers.notion.com/reference/list-custom-emojis.
package commands

import (
	"fmt"
	"os"
	"time"

	"github.com/Coastal-Programs/notion-cli/v6/internal/notion"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterCustomEmojiCommands registers all custom-emoji subcommands under root.
func RegisterCustomEmojiCommands(root *cobra.Command) {
	emojiCmd := &cobra.Command{
		Use:     "custom-emoji",
		Aliases: []string{"custom-emojis", "emoji"},
		Short:   "Custom emoji operations",
		Long:    "List and retrieve workspace custom emojis (Notion API 2026-03-11).",
	}

	emojiCmd.AddCommand(
		newCustomEmojiListCmd(),
		newCustomEmojiRetrieveCmd(),
	)

	root.AddCommand(emojiCmd)
}

// ---------------------------------------------------------------------------
// custom-emoji list
// ---------------------------------------------------------------------------

func newCustomEmojiListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List workspace custom emojis",
		Long:    "List custom emojis in the workspace.",
		Args:    cobra.NoArgs,
		RunE:    runCustomEmojiList,
	}

	cmd.Flags().Int("page-size", 0, "Number of results per page (max 100)")
	cmd.Flags().String("start-cursor", "", "Pagination cursor")
	cmd.Flags().Bool("all", false, "Paginate until all custom emojis are retrieved")

	addOutputFlags(cmd)
	return cmd
}

func runCustomEmojiList(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	qp := notion.QueryParams{}
	if pageSize, _ := cmd.Flags().GetInt("page-size"); pageSize > 0 {
		qp.PageSize = pageSize
	}
	if cursor, _ := cmd.Flags().GetString("start-cursor"); cursor != "" {
		qp.StartCursor = cursor
	}

	pageAll, _ := cmd.Flags().GetBool("all")
	ctx := cmd.Context()

	var rawResults []any

	if !pageAll {
		result, err := client.CustomEmojiList(ctx, qp)
		if err != nil {
			return handleError(cmd, err)
		}
		if rs, ok := result["results"].([]any); ok {
			rawResults = rs
		}
	} else {
		pageCount := 0
		for {
			pageCount++
			if pageCount > maxPaginationPages {
				fmt.Fprintf(os.Stderr, "Warning: reached maximum pagination limit (%d pages)\n", maxPaginationPages)
				break
			}

			result, err := client.CustomEmojiList(ctx, qp)
			if err != nil {
				return handleError(cmd, err)
			}

			if rs, ok := result["results"].([]any); ok {
				rawResults = append(rawResults, rs...)
			}

			hasMore, _ := result["has_more"].(bool)
			nextCursor, _ := result["next_cursor"].(string)
			if !hasMore || nextCursor == "" {
				break
			}
			qp.StartCursor = nextCursor
		}
	}

	// Shape results into name/id/url table rows.
	tableData := make([]map[string]any, 0, len(rawResults))
	for _, item := range rawResults {
		emoji, ok := item.(map[string]any)
		if !ok {
			continue
		}
		tableData = append(tableData, map[string]any{
			"id":   emoji["id"],
			"name": emoji["name"],
			"url":  emoji["url"],
		})
	}

	data := map[string]any{
		"results":      tableData,
		"result_count": len(tableData),
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "custom-emoji list", start)
	return nil
}

// ---------------------------------------------------------------------------
// custom-emoji retrieve
// ---------------------------------------------------------------------------

func newCustomEmojiRetrieveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "retrieve <custom_emoji_id>",
		Aliases: []string{"r", "get"},
		Short:   "Retrieve a custom emoji",
		Long:    "Retrieve a single workspace custom emoji by ID.",
		Args:    cobra.ExactArgs(1),
		RunE:    runCustomEmojiRetrieve,
	}

	addOutputFlags(cmd)
	return cmd
}

func runCustomEmojiRetrieve(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	id, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.CustomEmojiRetrieve(cmd.Context(), id)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "custom-emoji retrieve", start)
	return nil
}
