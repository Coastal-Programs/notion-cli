package commands

import (
	"context"
	"errors"
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
		newAuthListCmd(),
		newAuthDefaultCmd(),
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
			"where the local callback server cannot be reached by your browser.",
		RunE: runAuthLogin,
	}
	cmd.Flags().Bool("manual", false, "Skip the local callback server and paste the redirected URL by hand")
	cmd.Flags().String("slug", "", "Override the local workspace credential slug")
	addOutputFlags(cmd)
	return cmd
}

func runAuthLogin(cmd *cobra.Command, args []string) error {
	start := time.Now()

	explicitSlug, _ := cmd.Flags().GetString("slug")
	manual, _ := cmd.Flags().GetBool("manual")
	data, err := performOAuthLogin(cmd, explicitSlug, manual, true)
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "auth login", start)
	return nil
}

func performOAuthLogin(cmd *cobra.Command, explicitSlug string, manual bool, printProgress bool) (map[string]any, error) {
	clientID, clientSecret, ok := config.OAuthClientCredentials()
	if !ok {
		return nil, clierrors.OAuthNotConfigured()
	}

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
		if printProgress {
			fmt.Fprintln(os.Stderr, "Opening your browser to authorize with Notion...")
			fmt.Fprintln(os.Stderr, "Waiting for authorization (timeout: 5 minutes)...")
		}
		token, err = oauth.Login(ctx, clientID, clientSecret)
	}
	if err != nil {
		return nil, err
	}

	creds, err := config.LoadCredentials()
	if err != nil {
		return nil, clierrors.Wrap(clierrors.CodeInternalError, "Failed to load credentials", err)
	}

	slug, err := config.ChooseWorkspaceSlug(creds, explicitSlug, token.WorkspaceName, token.WorkspaceID)
	if err != nil {
		return nil, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: err.Error(),
		}
	}

	meta := config.WorkspaceCredential{
		Slug:          slug,
		AuthMethod:    config.AuthMethodOAuth,
		WorkspaceID:   token.WorkspaceID,
		WorkspaceName: token.WorkspaceName,
		WorkspaceIcon: token.WorkspaceIcon,
		BotID:         token.BotID,
	}
	if token.ExpiresIn > 0 {
		meta.OAuthTokenExpiresAt = time.Now().Add(
			time.Duration(token.ExpiresIn) * time.Second).UTC().Format(time.RFC3339)
	}

	if err := config.SaveWorkspaceCredential(meta, config.WorkspaceSecrets{
		OAuthAccessToken:  token.AccessToken,
		OAuthRefreshToken: token.RefreshToken,
	}, true); err != nil {
		return nil, clierrors.Wrap(clierrors.CodeInternalError, "Failed to save workspace credential", err)
	}

	data := map[string]any{
		"auth_method":    "oauth",
		"workspace":      slug,
		"workspace_id":   token.WorkspaceID,
		"workspace_name": token.WorkspaceName,
		"bot_id":         token.BotID,
	}

	if printProgress {
		fmt.Fprintf(os.Stderr, "Authenticated with workspace %q as %q\n", token.WorkspaceName, slug)
	}

	return data, nil
}

// --- auth logout ---

func newAuthLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout [workspace]",
		Short: "Log out and clear OAuth tokens",
		Long:  "Remove stored OAuth tokens from the config file. By default the token is also revoked on the Notion API.\n\nUse --local-only to skip the API call and only clear local config.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runAuthLogout,
	}
	cmd.Flags().Bool("local-only", false, "Clear local config only; do not call the Notion revoke API")
	addOutputFlags(cmd)
	return cmd
}

func runAuthLogout(cmd *cobra.Command, args []string) error {
	start := time.Now()

	explicitWorkspace := authWorkspaceFromCommand(cmd)
	if len(args) > 0 {
		explicitWorkspace = args[0]
	}

	cfg, active, err := config.LoadConfigForWorkspace(explicitWorkspace)
	if err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to load workspace credential", err))
	}

	if !cfg.HasOAuthToken() && active.Legacy {
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
	if !localOnly && cfg.OAuthAccessToken != "" {
		ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
		defer cancel()
		if clientID, clientSecret, ok := config.OAuthClientCredentials(); ok {
			if revokeErr := oauth.TokenRevoke(ctx, clientID, clientSecret, cfg.OAuthAccessToken); revokeErr != nil {
				// Non-fatal: log a warning and continue with local clear.
				fmt.Fprintf(os.Stderr, "Warning: token revoke API call failed: %s\n", revokeErr)
			}
		}
	}

	workspaceName := cfg.OAuthWorkspaceName
	if active.Legacy {
		cfg.ClearOAuth()
		if err := config.SaveConfig(cfg); err != nil {
			return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to save config", err))
		}
	} else if _, err := config.DeleteWorkspaceCredential(active.Slug); err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to remove workspace credential", err))
	}

	data := map[string]any{
		"logged_out":     true,
		"workspace":      active.DisplayName(),
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

	cfg, active, err := loadConfigForCommand(cmd)
	if err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to load workspace credential", err))
	}

	method := cfg.AuthMethod()
	data := map[string]any{
		"auth_method": method,
		"workspace":   active.DisplayName(),
	}

	switch method {
	case "env":
		data["source"] = "NOTION_TOKEN environment variable"
		data["token"] = maskToken(cfg.Token)
		if cfg.HasOAuthToken() || (!active.Legacy && active.Metadata != nil) {
			data["note"] = "Stored workspace credential is also configured but NOTION_TOKEN takes precedence"
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
		data["source"] = "workspace credential"
		if active.Legacy {
			data["source"] = "config file"
		}
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
	if remote && cfg.HasOAuthToken() {
		ctx, cancel := context.WithTimeout(cmd.Context(), 15*time.Second)
		defer cancel()
		if clientID, clientSecret, ok := config.OAuthClientCredentials(); ok {
			if introspect, iErr := oauth.TokenIntrospect(ctx, clientID, clientSecret, cfg.OAuthAccessToken); iErr == nil {
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
		} else {
			data["remote_error"] = "OAuth client credentials are not configured"
		}
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "auth status", start)
	return nil
}

// --- auth list ---

func newAuthListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stored workspace credentials",
		Long:  "List Notion workspace credentials stored by this CLI. Secrets are never printed.",
		Args:  cobra.NoArgs,
		RunE:  runAuthList,
	}
	addOutputFlags(cmd)
	return cmd
}

func runAuthList(cmd *cobra.Command, args []string) error {
	start := time.Now()

	creds, err := config.LoadCredentials()
	if err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to load credentials", err))
	}

	rows := []map[string]any{}
	for _, slug := range creds.SortedWorkspaceSlugs() {
		meta := creds.Workspaces[slug]
		rows = append(rows, map[string]any{
			"default":        slug == creds.DefaultWorkspace,
			"workspace":      slug,
			"workspace_name": meta.WorkspaceName,
			"workspace_id":   meta.WorkspaceID,
			"auth_method":    meta.AuthMethod,
			"bot_id":         meta.BotID,
			"expires_at":     meta.OAuthTokenExpiresAt,
		})
	}

	p := output.NewPrinter(outputFormat(cmd))
	if len(rows) == 0 {
		p.PrintSuccess(map[string]any{
			"workspaces": []map[string]any{},
			"message":    "No workspaces configured. Run 'notion-cli auth login' to add one.",
		}, "auth list", start)
		return nil
	}
	p.PrintSuccess(rows, "auth list", start)
	return nil
}

// --- auth default ---

func newAuthDefaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "default [workspace]",
		Short: "Show or set the default workspace credential",
		Long:  "Show or set the default Notion workspace credential. Use 'default' to clear the named default and fall back to legacy config.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runAuthDefault,
	}
	addOutputFlags(cmd)
	return cmd
}

func runAuthDefault(cmd *cobra.Command, args []string) error {
	start := time.Now()

	if len(args) == 0 {
		creds, err := config.LoadCredentials()
		if err != nil {
			return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to load credentials", err))
		}
		current := creds.DefaultWorkspace
		if current == "" {
			current = config.LegacyWorkspaceSlug
		}
		if shouldRunDefaultWorkspaceSelector(cmd, creds) {
			selected, err := selectDefaultWorkspace(cmd, creds)
			if err != nil {
				if errors.Is(err, errWorkspaceSelectionCancelled) {
					return handleError(cmd, &clierrors.NotionCLIError{
						Code:    clierrors.CodeInvalidRequest,
						Message: "Default workspace selection cancelled.",
					})
				}
				return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to select default workspace", err))
			}
			if err := config.SetDefaultWorkspace(selected); err != nil {
				return handleError(cmd, &clierrors.NotionCLIError{
					Code:    clierrors.CodeInvalidRequest,
					Message: err.Error(),
				})
			}
			p := output.NewPrinter(outputFormat(cmd))
			p.PrintSuccess(map[string]any{"workspace": selected, "default_set": true}, "auth default", start)
			return nil
		}
		p := output.NewPrinter(outputFormat(cmd))
		p.PrintSuccess(map[string]any{"workspace": current}, "auth default", start)
		return nil
	}

	if err := config.SetDefaultWorkspace(args[0]); err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: err.Error(),
		})
	}
	p := output.NewPrinter(outputFormat(cmd))
	if args[0] == config.LegacyWorkspaceSlug {
		p.PrintSuccess(map[string]any{
			"workspace":   config.LegacyWorkspaceSlug,
			"default_set": false,
			"message":     "Cleared named default; legacy config is now active.",
		}, "auth default", start)
		return nil
	}
	p.PrintSuccess(map[string]any{"workspace": args[0], "default_set": true}, "auth default", start)
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

	cfg, active, err := loadConfigForCommand(cmd)
	if err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to load workspace credential", err))
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

	clientID, clientSecret, ok := config.OAuthClientCredentials()
	if !ok {
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

	if err := config.SaveConfigForWorkspace(active, cfg); err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to save workspace credential", err))
	}

	data := map[string]any{
		"auth_method":    "oauth",
		"workspace":      active.DisplayName(),
		"workspace_id":   config.FirstNonEmpty(token.WorkspaceID, cfg.OAuthWorkspaceID),
		"workspace_name": config.FirstNonEmpty(token.WorkspaceName, cfg.OAuthWorkspaceName),
	}
	if cfg.OAuthTokenExpiresAt != "" {
		data["expires_at"] = cfg.OAuthTokenExpiresAt
	}

	fmt.Fprintln(os.Stderr, "OAuth token refreshed successfully.")

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "auth refresh", start)
	return nil
}
