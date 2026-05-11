package commands

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// page.go pure-function tests
// ---------------------------------------------------------------------------

func TestHeadingBlock_Level1(t *testing.T) {
	block := headingBlock(1, "My Title")
	if block["type"] != "heading_1" {
		t.Errorf("type = %v, want heading_1", block["type"])
	}
	h1, ok := block["heading_1"].(map[string]any)
	if !ok {
		t.Fatal("heading_1 key missing or wrong type")
	}
	rt, ok := h1["rich_text"].([]map[string]any)
	if !ok || len(rt) == 0 {
		t.Fatal("rich_text missing or empty")
	}
	text, _ := rt[0]["text"].(map[string]any)
	if text["content"] != "My Title" {
		t.Errorf("content = %v, want My Title", text["content"])
	}
}

func TestHeadingBlock_Level2(t *testing.T) {
	block := headingBlock(2, "Sub")
	if block["type"] != "heading_2" {
		t.Errorf("type = %v, want heading_2", block["type"])
	}
}

func TestHeadingBlock_Level3(t *testing.T) {
	block := headingBlock(3, "Sub-sub")
	if block["type"] != "heading_3" {
		t.Errorf("type = %v, want heading_3", block["type"])
	}
}

func TestSplitRichText_ShortString(t *testing.T) {
	segments := splitRichText("hello", 2000)
	if len(segments) != 1 {
		t.Fatalf("got %d segments, want 1", len(segments))
	}
	text, _ := segments[0]["text"].(map[string]any)
	if text["content"] != "hello" {
		t.Errorf("content = %v, want hello", text["content"])
	}
}

func TestSplitRichText_LongString(t *testing.T) {
	// Build a string longer than maxLen.
	long := strings.Repeat("a", 2500)
	segments := splitRichText(long, 2000)
	if len(segments) != 2 {
		t.Fatalf("got %d segments, want 2", len(segments))
	}
	first, _ := segments[0]["text"].(map[string]any)
	if len(first["content"].(string)) != 2000 {
		t.Errorf("first segment len = %d, want 2000", len(first["content"].(string)))
	}
	second, _ := segments[1]["text"].(map[string]any)
	if len(second["content"].(string)) != 500 {
		t.Errorf("second segment len = %d, want 500", len(second["content"].(string)))
	}
}

func TestRichText_ProducesSegments(t *testing.T) {
	rt := richText("hello world")
	if len(rt) == 0 {
		t.Fatal("richText should return at least one segment")
	}
}

func TestFileBaseName_WithExtension(t *testing.T) {
	if got := fileBaseName("/path/to/my-document.pdf"); got != "my document" {
		t.Errorf("fileBaseName = %q, want %q", got, "my document")
	}
}

func TestFileBaseName_WithUnderscores(t *testing.T) {
	if got := fileBaseName("report_2024_final.txt"); got != "report 2024 final" {
		t.Errorf("fileBaseName = %q, want %q", got, "report 2024 final")
	}
}

func TestFileBaseName_NoExtension(t *testing.T) {
	if got := fileBaseName("README"); got != "README" {
		t.Errorf("fileBaseName = %q, want %q", got, "README")
	}
}
