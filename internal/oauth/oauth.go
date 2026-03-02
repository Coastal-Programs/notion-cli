// Package oauth implements the Notion OAuth 2.0 authorization flow for the CLI.
// It starts a local HTTP server, opens the user's browser to the Notion
// authorization page, and exchanges the received code for an access token.
package oauth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/internal/errors"
)


const (
	// callbackPort is the port for the localhost OAuth callback server.
	callbackPort = 8080

	// redirectURI must match the one registered with Notion.
	redirectURI = "http://localhost:8080/callback"

	// authorizeURL is the Notion OAuth authorization endpoint.
	authorizeURL = "https://api.notion.com/v1/oauth/authorize"

	// maxResponseBody limits the token exchange response to 1 MB.
	maxResponseBody = 1 << 20
)

// tokenURL is the Notion OAuth token exchange endpoint. It is a var so tests
// can override it with an httptest server URL.
var tokenURL = "https://api.notion.com/v1/oauth/token"

// TokenResponse holds the result of a successful OAuth token exchange.
type TokenResponse struct {
	AccessToken   string `json:"access_token"`
	TokenType     string `json:"token_type"`
	BotID         string `json:"bot_id"`
	WorkspaceID   string `json:"workspace_id"`
	WorkspaceName string `json:"workspace_name"`
	WorkspaceIcon string `json:"workspace_icon"`
}

// callbackResult is sent from the HTTP handler back to the Login goroutine.
type callbackResult struct {
	code  string
	state string
	err   string
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
	// intercepting the OAuth callback.
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", callbackPort))
	if err != nil {
		return nil, clierrors.OAuthPortInUse(callbackPort)
	}

	resultCh := make(chan callbackResult, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()

		if errParam := q.Get("error"); errParam != "" {
			resultCh <- callbackResult{err: errParam}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, callbackHTML("Authorization Denied", "You denied the authorization request. You can close this tab."))
			return
		}

		resultCh <- callbackResult{
			code:  q.Get("code"),
			state: q.Get("state"),
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, callbackHTML("Authorization Successful", "You can close this tab and return to your terminal."))
	})

	server := &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server on the already-bound listener.
	go func() {
		if sErr := server.Serve(ln); sErr != nil && sErr != http.ErrServerClosed {
			resultCh <- callbackResult{err: sErr.Error()}
		}
	}()
	defer server.Shutdown(context.Background())

	// Build and open the authorization URL with properly encoded parameters.
	params := url.Values{
		"client_id":     {clientID},
		"response_type": {"code"},
		"owner":         {"user"},
		"redirect_uri":  {redirectURI},
		"state":         {state},
	}
	authURL := authorizeURL + "?" + params.Encode()

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
		return exchangeCode(ctx, clientID, clientSecret, result.code)
	}
}

// exchangeCode sends the authorization code to the Notion token endpoint
// and returns the parsed token response.
func exchangeCode(ctx context.Context, clientID, clientSecret, code string) (*TokenResponse, error) {
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
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, clierrors.OAuthFailed(err.Error())
	}
	defer resp.Body.Close()

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


// callbackHTML returns a simple HTML page for the OAuth callback.
func callbackHTML(title, message string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>%s</title>
<style>body{font-family:system-ui,sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#fafafa}
.card{text-align:center;padding:2rem;border-radius:8px;background:white;box-shadow:0 2px 8px rgba(0,0,0,0.1)}
h1{font-size:1.5rem;margin-bottom:0.5rem}p{color:#666}</style>
</head><body><div class="card"><h1>%s</h1><p>%s</p></div></body></html>`, title, title, message)
}
