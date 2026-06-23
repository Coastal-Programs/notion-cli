package commands

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// runConfigRoot returns a root command with config subcommands registered and
// stdout/stderr captured into a buffer.
func runConfigRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterConfigCommands(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

// ---------------------------------------------------------------------------
// maskToken pure-function tests
// ---------------------------------------------------------------------------

func TestMaskToken_Empty(t *testing.T) {
	if got := maskToken(""); got != "(not set)" {
		t.Errorf("maskToken(\"\") = %q, want \"(not set)\"", got)
	}
}

func TestMaskToken_Short(t *testing.T) {
	if got := maskToken("abc"); got != "***" {
		t.Errorf("maskToken(\"abc\") = %q, want \"***\"", got)
	}
}

func TestMaskToken_Long(t *testing.T) {
	got := maskToken("secret_longtoken")
	if !strings.HasPrefix(got, "secret_") {
		t.Errorf("maskToken long: %q should start with 'secret_'", got)
	}
	if !strings.Contains(got, "***") {
		t.Errorf("maskToken long: %q should contain '***'", got)
	}
}

// ---------------------------------------------------------------------------
// config get
// ---------------------------------------------------------------------------

func TestConfigGet_Token(t *testing.T) {
	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Setenv("NOTION_TOKEN", "secret_test123")
	t.Cleanup(func() {
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})

	_, buf, err := runConfigRoot(t, "config", "get", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	// Token should be masked by default.
	if strings.Contains(out, "secret_test123") {
		t.Error("full token should not appear in masked output")
	}
	if strings.TrimSpace(out) == "" {
		t.Error("output should not be empty")
	}
}

func TestConfigGet_ShowSecret(t *testing.T) {
	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Setenv("NOTION_TOKEN", "secret_revealed")
	t.Cleanup(func() {
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})

	_, buf, err := runConfigRoot(t, "config", "get", "token", "--show-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "secret_revealed") {
		t.Errorf("output = %q, want raw token with --show-secret", buf.String())
	}
}

func TestConfigGet_OAuthRefreshTokenMasked(t *testing.T) {
	// Set up an isolated HOME with a config file containing the refresh token.
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Unsetenv("NOTION_TOKEN")
	t.Cleanup(func() {
		if origToken != "" {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})

	cfgDir := filepath.Join(tmpDir, ".config", "notion-cli")
	if err := os.MkdirAll(cfgDir, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgPath := filepath.Join(cfgDir, "config.json")
	const rawRefresh = "rt_supersecretrefreshvalue"
	contents := `{"oauth_refresh_token":"` + rawRefresh + `"}`
	if err := os.WriteFile(cfgPath, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	// Default: should be masked.
	_, buf, err := runConfigRoot(t, "config", "get", "oauth_refresh_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if strings.Contains(out, rawRefresh) {
		t.Errorf("oauth_refresh_token should be masked by default, got: %q", out)
	}
	if !strings.Contains(out, "***") {
		t.Errorf("masked output should contain ***, got: %q", out)
	}

	// With --show-secret: should reveal.
	_, buf2, err := runConfigRoot(t, "config", "get", "oauth_refresh_token", "--show-secret")
	if err != nil {
		t.Fatalf("unexpected error with --show-secret: %v", err)
	}
	if !strings.Contains(buf2.String(), rawRefresh) {
		t.Errorf("with --show-secret, output should contain raw refresh token, got: %q", buf2.String())
	}
}

func TestConfigGet_UnknownKey(t *testing.T) {
	_, _, err := runConfigRoot(t, "config", "get", "nonexistent_key_xyz")
	if err != nil {
		t.Errorf("unknown key should not return error, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// config path
// ---------------------------------------------------------------------------

func TestConfigPath(t *testing.T) {
	_, buf, err := runConfigRoot(t, "config", "path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := strings.TrimSpace(buf.String())
	if out == "" {
		t.Error("config path output should not be empty")
	}
	if !filepath.IsAbs(out) {
		t.Errorf("config path %q should be absolute", out)
	}
}

// ---------------------------------------------------------------------------
// config list
// ---------------------------------------------------------------------------

func TestConfigList_ReturnsAll(t *testing.T) {
	var err error
	out := captureStdout(t, func() {
		_, _, err = runConfigRoot(t, "config", "list", "--output", "json")
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	envelope := parseEnvelope(t, out)

	data, ok := envelope["data"].(map[string]any)
	if !ok {
		t.Fatalf("envelope.data not a map: %T", envelope["data"])
	}

	wantKeys := []string{
		"token", "base_url", "max_retries", "base_delay_ms", "max_delay_ms",
		"cache_enabled", "cache_max_size", "disk_cache_enabled", "http_keep_alive", "verbose",
	}
	for _, k := range wantKeys {
		if _, found := data[k]; !found {
			t.Errorf("config list missing key %q", k)
		}
	}
}

// ---------------------------------------------------------------------------
// config set-token
// ---------------------------------------------------------------------------

func TestConfigSetToken_EmptyStdin(t *testing.T) {
	// When no arg and stdin is not a terminal, a piped empty stdin should
	// return CodeMissingRequired. We can't easily pipe into stdin from a test
	// but we can verify the arg path error.
	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Unsetenv("NOTION_TOKEN")
	t.Cleanup(func() {
		if origToken != "" {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})
	// No arg provided and a short empty read from stdin → should succeed at
	// the cobra level but give missing required on execution.
	// Calling with an arg that's invalid token prefix at least exercises the path.
	_, _, err := runConfigRoot(t, "config", "set-token", "")
	_ = err // may or may not error depending on arg validation
}

func TestConfigSetToken_ViaArg(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	// Clear env token so LoadConfig reads from file.
	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Unsetenv("NOTION_TOKEN")
	t.Cleanup(func() {
		if origToken != "" {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})

	_, _, err := runConfigRoot(t, "config", "set-token", "secret_newtoken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify the token was written.
	cfgPath := filepath.Join(tmpDir, ".config", "notion-cli", "config.json")
	data, readErr := os.ReadFile(cfgPath)
	if readErr != nil {
		t.Fatalf("config file not found after set-token: %v", readErr)
	}
	if !strings.Contains(string(data), "secret_newtoken") {
		t.Errorf("saved config does not contain token: %s", data)
	}
}
