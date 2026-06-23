package commands

import (
	"fmt"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterListCommand registers the list command under root.
func RegisterListCommand(root *cobra.Command) {
	root.AddCommand(newListCmd())
}

func newListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"db:list", "ls"},
		Short:   "List cached databases and data sources",
		Long:    "List all databases and data sources from the local workspace cache. Run 'sync' first to populate.",
		Args:    cobra.NoArgs,
		RunE:    runList,
	}

	addOutputFlags(cmd)
	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	wc, active, err := workspaceCacheForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}
	if err := wc.Load(); err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to load workspace cache", err))
	}

	if wc.Count() == 0 {
		return handleError(cmd, clierrors.WorkspaceNotSynced())
	}

	if wc.IsStale() {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Warning: Cache is stale (>24h old). Run 'notion-cli sync' to refresh.")
	}

	dbs := wc.GetDatabases()
	var dbRows []map[string]any
	for _, db := range dbs {
		dbRows = append(dbRows, map[string]any{
			"Title":      db.Title,
			"ID":         db.ID,
			"LastEdited": db.LastEdited.Format(time.RFC3339),
		})
	}

	dsEntries := wc.GetDataSources()
	var dsRows []map[string]any
	for _, ds := range dsEntries {
		dsRows = append(dsRows, map[string]any{
			"Title":      ds.Title,
			"ID":         ds.ID,
			"DatabaseID": ds.DatabaseID,
			"LastEdited": ds.LastEdited.Format(time.RFC3339),
		})
	}

	data := map[string]any{
		"databases":    dbRows,
		"data_sources": dsRows,
		"count":        len(dbRows),
		"last_sync":    wc.LastSyncTime().Format(time.RFC3339),
		"cache_stale":  wc.IsStale(),
		"workspace":    active.DisplayName(),
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "list", start)
	return nil
}
