package output

import (
	"fmt"
	"strings"
	"unicode/utf8"
)

// RenderTable renders headers and rows as an ASCII table with borders.
//
//	+-------+--------+
//	| Name  | Status |
//	+-------+--------+
//	| Task1 | Done   |
//	+-------+--------+
func RenderTable(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	colCount := len(headers)
	widths := make([]int, colCount)

	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
	}
	for _, row := range rows {
		for i := 0; i < colCount && i < len(row); i++ {
			if w := utf8.RuneCountInString(row[i]); w > widths[i] {
				widths[i] = w
			}
		}
	}

	var b strings.Builder
	writeSep := func() {
		b.WriteByte('+')
		for _, w := range widths {
			b.WriteString(strings.Repeat("-", w+2))
			b.WriteByte('+')
		}
		b.WriteByte('\n')
	}
	writeRow := func(cells []string) {
		b.WriteByte('|')
		for i := 0; i < colCount; i++ {
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			pad := widths[i] - utf8.RuneCountInString(cell)
			b.WriteString(fmt.Sprintf(" %s%s |", cell, strings.Repeat(" ", pad)))
		}
		b.WriteByte('\n')
	}

	writeSep()
	writeRow(headers)
	writeSep()
	for _, row := range rows {
		writeRow(row)
	}
	if len(rows) > 0 {
		writeSep()
	}

	return b.String()
}

// RenderCSV renders headers and rows as RFC 4180 CSV.
func RenderCSV(headers []string, rows [][]string) string {
	var b strings.Builder
	writeCSVRow := func(fields []string) {
		for i, f := range fields {
			if i > 0 {
				b.WriteByte(',')
			}
			b.WriteString(csvQuote(f))
		}
		b.WriteString("\r\n")
	}

	writeCSVRow(headers)
	for _, row := range rows {
		// Pad or trim to match header count
		cells := make([]string, len(headers))
		for i := 0; i < len(headers); i++ {
			if i < len(row) {
				cells[i] = row[i]
			}
		}
		writeCSVRow(cells)
	}
	return b.String()
}

// csvQuote quotes a field per RFC 4180: fields containing commas, double-quotes,
// or newlines are wrapped in double-quotes, with internal quotes doubled.
func csvQuote(field string) string {
	if strings.ContainsAny(field, ",\"\r\n") {
		return "\"" + strings.ReplaceAll(field, "\"", "\"\"") + "\""
	}
	return field
}

// RenderMarkdown renders headers and rows as a GitHub-flavored markdown table.
//
//	| Name  | Status |
//	| ----- | ------ |
//	| Task1 | Done   |
func RenderMarkdown(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}

	colCount := len(headers)
	widths := make([]int, colCount)

	for i, h := range headers {
		widths[i] = utf8.RuneCountInString(h)
		if widths[i] < 3 { // minimum "---" separator
			widths[i] = 3
		}
	}
	for _, row := range rows {
		for i := 0; i < colCount && i < len(row); i++ {
			if w := utf8.RuneCountInString(row[i]); w > widths[i] {
				widths[i] = w
			}
		}
	}

	var b strings.Builder
	writeRow := func(cells []string) {
		b.WriteByte('|')
		for i := 0; i < colCount; i++ {
			cell := ""
			if i < len(cells) {
				cell = cells[i]
			}
			pad := widths[i] - utf8.RuneCountInString(cell)
			b.WriteString(fmt.Sprintf(" %s%s |", cell, strings.Repeat(" ", pad)))
		}
		b.WriteByte('\n')
	}

	writeRow(headers)
	// Separator row
	b.WriteByte('|')
	for _, w := range widths {
		b.WriteString(fmt.Sprintf(" %s |", strings.Repeat("-", w)))
	}
	b.WriteByte('\n')

	for _, row := range rows {
		writeRow(row)
	}

	return b.String()
}
