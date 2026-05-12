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

// RegisterDataSourceCommands registers the `data-source` command group under root.
//
// This is the explicit, low-level command group for the 2025-09-03 Notion API
// data source endpoints. `db query --data-source` remains the convenience path
// for querying through a database alias.
func RegisterDataSourceCommands(root *cobra.Command) {
	dsCmd := &cobra.Command{
		Use:     "data-source",
		Aliases: []string{"ds", "data_source", "datasource"},
		Short:   "Data source operations (Notion API 2025-09-03)",
		Long: "Retrieve, create, update, and query Notion data sources. " +
			"Each Notion database has one or more data sources; these commands address them directly.",
	}

	dsCmd.AddCommand(
		newDataSourceRetrieveCmd(),
		newDataSourceCreateCmd(),
		newDataSourceUpdateCmd(),
		newDataSourceQueryCmd(),
		newDataSourceTemplatesCmd(),
		newDataSourcePropertiesCmd(),
	)

	root.AddCommand(dsCmd)
}

// --- data-source retrieve ---

func newDataSourceRetrieveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "retrieve <data_source_id>",
		Aliases: []string{"r", "get"},
		Short:   "Retrieve a data source",
		Long:    "Retrieve a Notion data source by ID, showing its schema and configuration.",
		Args:    cobra.ExactArgs(1),
		RunE:    runDataSourceRetrieve,
	}
	addOutputFlags(cmd)
	return cmd
}

func runDataSourceRetrieve(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	dsID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.DataSourceRetrieve(cmd.Context(), dsID)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "data-source retrieve", start)
	return nil
}

// --- data-source create ---

func newDataSourceCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create a data source",
		Long:    "Create a new data source attached to an existing database parent.",
		Args:    cobra.NoArgs,
		RunE:    runDataSourceCreate,
	}

	cmd.Flags().String("parent-database", "", "Parent database ID (required)")
	cmd.Flags().String("properties", "", "Properties schema as JSON (required)")
	cmd.Flags().String("title", "", "Data source title")
	_ = cmd.MarkFlagRequired("parent-database")
	_ = cmd.MarkFlagRequired("properties")
	addOutputFlags(cmd)

	return cmd
}

func runDataSourceCreate(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	parentRaw, _ := cmd.Flags().GetString("parent-database")
	parentID, err := resolveID(parentRaw)
	if err != nil {
		return handleError(cmd, err)
	}

	propsJSON, _ := cmd.Flags().GetString("properties")
	var properties map[string]any
	if err := json.Unmarshal([]byte(propsJSON), &properties); err != nil {
		return handleError(cmd, clierrors.InvalidJSON(fmt.Sprintf("--properties: %s", err)))
	}

	body := map[string]any{
		"parent": map[string]any{
			"type":        "database_id",
			"database_id": parentID,
		},
		"properties": properties,
	}

	if title, _ := cmd.Flags().GetString("title"); title != "" {
		body["title"] = []map[string]any{
			{"type": "text", "text": map[string]any{"content": title}},
		}
	}

	result, err := client.DataSourceCreate(cmd.Context(), body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "data-source create", start)
	return nil
}

// --- data-source update ---

func newDataSourceUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <data_source_id>",
		Aliases: []string{"u"},
		Short:   "Update a data source",
		Long:    "Update a Notion data source's title, properties schema, or trash state.",
		Args:    cobra.ExactArgs(1),
		RunE:    runDataSourceUpdate,
	}

	cmd.Flags().String("title", "", "New data source title")
	cmd.Flags().String("properties", "", "Updated properties schema as JSON")
	cmd.Flags().String("in-trash", "", "Move to or restore from trash (true/false)")
	addOutputFlags(cmd)

	return cmd
}

func runDataSourceUpdate(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	dsID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	body := map[string]any{}

	if title, _ := cmd.Flags().GetString("title"); title != "" {
		body["title"] = []map[string]any{
			{"type": "text", "text": map[string]any{"content": title}},
		}
	}

	if propsJSON, _ := cmd.Flags().GetString("properties"); propsJSON != "" {
		var properties map[string]any
		if err := json.Unmarshal([]byte(propsJSON), &properties); err != nil {
			return handleError(cmd, clierrors.InvalidJSON(fmt.Sprintf("--properties: %s", err)))
		}
		body["properties"] = properties
	}

	if trashRaw, _ := cmd.Flags().GetString("in-trash"); trashRaw != "" {
		switch trashRaw {
		case "true":
			body["in_trash"] = true
		case "false":
			body["in_trash"] = false
		default:
			return handleError(cmd, &clierrors.NotionCLIError{
				Code:    clierrors.CodeInvalidRequest,
				Message: fmt.Sprintf("Invalid --in-trash %q", trashRaw),
				Suggestions: []string{
					"Valid values: true, false",
				},
			})
		}
	}

	if len(body) == 0 {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: "data-source update requires at least one of --title, --properties, or --in-trash",
		})
	}

	result, err := client.DataSourceUpdate(cmd.Context(), dsID, body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "data-source update", start)
	return nil
}

// --- data-source query ---

func newDataSourceQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query <data_source_id>",
		Aliases: []string{"q"},
		Short:   "Query a data source",
		Long:    "Query a Notion data source directly by ID with optional filters, sorts, and pagination.",
		Args:    cobra.ExactArgs(1),
		RunE:    runDataSourceQuery,
	}

	cmd.Flags().String("filter", "", "Filter as JSON string")
	cmd.Flags().String("sorts", "", "Sorts as JSON array")
	cmd.Flags().Int("page-size", 0, "Number of results per page (1-100)")
	cmd.Flags().String("start-cursor", "", "Pagination cursor")
	addOutputFlags(cmd)

	return cmd
}

func runDataSourceQuery(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	dsID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	body := map[string]any{}

	if f, _ := cmd.Flags().GetString("filter"); f != "" {
		var filter map[string]any
		if err := json.Unmarshal([]byte(f), &filter); err != nil {
			return handleError(cmd, clierrors.InvalidJSON(fmt.Sprintf("--filter: %s", err)))
		}
		body["filter"] = filter
	}

	if s, _ := cmd.Flags().GetString("sorts"); s != "" {
		var sorts []any
		if err := json.Unmarshal([]byte(s), &sorts); err != nil {
			return handleError(cmd, clierrors.InvalidJSON(fmt.Sprintf("--sorts: %s", err)))
		}
		body["sorts"] = sorts
	}

	if ps, _ := cmd.Flags().GetInt("page-size"); ps > 0 {
		if ps > 100 {
			return handleError(cmd, &clierrors.NotionCLIError{
				Code:    clierrors.CodeInvalidRequest,
				Message: fmt.Sprintf("--page-size must be between 1 and 100, got %d", ps),
			})
		}
		body["page_size"] = ps
	}

	if c, _ := cmd.Flags().GetString("start-cursor"); c != "" {
		body["start_cursor"] = c
	}

	result, err := client.DataSourceQuery(cmd.Context(), dsID, body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "data-source query", start)
	return nil
}

// --- data-source templates ---

func newDataSourceTemplatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates <data_source_id>",
		Short: "List templates for a data source",
		Long:  "List the page templates associated with a Notion data source.",
		Args:  cobra.ExactArgs(1),
		RunE:  runDataSourceTemplates,
	}

	cmd.Flags().Int("page-size", 0, "Number of results per page (1-100)")
	cmd.Flags().Bool("all", false, "Fetch all pages (auto-paginate)")
	cmd.Flags().String("start-cursor", "", "Pagination cursor")
	addOutputFlags(cmd)

	return cmd
}

func runDataSourceTemplates(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	dsID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	pageAll, _ := cmd.Flags().GetBool("all")
	ps, _ := cmd.Flags().GetInt("page-size")
	cursor, _ := cmd.Flags().GetString("start-cursor")

	query := notion.QueryParams{
		PageSize:    ps,
		StartCursor: cursor,
	}

	var allResults []any
	pageCount := 0

	for {
		pageCount++
		if pageCount > maxPaginationPages {
			fmt.Fprintf(os.Stderr, "Warning: reached maximum pagination limit (%d pages)\n", maxPaginationPages)
			break
		}

		result, err := client.DataSourceTemplatesList(cmd.Context(), dsID, query)
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
		query.StartCursor = nextCursor
	}

	data := map[string]any{
		"results":      allResults,
		"result_count": len(allResults),
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "data-source templates", start)
	return nil
}

// --- data-source properties ---

func newDataSourcePropertiesCmd() *cobra.Command {
	propertiesCmd := &cobra.Command{
		Use:   "properties",
		Short: "Data source property operations",
		Long:  "Commands for managing data source properties.",
	}
	propertiesCmd.AddCommand(newDataSourcePropertiesUpdateCmd())
	return propertiesCmd
}

func newDataSourcePropertiesUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <data_source_id>",
		Aliases: []string{"u"},
		Short:   "Update data source properties schema",
		Long:    "Update the properties schema of a Notion data source.",
		Args:    cobra.ExactArgs(1),
		RunE:    runDataSourcePropertiesUpdate,
	}

	cmd.Flags().String("schema", "", "Properties schema as JSON string or @file path (required)")
	_ = cmd.MarkFlagRequired("schema")
	addOutputFlags(cmd)

	return cmd
}

func runDataSourcePropertiesUpdate(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	dsID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	schemaRaw, _ := cmd.Flags().GetString("schema")

	var schemaJSON []byte
	if len(schemaRaw) > 0 && schemaRaw[0] == '@' {
		schemaJSON, err = os.ReadFile(schemaRaw[1:])
		if err != nil {
			return handleError(cmd, &clierrors.NotionCLIError{
				Code:    clierrors.CodeInvalidRequest,
				Message: fmt.Sprintf("Cannot read schema file: %s", err),
			})
		}
	} else {
		schemaJSON = []byte(schemaRaw)
	}

	var properties map[string]any
	if err := json.Unmarshal(schemaJSON, &properties); err != nil {
		return handleError(cmd, clierrors.InvalidJSON(fmt.Sprintf("--schema: %s", err)))
	}

	body := map[string]any{
		"properties": properties,
	}

	result, err := client.DataSourcePropertiesUpdate(cmd.Context(), dsID, body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "data-source properties update", start)
	return nil
}
