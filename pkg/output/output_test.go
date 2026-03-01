package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// ─── Envelope Tests ──────────────────────────────────────────────────────────

func TestNewSuccessEnvelope(t *testing.T) {
	data := map[string]string{"id": "abc123"}
	env := NewSuccessEnvelope(data, "page retrieve", 42)

	if !env.Success {
		t.Fatal("expected Success=true")
	}
	if env.Metadata["command"] != "page retrieve" {
		t.Fatalf("expected command=page retrieve, got %v", env.Metadata["command"])
	}
	if env.Metadata["execution_time_ms"] != int64(42) {
		t.Fatalf("expected execution_time_ms=42, got %v", env.Metadata["execution_time_ms"])
	}
	if env.Metadata["version"] != Version {
		t.Fatalf("expected version=%s, got %v", Version, env.Metadata["version"])
	}
	ts, ok := env.Metadata["timestamp"].(string)
	if !ok || ts == "" {
		t.Fatal("expected non-empty timestamp string")
	}
	// Verify timestamp parses as RFC3339
	if _, err := time.Parse(time.RFC3339, ts); err != nil {
		t.Fatalf("timestamp not RFC3339: %v", err)
	}
}

func TestNewErrorEnvelope(t *testing.T) {
	suggestions := []string{"Check your token", "Run notion-cli init"}
	env := NewErrorEnvelope("AUTH_ERROR", "Unauthorized", nil, suggestions, "db query")

	if env.Success {
		t.Fatal("expected Success=false")
	}
	if env.Error.Code != "AUTH_ERROR" {
		t.Fatalf("expected code=AUTH_ERROR, got %s", env.Error.Code)
	}
	if env.Error.Message != "Unauthorized" {
		t.Fatalf("expected message=Unauthorized, got %s", env.Error.Message)
	}
	if len(env.Error.Suggestions) != 2 {
		t.Fatalf("expected 2 suggestions, got %d", len(env.Error.Suggestions))
	}
	if env.Metadata["command"] != "db query" {
		t.Fatalf("expected command=db query, got %v", env.Metadata["command"])
	}
}

func TestSuccessEnvelopeJSON(t *testing.T) {
	env := NewSuccessEnvelope("hello", "test", 0)
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if m["success"] != true {
		t.Fatal("expected success=true in JSON")
	}
	if m["data"] != "hello" {
		t.Fatalf("expected data=hello, got %v", m["data"])
	}
}

func TestErrorEnvelopeJSON(t *testing.T) {
	env := NewErrorEnvelope("NOT_FOUND", "Page not found", nil, nil, "page get")
	b, err := json.Marshal(env)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if m["success"] != false {
		t.Fatal("expected success=false in JSON")
	}
	errObj := m["error"].(map[string]any)
	if errObj["code"] != "NOT_FOUND" {
		t.Fatalf("expected code=NOT_FOUND, got %v", errObj["code"])
	}
	// details and suggestions should be omitted when nil
	if _, ok := errObj["details"]; ok {
		t.Fatal("expected details to be omitted")
	}
	if _, ok := errObj["suggestions"]; ok {
		t.Fatal("expected suggestions to be omitted")
	}
}

// ─── Table Formatter Tests ───────────────────────────────────────────────────

func TestRenderTableBasic(t *testing.T) {
	headers := []string{"Name", "Status"}
	rows := [][]string{
		{"Task1", "Done"},
		{"Task2", "In Progress"},
	}
	out := RenderTable(headers, rows)

	if !strings.Contains(out, "| Name  | Status      |") {
		t.Fatalf("missing header row, got:\n%s", out)
	}
	if !strings.Contains(out, "| Task1 | Done        |") {
		t.Fatalf("missing Task1 row, got:\n%s", out)
	}
	if !strings.Contains(out, "| Task2 | In Progress |") {
		t.Fatalf("missing Task2 row, got:\n%s", out)
	}
	// Should have separator lines
	if strings.Count(out, "+") < 6 {
		t.Fatalf("expected separator with + characters, got:\n%s", out)
	}
}

func TestRenderTableEmpty(t *testing.T) {
	if out := RenderTable(nil, nil); out != "" {
		t.Fatalf("expected empty string for nil headers, got %q", out)
	}
	if out := RenderTable([]string{}, nil); out != "" {
		t.Fatalf("expected empty string for empty headers, got %q", out)
	}
}

func TestRenderTableNoRows(t *testing.T) {
	out := RenderTable([]string{"Col"}, nil)
	// Should have header but no bottom separator (no data rows)
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	// Top sep, header, bottom sep for header = 3 lines, no trailing separator for empty data
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines for header-only table, got %d:\n%s", len(lines), out)
	}
}

func TestRenderTableShortRow(t *testing.T) {
	headers := []string{"A", "B", "C"}
	rows := [][]string{{"x"}} // row has fewer columns than headers
	out := RenderTable(headers, rows)
	if !strings.Contains(out, "| x |") {
		t.Fatalf("expected short row to be padded, got:\n%s", out)
	}
}

// ─── CSV Formatter Tests ────────────────────────────────────────────────────

func TestRenderCSVBasic(t *testing.T) {
	headers := []string{"Name", "Value"}
	rows := [][]string{{"a", "1"}, {"b", "2"}}
	out := RenderCSV(headers, rows)

	lines := strings.Split(strings.TrimRight(out, "\r\n"), "\r\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 CSV lines, got %d: %v", len(lines), lines)
	}
	if lines[0] != "Name,Value" {
		t.Fatalf("expected header Name,Value, got %q", lines[0])
	}
	if lines[1] != "a,1" {
		t.Fatalf("expected row a,1, got %q", lines[1])
	}
}

func TestRenderCSVQuoting(t *testing.T) {
	headers := []string{"Field"}
	rows := [][]string{
		{"has,comma"},
		{`has"quote`},
		{"has\nnewline"},
		{"plain"},
	}
	out := RenderCSV(headers, rows)

	if !strings.Contains(out, `"has,comma"`) {
		t.Fatalf("expected comma field to be quoted, got:\n%s", out)
	}
	if !strings.Contains(out, `"has""quote"`) {
		t.Fatalf("expected quote field to be double-quoted, got:\n%s", out)
	}
	if !strings.Contains(out, "\"has\nnewline\"") {
		t.Fatalf("expected newline field to be quoted, got:\n%s", out)
	}
	// plain should NOT be quoted
	if strings.Contains(out, `"plain"`) {
		t.Fatalf("expected plain field to NOT be quoted, got:\n%s", out)
	}
}

func TestRenderCSVPadding(t *testing.T) {
	headers := []string{"A", "B"}
	rows := [][]string{{"x"}} // short row
	out := RenderCSV(headers, rows)
	lines := strings.Split(strings.TrimRight(out, "\r\n"), "\r\n")
	if lines[1] != "x," {
		t.Fatalf("expected short row padded to x,, got %q", lines[1])
	}
}

// ─── Markdown Formatter Tests ───────────────────────────────────────────────

func TestRenderMarkdownBasic(t *testing.T) {
	headers := []string{"Name", "Status"}
	rows := [][]string{{"Task", "Done"}}
	out := RenderMarkdown(headers, rows)

	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("expected 3 markdown lines, got %d:\n%s", len(lines), out)
	}
	// Header line
	if !strings.Contains(lines[0], "| Name") && !strings.Contains(lines[0], "| Status") {
		t.Fatalf("missing header, got %q", lines[0])
	}
	// Separator line
	if !strings.Contains(lines[1], "---") {
		t.Fatalf("expected --- separator, got %q", lines[1])
	}
	// Data line
	if !strings.Contains(lines[2], "Task") {
		t.Fatalf("missing data, got %q", lines[2])
	}
}

func TestRenderMarkdownEmpty(t *testing.T) {
	if out := RenderMarkdown(nil, nil); out != "" {
		t.Fatalf("expected empty for nil headers, got %q", out)
	}
}

func TestRenderMarkdownMinWidth(t *testing.T) {
	headers := []string{"A"} // shorter than 3
	rows := [][]string{{"x"}}
	out := RenderMarkdown(headers, rows)
	// Separator should be at least "---" (3 dashes)
	if !strings.Contains(out, "---") {
		t.Fatalf("expected minimum 3-dash separator, got:\n%s", out)
	}
}

// ─── Printer Tests ──────────────────────────────────────────────────────────

func TestNewPrinter(t *testing.T) {
	p := NewPrinter(FormatJSON)
	if p.Format != FormatJSON {
		t.Fatalf("expected FormatJSON, got %s", p.Format)
	}
	if p.Writer == nil || p.ErrWriter == nil {
		t.Fatal("expected non-nil writers")
	}
}

func TestPrinterPrintSuccessJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatJSON, Writer: &buf, ErrWriter: &bytes.Buffer{}}

	data := map[string]string{"name": "Test"}
	p.PrintSuccess(data, "test cmd", time.Now())

	var env map[string]any
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("output is not valid JSON: %v\noutput: %s", err, buf.String())
	}
	if env["success"] != true {
		t.Fatal("expected success=true")
	}
	d := env["data"].(map[string]any)
	if d["name"] != "Test" {
		t.Fatalf("expected data.name=Test, got %v", d["name"])
	}
}

func TestPrinterPrintSuccessCompactJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatCompactJSON, Writer: &buf, ErrWriter: &bytes.Buffer{}}

	p.PrintSuccess("hello", "test", time.Now())

	out := strings.TrimSpace(buf.String())
	// Compact JSON should be one line
	if strings.Contains(out, "\n") {
		t.Fatalf("compact JSON should be one line, got:\n%s", out)
	}
	var env map[string]any
	if err := json.Unmarshal([]byte(out), &env); err != nil {
		t.Fatalf("not valid JSON: %v", err)
	}
}

func TestPrinterPrintSuccessPretty(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatPretty, Writer: &buf, ErrWriter: &bytes.Buffer{}}

	p.PrintSuccess("data", "cmd", time.Now())

	// Pretty is indented JSON
	if !strings.Contains(buf.String(), "  ") {
		t.Fatalf("pretty JSON should be indented, got:\n%s", buf.String())
	}
}

func TestPrinterPrintSuccessRaw(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatRaw, Writer: &buf, ErrWriter: &bytes.Buffer{}}

	data := map[string]string{"id": "123"}
	p.PrintSuccess(data, "cmd", time.Now())

	var m map[string]any
	if err := json.Unmarshal(buf.Bytes(), &m); err != nil {
		t.Fatalf("raw output not valid JSON: %v", err)
	}
	// Raw should NOT have envelope fields
	if _, ok := m["success"]; ok {
		t.Fatal("raw output should not have success field")
	}
	if m["id"] != "123" {
		t.Fatalf("expected id=123, got %v", m["id"])
	}
}

func TestPrinterPrintSuccessTable(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatTable, Writer: &buf, ErrWriter: &bytes.Buffer{}}

	data := []map[string]any{{"Name": "Task1", "Status": "Done"}}
	p.PrintSuccess(data, "cmd", time.Now())

	out := buf.String()
	if !strings.Contains(out, "Task1") {
		t.Fatalf("table output missing data, got:\n%s", out)
	}
	if !strings.Contains(out, "+") {
		t.Fatalf("expected table borders, got:\n%s", out)
	}
}

func TestPrinterPrintSuccessCSV(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatCSV, Writer: &buf, ErrWriter: &bytes.Buffer{}}

	data := []map[string]any{{"A": "1"}}
	p.PrintSuccess(data, "cmd", time.Now())

	out := buf.String()
	if !strings.Contains(out, "A") {
		t.Fatalf("CSV output missing header, got:\n%s", out)
	}
}

func TestPrinterPrintSuccessMarkdown(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatMarkdown, Writer: &buf, ErrWriter: &bytes.Buffer{}}

	data := []map[string]any{{"Col": "Val"}}
	p.PrintSuccess(data, "cmd", time.Now())

	out := buf.String()
	if !strings.Contains(out, "---") {
		t.Fatalf("markdown output missing separator, got:\n%s", out)
	}
}

func TestPrinterPrintError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	p := &Printer{Format: FormatJSON, Writer: &stdout, ErrWriter: &stderr}

	p.PrintError("RATE_LIMIT", "Too many requests", nil, []string{"Wait and retry"})

	if stdout.Len() > 0 {
		t.Fatal("error output should go to stderr, not stdout")
	}

	var env map[string]any
	if err := json.Unmarshal(stderr.Bytes(), &env); err != nil {
		t.Fatalf("error output not valid JSON: %v\noutput: %s", err, stderr.String())
	}
	if env["success"] != false {
		t.Fatal("expected success=false")
	}
	errObj := env["error"].(map[string]any)
	if errObj["code"] != "RATE_LIMIT" {
		t.Fatalf("expected code=RATE_LIMIT, got %v", errObj["code"])
	}
}

func TestPrinterPrintRaw(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatRaw, Writer: &buf, ErrWriter: &bytes.Buffer{}}

	p.PrintRaw(42)

	out := strings.TrimSpace(buf.String())
	if out != "42" {
		t.Fatalf("expected 42, got %q", out)
	}
}

func TestPrinterPrintTable(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatTable, Writer: &buf, ErrWriter: &bytes.Buffer{}}

	p.PrintTable([]string{"H"}, [][]string{{"V"}})
	if !strings.Contains(buf.String(), "| H |") {
		t.Fatalf("expected table output, got:\n%s", buf.String())
	}
}

// ─── ExtractTableData Tests ─────────────────────────────────────────────────

func TestExtractTableDataSliceOfMaps(t *testing.T) {
	data := []map[string]any{
		{"Name": "A", "Val": 1},
		{"Name": "B", "Val": 2},
	}
	headers, rows := ExtractTableData(data)

	if len(headers) != 2 {
		t.Fatalf("expected 2 headers, got %d", len(headers))
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
}

func TestExtractTableDataEmptySlice(t *testing.T) {
	data := []map[string]any{}
	headers, rows := ExtractTableData(data)
	if headers != nil || rows != nil {
		t.Fatal("expected nil for empty slice")
	}
}

func TestExtractTableDataSingleMap(t *testing.T) {
	data := map[string]any{"key": "value"}
	headers, rows := ExtractTableData(data)

	if len(headers) != 1 || headers[0] != "key" {
		t.Fatalf("expected [key], got %v", headers)
	}
	if len(rows) != 1 || rows[0][0] != "value" {
		t.Fatalf("expected [[value]], got %v", rows)
	}
}

func TestExtractTableDataOtherType(t *testing.T) {
	headers, rows := ExtractTableData(42)

	if len(headers) != 1 || headers[0] != "value" {
		t.Fatalf("expected [value], got %v", headers)
	}
	if len(rows) != 1 || rows[0][0] != "42" {
		t.Fatalf("expected [[42]], got %v", rows)
	}
}

func TestExtractTableDataString(t *testing.T) {
	headers, rows := ExtractTableData("hello world")

	if len(headers) != 1 || headers[0] != "value" {
		t.Fatalf("expected [value], got %v", headers)
	}
	if len(rows) != 1 || rows[0][0] != `"hello world"` {
		t.Fatalf("expected JSON-quoted string, got %v", rows)
	}
}

// ─── Integration Tests ──────────────────────────────────────────────────────

func TestRoundTripSuccessEnvelopeJSON(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: FormatJSON, Writer: &buf, ErrWriter: &bytes.Buffer{}}

	input := []map[string]any{
		{"id": "page-1", "title": "My Page"},
		{"id": "page-2", "title": "Other Page"},
	}
	start := time.Now().Add(-100 * time.Millisecond)
	p.PrintSuccess(input, "page list", start)

	var env SuccessEnvelope
	if err := json.Unmarshal(buf.Bytes(), &env); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !env.Success {
		t.Fatal("expected success=true")
	}
	if env.Metadata["command"] != "page list" {
		t.Fatalf("expected command=page list, got %v", env.Metadata["command"])
	}
	execMs, ok := env.Metadata["execution_time_ms"].(float64)
	if !ok || execMs < 0 {
		t.Fatalf("expected positive execution_time_ms, got %v", env.Metadata["execution_time_ms"])
	}
}

func TestCSVQuoteEdgeCases(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "simple"},
		{"with,comma", `"with,comma"`},
		{`with"quote`, `"with""quote"`},
		{"with\nnewline", "\"with\nnewline\""},
		{"with\r\nCRLF", "\"with\r\nCRLF\""},
		{"", ""},
	}
	for _, tc := range tests {
		got := csvQuote(tc.input)
		if got != tc.expected {
			t.Errorf("csvQuote(%q) = %q, want %q", tc.input, got, tc.expected)
		}
	}
}

func TestRenderTableUnicode(t *testing.T) {
	headers := []string{"Name"}
	rows := [][]string{{"cafe\u0301"}} // "cafe" + combining accent (5 runes)
	out := RenderTable(headers, rows)
	if !strings.Contains(out, "cafe\u0301") {
		t.Fatalf("expected unicode content, got:\n%s", out)
	}
}

func TestDefaultFormatFallback(t *testing.T) {
	var buf bytes.Buffer
	p := &Printer{Format: "unknown", Writer: &buf, ErrWriter: &bytes.Buffer{}}

	data := []map[string]any{{"x": "y"}}
	p.PrintSuccess(data, "cmd", time.Now())

	// Unknown format should fall back to table
	out := buf.String()
	if !strings.Contains(out, "+") {
		t.Fatalf("unknown format should fall back to table, got:\n%s", out)
	}
}
