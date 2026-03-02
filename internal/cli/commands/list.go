package commands

import (
	"fmt"
	"time"

	"github.com/Coastal-Programs/notion-cli/internal/cache"
	clierrors "github.com/Coastal-Programs/notion-cli/internal/errors"
	"github.com/Coastal-Programs/notion-cli/pkg/output"
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
		Short:   "List cached databases",
		Long:    "List all databases from the local workspace cache. Run 'sync' first to populate.",
		RunE:    runList,
	}

	addOutputFlags(cmd)
	return cmd
}

func runList(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	wc := cache.NewWorkspaceCache()
	if err := wc.Load(); err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: fmt.Sprintf("Failed to load workspace cache: %s", err),
		})
	}

	if wc.Count() == 0 {
		return handleError(cmd, clierrors.WorkspaceNotSynced())
	}

	if wc.IsStale() {
		fmt.Fprintln(cmd.ErrOrStderr(), "Warning: Cache is stale (>24h old). Run 'notion-cli sync' to refresh.")
	}

	dbs := wc.GetDatabases()
	var rows []map[string]any
	for _, db := range dbs {
		rows = append(rows, map[string]any{
			"Title":      db.Title,
			"ID":         db.ID,
			"LastEdited": db.LastEdited.Format(time.RFC3339),
		})
	}

	data := map[string]any{
		"databases":    rows,
		"count":        len(rows),
		"last_sync":    wc.LastSyncTime().Format(time.RFC3339),
		"cache_stale":  wc.IsStale(),
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "list", start)
	return nil
}
