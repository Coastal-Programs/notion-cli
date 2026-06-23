package commands

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Coastal-Programs/notion-cli/v6/internal/config"
	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
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
	cmd := &cobra.Command{
		Use:     "set-token [token]",
		Aliases: []string{"token"},
		Short:   "Set the Notion API token",
		Long: `Save the Notion API token.

If no argument is provided, the token is read from stdin.
You can pipe a token via stdin to avoid exposing it in process listings:

  echo "$NOTION_TOKEN" | notion-cli config set-token
  notion-cli config set-token < token-file.txt`,
		Args: cobra.MaximumNArgs(1),
		RunE: runConfigSetToken,
	}
	addOutputFlags(cmd)
	return cmd
}

func runConfigSetToken(cmd *cobra.Command, args []string) error {
	start := time.Now()

	var token string
	if len(args) == 1 {
		token = args[0]
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "Warning: passing tokens as arguments exposes them in process listings. Consider piping via stdin instead.")
	} else {
		// Read token from stdin.
		// If stdin is a terminal, prompt the user.
		info, err := os.Stdin.Stat()
		if err == nil && (info.Mode()&os.ModeCharDevice) != 0 {
			_, _ = fmt.Fprint(cmd.ErrOrStderr(), "Enter token: ")
		}
		scanner := bufio.NewScanner(os.Stdin)
		if scanner.Scan() {
			token = strings.TrimSpace(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to read stdin", err))
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
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to load config", err))
	}

	cfg.Token = token
	if err := config.SaveConfig(cfg); err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to save config", err))
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(map[string]any{
		"message":     "Token saved successfully",
		"workspace":   config.LegacyWorkspaceSlug,
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
	cfg, _, err := loadConfigForCommand(cmd)
	if err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to load config", err))
	}
	value := configValue(cfg, args[0])

	// Mask sensitive values unless --show-secret is set.
	sensitiveKeys := map[string]bool{
		"token":               true,
		"oauth_access_token":  true,
		"oauth_refresh_token": true,
	}
	if sensitiveKeys[args[0]] {
		showSecret, _ := cmd.Flags().GetBool("show-secret")
		if !showSecret {
			value = maskToken(value)
		}
	}

	_, _ = fmt.Fprintln(cmd.OutOrStdout(), value)
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
	_, active, err := loadConfigForCommand(cmd)
	if err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to load config", err))
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), config.GetConfigPathForWorkspace(active.Slug))
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

	cfg, active, err := loadConfigForCommand(cmd)
	if err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to load config", err))
	}

	// Mask token.
	token := maskToken(cfg.Token)

	data := map[string]any{
		"token":               token,
		"oauth_access_token":  maskToken(cfg.OAuthAccessToken),
		"oauth_refresh_token": maskToken(cfg.OAuthRefreshToken),
		"base_url":            cfg.BaseURL,
		"max_retries":         cfg.MaxRetries,
		"base_delay_ms":       cfg.BaseDelayMs,
		"max_delay_ms":        cfg.MaxDelayMs,
		"cache_enabled":       cfg.CacheEnabled,
		"cache_max_size":      cfg.CacheMaxSize,
		"disk_cache_enabled":  cfg.DiskCacheEnabled,
		"http_keep_alive":     cfg.HTTPKeepAlive,
		"verbose":             cfg.Verbose,
		"workspace":           active.DisplayName(),
		"config_path":         config.GetConfigPathForWorkspace(active.Slug),
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "config list", start)
	return nil
}

func configValue(cfg *config.Config, key string) string {
	switch key {
	case "token":
		return cfg.Token
	case "base_url":
		return cfg.BaseURL
	case "max_retries":
		return fmt.Sprintf("%d", cfg.MaxRetries)
	case "base_delay_ms":
		return fmt.Sprintf("%d", cfg.BaseDelayMs)
	case "max_delay_ms":
		return fmt.Sprintf("%d", cfg.MaxDelayMs)
	case "cache_enabled":
		return fmt.Sprintf("%t", cfg.CacheEnabled)
	case "cache_max_size":
		return fmt.Sprintf("%d", cfg.CacheMaxSize)
	case "disk_cache_enabled":
		return fmt.Sprintf("%t", cfg.DiskCacheEnabled)
	case "http_keep_alive":
		return fmt.Sprintf("%t", cfg.HTTPKeepAlive)
	case "verbose":
		return fmt.Sprintf("%t", cfg.Verbose)
	case "oauth_access_token":
		return cfg.OAuthAccessToken
	case "oauth_workspace_id":
		return cfg.OAuthWorkspaceID
	case "oauth_workspace_name":
		return cfg.OAuthWorkspaceName
	case "oauth_bot_id":
		return cfg.OAuthBotID
	case "oauth_refresh_token":
		return cfg.OAuthRefreshToken
	case "oauth_token_expires_at":
		return cfg.OAuthTokenExpiresAt
	case "auth_method":
		return cfg.AuthMethod()
	default:
		return ""
	}
}
