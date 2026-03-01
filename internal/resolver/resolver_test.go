package resolver

import (
	"testing"
)

func TestExtractID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			"hyphenated UUID passthrough",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"raw 32-char hex",
			"8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"uppercase hex",
			"8C4D6E5FA1B23C4D5E6F7A8B9C0D1E2F",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"Notion URL with page title",
			"https://www.notion.so/My-Page-Title-8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"Notion URL with query params",
			"https://notion.so/workspace/8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f?v=abc123",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"Notion URL with hash",
			"https://notion.so/page-8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f#section",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"input with whitespace",
			"  8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f  ",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"notion.site URL",
			"https://mysite.notion.site/Page-8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			false,
		},
		{
			"empty input",
			"",
			"",
			true,
		},
		{
			"whitespace only",
			"   ",
			"",
			true,
		},
		{
			"invalid input",
			"not-an-id",
			"",
			true,
		},
		{
			"too short hex",
			"8c4d6e5f",
			"",
			true,
		},
		{
			"Notion URL without ID",
			"https://notion.so/just-a-page-title",
			"",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ExtractID(%q) expected error, got %q", tt.input, got)
				}
				return
			}
			if err != nil {
				t.Errorf("ExtractID(%q) unexpected error: %v", tt.input, err)
				return
			}
			if got != tt.want {
				t.Errorf("ExtractID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			"32 hex chars",
			"8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
		},
		{
			"already formatted",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
		},
		{
			"uppercase input",
			"8C4D6E5FA1B23C4D5E6F7A8B9C0D1E2F",
			"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f",
		},
		{
			"too short returns as-is (lowercased)",
			"abc123",
			"abc123",
		},
		{
			"not hex returns as-is (lowercased)",
			"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
			"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatID(tt.input)
			if got != tt.want {
				t.Errorf("FormatID(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidID(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f", true},
		{"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f", true},
		{"8C4D6E5FA1B23C4D5E6F7A8B9C0D1E2F", true},
		{"  8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f  ", true},
		{"abc123", false},
		{"not-a-valid-id", false},
		{"", false},
		{"zzzzzzzzzzzzzzzzzzzzzzzzzzzzzzzz", false},
		{"8c4d6e5f-a1b2-3c4d-5e6f", false}, // too short UUID
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := IsValidID(tt.input)
			if got != tt.want {
				t.Errorf("IsValidID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestStripHyphens(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f", "8c4d6e5fa1b23c4d5e6f7a8b9c0d1e2f"},
		{"no-hyphens-at-all", "nohyphensatall"},
		{"already-clean", "alreadyclean"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := StripHyphens(tt.input)
			if got != tt.want {
				t.Errorf("StripHyphens(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestExtractID_AllZeros(t *testing.T) {
	id, err := ExtractID("00000000000000000000000000000000")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "00000000-0000-0000-0000-000000000000"
	if id != want {
		t.Errorf("got %q, want %q", id, want)
	}
}

func TestExtractID_MixedCaseUUID(t *testing.T) {
	id, err := ExtractID("8C4D6E5F-A1B2-3C4D-5E6F-7A8B9C0D1E2F")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "8c4d6e5f-a1b2-3c4d-5e6f-7a8b9c0d1e2f"
	if id != want {
		t.Errorf("got %q, want %q", id, want)
	}
}
