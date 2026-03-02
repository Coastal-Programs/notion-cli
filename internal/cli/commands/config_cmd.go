package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Coastal-Programs/notion-cli/internal/config"
	clierrors "github.com/Coastal-Programs/notion-cli/internal/errors"
	"github.com/Coastal-Programs/notion-cli/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterConfigCommands registers config subcommands.
func RegisterConfigCommands(root *cobra.Command) {
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Configuration management",
		Long:  "View and manage CLI configuration.",
	}

	configCmd.AddCommand(
		newConfigSetTokenCmd(),
		newConfigGetCmd(),
		newConfigPathCmd(),
		newConfigListCmd(),
	)

	root.AddCommand(configCmd)
}

// maskToken masks a token for safe display.
// If token is empty, returns "(not set)".
// If len < 10, returns "***".
// Otherwise returns first 7 chars + "***..." + last 3 chars.
func maskToken(token string) string {
	if token == "" {
		return "(not set)"
	}
	if len(token) < 10 {
		return "***"
	}
	return token[:7] + "***..." + token[len(token)-3:]
}

// --- config set-token ---

func newConfigSetTokenCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "set-token [token]",
		Aliases: []string{"token"},
		Short:   "Set the Notion API token",
		Long: `Save the Notion API token to the config file.

If no argument is provided, the token is read from stdin.
You can pipe a token via stdin to avoid exposing it in process listings:

  echo "$NOTION_TOKEN" | notion-cli config set-token
  notion-cli config set-token < token-file.txt`,
		Args: cobra.MaximumNArgs(1),
		RunE: runConfigSetToken,
	}
}

func runConfigSetToken(cmd *cobra.Command, args []string) error {
	start := time.Now()

	var token string
	if len(args) == 1 {
		token = args[0]
		fmt.Fprintln(cmd.ErrOrStderr(), "Warning: passing tokens as arguments exposes them in process listings. Consider piping via stdin instead.")
	} else {
		// Read token from stdin.
		// If stdin is a terminal, prompt the user.
		info, err := os.Stdin.Stat()
		if err == nil && (info.Mode()&os.ModeCharDevice) != 0 {
			fmt.Fprint(cmd.ErrOrStderr(), "Enter token: ")
		}
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			token = strings.TrimSpace(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return handleError(cmd, &clierrors.NotionCLIError{
				Code:    clierrors.CodeInternalError,
				Message: fmt.Sprintf("Failed to read stdin: %s", err),
			})
		}
		if token == "" {
			return handleError(cmd, &clierrors.NotionCLIError{
				Code:    clierrors.CodeMissingRequired,
				Message: "No token provided",
				Suggestions: []string{
					"Pass token as argument: notion-cli config set-token <token>",
					"Or pipe via stdin: echo $NOTION_TOKEN | notion-cli config set-token",
				},
			})
		}
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: fmt.Sprintf("Failed to load config: %s", err),
		})
	}

	cfg.Token = token
	if err := config.SaveConfig(cfg); err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: fmt.Sprintf("Failed to save config: %s", err),
		})
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(map[string]any{
		"message":     "Token saved successfully",
		"config_path": config.GetConfigPath(),
	}, "config set-token", start)
	return nil
}

// --- config get ---

func newConfigGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <key>",
		Short: "Get a config value",
		Long:  "Get a configuration value by key.",
		Args:  cobra.ExactArgs(1),
		RunE:  runConfigGet,
	}
	cmd.Flags().Bool("show-secret", false, "Show unmasked token value")
	return cmd
}

func runConfigGet(cmd *cobra.Command, args []string) error {
	value := config.GetConfigValue(args[0])

	// Mask token unless --show-secret is set.
	if args[0] == "token" {
		showSecret, _ := cmd.Flags().GetBool("show-secret")
		if !showSecret {
			value = maskToken(value)
		}
	}

	fmt.Fprintln(cmd.OutOrStdout(), value)
	return nil
}

// --- config path ---

func newConfigPathCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "path",
		Short: "Show config file path",
		RunE:  runConfigPath,
	}
}

func runConfigPath(cmd *cobra.Command, args []string) error {
	fmt.Fprintln(cmd.OutOrStdout(), config.GetConfigPath())
	return nil
}

// --- config list ---

func newConfigListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List all config values",
		RunE:    runConfigList,
	}
	addOutputFlags(cmd)
	return cmd
}

func runConfigList(cmd *cobra.Command, args []string) error {
	start := time.Now()

	cfg, err := config.LoadConfig()
	if err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: fmt.Sprintf("Failed to load config: %s", err),
		})
	}

	// Mask token.
	token := maskToken(cfg.Token)

	data := map[string]any{
		"token":              token,
		"base_url":           cfg.BaseURL,
		"max_retries":        cfg.MaxRetries,
		"base_delay_ms":      cfg.BaseDelayMs,
		"max_delay_ms":       cfg.MaxDelayMs,
		"cache_enabled":      cfg.CacheEnabled,
		"cache_max_size":     cfg.CacheMaxSize,
		"disk_cache_enabled": cfg.DiskCacheEnabled,
		"http_keep_alive":    cfg.HTTPKeepAlive,
		"verbose":            cfg.Verbose,
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "config list", start)
	return nil
}
