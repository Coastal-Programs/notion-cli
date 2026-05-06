package commands

import (
	"bytes"
	"strings"
	"testing"

	clierrors "github.com/Coastal-Programs/notion-cli/internal/errors"
	"github.com/spf13/cobra"
)

func newIconCoverFlagSet(t *testing.T) *cobra.Command {
	t.Helper()
	c := &cobra.Command{Use: "test"}
	c.Flags().String("icon-emoji", "", "")
	c.Flags().String("icon-url", "", "")
	c.Flags().String("cover-url", "", "")
	return c
}

func TestBuildIconCover_Emoji(t *testing.T) {
	c := newIconCoverFlagSet(t)
	_ = c.Flags().Set("icon-emoji", "💰")

	icon, cover, hasIcon, hasCover, err := buildIconCover(c, false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !hasIcon || hasCover {
		t.Fatalf("hasIcon=%v hasCover=%v want true,false", hasIcon, hasCover)
	}
	m, ok := icon.(map[string]any)
	if !ok {
		t.Fatalf("icon not a map: %T", icon)
	}
	if m["type"] != "emoji" || m["emoji"] != "💰" {
		t.Errorf("icon = %v", m)
	}
	if cover != nil {
		t.Errorf("cover should be nil, got %v", cover)
	}
}

func TestBuildIconCover_IconURLAndCover(t *testing.T) {
	c := newIconCoverFlagSet(t)
	_ = c.Flags().Set("icon-url", "https://example.com/i.png")
	_ = c.Flags().Set("cover-url", "https://example.com/c.jpg")

	icon, cover, hasIcon, hasCover, err := buildIconCover(c, false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !hasIcon || !hasCover {
		t.Fatalf("hasIcon=%v hasCover=%v want both true", hasIcon, hasCover)
	}
	im := icon.(map[string]any)
	if im["type"] != "external" {
		t.Errorf("icon type = %v", im["type"])
	}
	if im["external"].(map[string]any)["url"] != "https://example.com/i.png" {
		t.Errorf("icon url wrong: %v", im["external"])
	}
	cm := cover.(map[string]any)
	if cm["type"] != "external" {
		t.Errorf("cover type = %v", cm["type"])
	}
	if cm["external"].(map[string]any)["url"] != "https://example.com/c.jpg" {
		t.Errorf("cover url wrong: %v", cm["external"])
	}
}

func TestBuildIconCover_InvalidURL(t *testing.T) {
	c := newIconCoverFlagSet(t)
	_ = c.Flags().Set("icon-url", "not-a-url")

	_, _, _, _, err := buildIconCover(c, false)
	if err == nil {
		t.Fatal("expected validation error")
	}
	cliErr, ok := err.(*clierrors.NotionCLIError)
	if !ok {
		t.Fatalf("err type = %T", err)
	}
	if cliErr.Code != clierrors.CodeInvalidRequest {
		t.Errorf("code = %q", cliErr.Code)
	}
}

func TestBuildIconCover_CoverWrongScheme(t *testing.T) {
	c := newIconCoverFlagSet(t)
	_ = c.Flags().Set("cover-url", "ftp://example.com/c.jpg")

	_, _, _, _, err := buildIconCover(c, false)
	if err == nil {
		t.Fatal("expected scheme error")
	}
}

func TestBuildIconCover_ClearWithNone(t *testing.T) {
	c := newIconCoverFlagSet(t)
	_ = c.Flags().Set("icon-emoji", "none")
	_ = c.Flags().Set("cover-url", "none")

	icon, cover, hasIcon, hasCover, err := buildIconCover(c, true)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !hasIcon || !hasCover {
		t.Fatalf("hasIcon/hasCover both should be true")
	}
	if icon != nil {
		t.Errorf("icon should be nil for clear, got %v", icon)
	}
	if cover != nil {
		t.Errorf("cover should be nil for clear, got %v", cover)
	}
}

func TestBuildIconCover_NoneNotAllowedOnCreate(t *testing.T) {
	// On create (allowClear=false), "none" should be treated as a literal
	// emoji — current behaviour. Just confirm no panic and field set.
	c := newIconCoverFlagSet(t)
	_ = c.Flags().Set("icon-emoji", "none")

	icon, _, hasIcon, _, err := buildIconCover(c, false)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if !hasIcon {
		t.Fatal("expected hasIcon true")
	}
	m := icon.(map[string]any)
	if m["emoji"] != "none" {
		t.Errorf("expected literal 'none', got %v", m["emoji"])
	}
}

func TestPageCreate_IconEmojiAndURLMutuallyExclusive(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterPageCommands(root)

	root.SetArgs([]string{"page", "create", "-d", "abc", "--icon-emoji", "💰", "--icon-url", "https://x/y.png"})
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)

	err := root.Execute()
	if err == nil {
		t.Fatal("expected mutually-exclusive error")
	}
	combined := strings.ToLower(err.Error() + " " + out.String())
	if !strings.Contains(combined, "icon-emoji") || !strings.Contains(combined, "icon-url") {
		t.Errorf("expected error mentioning icon-emoji and icon-url, got %v / %s", err, out.String())
	}
}

func TestPageUpdate_FlagsRegistered(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterPageCommands(root)

	updateCmd, _, err := root.Find([]string{"page", "update"})
	if err != nil {
		t.Fatalf("page update not found: %v", err)
	}
	for _, name := range []string{"icon-emoji", "icon-url", "cover-url"} {
		if updateCmd.Flag(name) == nil {
			t.Errorf("flag %q not registered on page update", name)
		}
	}
}
