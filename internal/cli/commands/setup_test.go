package commands

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Coastal-Programs/notion-cli/v6/internal/config"
	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
)

func withInitialSetupTestEnv(t *testing.T) {
	t.Helper()
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })
	t.Setenv("NOTION_TOKEN", "")
	t.Setenv("NOTION_WORKSPACE", "")

	origID, origSecret := config.OAuthClientID, config.OAuthClientSecret
	config.OAuthClientID = "client-id"
	config.OAuthClientSecret = "client-secret"
	t.Cleanup(func() {
		config.OAuthClientID = origID
		config.OAuthClientSecret = origSecret
	})

	origIsTerminal := isTerminal
	isTerminal = func() bool { return true }
	t.Cleanup(func() { isTerminal = origIsTerminal })
}

func TestShouldRunInitialSetup(t *testing.T) {
	withInitialSetupTestEnv(t)

	if !shouldRunInitialSetup("", clierrors.TokenMissing()) {
		t.Fatal("shouldRunInitialSetup() = false, want true")
	}
}

func TestShouldRunInitialSetup_FalseCases(t *testing.T) {
	t.Run("non token missing error", func(t *testing.T) {
		withInitialSetupTestEnv(t)
		if shouldRunInitialSetup("", errors.New("boom")) {
			t.Fatal("shouldRunInitialSetup() = true for generic error")
		}
	})

	t.Run("explicit workspace", func(t *testing.T) {
		withInitialSetupTestEnv(t)
		if shouldRunInitialSetup("haven", clierrors.TokenMissing()) {
			t.Fatal("shouldRunInitialSetup() = true with explicit workspace")
		}
	})

	t.Run("workspace env", func(t *testing.T) {
		withInitialSetupTestEnv(t)
		t.Setenv("NOTION_WORKSPACE", "haven")
		if shouldRunInitialSetup("", clierrors.TokenMissing()) {
			t.Fatal("shouldRunInitialSetup() = true with NOTION_WORKSPACE")
		}
	})

	t.Run("non interactive", func(t *testing.T) {
		withInitialSetupTestEnv(t)
		isTerminal = func() bool { return false }
		if shouldRunInitialSetup("", clierrors.TokenMissing()) {
			t.Fatal("shouldRunInitialSetup() = true when non-interactive")
		}
	})

	t.Run("oauth unavailable", func(t *testing.T) {
		withInitialSetupTestEnv(t)
		config.OAuthClientID = ""
		config.OAuthClientSecret = ""
		if shouldRunInitialSetup("", clierrors.TokenMissing()) {
			t.Fatal("shouldRunInitialSetup() = true without OAuth credentials")
		}
	})

	t.Run("named credentials exist", func(t *testing.T) {
		withInitialSetupTestEnv(t)
		if err := config.SaveCredentials(&config.CredentialsFile{
			Workspaces: map[string]config.WorkspaceCredential{
				"haven": {Slug: "haven", AuthMethod: config.AuthMethodOAuth},
			},
		}); err != nil {
			t.Fatalf("SaveCredentials: %v", err)
		}
		if shouldRunInitialSetup("", clierrors.TokenMissing()) {
			t.Fatal("shouldRunInitialSetup() = true with existing workspace credentials")
		}
	})

	t.Run("legacy token exists", func(t *testing.T) {
		withInitialSetupTestEnv(t)
		cfgDir := filepath.Join(os.Getenv("HOME"), ".config", "notion-cli")
		if err := os.MkdirAll(cfgDir, 0o700); err != nil {
			t.Fatalf("MkdirAll: %v", err)
		}
		if err := config.SaveConfig(&config.Config{Token: "secret_legacy"}); err != nil {
			t.Fatalf("SaveConfig: %v", err)
		}
		if shouldRunInitialSetup("", clierrors.TokenMissing()) {
			t.Fatal("shouldRunInitialSetup() = true with legacy token")
		}
	})
}

func TestShouldRunInitialSetup_DevNullStdin(t *testing.T) {
	realIsTerminal := isTerminal
	withInitialSetupTestEnv(t)
	isTerminal = realIsTerminal

	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("open %s: %v", os.DevNull, err)
	}
	t.Cleanup(func() { _ = devNull.Close() })
	origStdin := os.Stdin
	os.Stdin = devNull
	t.Cleanup(func() { os.Stdin = origStdin })

	// Cron, CI and agent harnesses run commands with stdin from /dev/null (#132).
	if shouldRunInitialSetup("", clierrors.TokenMissing()) {
		t.Fatal("shouldRunInitialSetup() = true with stdin from /dev/null, want false")
	}
}
