package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
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
				os.Unsetenv(k)
			} else {
				os.Setenv(k, saved[k])
			}
		}
	})

	os.Setenv("NOTION_TOKEN", "secret_test123")
	os.Setenv("NOTION_CLI_BASE_URL", "https://custom.api.com")
	os.Setenv("NOTION_CLI_MAX_RETRIES", "5")
	os.Setenv("NOTION_CLI_BASE_DELAY", "2000")
	os.Setenv("NOTION_CLI_MAX_DELAY", "60000")
	os.Setenv("NOTION_CLI_CACHE_ENABLED", "false")
	os.Setenv("NOTION_CLI_CACHE_MAX_SIZE", "500")
	os.Setenv("NOTION_CLI_DISK_CACHE_ENABLED", "0")
	os.Setenv("NOTION_CLI_HTTP_KEEP_ALIVE", "no")
	os.Setenv("NOTION_CLI_VERBOSE", "1")

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
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	// Clear env vars that would override file values.
	for _, k := range []string{
		"NOTION_TOKEN", "NOTION_CLI_BASE_URL", "NOTION_CLI_MAX_RETRIES",
		"NOTION_CLI_BASE_DELAY", "NOTION_CLI_MAX_DELAY",
		"NOTION_CLI_CACHE_ENABLED", "NOTION_CLI_CACHE_MAX_SIZE",
		"NOTION_CLI_DISK_CACHE_ENABLED", "NOTION_CLI_HTTP_KEEP_ALIVE",
		"NOTION_CLI_VERBOSE",
	} {
		origVal := os.Getenv(k)
		os.Unsetenv(k)
		t.Cleanup(func() {
			if origVal != "" {
				os.Setenv(k, origVal)
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
	os.Setenv("NOTION_TOKEN", "secret_getval")
	t.Cleanup(func() {
		if origToken == "" {
			os.Unsetenv("NOTION_TOKEN")
		} else {
			os.Setenv("NOTION_TOKEN", origToken)
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
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	// Write invalid JSON to config path.
	cfgDir := filepath.Join(tmpDir, ".config", "notion-cli")
	os.MkdirAll(cfgDir, 0o755)
	os.WriteFile(filepath.Join(cfgDir, "config.json"), []byte("{invalid json"), 0o600)

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

func TestLoadConfig_FileWithPartialValues(t *testing.T) {
	tmpDir := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmpDir)
	t.Cleanup(func() { os.Setenv("HOME", origHome) })

	// Clear env vars.
	for _, k := range []string{"NOTION_TOKEN", "NOTION_CLI_MAX_RETRIES"} {
		origVal := os.Getenv(k)
		os.Unsetenv(k)
		t.Cleanup(func() {
			if origVal != "" {
				os.Setenv(k, origVal)
			}
		})
	}

	// Write config with only token set.
	cfgDir := filepath.Join(tmpDir, ".config", "notion-cli")
	os.MkdirAll(cfgDir, 0o755)
	os.WriteFile(filepath.Join(cfgDir, "config.json"),
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
