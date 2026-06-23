package oauth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
)

func TestParseCallbackInput(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCode  string
		wantState string
		wantErr   bool
	}{
		{"full URL with code+state", "http://localhost:8080/callback?code=abc&state=xyz", "abc", "xyz", false},
		{"URL with code only", "http://localhost:8081/callback?code=abc", "abc", "", false},
		{"https URL", "https://localhost:8080/callback?code=def&state=foo", "def", "foo", false},
		{"bare code", "raw-auth-code-123", "raw-auth-code-123", "", false},
		{"URL with whitespace", "  http://localhost:8080/callback?code=trim  \n", "trim", "", false},
		{"URL with error param", "http://localhost:8080/callback?error=invalid_request", "", "", true},
		{"URL with no code", "http://localhost:8080/callback?state=xyz", "", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, state, err := parseCallbackInput(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if code != tt.wantCode {
				t.Errorf("code = %q, want %q", code, tt.wantCode)
			}
			if state != tt.wantState {
				t.Errorf("state = %q, want %q", state, tt.wantState)
			}
		})
	}
}

func TestBindCallbackListener_FallsBackWhenFirstPortBusy(t *testing.T) {
	orig := CallbackPorts
	defer func() { CallbackPorts = orig }()

	// Pick two free ports, then occupy the first to force fallback.
	probe1, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("probe1 listen: %v", err)
	}
	port1 := probe1.Addr().(*net.TCPAddr).Port

	probe2, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		_ = probe1.Close()
		t.Fatalf("probe2 listen: %v", err)
	}
	port2 := probe2.Addr().(*net.TCPAddr).Port
	_ = probe2.Close() // free port2 so bindCallbackListener can take it

	CallbackPorts = []int{port1, port2}

	ln, gotPort, err := bindCallbackListener()
	if err != nil {
		_ = probe1.Close()
		t.Fatalf("bindCallbackListener err: %v", err)
	}
	defer ln.Close() //nolint:errcheck
	_ = probe1.Close()

	if gotPort != port2 {
		t.Errorf("got port %d, want %d (the fallback)", gotPort, port2)
	}
}

func TestBindCallbackListener_AllBusy(t *testing.T) {
	orig := CallbackPorts
	defer func() { CallbackPorts = orig }()

	probe, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("probe listen: %v", err)
	}
	defer probe.Close() //nolint:errcheck
	port := probe.Addr().(*net.TCPAddr).Port

	CallbackPorts = []int{port}

	_, _, err = bindCallbackListener()
	if err == nil {
		t.Fatal("expected error when all ports busy")
	}
	cliErr, ok := err.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("want NotionCLIError, got %T", err)
	}
	if cliErr.Code != clierrors.CodeOAuthPortInUse {
		t.Errorf("code = %q, want %q", cliErr.Code, clierrors.CodeOAuthPortInUse)
	}
}

func TestBuildAuthorizeURL(t *testing.T) {
	u := buildAuthorizeURL("client-1", "http://localhost:8081/callback", "state-x")
	for _, want := range []string{
		"client_id=client-1",
		"redirect_uri=http%3A%2F%2Flocalhost%3A8081%2Fcallback",
		"state=state-x",
		"response_type=code",
		"owner=user",
	} {
		if !contains(u, want) {
			t.Errorf("authorize URL missing %q\ngot: %s", want, u)
		}
	}
}

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
	for _, success := range []bool{true, false} {
		html := callbackHTML(success, "Test Title", "Test message")
		if html == "" {
			t.Errorf("callbackHTML(success=%v) returned empty string", success)
		}
		if len(html) < 200 {
			t.Errorf("callbackHTML(success=%v) returned unexpectedly short HTML", success)
		}
		if !contains(html, "Test Title") || !contains(html, "Test message") {
			t.Errorf("callbackHTML(success=%v) missing title or message", success)
		}
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
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

		// Verify Accept header (defensive — some OAuth servers content-negotiate).
		if a := r.Header.Get("Accept"); a != "application/json" {
			t.Errorf("Accept = %q, want application/json", a)
		}

		// Verify Notion-Version header (matches official notion-sdk-js behavior).
		if v := r.Header.Get("Notion-Version"); v != notionVersion {
			t.Errorf("Notion-Version = %q, want %q", v, notionVersion)
		}

		resp := TokenResponse{
			AccessToken:   "ntn_test_token_abc123",
			TokenType:     "bearer",
			BotID:         "bot-123",
			WorkspaceID:   "ws-456",
			WorkspaceName: "Test Workspace",
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	// Override tokenURL for test.
	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	token, err := exchangeCode(context.Background(), "test-client-id", "test-client-secret", "http://localhost:8080/callback", "auth-code-123")
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
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
	}))
	defer server.Close()

	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	_, err := exchangeCode(context.Background(), "id", "secret", "http://localhost:8080/callback", "bad-code")
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
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": ""})
	}))
	defer server.Close()

	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	_, err := exchangeCode(context.Background(), "id", "secret", "http://localhost:8080/callback", "code")
	if err == nil {
		t.Fatal("exchangeCode() should return error for empty access token")
	}
}

func TestExchangeCode_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, "not json")
	}))
	defer server.Close()

	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	_, err := exchangeCode(context.Background(), "id", "secret", "http://localhost:8080/callback", "code")
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

// assertOAuthHeaders fails the test if the request is missing any of the
// three headers we send on every OAuth endpoint call: Content-Type,
// Accept, and Notion-Version.
func assertOAuthHeaders(t *testing.T, r *http.Request) {
	t.Helper()
	if ct := r.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", ct)
	}
	if a := r.Header.Get("Accept"); a != "application/json" {
		t.Errorf("Accept = %q, want application/json", a)
	}
	if v := r.Header.Get("Notion-Version"); v != notionVersion {
		t.Errorf("Notion-Version = %q, want %q", v, notionVersion)
	}
}

func TestTokenRefresh_SendsHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertOAuthHeaders(t, r)
		_ = json.NewEncoder(w).Encode(TokenResponse{AccessToken: "new-token"})
	}))
	defer server.Close()

	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	_, err := TokenRefresh(context.Background(), "id", "secret", "refresh-tok")
	if err != nil {
		t.Fatalf("TokenRefresh() error: %v", err)
	}
}

func TestTokenIntrospect_SendsHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertOAuthHeaders(t, r)
		_ = json.NewEncoder(w).Encode(IntrospectResponse{Active: true})
	}))
	defer server.Close()

	origURL := introspectURL
	introspectURL = server.URL
	defer func() { introspectURL = origURL }()

	_, err := TokenIntrospect(context.Background(), "id", "secret", "tok")
	if err != nil {
		t.Fatalf("TokenIntrospect() error: %v", err)
	}
}

func TestTokenRevoke_SendsHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertOAuthHeaders(t, r)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	origURL := revokeURL
	revokeURL = server.URL
	defer func() { revokeURL = origURL }()

	if err := TokenRevoke(context.Background(), "id", "secret", "tok"); err != nil {
		t.Fatalf("TokenRevoke() error: %v", err)
	}
}

// TestLoginManual_RejectsBareCode verifies that pasting only a bare
// authorization code (no state) into the manual flow is rejected with an
// OAuth state-mismatch error and that the token endpoint is never called.
func TestLoginManual_RejectsBareCode(t *testing.T) {
	tokenServerHit := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenServerHit = true
		_ = json.NewEncoder(w).Encode(TokenResponse{AccessToken: "should-not-be-returned"})
	}))
	defer server.Close()

	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	in := strings.NewReader("bare-auth-code-xyz\n")
	var out bytes.Buffer

	_, err := LoginManual(context.Background(), "client-id", "client-secret", in, &out)
	if err == nil {
		t.Fatal("LoginManual should reject a bare authorization code")
	}

	cliErr, ok := err.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("expected NotionCLIError, got %T: %v", err, err)
	}
	if cliErr.Code != clierrors.CodeOAuthStateMismatch {
		t.Errorf("error code = %q, want %q", cliErr.Code, clierrors.CodeOAuthStateMismatch)
	}
	// The bare-code path must surface a message that points the user at the
	// real mistake (pasting just the code), not the generic CSRF text.
	if !strings.Contains(cliErr.Message, "missing") {
		t.Errorf("error message = %q, want it to mention %q", cliErr.Message, "missing")
	}
	foundFullHint := false
	for _, s := range cliErr.Suggestions {
		if strings.Contains(s, "FULL") {
			foundFullHint = true
			break
		}
	}
	if !foundFullHint {
		t.Errorf("expected a suggestion mentioning %q, got %v", "FULL", cliErr.Suggestions)
	}
	if tokenServerHit {
		t.Error("token endpoint was called \u2014 the bare code should have been rejected before exchangeCode")
	}
}

// TestLoginManual_RejectsMismatchedState verifies that pasting a redirected
// URL whose state does not match the generated state is rejected.
func TestLoginManual_RejectsMismatchedState(t *testing.T) {
	tokenServerHit := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenServerHit = true
		_ = json.NewEncoder(w).Encode(TokenResponse{AccessToken: "should-not-be-returned"})
	}))
	defer server.Close()

	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	// Attacker-supplied state that won't match the freshly generated one.
	in := strings.NewReader("http://localhost:8080/callback?code=abc&state=attacker-state\n")
	var out bytes.Buffer

	_, err := LoginManual(context.Background(), "client-id", "client-secret", in, &out)
	if err == nil {
		t.Fatal("LoginManual should reject a mismatched state")
	}
	cliErr, ok := err.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("expected NotionCLIError, got %T: %v", err, err)
	}
	if cliErr.Code != clierrors.CodeOAuthStateMismatch {
		t.Errorf("error code = %q, want %q", cliErr.Code, clierrors.CodeOAuthStateMismatch)
	}
	// The real-mismatch path must keep using the original CSRF message —
	// the bare-code diagnostic must not bleed into this branch.
	if !strings.Contains(cliErr.Message, "CSRF") {
		t.Errorf("error message = %q, want it to mention %q", cliErr.Message, "CSRF")
	}
	if tokenServerHit {
		t.Error("token endpoint was called \u2014 mismatched state should have been rejected")
	}
}

func TestExchangeCode_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay longer than the context timeout.
		time.Sleep(500 * time.Millisecond)
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "tok"})
	}))
	defer server.Close()

	origURL := tokenURL
	tokenURL = server.URL
	defer func() { tokenURL = origURL }()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := exchangeCode(ctx, "id", "secret", "http://localhost:8080/callback", "code")
	if err == nil {
		t.Fatal("exchangeCode() should return error when context is cancelled")
	}
}

// TestCallback_RejectsNonGet verifies the OAuth callback handler returns 405
// for any method other than GET, sets Allow: GET, and does not enqueue a
// callback result. Without this guard a CSRF-like POST from a malicious origin
// could be processed.
func TestCallback_RejectsNonGet(t *testing.T) {
	resultCh := make(chan callbackResult, 1)
	h := newCallbackHandler(8080, resultCh)

	req := httptest.NewRequest(http.MethodPost, "/callback?code=xyz&state=expected", nil)
	req.Host = "localhost:8080"
	w := httptest.NewRecorder()

	h(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
	if got := w.Header().Get("Allow"); got != "GET" {
		t.Errorf("Allow header = %q, want %q", got, "GET")
	}
	select {
	case r := <-resultCh:
		t.Errorf("unexpected callback result sent: %+v", r)
	default:
	}
}

// TestCallback_RejectsMismatchedHost verifies the DNS-rebinding defense:
// requests whose Host header is neither localhost:PORT nor 127.0.0.1:PORT must
// receive 421 Misdirected Request and never reach the result channel.
func TestCallback_RejectsMismatchedHost(t *testing.T) {
	resultCh := make(chan callbackResult, 1)
	h := newCallbackHandler(8080, resultCh)

	req := httptest.NewRequest(http.MethodGet, "/callback?code=xyz&state=expected", nil)
	req.Host = "evil.example.com:8080"
	w := httptest.NewRecorder()

	h(w, req)

	if w.Code != http.StatusMisdirectedRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusMisdirectedRequest)
	}
	select {
	case r := <-resultCh:
		t.Errorf("unexpected callback result sent: %+v", r)
	default:
	}
}

// TestCallback_AcceptsValidHosts verifies the handler accepts the two Host
// values Go's HTTP server can present for a listener bound to 127.0.0.1 with
// a redirect URI of http://localhost:PORT/callback: literal "localhost:PORT"
// (what the browser sends after following the Notion redirect) and the IPv4
// form "127.0.0.1:PORT". Both must enqueue the code and state on resultCh.
func TestCallback_AcceptsValidHosts(t *testing.T) {
	tests := []struct {
		name string
		host string
	}{
		{"localhost form", "localhost:8080"},
		{"ipv4 form", "127.0.0.1:8080"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultCh := make(chan callbackResult, 1)
			h := newCallbackHandler(8080, resultCh)

			req := httptest.NewRequest(http.MethodGet, "/callback?code=xyz&state=expected", nil)
			req.Host = tt.host
			w := httptest.NewRecorder()

			h(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
			}
			select {
			case r := <-resultCh:
				if r.code != "xyz" {
					t.Errorf("result.code = %q, want %q", r.code, "xyz")
				}
				if r.state != "expected" {
					t.Errorf("result.state = %q, want %q", r.state, "expected")
				}
				if r.err != "" {
					t.Errorf("result.err = %q, want empty", r.err)
				}
			default:
				t.Error("expected callback result on channel, got none")
			}
		})
	}
}
