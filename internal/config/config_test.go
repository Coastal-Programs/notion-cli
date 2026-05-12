package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()

	if cfg.BaseURL != "https://api.notion.com/v1" {
		t.Errorf("BaseURL = %q, want default", cfg.BaseURL)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", cfg.MaxRetries)
	}
	if cfg.BaseDelayMs != 1000 {
		t.Errorf("BaseDelayMs = %d, want 1000", cfg.BaseDelayMs)
	}
	if cfg.MaxDelayMs != 30000 {
		t.Errorf("MaxDelayMs = %d, want 30000", cfg.MaxDelayMs)
	}
	if !cfg.CacheEnabled {
		t.Error("CacheEnabled should default to true")
	}
	if cfg.CacheMaxSize != 1000 {
		t.Errorf("CacheMaxSize = %d, want 1000", cfg.CacheMaxSize)
	}
	if !cfg.DiskCacheEnabled {
		t.Error("DiskCacheEnabled should default to true")
	}
	if !cfg.HTTPKeepAlive {
		t.Error("HTTPKeepAlive should default to true")
	}
	if cfg.Verbose {
		t.Error("Verbose should default to false")
	}
}

func TestLoadConfig_EnvVarsOverrideDefaults(t *testing.T) {
	// Save and restore environment.
	envVars := []string{
		"NOTION_TOKEN", "NOTION_CLI_BASE_URL", "NOTION_CLI_MAX_RETRIES",
		"NOTION_CLI_BASE_DELAY", "NOTION_CLI_MAX_DELAY",
		"NOTION_CLI_CACHE_ENABLED", "NOTION_CLI_CACHE_MAX_SIZE",
		"NOTION_CLI_DISK_CACHE_ENABLED", "NOTION_CLI_HTTP_KEEP_ALIVE",
		"NOTION_CLI_VERBOSE",
	}
	saved := make(map[string]string)
	for _, k := range envVars {
		saved[k] = os.Getenv(k)
	}
	t.Cleanup(func() {
		for _, k := range envVars {
			if saved[k] == "" {
				_ = os.Unsetenv(k)
			} else {
				_ = os.Setenv(k, saved[k])
			}
		}
	})

	_ = os.Setenv("NOTION_TOKEN", "secret_test123")
	_ = os.Setenv("NOTION_CLI_BASE_URL", "https://custom.api.com")
	_ = os.Setenv("NOTION_CLI_MAX_RETRIES", "5")
	_ = os.Setenv("NOTION_CLI_BASE_DELAY", "2000")
	_ = os.Setenv("NOTION_CLI_MAX_DELAY", "60000")
	_ = os.Setenv("NOTION_CLI_CACHE_ENABLED", "false")
	_ = os.Setenv("NOTION_CLI_CACHE_MAX_SIZE", "500")
	_ = os.Setenv("NOTION_CLI_DISK_CACHE_ENABLED", "0")
	_ = os.Setenv("NOTION_CLI_HTTP_KEEP_ALIVE", "no")
	_ = os.Setenv("NOTION_CLI_VERBOSE", "1")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if cfg.Token != "secret_test123" {
		t.Errorf("Token = %q, want %q", cfg.Token, "secret_test123")
	}
	if cfg.BaseURL != "https://custom.api.com" {
		t.Errorf("BaseURL = %q", cfg.BaseURL)
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", cfg.MaxRetries)
	}
	if cfg.BaseDelayMs != 2000 {
		t.Errorf("BaseDelayMs = %d, want 2000", cfg.BaseDelayMs)
	}
	if cfg.MaxDelayMs != 60000 {
		t.Errorf("MaxDelayMs = %d, want 60000", cfg.MaxDelayMs)
	}
	if cfg.CacheEnabled {
		t.Error("CacheEnabled should be false")
	}
	if cfg.CacheMaxSize != 500 {
		t.Errorf("CacheMaxSize = %d, want 500", cfg.CacheMaxSize)
	}
	if cfg.DiskCacheEnabled {
		t.Error("DiskCacheEnabled should be false")
	}
	if cfg.HTTPKeepAlive {
		t.Error("HTTPKeepAlive should be false")
	}
	if !cfg.Verbose {
		t.Error("Verbose should be true")
	}
}

func TestParseBool(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"1", true},
		{"true", true},
		{"yes", true},
		{"on", true},
		{"0", false},
		{"false", false},
		{"no", false},
		{"off", false},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseBool(tt.input, !tt.want) // default is opposite
			if got != tt.want {
				t.Errorf("parseBool(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}

	t.Run("unknown falls back to default", func(t *testing.T) {
		if parseBool("maybe", true) != true {
			t.Error("unknown input should return default (true)")
		}
		if parseBool("maybe", false) != false {
			t.Error("unknown input should return default (false)")
		}
	})
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Use a temp dir as home to avoid touching real config.
	tmpDir := t.TempDir()

	// Override GetConfigPath by setting HOME.
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	// Clear env vars that would override file values.
	for _, k := range []string{
		"NOTION_TOKEN", "NOTION_CLI_BASE_URL", "NOTION_CLI_MAX_RETRIES",
		"NOTION_CLI_BASE_DELAY", "NOTION_CLI_MAX_DELAY",
		"NOTION_CLI_CACHE_ENABLED", "NOTION_CLI_CACHE_MAX_SIZE",
		"NOTION_CLI_DISK_CACHE_ENABLED", "NOTION_CLI_HTTP_KEEP_ALIVE",
		"NOTION_CLI_VERBOSE",
	} {
		origVal := os.Getenv(k)
		_ = os.Unsetenv(k)
		t.Cleanup(func() {
			if origVal != "" {
				_ = os.Setenv(k, origVal)
			}
		})
	}

	cfg := &Config{
		Token:            "secret_saved",
		BaseURL:          "https://api.notion.com/v1",
		MaxRetries:       7,
		BaseDelayMs:      500,
		MaxDelayMs:       10000,
		CacheEnabled:     false,
		CacheMaxSize:     200,
		DiskCacheEnabled: false,
		HTTPKeepAlive:    false,
		Verbose:          true,
	}

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	// Verify file exists and is readable.
	path := GetConfigPath()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("config file not readable: %v", err)
	}

	var saved Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("config file is not valid JSON: %v", err)
	}
	if saved.Token != "secret_saved" {
		t.Errorf("saved Token = %q", saved.Token)
	}

	// LoadConfig should read it back.
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}
	if loaded.Token != "secret_saved" {
		t.Errorf("loaded Token = %q, want %q", loaded.Token, "secret_saved")
	}
	if loaded.MaxRetries != 7 {
		t.Errorf("loaded MaxRetries = %d, want 7", loaded.MaxRetries)
	}
	if loaded.CacheEnabled {
		t.Error("loaded CacheEnabled should be false")
	}
	if loaded.HTTPKeepAlive {
		t.Error("loaded HTTPKeepAlive should be false")
	}
	if !loaded.Verbose {
		t.Error("loaded Verbose should be true")
	}
}

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Skip("could not determine home directory")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("GetConfigPath() = %q, want absolute path", path)
	}
	if filepath.Base(path) != "config.json" {
		t.Errorf("config file should be named config.json, got %q", filepath.Base(path))
	}
}

func TestGetDataDir(t *testing.T) {
	dir := GetDataDir()
	if dir == "" {
		t.Skip("could not determine home directory")
	}
	if !filepath.IsAbs(dir) {
		t.Errorf("GetDataDir() = %q, want absolute path", dir)
	}
}

func TestGetConfigValue(t *testing.T) {
	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Setenv("NOTION_TOKEN", "secret_getval")
	t.Cleanup(func() {
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})

	if got := GetConfigValue("token"); got != "secret_getval" {
		t.Errorf("GetConfigValue(token) = %q, want %q", got, "secret_getval")
	}
	if got := GetConfigValue("max_retries"); got == "" {
		t.Error("GetConfigValue(max_retries) should return a value")
	}
	if got := GetConfigValue("nonexistent_key"); got != "" {
		t.Errorf("GetConfigValue(nonexistent) = %q, want empty", got)
	}

	// Test all key names return something.
	keys := []string{
		"base_url", "base_delay_ms", "max_delay_ms",
		"cache_enabled", "cache_max_size", "disk_cache_enabled",
		"http_keep_alive", "verbose",
	}
	for _, k := range keys {
		if v := GetConfigValue(k); v == "" {
			t.Errorf("GetConfigValue(%q) returned empty string", k)
		}
	}
}

func TestLoadConfig_InvalidConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	// Write invalid JSON to config path.
	cfgDir := filepath.Join(tmpDir, ".config", "notion-cli")
	_ = os.MkdirAll(cfgDir, 0o755)
	_ = os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte("{invalid json"), 0o600)

	_, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig should return error for invalid JSON config file")
	}
}

func TestBuildVarsExist(t *testing.T) {
	// These are set by ldflags at build time; verify they have defaults.
	if Version == "" {
		t.Error("Version should have a default value")
	}
	if Commit == "" {
		t.Error("Commit should have a default value")
	}
	if Date == "" {
		t.Error("Date should have a default value")
	}
}

func TestOAuthClientCredentials(t *testing.T) {
	origID, origSecret := OAuthClientID, OAuthClientSecret
	t.Cleanup(func() {
		OAuthClientID = origID
		OAuthClientSecret = origSecret
	})

	t.Run("build time values", func(t *testing.T) {
		t.Setenv("NOTION_OAUTH_CLIENT_ID", "")
		t.Setenv("NOTION_OAUTH_SECRET", "")
		OAuthClientID = "  \"client-id\"  "
		OAuthClientSecret = "  'client-secret'  "

		clientID, clientSecret, ok := OAuthClientCredentials()
		if !ok {
			t.Fatal("OAuthClientCredentials() ok = false")
		}
		if clientID != "client-id" || clientSecret != "client-secret" {
			t.Fatalf("credentials = %q/%q", clientID, clientSecret)
		}
	})

	t.Run("runtime env fallback", func(t *testing.T) {
		OAuthClientID = ""
		OAuthClientSecret = ""
		t.Setenv("NOTION_OAUTH_CLIENT_ID", "env-client")
		t.Setenv("NOTION_OAUTH_SECRET", "env-secret")

		clientID, clientSecret, ok := OAuthClientCredentials()
		if !ok {
			t.Fatal("OAuthClientCredentials() ok = false")
		}
		if clientID != "env-client" || clientSecret != "env-secret" {
			t.Fatalf("credentials = %q/%q", clientID, clientSecret)
		}
	})

	t.Run("reject placeholders", func(t *testing.T) {
		OAuthClientID = "<from Notion integration settings>"
		OAuthClientSecret = "secret"
		t.Setenv("NOTION_OAUTH_CLIENT_ID", "")
		t.Setenv("NOTION_OAUTH_SECRET", "")

		if _, _, ok := OAuthClientCredentials(); ok {
			t.Fatal("OAuthClientCredentials() ok = true for placeholder client id")
		}
	})
}

func TestOAuthConfigFields(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	// Clear env vars.
	for _, k := range []string{"NOTION_TOKEN"} {
		origVal := os.Getenv(k)
		_ = os.Unsetenv(k)
		t.Cleanup(func() {
			if origVal != "" {
				_ = os.Setenv(k, origVal)
			}
		})
	}

	cfg := &Config{
		Token:              "secret_manual",
		OAuthAccessToken:   "ntn_oauth_token_123",
		OAuthWorkspaceID:   "ws-123",
		OAuthWorkspaceName: "Test Workspace",
		OAuthBotID:         "bot-456",
	}

	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if loaded.OAuthAccessToken != "ntn_oauth_token_123" {
		t.Errorf("OAuthAccessToken = %q, want %q", loaded.OAuthAccessToken, "ntn_oauth_token_123")
	}
	if loaded.OAuthWorkspaceID != "ws-123" {
		t.Errorf("OAuthWorkspaceID = %q, want %q", loaded.OAuthWorkspaceID, "ws-123")
	}
	if loaded.OAuthWorkspaceName != "Test Workspace" {
		t.Errorf("OAuthWorkspaceName = %q, want %q", loaded.OAuthWorkspaceName, "Test Workspace")
	}
	if loaded.OAuthBotID != "bot-456" {
		t.Errorf("OAuthBotID = %q, want %q", loaded.OAuthBotID, "bot-456")
	}
}

func TestHasOAuthToken(t *testing.T) {
	cfg := &Config{}
	if cfg.HasOAuthToken() {
		t.Error("HasOAuthToken() should be false for empty config")
	}

	cfg.OAuthAccessToken = "ntn_token"
	if !cfg.HasOAuthToken() {
		t.Error("HasOAuthToken() should be true when token is set")
	}
}

func TestClearOAuth(t *testing.T) {
	cfg := &Config{
		Token:              "secret_keep",
		OAuthAccessToken:   "ntn_clear",
		OAuthWorkspaceID:   "ws-clear",
		OAuthWorkspaceName: "Clear Me",
		OAuthBotID:         "bot-clear",
	}

	cfg.ClearOAuth()

	if cfg.OAuthAccessToken != "" {
		t.Errorf("OAuthAccessToken should be empty after ClearOAuth, got %q", cfg.OAuthAccessToken)
	}
	if cfg.OAuthWorkspaceID != "" {
		t.Errorf("OAuthWorkspaceID should be empty after ClearOAuth, got %q", cfg.OAuthWorkspaceID)
	}
	if cfg.OAuthWorkspaceName != "" {
		t.Errorf("OAuthWorkspaceName should be empty after ClearOAuth, got %q", cfg.OAuthWorkspaceName)
	}
	if cfg.OAuthBotID != "" {
		t.Errorf("OAuthBotID should be empty after ClearOAuth, got %q", cfg.OAuthBotID)
	}
	// Token should not be cleared.
	if cfg.Token != "secret_keep" {
		t.Errorf("Token should remain after ClearOAuth, got %q", cfg.Token)
	}
}

func TestAuthMethod(t *testing.T) {
	// Save and restore NOTION_TOKEN.
	origToken := os.Getenv("NOTION_TOKEN")
	t.Cleanup(func() {
		if origToken != "" {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		} else {
			_ = os.Unsetenv("NOTION_TOKEN")
		}
	})

	t.Run("env takes precedence", func(t *testing.T) {
		_ = os.Setenv("NOTION_TOKEN", "secret_env")
		cfg := &Config{OAuthAccessToken: "ntn_oauth", Token: "secret_manual"}
		if m := cfg.AuthMethod(); m != "env" {
			t.Errorf("AuthMethod() = %q, want %q", m, "env")
		}
	})

	t.Run("oauth when no env", func(t *testing.T) {
		_ = os.Unsetenv("NOTION_TOKEN")
		cfg := &Config{OAuthAccessToken: "ntn_oauth", Token: "secret_manual"}
		if m := cfg.AuthMethod(); m != "oauth" {
			t.Errorf("AuthMethod() = %q, want %q", m, "oauth")
		}
	})

	t.Run("token when no oauth", func(t *testing.T) {
		_ = os.Unsetenv("NOTION_TOKEN")
		cfg := &Config{Token: "secret_manual"}
		if m := cfg.AuthMethod(); m != "token" {
			t.Errorf("AuthMethod() = %q, want %q", m, "token")
		}
	})

	t.Run("none when nothing set", func(t *testing.T) {
		_ = os.Unsetenv("NOTION_TOKEN")
		cfg := &Config{}
		if m := cfg.AuthMethod(); m != "none" {
			t.Errorf("AuthMethod() = %q, want %q", m, "none")
		}
	})
}

func TestGetConfigValue_OAuthKeys(t *testing.T) {
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

	cfg := &Config{
		OAuthAccessToken:   "ntn_test",
		OAuthWorkspaceID:   "ws-test",
		OAuthWorkspaceName: "Test WS",
		OAuthBotID:         "bot-test",
	}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	tests := map[string]string{
		"oauth_access_token":   "ntn_test",
		"oauth_workspace_id":   "ws-test",
		"oauth_workspace_name": "Test WS",
		"oauth_bot_id":         "bot-test",
		"auth_method":          "oauth",
	}
	for key, want := range tests {
		if got := GetConfigValue(key); got != want {
			t.Errorf("GetConfigValue(%q) = %q, want %q", key, got, want)
		}
	}
}

func TestNeedsRefresh_NoRefreshToken(t *testing.T) {
	cfg := &Config{OAuthTokenExpiresAt: time.Now().Add(-1 * time.Minute).Format(time.RFC3339)}
	if cfg.NeedsRefresh() {
		t.Error("NeedsRefresh() should be false when no refresh token")
	}
}

func TestNeedsRefresh_NoExpiresAt(t *testing.T) {
	cfg := &Config{OAuthRefreshToken: "rt_abc"}
	if cfg.NeedsRefresh() {
		t.Error("NeedsRefresh() should be false when no expires_at")
	}
}

func TestNeedsRefresh_InvalidTimestamp(t *testing.T) {
	cfg := &Config{OAuthRefreshToken: "rt_abc", OAuthTokenExpiresAt: "not-a-date"}
	if cfg.NeedsRefresh() {
		t.Error("NeedsRefresh() should be false on parse error")
	}
}

func TestNeedsRefresh_ExpiresInFuture_NotSoon(t *testing.T) {
	cfg := &Config{
		OAuthRefreshToken:   "rt_abc",
		OAuthTokenExpiresAt: time.Now().Add(10 * time.Minute).Format(time.RFC3339),
	}
	if cfg.NeedsRefresh() {
		t.Error("NeedsRefresh() should be false when expires in 10 min")
	}
}

func TestNeedsRefresh_ExpiringSoon(t *testing.T) {
	cfg := &Config{
		OAuthRefreshToken:   "rt_abc",
		OAuthTokenExpiresAt: time.Now().Add(4 * time.Minute).Format(time.RFC3339),
	}
	if !cfg.NeedsRefresh() {
		t.Error("NeedsRefresh() should be true when expires in 4 min")
	}
}

func TestNeedsRefresh_AlreadyExpired(t *testing.T) {
	cfg := &Config{
		OAuthRefreshToken:   "rt_abc",
		OAuthTokenExpiresAt: time.Now().Add(-1 * time.Minute).Format(time.RFC3339),
	}
	if !cfg.NeedsRefresh() {
		t.Error("NeedsRefresh() should be true when already expired")
	}
}

func TestGetConfigValue_OAuthRefreshToken(t *testing.T) {
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

	cfg := &Config{OAuthRefreshToken: "rt_refresh_123"}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	if got := GetConfigValue("oauth_refresh_token"); got != "rt_refresh_123" {
		t.Errorf("GetConfigValue(oauth_refresh_token) = %q, want %q", got, "rt_refresh_123")
	}
}

func TestGetConfigValue_OAuthTokenExpiresAt(t *testing.T) {
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

	expiry := time.Now().Add(1 * time.Hour).Format(time.RFC3339)
	cfg := &Config{OAuthTokenExpiresAt: expiry}
	if err := SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig() error: %v", err)
	}

	if got := GetConfigValue("oauth_token_expires_at"); got != expiry {
		t.Errorf("GetConfigValue(oauth_token_expires_at) = %q, want %q", got, expiry)
	}
}

func TestLoadConfig_FileWithPartialValues(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	// Clear env vars.
	for _, k := range []string{"NOTION_TOKEN", "NOTION_CLI_MAX_RETRIES"} {
		origVal := os.Getenv(k)
		_ = os.Unsetenv(k)
		t.Cleanup(func() {
			if origVal != "" {
				_ = os.Setenv(k, origVal)
			}
		})
	}

	// Write config with only token set.
	cfgDir := filepath.Join(tmpDir, ".config", "notion-cli")
	_ = os.MkdirAll(cfgDir, 0o755)
	_ = os.WriteFile(filepath.Join(cfgDir, "config.json"),
		[]byte(`{"token": "secret_partial"}`), 0o600)

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error: %v", err)
	}

	if cfg.Token != "secret_partial" {
		t.Errorf("Token = %q, want %q", cfg.Token, "secret_partial")
	}
	// Defaults should still be in place.
	if cfg.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want default 3", cfg.MaxRetries)
	}
	if !cfg.CacheEnabled {
		t.Error("CacheEnabled should still be default true")
	}
}

func TestSlugFromWorkspaceName(t *testing.T) {
	tests := []struct {
		name        string
		workspaceID string
		want        string
	}{
		{"Haven", "abcd1234-0000", "haven"},
		{"Ross Berger's Notion", "abcd1234-0000", "ross-berger-s-notion"},
		{"  Work  / Team!! ", "abcd1234-0000", "work-team"},
		{"!!!", "17ab3186-873d-418f-b899-c3f6a43f68de", "notion-17ab3186"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SlugFromWorkspaceName(tt.name, tt.workspaceID); got != tt.want {
				t.Errorf("SlugFromWorkspaceName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestChooseWorkspaceSlug_CollisionAndReauth(t *testing.T) {
	creds := &CredentialsFile{Workspaces: map[string]WorkspaceCredential{
		"haven": {Slug: "haven", WorkspaceID: "ws-1"},
	}}
	if got, err := ChooseWorkspaceSlug(creds, "", "Haven", "ws-1"); err != nil || got != "haven" {
		t.Fatalf("reauth slug = %q, %v; want haven", got, err)
	}
	if got, err := ChooseWorkspaceSlug(creds, "", "Haven", "abcdef12-9999"); err != nil || got != "haven-abcdef12" {
		t.Fatalf("collision slug = %q, %v; want haven-abcdef12", got, err)
	}
	if got, err := ChooseWorkspaceSlug(creds, "personal", "Haven", "ws-2"); err != nil || got != "personal" {
		t.Fatalf("explicit slug = %q, %v; want personal", got, err)
	}
}

func TestChooseWorkspaceSlug_CollisionExhaustion(t *testing.T) {
	creds := &CredentialsFile{Workspaces: map[string]WorkspaceCredential{
		"haven":     {Slug: "haven", WorkspaceID: "ws-1"},
		"haven-new": {Slug: "haven-new", WorkspaceID: "ws-prefix"},
	}}
	for i := 2; i < 1000; i++ {
		slug := "haven-" + strconv.Itoa(i)
		creds.Workspaces[slug] = WorkspaceCredential{Slug: slug, WorkspaceID: "ws-collision"}
	}
	if _, err := ChooseWorkspaceSlug(creds, "", "Haven", "new"); err == nil {
		t.Fatal("ChooseWorkspaceSlug succeeded after exhausting generated slugs")
	}
}

func TestWorkspaceCredential_LoadConfigForWorkspace(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })
	t.Setenv("NOTION_WORKSPACE", "")
	t.Setenv("NOTION_TOKEN", "")

	store := NewMemorySecretStore()
	restore := SetSecretStoreForTest(store)
	t.Cleanup(restore)

	err := SaveWorkspaceCredential(WorkspaceCredential{
		Slug:                "haven",
		AuthMethod:          AuthMethodOAuth,
		WorkspaceID:         "ws-haven",
		WorkspaceName:       "Haven",
		BotID:               "bot-haven",
		OAuthTokenExpiresAt: "2030-01-01T00:00:00Z",
	}, WorkspaceSecrets{
		OAuthAccessToken:  "ntn_access",
		OAuthRefreshToken: "rt_refresh",
	}, true)
	if err != nil {
		t.Fatalf("SaveWorkspaceCredential: %v", err)
	}

	cfg, active, err := LoadConfigForWorkspace("")
	if err != nil {
		t.Fatalf("LoadConfigForWorkspace: %v", err)
	}
	if active.DisplayName() != "haven" {
		t.Fatalf("active workspace = %q, want haven", active.DisplayName())
	}
	if cfg.OAuthAccessToken != "ntn_access" || cfg.OAuthRefreshToken != "rt_refresh" {
		t.Fatalf("oauth secrets not loaded: access=%q refresh=%q", cfg.OAuthAccessToken, cfg.OAuthRefreshToken)
	}
	if cfg.OAuthWorkspaceName != "Haven" || cfg.OAuthWorkspaceID != "ws-haven" {
		t.Fatalf("workspace metadata not loaded: %#v", cfg)
	}
}

func TestLoadConfigForWorkspace_PrecedenceCascade(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })
	t.Setenv("NOTION_TOKEN", "")
	t.Setenv("NOTION_WORKSPACE", "")

	store := NewMemorySecretStore()
	restore := SetSecretStoreForTest(store)
	t.Cleanup(restore)

	if err := SaveConfig(&Config{Token: "secret_legacy"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	if err := SaveWorkspaceCredential(WorkspaceCredential{
		Slug:       "personal",
		AuthMethod: AuthMethodOAuth,
	}, WorkspaceSecrets{OAuthAccessToken: "ntn_personal", OAuthRefreshToken: "rt_personal"}, true); err != nil {
		t.Fatalf("SaveWorkspaceCredential personal: %v", err)
	}
	if err := SaveWorkspaceCredential(WorkspaceCredential{
		Slug:       "work",
		AuthMethod: AuthMethodOAuth,
	}, WorkspaceSecrets{OAuthAccessToken: "ntn_work", OAuthRefreshToken: "rt_work"}, false); err != nil {
		t.Fatalf("SaveWorkspaceCredential work: %v", err)
	}

	t.Setenv("NOTION_TOKEN", "secret_env")
	cfg, active, err := LoadConfigForWorkspace("work")
	if err != nil {
		t.Fatalf("env explicit LoadConfigForWorkspace: %v", err)
	}
	if active.DisplayName() != "work" || cfg.AuthMethod() != "env" || cfg.Token != "secret_env" {
		t.Fatalf("env should mask explicit workspace: active=%s method=%s token=%q", active.DisplayName(), cfg.AuthMethod(), cfg.Token)
	}

	t.Setenv("NOTION_TOKEN", "")
	cfg, active, err = LoadConfigForWorkspace("work")
	if err != nil {
		t.Fatalf("explicit LoadConfigForWorkspace: %v", err)
	}
	if active.DisplayName() != "work" || cfg.AuthMethod() != "oauth" || cfg.OAuthAccessToken != "ntn_work" {
		t.Fatalf("explicit workspace not selected: active=%s method=%s token=%q", active.DisplayName(), cfg.AuthMethod(), cfg.OAuthAccessToken)
	}

	t.Setenv("NOTION_WORKSPACE", "work")
	cfg, active, err = LoadConfigForWorkspace("")
	if err != nil {
		t.Fatalf("env workspace LoadConfigForWorkspace: %v", err)
	}
	if active.DisplayName() != "work" || cfg.OAuthAccessToken != "ntn_work" {
		t.Fatalf("NOTION_WORKSPACE not selected: active=%s token=%q", active.DisplayName(), cfg.OAuthAccessToken)
	}

	t.Setenv("NOTION_WORKSPACE", "")
	cfg, active, err = LoadConfigForWorkspace("")
	if err != nil {
		t.Fatalf("default workspace LoadConfigForWorkspace: %v", err)
	}
	if active.DisplayName() != "personal" || cfg.AuthMethod() != "oauth" || cfg.OAuthAccessToken != "ntn_personal" {
		t.Fatalf("default workspace not selected: active=%s method=%s token=%q", active.DisplayName(), cfg.AuthMethod(), cfg.OAuthAccessToken)
	}

	if _, err := DeleteWorkspaceCredential("personal"); err != nil {
		t.Fatalf("DeleteWorkspaceCredential personal: %v", err)
	}
	if _, err := DeleteWorkspaceCredential("work"); err != nil {
		t.Fatalf("DeleteWorkspaceCredential work: %v", err)
	}
	cfg, active, err = LoadConfigForWorkspace("")
	if err != nil {
		t.Fatalf("legacy fallback LoadConfigForWorkspace: %v", err)
	}
	if !active.Legacy || cfg.AuthMethod() != "token" || cfg.Token != "secret_legacy" {
		t.Fatalf("legacy fallback not selected: active=%#v method=%s token=%q", active, cfg.AuthMethod(), cfg.Token)
	}
}

type failingSecretStore struct {
	err error
}

func (f failingSecretStore) Get(string) (string, error) {
	return "", f.err
}

func (f failingSecretStore) Set(string, string) error {
	return f.err
}

func (f failingSecretStore) Delete(string) error {
	return f.err
}

func TestWorkspaceCredential_SecretStoreErrors(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })
	t.Setenv("NOTION_TOKEN", "")
	t.Setenv("NOTION_WORKSPACE", "")

	storeErr := errors.New("keychain unavailable")
	restore := SetSecretStoreForTest(failingSecretStore{err: storeErr})
	t.Cleanup(restore)

	if err := SaveWorkspaceCredential(WorkspaceCredential{
		Slug:       "haven",
		AuthMethod: AuthMethodOAuth,
	}, WorkspaceSecrets{OAuthAccessToken: "ntn_haven"}, true); !errors.Is(err, storeErr) {
		t.Fatalf("SaveWorkspaceCredential error = %v, want %v", err, storeErr)
	}

	if err := SaveCredentials(&CredentialsFile{
		DefaultWorkspace: "haven",
		Workspaces: map[string]WorkspaceCredential{
			"haven": {Slug: "haven", AuthMethod: AuthMethodOAuth},
		},
	}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}
	if _, err := LoadWorkspaceSecrets("haven"); !errors.Is(err, storeErr) {
		t.Fatalf("LoadWorkspaceSecrets error = %v, want %v", err, storeErr)
	}
	if _, _, err := LoadConfigForWorkspace("haven"); !errors.Is(err, storeErr) {
		t.Fatalf("LoadConfigForWorkspace error = %v, want %v", err, storeErr)
	}
}

func TestWorkspaceCredential_RequiresNamedDefaultWhenWorkspacesExist(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })
	t.Setenv("NOTION_WORKSPACE", "")
	t.Setenv("NOTION_TOKEN", "")

	if _, active, err := LoadConfigForWorkspace(""); err != nil || !active.Legacy {
		t.Fatalf("empty credentials should use legacy fallback: active=%#v err=%v", active, err)
	}

	if err := SaveCredentials(&CredentialsFile{
		Workspaces: map[string]WorkspaceCredential{
			"personal": {Slug: "personal", AuthMethod: AuthMethodOAuth, WorkspaceID: "ws-personal"},
			"work":     {Slug: "work", AuthMethod: AuthMethodOAuth, WorkspaceID: "ws-work"},
		},
	}); err != nil {
		t.Fatalf("SaveCredentials: %v", err)
	}
	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if creds.DefaultWorkspace != "personal" {
		t.Fatalf("default workspace = %q, want first sorted slug personal", creds.DefaultWorkspace)
	}
	if err := SetDefaultWorkspace(LegacyWorkspaceSlug); err == nil {
		t.Fatal("SetDefaultWorkspace(default) succeeded with named workspaces")
	}
}

func TestDeleteWorkspaceCredential_PromotesDefault(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	_ = os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { _ = os.Setenv("HOME", origHome) })

	store := NewMemorySecretStore()
	restore := SetSecretStoreForTest(store)
	t.Cleanup(restore)

	if err := SaveWorkspaceCredential(WorkspaceCredential{
		Slug:       "alpha",
		AuthMethod: AuthMethodOAuth,
	}, WorkspaceSecrets{OAuthAccessToken: "ntn_alpha", OAuthRefreshToken: "rt_alpha"}, true); err != nil {
		t.Fatalf("SaveWorkspaceCredential alpha: %v", err)
	}
	if err := SaveWorkspaceCredential(WorkspaceCredential{
		Slug:       "bravo",
		AuthMethod: AuthMethodOAuth,
	}, WorkspaceSecrets{OAuthAccessToken: "ntn_bravo", OAuthRefreshToken: "rt_bravo"}, false); err != nil {
		t.Fatalf("SaveWorkspaceCredential bravo: %v", err)
	}

	if _, err := DeleteWorkspaceCredential("alpha"); err != nil {
		t.Fatalf("DeleteWorkspaceCredential: %v", err)
	}
	creds, err := LoadCredentials()
	if err != nil {
		t.Fatalf("LoadCredentials: %v", err)
	}
	if creds.DefaultWorkspace != "bravo" {
		t.Fatalf("default workspace = %q, want bravo", creds.DefaultWorkspace)
	}
}
