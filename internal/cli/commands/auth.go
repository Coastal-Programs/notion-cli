package commands

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/Coastal-Programs/notion-cli/v6/internal/config"
	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/internal/oauth"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
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
		newAuthRefreshCmd(),
	)

	root.AddCommand(authCmd)
}

// --- auth login ---

func newAuthLoginCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in via Notion OAuth",
		Long: "Authenticate with Notion by opening your browser and completing the OAuth flow.\n\n" +
			"Use --manual when running over SSH, in a container, or behind a firewall " +
			"where the local callback server cannot be reached by your browser. In manual " +
			"mode you must paste the FULL redirected URL (the authorization code alone is " +
			"not sufficient \u2014 the state parameter must be present for CSRF protection).",
		RunE: runAuthLogin,
	}
	cmd.Flags().Bool("manual", false, "Skip the local callback server and paste the FULL redirected URL by hand")
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

	manual, _ := cmd.Flags().GetBool("manual")
	// Auto-enable manual mode when running over SSH; the user's browser
	// almost certainly cannot reach our local callback server.
	if !manual && os.Getenv("SSH_TTY") != "" {
		fmt.Fprintln(os.Stderr, "SSH session detected \u2014 falling back to manual flow.")
		manual = true
	}

	// 5-minute timeout for the full OAuth flow. Generous to accommodate 2FA,
	// account switching, workspace selection, and slow networks.
	ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Minute)
	defer cancel()

	var (
		token *oauth.TokenResponse
		err   error
	)
	if manual {
		token, err = oauth.LoginManual(ctx, clientID, clientSecret, os.Stdin, os.Stderr)
	} else {
		fmt.Fprintln(os.Stderr, "Opening your browser to authorize with Notion...")
		fmt.Fprintln(os.Stderr, "Waiting for authorization (timeout: 5 minutes)...")
		token, err = oauth.Login(ctx, clientID, clientSecret)
	}
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
	cfg.OAuthRefreshToken = token.RefreshToken
	if token.ExpiresIn > 0 {
		cfg.OAuthTokenExpiresAt = time.Now().Add(
			time.Duration(token.ExpiresIn) * time.Second).UTC().Format(time.RFC3339)
	}

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
		Long:  "Remove stored OAuth tokens from the config file. By default the token is also revoked on the Notion API.\n\nUse --local-only to skip the API call and only clear local config.",
		RunE:  runAuthLogout,
	}
	cmd.Flags().Bool("local-only", false, "Clear local config only; do not call the Notion revoke API")
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

	localOnly, _ := cmd.Flags().GetBool("local-only")

	// Revoke the token on the API unless --local-only was requested.
	if !localOnly && config.OAuthClientID != "" {
		ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
		defer cancel()
		if revokeErr := oauth.TokenRevoke(ctx, config.OAuthClientID, config.OAuthClientSecret, cfg.OAuthAccessToken); revokeErr != nil {
			// Non-fatal: log a warning and continue with local clear.
			fmt.Fprintf(os.Stderr, "Warning: token revoke API call failed: %s\n", revokeErr)
		}
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
		Long:  "Display the current authentication method and details.\n\nUse --remote to also call the Notion token introspect API and show active status, scope, and issued_at.",
		RunE:  runAuthStatus,
	}
	cmd.Flags().Bool("remote", false, "Also call Notion token introspect API to verify active status")
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
		if cfg.OAuthTokenExpiresAt != "" {
			data["expires_at"] = cfg.OAuthTokenExpiresAt
		}
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

	// --remote: introspect the token via the Notion API.
	remote, _ := cmd.Flags().GetBool("remote")
	if remote && cfg.HasOAuthToken() && config.OAuthClientID != "" {
		ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
		defer cancel()
		if introspect, iErr := oauth.TokenIntrospect(ctx, config.OAuthClientID, config.OAuthClientSecret, cfg.OAuthAccessToken); iErr == nil {
			data["active"] = introspect.Active
			if introspect.Scope != "" {
				data["scope"] = introspect.Scope
			}
			if introspect.IssuedAt != 0 {
				data["issued_at"] = time.Unix(introspect.IssuedAt, 0).UTC().Format(time.RFC3339)
			}
		} else {
			data["remote_error"] = iErr.Error()
		}
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "auth status", start)
	return nil
}

// --- auth refresh ---

func newAuthRefreshCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "refresh",
		Short: "Refresh the OAuth access token",
		Long:  "Exchange the stored refresh token for a new OAuth access token and persist it to config.",
		RunE:  runAuthRefresh,
	}
	addOutputFlags(cmd)
	return cmd
}

func runAuthRefresh(cmd *cobra.Command, args []string) error {
	start := time.Now()

	cfg, err := config.LoadConfig()
	if err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: fmt.Sprintf("Failed to load config: %s", err),
		})
	}

	if cfg.OAuthRefreshToken == "" {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: "No refresh token stored in config",
			Suggestions: []string{
				"Run 'notion-cli auth login' to authenticate via OAuth first",
			},
		})
	}

	clientID := config.OAuthClientID
	clientSecret := config.OAuthClientSecret
	if clientID == "" || clientSecret == "" {
		return handleError(cmd, clierrors.OAuthNotConfigured())
	}

	ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
	defer cancel()

	token, err := oauth.TokenRefresh(ctx, clientID, clientSecret, cfg.OAuthRefreshToken)
	if err != nil {
		return handleError(cmd, err)
	}

	cfg.OAuthAccessToken = token.AccessToken
	if token.RefreshToken != "" {
		cfg.OAuthRefreshToken = token.RefreshToken
	}
	if token.ExpiresIn > 0 {
		cfg.OAuthTokenExpiresAt = time.Now().Add(
			time.Duration(token.ExpiresIn) * time.Second).UTC().Format(time.RFC3339)
	}

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
	}
	if cfg.OAuthTokenExpiresAt != "" {
		data["expires_at"] = cfg.OAuthTokenExpiresAt
	}

	fmt.Fprintln(os.Stderr, "OAuth token refreshed successfully.")

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "auth refresh", start)
	return nil
}
