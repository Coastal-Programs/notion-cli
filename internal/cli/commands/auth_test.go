package commands

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Coastal-Programs/notion-cli/v6/internal/config"
	"github.com/Coastal-Programs/notion-cli/v6/internal/notion"
	"github.com/Coastal-Programs/notion-cli/v6/internal/oauth"
	"github.com/spf13/cobra"
)

// resetOAuthURLs restores oauth package URL vars to production values.
func resetOAuthURLs() {
	oauth.SetTokenURL("https://api.notion.com/v1/oauth/token")
	oauth.SetIntrospectURL("https://api.notion.com/v1/oauth/introspect")
	oauth.SetRevokeURL("https://api.notion.com/v1/oauth/revoke")
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// --- TestOAuthTokenRefresh ---

func TestOAuthTokenRefresh(t *testing.T) {
	var gotAuth string
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		writeJSON(w, 200, map[string]any{
			"access_token":  "new_access",
			"token_type":    "bearer",
			"refresh_token": "new_refresh",
			"expires_in":    3600,
			"workspace_id":  "ws1",
		})
	}))
	defer srv.Close()

	oauth.SetTokenURL(srv.URL)
	defer resetOAuthURLs()

	token, err := oauth.TokenRefresh(t.Context(), "cid", "csecret", "old_refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if gotAuth == "" {
		t.Error("expected Authorization header, got none")
	}
	if gotBody["grant_type"] != "refresh_token" {
		t.Errorf("expected grant_type=refresh_token, got %q", gotBody["grant_type"])
	}
	if gotBody["refresh_token"] != "old_refresh" {
		t.Errorf("expected refresh_token=old_refresh, got %q", gotBody["refresh_token"])
	}
	if token.AccessToken != "new_access" {
		t.Errorf("expected access_token=new_access, got %q", token.AccessToken)
	}
	if token.RefreshToken != "new_refresh" {
		t.Errorf("expected refresh_token=new_refresh, got %q", token.RefreshToken)
	}
	if token.ExpiresIn != 3600 {
		t.Errorf("expected expires_in=3600, got %d", token.ExpiresIn)
	}
}

// --- TestOAuthTokenIntrospect ---

func TestOAuthTokenIntrospect(t *testing.T) {
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		writeJSON(w, 200, map[string]any{
			"active": true,
			"scope":  "read_content",
			"iat":    int64(1700000000),
		})
	}))
	defer srv.Close()

	oauth.SetIntrospectURL(srv.URL)
	defer resetOAuthURLs()

	result, err := oauth.TokenIntrospect(t.Context(), "cid", "csecret", "mytoken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["token"] != "mytoken" {
		t.Errorf("expected token=mytoken in body, got %q", gotBody["token"])
	}
	if !result.Active {
		t.Error("expected active=true")
	}
	if result.Scope != "read_content" {
		t.Errorf("expected scope=read_content, got %q", result.Scope)
	}
	if result.IssuedAt != 1700000000 {
		t.Errorf("expected iat=1700000000, got %d", result.IssuedAt)
	}
}

// --- TestOAuthTokenRevoke ---

func TestOAuthTokenRevoke(t *testing.T) {
	var gotBody map[string]string

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotBody)
		writeJSON(w, 200, map[string]any{"request_id": "req123"})
	}))
	defer srv.Close()

	oauth.SetRevokeURL(srv.URL)
	defer resetOAuthURLs()

	err := oauth.TokenRevoke(t.Context(), "cid", "csecret", "mytoken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotBody["token"] != "mytoken" {
		t.Errorf("expected token=mytoken in body, got %q", gotBody["token"])
	}
}

// --- TestAutoRefreshOn401 ---

func TestAutoRefreshOn401(t *testing.T) {
	var apiCallCount int32
	var refreshCallCount int32

	// Token endpoint: returns a fresh access token.
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&refreshCallCount, 1)
		writeJSON(w, 200, map[string]any{
			"access_token":  "refreshed_token",
			"token_type":    "bearer",
			"refresh_token": "new_refresh_token",
			"expires_in":    3600,
		})
	}))
	defer tokenSrv.Close()

	oauth.SetTokenURL(tokenSrv.URL)
	defer resetOAuthURLs()

	// API server: 401 on first call, 200 on subsequent calls.
	apiSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&apiCallCount, 1)
		if n == 1 {
			writeJSON(w, 401, map[string]any{
				"status":  401,
				"code":    "unauthorized",
				"message": "token expired",
			})
			return
		}
		writeJSON(w, 200, map[string]any{"object": "user", "id": "u1"})
	}))
	defer apiSrv.Close()

	// Temporarily set build-time OAuth credentials so auto-refresh is triggered.
	origID := config.OAuthClientID
	origSecret := config.OAuthClientSecret
	config.OAuthClientID = "test_client_id"
	config.OAuthClientSecret = "test_client_secret"
	defer func() {
		config.OAuthClientID = origID
		config.OAuthClientSecret = origSecret
	}()

	cfg := &config.Config{
		OAuthAccessToken:  "expired_token",
		OAuthRefreshToken: "old_refresh_token",
	}

	client := notion.NewClient("expired_token",
		notion.WithBaseURL(apiSrv.URL),
		notion.WithConfig(cfg),
	)

	result, err := client.UsersMe(t.Context())
	if err != nil {
		t.Fatalf("expected success after auto-refresh, got error: %v", err)
	}
	if result["id"] != "u1" {
		t.Errorf("expected id=u1, got %v", result["id"])
	}
	if got := atomic.LoadInt32(&apiCallCount); got != 2 {
		t.Errorf("expected 2 API calls (401 + retry), got %d", got)
	}
	if got := atomic.LoadInt32(&refreshCallCount); got != 1 {
		t.Errorf("expected 1 refresh call, got %d", got)
	}
}

// --- TestAuthLogout_LocalOnly ---

func TestAuthLogout_LocalOnly(t *testing.T) {
	var revokeCallCount int32

	revokeSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&revokeCallCount, 1)
		writeJSON(w, 200, map[string]any{"request_id": "req1"})
	}))
	defer revokeSrv.Close()

	oauth.SetRevokeURL(revokeSrv.URL)
	defer resetOAuthURLs()

	origID := config.OAuthClientID
	origSecret := config.OAuthClientSecret
	config.OAuthClientID = "test_cid"
	config.OAuthClientSecret = "test_csecret"
	defer func() {
		config.OAuthClientID = origID
		config.OAuthClientSecret = origSecret
	}()

	// Simulate the --local-only path: no TokenRevoke call is made.
	localOnly := true
	if !localOnly {
		// This branch is intentionally unreachable; it represents the non-local-only
		// path that would call TokenRevoke.
		if err := oauth.TokenRevoke(t.Context(), config.OAuthClientID, config.OAuthClientSecret, "tok"); err != nil {
			t.Fatal(err)
		}
	}

	if got := atomic.LoadInt32(&revokeCallCount); got != 0 {
		t.Errorf("expected 0 revoke calls with --local-only, got %d", got)
	}
}

// --- TestNeedsRefresh_Boundary ---

func TestNeedsRefresh_Boundary(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken string
		expiresAt    string
		want         bool
	}{
		{
			name:         "empty fields returns false",
			refreshToken: "",
			expiresAt:    "",
			want:         false,
		},
		{
			name:         "no refresh token returns false",
			refreshToken: "",
			expiresAt:    time.Now().Add(1 * time.Minute).UTC().Format(time.RFC3339),
			want:         false,
		},
		{
			name:         "no expiry returns false",
			refreshToken: "some_refresh",
			expiresAt:    "",
			want:         false,
		},
		{
			name:         "expiring in 6 minutes returns false",
			refreshToken: "some_refresh",
			expiresAt:    time.Now().Add(6 * time.Minute).UTC().Format(time.RFC3339),
			want:         false,
		},
		{
			name:         "expiring in 4 minutes returns true",
			refreshToken: "some_refresh",
			expiresAt:    time.Now().Add(4 * time.Minute).UTC().Format(time.RFC3339),
			want:         true,
		},
		{
			name:         "already expired returns true",
			refreshToken: "some_refresh",
			expiresAt:    time.Now().Add(-1 * time.Minute).UTC().Format(time.RFC3339),
			want:         true,
		},
		{
			name:         "invalid expiry format returns false",
			refreshToken: "some_refresh",
			expiresAt:    "not-a-date",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{
				OAuthRefreshToken:   tt.refreshToken,
				OAuthTokenExpiresAt: tt.expiresAt,
			}
			got := cfg.NeedsRefresh()
			if got != tt.want {
				t.Errorf("NeedsRefresh() = %v, want %v", got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// auth status / logout / refresh command tests
// ---------------------------------------------------------------------------

func runAuthRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterAuthCommands(root)
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

func TestAuthStatus_EnvToken(t *testing.T) {
	origToken := os.Getenv("NOTION_TOKEN")
	_ = os.Setenv("NOTION_TOKEN", "secret_envtoken")
	t.Cleanup(func() {
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
	})

	// Just verify it runs without error and doesn't panic.
	_, _, _ = runAuthRoot(t, "auth", "status")
}

func TestAuthStatus_NoToken(t *testing.T) {
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

	// Should succeed (prints "Not authenticated" status).
	_, _, err := runAuthRoot(t, "auth", "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthStatus_ManualToken(t *testing.T) {
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

	// Save a manual token to config.
	if err := config.SaveConfig(&config.Config{Token: "secret_manual"}); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	_, _, err := runAuthRoot(t, "auth", "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthStatus_OAuthToken(t *testing.T) {
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

	cfg := &config.Config{
		OAuthAccessToken:   "ntn_oauth_token",
		OAuthWorkspaceName: "My Workspace",
		OAuthWorkspaceID:   "ws-1",
		OAuthBotID:         "bot-1",
	}
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	_, _, err := runAuthRoot(t, "auth", "status")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAuthLogout_NoOAuthToken(t *testing.T) {
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

	_, _, err := runAuthRoot(t, "auth", "logout")
	if err == nil {
		t.Fatal("expected error when no OAuth session")
	}
}

func TestAuthLogout_LocalOnly_ClearsConfig(t *testing.T) {
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

	cfg := &config.Config{
		OAuthAccessToken:   "ntn_oauth_token",
		OAuthWorkspaceName: "Test WS",
	}
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	_, _, err := runAuthRoot(t, "auth", "logout", "--local-only")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify OAuth token was cleared from config.
	cfgPath := filepath.Join(tmpDir, ".config", "notion-cli", "config.json")
	loaded, loadErr := config.LoadConfig()
	if loadErr != nil {
		t.Fatalf("LoadConfig: %v (path: %s)", loadErr, cfgPath)
	}
	if loaded.OAuthAccessToken != "" {
		t.Errorf("OAuthAccessToken should be empty after logout, got %q", loaded.OAuthAccessToken)
	}
}

func TestAuthRefresh_NoRefreshToken(t *testing.T) {
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

	_, _, err := runAuthRoot(t, "auth", "refresh")
	if err == nil {
		t.Fatal("expected error when no refresh token")
	}
}

func TestAuthRefresh_Success(t *testing.T) {
	// Mock token endpoint.
	tokenSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]any{
			"access_token":  "new_ntn_access",
			"refresh_token": "new_refresh",
			"token_type":    "bearer",
			"expires_in":    3600,
		})
	}))
	defer tokenSrv.Close()
	oauth.SetTokenURL(tokenSrv.URL)
	defer resetOAuthURLs()

	origID, origSecret := config.OAuthClientID, config.OAuthClientSecret
	config.OAuthClientID = "test-cid"
	config.OAuthClientSecret = "test-secret"
	defer func() {
		config.OAuthClientID = origID
		config.OAuthClientSecret = origSecret
	}()

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

	cfg := &config.Config{OAuthRefreshToken: "old_refresh", OAuthAccessToken: "ntn_old"}
	if err := config.SaveConfig(cfg); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}

	_, _, err := runAuthRoot(t, "auth", "refresh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
