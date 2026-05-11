// Package commands implements the Notion CLI subcommands.
//
// This file implements the `view` command group, which exposes the Notion
// Views API (Notion-Version 2026-03-11):
//
//	POST   /v1/views
//	GET    /v1/views?database_id=...|data_source_id=...
//	GET    /v1/views/{view_id}
//	PATCH  /v1/views/{view_id}
//	DELETE /v1/views/{view_id}
//	POST   /v1/views/{view_id}/queries
//	GET    /v1/views/{view_id}/queries/{query_id}
//	DELETE /v1/views/{view_id}/queries/{query_id}
//
// Cached query results expire 15 minutes after creation; clients should
// re-create the query if `expires_at` has passed before paginating.
package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/internal/notion"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
	"github.com/spf13/cobra"
)

// validViewTypes is the set of view type values accepted by the Notion API.
var validViewTypes = map[string]bool{
	"table":     true,
	"board":     true,
	"list":      true,
	"calendar":  true,
	"timeline":  true,
	"gallery":   true,
	"form":      true,
	"chart":     true,
	"map":       true,
	"dashboard": true,
}

// loadJSONFlag reads a JSON value from either a raw JSON string or a file path
// prefixed with "@". dest must be a pointer (e.g. *map[string]any, *[]any).
func loadJSONFlag(flagValue string, dest any) error {
	var raw []byte
	if strings.HasPrefix(flagValue, "@") {
		path := flagValue[1:]
		data, err := os.ReadFile(path)
		if err != nil {
			return &clierrors.NotionCLIError{
				Code:    clierrors.CodeInvalidRequest,
				Message: fmt.Sprintf("Cannot read file %q: %v", path, err),
				Suggestions: []string{
					"Verify the path exists and is readable",
				},
			}
		}
		raw = data
	} else {
		raw = []byte(flagValue)
	}
	if err := json.Unmarshal(raw, dest); err != nil {
		return &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidJSON,
			Message: fmt.Sprintf("Invalid JSON: %v", err),
			Suggestions: []string{
				"Provide valid JSON or use @<file> to load from a file",
			},
		}
	}
	return nil
}

// RegisterViewCommands registers all view subcommands under root.
func RegisterViewCommands(root *cobra.Command) {
	viewCmd := &cobra.Command{
		Use:   "view",
		Short: "View operations",
		Long:  "Create, list, retrieve, update, delete, and query Notion database views (Notion-Version 2026-03-11).",
	}

	viewCmd.AddCommand(
		newViewCreateCmd(),
		newViewListCmd(),
		newViewRetrieveCmd(),
		newViewUpdateCmd(),
		newViewDeleteCmd(),
		newViewQueryCmd(),
		newViewResultsCmd(),
		newViewDeleteQueryCmd(),
	)

	root.AddCommand(viewCmd)
}

// ---------------------------------------------------------------------------
// view create
// ---------------------------------------------------------------------------

func newViewCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Aliases: []string{"c"},
		Short:   "Create a view",
		Long:    "Create a new Notion database view. Requires --data-source, --name, and --type.",
		RunE:    runViewCreate,
	}

	cmd.Flags().String("data-source", "", "Data source ID the view belongs to (required)")
	cmd.Flags().String("name", "", "Display name for the view (required)")
	cmd.Flags().String("type", "", "View type: table|board|list|calendar|timeline|gallery|form|chart|map|dashboard (required)")
	cmd.Flags().String("database", "", "Database ID to associate with the view")
	cmd.Flags().String("filter", "", "Filter as a JSON object or @<file> path")
	cmd.Flags().String("sorts", "", "Sorts as a JSON array or @<file> path")

	_ = cmd.MarkFlagRequired("data-source")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("type")

	addOutputFlags(cmd)
	return cmd
}

func runViewCreate(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	dataSource, _ := cmd.Flags().GetString("data-source")
	name, _ := cmd.Flags().GetString("name")
	viewType, _ := cmd.Flags().GetString("type")
	database, _ := cmd.Flags().GetString("database")
	filterStr, _ := cmd.Flags().GetString("filter")
	sortsStr, _ := cmd.Flags().GetString("sorts")

	if !validViewTypes[viewType] {
		types := "table, board, list, calendar, timeline, gallery, form, chart, map, dashboard"
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:        clierrors.CodeInvalidRequest,
			Message:     fmt.Sprintf("Invalid view type %q. Must be one of: %s", viewType, types),
			Suggestions: []string{"Use one of: " + types},
		})
	}

	dsID, err := resolveID(dataSource)
	if err != nil {
		return handleError(cmd, err)
	}

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	body := map[string]any{
		"data_source_id": dsID,
		"name":           name,
		"type":           viewType,
	}

	if database != "" {
		dbID, err := resolveID(database)
		if err != nil {
			return handleError(cmd, err)
		}
		body["database_id"] = dbID
	}

	if filterStr != "" {
		var filter map[string]any
		if err := loadJSONFlag(filterStr, &filter); err != nil {
			return handleError(cmd, err)
		}
		body["filter"] = filter
	}

	if sortsStr != "" {
		var sorts []any
		if err := loadJSONFlag(sortsStr, &sorts); err != nil {
			return handleError(cmd, err)
		}
		body["sorts"] = sorts
	}

	result, err := client.ViewCreate(cmd.Context(), body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "view create", start)
	return nil
}

// ---------------------------------------------------------------------------
// view list
// ---------------------------------------------------------------------------

func newViewListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List views in a database or data source",
		Long:    "List all views in a database (or every linked view across the workspace for a data source). Exactly one of --data-source or --database is required.",
		RunE:    runViewList,
	}

	cmd.Flags().String("data-source", "", "Data source ID (lists all linked views across the workspace)")
	cmd.Flags().String("database", "", "Database ID (lists views defined on that database)")
	cmd.Flags().Int("page-size", 0, "Number of results per page (max 100)")
	cmd.Flags().String("start-cursor", "", "Pagination cursor")

	cmd.MarkFlagsMutuallyExclusive("data-source", "database")

	addOutputFlags(cmd)
	return cmd
}

func runViewList(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	dataSource, _ := cmd.Flags().GetString("data-source")
	database, _ := cmd.Flags().GetString("database")

	if dataSource == "" && database == "" {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "Either --data-source or --database is required",
			Suggestions: []string{
				"Use --data-source <data_source_id> to list every linked view across the workspace",
				"Use --database <database_id> to list views defined on that database",
			},
		})
	}

	var dsID, dbID string
	if dataSource != "" {
		id, err := resolveID(dataSource)
		if err != nil {
			return handleError(cmd, err)
		}
		dsID = id
	}
	if database != "" {
		id, err := resolveID(database)
		if err != nil {
			return handleError(cmd, err)
		}
		dbID = id
	}

	qp := notion.QueryParams{}
	if pageSize, _ := cmd.Flags().GetInt("page-size"); pageSize > 0 {
		qp.PageSize = pageSize
	}
	if cursor, _ := cmd.Flags().GetString("start-cursor"); cursor != "" {
		qp.StartCursor = cursor
	}

	result, err := client.ViewsList(cmd.Context(), dbID, dsID, qp)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "view list", start)
	return nil
}

// ---------------------------------------------------------------------------
// view retrieve
// ---------------------------------------------------------------------------

func newViewRetrieveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "retrieve <view_id>",
		Aliases: []string{"r", "get"},
		Short:   "Retrieve a view",
		Long:    "Retrieve a single Notion view by ID.",
		Args:    cobra.ExactArgs(1),
		RunE:    runViewRetrieve,
	}

	addOutputFlags(cmd)
	return cmd
}

func runViewRetrieve(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	id, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.ViewRetrieve(cmd.Context(), id)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "view retrieve", start)
	return nil
}

// ---------------------------------------------------------------------------
// view update
// ---------------------------------------------------------------------------

func newViewUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <view_id>",
		Aliases: []string{"u"},
		Short:   "Update a view",
		Long:    "Update an existing Notion view. At least one of --name, --filter, or --sorts is required.",
		Args:    cobra.ExactArgs(1),
		RunE:    runViewUpdate,
	}

	cmd.Flags().String("name", "", "New display name for the view")
	cmd.Flags().String("filter", "", "Filter as a JSON object or @<file> path")
	cmd.Flags().String("sorts", "", "Sorts as a JSON array or @<file> path")

	addOutputFlags(cmd)
	return cmd
}

func runViewUpdate(cmd *cobra.Command, args []string) error {
	start := time.Now()

	name, _ := cmd.Flags().GetString("name")
	filterStr, _ := cmd.Flags().GetString("filter")
	sortsStr, _ := cmd.Flags().GetString("sorts")

	if name == "" && filterStr == "" && sortsStr == "" {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeMissingRequired,
			Message: "At least one of --name, --filter, or --sorts is required",
			Suggestions: []string{
				"Use --name to rename the view",
				"Use --filter <json-or-@file> to replace the view filter",
				"Use --sorts <json-or-@file> to replace the view sorts",
			},
		})
	}

	id, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	body := map[string]any{}

	if name != "" {
		body["name"] = name
	}

	if filterStr != "" {
		var filter map[string]any
		if err := loadJSONFlag(filterStr, &filter); err != nil {
			return handleError(cmd, err)
		}
		body["filter"] = filter
	}

	if sortsStr != "" {
		var sorts []any
		if err := loadJSONFlag(sortsStr, &sorts); err != nil {
			return handleError(cmd, err)
		}
		body["sorts"] = sorts
	}

	result, err := client.ViewUpdate(cmd.Context(), id, body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "view update", start)
	return nil
}

// ---------------------------------------------------------------------------
// view delete
// ---------------------------------------------------------------------------

func newViewDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete <view_id>",
		Aliases: []string{"d", "rm"},
		Short:   "Delete a view",
		Long:    "Delete a Notion view by ID.",
		Args:    cobra.ExactArgs(1),
		RunE:    runViewDelete,
	}

	addOutputFlags(cmd)
	return cmd
}

func runViewDelete(cmd *cobra.Command, args []string) error {
	start := time.Now()

	id, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.ViewDelete(cmd.Context(), id)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "view delete", start)
	return nil
}

// ---------------------------------------------------------------------------
// view query
// ---------------------------------------------------------------------------

func newViewQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query <view_id>",
		Aliases: []string{"q"},
		Short:   "Create a view query",
		Long: "Execute a view's filter and sort configuration against its data source, caching the full result set on the server. " +
			"Without --all, returns the first page of results plus a query_id for paginating via `view results`. " +
			"With --all, paginates through all results automatically and deletes the query when done. " +
			"Cached results expire 15 minutes after creation.",
		Args: cobra.ExactArgs(1),
		RunE: runViewQuery,
	}

	cmd.Flags().Int("page-size", 0, "Number of results to return per page (max 100)")
	cmd.Flags().Bool("all", false, "Paginate through all results automatically and delete the query when done")

	addOutputFlags(cmd)
	return cmd
}

func runViewQuery(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	id, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	pageSize, _ := cmd.Flags().GetInt("page-size")
	all, _ := cmd.Flags().GetBool("all")

	body := map[string]any{}
	if pageSize > 0 {
		body["page_size"] = pageSize
	}

	// Create the initial query — returns first page + query_id.
	first, err := client.ViewQueryCreate(cmd.Context(), id, body)
	if err != nil {
		return handleError(cmd, err)
	}

	if !all {
		p := output.NewPrinter(outputFormat(cmd))
		p.PrintSuccess(first, "view query", start)
		return nil
	}

	// --all: paginate through every page, then delete the query.
	queryID, _ := first["query_id"].(string)
	allResults, _ := first["results"].([]any)

	pageCount := 1
	qp := notion.QueryParams{}
	if pageSize > 0 {
		qp.PageSize = pageSize
	}

	for {
		hasMore, _ := first["has_more"].(bool)
		nextCursor, _ := first["next_cursor"].(string)
		if !hasMore || nextCursor == "" {
			break
		}

		pageCount++
		if pageCount > maxPaginationPages {
			fmt.Fprintf(os.Stderr, "Warning: reached maximum pagination limit (%d pages)\n", maxPaginationPages)
			break
		}

		qp.StartCursor = nextCursor
		first, err = client.ViewQueryResults(cmd.Context(), id, queryID, qp)
		if err != nil {
			// Best-effort cleanup before returning error.
			if queryID != "" {
				_, _ = client.ViewQueryDelete(cmd.Context(), id, queryID)
			}
			return handleError(cmd, err)
		}

		page, _ := first["results"].([]any)
		allResults = append(allResults, page...)
	}

	// Clean up the server-side query cache.
	if queryID != "" {
		_, _ = client.ViewQueryDelete(cmd.Context(), id, queryID)
	}

	data := map[string]any{
		"object":       "list",
		"results":      allResults,
		"has_more":     false,
		"next_cursor":  nil,
		"result_count": len(allResults),
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "view query", start)
	return nil
}

// ---------------------------------------------------------------------------
// view results
// ---------------------------------------------------------------------------

func newViewResultsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "results <view_id> <query_id>",
		Short: "Paginate through cached view query results",
		Long:  "Paginate through results of a previously created view query. Cached results expire 15 minutes after the query was created; expired queries return a 404.",
		Args:  cobra.ExactArgs(2),
		RunE:  runViewResults,
	}

	cmd.Flags().Int("page-size", 0, "Number of results per page (max 100)")
	cmd.Flags().String("start-cursor", "", "Pagination cursor")

	addOutputFlags(cmd)
	return cmd
}

func runViewResults(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	viewID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}
	queryID, err := resolveID(args[1])
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

	result, err := client.ViewQueryResults(cmd.Context(), viewID, queryID, qp)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "view results", start)
	return nil
}

// ---------------------------------------------------------------------------
// view delete-query
// ---------------------------------------------------------------------------

func newViewDeleteQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "delete-query <view_id> <query_id>",
		Aliases: []string{"dq"},
		Short:   "Delete a cached view query",
		Long:    "Delete a cached view query. Idempotent — returns success even if the query doesn't exist or has already expired.",
		Args:    cobra.ExactArgs(2),
		RunE:    runViewDeleteQuery,
	}

	addOutputFlags(cmd)
	return cmd
}

func runViewDeleteQuery(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	viewID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}
	queryID, err := resolveID(args[1])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.ViewQueryDelete(cmd.Context(), viewID, queryID)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "view delete-query", start)
	return nil
}
