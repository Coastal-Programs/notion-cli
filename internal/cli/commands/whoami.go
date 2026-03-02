package commands

import (
	"fmt"
	"time"

	"github.com/Coastal-Programs/notion-cli/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterWhoamiCommand registers the whoami command.
func RegisterWhoamiCommand(root *cobra.Command) {
	cmd := &cobra.Command{
		Use:     "whoami",
		Aliases: []string{"test", "health", "connectivity"},
		Short:   "Check API connectivity",
		Long:    "Verify your Notion API token and show bot info, workspace, and API latency.",
		Args:    cobra.NoArgs,
		RunE:    runWhoami,
	}
	addOutputFlags(cmd)
	root.AddCommand(cmd)
}

func runWhoami(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	apiStart := time.Now()
	result, err := client.UsersMe(cmd.Context())
	apiLatency := time.Since(apiStart)

	if err != nil {
		return handleError(cmd, err)
	}

	// Extract bot info.
	data := map[string]any{
		"id":          result["id"],
		"name":        result["name"],
		"type":        result["type"],
		"api_latency": fmt.Sprintf("%dms", apiLatency.Milliseconds()),
	}

	if bot, ok := result["bot"].(map[string]any); ok {
		if owner, ok := bot["owner"].(map[string]any); ok {
			if ws, ok := owner["workspace"].(bool); ok && ws {
				data["workspace"] = "workspace-level integration"
			}
			data["owner_type"] = owner["type"]
		}
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "whoami", start)
	return nil
}
