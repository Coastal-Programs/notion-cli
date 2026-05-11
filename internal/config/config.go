// Package config handles configuration loading from environment variables
// and a JSON config file. Environment variables take precedence over the
// config file. Build-time variables (Version, Commit, Date) are set via
// ldflags.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// Build-time variables set via ldflags.
var (
	Version           = "dev"
	Commit            = "none"
	Date              = "unknown"
	OAuthClientID     = ""
	OAuthClientSecret = ""
)

// Config holds all CLI configuration values.
type Config struct {
	Token       string `json:"token,omitempty"`
	BaseURL     string `json:"base_url,omitempty"`
	MaxRetries  int    `json:"max_retries,omitempty"`
	BaseDelayMs int    `json:"base_delay_ms,omitempty"`
	MaxDelayMs  int    `json:"max_delay_ms,omitempty"`
	// CacheEnabled, CacheMaxSize, DiskCacheEnabled are reserved for a future
	// in-memory/disk response cache. They are persisted and exposed via
	// `config get/set` for forward compatibility but currently have no effect
	// on runtime behavior. Only the workspace database cache (see
	// internal/cache/workspace.go) is wired up today.
	CacheEnabled     bool `json:"cache_enabled"`
	CacheMaxSize     int  `json:"cache_max_size,omitempty"`
	DiskCacheEnabled bool `json:"disk_cache_enabled"`
	HTTPKeepAlive    bool `json:"http_keep_alive"`
	Verbose          bool `json:"verbose,omitempty"`

	// OAuth fields (populated by 'auth login').
	OAuthAccessToken    string `json:"oauth_access_token,omitempty"`
	OAuthWorkspaceID    string `json:"oauth_workspace_id,omitempty"`
	OAuthWorkspaceName  string `json:"oauth_workspace_name,omitempty"`
	OAuthBotID          string `json:"oauth_bot_id,omitempty"`
	OAuthRefreshToken   string `json:"oauth_refresh_token,omitempty"`
	OAuthTokenExpiresAt string `json:"oauth_token_expires_at,omitempty"` // RFC3339
}

// HasOAuthToken reports whether an OAuth access token is configured.
func (c *Config) HasOAuthToken() bool {
	return c.OAuthAccessToken != ""
}

// ClearOAuth removes all OAuth-related fields from the config.
func (c *Config) ClearOAuth() {
	c.OAuthAccessToken = ""
	c.OAuthWorkspaceID = ""
	c.OAuthWorkspaceName = ""
	c.OAuthBotID = ""
	c.OAuthRefreshToken = ""
	c.OAuthTokenExpiresAt = ""
}

// NeedsRefresh reports whether the OAuth access token expires within 5 minutes
// and a refresh token is available. Returns false when either token is absent
// or the expiry timestamp cannot be parsed.
func (c *Config) NeedsRefresh() bool {
	if c.OAuthRefreshToken == "" || c.OAuthTokenExpiresAt == "" {
		return false
	}
	t, err := time.Parse(time.RFC3339, c.OAuthTokenExpiresAt)
	if err != nil {
		return false
	}
	return time.Until(t) < 5*time.Minute
}

// AuthMethod returns "oauth", "token", or "none" describing how the CLI
// is currently authenticated.
func (c *Config) AuthMethod() string {
	if os.Getenv("NOTION_TOKEN") != "" {
		return "env"
	}
	if c.OAuthAccessToken != "" {
		return "oauth"
	}
	if c.Token != "" {
		return "token"
	}
	return "none"
}

// defaults returns a Config with default values.
func defaults() *Config {
	return &Config{
		BaseURL:          "https://api.notion.com/v1",
		MaxRetries:       3,
		BaseDelayMs:      1000,
		MaxDelayMs:       30000,
		CacheEnabled:     true,
		CacheMaxSize:     1000,
		DiskCacheEnabled: true,
		HTTPKeepAlive:    true,
		Verbose:          false,
	}
}

// GetConfigPath returns the path to the JSON config file.
func GetConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "notion-cli", "config.json")
}

// GetDataDir returns the path to the CLI data directory (cache, workspace).
func GetDataDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".notion-cli")
}

// LoadConfig reads configuration from the config file and environment
// variables. Environment variables always take precedence.
func LoadConfig() (*Config, error) {
	cfg := defaults()

	// Layer 1: config file.
	if err := loadFromFile(cfg); err != nil {
		// File not existing is fine; other errors are reported.
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	// Layer 2: environment variables (override file values).
	loadFromEnv(cfg)

	return cfg, nil
}

// loadFromFile reads the JSON config file into cfg, overwriting only fields
// that are present in the file.
func loadFromFile(cfg *Config) error {
	path := GetConfigPath()
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Decode into a separate struct so we only overwrite fields the file
	// actually contains.
	var fileCfg Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return err
	}

	if fileCfg.Token != "" {
		cfg.Token = fileCfg.Token
	}
	if fileCfg.BaseURL != "" {
		cfg.BaseURL = fileCfg.BaseURL
	}
	if fileCfg.OAuthAccessToken != "" {
		cfg.OAuthAccessToken = fileCfg.OAuthAccessToken
	}
	if fileCfg.OAuthWorkspaceID != "" {
		cfg.OAuthWorkspaceID = fileCfg.OAuthWorkspaceID
	}
	if fileCfg.OAuthWorkspaceName != "" {
		cfg.OAuthWorkspaceName = fileCfg.OAuthWorkspaceName
	}
	if fileCfg.OAuthBotID != "" {
		cfg.OAuthBotID = fileCfg.OAuthBotID
	}
	if fileCfg.OAuthRefreshToken != "" {
		cfg.OAuthRefreshToken = fileCfg.OAuthRefreshToken
	}
	if fileCfg.OAuthTokenExpiresAt != "" {
		cfg.OAuthTokenExpiresAt = fileCfg.OAuthTokenExpiresAt
	}
	// Use a raw map to detect explicitly set fields, including zero values.
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err == nil {
		if _, exists := raw["max_retries"]; exists {
			cfg.MaxRetries = fileCfg.MaxRetries
		}
		if _, exists := raw["base_delay_ms"]; exists {
			cfg.BaseDelayMs = fileCfg.BaseDelayMs
		}
		if _, exists := raw["max_delay_ms"]; exists {
			cfg.MaxDelayMs = fileCfg.MaxDelayMs
		}
		if _, exists := raw["cache_max_size"]; exists {
			cfg.CacheMaxSize = fileCfg.CacheMaxSize
		}
		if _, exists := raw["cache_enabled"]; exists {
			cfg.CacheEnabled = fileCfg.CacheEnabled
		}
		if _, exists := raw["disk_cache_enabled"]; exists {
			cfg.DiskCacheEnabled = fileCfg.DiskCacheEnabled
		}
		if _, exists := raw["http_keep_alive"]; exists {
			cfg.HTTPKeepAlive = fileCfg.HTTPKeepAlive
		}
		if _, exists := raw["verbose"]; exists {
			cfg.Verbose = fileCfg.Verbose
		}
	}

	return nil
}

// loadFromEnv applies environment variable overrides to cfg.
func loadFromEnv(cfg *Config) {
	if v := os.Getenv("NOTION_TOKEN"); v != "" {
		cfg.Token = v
	}
	if v := os.Getenv("NOTION_CLI_BASE_URL"); v != "" {
		cfg.BaseURL = v
	}
	if v, err := strconv.Atoi(os.Getenv("NOTION_CLI_MAX_RETRIES")); err == nil {
		cfg.MaxRetries = v
	}
	if v, err := strconv.Atoi(os.Getenv("NOTION_CLI_BASE_DELAY")); err == nil {
		cfg.BaseDelayMs = v
	}
	if v, err := strconv.Atoi(os.Getenv("NOTION_CLI_MAX_DELAY")); err == nil {
		cfg.MaxDelayMs = v
	}
	if v := os.Getenv("NOTION_CLI_CACHE_ENABLED"); v != "" {
		cfg.CacheEnabled = parseBool(v, cfg.CacheEnabled)
	}
	if v, err := strconv.Atoi(os.Getenv("NOTION_CLI_CACHE_MAX_SIZE")); err == nil {
		cfg.CacheMaxSize = v
	}
	if v := os.Getenv("NOTION_CLI_DISK_CACHE_ENABLED"); v != "" {
		cfg.DiskCacheEnabled = parseBool(v, cfg.DiskCacheEnabled)
	}
	if v := os.Getenv("NOTION_CLI_HTTP_KEEP_ALIVE"); v != "" {
		cfg.HTTPKeepAlive = parseBool(v, cfg.HTTPKeepAlive)
	}
	if v := os.Getenv("NOTION_CLI_VERBOSE"); v != "" {
		cfg.Verbose = parseBool(v, cfg.Verbose)
	}
}

// parseBool parses common boolean representations, falling back to def on
// unrecognized input.
func parseBool(s string, def bool) bool {
	switch s {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	default:
		return def
	}
}

// SaveConfig writes the config to the JSON config file, creating directories
// as needed.
func SaveConfig(cfg *Config) error {
	path := GetConfigPath()
	if path == "" {
		return os.ErrNotExist
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// GetConfigValue returns a single config value by key name. It loads the
// full config and returns the string representation of the requested field.
func GetConfigValue(key string) string {
	cfg, err := LoadConfig()
	if err != nil {
		return ""
	}

	switch key {
	case "token":
		return cfg.Token
	case "base_url":
		return cfg.BaseURL
	case "max_retries":
		return strconv.Itoa(cfg.MaxRetries)
	case "base_delay_ms":
		return strconv.Itoa(cfg.BaseDelayMs)
	case "max_delay_ms":
		return strconv.Itoa(cfg.MaxDelayMs)
	case "cache_enabled":
		return strconv.FormatBool(cfg.CacheEnabled)
	case "cache_max_size":
		return strconv.Itoa(cfg.CacheMaxSize)
	case "disk_cache_enabled":
		return strconv.FormatBool(cfg.DiskCacheEnabled)
	case "http_keep_alive":
		return strconv.FormatBool(cfg.HTTPKeepAlive)
	case "verbose":
		return strconv.FormatBool(cfg.Verbose)
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
