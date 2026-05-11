// Package commands implements the Notion CLI subcommands.
//
// This file implements the `files` command group, which exposes the Notion
// File Uploads API:
//
//	POST   /v1/file_uploads
//	POST   /v1/file_uploads/{id}/send
//	POST   /v1/file_uploads/{id}/complete
//	GET    /v1/file_uploads/{id}
//	GET    /v1/file_uploads
//
// Single-part uploads (≤20 MB): create → send.
// Multi-part uploads (>20 MB): create with mode=multi_part → send each chunk
// (1-indexed part_number) → complete.
package commands

import (
	"fmt"
	"io"
	"math"
	"mime"
	"os"
	"path/filepath"
	"strings"
	"time"

	clierrors "github.com/Coastal-Programs/notion-cli/v6/internal/errors"
	"github.com/Coastal-Programs/notion-cli/v6/internal/notion"
	"github.com/Coastal-Programs/notion-cli/v6/pkg/output"
	"github.com/spf13/cobra"
)

// singlePartMaxBytes is the threshold below which single-part upload is used.
// It is a variable so tests can override it.
var singlePartMaxBytes int64 = 20 * 1024 * 1024 // 20 MB

// isStderrTTY reports whether stderr is an interactive terminal. It is a
// variable so tests can override it without patching os.Stderr.
var isStderrTTY = func() bool {
	stat, err := os.Stderr.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}

// RegisterFilesCommands registers `files upload`, `files retrieve`, and
// `files list` under root.
func RegisterFilesCommands(root *cobra.Command) {
	filesCmd := &cobra.Command{
		Use:   "files",
		Short: "File upload operations",
		Long:  "Upload, retrieve, and list Notion file uploads.",
	}

	filesCmd.AddCommand(
		newFilesUploadCmd(),
		newFilesRetrieveCmd(),
		newFilesListCmd(),
	)

	root.AddCommand(filesCmd)
}

// ---------------------------------------------------------------------------
// files upload
// ---------------------------------------------------------------------------

func newFilesUploadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload <path>",
		Short: "Upload a file",
		Long: `Upload a local file to the Notion File Uploads API.

Files ≤20 MB are uploaded in a single part. Larger files are split into
chunks and uploaded using the multi-part API. The file_upload ID is printed
to stdout for piping; use --json for the full response envelope.`,
		Args: cobra.ExactArgs(1),
		RunE: runFilesUpload,
	}

	cmd.Flags().String("chunk-size", "10MB", "Chunk size for multi-part uploads (5MB–20MB)")

	addOutputFlags(cmd)
	return cmd
}

// parseChunkSize parses a human-readable chunk size string (e.g. "10MB") and
// returns the number of bytes. Accepts values in the range [5MB, 20MB].
func parseChunkSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	var multiplier int64
	switch {
	case strings.HasSuffix(s, "MB"):
		multiplier = 1024 * 1024
		s = strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "KB"):
		multiplier = 1024
		s = strings.TrimSuffix(s, "KB")
	default:
		return 0, fmt.Errorf("unrecognised chunk size unit in %q; use MB (e.g. 10MB)", s)
	}

	var n int64
	if _, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n); err != nil {
		return 0, fmt.Errorf("cannot parse chunk size %q: %w", s, err)
	}
	bytes := n * multiplier
	const minBytes = 5 * 1024 * 1024
	const maxBytes = 20 * 1024 * 1024
	if bytes < minBytes || bytes > maxBytes {
		return 0, fmt.Errorf("chunk size %q is out of range; must be between 5MB and 20MB", s)
	}
	return bytes, nil
}

// detectContentType returns the MIME type for the given filename, falling back
// to application/octet-stream for unknown extensions.
func detectContentType(filename string) string {
	ct := mime.TypeByExtension(filepath.Ext(filename))
	if ct == "" {
		return "application/octet-stream"
	}
	return ct
}

// progressBar writes a simple progress line to stderr when stderr is a TTY.
// It overwrites the current line using a carriage return.
func progressBar(done, total int64) {
	if total <= 0 {
		return
	}
	pct := float64(done) / float64(total) * 100
	barWidth := 30
	filled := int(math.Round(float64(barWidth) * float64(done) / float64(total)))
	if filled > barWidth {
		filled = barWidth
	}
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", barWidth-filled)
	doneMB := float64(done) / (1024 * 1024)
	totalMB := float64(total) / (1024 * 1024)
	fmt.Fprintf(os.Stderr, "\r[%s] %3.0f%%  %.1f / %.1f MB", bar, pct, doneMB, totalMB)
}

func runFilesUpload(cmd *cobra.Command, args []string) error {
	start := time.Now()
	path := args[0]

	chunkSizeStr, _ := cmd.Flags().GetString("chunk-size")
	chunkSize, err := parseChunkSize(chunkSizeStr)
	if err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: err.Error(),
			Suggestions: []string{
				"Provide a chunk size between 5MB and 20MB, e.g. --chunk-size 10MB",
			},
		})
	}

	info, err := os.Stat(path)
	if err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: fmt.Sprintf("Cannot stat file %q: %v", path, err),
			Suggestions: []string{
				"Verify the path exists and is readable",
			},
		})
	}

	fileSize := info.Size()
	base := filepath.Base(path)
	ct := detectContentType(base)

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	ctx := cmd.Context()
	showProgress := isStderrTTY()

	if fileSize <= singlePartMaxBytes {
		// --- single-part upload ---
		createResult, err := client.FileUploadCreate(ctx, map[string]any{
			"mode":         "single_part",
			"filename":     base,
			"content_type": ct,
		})
		if err != nil {
			return handleError(cmd, err)
		}

		id, _ := createResult["id"].(string)
		if id == "" {
			return handleError(cmd, &clierrors.NotionCLIError{
				Code:    clierrors.CodeInternalError,
				Message: "File upload create response missing id",
			})
		}

		data, err := os.ReadFile(path)
		if err != nil {
			return handleError(cmd, &clierrors.NotionCLIError{
				Code:    clierrors.CodeInvalidRequest,
				Message: fmt.Sprintf("Cannot read file %q: %v", path, err),
			})
		}

		if showProgress {
			progressBar(int64(len(data)), fileSize)
		}

		sendResult, err := client.FileUploadSend(ctx, id, 0, base, ct, data)
		if err != nil {
			return handleError(cmd, err)
		}

		if showProgress {
			progressBar(fileSize, fileSize)
			_, _ = fmt.Fprintln(os.Stderr)
		}

		return printUploadResult(cmd, sendResult, start)
	}

	// --- multi-part upload ---
	numParts := int((fileSize + chunkSize - 1) / chunkSize)

	createResult, err := client.FileUploadCreate(ctx, map[string]any{
		"mode":            "multi_part",
		"filename":        base,
		"content_type":    ct,
		"number_of_parts": numParts,
	})
	if err != nil {
		return handleError(cmd, err)
	}

	id, _ := createResult["id"].(string)
	if id == "" {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInternalError,
			Message: "File upload create response missing id",
		})
	}

	f, err := os.Open(path)
	if err != nil {
		return handleError(cmd, &clierrors.NotionCLIError{
			Code:    clierrors.CodeInvalidRequest,
			Message: fmt.Sprintf("Cannot open file %q: %v", path, err),
		})
	}
	defer f.Close() //nolint:errcheck

	var bytesSent int64
	chunk := make([]byte, chunkSize)

	for i := 1; i <= numParts; i++ {
		n, err := io.ReadFull(f, chunk)
		if err != nil && err != io.ErrUnexpectedEOF && err != io.EOF {
			return handleError(cmd, &clierrors.NotionCLIError{
				Code:    clierrors.CodeInvalidRequest,
				Message: fmt.Sprintf("Error reading file %q at part %d: %v", path, i, err),
			})
		}
		if n == 0 {
			break
		}

		partData := chunk[:n]
		if _, err := client.FileUploadSend(ctx, id, i, base, ct, partData); err != nil {
			return handleError(cmd, err)
		}

		bytesSent += int64(n)
		if showProgress {
			progressBar(bytesSent, fileSize)
		}
	}

	completeResult, err := client.FileUploadComplete(ctx, id)
	if err != nil {
		return handleError(cmd, err)
	}

	if showProgress {
		_, _ = fmt.Fprintln(os.Stderr)
	}

	return printUploadResult(cmd, completeResult, start)
}

// printUploadResult prints the upload result. Without --json it prints only
// the bare file_upload ID (suitable for piping); with --json it prints the
// full envelope.
func printUploadResult(cmd *cobra.Command, result map[string]any, start time.Time) error {
	jsonFlag, _ := cmd.Flags().GetBool("json")
	if jsonFlag {
		p := output.NewPrinter(outputFormat(cmd))
		p.PrintSuccess(result, "files upload", start)
		return nil
	}

	// Bare ID for piping.
	id, _ := result["id"].(string)
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), id)
	return nil
}

// ---------------------------------------------------------------------------
// files retrieve
// ---------------------------------------------------------------------------

func newFilesRetrieveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "retrieve <file_upload_id>",
		Aliases: []string{"r", "get"},
		Short:   "Retrieve a file upload",
		Long:    "Retrieve the status and metadata of a Notion file upload by ID.",
		Args:    cobra.ExactArgs(1),
		RunE:    runFilesRetrieve,
	}

	addOutputFlags(cmd)
	return cmd
}

func runFilesRetrieve(cmd *cobra.Command, args []string) error {
	start := time.Now()

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	result, err := client.FileUploadRetrieve(cmd.Context(), args[0])
	if err != nil {
		return handleError(cmd, err)
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(result, "files retrieve", start)
	return nil
}

// ---------------------------------------------------------------------------
// files list
// ---------------------------------------------------------------------------

func newFilesListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"l", "ls"},
		Short:   "List file uploads",
		Long:    "List Notion file uploads with optional pagination.",
		RunE:    runFilesList,
	}

	cmd.Flags().Int("page-size", 50, "Number of results per page")
	cmd.Flags().Bool("all", false, "Paginate until all file uploads are retrieved")

	addOutputFlags(cmd)
	return cmd
}

func runFilesList(cmd *cobra.Command, _ []string) error {
	start := time.Now()

	pageSize, _ := cmd.Flags().GetInt("page-size")
	pageAll, _ := cmd.Flags().GetBool("all")

	client, err := newClient()
	if err != nil {
		return handleError(cmd, err)
	}

	ctx := cmd.Context()
	qp := notion.QueryParams{PageSize: pageSize}

	if !pageAll {
		result, err := client.FileUploadList(ctx, qp)
		if err != nil {
			return handleError(cmd, err)
		}
		p := output.NewPrinter(outputFormat(cmd))
		p.PrintSuccess(result, "files list", start)
		return nil
	}

	var allResults []any
	pageCount := 0
	for {
		pageCount++
		if pageCount > maxPaginationPages {
			fmt.Fprintf(os.Stderr, "Warning: reached maximum pagination limit (%d pages)\n", maxPaginationPages)
			break
		}

		result, err := client.FileUploadList(ctx, qp)
		if err != nil {
			return handleError(cmd, err)
		}

		results, _ := result["results"].([]any)
		allResults = append(allResults, results...)

		hasMore, _ := result["has_more"].(bool)
		nextCursor, _ := result["next_cursor"].(string)
		if !hasMore || nextCursor == "" {
			break
		}
		qp.StartCursor = nextCursor
	}

	data := map[string]any{
		"object":       "list",
		"results":      allResults,
		"has_more":     false,
		"next_cursor":  nil,
		"result_count": len(allResults),
	}

	p := output.NewPrinter(outputFormat(cmd))
	p.PrintSuccess(data, "files list", start)
	return nil
}
