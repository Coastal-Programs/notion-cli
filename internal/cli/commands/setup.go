package commands

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/Coastal-Programs/notion-cli/v6/internal/config"
	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/spf13/cobra"
)

func shouldRunInitialSetup(workspaceSlug string, err error) bool {
	var cliErr *clierrors.NotionCLIError
	if !errors.As(err, &cliErr) || cliErr.Code != clierrors.CodeTokenMissing {
		return false
	}
	if strings.TrimSpace(workspaceSlug) != "" || strings.TrimSpace(os.Getenv("NOTION_WORKSPACE")) != "" {
		return false
	}
	if !isTerminal() {
		return false
	}
	if _, _, ok := config.OAuthClientCredentials(); !ok {
		return false
	}
	creds, err := config.LoadCredentials()
	if err != nil || len(creds.Workspaces) > 0 {
		return false
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		return false
	}
	return cfg.Token == "" && cfg.OAuthAccessToken == "" && cfg.OAuthRefreshToken == ""
}

func runInitialWorkspaceSetup(cmd *cobra.Command) error {
	fmt.Fprintln(cmd.ErrOrStderr(), "No Notion workspace is configured. Starting first-time setup...")
	data, err := performOAuthLogin(cmd, "", false, true)
	if err != nil {
		return err
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Default Notion workspace set to %q.\n", data["workspace"])
	return nil
}
