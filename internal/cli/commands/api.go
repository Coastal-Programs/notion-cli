package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/internal/notion"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
	"github.com/spf13/cobra"
)

// RegisterAPICommands registers the ntn-compatible generic API command.
func RegisterAPICommands(root *cobra.Command) {
	apiCmd := &cobra.Command{
		Use:   "api <path> [field=value ...]",
		Short: "Make authenticated Notion API requests",
		Long: "Make authenticated Notion API requests with ntn-compatible inline body, query, and header syntax.\n\n" +
			"Forms:\n" +
			"  path=value       body field with string value\n" +
			"  path:=json       body field parsed as JSON\n" +
			"  name==value      query parameter\n" +
			"  Header:Value     request header",
		Args: cobra.MinimumNArgs(1),
		RunE: runAPI,
	}
	apiCmd.Flags().StringP("method", "X", "", "HTTP method")
	apiCmd.Flags().String("data", "", "JSON request body")
	apiCmd.Flags().String("file", "", "Send a multipart form-data request with this file as field \"file\"")
	apiCmd.Flags().Bool("spec", false, "Print the embedded reduced OpenAPI fragment for this endpoint")
	apiCmd.Flags().Bool("docs", false, "Print the embedded markdown docs for this endpoint")
	apiCmd.Flags().Bool("unsafe-verbose", false, "Do not redact Authorization in verbose logs")
	_ = apiCmd.Flags().MarkHidden("unsafe-verbose")
	addOutputFlags(apiCmd)

	lsCmd := &cobra.Command{
		Use:   "ls",
		Short: "List embedded Notion API endpoints",
		Args:  cobra.NoArgs,
		RunE:  runAPILS,
	}
	addOutputFlags(lsCmd)
	apiCmd.AddCommand(lsCmd)

	root.AddCommand(apiCmd)
}

func runAPILS(cmd *cobra.Command, _ []string) error {
	start := time.Now()
	p := output.NewPrinter(outputFormat(cmd))
	p.Writer = cmd.OutOrStdout()
	p.PrintSuccess(apiCatalogRows(), "api ls", start)
	return nil
}

func runAPI(cmd *cobra.Command, args []string) error {
	start := time.Now()
	path := args[0]
	inlineArgs := args[1:]

	method, _ := cmd.Flags().GetString("method")
	method = strings.ToUpper(strings.TrimSpace(method))
	methodExplicit := method != ""

	if spec, _ := cmd.Flags().GetBool("spec"); spec {
		return printAPISpec(cmd, path, method, methodExplicit)
	}
	if docs, _ := cmd.Flags().GetBool("docs"); docs {
		return printAPIDocs(cmd, path, method, methodExplicit)
	}

	parsed, err := parseAPIInlineArgs(inlineArgs)
	if err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{Code: clierrors.CodeInvalidRequest, Message: err.Error()})
	}

	body, contentType, err := buildAPIPayload(cmd, parsed)
	if err != nil {
		return handleError(cmd, err)
	}
	if method == "" {
		if len(body) > 0 {
			method = http.MethodPost
		} else {
			method = http.MethodGet
		}
	}

	client, err := newClientForCommand(cmd)
	if err != nil {
		return handleError(cmd, err)
	}

	if apiVerbose(cmd) {
		printAPIRequestDebug(cmd, method, path, parsed.Headers, body, contentType)
	}

	resp, err := client.Raw(cmd.Context(), notion.RawRequest{
		Method:      method,
		Path:        path,
		Query:       parsed.Query,
		Headers:     parsed.Headers,
		Body:        body,
		ContentType: contentType,
	})
	if err != nil {
		return handleError(cmd, err)
	}
	if apiVerbose(cmd) {
		printAPIResponseDebug(cmd, resp)
	}
	return printAPIResponse(cmd, resp, start)
}

func printAPISpec(cmd *cobra.Command, path, method string, methodExplicit bool) error {
	endpoint, err := findCatalogEndpoint(path, method, methodExplicit)
	if err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{Code: clierrors.CodeNotFound, Message: err.Error()})
	}
	b, err := endpointSpecJSON(endpoint)
	if err != nil {
		return handleError(cmd, clierrors.Wrap(clierrors.CodeInternalError, "Failed to render endpoint spec", err))
	}
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), string(b))
	return nil
}

func printAPIDocs(cmd *cobra.Command, path, method string, methodExplicit bool) error {
	endpoint, err := findCatalogEndpoint(path, method, methodExplicit)
	if err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{Code: clierrors.CodeNotFound, Message: err.Error()})
	}
	_, _ = fmt.Fprint(cmd.OutOrStdout(), endpoint.Docs)
	return nil
}

func buildAPIPayload(cmd *cobra.Command, parsed *parsedAPIInline) ([]byte, string, error) {
	data, _ := cmd.Flags().GetString("data")
	filePath, _ := cmd.Flags().GetString("file")
	hasInlineBody := len(parsed.BodyFields) > 0
	stdinBody, stdinHasBody, err := readAPIStdin(cmd)
	if err != nil {
		return nil, "", err
	}

	bodySources := 0
	if strings.TrimSpace(data) != "" {
		bodySources++
	}
	if stdinHasBody {
		bodySources++
	}
	if hasInlineBody && filePath == "" {
		bodySources++
	}
	if filePath != "" {
		bodySources++
	}
	if bodySources > 1 {
		return nil, "", &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: "Use only one request body source: stdin JSON, --data, inline body fields, or --file",
		}
	}

	switch {
	case filePath != "":
		formFields, err := apiFieldsAsFormFields(parsed.BodyFields)
		if err != nil {
			return nil, "", clierrors.Wrap(clierrors.CodeInvalidRequest, "Invalid multipart form fields", err)
		}
		return buildAPIMultipart(filePath, formFields)
	case strings.TrimSpace(data) != "":
		if err := validateJSONBody([]byte(data)); err != nil {
			return nil, "", err
		}
		return []byte(data), "application/json", nil
	case stdinHasBody:
		if err := validateJSONBody(stdinBody); err != nil {
			return nil, "", err
		}
		return stdinBody, "application/json", nil
	case hasInlineBody:
		body, err := buildAPIBody(parsed.BodyFields)
		if err != nil {
			return nil, "", &clierrors.NotionCLIError{Code: clierrors.CodeInvalidRequest, Message: err.Error()}
		}
		b, err := json.Marshal(body)
		if err != nil {
			return nil, "", clierrors.Wrap(clierrors.CodeInternalError, "Failed to encode request body", err)
		}
		return b, "application/json", nil
	default:
		return nil, "", nil
	}
}

func readAPIStdin(cmd *cobra.Command) ([]byte, bool, error) {
	in := cmd.InOrStdin()
	if file, ok := in.(*os.File); ok {
		stat, err := file.Stat()
		if err != nil {
			return nil, false, clierrors.Wrap(clierrors.CodeInternalError, "Failed to stat stdin", err)
		}
		if stat.Mode()&os.ModeCharDevice != 0 {
			return nil, false, nil
		}
	}
	b, err := io.ReadAll(in)
	if err != nil {
		return nil, false, clierrors.Wrap(clierrors.CodeInternalError, "Failed to read stdin", err)
	}
	if strings.TrimSpace(string(b)) == "" {
		return nil, false, nil
	}
	return b, true, nil
}

func validateJSONBody(body []byte) error {
	var v any
	if err := json.Unmarshal(body, &v); err != nil {
		return clierrors.InvalidJSON(err.Error())
	}
	return nil
}

func buildAPIMultipart(path string, fields map[string]string) ([]byte, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, "", &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: fmt.Sprintf("Cannot read file %q: %s", path, err),
		}
	}
	var buf bytes.Buffer
	mw := multipart.NewWriter(&buf)
	for k, v := range fields {
		if err := mw.WriteField(k, v); err != nil {
			return nil, "", clierrors.Wrap(clierrors.CodeInternalError, "Failed to write multipart field", err)
		}
	}
	part, err := mw.CreateFormFile("file", filepath.Base(path))
	if err != nil {
		return nil, "", clierrors.Wrap(clierrors.CodeInternalError, "Failed to create multipart file field", err)
	}
	if _, err := part.Write(data); err != nil {
		return nil, "", clierrors.Wrap(clierrors.CodeInternalError, "Failed to write multipart file bytes", err)
	}
	if err := mw.Close(); err != nil {
		return nil, "", clierrors.Wrap(clierrors.CodeInternalError, "Failed to close multipart body", err)
	}
	return buf.Bytes(), mw.FormDataContentType(), nil
}

func printAPIResponse(cmd *cobra.Command, resp *notion.RawResponse, start time.Time) error {
	if outputFormatExplicit(cmd) && outputFormat(cmd) != output.FormatRaw {
		var decoded any
		if len(resp.Body) > 0 {
			if err := json.Unmarshal(resp.Body, &decoded); err != nil {
				decoded = string(resp.Body)
			}
		} else {
			decoded = map[string]any{}
		}
		p := output.NewPrinter(outputFormat(cmd))
		p.Writer = cmd.OutOrStdout()
		p.PrintSuccess(decoded, "api", start)
		return nil
	}
	if len(resp.Body) > 0 {
		_, _ = cmd.OutOrStdout().Write(resp.Body)
		if resp.Body[len(resp.Body)-1] != '\n' {
			_, _ = fmt.Fprintln(cmd.OutOrStdout())
		}
	}
	return nil
}

func apiVerbose(cmd *cobra.Command) bool {
	v, _ := cmd.Flags().GetBool("verbose")
	return v
}

func printAPIRequestDebug(cmd *cobra.Command, method, path string, headers http.Header, body []byte, contentType string) {
	unsafe, _ := cmd.Flags().GetBool("unsafe-verbose")
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "> %s %s\n", method, path)
	_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "> Authorization:", redactedAuth(unsafe))
	if version, _ := cmd.Flags().GetString("notion-version"); version != "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "> Notion-Version:", version)
	}
	if contentType != "" {
		_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "> Content-Type:", contentType)
	}
	for k, values := range headers {
		for _, v := range values {
			if strings.EqualFold(k, "Authorization") && !unsafe {
				v = "<redacted>"
			}
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "> %s: %s\n", k, v)
		}
	}
	if len(body) > 0 {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "> Body: %s\n", string(body))
	}
}

func printAPIResponseDebug(cmd *cobra.Command, resp *notion.RawResponse) {
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "< Status: %d\n", resp.StatusCode)
	for k, values := range resp.Headers {
		for _, v := range values {
			_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "< %s: %s\n", k, v)
		}
	}
	if len(resp.Body) > 0 {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "< Body: %s\n", string(resp.Body))
	}
}

func redactedAuth(unsafe bool) string {
	if unsafe {
		return "<not redacted>"
	}
	return "<redacted>"
}
