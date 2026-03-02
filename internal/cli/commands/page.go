package commands

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/internal/errors"
	"github.com/Coastal-Programs/notion-cli/internal/notion"
	"github.com/Coastal-Programs/notion-cli/pkg/output"
	"github.com/spf13/cobra"
)

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
	cmd.Flags().BoolP("simple-properties", "S", false, "Use simple flat properties format (phase 2)")
	cmd.Flags().MarkHidden("simple-properties")
	addOutputFlags(cmd)

	return cmd
}

func runPageCreate(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	client, err := newClient()
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

	cmd.Flags().Bool("map", false, "Output property map (phase 2)")
	cmd.Flags().BoolP("recursive", "R", false, "Recursively retrieve child blocks (phase 2)")
	cmd.Flags().Int("max-depth", 3, "Maximum recursion depth (1-10)")
	cmd.Flags().MarkHidden("map")
	cmd.Flags().MarkHidden("recursive")
	addOutputFlags(cmd)

	return cmd
}

func runPageRetrieve(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
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
	cmd.Flags().BoolP("simple-properties", "S", false, "Use simple flat properties format (phase 2)")
	cmd.Flags().MarkHidden("simple-properties")
	cmd.MarkFlagsMutuallyExclusive("archived", "unarchive")
	addOutputFlags(cmd)

	return cmd
}

func runPageUpdate(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	pageID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	body := map[string]any{}

	// Archive / unarchive.
	if archived, _ := cmd.Flags().GetBool("archived"); archived {
		body["archived"] = true
	}
	if unarchive, _ := cmd.Flags().GetBool("unarchive"); unarchive {
		body["archived"] = false
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

	if len(body) == 0 {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "No updates specified",
			Suggestions: []string{
				"Use --properties to update page properties",
				"Use --archived to archive the page",
				"Use --unarchive to unarchive the page",
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

	client, err := newClient()
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
	defer f.Close()

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
					"rich_text": richText(strings.TrimLeft(line, "# ")),
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
