package notion

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Coastal-Programs/notion-cli/v6/internal/config"
	"github.com/Coastal-Programs/notion-cli/v6/internal/oauth"
	"github.com/Coastal-Programs/notion-cli/v6/internal/retry"
)

const (
	defaultBaseURL       = "https://api.notion.com/v1"
	defaultNotionVersion = "2026-03-11"
)

// UserAgent is set during init to include the build-time version.
var userAgent string

func init() {
	userAgent = "notion-cli/" + config.Version + " (Go/" + runtime.Version() + ")"
}

// APIError represents an error response from the Notion API.
type APIError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("notion api error (status %d, code %q): %s", e.Status, e.Code, e.Message)
}

// QueryParams holds pagination parameters for list endpoints. BlockID and
// PageID are used by endpoints that filter by a parent identifier in the
// query string (e.g. GET /v1/comments).
type QueryParams struct {
	StartCursor string
	PageSize    int
	BlockID     string
	PageID      string
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
	if q.BlockID != "" {
		v.Set("block_id", q.BlockID)
	}
	if q.PageID != "" {
		v.Set("page_id", q.PageID)
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

// WithKeepAlive controls whether the underlying HTTP transport reuses
// connections. When false, a transport with DisableKeepAlives=true is
// installed; when true, the default transport is left in place (keep-alive
// is the standard library default).
func WithKeepAlive(enabled bool) ClientOption {
	return func(c *Client) {
		if enabled {
			return
		}
		base, ok := c.httpClient.Transport.(*http.Transport)
		if !ok || base == nil {
			base, ok = http.DefaultTransport.(*http.Transport)
			if !ok {
				return
			}
		}
		t := base.Clone()
		t.DisableKeepAlives = true
		c.httpClient.Transport = t
	}
}

// Client is an HTTP client for the Notion API.
type Client struct {
	httpClient    *http.Client
	token         string
	baseURL       string
	notionVersion string
	retryConfig   *retry.RetryConfig
	cfg           *config.Config // optional; enables auto-refresh on 401
}

// WithConfig attaches a Config to the client so it can automatically refresh
// the OAuth access token when a 401 response is received.
func WithConfig(cfg *config.Config) ClientOption {
	return func(c *Client) { c.cfg = cfg }
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

// --- Data Sources ---
//
// Per the 2025-09-03 Notion API contract, every database has one or more
// data sources. These endpoints address an individual data source by ID.
// See https://developers.notion.com/reference/data-source.

// DataSourceRetrieve retrieves a data source by ID.
func (c *Client) DataSourceRetrieve(ctx context.Context, dataSourceID string) (map[string]any, error) {
	return c.get(ctx, "/data_sources/"+dataSourceID, nil)
}

// DataSourceQuery queries a data source with optional filters, sorts, and pagination.
func (c *Client) DataSourceQuery(ctx context.Context, dataSourceID string, body map[string]any) (map[string]any, error) {
	return c.post(ctx, "/data_sources/"+dataSourceID+"/query", body)
}

// DataSourceCreate creates a new data source.
func (c *Client) DataSourceCreate(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.post(ctx, "/data_sources", body)
}

// DataSourceUpdate updates an existing data source.
func (c *Client) DataSourceUpdate(ctx context.Context, dataSourceID string, body map[string]any) (map[string]any, error) {
	return c.patch(ctx, "/data_sources/"+dataSourceID, body)
}

// DataSourceTemplatesList lists the templates available for a data source.
// Pagination is controlled via QueryParams (StartCursor, PageSize).
func (c *Client) DataSourceTemplatesList(ctx context.Context, dataSourceID string, query QueryParams) (map[string]any, error) {
	return c.get(ctx, "/data_sources/"+dataSourceID+"/templates", query.Values())
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

// PageMarkdownGet retrieves a page's content rendered as enhanced markdown.
// See https://developers.notion.com/reference/retrieve-page-markdown.
func (c *Client) PageMarkdownGet(ctx context.Context, pageID string, query QueryParams) (map[string]any, error) {
	return c.get(ctx, "/pages/"+pageID+"/markdown", query.Values())
}

// PageMarkdownUpdate inserts or replaces page content using enhanced markdown.
// The body shape is determined by the `type` field (e.g. replace_content,
// insert_content, update_content, replace_content_range).
// See https://developers.notion.com/reference/update-page-markdown.
func (c *Client) PageMarkdownUpdate(ctx context.Context, pageID string, body map[string]any) (map[string]any, error) {
	return c.patch(ctx, "/pages/"+pageID+"/markdown", body)
}

// PageTrash moves a page to trash.
func (c *Client) PageTrash(ctx context.Context, pageID string) (map[string]any, error) {
	return c.PageUpdate(ctx, pageID, map[string]any{"in_trash": true})
}

// PageRestore restores a trashed page.
func (c *Client) PageRestore(ctx context.Context, pageID string) (map[string]any, error) {
	return c.PageUpdate(ctx, pageID, map[string]any{"in_trash": false})
}

// PageMove moves a page to a new parent via POST /pages/{id}/move.
// body must contain a "parent" key per the Notion Move Page API.
func (c *Client) PageMove(ctx context.Context, pageID string, body map[string]any) (map[string]any, error) {
	return c.post(ctx, "/pages/"+pageID+"/move", body)
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

// --- Comments ---

// CommentCreate creates a new comment on a page or discussion thread.
// Body must include either a `parent` (with page_id) or `discussion_id`,
// plus exactly one of `rich_text` or `markdown`.
func (c *Client) CommentCreate(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.post(ctx, "/comments", body)
}

// CommentList lists open (unresolved) comments for a block or page.
// query.BlockID or query.PageID is required.
func (c *Client) CommentList(ctx context.Context, query QueryParams) (map[string]any, error) {
	return c.get(ctx, "/comments", query.Values())
}

// CommentRetrieve retrieves a single comment by ID.
func (c *Client) CommentRetrieve(ctx context.Context, id string) (map[string]any, error) {
	return c.get(ctx, "/comments/"+id, nil)
}

// CommentUpdate updates the content of an existing comment. Body must contain
// exactly one of `rich_text` or `markdown`.
func (c *Client) CommentUpdate(ctx context.Context, id string, body map[string]any) (map[string]any, error) {
	return c.patch(ctx, "/comments/"+id, body)
}

// CommentDelete deletes a comment by ID.
func (c *Client) CommentDelete(ctx context.Context, id string) (map[string]any, error) {
	return c.delete(ctx, "/comments/"+id)
}

// --- Views ---
//
// Per the 2026-03-11 Notion API contract, views expose a database's
// configured filters, sorts, and grouping. View queries cache the result
// set on the server for paginated retrieval (cache TTL: 15 minutes).
// See https://developers.notion.com/reference/list-views.

// ViewsList lists views for a database or data source. Exactly one of
// databaseID or dataSourceID must be supplied. Pagination is controlled
// via QueryParams.
func (c *Client) ViewsList(ctx context.Context, databaseID, dataSourceID string, query QueryParams) (map[string]any, error) {
	params := query.Values()
	if databaseID != "" {
		params.Set("database_id", databaseID)
	}
	if dataSourceID != "" {
		params.Set("data_source_id", dataSourceID)
	}
	return c.get(ctx, "/views", params)
}

// ViewRetrieve retrieves a view by ID.
func (c *Client) ViewRetrieve(ctx context.Context, viewID string) (map[string]any, error) {
	return c.get(ctx, "/views/"+viewID, nil)
}

// ViewQueryCreate executes a view's filter and sort configuration against
// its data source, caches the full result set, and returns the first page
// of page references along with a query_id for paginating through results.
func (c *Client) ViewQueryCreate(ctx context.Context, viewID string, body map[string]any) (map[string]any, error) {
	if body == nil {
		body = map[string]any{}
	}
	return c.post(ctx, "/views/"+viewID+"/queries", body)
}

// ViewQueryResults paginates through cached view query results.
func (c *Client) ViewQueryResults(ctx context.Context, viewID, queryID string, query QueryParams) (map[string]any, error) {
	return c.get(ctx, "/views/"+viewID+"/queries/"+queryID, query.Values())
}

// ViewCreate creates a new view. body must include data_source_id, name, and
// type. Optionally include database_id, filter, and sorts.
// See https://developers.notion.com/reference/create-view.
func (c *Client) ViewCreate(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.post(ctx, "/views", body)
}

// ViewUpdate updates an existing view. All fields are optional: name, filter, sorts.
// See https://developers.notion.com/reference/update-view.
func (c *Client) ViewUpdate(ctx context.Context, viewID string, body map[string]any) (map[string]any, error) {
	return c.patch(ctx, "/views/"+viewID, body)
}

// ViewDelete deletes a view by ID. Returns the deleted View object.
// See https://developers.notion.com/reference/delete-view.
func (c *Client) ViewDelete(ctx context.Context, viewID string) (map[string]any, error) {
	return c.delete(ctx, "/views/"+viewID)
}

// ViewQueryDelete deletes a cached view query. Idempotent: returns success
// even if the query doesn't exist or has already expired.
func (c *Client) ViewQueryDelete(ctx context.Context, viewID, queryID string) (map[string]any, error) {
	return c.delete(ctx, "/views/"+viewID+"/queries/"+queryID)
}

// --- Custom Emojis ---
//
// Per the 2026-03-11 Notion API contract, custom emojis are workspace-managed
// emoji icons addressed by ID. See https://developers.notion.com/reference/list-custom-emojis.

// CustomEmojiList lists custom emojis in the workspace with pagination.
func (c *Client) CustomEmojiList(ctx context.Context, query QueryParams) (map[string]any, error) {
	return c.get(ctx, "/custom_emojis", query.Values())
}

// CustomEmojiRetrieve retrieves a single custom emoji by ID.
func (c *Client) CustomEmojiRetrieve(ctx context.Context, id string) (map[string]any, error) {
	return c.get(ctx, "/custom_emojis/"+id, nil)
}

// --- Search ---

// Search searches across all pages and databases the integration has access to.
func (c *Client) Search(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.post(ctx, "/search", body)
}

// --- File Uploads ---

// FileUploadCreate creates a new file upload. body must include "filename",
// "content_type", and "mode" ("single_part" or "multi_part"). For multi-part
// uploads also include "number_of_parts".
func (c *Client) FileUploadCreate(ctx context.Context, body map[string]any) (map[string]any, error) {
	return c.post(ctx, "/file_uploads", body)
}

// FileUploadSend sends file data for an upload. partNumber==0 means single-part
// (the part_number field is omitted). partNumber>0 is used for multi-part
// uploads (1-indexed). data must be ≤20 MB.
func (c *Client) FileUploadSend(ctx context.Context, id string, partNumber int, filename, contentType string, data []byte) (map[string]any, error) {
	fields := map[string]string{}
	if partNumber > 0 {
		fields["part_number"] = strconv.Itoa(partNumber)
	}
	return c.doFormData(ctx, "/file_uploads/"+id+"/send", fields, "file", filename, contentType, data)
}

// FileUploadComplete completes a multi-part upload after all parts have been sent.
func (c *Client) FileUploadComplete(ctx context.Context, id string) (map[string]any, error) {
	return c.post(ctx, "/file_uploads/"+id+"/complete", map[string]any{})
}

// FileUploadRetrieve retrieves the status of a file upload by ID.
func (c *Client) FileUploadRetrieve(ctx context.Context, id string) (map[string]any, error) {
	return c.get(ctx, "/file_uploads/"+id, nil)
}

// FileUploadList lists file uploads with optional pagination.
func (c *Client) FileUploadList(ctx context.Context, query QueryParams) (map[string]any, error) {
	return c.get(ctx, "/file_uploads", query.Values())
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
	result, err := c.doInternal(ctx, method, path, params, body)
	if err == nil {
		return result, nil
	}

	// Auto-refresh on 401 when a refresh token is available.
	var apiErr *APIError
	if !errors.As(err, &apiErr) || apiErr.Status != 401 {
		return nil, err
	}
	if c.cfg == nil || c.cfg.OAuthRefreshToken == "" || config.OAuthClientID == "" {
		return nil, err
	}

	newToken, refreshErr := oauth.TokenRefresh(ctx, config.OAuthClientID, config.OAuthClientSecret, c.cfg.OAuthRefreshToken)
	if refreshErr != nil {
		return nil, err // return original 401
	}

	c.token = newToken.AccessToken
	c.cfg.OAuthAccessToken = newToken.AccessToken
	// Only overwrite the stored refresh token if the server returned a new one.
	// Notion does not always rotate refresh tokens; an empty response field
	// must not blank out the credential we already have on disk.
	if newToken.RefreshToken != "" {
		c.cfg.OAuthRefreshToken = newToken.RefreshToken
	}
	if newToken.ExpiresIn > 0 {
		c.cfg.OAuthTokenExpiresAt = time.Now().Add(
			time.Duration(newToken.ExpiresIn) * time.Second).UTC().Format(time.RFC3339)
	}
	_ = config.SaveConfig(c.cfg) // best-effort

	return c.doInternal(ctx, method, path, params, body) // retry once
}

func (c *Client) doInternal(ctx context.Context, method, path string, params url.Values, body map[string]any) (map[string]any, error) {
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
		defer resp.Body.Close() //nolint:errcheck

		// Handle gzip-encoded responses.
		var reader io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(resp.Body)
			if err != nil {
				return fmt.Errorf("create gzip reader: %w", err)
			}
			defer gr.Close() //nolint:errcheck
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

// doFormData sends a multipart/form-data POST request. String key-value pairs
// are written as form fields; data is written as a file part named fileField
// with the given filename and contentType. The multipart body is buffered once
// and replayed across retry attempts.
func (c *Client) doFormData(ctx context.Context, path string, fields map[string]string, fileField, filename, contentType string, data []byte) (map[string]any, error) {
	// Build the multipart body once so it can be replayed across retries.
	var bodyBuf bytes.Buffer
	mw := multipart.NewWriter(&bodyBuf)

	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			return nil, fmt.Errorf("write form field %q: %w", k, err)
		}
	}

	h := map[string][]string{
		"Content-Disposition": {fmt.Sprintf(`form-data; name=%q; filename=%q`, fileField, filename)},
		"Content-Type":        {contentType},
	}
	part, err := mw.CreatePart(h)
	if err != nil {
		return nil, fmt.Errorf("create file part: %w", err)
	}
	if _, err := part.Write(data); err != nil {
		return nil, fmt.Errorf("write file data: %w", err)
	}
	if err := mw.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	bodyBytes := bodyBuf.Bytes()
	contentTypeHeader := mw.FormDataContentType()
	u := c.baseURL + path

	var result map[string]any

	err = retry.Do(ctx, *c.retryConfig, func() error {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(bodyBytes))
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}

		req.Header.Set("Authorization", "Bearer "+c.token)
		req.Header.Set("Notion-Version", c.notionVersion)
		req.Header.Set("User-Agent", userAgent)
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", contentTypeHeader)

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("execute request: %w", err)
		}
		defer resp.Body.Close() //nolint:errcheck

		var reader io.Reader = resp.Body
		if resp.Header.Get("Content-Encoding") == "gzip" {
			gr, err := gzip.NewReader(resp.Body)
			if err != nil {
				return fmt.Errorf("create gzip reader: %w", err)
			}
			defer gr.Close() //nolint:errcheck
			reader = gr
		}

		const maxResponseSize = 50 * 1024 * 1024
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
