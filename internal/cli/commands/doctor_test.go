package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Coastal-Programs/notion-cli/v6/internal/config"
	"github.com/spf13/cobra"
)

// TestOAuthCredentialsEmbedded_DefaultBuild ensures the helper returns false
// when ldflags are not set (the case for `go test` builds). This guards
// against accidental hard-coded defaults in `internal/config`.
func runDoctorRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterDoctorCommand(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

func TestDoctorCmd_NoToken(t *testing.T) {
	// With no token configured the doctor still runs but reports failures.
	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Unsetenv("NOTION_TOKEN")
	t.Cleanup(func() {
		if origToken != "" {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})

	// Point HOME to empty temp dir so no config file is found.
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	// Doctor command returns error when checks fail, but should not panic.
	_, _, _ = runDoctorRoot(t, "doctor")
}

func TestDoctorCmd_WithToken(t *testing.T) {
	// Provide a valid-looking token via env + mock API server.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"object":"user","id":"bot-1","type":"bot"}`))
	}))
	defer srv.Close()

	origToken := os.Getenv("NOTION_TOKEN")
	origBase := os.Getenv("NOTION_CLI_BASE_URL")
	_ = os.Setenv("NOTION_TOKEN", "secret_test_token")
	_ = os.Setenv("NOTION_CLI_BASE_URL", srv.URL)
	t.Cleanup(func() {
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
		if origBase == "" {
			_ = os.Unsetenv("NOTION_CLI_BASE_URL")
		} else {
			_ = os.Setenv("NOTION_CLI_BASE_URL", origBase)
		}
	})

	// Doctor may pass or fail depending on network connectivity; just ensure no panic.
	_, _, _ = runDoctorRoot(t, "doctor")
}

func TestOAuthCredentialsAvailable_DefaultBuild(t *testing.T) {
	origID, origSecret := config.OAuthClientID, config.OAuthClientSecret
	t.Cleanup(func() {
		config.OAuthClientID = origID
		config.OAuthClientSecret = origSecret
	})
	t.Setenv("NOTION_OAUTH_CLIENT_ID", "")
	t.Setenv("NOTION_OAUTH_SECRET", "")

	config.OAuthClientID = ""
	config.OAuthClientSecret = ""
	if oauthCredentialsAvailable() {
		t.Error("expected false when both vars empty")
	}

	config.OAuthClientID = "id"
	if oauthCredentialsAvailable() {
		t.Error("expected false when only client id set")
	}

	config.OAuthClientID = ""
	config.OAuthClientSecret = "secret"
	if oauthCredentialsAvailable() {
		t.Error("expected false when only client secret set")
	}

	config.OAuthClientID = "id"
	config.OAuthClientSecret = "secret"
	if !oauthCredentialsAvailable() {
		t.Error("expected true when both set")
	}
}
