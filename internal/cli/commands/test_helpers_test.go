package commands

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"testing"
)

// captureStdout redirects os.Stdout for the duration of fn() and returns
// whatever was written to it. It is safe to call from parallel tests as long
// as each test restores os.Stdout on return (which this helper guarantees).
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	origStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe(): %v", err)
	}
	os.Stdout = w

	fn()

	_ = w.Close()
	os.Stdout = origStdout

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy: %v", err)
	}
	_ = r.Close()
	return buf.String()
}

// parseEnvelope parses a JSON envelope from s and returns the data map.
func parseEnvelope(t *testing.T, s string) map[string]any {
	t.Helper()
	var envelope map[string]any
	if err := json.Unmarshal([]byte(s), &envelope); err != nil {
		t.Fatalf("output not valid JSON: %v\noutput: %s", err, s)
	}
	return envelope
}
