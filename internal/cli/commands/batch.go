package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Coastal-Programs/notion-cli/internal/notion"
	"github.com/Coastal-Programs/notion-cli/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterBatchCommands registers all batch subcommands under root.
func RegisterBatchCommands(root *cobra.Command) {
	batchCmd := &cobra.Command{
		Use:     "batch",
		Aliases: []string{"b"},
		Short:   "Batch operations",
		Long:    "Execute batch operations on multiple Notion resources.",
	}

	batchCmd.AddCommand(newBatchRetrieveCmd())
	root.AddCommand(batchCmd)
}

func newBatchRetrieveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "retrieve [ids...]",
		Aliases: []string{"r", "get"},
		Short:   "Batch retrieve multiple resources",
		Long:    "Retrieve multiple pages, blocks, or databases by ID. IDs can be passed as arguments, via --ids flag, or piped via stdin.",
		Args:    cobra.ArbitraryArgs,
		RunE:    runBatchRetrieve,
	}

	cmd.Flags().String("ids", "", "Comma-separated list of IDs to retrieve")
	cmd.Flags().String("type", "page", "Resource type: page, block, or database")
	addOutputFlags(cmd)

	return cmd
}

func runBatchRetrieve(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	// Collect IDs from all sources
	var ids []string

	// From positional args
	for _, arg := range args {
		for _, id := range strings.Split(arg, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				ids = append(ids, id)
			}
		}
	}

	// From --ids flag
	if idsFlag, _ := cmd.Flags().GetString("ids"); idsFlag != "" {
		for _, id := range strings.Split(idsFlag, ",") {
			id = strings.TrimSpace(id)
			if id != "" {
				ids = append(ids, id)
			}
		}
	}

	// From stdin (non-blocking: only read if data is available)
	if len(ids) == 0 {
		ids = append(ids, readStdinIDs()...)
	}

	if len(ids) == 0 {
		return handleError(cmd, fmt.Errorf("no IDs provided; pass IDs as arguments, via --ids, or pipe through stdin"))
	}

	resourceType, _ := cmd.Flags().GetString("type")

	// Resolve all IDs
	resolvedIDs := make([]string, len(ids))
	for i, raw := range ids {
		resolved, err := resolveID(raw)
		if err != nil {
			return handleError(cmd, err)
		}
		resolvedIDs[i] = resolved
	}

	// Retrieve concurrently with bounded concurrency
	type result struct {
		index int
		data  map[string]any
		err   error
	}

	results := make([]result, len(resolvedIDs))
	var wg sync.WaitGroup

	const maxConcurrency = 10
	sem := make(chan struct{}, maxConcurrency)

	for i, id := range resolvedIDs {
		wg.Add(1)
		sem <- struct{}{} // acquire
		go func(idx int, resourceID string) {
			defer wg.Done()
			defer func() { <-sem }() // release
			data, err := retrieveResource(cmd, client, resourceType, resourceID)
			results[idx] = result{index: idx, data: data, err: err}
		}(i, id)
	}

	wg.Wait()

	// Collect results
	var successResults []any
	var errors []map[string]any

	for _, r := range results {
		if r.err != nil {
			errors = append(errors, map[string]any{
				"id":    resolvedIDs[r.index],
				"error": r.err.Error(),
			})
		} else {
			successResults = append(successResults, r.data)
		}
	}

	data := map[string]any{
		"results":       successResults,
		"result_count":  len(successResults),
		"error_count":   len(errors),
		"total_requested": len(resolvedIDs),
	}
	if len(errors) > 0 {
		data["errors"] = errors
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "batch retrieve", start)
	return nil
}

// retrieveResource calls the appropriate Notion API method based on resource type.
func retrieveResource(cmd *cobra.Command, client *notion.Client, resourceType, id string) (map[string]any, error) {
	ctx := cmd.Context()
	switch strings.ToLower(resourceType) {
	case "page", "pages":
		return client.PageRetrieve(ctx, id)
	case "block", "blocks":
		return client.BlockRetrieve(ctx, id)
	case "database", "databases", "db":
		return client.DatabaseRetrieve(ctx, id)
	default:
		return nil, fmt.Errorf("unsupported resource type: %s (use page, block, or database)", resourceType)
	}
}

// readStdinIDs reads IDs from stdin if available (one per line).
// Returns immediately if stdin is a terminal (not piped).
func readStdinIDs() []string {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil
	}

	// Only read if stdin has data piped to it
	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil
	}

	var ids []string
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			ids = append(ids, line)
		}
	}
	return ids
}
