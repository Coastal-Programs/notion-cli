package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Coastal-Programs/notion-cli/internal/config"
	clierrors "github.com/Coastal-Programs/notion-cli/internal/errors"
	"github.com/Coastal-Programs/notion-cli/internal/oauth"
	"github.com/Coastal-Programs/notion-cli/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterAuthCommands registers all auth subcommands under root.
func RegisterAuthCommands(root *cobra.Command) {
	authCmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication commands",
		Long:  "Log in, log out, and check authentication status via OAuth or manual token.",
	}

	authCmd.AddCommand(
		newAuthLoginCmd(),
		newAuthLogoutCmd(),
		newAuthStatusCmd(),
	)

	root.AddCommand(authCmd)
}

// --- auth login ---

func newAuthLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in via Notion OAuth",
		Long:  "Authenticate with Notion by opening your browser and completing the OAuth flow.",
		RunE:  runAuthLogin,
	}
	addOutputFlags(cmd)
	return cmd
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	start := time.Now()

	clientID := config.OAuthClientID
	clientSecret := config.OAuthClientSecret

	if clientID == "" || clientSecret == "" {
		return handleError(cmd, clierrors.OAuthNotConfigured())
	}

	fmt.Fprintln(os.Stderr, "Opening your browser to authorize with Notion...")
	fmt.Fprintln(os.Stderr, "Waiting for authorization (timeout: 2 minutes)...")

	// 2-minute timeout for the full OAuth flow.
	ctx, cancel := context.WithTimeout(cmd.Context(), 2*time.Minute)
	defer cancel()

	token, err := oauth.Login(ctx, clientID, clientSecret)
	if err != nil {
		return handleError(cmd, err)
	}

	// Save token to config.
	cfg, err := config.LoadConfig()
	if err != nil {
		cfg = &config.Config{}
	}

	cfg.OAuthAccessToken = token.AccessToken
	cfg.OAuthWorkspaceID = token.WorkspaceID
	cfg.OAuthWorkspaceName = token.WorkspaceName
	cfg.OAuthBotID = token.BotID

	if err := config.SaveConfig(cfg); err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: fmt.Sprintf("Failed to save config: %s", err),
		})
	}

	data := map[string]any{
		"auth_method":    "oauth",
		"workspace_id":   token.WorkspaceID,
		"workspace_name": token.WorkspaceName,
		"bot_id":         token.BotID,
	}

	fmt.Fprintf(os.Stderr, "Authenticated with workspace: %s\n", token.WorkspaceName)

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "auth login", start)
	return nil
}

// --- auth logout ---

func newAuthLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out and clear OAuth tokens",
		Long:  "Remove stored OAuth tokens from the config file.",
		RunE:  runAuthLogout,
	}
	addOutputFlags(cmd)
	return cmd
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	start := time.Now()

	cfg, err := config.LoadConfig()
	if err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: fmt.Sprintf("Failed to load config: %s", err),
		})
	}

	if !cfg.HasOAuthToken() {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: "No OAuth session to log out from",
			Suggestions: []string{
				"You are not currently logged in via OAuth",
				"If using NOTION_TOKEN env var, unset it instead",
			},
		})
	}

	workspaceName := cfg.OAuthWorkspaceName
	cfg.ClearOAuth()

	if err := config.SaveConfig(cfg); err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: fmt.Sprintf("Failed to save config: %s", err),
		})
	}

	data := map[string]any{
		"logged_out":     true,
		"workspace_name": workspaceName,
	}

	fmt.Fprintf(os.Stderr, "Logged out from workspace: %s\n", workspaceName)

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "auth logout", start)
	return nil
}

// --- auth status ---

func newAuthStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  "Display the current authentication method and details.",
		RunE:  runAuthStatus,
	}
	addOutputFlags(cmd)
	return cmd
}

func runAuthStatus(cmd *cobra.Command, args []string) error {
	start := time.Now()

	cfg, err := config.LoadConfig()
	if err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: fmt.Sprintf("Failed to load config: %s", err),
		})
	}

	method := cfg.AuthMethod()
	data := map[string]any{
		"auth_method": method,
	}

	switch method {
	case "env":
		token := os.Getenv("NOTION_TOKEN")
		data["source"] = "NOTION_TOKEN environment variable"
		data["token"] = maskToken(token)
		if cfg.HasOAuthToken() {
			data["note"] = "OAuth token is also configured but env var takes precedence"
		}
	case "oauth":
		data["workspace_id"] = cfg.OAuthWorkspaceID
		data["workspace_name"] = cfg.OAuthWorkspaceName
		data["bot_id"] = cfg.OAuthBotID
		data["token"] = maskToken(cfg.OAuthAccessToken)
	case "token":
		data["source"] = "config file"
		data["token"] = maskToken(cfg.Token)
	case "none":
		data["message"] = "Not authenticated"
		data["suggestions"] = []string{
			"Run 'notion-cli auth login' to authenticate via OAuth",
			"Or set NOTION_TOKEN environment variable",
		}
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "auth status", start)
	return nil
}

