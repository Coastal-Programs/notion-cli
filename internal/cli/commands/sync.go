package commands

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Coastal-Programs/notion-cli/internal/cache"
	clierrors "github.com/Coastal-Programs/notion-cli/internal/errors"
	"github.com/Coastal-Programs/notion-cli/pkg/output"
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
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: fmt.Sprintf("Failed to load workspace cache: %s", err),
		})
	}

	force, _ := cmd.Flags().GetBool("force")
	if !force && !wc.IsStale() && wc.Count() > 0 {
		p := output.NewPrinter(outputFormat(cmd))
		p.PrintSuccess(map[string]any{
			"message":   "Cache is fresh, skipping sync (use --force to override)",
			"databases": wc.Count(),
			"last_sync": wc.LastSyncTime().Format(time.RFC3339),
		}, "sync", start)
		return nil
	}

	fmt.Fprintln(cmd.OutOrStderr(), "Syncing workspace databases...")

	var allDatabases []cache.DatabaseEntry
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
				"value":    "database",
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
			db, ok := r.(map[string]any)
			if !ok {
				continue
			}

			entry := cache.DatabaseEntry{
				ID: fmt.Sprintf("%v", db["id"]),
			}

			// Extract title.
			if titleArr, ok := db["title"].([]any); ok {
				var parts []string
				for _, t := range titleArr {
					if tm, ok := t.(map[string]any); ok {
						if pt, ok := tm["plain_text"].(string); ok {
							parts = append(parts, pt)
						}
					}
				}
				entry.Title = strings.Join(parts, "")
			}

			if url, ok := db["url"].(string); ok {
				entry.URL = url
			}

			if let, ok := db["last_edited_time"].(string); ok {
				if t, err := time.Parse(time.RFC3339, let); err == nil {
					entry.LastEdited = t
				}
			}

			// Generate aliases from title.
			entry.Aliases = generateAliases(entry.Title)

			allDatabases = append(allDatabases, entry)
		}

		fmt.Fprintf(cmd.OutOrStderr(), "  Found %d databases so far...\n", len(allDatabases))

		hasMore, _ := result["has_more"].(bool)
		next, _ := result["next_cursor"].(string)
		if !hasMore || next == "" {
			break
		}
		startCursor = next
	}

	wc.SetDatabases(allDatabases)
	if err := wc.Save(); err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: fmt.Sprintf("Failed to save workspace cache: %s", err),
		})
	}

	elapsed := time.Since(start)
	fmt.Fprintf(cmd.OutOrStderr(), "Synced %d databases in %s\n", len(allDatabases), elapsed.Round(time.Millisecond))

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(map[string]any{
		"databases":    len(allDatabases),
		"last_sync":    time.Now().Format(time.RFC3339),
		"elapsed_ms":   elapsed.Milliseconds(),
	}, "sync", start)
	return nil
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
