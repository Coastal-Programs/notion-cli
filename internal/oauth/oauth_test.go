package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/internal/errors"
)

func TestRandomState(t *testing.T) {
	s1, err := randomState()
	if err != nil {
		t.Fatalf("randomState() error: %v", err)
	}
	if len(s1) != 32 {
		t.Errorf("randomState() length = %d, want 32", len(s1))
	}

	s2, err := randomState()
	if err != nil {
		t.Fatalf("randomState() error: %v", err)
	}
	if s1 == s2 {
		t.Error("randomState() returned same value twice")
	}
}

func TestCallbackHTML(t *testing.T) {
	html := callbackHTML("Test Title", "Test message")
	if html == "" {
		t.Error("callbackHTML() returned empty string")
	}
	if len(html) < 50 {
		t.Error("callbackHTML() returned unexpectedly short HTML")
	}
}

func TestExchangeCode_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Basic auth is present.
		user, pass, ok := r.BasicAuth()
		if !ok || user != "test-client-id" || pass != "test-client-secret" {
			t.Errorf("Expected Basic auth with test credentials, got user=%q pass=%q ok=%v", user, pass, ok)
		}

		// Verify content type.
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		resp := TokenResponse{
			AccessToken:   "ntn_test_token_abc123",
			TokenType:     "bearer",
			BotID:         "bot-123",
			WorkspaceID:   "ws-456",
			WorkspaceName: "Test Workspace",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Override tokenURL for test.
	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	token, err := exchangeCode(context.Background(), "test-client-id", "test-client-secret", "auth-code-123")
	if err != nil {
		t.Fatalf("exchangeCode() error: %v", err)
	}

	if token.AccessToken != "ntn_test_token_abc123" {
		t.Errorf("AccessToken = %q, want %q", token.AccessToken, "ntn_test_token_abc123")
	}
	if token.WorkspaceName != "Test Workspace" {
		t.Errorf("WorkspaceName = %q, want %q", token.WorkspaceName, "Test Workspace")
	}
	if token.BotID != "bot-123" {
		t.Errorf("BotID = %q, want %q", token.BotID, "bot-123")
	}
}

func TestExchangeCode_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
	}))
	defer server.Close()

	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	_, err := exchangeCode(context.Background(), "id", "secret", "bad-code")
	if err == nil {
		t.Fatal("exchangeCode() should return error for HTTP 400")
	}

	cliErr, ok := err.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("expected NotionCLIError, got %T", err)
	}
	if cliErr.Code != clierrors.CodeOAuthFailed {
		t.Errorf("error code = %q, want %q", cliErr.Code, clierrors.CodeOAuthFailed)
	}
}

func TestExchangeCode_EmptyAccessToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"access_token": ""})
	}))
	defer server.Close()

	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	_, err := exchangeCode(context.Background(), "id", "secret", "code")
	if err == nil {
		t.Fatal("exchangeCode() should return error for empty access token")
	}
}

func TestExchangeCode_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, "not json")
	}))
	defer server.Close()

	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	_, err := exchangeCode(context.Background(), "id", "secret", "code")
	if err == nil {
		t.Fatal("exchangeCode() should return error for invalid JSON")
	}
}

func TestLogin_EmptyCredentials(t *testing.T) {
	_, err := Login(context.Background(), "", "")
	if err == nil {
		t.Fatal("Login() should return error for empty credentials")
	}

	cliErr, ok := err.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("expected NotionCLIError, got %T", err)
	}
	if cliErr.Code != clierrors.CodeOAuthNotConfigured {
		t.Errorf("error code = %q, want %q", cliErr.Code, clierrors.CodeOAuthNotConfigured)
	}
}

func TestLogin_ContextCancelled(t *testing.T) {
	// Use a very short timeout so the test doesn't block.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Use a non-standard port that's likely free but won't have a browser connecting.
	// We test with real credentials to exercise the server startup path,
	// but the context will time out before a browser could connect.
	_, err := Login(ctx, "test-id", "test-secret")
	if err == nil {
		t.Fatal("Login() should return error when context times out")
	}

	cliErr, ok := err.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("expected NotionCLIError, got %T", err)
	}
	if cliErr.Code != clierrors.CodeOAuthTimeout {
		t.Errorf("error code = %q, want %q", cliErr.Code, clierrors.CodeOAuthTimeout)
	}
}

func TestExchangeCode_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay longer than the context timeout.
		time.Sleep(500 * time.Millisecond)
		json.NewEncoder(w).Encode(map[string]string{"access_token": "tok"})
	}))
	defer server.Close()

	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := exchangeCode(ctx, "id", "secret", "code")
	if err == nil {
		t.Fatal("exchangeCode() should return error when context is cancelled")
	}
}
