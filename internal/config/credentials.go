package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

const (
	// LegacyWorkspaceSlug is accepted in flags/env as an explicit request to
	// use the pre-workspace config.json credential path.
	LegacyWorkspaceSlug = "default"

	AuthMethodOAuth = "oauth"
)

var validWorkspaceSlug = regexp.MustCompile(`^[a-z0-9](?:[a-z0-9-]*[a-z0-9])?$`)

// CredentialsFile stores non-secret workspace metadata. Tokens live in the
// system keychain via SecretStore.
type CredentialsFile struct {
	DefaultWorkspace string                         `json:"default_workspace,omitempty"`
	Workspaces       map[string]WorkspaceCredential `json:"workspaces,omitempty"`
}

// WorkspaceCredential is safe to write to disk. Do not add token fields here.
type WorkspaceCredential struct {
	Slug                string `json:"slug"`
	AuthMethod          string `json:"auth_method"`
	WorkspaceID         string `json:"workspace_id,omitempty"`
	WorkspaceName       string `json:"workspace_name,omitempty"`
	WorkspaceIcon       string `json:"workspace_icon,omitempty"`
	BotID               string `json:"bot_id,omitempty"`
	OAuthTokenExpiresAt string `json:"oauth_token_expires_at,omitempty"`
	UpdatedAt           string `json:"updated_at,omitempty"`
}

// WorkspaceSecrets are stored in the system keychain.
type WorkspaceSecrets struct {
	OAuthAccessToken  string
	OAuthRefreshToken string
}

// ActiveWorkspace describes the resolved credential target for a command.
type ActiveWorkspace struct {
	Slug     string
	Source   string
	Legacy   bool
	Metadata *WorkspaceCredential
}

func (a *ActiveWorkspace) DisplayName() string {
	if a == nil || a.Legacy || a.Slug == "" {
		return LegacyWorkspaceSlug
	}
	return a.Slug
}

func emptyCredentialsFile() *CredentialsFile {
	return &CredentialsFile{Workspaces: map[string]WorkspaceCredential{}}
}

// GetCredentialsPath returns the workspace credential metadata file path.
func GetCredentialsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "notion-cli", "credentials.json")
}

// GetConfigPathForWorkspace returns the most relevant on-disk config metadata
// path for an active workspace. Named workspaces store metadata in
// credentials.json; the legacy default stores config in config.json.
func GetConfigPathForWorkspace(slug string) string {
	if slug == "" || slug == LegacyWorkspaceSlug {
		return GetConfigPath()
	}
	return GetCredentialsPath()
}

// GetDataDirForWorkspace returns the workspace-scoped data directory.
func GetDataDirForWorkspace(slug string) string {
	base := GetDataDir()
	if base == "" || slug == "" || slug == LegacyWorkspaceSlug {
		return base
	}
	return filepath.Join(base, "workspaces", slug)
}

// GetWorkspaceCachePath returns the workspace database cache path.
func GetWorkspaceCachePath(slug string) string {
	dataDir := GetDataDirForWorkspace(slug)
	if dataDir == "" {
		return ""
	}
	return filepath.Join(dataDir, "databases.json")
}

// LoadCredentials reads non-secret workspace metadata from disk.
func LoadCredentials() (*CredentialsFile, error) {
	path := GetCredentialsPath()
	if path == "" {
		return emptyCredentialsFile(), nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return emptyCredentialsFile(), nil
		}
		return nil, err
	}
	var creds CredentialsFile
	if err := json.Unmarshal(data, &creds); err != nil {
		return nil, err
	}
	if creds.Workspaces == nil {
		creds.Workspaces = map[string]WorkspaceCredential{}
	}
	creds.ensureDefaultWorkspace()
	return &creds, nil
}

// SaveCredentials writes non-secret workspace metadata atomically.
func SaveCredentials(creds *CredentialsFile) error {
	path := GetCredentialsPath()
	if path == "" {
		return os.ErrNotExist
	}
	if creds == nil {
		creds = emptyCredentialsFile()
	}
	if creds.Workspaces == nil {
		creds.Workspaces = map[string]WorkspaceCredential{}
	}
	creds.ensureDefaultWorkspace()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(creds, "", "  ")
	if err != nil {
		return err
	}
	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmpPath, path)
}

// SortedWorkspaceSlugs returns workspace slugs in stable order.
func (c *CredentialsFile) SortedWorkspaceSlugs() []string {
	if c == nil {
		return nil
	}
	slugs := make([]string, 0, len(c.Workspaces))
	for slug := range c.Workspaces {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	return slugs
}

func (c *CredentialsFile) ensureDefaultWorkspace() {
	if c == nil {
		return
	}
	if c.Workspaces == nil {
		c.Workspaces = map[string]WorkspaceCredential{}
	}
	if len(c.Workspaces) == 0 {
		c.DefaultWorkspace = ""
		return
	}
	if c.DefaultWorkspace != "" {
		if _, ok := c.Workspaces[c.DefaultWorkspace]; ok {
			return
		}
	}
	c.DefaultWorkspace = c.SortedWorkspaceSlugs()[0]
}

func (c *CredentialsFile) findSlugByWorkspaceID(workspaceID string) (string, bool) {
	if c == nil || workspaceID == "" {
		return "", false
	}
	for slug, meta := range c.Workspaces {
		if meta.WorkspaceID == workspaceID {
			return slug, true
		}
	}
	return "", false
}

// ValidateWorkspaceSlug validates a local workspace credential slug.
func ValidateWorkspaceSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("workspace slug cannot be empty")
	}
	if slug == LegacyWorkspaceSlug {
		return fmt.Errorf("%q is reserved for the legacy default credential", LegacyWorkspaceSlug)
	}
	if !validWorkspaceSlug.MatchString(slug) {
		return fmt.Errorf("workspace slug %q must contain only lowercase letters, numbers, and single hyphen-separated segments", slug)
	}
	return nil
}

// SlugFromWorkspaceName derives a stable local slug from Notion workspace
// metadata.
func SlugFromWorkspaceName(workspaceName, workspaceID string) string {
	var b strings.Builder
	lastHyphen := false
	for _, r := range strings.ToLower(workspaceName) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastHyphen = false
			continue
		}
		if !lastHyphen {
			b.WriteByte('-')
			lastHyphen = true
		}
	}
	slug := strings.Trim(b.String(), "-")
	if slug == "" {
		prefix := workspaceIDPrefix(workspaceID)
		if prefix == "" {
			return ""
		}
		return "notion-" + prefix
	}
	return slug
}

func workspaceIDPrefix(workspaceID string) string {
	clean := strings.ToLower(strings.ReplaceAll(workspaceID, "-", ""))
	if len(clean) > 8 {
		return clean[:8]
	}
	return clean
}

// ChooseWorkspaceSlug returns the slug to use for workspace metadata, handling
// re-auth, explicit overrides, and generated-slug collisions.
func ChooseWorkspaceSlug(creds *CredentialsFile, explicitSlug, workspaceName, workspaceID string) (string, error) {
	explicitSlug = strings.TrimSpace(explicitSlug)
	if explicitSlug != "" {
		if err := ValidateWorkspaceSlug(explicitSlug); err != nil {
			return "", err
		}
		return explicitSlug, nil
	}
	if existing, ok := creds.findSlugByWorkspaceID(workspaceID); ok {
		return existing, nil
	}
	slug := SlugFromWorkspaceName(workspaceName, workspaceID)
	if slug == "" {
		return "", fmt.Errorf("could not derive workspace slug; pass --slug")
	}
	if err := ValidateWorkspaceSlug(slug); err != nil {
		return "", err
	}
	if creds == nil {
		return slug, nil
	}
	if current, exists := creds.Workspaces[slug]; !exists || current.WorkspaceID == "" || current.WorkspaceID == workspaceID {
		return slug, nil
	}
	prefix := workspaceIDPrefix(workspaceID)
	if prefix != "" {
		withPrefix := slug + "-" + prefix
		if current, exists := creds.Workspaces[withPrefix]; !exists || current.WorkspaceID == workspaceID {
			return withPrefix, nil
		}
	}
	for i := 2; i < 1000; i++ {
		candidate := fmt.Sprintf("%s-%d", slug, i)
		if current, exists := creds.Workspaces[candidate]; !exists || current.WorkspaceID == workspaceID {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not derive unique workspace slug for %q after 999 attempts; pass --slug", slug)
}

// ResolveWorkspace resolves a command's active workspace from flag/env/default.
func ResolveWorkspace(explicitSlug string) (*ActiveWorkspace, error) {
	slug := strings.TrimSpace(explicitSlug)
	source := "flag"
	if slug == "" {
		slug = strings.TrimSpace(os.Getenv("NOTION_WORKSPACE"))
		source = "env"
	}
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}
	if slug == "" && creds.DefaultWorkspace != "" {
		slug = creds.DefaultWorkspace
		source = "default"
	}
	if slug == LegacyWorkspaceSlug {
		return &ActiveWorkspace{Slug: "", Source: "legacy", Legacy: true}, nil
	}
	if slug == "" {
		if len(creds.Workspaces) > 0 {
			return nil, fmt.Errorf("no default Notion workspace selected; run 'notion-cli auth default <workspace>' or use --auth-workspace <workspace>")
		}
		return &ActiveWorkspace{Slug: "", Source: "legacy", Legacy: true}, nil
	}
	if err := ValidateWorkspaceSlug(slug); err != nil {
		return nil, err
	}
	meta, ok := creds.Workspaces[slug]
	if !ok {
		return nil, fmt.Errorf("unknown Notion workspace credential %q; run 'notion-cli auth list'", slug)
	}
	return &ActiveWorkspace{Slug: slug, Source: source, Metadata: &meta}, nil
}

// SaveWorkspaceCredential stores metadata and secrets for a named workspace.
func SaveWorkspaceCredential(meta WorkspaceCredential, secrets WorkspaceSecrets, makeDefaultIfEmpty bool) error {
	if err := ValidateWorkspaceSlug(meta.Slug); err != nil {
		return err
	}
	if meta.UpdatedAt == "" {
		meta.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if secrets.OAuthAccessToken != "" {
		if err := secretStore.Set(secretKey(meta.Slug, "oauth_access_token"), secrets.OAuthAccessToken); err != nil {
			return err
		}
	}
	if secrets.OAuthRefreshToken != "" {
		if err := secretStore.Set(secretKey(meta.Slug, "oauth_refresh_token"), secrets.OAuthRefreshToken); err != nil {
			return err
		}
	}
	creds, err := LoadCredentials()
	if err != nil {
		return err
	}
	if meta.WorkspaceID != "" {
		for slug, existing := range creds.Workspaces {
			if slug != meta.Slug && existing.WorkspaceID == meta.WorkspaceID {
				delete(creds.Workspaces, slug)
				_ = DeleteWorkspaceSecrets(slug)
			}
		}
	}
	creds.Workspaces[meta.Slug] = meta
	if makeDefaultIfEmpty && creds.DefaultWorkspace == "" {
		creds.DefaultWorkspace = meta.Slug
	}
	return SaveCredentials(creds)
}

// LoadWorkspaceSecrets reads secrets for a named workspace from the keychain.
func LoadWorkspaceSecrets(slug string) (WorkspaceSecrets, error) {
	var secrets WorkspaceSecrets
	for key, dest := range map[string]*string{
		"oauth_access_token":  &secrets.OAuthAccessToken,
		"oauth_refresh_token": &secrets.OAuthRefreshToken,
	} {
		value, err := secretStore.Get(secretKey(slug, key))
		if err != nil {
			if err == ErrSecretNotFound {
				continue
			}
			return secrets, err
		}
		*dest = value
	}
	return secrets, nil
}

// DeleteWorkspaceSecrets removes all keychain items for a named workspace.
func DeleteWorkspaceSecrets(slug string) error {
	for _, key := range []string{"oauth_access_token", "oauth_refresh_token"} {
		if err := secretStore.Delete(secretKey(slug, key)); err != nil {
			return err
		}
	}
	return nil
}

// DeleteWorkspaceCredential removes metadata and secrets for a named workspace.
func DeleteWorkspaceCredential(slug string) (*WorkspaceCredential, error) {
	if err := ValidateWorkspaceSlug(slug); err != nil {
		return nil, err
	}
	creds, err := LoadCredentials()
	if err != nil {
		return nil, err
	}
	meta, ok := creds.Workspaces[slug]
	if !ok {
		return nil, fmt.Errorf("unknown Notion workspace credential %q", slug)
	}
	delete(creds.Workspaces, slug)
	if creds.DefaultWorkspace == slug {
		creds.DefaultWorkspace = ""
	}
	if err := DeleteWorkspaceSecrets(slug); err != nil {
		return nil, err
	}
	if err := SaveCredentials(creds); err != nil {
		return nil, err
	}
	return &meta, nil
}

// SetDefaultWorkspace sets the default named workspace.
func SetDefaultWorkspace(slug string) error {
	creds, err := LoadCredentials()
	if err != nil {
		return err
	}
	if slug == LegacyWorkspaceSlug {
		if len(creds.Workspaces) > 0 {
			return fmt.Errorf("cannot set legacy config as the default while named workspace credentials exist; use --auth-workspace default for legacy commands")
		}
		creds.DefaultWorkspace = ""
		return SaveCredentials(creds)
	}
	if err := ValidateWorkspaceSlug(slug); err != nil {
		return err
	}
	if _, ok := creds.Workspaces[slug]; !ok {
		return fmt.Errorf("unknown Notion workspace credential %q", slug)
	}
	creds.DefaultWorkspace = slug
	return SaveCredentials(creds)
}

// LoadConfigForWorkspace loads config settings plus credentials for a resolved
// workspace. For named workspaces, non-secret settings still come from the
// legacy config.json and environment variables, but token fields come from the
// selected keychain entry.
func LoadConfigForWorkspace(explicitSlug string) (*Config, *ActiveWorkspace, error) {
	active, err := ResolveWorkspace(explicitSlug)
	if err != nil {
		return nil, nil, err
	}
	cfg := defaults()
	if err := loadFromFile(cfg); err != nil && !os.IsNotExist(err) {
		return nil, nil, err
	}
	if !active.Legacy {
		cfg.Token = ""
		cfg.ClearOAuth()
		if active.Metadata != nil {
			cfg.OAuthWorkspaceID = active.Metadata.WorkspaceID
			cfg.OAuthWorkspaceName = active.Metadata.WorkspaceName
			cfg.OAuthBotID = active.Metadata.BotID
			cfg.OAuthTokenExpiresAt = active.Metadata.OAuthTokenExpiresAt
		}
		secrets, secErr := LoadWorkspaceSecrets(active.Slug)
		if secErr != nil {
			return nil, nil, secErr
		}
		switch active.Metadata.AuthMethod {
		case AuthMethodOAuth:
			cfg.OAuthAccessToken = secrets.OAuthAccessToken
			cfg.OAuthRefreshToken = secrets.OAuthRefreshToken
		}
	}
	loadFromEnv(cfg)
	return cfg, active, nil
}

// SaveConfigForWorkspace persists token fields back to the selected credential
// location. Named workspaces update metadata plus keychain secrets; the legacy
// default writes config.json.
func SaveConfigForWorkspace(active *ActiveWorkspace, cfg *Config) error {
	if active == nil || active.Legacy || active.Slug == "" {
		return SaveConfig(cfg)
	}
	meta := WorkspaceCredential{Slug: active.Slug}
	if active.Metadata != nil {
		meta = *active.Metadata
	}
	meta.Slug = active.Slug
	meta.WorkspaceID = FirstNonEmpty(cfg.OAuthWorkspaceID, meta.WorkspaceID)
	meta.WorkspaceName = FirstNonEmpty(cfg.OAuthWorkspaceName, meta.WorkspaceName)
	meta.BotID = FirstNonEmpty(cfg.OAuthBotID, meta.BotID)
	meta.OAuthTokenExpiresAt = cfg.OAuthTokenExpiresAt
	secrets := WorkspaceSecrets{}
	if cfg.OAuthAccessToken != "" {
		meta.AuthMethod = AuthMethodOAuth
		secrets.OAuthAccessToken = cfg.OAuthAccessToken
		secrets.OAuthRefreshToken = cfg.OAuthRefreshToken
	}
	return SaveWorkspaceCredential(meta, secrets, false)
}

// FirstNonEmpty returns the first non-empty string from values.
func FirstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
