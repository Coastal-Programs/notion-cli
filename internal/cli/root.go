package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Coastal-Programs/notion-cli/v6/internal/cli/commands"
	"github.com/Coastal-Programs/notion-cli/v6/internal/config"
	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/spf13/cobra"
)

var startTime time.Time

var rootCmd = &cobra.Command{
	Use:           "notion-cli",
	Short:         "Unofficial CLI for the Notion API",
	Long:          "Unofficial CLI for the Notion API, optimized for AI agents and automation.",
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		startTime = time.Now()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		if !verbose {
			// Fall back to config (env var or config file) when the flag
			// is not explicitly set. This makes NOTION_CLI_VERBOSE work.
			if !cmd.Flags().Changed("verbose") {
				authWorkspace, _ := cmd.Root().PersistentFlags().GetString("auth-workspace")
				if cfg, _, err := config.LoadConfigForWorkspace(authWorkspace); err == nil && cfg != nil {
					verbose = cfg.Verbose
				}
			}
		}
		if verbose {
			elapsed := time.Since(startTime)
			fmt.Fprintf(os.Stderr, "Execution time: %s\n", elapsed.Round(time.Millisecond))
		}
	},
	Version: config.Version,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// ExecuteContext runs the root command with a context.
func ExecuteContext(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	// Global persistent flags (non-output flags only; output format flags
	// are added per-command via addOutputFlags to avoid redefinition panics).
	pf := rootCmd.PersistentFlags()
	pf.BoolP("verbose", "v", false, "Enable verbose stderr logging")
	commands.AddAuthWorkspaceFlag(rootCmd)

	// Version template.
	rootCmd.SetVersionTemplate(fmt.Sprintf("notion-cli version %s (commit: %s, built: %s)\n",
		config.Version, config.Commit, config.Date))

	// Register all command groups.
	commands.RegisterDBCommands(rootCmd)
	commands.RegisterDataSourceCommands(rootCmd)
	commands.RegisterPageCommands(rootCmd)
	commands.RegisterBlockCommands(rootCmd)
	commands.RegisterUserCommands(rootCmd)
	commands.RegisterSearchCommand(rootCmd)
	commands.RegisterBatchCommands(rootCmd)
	commands.RegisterSyncCommand(rootCmd)
	commands.RegisterListCommand(rootCmd)
	commands.RegisterWhoamiCommand(rootCmd)
	commands.RegisterDoctorCommand(rootCmd)
	commands.RegisterConfigCommands(rootCmd)
	commands.RegisterCacheCommands(rootCmd)
	commands.RegisterAuthCommands(rootCmd)
	commands.RegisterCommentCommands(rootCmd)
	commands.RegisterViewCommands(rootCmd)
	commands.RegisterCustomEmojiCommands(rootCmd)
	commands.RegisterMarkdownCommands(rootCmd)
	commands.RegisterFilesCommands(rootCmd)
}

// ExitCode returns the appropriate exit code for an error.
func ExitCode(err error) int {
	if err == nil {
		return clierrors.ExitSuccess
	}
	if cliErr, ok := err.(*clierrors.NotionCLIError); ok {
		return cliErr.ExitCode()
	}
	return clierrors.ExitCLIError
}
