package commands

import (
	"fmt"
	"io"
	"strings"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterPagesCommands adds official ntn-style page aliases without changing
// the existing page command group.
func RegisterPagesCommands(root *cobra.Command) {
	pagesCmd := &cobra.Command{
		Use:   "pages",
		Short: "Official-style page operations",
		Long:  "Official ntn-compatible page helpers for markdown get/create/update/trash workflows.",
	}
	pagesCmd.AddCommand(
		newPagesGetCmd(),
		newPagesCreateCmd(),
		newPagesUpdateCmd(),
		newPagesTrashCmd(),
	)
	root.AddCommand(pagesCmd)
}

func newPagesGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <page-id-or-url>",
		Short: "Retrieve a page as Markdown",
		Args:  cobra.ExactArgs(1),
		RunE:  runMarkdownGet,
	}
	addOutputFlags(cmd)
	return cmd
}

func newPagesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <page-id-or-url>",
		Short: "Update a page from Markdown",
		Args:  cobra.ExactArgs(1),
		RunE:  runMarkdownSet,
	}
	cmd.Flags().String("content", "", "Markdown content; reads stdin when omitted")
	cmd.Flags().Bool("allow-deleting-content", false, "Allow deletion of child pages and databases")
	addOutputFlags(cmd)
	return cmd
}

func newPagesTrashCmd() *cobra.Command {
	cmd := newPageTrashCmd()
	cmd.Use = "trash <page-id-or-url>"
	cmd.Short = "Trash a page"
	return cmd
}

func newPagesCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a page from Markdown",
		Args:  cobra.NoArgs,
		RunE:  runPagesCreate,
	}
	cmd.Flags().String("parent", "", "Parent target: page:<id>, database:<id>, or data-source:<id>")
	cmd.Flags().String("content", "", "Markdown content; reads stdin when omitted")
	_ = cmd.MarkFlagRequired("parent")
	addOutputFlags(cmd)
	return cmd
}

func runPagesCreate(cmd *cobra.Command, _ []string) error {
	start := time.Now()
	parent, _ := cmd.Flags().GetString("parent")
	content, _ := cmd.Flags().GetString("content")
	if content == "" {
		b, err := io.ReadAll(cmd.InOrStdin())
		if err != nil {
			return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Cannot read stdin", err))
		}
		content = string(b)
	}
	if strings.TrimSpace(content) == "" {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "No markdown content provided",
		})
	}

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}
	parentBody, err := pagesParentBody(parent, content)
	if err != nil {
		return handleError(cmd, err)
	}

	created, err := client.PageCreate(cmd.Context(), parentBody)
	if err != nil {
		return handleError(cmd, err)
	}
	pageID, _ := created["id"].(string)
	if pageID == "" {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: "Page create response missing id",
		})
	}
	updateBody := map[string]any{
		"type":            "replace_content",
		"replace_content": map[string]any{"new_str": content},
	}
	updated, err := client.PageMarkdownUpdate(cmd.Context(), pageID, updateBody)
	if err != nil {
		return handleError(cmd, err)
	}
	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(map[string]any{
		"page":     created,
		"markdown": updated,
	}, "pages create", start)
	return nil
}

func pagesParentBody(parent, content string) (map[string]any, error) {
	kind, rawID, ok := strings.Cut(parent, ":")
	if !ok || rawID == "" {
		return nil, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: "--parent must be page:<id>, database:<id>, or data-source:<id>",
		}
	}
	id, err := resolveID(rawID)
	if err != nil {
		return nil, err
	}
	switch kind {
	case "page":
		return map[string]any{
			"parent": map[string]any{
				"type":    "page_id",
				"page_id": id,
			},
		}, nil
	case "database":
		return map[string]any{
			"parent": map[string]any{
				"type":        "database_id",
				"database_id": id,
			},
			"properties": defaultPageProperties(content),
		}, nil
	case "data-source", "data_source", "datasource":
		return map[string]any{
			"parent": map[string]any{
				"type":           "data_source_id",
				"data_source_id": id,
			},
			"properties": defaultPageProperties(content),
		}, nil
	default:
		return nil, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: fmt.Sprintf("Unsupported parent type %q", kind),
		}
	}
}

func defaultPageProperties(content string) map[string]any {
	title := firstMarkdownHeading(content)
	if title == "" {
		title = "Untitled"
	}
	return map[string]any{
		"Name": map[string]any{
			"title": []map[string]any{
				{"text": map[string]any{"content": title}},
			},
		},
	}
}

func firstMarkdownHeading(content string) string {
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			return strings.TrimSpace(strings.TrimLeft(line, "#"))
		}
	}
	return ""
}
