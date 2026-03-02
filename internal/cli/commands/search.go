package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/internal/errors"
	"github.com/Coastal-Programs/notion-cli/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterSearchCommand registers the search command under root.
func RegisterSearchCommand(root *cobra.Command) {
	root.AddCommand(newSearchCmd())
}

func newSearchCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "search",
		Aliases: []string{"s", "find"},
		Short:   "Search by title",
		Long:    "Search across all pages and databases the integration has access to.",
		Args:    cobra.NoArgs,
		RunE:    runSearch,
	}

	cmd.Flags().StringP("query", "q", "", "Search query text")
	cmd.Flags().StringP("sort-direction", "d", "desc", "Sort direction (asc or desc)")
	cmd.Flags().StringP("property", "p", "", "Filter by object type (database or page)")
	cmd.Flags().StringP("start-cursor", "c", "", "Pagination cursor")
	cmd.Flags().IntP("page-size", "s", 5, "Number of results per page")
	cmd.Flags().String("database", "", "Filter results by database ID")
	cmd.Flags().String("created-after", "", "Filter: created after date (YYYY-MM-DD)")
	cmd.Flags().String("created-before", "", "Filter: created before date (YYYY-MM-DD)")
	cmd.Flags().String("edited-after", "", "Filter: last edited after date (YYYY-MM-DD)")
	cmd.Flags().String("edited-before", "", "Filter: last edited before date (YYYY-MM-DD)")
	cmd.Flags().Int("limit", 0, "Maximum total results to return")
	addOutputFlags(cmd)

	return cmd
}

func runSearch(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	body := map[string]any{}

	if q, _ := cmd.Flags().GetString("query"); q != "" {
		body["query"] = q
	}

	sortDir, _ := cmd.Flags().GetString("sort-direction")
	if sortDir == "asc" || sortDir == "ascending" {
		body["sort"] = map[string]any{
			"direction": "ascending",
			"timestamp": "last_edited_time",
		}
	} else {
		body["sort"] = map[string]any{
			"direction": "descending",
			"timestamp": "last_edited_time",
		}
	}

	if prop, _ := cmd.Flags().GetString("property"); prop != "" {
		switch strings.ToLower(prop) {
		case "database", "databases", "db":
			body["filter"] = map[string]any{
				"value":    "database",
				"property": "object",
			}
		case "page", "pages":
			body["filter"] = map[string]any{
				"value":    "page",
				"property": "object",
			}
		default:
			return handleError(cmd, &clierrors.NotionCLIError{
				Code:    clierrors.CodeInvalidRequest,
				Message: fmt.Sprintf("Invalid --property value %q: must be 'database' or 'page'", prop),
				Suggestions: []string{
					"Use --property database to filter for databases",
					"Use --property page to filter for pages",
				},
			})
		}
	}

	if ps, _ := cmd.Flags().GetInt("page-size"); ps > 0 {
		body["page_size"] = ps
	}

	if sc, _ := cmd.Flags().GetString("start-cursor"); sc != "" {
		body["start_cursor"] = sc
	}

	if ps, _ := cmd.Flags().GetInt("page-size"); ps > 100 {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: fmt.Sprintf("--page-size must be between 1 and 100, got %d", ps),
		})
	}

	limit, _ := cmd.Flags().GetInt("limit")
	dbFilter, _ := cmd.Flags().GetString("database")
	createdAfter, _ := cmd.Flags().GetString("created-after")
	createdBefore, _ := cmd.Flags().GetString("created-before")
	editedAfter, _ := cmd.Flags().GetString("edited-after")
	editedBefore, _ := cmd.Flags().GetString("edited-before")

	// Validate date filter formats upfront.
	for _, df := range []struct{ flag, val string }{
		{"--created-after", createdAfter},
		{"--created-before", createdBefore},
		{"--edited-after", editedAfter},
		{"--edited-before", editedBefore},
	} {
		if df.val != "" {
			if _, err := time.Parse("2006-01-02", df.val); err != nil {
				return handleError(cmd, &clierrors.NotionCLIError{
					Code:    clierrors.CodeInvalidRequest,
					Message: fmt.Sprintf("Invalid date for %s: %q (expected YYYY-MM-DD)", df.flag, df.val),
				})
			}
		}
	}

	needsClientFilter := dbFilter != "" || createdAfter != "" || createdBefore != "" || editedAfter != "" || editedBefore != ""

	p := output.NewPrinter(outputFormat(cmd))
	var allResults []any
	pageCount := 0

	for {
		pageCount++
		if pageCount > maxPaginationPages {
			fmt.Fprintf(os.Stderr, "Warning: reached maximum pagination limit (%d pages)\n", maxPaginationPages)
			break
		}

		result, err := client.Search(cmd.Context(), body)
		if err != nil {
			return handleError(cmd, err)
		}

		results, _ := result["results"].([]any)

		if needsClientFilter {
			results = clientSideFilter(results, dbFilter, createdAfter, createdBefore, editedAfter, editedBefore)
		}

		allResults = append(allResults, results...)

		if limit > 0 && len(allResults) >= limit {
			allResults = allResults[:limit]
			break
		}

		hasMore, _ := result["has_more"].(bool)
		nextCursor, _ := result["next_cursor"].(string)

		if !hasMore || nextCursor == "" {
			break
		}

		// Only continue paginating if we need more results for limit or client-side filters
		if limit == 0 && !needsClientFilter {
			break
		}
		body["start_cursor"] = nextCursor
	}

	data := map[string]any{
		"results":      allResults,
		"result_count": len(allResults),
	}

	p.PrintSuccess(data, "search", start)
	return nil
}

// clientSideFilter applies date and database filters to search results.
func clientSideFilter(results []any, dbFilter, createdAfter, createdBefore, editedAfter, editedBefore string) []any {
	var filtered []any
	for _, r := range results {
		item, ok := r.(map[string]any)
		if !ok {
			continue
		}

		// Database filter: check if page belongs to specified database
		if dbFilter != "" {
			parent, _ := item["parent"].(map[string]any)
			parentDB, _ := parent["database_id"].(string)
			if parentDB != dbFilter {
				continue
			}
		}

		// Date filters
		createdTime, _ := item["created_time"].(string)
		editedTime, _ := item["last_edited_time"].(string)

		if createdAfter != "" {
			afterDate, err := time.Parse("2006-01-02", createdAfter)
			if err == nil {
				created, err := time.Parse(time.RFC3339, createdTime)
				if err == nil && created.Before(afterDate) {
					continue
				}
			}
		}
		if createdBefore != "" {
			beforeDate, err := time.Parse("2006-01-02", createdBefore)
			if err == nil {
				created, err := time.Parse(time.RFC3339, createdTime)
				if err == nil && !created.Before(beforeDate) {
					continue
				}
			}
		}
		if editedAfter != "" {
			afterDate, err := time.Parse("2006-01-02", editedAfter)
			if err == nil {
				edited, err := time.Parse(time.RFC3339, editedTime)
				if err == nil && edited.Before(afterDate) {
					continue
				}
			}
		}
		if editedBefore != "" {
			beforeDate, err := time.Parse("2006-01-02", editedBefore)
			if err == nil {
				edited, err := time.Parse(time.RFC3339, editedTime)
				if err == nil && !edited.Before(beforeDate) {
					continue
				}
			}
		}

		filtered = append(filtered, item)
	}
	return filtered
}


