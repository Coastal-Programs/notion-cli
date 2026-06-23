package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Coastal-Programs/notion-cli/v6/internal/cache"
	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterSyncCommand registers the sync command under root.
func RegisterSyncCommand(root *cobra.Command) {
	root.AddCommand(newSyncCmd())
}

func newSyncCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "sync",
		Aliases: []string{"db:sync"},
		Short:   "Sync workspace databases to local cache",
		Long:    "Search for all databases shared with the integration and cache them locally.",
		Args:    cobra.NoArgs,
		RunE:    runSync,
	}

	cmd.Flags().BoolP("force", "f", false, "Force re-sync even if cache is fresh")
	addOutputFlags(cmd)

	return cmd
}

func runSync(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	wc := cache.NewWorkspaceCache()
	if err := wc.Load(); err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to load workspace cache", err))
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force && !wc.IsStale() && wc.Count() > 0 {
		p := output.NewPrinter(outputFormat(cmd))
		p.PrintSuccess(map[string]any{
			"message":      "Cache is fresh, skipping sync (use --force to override)",
			"databases":    wc.Count(),
			"data_sources": wc.DataSourceCount(),
			"last_sync":    wc.LastSyncTime().Format(time.RFC3339),
		}, "sync", start)
		return nil
	}

	_, _ = fmt.Fprintln(cmd.OutOrStderr(), "Syncing workspace databases...")

	var allDatabases []cache.DatabaseEntry
	var allDataSources []cache.DataSourceEntry
	var startCursor string
	pageCount := 0

	for {
		pageCount++
		if pageCount > maxPaginationPages {
			fmt.Fprintf(os.Stderr, "Warning: reached maximum pagination limit (%d pages)\n", maxPaginationPages)
			break
		}

		body := map[string]any{
			"filter": map[string]any{
				"value":    "data_source",
				"property": "object",
			},
			"page_size": 100,
		}
		if startCursor != "" {
			body["start_cursor"] = startCursor
		}

		result, err := client.Search(cmd.Context(), body)
		if err != nil {
			return handleError(cmd, err)
		}

		results, _ := result["results"].([]any)
		for _, r := range results {
			resultMap, ok := r.(map[string]any)
			if !ok {
				continue
			}
			entry, dataSources, ok := cacheEntriesFromSearchResult(resultMap)
			if !ok {
				continue
			}
			allDatabases = append(allDatabases, entry)
			allDataSources = append(allDataSources, dataSources...)
		}

		_, _ = fmt.Fprintf(cmd.OutOrStderr(), "  Found %d databases so far...\n", len(allDatabases))

		hasMore, _ := result["has_more"].(bool)
		next, _ := result["next_cursor"].(string)
		if !hasMore || next == "" {
			break
		}
		startCursor = next
	}

	wc.SetDatabases(allDatabases)
	wc.SetDataSources(allDataSources)
	if err := wc.Save(); err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to save workspace cache", err))
	}

	elapsed := time.Since(start)
	_, _ = fmt.Fprintf(cmd.OutOrStderr(), "Synced %d databases (%d data sources) in %s\n", len(allDatabases), len(allDataSources), elapsed.Round(time.Millisecond))

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(map[string]any{
		"databases":    len(allDatabases),
		"data_sources": len(allDataSources),
		"last_sync":    time.Now().Format(time.RFC3339),
		"elapsed_ms":   elapsed.Milliseconds(),
	}, "sync", start)
	return nil
}

func cacheEntriesFromSearchResult(result map[string]any) (cache.DatabaseEntry, []cache.DataSourceEntry, bool) {
	switch result["object"] {
	case "data_source":
		return cacheEntriesFromDataSourceSearchResult(result)
	case "database":
		return cacheEntriesFromDatabaseSearchResult(result)
	default:
		return cache.DatabaseEntry{}, nil, false
	}
}

func cacheEntriesFromDataSourceSearchResult(ds map[string]any) (cache.DatabaseEntry, []cache.DataSourceEntry, bool) {
	dataSourceID := stringField(ds, "id")
	if dataSourceID == "" {
		return cache.DatabaseEntry{}, nil, false
	}

	databaseID := parentDatabaseID(ds)
	if databaseID == "" {
		databaseID = dataSourceID
	}

	title := titleFromSearchResult(ds)
	lastEdited := parseNotionEditedTime(ds)
	url := stringField(ds, "url")
	dbEntry := cache.DatabaseEntry{
		ID:           databaseID,
		Title:        title,
		DataSourceID: dataSourceID,
		URL:          url,
		Aliases:      generateAliases(title),
		LastEdited:   lastEdited,
	}
	dsEntry := cache.DataSourceEntry{
		ID:         dataSourceID,
		DatabaseID: databaseID,
		Title:      title,
		URL:        url,
		LastEdited: lastEdited,
	}
	return dbEntry, []cache.DataSourceEntry{dsEntry}, true
}

func cacheEntriesFromDatabaseSearchResult(db map[string]any) (cache.DatabaseEntry, []cache.DataSourceEntry, bool) {
	databaseID := stringField(db, "id")
	if databaseID == "" {
		return cache.DatabaseEntry{}, nil, false
	}

	title := titleFromSearchResult(db)
	lastEdited := parseNotionEditedTime(db)
	url := stringField(db, "url")
	dbEntry := cache.DatabaseEntry{
		ID:         databaseID,
		Title:      title,
		URL:        url,
		Aliases:    generateAliases(title),
		LastEdited: lastEdited,
	}

	var dataSources []cache.DataSourceEntry
	if dsArr, ok := db["data_sources"].([]any); ok {
		for _, rawDS := range dsArr {
			dsMap, ok := rawDS.(map[string]any)
			if !ok {
				continue
			}
			dataSourceID := stringField(dsMap, "id")
			if dataSourceID == "" {
				continue
			}
			if dbEntry.DataSourceID == "" {
				dbEntry.DataSourceID = dataSourceID
			}
			dsTitle := titleFromSearchResult(dsMap)
			if dsTitle == "" {
				dsTitle = title
			}
			dsURL := stringField(dsMap, "url")
			if dsURL == "" {
				dsURL = url
			}
			dsLastEdited := parseNotionEditedTime(dsMap)
			if dsLastEdited.IsZero() {
				dsLastEdited = lastEdited
			}
			dataSources = append(dataSources, cache.DataSourceEntry{
				ID:         dataSourceID,
				DatabaseID: databaseID,
				Title:      dsTitle,
				URL:        dsURL,
				LastEdited: dsLastEdited,
			})
		}
	}
	return dbEntry, dataSources, true
}

func titleFromSearchResult(item map[string]any) string {
	titleArr, ok := item["title"].([]any)
	if !ok {
		return ""
	}
	var parts []string
	for _, t := range titleArr {
		tm, ok := t.(map[string]any)
		if !ok {
			continue
		}
		if plainText, ok := tm["plain_text"].(string); ok {
			parts = append(parts, plainText)
		}
	}
	return strings.Join(parts, "")
}

func parentDatabaseID(item map[string]any) string {
	parent, ok := item["parent"].(map[string]any)
	if !ok {
		return ""
	}
	return stringField(parent, "database_id")
}

func parseNotionEditedTime(item map[string]any) time.Time {
	edited := stringField(item, "last_edited_time")
	if edited == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, edited)
	if err != nil {
		return time.Time{}
	}
	return t
}

func stringField(item map[string]any, key string) string {
	value, ok := item[key].(string)
	if !ok {
		return ""
	}
	return value
}

// generateAliases creates search aliases from a database title.
func generateAliases(title string) []string {
	if title == "" {
		return nil
	}

	aliases := []string{strings.ToLower(title)}

	// Generate acronym from words.
	words := strings.Fields(title)
	if len(words) > 1 {
		var acronym strings.Builder
		for _, w := range words {
			if len(w) > 0 {
				acronym.WriteByte(w[0])
			}
		}
		aliases = append(aliases, strings.ToLower(acronym.String()))
	}

	return aliases
}
