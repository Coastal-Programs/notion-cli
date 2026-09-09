package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/internal/notion"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterAPICommand registers the raw `api` escape hatch under root.
func RegisterAPICommand(root *cobra.Command) {
	root.AddCommand(newAPICmd())
}

func newAPICmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api <endpoint>",
		Short: "Make an authenticated request to any Notion API endpoint",
		Long: `Make an authenticated HTTP request to the Notion API and print the response.

The endpoint is a path under https://api.notion.com/v1, such as "users/me" or
"data_sources/<id>/query". Authentication, Notion-Version, gzip, retries, and
token refresh are handled exactly as for every other command, so endpoints
this CLI has no dedicated command for (agents, sessions, meeting notes, async
tasks, betas) are reachable today.

The method defaults to GET, or POST when any body input is supplied. Body
fields come from --input (a JSON file, or "-" for stdin) and/or repeated
-f / -F flags; flags are merged on top of --input. With an explicit -X GET or
-X DELETE, -f / -F values become query parameters instead. Keys may use dots to nest
(parent.page_id=...). -f values are strings; -F values are parsed as JSON
(numbers, booleans, null, objects), and "@file" or "@-" reads the value from a
file or stdin.

--paginate follows next_cursor until has_more is false and merges every
page's results into one list.`,
		Example: `  # Who am I?
  notion-cli api users/me

  # List users, 10 per page (fields on an explicit GET become query params)
  notion-cli api users -X GET -f page_size=10

  # Query a data source with a filter from a file, fetching every page
  notion-cli api data_sources/<id>/query --input filter.json --paginate

  # Create a page with nested fields
  notion-cli api pages -f parent.page_id=<id> -F 'properties={"title":[{"text":{"content":"Hi"}}]}'

  # Retrieve an async task
  notion-cli api async_tasks/<task_id>

  # Opt into a beta for one call
  notion-cli api -H 'Notion-Beta: workspace-as-code' workspaces/<id>/export

  # Move a page to trash
  notion-cli api -X PATCH pages/<id> -F in_trash=true`,
		Args: cobra.ExactArgs(1),
		RunE: runAPI,
	}
	cmd.Flags().StringP("method", "X", "", "HTTP method (default GET, or POST when a body is given)")
	cmd.Flags().StringArrayP("raw-field", "f", nil, "Add a string field: key=value")
	cmd.Flags().StringArrayP("field", "F", nil, "Add a typed field: key=value, value is JSON (@file or @- reads it)")
	cmd.Flags().StringArrayP("header", "H", nil, "Add an HTTP header: 'Name: value'")
	cmd.Flags().String("input", "", "Read the JSON request body from a file, or '-' for stdin")
	cmd.Flags().Bool("paginate", false, "Follow next_cursor and merge all pages' results")
	addOutputFlags(cmd)
	return cmd
}

func runAPI(cmd *cobra.Command, args []string) error {
	start := time.Now()
	flags := cmd.Flags()
	method, _ := flags.GetString("method")
	rawFields, _ := flags.GetStringArray("raw-field")
	typedFields, _ := flags.GetStringArray("field")
	headerFlags, _ := flags.GetStringArray("header")
	input, _ := flags.GetString("input")
	paginate, _ := flags.GetBool("paginate")

	path, params, err := parseAPIEndpoint(args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	body, err := parseAPIBody(input, os.Stdin)
	if err != nil {
		return handleError(cmd, err)
	}

	fields, err := parseAPIFields(rawFields, typedFields, os.Stdin)
	if err != nil {
		return handleError(cmd, err)
	}

	method = strings.ToUpper(method)
	if method == "" {
		method = http.MethodGet
		if body != nil || len(fields) > 0 {
			method = http.MethodPost
		}
	}

	// gh convention: fields on a GET/DELETE are query parameters.
	if method == http.MethodGet || method == http.MethodDelete {
		for k, v := range fields {
			params.Set(k, fmt.Sprint(v))
		}
	} else if len(fields) > 0 {
		if body == nil {
			body = map[string]any{}
		}
		for k, v := range fields {
			if err := setNestedField(body, k, v); err != nil {
				return handleError(cmd, err)
			}
		}
	}

	headers, err := parseAPIHeaders(headerFlags)
	if err != nil {
		return handleError(cmd, err)
	}

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	req := notion.RawRequest{Method: method, Path: path, Params: params, Body: body, Headers: headers}
	p := output.NewPrinter(outputFormat(cmd))

	if !paginate {
		result, err := client.Request(cmd.Context(), req)
		if err != nil {
			return handleError(cmd, err)
		}
		p.PrintSuccess(result, "api", start)
		return nil
	}

	var results []any
	for page := 1; ; page++ {
		if page > maxPaginationPages {
			fmt.Fprintf(os.Stderr, "Warning: reached maximum pagination limit (%d pages)\n", maxPaginationPages)
			break
		}
		result, err := client.Request(cmd.Context(), req)
		if err != nil {
			return handleError(cmd, err)
		}
		pageResults, isList := result["results"].([]any)
		if !isList && page == 1 {
			// Not a paginated endpoint: print the response rather than dropping it.
			p.PrintSuccess(result, "api", start)
			return nil
		}
		results = append(results, pageResults...)

		hasMore, _ := result["has_more"].(bool)
		nextCursor, _ := result["next_cursor"].(string)
		if !hasMore || nextCursor == "" {
			break
		}
		// Notion reads the cursor from the query on GET and from the body on POST.
		if method == http.MethodGet {
			req.Params.Set("start_cursor", nextCursor)
		} else {
			if req.Body == nil {
				req.Body = map[string]any{}
			}
			req.Body["start_cursor"] = nextCursor
		}
	}
	p.PrintSuccess(map[string]any{
		"object":       "list",
		"results":      results,
		"result_count": len(results),
		"has_more":     false,
	}, "api", start)
	return nil
}

// parseAPIEndpoint normalises a user-supplied endpoint into a path relative to
// the API base plus any query string it carried. Absolute URLs are rejected so
// the bearer token can only ever be sent to the configured API host.
func parseAPIEndpoint(endpoint string) (string, url.Values, error) {
	endpoint = strings.TrimSpace(endpoint)
	u, err := url.Parse(endpoint)
	if err != nil || u.Scheme != "" || u.Host != "" || strings.HasPrefix(endpoint, "//") {
		return "", nil, invalidRequest(
			fmt.Sprintf("endpoint must be a path relative to /v1, such as users/me (got %q)", endpoint))
	}
	path := strings.Trim(u.Path, "/")
	path = strings.TrimPrefix(path, "v1/")
	if path == "" || path == "v1" {
		return "", nil, invalidRequest("endpoint path is empty")
	}
	return path, u.Query(), nil
}

// parseAPIBody reads the request body from --input: a file path or "-" for
// stdin. The body must be a JSON object, which is what every Notion endpoint
// accepts.
func parseAPIBody(input string, stdin io.Reader) (map[string]any, error) {
	if input == "" {
		return nil, nil
	}
	var raw []byte
	var err error
	if input == "-" {
		raw, err = io.ReadAll(stdin)
	} else {
		raw, err = os.ReadFile(input)
	}
	if err != nil {
		return nil, invalidRequest(fmt.Sprintf("read --input: %v", err))
	}
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, clierrors.InvalidJSON(fmt.Sprintf("--input must be a JSON object: %v", err))
	}
	return body, nil
}

// parseAPIFields turns -f (string) and -F (typed JSON) flags into a flat
// key→value map. Dots in keys are resolved later by setNestedField.
func parseAPIFields(rawFields, typedFields []string, stdin io.Reader) (map[string]any, error) {
	fields := map[string]any{}
	for _, f := range rawFields {
		k, v, ok := strings.Cut(f, "=")
		if !ok || k == "" {
			return nil, invalidRequest(fmt.Sprintf("-f expects key=value, got %q", f))
		}
		fields[k] = v
	}
	for _, f := range typedFields {
		k, v, ok := strings.Cut(f, "=")
		if !ok || k == "" {
			return nil, invalidRequest(fmt.Sprintf("-F expects key=value, got %q", f))
		}
		if strings.HasPrefix(v, "@") {
			var raw []byte
			var err error
			if v == "@-" {
				raw, err = io.ReadAll(stdin)
			} else {
				raw, err = os.ReadFile(v[1:])
			}
			if err != nil {
				return nil, invalidRequest(fmt.Sprintf("-F %s: %v", k, err))
			}
			v = string(raw)
		}
		var parsed any
		if err := json.Unmarshal([]byte(v), &parsed); err != nil {
			return nil, clierrors.InvalidJSON(fmt.Sprintf("-F %s: value is not valid JSON (use -f for plain strings): %v", k, err))
		}
		fields[k] = parsed
	}
	return fields, nil
}

// setNestedField assigns val at a dotted key path inside m, creating
// intermediate objects as needed: "parent.page_id" → m["parent"]["page_id"].
func setNestedField(m map[string]any, key string, val any) error {
	parts := strings.Split(key, ".")
	for _, part := range parts[:len(parts)-1] {
		child, ok := m[part].(map[string]any)
		if !ok {
			if _, exists := m[part]; exists {
				return invalidRequest(fmt.Sprintf("field %q: %q is not an object", key, part))
			}
			child = map[string]any{}
			m[part] = child
		}
		m = child
	}
	m[parts[len(parts)-1]] = val
	return nil
}

// parseAPIHeaders parses repeated "Name: value" flags.
func parseAPIHeaders(flags []string) (http.Header, error) {
	if len(flags) == 0 {
		return nil, nil
	}
	h := http.Header{}
	for _, f := range flags {
		name, value, ok := strings.Cut(f, ":")
		name = strings.TrimSpace(name)
		if !ok || name == "" {
			return nil, invalidRequest(fmt.Sprintf("-H expects 'Name: value', got %q", f))
		}
		h.Add(name, strings.TrimSpace(value))
	}
	return h, nil
}

// invalidRequest builds the INVALID_REQUEST error used for bad flag input.
func invalidRequest(msg string) *clierrors.NotionCLIError {
	return &clierrors.NotionCLIError{Code: clierrors.CodeInvalidRequest, Message: msg}
}
