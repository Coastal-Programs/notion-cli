package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

// Format represents an output format.
type Format string

const (
	FormatJSON        Format = "json"
	FormatCompactJSON Format = "compact-json"
	FormatRaw         Format = "raw"
	FormatTable       Format = "table"
	FormatCSV         Format = "csv"
	FormatMarkdown    Format = "markdown"
	FormatPretty      Format = "pretty"
)

// Printer handles formatted output for CLI commands.
type Printer struct {
	Format    Format
	Writer    io.Writer
	ErrWriter io.Writer
}

// NewPrinter creates a Printer with the given format.
// Output defaults to os.Stdout, errors to os.Stderr.
func NewPrinter(format Format) *Printer {
	return &Printer{
		Format:    format,
		Writer:    os.Stdout,
		ErrWriter: os.Stderr,
	}
}

// PrintSuccess outputs data wrapped in a success envelope.
// For JSON formats it serialises the envelope; for table/csv/markdown it extracts
// tabular data and formats it; for raw it prints data directly.
func (p *Printer) PrintSuccess(data any, command string, startTime time.Time) {
	elapsed := time.Since(startTime).Milliseconds()

	switch p.Format {
	case FormatJSON, FormatPretty:
		env := NewSuccessEnvelope(data, command, elapsed)
		b, err := json.MarshalIndent(env, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal output: %v\n", err)
			return
		}
		fmt.Fprintln(p.Writer, string(b))

	case FormatCompactJSON:
		env := NewSuccessEnvelope(data, command, elapsed)
		b, err := json.Marshal(env)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: failed to marshal output: %v\n", err)
			return
		}
		fmt.Fprintln(p.Writer, string(b))

	case FormatRaw:
		p.PrintRaw(data)

	case FormatCSV:
		headers, rows := ExtractTableData(data)
		fmt.Fprint(p.Writer, RenderCSV(headers, rows))

	case FormatMarkdown:
		headers, rows := ExtractTableData(data)
		fmt.Fprint(p.Writer, RenderMarkdown(headers, rows))

	default: // FormatTable and anything else
		headers, rows := ExtractTableData(data)
		fmt.Fprint(p.Writer, RenderTable(headers, rows))
	}
}

// PrintError outputs a structured error envelope to stderr.
func (p *Printer) PrintError(code, message string, details any, suggestions []string) {
	env := NewErrorEnvelope(code, message, details, suggestions, "")
	b, err := json.MarshalIndent(env, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal output: %v\n", err)
		return
	}
	fmt.Fprintln(p.ErrWriter, string(b))
}

// PrintRaw outputs data as JSON without an envelope wrapper.
func (p *Printer) PrintRaw(data any) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: failed to marshal output: %v\n", err)
		return
	}
	fmt.Fprintln(p.Writer, string(b))
}

// PrintTable is a convenience method that formats and prints a table.
func (p *Printer) PrintTable(headers []string, rows [][]string) {
	fmt.Fprint(p.Writer, RenderTable(headers, rows))
}

// ExtractTableData does best-effort extraction of tabular data from
// a map[string]any (single row) or []map[string]any (multiple rows).
// For other types it returns a single "value" column with the JSON representation.
func ExtractTableData(data any) (headers []string, rows [][]string) {
	switch v := data.(type) {
	case []map[string]any:
		if len(v) == 0 {
			return nil, nil
		}
		// Collect all keys across all rows for stable header set.
		seen := map[string]bool{}
		for _, m := range v {
			for k := range m {
				if !seen[k] {
					seen[k] = true
					headers = append(headers, k)
				}
			}
		}
		sort.Strings(headers)
		for _, m := range v {
			row := make([]string, len(headers))
			for i, h := range headers {
				if val, ok := m[h]; ok {
					row[i] = fmt.Sprintf("%v", val)
				}
			}
			rows = append(rows, row)
		}

	case map[string]any:
		for k := range v {
			headers = append(headers, k)
		}
		sort.Strings(headers)
		row := make([]string, len(headers))
		for i, h := range headers {
			row[i] = fmt.Sprintf("%v", v[h])
		}
		rows = [][]string{row}

	default:
		headers = []string{"value"}
		b, err := json.Marshal(data)
		if err != nil {
			rows = [][]string{{fmt.Sprintf("%v", data)}}
		} else {
			rows = [][]string{{string(b)}}
		}
	}
	return
}
