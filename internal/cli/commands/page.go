package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/internal/notion"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
	"github.com/spf13/cobra"
)

// isTerminal reports whether stdin is an interactive terminal.
// It is a variable so tests can override it.
var isTerminal = func() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// RegisterPageCommands registers all page subcommands under root.
func RegisterPageCommands(root *cobra.Command) {
	pageCmd := &cobra.Command{
		Use:     "page",
		Aliases: []string{"p"},
		Short:   "Page operations",
		Long:    "Create, retrieve, update, and inspect Notion pages.",
	}

	pageCmd.AddCommand(
		newPageCreateCmd(),
		newPageRetrieveCmd(),
		newPageUpdateCmd(),
		newPagePropertyItemCmd(),
		newPageTrashCmd(),
		newPageRestoreCmd(),
		newPageMoveCmd(),
	)

	root.AddCommand(pageCmd)
}

// --- page create ---

func newPageCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create a page",
		Long:    "Create a new Notion page under a parent page or database.",
		RunE:    runPageCreate,
	}

	cmd.Flags().StringP("parent-page-id", "p", "", "Parent page ID")
	cmd.Flags().StringP("parent-data-source-id", "d", "", "Parent database ID")
	cmd.Flags().StringP("file-path", "f", "", "Path to a markdown file for page content")
	cmd.Flags().StringP("title-property", "t", "Name", "Title property name")
	cmd.Flags().String("properties", "", "Properties as JSON string")
	cmd.Flags().String("icon-emoji", "", "Page icon as emoji (e.g. 💰)")
	cmd.Flags().String("icon-url", "", "Page icon as external image URL (https://...)")
	cmd.Flags().String("cover-url", "", "Page cover as external image URL (https://...)")
	cmd.MarkFlagsMutuallyExclusive("icon-emoji", "icon-url")
	addOutputFlags(cmd)

	return cmd
}

func runPageCreate(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	parentPageID, _ := cmd.Flags().GetString("parent-page-id")
	parentDBID, _ := cmd.Flags().GetString("parent-data-source-id")

	if parentPageID == "" && parentDBID == "" {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "Either --parent-page-id (-p) or --parent-data-source-id (-d) is required",
			Suggestions: []string{
				"Use -p <page_id> to create a child page",
				"Use -d <database_id> to create a database row",
			},
		})
	}

	body := map[string]any{}

	// Set parent.
	if parentDBID != "" {
		id, err := resolveID(parentDBID)
		if err != nil {
			return handleError(cmd, err)
		}
		body["parent"] = map[string]any{
			"type":        "database_id",
			"database_id": id,
		}
	} else {
		id, err := resolveID(parentPageID)
		if err != nil {
			return handleError(cmd, err)
		}
		body["parent"] = map[string]any{
			"type":    "page_id",
			"page_id": id,
		}
	}

	// Parse properties.
	propsJSON, _ := cmd.Flags().GetString("properties")
	if propsJSON != "" {
		var props map[string]any
		if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
			return handleError(cmd, clierrors.InvalidJSON(fmt.Sprintf("--properties: %s", err)))
		}
		body["properties"] = props
	}

	// Read markdown file and convert to blocks.
	filePath, _ := cmd.Flags().GetString("file-path")
	if filePath != "" {
		blocks, err := markdownFileToBlocks(filePath)
		if err != nil {
			return handleError(cmd, err)
		}
		body["children"] = blocks

		// If no properties set, create a title from the filename.
		if propsJSON == "" {
			titleProp, _ := cmd.Flags().GetString("title-property")
			body["properties"] = map[string]any{
				titleProp: map[string]any{
					"title": []map[string]any{
						{"text": map[string]any{"content": fileBaseName(filePath)}},
					},
				},
			}
		}
	}

	// Ensure at least empty properties for database pages.
	if _, ok := body["properties"]; !ok && parentDBID != "" {
		body["properties"] = map[string]any{}
	}

	// Icon / cover.
	icon, cover, hasIcon, hasCover, err := buildIconCover(cmd, false)
	if err != nil {
		return handleError(cmd, err)
	}
	if hasIcon {
		body["icon"] = icon
	}
	if hasCover {
		body["cover"] = cover
	}

	result, err := client.PageCreate(cmd.Context(), body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "page create", start)
	return nil
}

// --- page retrieve ---

func newPageRetrieveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "retrieve <page_id>",
		Aliases: []string{"r", "get"},
		Short:   "Retrieve a page",
		Long:    "Retrieve a Notion page by ID.",
		Args:    cobra.ExactArgs(1),
		RunE:    runPageRetrieve,
	}

	addOutputFlags(cmd)

	return cmd
}

func runPageRetrieve(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	pageID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.PageRetrieve(cmd.Context(), pageID)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "page retrieve", start)
	return nil
}

// --- page update ---

func newPageUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <page_id>",
		Aliases: []string{"u"},
		Short:   "Update a page",
		Long:    "Update an existing Notion page's properties or archive state.",
		Args:    cobra.ExactArgs(1),
		RunE:    runPageUpdate,
	}

	cmd.Flags().BoolP("archived", "a", false, "Archive the page")
	cmd.Flags().BoolP("unarchive", "u", false, "Unarchive the page")
	cmd.Flags().String("properties", "", "Properties as JSON string")
	cmd.Flags().String("icon-emoji", "", "Page icon as emoji, or 'none' to clear")
	cmd.Flags().String("icon-url", "", "Page icon as external image URL, or 'none' to clear")
	cmd.Flags().String("cover-url", "", "Page cover as external image URL, or 'none' to clear")
	cmd.MarkFlagsMutuallyExclusive("archived", "unarchive")
	cmd.MarkFlagsMutuallyExclusive("icon-emoji", "icon-url")
	addOutputFlags(cmd)

	return cmd
}

func runPageUpdate(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	pageID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	body := map[string]any{}

	// Archive / unarchive.
	// Notion API 2026-03-11 renamed `archived` -> `in_trash`.
	if archived, _ := cmd.Flags().GetBool("archived"); archived {
		body["in_trash"] = true
	}
	if unarchive, _ := cmd.Flags().GetBool("unarchive"); unarchive {
		body["in_trash"] = false
	}

	// Properties.
	propsJSON, _ := cmd.Flags().GetString("properties")
	if propsJSON != "" {
		var props map[string]any
		if err := json.Unmarshal([]byte(propsJSON), &props); err != nil {
			return handleError(cmd, clierrors.InvalidJSON(fmt.Sprintf("--properties: %s", err)))
		}
		body["properties"] = props
	}

	// Icon / cover (allow 'none' to clear).
	icon, cover, hasIcon, hasCover, err := buildIconCover(cmd, true)
	if err != nil {
		return handleError(cmd, err)
	}
	if hasIcon {
		body["icon"] = icon
	}
	if hasCover {
		body["cover"] = cover
	}

	if len(body) == 0 {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "No updates specified",
			Suggestions: []string{
				"Use --properties to update page properties",
				"Use --archived to archive the page",
				"Use --unarchive to unarchive the page",
				"Use --icon-emoji, --icon-url, or --cover-url to set/clear icon/cover",
			},
		})
	}

	result, err := client.PageUpdate(cmd.Context(), pageID, body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "page update", start)
	return nil
}

// --- page property-item ---

func newPagePropertyItemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "property-item <page_id> <property_id>",
		Aliases: []string{"pi", "prop"},
		Short:   "Retrieve a page property item",
		Long:    "Retrieve the value of a specific property from a page.",
		Args:    cobra.ExactArgs(2),
		RunE:    runPagePropertyItem,
	}

	addOutputFlags(cmd)

	return cmd
}

func runPagePropertyItem(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	pageID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	propID := args[1] // property_id is used as-is (not a UUID)
	if strings.ContainsAny(propID, "/\\") {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: "property ID contains invalid characters",
		})
	}

	result, err := client.PagePropertyRetrieve(cmd.Context(), pageID, propID, notion.QueryParams{})
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "page property-item", start)
	return nil
}

// --- page trash ---

func newPageTrashCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trash <page_id>",
		Short: "Move a page to trash",
		Long:  "Move a Notion page to trash (sets in_trash: true). Requires --yes in non-interactive environments.",
		Args:  cobra.ExactArgs(1),
		RunE:  runPageTrash,
	}

	cmd.Flags().Bool("yes", false, "Skip confirmation prompt")
	addOutputFlags(cmd)

	return cmd
}

func runPageTrash(cmd *cobra.Command, args []string) error {
	start := time.Now()

	yes, _ := cmd.Flags().GetBool("yes")

	if !yes && !isTerminal() {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "--yes flag is required in non-interactive environments",
			Suggestions: []string{
				"Add --yes to confirm the operation: notion-cli page trash <page_id> --yes",
			},
		})
	}

	if !yes && isTerminal() {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Trash page %s? [y/N]: ", args[0])
		scanner := bufio.NewScanner(os.Stdin)
		scanner.Scan()
		response := strings.TrimSpace(strings.ToLower(scanner.Text()))
		if response != "y" && response != "yes" {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
			return nil
		}
	}

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	pageID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.PageTrash(cmd.Context(), pageID)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "page trash", start)
	return nil
}

// --- page restore ---

func newPageRestoreCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore <page_id>",
		Short: "Restore a trashed page",
		Long:  "Restore a trashed Notion page (sets in_trash: false).",
		Args:  cobra.ExactArgs(1),
		RunE:  runPageRestore,
	}

	addOutputFlags(cmd)

	return cmd
}

func runPageRestore(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	pageID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.PageRestore(cmd.Context(), pageID)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "page restore", start)
	return nil
}

// --- page move ---

func newPageMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "move <page_id>",
		Short: "Move a page to a new parent",
		Long:  "Move a Notion page to a new parent via POST /pages/{id}/move.",
		Args:  cobra.ExactArgs(1),
		RunE:  runPageMove,
	}

	cmd.Flags().String("parent", "", "New parent page ID")
	cmd.Flags().String("data-source", "", "New parent data source ID")
	cmd.Flags().Bool("workspace", false, "Move to workspace root")
	cmd.MarkFlagsMutuallyExclusive("parent", "data-source", "workspace")
	addOutputFlags(cmd)

	return cmd
}

func runPageMove(cmd *cobra.Command, args []string) error {
	start := time.Now()

	parentID, _ := cmd.Flags().GetString("parent")
	dataSourceID, _ := cmd.Flags().GetString("data-source")
	workspace, _ := cmd.Flags().GetBool("workspace")

	var parentBody map[string]any
	switch {
	case parentID != "":
		id, err := resolveID(parentID)
		if err != nil {
			return handleError(cmd, err)
		}
		parentBody = map[string]any{
			"parent": map[string]any{
				"type":    "page_id",
				"page_id": id,
			},
		}
	case dataSourceID != "":
		id, err := resolveID(dataSourceID)
		if err != nil {
			return handleError(cmd, err)
		}
		parentBody = map[string]any{
			"parent": map[string]any{
				"type":           "data_source_id",
				"data_source_id": id,
			},
		}
	case workspace:
		parentBody = map[string]any{
			"parent": map[string]any{
				"type":      "workspace",
				"workspace": true,
			},
		}
	default:
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "one of --parent, --data-source, or --workspace is required",
			Suggestions: []string{
				"Use --parent <page_id> to move under a page",
				"Use --data-source <data_source_id> to move under a data source",
				"Use --workspace to move to the workspace root",
			},
		})
	}

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	pageID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.PageMove(cmd.Context(), pageID, parentBody)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "page move", start)
	return nil
}

// buildIconCover reads the --icon-emoji, --icon-url, and --cover-url flags and
// returns the corresponding Notion API values. When allowClear is true the
// literal value "none" produces a JSON null payload (used by `page update` to
// clear a previously set icon or cover). The hasIcon/hasCover return values
// indicate whether the caller should set the field on the request body at all.
func buildIconCover(cmd *cobra.Command, allowClear bool) (icon, cover any, hasIcon, hasCover bool, err error) {
	emoji, _ := cmd.Flags().GetString("icon-emoji")
	iconURL, _ := cmd.Flags().GetString("icon-url")
	coverURL, _ := cmd.Flags().GetString("cover-url")

	if emoji != "" {
		hasIcon = true
		if allowClear && emoji == "none" {
			icon = nil
		} else {
			icon = map[string]any{"type": "emoji", "emoji": emoji}
		}
	} else if iconURL != "" {
		hasIcon = true
		if allowClear && iconURL == "none" {
			icon = nil
		} else {
			if vErr := validateExternalURL(iconURL, "--icon-url"); vErr != nil {
				return nil, nil, false, false, vErr
			}
			icon = map[string]any{
				"type":     "external",
				"external": map[string]any{"url": iconURL},
			}
		}
	}

	if coverURL != "" {
		hasCover = true
		if allowClear && coverURL == "none" {
			cover = nil
		} else {
			if vErr := validateExternalURL(coverURL, "--cover-url"); vErr != nil {
				return nil, nil, false, false, vErr
			}
			cover = map[string]any{
				"type":     "external",
				"external": map[string]any{"url": coverURL},
			}
		}
	}

	return icon, cover, hasIcon, hasCover, nil
}

// validateExternalURL ensures the value is a parseable http(s) URL.
func validateExternalURL(raw, flag string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: fmt.Sprintf("%s must be a valid http(s) URL", flag),
		}
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: fmt.Sprintf("%s scheme must be http or https (got %q)", flag, u.Scheme),
		}
	}
	return nil
}

// --- Markdown to blocks converter ---

// markdownFileToBlocks reads a markdown file and converts it to Notion blocks.
func markdownFileToBlocks(path string) ([]map[string]any, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: fmt.Sprintf("Cannot read file: %s", err),
		}
	}
	defer f.Close() //nolint:errcheck

	var blocks []map[string]any
	scanner := bufio.NewScanner(f)
	var inCodeBlock bool
	var codeLines []string
	var codeLang string

	for scanner.Scan() {
		line := scanner.Text()

		// Code block fences.
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				// End code block.
				blocks = append(blocks, map[string]any{
					"type": "code",
					"code": map[string]any{
						"rich_text": []map[string]any{
							{"type": "text", "text": map[string]any{"content": strings.Join(codeLines, "\n")}},
						},
						"language": codeLang,
					},
				})
				inCodeBlock = false
				codeLines = nil
				codeLang = ""
			} else {
				inCodeBlock = true
				codeLang = strings.TrimPrefix(line, "```")
				codeLang = strings.TrimSpace(codeLang)
				if codeLang == "" {
					codeLang = "plain text"
				}
			}
			continue
		}

		if inCodeBlock {
			codeLines = append(codeLines, line)
			continue
		}

		// Horizontal rules.
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine == "---" || trimmedLine == "***" || trimmedLine == "___" {
			blocks = append(blocks, map[string]any{
				"type":    "divider",
				"divider": map[string]any{},
			})
			continue
		}

		// Headings (Notion only supports h1-h3; deeper headings become paragraphs).
		if strings.HasPrefix(line, "#### ") {
			// Notion doesn't support h4+, render as bold paragraph.
			blocks = append(blocks, map[string]any{
				"type": "paragraph",
				"paragraph": map[string]any{
					"rich_text": richText(strings.TrimPrefix(line, "#### ")),
				},
			})
			continue
		}
		if strings.HasPrefix(line, "### ") {
			blocks = append(blocks, headingBlock(3, strings.TrimPrefix(line, "### ")))
			continue
		}
		if strings.HasPrefix(line, "## ") {
			blocks = append(blocks, headingBlock(2, strings.TrimPrefix(line, "## ")))
			continue
		}
		if strings.HasPrefix(line, "# ") {
			blocks = append(blocks, headingBlock(1, strings.TrimPrefix(line, "# ")))
			continue
		}

		// Blockquotes.
		if strings.HasPrefix(line, "> ") {
			blocks = append(blocks, map[string]any{
				"type": "quote",
				"quote": map[string]any{
					"rich_text": richText(strings.TrimPrefix(line, "> ")),
				},
			})
			continue
		}

		// Unordered list items.
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			blocks = append(blocks, map[string]any{
				"type": "bulleted_list_item",
				"bulleted_list_item": map[string]any{
					"rich_text": richText(line[2:]),
				},
			})
			continue
		}

		// Numbered list items.
		if len(line) > 2 && line[0] >= '0' && line[0] <= '9' {
			if idx := strings.Index(line, ". "); idx > 0 && idx < 4 {
				blocks = append(blocks, map[string]any{
					"type": "numbered_list_item",
					"numbered_list_item": map[string]any{
						"rich_text": richText(line[idx+2:]),
					},
				})
				continue
			}
		}

		// Empty lines and paragraphs.
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		blocks = append(blocks, map[string]any{
			"type": "paragraph",
			"paragraph": map[string]any{
				"rich_text": richText(trimmed),
			},
		})
	}

	// If file ended mid code block, flush remaining code.
	if inCodeBlock && len(codeLines) > 0 {
		blocks = append(blocks, map[string]any{
			"type": "code",
			"code": map[string]any{
				"rich_text": []map[string]any{
					{"type": "text", "text": map[string]any{"content": strings.Join(codeLines, "\n")}},
				},
				"language": codeLang,
			},
		})
	}

	return blocks, scanner.Err()
}

// headingBlock creates a Notion heading block (1, 2, or 3).
func headingBlock(level int, text string) map[string]any {
	key := fmt.Sprintf("heading_%d", level)
	return map[string]any{
		"type": key,
		key: map[string]any{
			"rich_text": richText(text),
		},
	}
}

// splitRichText splits text into chunks of maxLen for Notion's 2000-char limit.
func splitRichText(s string, maxLen int) []map[string]any {
	if len(s) <= maxLen {
		return []map[string]any{
			{"type": "text", "text": map[string]any{"content": s}},
		}
	}
	var segments []map[string]any
	for len(s) > 0 {
		end := maxLen
		if end > len(s) {
			end = len(s)
		}
		segments = append(segments, map[string]any{
			"type": "text",
			"text": map[string]any{"content": s[:end]},
		})
		s = s[end:]
	}
	return segments
}

// richText creates a simple rich text array from a plain string.
func richText(s string) []map[string]any {
	return splitRichText(s, 2000)
}

// fileBaseName extracts a human-readable name from a file path.
func fileBaseName(path string) string {
	name := filepath.Base(path)
	// Strip extension.
	if dot := strings.LastIndex(name, "."); dot > 0 {
		name = name[:dot]
	}
	// Replace dashes/underscores with spaces.
	name = strings.ReplaceAll(name, "-", " ")
	name = strings.ReplaceAll(name, "_", " ")
	return name
}
