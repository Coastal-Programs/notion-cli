package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Coastal-Programs/notion-cli/internal/config"
	clierrors "github.com/Coastal-Programs/notion-cli/internal/errors"
	"github.com/Coastal-Programs/notion-cli/internal/notion"
	"github.com/Coastal-Programs/notion-cli/internal/resolver"
	"github.com/Coastal-Programs/notion-cli/pkg/output"
	"github.com/spf13/cobra"
)

// newClient creates a Notion API client from the NOTION_TOKEN env var,
// falling back to the config file token if the env var is not set.
func newClient() (*notion.Client, error) {
	token := os.Getenv("NOTION_TOKEN")
	if token == "" {
		// Fallback to config file.
		cfg, err := config.LoadConfig()
		if err == nil && cfg.Token != "" {
			token = cfg.Token
		}
	}
	if token == "" {
		return nil, &clierrors.NotionCLIError{
			Code:    clierrors.CodeTokenMissing,
			Message: "NOTION_TOKEN environment variable is not set",
			Suggestions: []string{
				"Export your token: export NOTION_TOKEN=secret_...",
				"Or run: notion-cli config set-token <your_token>",
			},
		}
	}
	return notion.NewClient(token), nil
}

// resolveID extracts and validates a Notion ID from the provided argument.
func resolveID(raw string) (string, error) {
	id, err := resolver.ExtractID(raw)
	if err != nil {
		return "", clierrors.InvalidIDFormat(raw)
	}
	return id, nil
}

// outputFormat returns the output.Format from command flags.
func outputFormat(cmd *cobra.Command) output.Format {
	if v, _ := cmd.Flags().GetBool("json"); v {
		return output.FormatJSON
	}
	if v, _ := cmd.Flags().GetBool("compact-json"); v {
		return output.FormatCompactJSON
	}
	if v, _ := cmd.Flags().GetBool("raw"); v {
		return output.FormatRaw
	}
	if v, _ := cmd.Flags().GetBool("csv"); v {
		return output.FormatCSV
	}
	if v, _ := cmd.Flags().GetBool("markdown"); v {
		return output.FormatMarkdown
	}
	if v, _ := cmd.Flags().GetBool("pretty"); v {
		return output.FormatPretty
	}
	return output.FormatTable
}

// handleError prints an error using the output printer and returns it for
// cobra to handle the exit code.
func handleError(cmd *cobra.Command, err error) error {
	p := output.NewPrinter(outputFormat(cmd))

	// Check for NotionCLIError first.
	if cliErr, ok := err.(*clierrors.NotionCLIError); ok {
		p.PrintError(cliErr.Code, cliErr.Message, cliErr.Details, cliErr.Suggestions)
		return err
	}

	// Check for Notion API errors and translate them.
	if apiErr, ok := err.(*notion.APIError); ok {
		body := map[string]any{
			"code":    apiErr.Code,
			"message": apiErr.Message,
		}
		cliErr := clierrors.FromNotionAPI(apiErr.Status, body)
		p.PrintError(cliErr.Code, cliErr.Message, cliErr.Details, cliErr.Suggestions)
		return cliErr
	}

	// Generic error.
	cliErr := &clierrors.NotionCLIError{
		Code:    clierrors.CodeInternalError,
		Message: err.Error(),
	}
	p.PrintError(cliErr.Code, cliErr.Message, nil, nil)
	return cliErr
}

// addOutputFlags adds the shared output format flags to a command.
func addOutputFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Output as JSON envelope")
	cmd.Flags().Bool("compact-json", false, "Output as compact JSON (single line)")
	cmd.Flags().Bool("raw", false, "Output raw API response without envelope")
	cmd.Flags().Bool("csv", false, "Output as CSV")
	cmd.Flags().Bool("markdown", false, "Output as markdown table")
	cmd.Flags().Bool("pretty", false, "Output as pretty-printed JSON")
	cmd.MarkFlagsMutuallyExclusive("json", "compact-json", "csv", "markdown", "raw", "pretty")
}

// maxPaginationPages is the safety limit for pagination loops.
const maxPaginationPages = 1000

// readOnlyPropertyTypes identifies Notion property types that cannot be written to.
var readOnlyPropertyTypes = map[string]bool{
	"formula":          true,
	"rollup":           true,
	"created_time":     true,
	"created_by":       true,
	"last_edited_time": true,
	"last_edited_by":   true,
	"unique_id":        true,
}

// RegisterDBCommands registers all database subcommands under root.
func RegisterDBCommands(root *cobra.Command) {
	dbCmd := &cobra.Command{
		Use:     "db",
		Aliases: []string{"ds", "database"},
		Short:   "Database operations",
		Long:    "Query, retrieve, create, update, and inspect Notion databases.",
	}

	dbCmd.AddCommand(
		newDBQueryCmd(),
		newDBRetrieveCmd(),
		newDBSchemaCmd(),
		newDBCreateCmd(),
		newDBUpdateCmd(),
	)

	root.AddCommand(dbCmd)
}

// --- db query ---

func newDBQueryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "query <database_id>",
		Aliases: []string{"q"},
		Short:   "Query a database",
		Long:    "Query a Notion database with optional filters, sorts, and pagination.",
		Args:    cobra.ExactArgs(1),
		RunE:    runDBQuery,
	}

	cmd.Flags().Int("page-size", 10, "Number of results per page")
	cmd.Flags().Bool("page-all", false, "Fetch all pages (auto-paginate)")
	cmd.Flags().String("sort-property", "", "Property to sort by")
	cmd.Flags().String("sort-direction", "", "Sort direction (asc or desc)")
	cmd.Flags().String("filter", "", "Filter as JSON string")
	cmd.Flags().String("file-filter", "", "Path to JSON file containing filter")
	cmd.Flags().String("search", "", "Search text within the database")
	cmd.Flags().String("select", "", "Comma-separated list of properties to include")
	addOutputFlags(cmd)

	return cmd
}

func runDBQuery(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	dbID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	body := map[string]any{}

	if ps, _ := cmd.Flags().GetInt("page-size"); ps > 0 {
		if ps > 100 {
			return handleError(cmd, &clierrors.NotionCLIError{
				Code:    clierrors.CodeInvalidRequest,
				Message: fmt.Sprintf("--page-size must be between 1 and 100, got %d", ps),
			})
		}
		body["page_size"] = ps
	}

	if sp, _ := cmd.Flags().GetString("sort-property"); sp != "" {
		dir, _ := cmd.Flags().GetString("sort-direction")
		if dir == "" {
			dir = "ascending"
		} else if dir == "asc" {
			dir = "ascending"
		} else if dir == "desc" {
			dir = "descending"
		}
		body["sorts"] = []map[string]any{
			{"property": sp, "direction": dir},
		}
	}

	// Filter: from --filter flag or --file-filter
	if f, _ := cmd.Flags().GetString("filter"); f != "" {
		var filter map[string]any
		if err := json.Unmarshal([]byte(f), &filter); err != nil {
			return handleError(cmd, clierrors.InvalidJSON(fmt.Sprintf("--filter: %s", err)))
		}
		body["filter"] = filter
	} else if ff, _ := cmd.Flags().GetString("file-filter"); ff != "" {
		data, err := os.ReadFile(ff)
		if err != nil {
			return handleError(cmd, &clierrors.NotionCLIError{
				Code:    clierrors.CodeInvalidRequest,
				Message: fmt.Sprintf("Cannot read filter file: %s", err),
			})
		}
		var filter map[string]any
		if err := json.Unmarshal(data, &filter); err != nil {
			return handleError(cmd, clierrors.InvalidJSON(fmt.Sprintf("--file-filter: %s", err)))
		}
		body["filter"] = filter
	}

	pageAll, _ := cmd.Flags().GetBool("page-all")
	p := output.NewPrinter(outputFormat(cmd))

	var allResults []any
	pageCount := 0

	for {
		pageCount++
		if pageCount > maxPaginationPages {
			fmt.Fprintf(os.Stderr, "Warning: reached maximum pagination limit (%d pages)\n", maxPaginationPages)
			break
		}

		result, err := client.DatabaseQuery(cmd.Context(), dbID, body)
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
		body["start_cursor"] = nextCursor
	}

	// Apply --select filter on properties.
	selectProp, _ := cmd.Flags().GetString("select")
	if selectProp != "" {
		selected := strings.Split(selectProp, ",")
		for i := range selected {
			selected[i] = strings.TrimSpace(selected[i])
		}
		allResults = filterProperties(allResults, selected)
	}

	data := map[string]any{
		"results":     allResults,
		"result_count": len(allResults),
	}

	p.PrintSuccess(data, "db query", start)
	return nil
}

// filterProperties removes non-selected properties from page results.
func filterProperties(results []any, selected []string) []any {
	want := map[string]bool{}
	for _, s := range selected {
		want[s] = true
	}

	for _, r := range results {
		page, ok := r.(map[string]any)
		if !ok {
			continue
		}
		props, ok := page["properties"].(map[string]any)
		if !ok {
			continue
		}
		for key := range props {
			if !want[key] {
				delete(props, key)
			}
		}
	}
	return results
}

// --- db retrieve ---

func newDBRetrieveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "retrieve <database_id>",
		Aliases: []string{"r", "get"},
		Short:   "Retrieve a database",
		Long:    "Retrieve a Notion database by ID, showing its schema and configuration.",
		Args:    cobra.ExactArgs(1),
		RunE:    runDBRetrieve,
	}
	addOutputFlags(cmd)
	return cmd
}

func runDBRetrieve(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	dbID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.DatabaseRetrieve(cmd.Context(), dbID)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "db retrieve", start)
	return nil
}

// --- db schema ---

func newDBSchemaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "schema <database_id>",
		Aliases: []string{"s"},
		Short:   "Extract database schema",
		Long:    "Extract a clean, AI-parseable schema from a Notion database.",
		Args:    cobra.ExactArgs(1),
		RunE:    runDBSchema,
	}

	cmd.Flags().String("properties", "", "Comma-separated list of properties to include")
	cmd.Flags().Bool("with-examples", false, "Include example property payloads")
	addOutputFlags(cmd)

	return cmd
}

func runDBSchema(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	dbID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.DatabaseRetrieve(cmd.Context(), dbID)
	if err != nil {
		return handleError(cmd, err)
	}

	schema := extractSchema(result)

	// Filter properties if --properties is specified.
	if propFilter, _ := cmd.Flags().GetString("properties"); propFilter != "" {
		names := strings.Split(propFilter, ",")
		for i := range names {
			names[i] = strings.TrimSpace(names[i])
		}
		schema["properties"] = filterSchemaProperties(schema["properties"].([]map[string]any), names)
	}

	// Add examples if requested.
	if withExamples, _ := cmd.Flags().GetBool("with-examples"); withExamples {
		props, _ := schema["properties"].([]map[string]any)
		for i, prop := range props {
			props[i]["example"] = generateExample(prop)
		}
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(schema, "db schema", start)
	return nil
}

// extractSchema transforms a Notion database object into a clean schema.
func extractSchema(db map[string]any) map[string]any {
	schema := map[string]any{
		"id":    db["id"],
		"title": extractPlainText(db["title"]),
	}

	props, _ := db["properties"].(map[string]any)
	var propList []map[string]any

	for name, v := range props {
		prop, ok := v.(map[string]any)
		if !ok {
			continue
		}
		entry := map[string]any{
			"name": name,
			"type": prop["type"],
		}

		propType, _ := prop["type"].(string)

		// Extract options for select/multi_select/status.
		if propType == "select" || propType == "multi_select" || propType == "status" {
			if typeData, ok := prop[propType].(map[string]any); ok {
				if options, ok := typeData["options"].([]any); ok {
					var optNames []string
					for _, opt := range options {
						if o, ok := opt.(map[string]any); ok {
							if n, ok := o["name"].(string); ok {
								optNames = append(optNames, n)
							}
						}
					}
					entry["options"] = optNames
				}
				// Status groups.
				if propType == "status" {
					if groups, ok := typeData["groups"].([]any); ok {
						var groupNames []string
						for _, g := range groups {
							if gm, ok := g.(map[string]any); ok {
								if n, ok := gm["name"].(string); ok {
									groupNames = append(groupNames, n)
								}
							}
						}
						entry["groups"] = groupNames
					}
				}
			}
		}

		// Extract relation target for relation properties.
		if propType == "relation" {
			if rel, ok := prop["relation"].(map[string]any); ok {
				entry["relation_database_id"] = rel["database_id"]
				entry["relation_type"] = rel["type"]
			}
		}

		// Extract formula expression.
		if propType == "formula" {
			if formula, ok := prop["formula"].(map[string]any); ok {
				entry["expression"] = formula["expression"]
			}
		}

		// Extract rollup config.
		if propType == "rollup" {
			if rollup, ok := prop["rollup"].(map[string]any); ok {
				entry["rollup_property"] = rollup["rollup_property_name"]
				entry["rollup_relation"] = rollup["relation_property_name"]
				entry["rollup_function"] = rollup["function"]
			}
		}

		// Mark read-only properties.
		if readOnlyPropertyTypes[propType] {
			entry["read_only"] = true
		}

		propList = append(propList, entry)
	}

	sort.Slice(propList, func(i, j int) bool {
		return propList[i]["name"].(string) < propList[j]["name"].(string)
	})

	schema["properties"] = propList
	schema["property_count"] = len(propList)
	return schema
}

// extractPlainText extracts plain text from a Notion title/rich_text array.
func extractPlainText(v any) string {
	arr, ok := v.([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, item := range arr {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if pt, ok := m["plain_text"].(string); ok {
			parts = append(parts, pt)
		}
	}
	return strings.Join(parts, "")
}

// filterSchemaProperties filters the property list to only include named properties.
func filterSchemaProperties(props []map[string]any, names []string) []map[string]any {
	want := map[string]bool{}
	for _, n := range names {
		want[strings.ToLower(n)] = true
	}
	var filtered []map[string]any
	for _, p := range props {
		name, _ := p["name"].(string)
		if want[strings.ToLower(name)] {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// generateExample creates an example property payload based on type.
func generateExample(prop map[string]any) any {
	propType, _ := prop["type"].(string)
	switch propType {
	case "title":
		return map[string]any{
			"title": []map[string]any{{"text": map[string]any{"content": "Example Title"}}},
		}
	case "rich_text":
		return map[string]any{
			"rich_text": []map[string]any{{"text": map[string]any{"content": "Example text"}}},
		}
	case "number":
		return map[string]any{"number": 42}
	case "checkbox":
		return map[string]any{"checkbox": true}
	case "select":
		options, _ := prop["options"].([]string)
		value := "Option"
		if len(options) > 0 {
			value = options[0]
		}
		return map[string]any{"select": map[string]any{"name": value}}
	case "multi_select":
		options, _ := prop["options"].([]string)
		value := "Option"
		if len(options) > 0 {
			value = options[0]
		}
		return map[string]any{"multi_select": []map[string]any{{"name": value}}}
	case "status":
		options, _ := prop["options"].([]string)
		value := "Not started"
		if len(options) > 0 {
			value = options[0]
		}
		return map[string]any{"status": map[string]any{"name": value}}
	case "date":
		return map[string]any{"date": map[string]any{"start": "2024-01-15"}}
	case "url":
		return map[string]any{"url": "https://example.com"}
	case "email":
		return map[string]any{"email": "user@example.com"}
	case "phone_number":
		return map[string]any{"phone_number": "+1-555-0100"}
	case "people":
		return map[string]any{"people": []map[string]any{{"id": "<user_id>"}}}
	case "files":
		return map[string]any{"files": []map[string]any{{"name": "file.pdf", "external": map[string]any{"url": "https://example.com/file.pdf"}}}}
	case "relation":
		return map[string]any{"relation": []map[string]any{{"id": "<page_id>"}}}
	default:
		return map[string]any{"note": fmt.Sprintf("%s is read-only", propType)}
	}
}

// --- db create ---

func newDBCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create <parent_page_id>",
		Aliases: []string{"c"},
		Short:   "Create a database",
		Long:    "Create a new Notion database under the specified parent page.",
		Args:    cobra.ExactArgs(1),
		RunE:    runDBCreate,
	}

	cmd.Flags().String("title", "", "Database title (required)")
	cmd.MarkFlagRequired("title")
	addOutputFlags(cmd)

	return cmd
}

func runDBCreate(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	pageID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	title, _ := cmd.Flags().GetString("title")

	body := map[string]any{
		"parent": map[string]any{
			"type":    "page_id",
			"page_id": pageID,
		},
		"title": []map[string]any{
			{"type": "text", "text": map[string]any{"content": title}},
		},
		"properties": map[string]any{
			"Name": map[string]any{
				"title": map[string]any{},
			},
		},
	}

	result, err := client.DatabaseCreate(cmd.Context(), body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "db create", start)
	return nil
}

// --- db update ---

func newDBUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "update <database_id>",
		Aliases: []string{"u"},
		Short:   "Update a database",
		Long:    "Update a Notion database's title.",
		Args:    cobra.ExactArgs(1),
		RunE:    runDBUpdate,
	}

	cmd.Flags().String("title", "", "New database title (required)")
	cmd.MarkFlagRequired("title")
	addOutputFlags(cmd)

	return cmd
}

func runDBUpdate(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	dbID, err := resolveID(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	title, _ := cmd.Flags().GetString("title")

	body := map[string]any{
		"title": []map[string]any{
			{"type": "text", "text": map[string]any{"content": title}},
		},
	}

	result, err := client.DatabaseUpdate(cmd.Context(), dbID, body)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "db update", start)
	return nil
}
