package commands

import "testing"

// ---------------------------------------------------------------------------
// page.go pure-function tests
// ---------------------------------------------------------------------------

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
