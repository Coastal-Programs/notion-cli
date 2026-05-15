// Package oauth implements the Notion OAuth 2.0 authorization flow for the CLI.
// It starts a local HTTP server, opens the user's browser to the Notion
// authorization page, and exchanges the received code for an access token.
package oauth

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
)

// CallbackPorts is the ordered list of localhost ports the CLI will attempt
// to bind for the OAuth callback. Each must be registered as a redirect URI
// in the Notion integration's "OAuth Domain & URIs" settings exactly as
// http://localhost:PORT/callback. The CLI tries them in order and uses the
// first one it can bind. This avoids the most common failure mode of
// localhost OAuth flows: another process already holding the default port.
var CallbackPorts = []int{8080, 8081, 8089}

const (
	// authorizeURL is the Notion OAuth authorization endpoint.
	authorizeURL = "https://api.notion.com/v1/oauth/authorize"

	// maxResponseBody limits the token exchange response to 1 MB.
	maxResponseBody = 1 << 20

	// notionVersion is the Notion API version sent on OAuth requests.
	// Matches the default in internal/notion/client.go and the official
	// makenotion/notion-sdk-js, which sends Notion-Version on every request
	// including OAuth endpoints. Keep in sync with defaultNotionVersion.
	notionVersion = "2026-03-11"
)

// redirectURIFor returns the canonical redirect URI for the given port. It
// must match what is registered in the Notion integration settings.
func redirectURIFor(port int) string {
	return fmt.Sprintf("http://localhost:%d/callback", port)
}

// bindCallbackListener tries each port in CallbackPorts and returns the
// first listener it can open along with the bound port. If none are
// available it returns an OAuthPortInUse error listing all attempted ports.
func bindCallbackListener() (net.Listener, int, error) {
	for _, port := range CallbackPorts {
		ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err == nil {
			return ln, port, nil
		}
	}
	return nil, 0, clierrors.OAuthPortInUse(CallbackPorts)
}

// tokenURL is the Notion OAuth token exchange endpoint. It is a var so tests
// can override it with an httptest server URL.
var tokenURL = "https://api.notion.com/v1/oauth/token"

// introspectURL is the Notion OAuth token introspect endpoint. It is a var so
// tests can override it with an httptest server URL.
var introspectURL = "https://api.notion.com/v1/oauth/introspect"

// revokeURL is the Notion OAuth token revoke endpoint. It is a var so tests
// can override it with an httptest server URL.
var revokeURL = "https://api.notion.com/v1/oauth/revoke"

// TokenResponse holds the result of a successful OAuth token exchange or refresh.
type TokenResponse struct {
	AccessToken   string `json:"access_token"`
	TokenType     string `json:"token_type"`
	RefreshToken  string `json:"refresh_token,omitempty"`
	ExpiresIn     int    `json:"expires_in,omitempty"` // seconds
	BotID         string `json:"bot_id"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	WorkspaceIcon string `json:"workspace_icon"`
}

// IntrospectResponse holds the result of a token introspect call.
type IntrospectResponse struct {
	Active   bool   `json:"active"`
	Scope    string `json:"scope,omitempty"`
	IssuedAt int64  `json:"iat,omitempty"` // Unix seconds
}

// callbackResult is sent from the HTTP handler back to the Login goroutine.
type callbackResult struct {
	code  string
	state string
	err   string
}

// newCallbackHandler returns the HTTP handler for the OAuth /callback path.
// It is extracted from Login so its DNS-rebinding and method guards can be
// exercised in unit tests without spinning up the full login flow.
//
// The handler accepts only GET, and only requests whose Host header matches
// localhost:port or 127.0.0.1:port — Go's HTTP server will accept connections
// to 127.0.0.1 on the same listener bound for "localhost", and on dual-stack
// systems "localhost" can resolve to ::1. Anything else is rejected to defeat
// DNS-rebinding attacks even though the state parameter already guards CSRF.
//
// Successful and error results are delivered on resultCh with a non-blocking
// send so a slow consumer (or an unexpected duplicate request) cannot wedge
// the handler.
func newCallbackHandler(port int, resultCh chan<- callbackResult) http.HandlerFunc {
	wantHostLocalhost := fmt.Sprintf("localhost:%d", port)
	wantHostIPv4 := fmt.Sprintf("127.0.0.1:%d", port)
	return func(w http.ResponseWriter, r *http.Request) {
		// Only GET is valid on the OAuth callback.
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", "GET")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Reject mismatched Host headers (DNS rebinding defense).
		if r.Host != wantHostLocalhost && r.Host != wantHostIPv4 {
			http.Error(w, "misdirected request", http.StatusMisdirectedRequest)
			return
		}

		q := r.URL.Query()

		if errParam := q.Get("error"); errParam != "" {
			select {
			case resultCh <- callbackResult{err: errParam}:
			default:
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = fmt.Fprint(w, callbackHTML(false, "Authorization denied", "You denied the authorization request. You can safely close this tab."))
			return
		}

		select {
		case resultCh <- callbackResult{
			code:  q.Get("code"),
			state: q.Get("state"),
		}:
		default:
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, callbackHTML(true, "You're all set", "Authorization complete. Return to your terminal — notion-cli is ready to use."))
	}
}

// Login performs the full OAuth authorization flow:
//  1. Generate a random state parameter
//  2. Start a localhost HTTP server on callbackPort
//  3. Open the user's browser to the Notion authorization URL
//  4. Wait for the callback with the authorization code
//  5. Exchange the code for an access token
//  6. Return the token response
//
// The context controls the overall timeout. If the context is cancelled
// (e.g. Ctrl+C), the server is shut down and an error is returned.
func Login(ctx context.Context, clientID, clientSecret string) (*TokenResponse, error) {
	if clientID == "" || clientSecret == "" {
		return nil, clierrors.OAuthNotConfigured()
	}

	// Generate random state for CSRF protection.
	state, err := randomState()
	if err != nil {
		return nil, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: "Failed to generate random state",
			Err:     err,
		}
	}

	// Bind to localhost only to prevent other users on shared machines from
	// intercepting the OAuth callback. Try each registered port in order so a
	// busy 8080 (Jenkins, Tomcat, dev server) doesn't break the flow.
	ln, port, err := bindCallbackListener()
	if err != nil {
		return nil, err
	}
	redirectURI := redirectURIFor(port)

	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", newCallbackHandler(port, resultCh))

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server on the already-bound listener.
	go func() {
		if sErr := server.Serve(ln); sErr != nil && sErr != http.ErrServerClosed {
			select {
			case resultCh <- callbackResult{err: sErr.Error()}:
			default:
			}
		}
	}()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer server.Shutdown(shutdownCtx) //nolint:errcheck

	// Build and open the authorization URL with properly encoded parameters.
	authURL := buildAuthorizeURL(clientID, redirectURI, state)

	if err := openBrowser(authURL); err != nil {
		// Non-fatal: print the URL so the user can open it manually.
		fmt.Fprintf(os.Stderr, "Could not open browser automatically.\nOpen this URL in your browser:\n\n  %s\n\n", authURL)
	}

	// Wait for callback, context cancellation, or timeout.
	select {
	case <-ctx.Done():
		return nil, clierrors.OAuthTimeout()
	case result := <-resultCh:
		if result.err != "" {
			if result.err == "access_denied" {
				return nil, clierrors.OAuthCancelled()
			}
			return nil, clierrors.OAuthFailed(result.err)
		}

		if result.state != state {
			return nil, clierrors.OAuthStateMismatch()
		}

		if result.code == "" {
			return nil, clierrors.OAuthFailed("no authorization code received")
		}

		// Exchange code for token.
		return exchangeCode(ctx, clientID, clientSecret, redirectURI, result.code)
	}
}

// LoginManual performs the OAuth flow without binding a localhost callback
// server. It prints the authorize URL, the user opens it manually (useful
// over SSH, in containers, or behind firewalls), authorizes in their
// browser, then pastes the FULL redirected URL back into the terminal. The
// full URL is required because it carries the `state` query parameter the
// CLI uses to defeat CSRF; a bare authorization code is no longer accepted.
// The redirect URI is chosen from CallbackPorts[0] to keep it inside the
// integration's whitelist.
func LoginManual(ctx context.Context, clientID, clientSecret string, in io.Reader, out io.Writer) (*TokenResponse, error) {
	if clientID == "" || clientSecret == "" {
		return nil, clierrors.OAuthNotConfigured()
	}

	state, err := randomState()
	if err != nil {
		return nil, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: "Failed to generate random state",
			Err:     err,
		}
	}

	redirectURI := redirectURIFor(CallbackPorts[0])
	authURL := buildAuthorizeURL(clientID, redirectURI, state)

	_, _ = fmt.Fprintln(out, "Manual OAuth flow:")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "  1. Open this URL in any browser:")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintf(out, "     %s\n", authURL)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "  2. Authorize the integration. Your browser will be redirected")
	_, _ = fmt.Fprintf(out, "     to %s (which will fail to load \u2014 that is expected).\n", redirectURI)
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprintln(out, "  3. Copy the FULL URL from your browser's address bar and paste")
	_, _ = fmt.Fprintln(out, "     it below, then press Enter. The full URL is required \u2014 the")
	_, _ = fmt.Fprintln(out, "     authorization code alone is not sufficient.")
	_, _ = fmt.Fprintln(out)
	_, _ = fmt.Fprint(out, "Paste FULL redirected URL (the code alone is not sufficient): ")

	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && err != io.EOF {
		return nil, clierrors.OAuthFailed(fmt.Sprintf("failed to read input: %s", err))
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return nil, clierrors.OAuthFailed("no input received")
	}

	code, gotState, err := parseCallbackInput(line)
	if err != nil {
		return nil, err
	}
	// Require state on the manual flow. parseCallbackInput is permissive (it
	// accepts a bare code) but the manual flow demands the full redirected URL
	// so we can verify the state parameter — there's no other CSRF defense in
	// the manual path.
	if gotState == "" {
		// Most common cause: user pasted only the bare code. Give them the
		// specific diagnostic; the generic CSRF message is reserved for the
		// real mismatch path below.
		return nil, &clierrors.NotionCLIError{
			Code:    clierrors.CodeOAuthStateMismatch,
			Message: "OAuth state parameter is missing from the pasted input",
			Suggestions: []string{
				"Paste the FULL redirected URL, not just the authorization code.",
				"Look for ...&state=... at the end of the URL in your browser's address bar.",
				"Then run 'notion-cli auth login --manual' again.",
			},
		}
	}
	if gotState != state {
		return nil, clierrors.OAuthStateMismatch()
	}

	return exchangeCode(ctx, clientID, clientSecret, redirectURI, code)
}

// parseCallbackInput accepts either a full redirected URL containing
// ?code=...&state=... or a bare authorization code, and returns (code,
// state). State is empty when the user pasted only a code.
func parseCallbackInput(input string) (string, string, error) {
	input = strings.TrimSpace(input)
	if strings.HasPrefix(input, "http://") || strings.HasPrefix(input, "https://") {
		u, err := url.Parse(input)
		if err != nil {
			return "", "", clierrors.OAuthFailed(fmt.Sprintf("invalid URL: %s", err))
		}
		q := u.Query()
		if errParam := q.Get("error"); errParam != "" {
			if errParam == "access_denied" {
				return "", "", clierrors.OAuthCancelled()
			}
			return "", "", clierrors.OAuthFailed(errParam)
		}
		code := q.Get("code")
		if code == "" {
			return "", "", clierrors.OAuthFailed("no authorization code in URL")
		}
		return code, q.Get("state"), nil
	}
	// Treat as bare code.
	return input, "", nil
}

// buildAuthorizeURL constructs the Notion authorization URL with all
// required query parameters.
func buildAuthorizeURL(clientID, redirectURI, state string) string {
	params := url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"owner":         {"user"},
		"redirect_uri":  {redirectURI},
		"state":         {state},
	}
	return authorizeURL + "?" + params.Encode()
}

// exchangeCode sends the authorization code to the Notion token endpoint
// and returns the parsed token response. redirectURI must match the value
// used at the /authorize step (Notion validates the pair).
func exchangeCode(ctx context.Context, clientID, clientSecret, redirectURI, code string) (*TokenResponse, error) {
	bodyMap := map[string]string{
		"grant_type":   "authorization_code",
		"code":         code,
		"redirect_uri": redirectURI,
	}
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, clierrors.OAuthFailed(err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, clierrors.OAuthFailed(err.Error())
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Notion-Version", notionVersion)
	req.SetBasicAuth(clientID, clientSecret)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, clierrors.OAuthFailed(err.Error())
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, clierrors.OAuthFailed("failed to read response")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		if json.Unmarshal(respBody, &errResp) == nil {
			if msg, ok := errResp["error"].(string); ok {
				return nil, clierrors.OAuthFailed(msg)
			}
		}
		return nil, clierrors.OAuthFailed(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)))
	}

	var token TokenResponse
	if err := json.Unmarshal(respBody, &token); err != nil {
		return nil, clierrors.OAuthFailed("failed to parse token response")
	}

	if token.AccessToken == "" {
		return nil, clierrors.OAuthFailed("empty access token in response")
	}

	return &token, nil
}

// SetTokenURL overrides the token endpoint URL. Intended for use in tests only.
func SetTokenURL(u string) { tokenURL = u }

// SetIntrospectURL overrides the introspect endpoint URL. Intended for use in tests only.
func SetIntrospectURL(u string) { introspectURL = u }

// SetRevokeURL overrides the revoke endpoint URL. Intended for use in tests only.
func SetRevokeURL(u string) { revokeURL = u }

// TokenRefresh exchanges a refresh token for a new access token.
// It POSTs to tokenURL with grant_type=refresh_token and HTTP Basic auth.
func TokenRefresh(ctx context.Context, clientID, clientSecret, refreshToken string) (*TokenResponse, error) {
	bodyMap := map[string]string{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
	}
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, clierrors.OAuthFailed(err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, clierrors.OAuthFailed(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Notion-Version", notionVersion)
	req.SetBasicAuth(clientID, clientSecret)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, clierrors.OAuthFailed(err.Error())
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, clierrors.OAuthFailed("failed to read response")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		if json.Unmarshal(respBody, &errResp) == nil {
			if msg, ok := errResp["error"].(string); ok {
				return nil, clierrors.OAuthFailed(msg)
			}
		}
		return nil, clierrors.OAuthFailed(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)))
	}

	var token TokenResponse
	if err := json.Unmarshal(respBody, &token); err != nil {
		return nil, clierrors.OAuthFailed("failed to parse token response")
	}
	if token.AccessToken == "" {
		return nil, clierrors.OAuthFailed("empty access token in response")
	}
	return &token, nil
}

// TokenIntrospect queries the active status and metadata of an OAuth token.
// It POSTs to introspectURL with HTTP Basic auth.
func TokenIntrospect(ctx context.Context, clientID, clientSecret, token string) (*IntrospectResponse, error) {
	bodyMap := map[string]string{"token": token}
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return nil, clierrors.OAuthFailed(err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, introspectURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, clierrors.OAuthFailed(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Notion-Version", notionVersion)
	req.SetBasicAuth(clientID, clientSecret)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, clierrors.OAuthFailed(err.Error())
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return nil, clierrors.OAuthFailed("failed to read response")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		if json.Unmarshal(respBody, &errResp) == nil {
			if msg, ok := errResp["error"].(string); ok {
				return nil, clierrors.OAuthFailed(msg)
			}
		}
		return nil, clierrors.OAuthFailed(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)))
	}

	var result IntrospectResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, clierrors.OAuthFailed("failed to parse introspect response")
	}
	return &result, nil
}

// TokenRevoke revokes an OAuth token. Returns nil on success.
// It POSTs to revokeURL with HTTP Basic auth.
func TokenRevoke(ctx context.Context, clientID, clientSecret, token string) error {
	bodyMap := map[string]string{"token": token}
	bodyBytes, err := json.Marshal(bodyMap)
	if err != nil {
		return clierrors.OAuthFailed(err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, revokeURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return clierrors.OAuthFailed(err.Error())
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Notion-Version", notionVersion)
	req.SetBasicAuth(clientID, clientSecret)

	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return clierrors.OAuthFailed(err.Error())
	}
	defer resp.Body.Close() //nolint:errcheck

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBody))
	if err != nil {
		return clierrors.OAuthFailed("failed to read response")
	}

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]any
		if json.Unmarshal(respBody, &errResp) == nil {
			if msg, ok := errResp["error"].(string); ok {
				return clierrors.OAuthFailed(msg)
			}
		}
		return clierrors.OAuthFailed(fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBody)))
	}
	return nil
}

// randomState generates a 32-byte hex-encoded random string.
func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// openBrowser opens the given URL in the user's default browser.
func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	return cmd.Start()
}

// callbackHTML renders the OAuth callback page shown to the user in their
// browser after Notion redirects back to localhost. The page is intentionally
// self-contained (no external assets) so it works offline and instantly.
func callbackHTML(success bool, title, message string) string {
	accent := "#16a34a" // green for success
	icon := `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"></polyline></svg>`
	if !success {
		accent = "#dc2626" // red for failure
		icon = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line></svg>`
	}

	safeTitle := html.EscapeString(title)
	safeMessage := html.EscapeString(message)
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>%[2]s · notion-cli</title>
<style>
  *,*::before,*::after{box-sizing:border-box}
  html,body{margin:0;padding:0;height:100%%}
  body{
    font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","Inter",Roboto,sans-serif;
    display:flex;align-items:center;justify-content:center;
    min-height:100vh;padding:1.5rem;
    background:#fafaf9;color:#0f0f0f;
    -webkit-font-smoothing:antialiased;
  }
  .card{
    width:100%%;max-width:440px;
    background:#fff;border:1px solid rgba(0,0,0,0.06);
    border-radius:14px;padding:2.5rem 2rem;
    box-shadow:0 1px 2px rgba(0,0,0,0.04),0 8px 24px rgba(0,0,0,0.06);
    text-align:center;
  }
  .icon{
    width:56px;height:56px;margin:0 auto 1.25rem;
    border-radius:50%%;display:flex;align-items:center;justify-content:center;
    background:%[1]s14;color:%[1]s;
  }
  .icon svg{width:30px;height:30px}
  h1{font-size:1.375rem;font-weight:600;margin:0 0 .5rem;letter-spacing:-0.01em}
  p{font-size:.95rem;line-height:1.55;color:#525252;margin:0 0 1.5rem}
  .terminal{
    font-family:ui-monospace,SFMono-Regular,"SF Mono",Menlo,Consolas,monospace;
    font-size:.8rem;color:#737373;
    background:#f5f5f4;border:1px solid rgba(0,0,0,0.05);
    border-radius:8px;padding:.6rem .75rem;
    display:inline-block;
  }
  .footer{margin-top:1.5rem;font-size:.75rem;color:#a3a3a3;letter-spacing:.02em}
  .footer a{color:#737373;text-decoration:none;border-bottom:1px solid rgba(0,0,0,0.1)}
  @media (prefers-color-scheme: dark){
    body{background:#0a0a0a;color:#fafafa}
    .card{background:#171717;border-color:rgba(255,255,255,0.06);box-shadow:0 1px 2px rgba(0,0,0,0.4),0 8px 24px rgba(0,0,0,0.5)}
    p{color:#a3a3a3}
    .terminal{background:#0a0a0a;border-color:rgba(255,255,255,0.08);color:#a3a3a3}
    .footer{color:#525252}
    .footer a{color:#a3a3a3;border-bottom-color:rgba(255,255,255,0.15)}
  }
</style>
</head>
<body>
  <div class="card">
    <div class="icon">%[4]s</div>
    <h1>%[2]s</h1>
    <p>%[3]s</p>
    <div class="terminal">$ notion-cli auth status</div>
    <div class="footer">notion-cli · <a href="https://github.com/Coastal-Programs/notion-cli">github.com/Coastal-Programs/notion-cli</a></div>
  </div>
</body>
</html>`, accent, safeTitle, safeMessage, icon)
}
