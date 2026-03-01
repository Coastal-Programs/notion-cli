package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Coastal-Programs/notion-cli/internal/cli/commands"
	"github.com/Coastal-Programs/notion-cli/internal/config"
	clierrors "github.com/Coastal-Programs/notion-cli/internal/errors"
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
	pf.Int("timeout", 30000, "Request timeout in milliseconds")

	// Version template.
	rootCmd.SetVersionTemplate(fmt.Sprintf("notion-cli version %s (commit: %s, built: %s)\n",
		config.Version, config.Commit, config.Date))

	// Register all command groups.
	commands.RegisterDBCommands(rootCmd)
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
