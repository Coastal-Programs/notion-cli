package commands

import (
	"fmt"
	"io"
	"os"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/internal/notion"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterMarkdownCommands registers the `markdown` subcommand tree under root.
func RegisterMarkdownCommands(root *cobra.Command) {
	mdCmd := &cobra.Command{
		Use:     "markdown",
		Aliases: []string{"md"},
		Short:   "Page content as enhanced markdown",
		Long: "Read and write Notion page content using the enhanced markdown\n" +
			"endpoints (GET /pages/{id}/markdown, PATCH /pages/{id}/markdown).",
	}

	mdCmd.AddCommand(
		newMarkdownGetCmd(),
		newMarkdownSetCmd(),
	)

	root.AddCommand(mdCmd)
}

// --- markdown get ---

func newMarkdownGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <page-id-or-url>",
		Short: "Retrieve a page as enhanced markdown",
		Long: "Retrieve the full content of a Notion page as enhanced markdown.\n\n" +
			"By default, prints the raw markdown to stdout. Use --output json|csv|table\n" +
			"to wrap the response in the standard envelope (markdown is exposed under\n" +
			"the `content` field). Use --output PATH to write the markdown to a file.",
		Args: cobra.ExactArgs(1),
		RunE: runMarkdownGet,
	}

	cmd.Flags().String("file", "", "Write markdown to this file path (instead of stdout)")
	addOutputFlags(cmd)

	return cmd
}

func runMarkdownGet(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	pageID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.PageMarkdownGet(cmd.Context(), pageID, notion.QueryParams{})
	if err != nil {
		return handleError(cmd, err)
	}

	// Decide output mode. If any structured output flag is explicitly set
	// (--output, --json, --csv, etc.), wrap in envelope. Otherwise default
	// to printing raw markdown to stdout (or the --file path).
	if isStructuredOutputRequested(cmd) {
		p := output.NewPrinter(outputFormat(cmd))
		p.Writer = cmd.OutOrStdout()
		envelope := map[string]any{
			"id":                result["id"],
			"object":            result["object"],
			"content":           result["markdown"],
			"truncated":         result["truncated"],
			"unknown_block_ids": result["unknown_block_ids"],
		}
		p.PrintSuccess(envelope, "markdown get", start)
		return nil
	}

	md, _ := result["markdown"].(string)

	filePath, _ := cmd.Flags().GetString("file")
	if filePath != "" {
		if err := os.WriteFile(filePath, []byte(md), 0o644); err != nil {
			return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, fmt.Sprintf("Cannot write file %q", filePath), err))
		}
		return nil
	}

	_, _ = fmt.Fprint(cmd.OutOrStdout(), md)
	if md != "" && md[len(md)-1] != '\n' {
		_, _ = fmt.Fprintln(cmd.OutOrStdout())
	}
	return nil
}

// isStructuredOutputRequested reports whether the caller explicitly asked for
// a structured output format. When false, `markdown get` prints raw markdown.
func isStructuredOutputRequested(cmd *cobra.Command) bool {
	return outputFormatExplicit(cmd)
}

// --- markdown set ---

func newMarkdownSetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <page-id-or-url>",
		Short: "Replace or append a page's markdown content",
		Long: "Replace or append a Notion page's content using enhanced markdown.\n\n" +
			"Provide content via --content STR, --file PATH, or stdin (default).\n" +
			"By default, the page content is replaced; pass --append to append instead.",
		Args: cobra.ExactArgs(1),
		RunE: runMarkdownSet,
	}

	cmd.Flags().String("content", "", "Markdown content as a literal string")
	cmd.Flags().String("file", "", "Path to a markdown file")
	cmd.Flags().Bool("append", false, "Append content instead of replacing")
	cmd.Flags().Bool("allow-deleting-content", false, "Permit removal of child pages or databases")
	cmd.MarkFlagsMutuallyExclusive("content", "file")
	addOutputFlags(cmd)

	return cmd
}

func runMarkdownSet(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	pageID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	content, err := readMarkdownInput(cmd)
	if err != nil {
		return handleError(cmd, err)
	}
	if content == "" {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "No content provided",
			Suggestions: []string{
				"Pass --content \"...\" to supply markdown inline",
				"Pass --file PATH to read from a file",
				"Pipe markdown via stdin",
			},
		})
	}

	appendMode, _ := cmd.Flags().GetBool("append")
	allowDelete, _ := cmd.Flags().GetBool("allow-deleting-content")

	var body map[string]any
	if appendMode {
		inner := map[string]any{"new_str": content}
		if allowDelete {
			inner["allow_deleting_content"] = true
		}
		body = map[string]any{
			"type":           "insert_content",
			"insert_content": inner,
		}
	} else {
		inner := map[string]any{"new_str": content}
		if allowDelete {
			inner["allow_deleting_content"] = true
		}
		body = map[string]any{
			"type":            "replace_content",
			"replace_content": inner,
		}
	}

	result, err := client.PageMarkdownUpdate(cmd.Context(), pageID, body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.Writer = cmd.OutOrStdout()
	p.PrintSuccess(result, "markdown set", start)
	return nil
}

// readMarkdownInput resolves the markdown source, honoring the precedence
// --content > --file > stdin. Mutual exclusion of --content / --file is
// enforced by cobra.
func readMarkdownInput(cmd *cobra.Command) (string, error) {
	if v, _ := cmd.Flags().GetString("content"); v != "" {
		return v, nil
	}
	if path, _ := cmd.Flags().GetString("file"); path != "" {
		b, err := os.ReadFile(path)
		if err != nil {
			return "", &clierrors.NotionCLIError{
				Code:    clierrors.CodeInvalidRequest,
				Message: fmt.Sprintf("Cannot read file %q: %s", path, err),
			}
		}
		return string(b), nil
	}
	// Read from stdin.
	in := cmd.InOrStdin()
	b, err := io.ReadAll(in)
	if err != nil {
		return "", &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: fmt.Sprintf("Cannot read stdin: %s", err),
		}
	}
	return string(b), nil
}
