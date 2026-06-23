package commands

import (
	"strings"

	"github.com/Coastal-Programs/notion-cli/v6/internal/cache"
	"github.com/Coastal-Programs/notion-cli/v6/internal/config"
	"github.com/spf13/cobra"
)

// AddAuthWorkspaceFlag adds the global auth workspace selector. It intentionally
// does not use --workspace because page move already uses that flag for moving
// a page to the Notion workspace root.
func AddAuthWorkspaceFlag(cmd *cobra.Command) {
	cmd.PersistentFlags().String("auth-workspace", "", "Target stored Notion workspace credential")
}

func authWorkspaceFromCommand(cmd *cobra.Command) string {
	if cmd == nil {
		return ""
	}
	flag := cmd.Flags().Lookup("auth-workspace")
	if flag == nil {
		flag = cmd.InheritedFlags().Lookup("auth-workspace")
	}
	if flag == nil && cmd.Root() != nil {
		flag = cmd.Root().PersistentFlags().Lookup("auth-workspace")
	}
	if flag == nil {
		return ""
	}
	return strings.TrimSpace(flag.Value.String())
}

func loadConfigForCommand(cmd *cobra.Command) (*config.Config, *config.ActiveWorkspace, error) {
	return config.LoadConfigForWorkspace(authWorkspaceFromCommand(cmd))
}

func workspaceCacheForCommand(cmd *cobra.Command) (*cache.WorkspaceCache, *config.ActiveWorkspace, error) {
	_, active, err := loadConfigForCommand(cmd)
	if err != nil {
		return nil, nil, err
	}
	return cache.NewWorkspaceCacheForWorkspace(active.Slug), active, nil
}
