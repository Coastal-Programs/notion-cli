package notion

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Coastal-Programs/notion-cli/internal/retry"
)

const (
	defaultBaseURL       = "https://api.notion.com/v1"
	defaultNotionVersion = "2022-06-28"
)

var userAgent = "notion-cli/1.0 (Go/" + runtime.Version() + ")"

// APIError represents an error response from the Notion API.
type APIError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("notion api error (status %d, code %q): %s", e.Status, e.Code, e.Message)
}

// QueryParams holds pagination parameters for list endpoints.
type QueryParams struct {
	StartCursor string
	PageSize    int
}

// Values converts QueryParams to url.Values for use in HTTP requests.
func (q QueryParams) Values() url.Values {
	v := url.Values{}
	if q.StartCursor != "" {
		v.Set("start_cursor", q.StartCursor)
	}
	if q.PageSize > 0 {
		v.Set("page_size", strconv.Itoa(q.PageSize))
	}
	return v
}

// ClientOption configures a Client.
type ClientOption func(*Client)

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(c *Client) {
		c.httpClient = hc
	}
}

// WithBaseURL overrides the default Notion API base URL.
func WithBaseURL(u string) ClientOption {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(u, "/")
	}
}

// WithNotionVersion overrides the Notion-Version header.
func WithNotionVersion(v string) ClientOption {
	return func(c *Client) {
		c.notionVersion = v
	}
}

// WithTimeout sets the HTTP client timeout.
func WithTimeout(d time.Duration) ClientOption {
	return func(c *Client) {
		c.httpClient.Timeout = d
	}
}

// WithRetryConfig overrides the default retry configuration.
func WithRetryConfig(cfg retry.RetryConfig) ClientOption {
	return func(c *Client) {
		c.retryConfig = &cfg
	}
}

// Client is an HTTP client for the Notion API.
type Client struct {
	httpClient    *http.Client
	token         string
	baseURL       string
	notionVersion string
	retryConfig   *retry.RetryConfig
}

// NewClient creates a new Notion API client.
func NewClient(token string, opts ...ClientOption) *Client {
	rc := retry.DefaultRetryConfig()
	c := &Client{
		httpClient:    &http.Client{Timeout: 30 * time.Second},
		token:         token,
		baseURL:       defaultBaseURL,
		notionVersion: defaultNotionVersion,
		retryConfig:   &rc,
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// --- Databases ---

// DatabaseRetrieve retrieves a database by ID.
func (c *Client) DatabaseRetrieve(ctx context.Context, dbID string) (map[string]any, error) {
	return c.get(ctx, "/databases/"+dbID, nil)
}

// DatabaseQuery queries a database with optional filters, sorts, and pagination.
func (c *Client) DatabaseQuery(ctx context.Context, dbID string, body map[string]any) (map[string]any, error) {
	return c.post(ctx, "/databases/"+dbID+"/query", body)
}

// DatabaseCreate creates a new database.
func (c *Client) DatabaseCreate(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.post(ctx, "/databases", body)
}

// DatabaseUpdate updates an existing database.
func (c *Client) DatabaseUpdate(ctx context.Context, dbID string, body map[string]any) (map[string]any, error) {
	return c.patch(ctx, "/databases/"+dbID, body)
}

// --- Pages ---

// PageCreate creates a new page.
func (c *Client) PageCreate(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.post(ctx, "/pages", body)
}

// PageRetrieve retrieves a page by ID.
func (c *Client) PageRetrieve(ctx context.Context, pageID string) (map[string]any, error) {
	return c.get(ctx, "/pages/"+pageID, nil)
}

// PageUpdate updates an existing page.
func (c *Client) PageUpdate(ctx context.Context, pageID string, body map[string]any) (map[string]any, error) {
	return c.patch(ctx, "/pages/"+pageID, body)
}

// PagePropertyRetrieve retrieves a page property item.
func (c *Client) PagePropertyRetrieve(ctx context.Context, pageID, propID string, query QueryParams) (map[string]any, error) {
	return c.get(ctx, "/pages/"+pageID+"/properties/"+propID, query.Values())
}

// --- Blocks ---

// BlockRetrieve retrieves a block by ID.
func (c *Client) BlockRetrieve(ctx context.Context, blockID string) (map[string]any, error) {
	return c.get(ctx, "/blocks/"+blockID, nil)
}

// BlockUpdate updates an existing block.
func (c *Client) BlockUpdate(ctx context.Context, blockID string, body map[string]any) (map[string]any, error) {
	return c.patch(ctx, "/blocks/"+blockID, body)
}

// BlockDelete deletes a block by ID.
func (c *Client) BlockDelete(ctx context.Context, blockID string) (map[string]any, error) {
	return c.delete(ctx, "/blocks/"+blockID)
}

// BlockChildrenList lists children of a block with pagination.
func (c *Client) BlockChildrenList(ctx context.Context, blockID string, query QueryParams) (map[string]any, error) {
	return c.get(ctx, "/blocks/"+blockID+"/children", query.Values())
}

// BlockChildrenAppend appends children blocks to a parent block.
func (c *Client) BlockChildrenAppend(ctx context.Context, blockID string, body map[string]any) (map[string]any, error) {
	return c.patch(ctx, "/blocks/"+blockID+"/children", body)
}

// --- Users ---

// UsersList lists all users in the workspace with pagination.
func (c *Client) UsersList(ctx context.Context, query QueryParams) (map[string]any, error) {
	return c.get(ctx, "/users", query.Values())
}

// UserRetrieve retrieves a user by ID.
func (c *Client) UserRetrieve(ctx context.Context, userID string) (map[string]any, error) {
	return c.get(ctx, "/users/"+userID, nil)
}

// UsersMe retrieves the bot user associated with the token.
func (c *Client) UsersMe(ctx context.Context) (map[string]any, error) {
	return c.get(ctx, "/users/me", nil)
}

// --- Search ---

// Search searches across all pages and databases the integration has access to.
func (c *Client) Search(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.post(ctx, "/search", body)
}

// --- Internal HTTP helpers ---

func (c *Client) get(ctx context.Context, path string, params url.Values) (map[string]any, error) {
	return c.do(ctx, http.MethodGet, path, params, nil)
}

func (c *Client) post(ctx context.Context, path string, body map[string]any) (map[string]any, error) {
	return c.do(ctx, http.MethodPost, path, nil, body)
}

func (c *Client) patch(ctx context.Context, path string, body map[string]any) (map[string]any, error) {
	return c.do(ctx, http.MethodPatch, path, nil, body)
}

func (c *Client) delete(ctx context.Context, path string) (map[string]any, error) {
	return c.do(ctx, http.MethodDelete, path, nil, nil)
}

func (c *Client) do(ctx context.Context, method, path string, params url.Values, body map[string]any) (map[string]any, error) {
	u := c.baseURL + path
	if len(params) > 0 {
		u += "?" + params.Encode()
	}

	// Marshal body once so it can be reused across retry attempts.
	var bodyBytes []byte
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyBytes = b
	}

	var result map[string]any

	err := retry.Do(ctx, *c.retryConfig, func() error {
		var bodyReader io.Reader
		if bodyBytes != nil {
			bodyReader = bytes.NewReader(bodyBytes)
		}

		req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Notion-Version", c.notionVersion)
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept-Encoding", "gzip")
		if bodyBytes != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			// Network errors are retryable for GET requests.
			if method == http.MethodGet {
				return &retry.RetryableError{Err: fmt.Errorf("execute request: %w", err)}
			}
			return fmt.Errorf("execute request: %w", err)
		}
		defer resp.Body.Close()

		// Handle gzip-encoded responses.
		var reader io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(resp.Body)
			if err != nil {
				return fmt.Errorf("create gzip reader: %w", err)
			}
			defer gr.Close()
			reader = gr
		}

		const maxResponseSize = 50 * 1024 * 1024 // 50MB
		respBody, err := io.ReadAll(io.LimitReader(reader, maxResponseSize))
		if err != nil {
			return fmt.Errorf("read response body: %w", err)
		}

		if resp.StatusCode >= 400 {
			var apiErr APIError
			if jsonErr := json.Unmarshal(respBody, &apiErr); jsonErr != nil {
				apiErr = APIError{
					Status:  resp.StatusCode,
					Code:    "unknown",
					Message: string(respBody),
				}
			}
			apiErr.Status = resp.StatusCode

			// Retry on 429/5xx for any method, and all retryable status codes for GET.
			if retry.IsRetryable(resp.StatusCode) {
				retryAfter := parseDuration(resp.Header.Get("Retry-After"))
				return &retry.RetryableError{
					Err:        &apiErr,
					StatusCode: resp.StatusCode,
					RetryAfter: retryAfter,
				}
			}

			return &apiErr
		}

		// DELETE may return 200 with an object body.
		if len(respBody) == 0 {
			result = map[string]any{}
			return nil
		}

		if err := json.Unmarshal(respBody, &result); err != nil {
			return fmt.Errorf("unmarshal response: %w", err)
		}
		return nil
	})

	if err != nil {
		// Unwrap RetryableError to return the original APIError.
		if retryErr, ok := err.(*retry.RetryableError); ok {
			return nil, retryErr.Err
		}
		return nil, err
	}

	return result, nil
}

// parseDuration parses a Retry-After header value (in seconds) into a
// time.Duration. Returns 0 if the value cannot be parsed.
func parseDuration(s string) time.Duration {
	if s == "" {
		return 0
	}
	secs, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return time.Duration(secs) * time.Second
}
