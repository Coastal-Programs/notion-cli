package commands

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/spf13/cobra"
)

func testCustomEmojiServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, func()) {
	t.Helper()
	srv := httptest.NewServer(handler)

	origToken := os.Getenv("NOTION_TOKEN")
	origBase := os.Getenv("NOTION_CLI_BASE_URL")
	_ = os.Setenv("NOTION_TOKEN", "secret_test_token")
	_ = os.Setenv("NOTION_CLI_BASE_URL", srv.URL)

	return srv, func() {
		srv.Close()
		if origToken == "" {
			_ = os.Unsetenv("NOTION_TOKEN")
		} else {
			_ = os.Setenv("NOTION_TOKEN", origToken)
		}
		if origBase == "" {
			_ = os.Unsetenv("NOTION_CLI_BASE_URL")
		} else {
			_ = os.Setenv("NOTION_CLI_BASE_URL", origBase)
		}
	}
}

func runCustomEmojiRoot(t *testing.T, args ...string) (*cobra.Command, *bytes.Buffer, error) {
	t.Helper()
	root := &cobra.Command{Use: "notion-cli", SilenceErrors: true, SilenceUsage: true}
	RegisterCustomEmojiCommands(root)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(args)
	err := root.Execute()
	return root, &buf, err
}

func TestCustomEmojiRetrieve_RequiresID(t *testing.T) {
	_, _, err := runCustomEmojiRoot(t, "custom-emoji", "retrieve")
	if err == nil {
		t.Fatal("expected error when no ID is provided")
	}
}

func TestCustomEmojiSubcommands_Registered(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterCustomEmojiCommands(root)
	for _, sub := range []string{"list", "retrieve"} {
		c, _, err := root.Find([]string{"custom-emoji", sub})
		if err != nil || c == nil || c.Name() != sub {
			t.Errorf("subcommand %q not found: %v", sub, err)
		}
	}
}

func TestCustomEmojiList_AcceptsPaginationFlags(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterCustomEmojiCommands(root)
	c, _, err := root.Find([]string{"custom-emoji", "list"})
	if err != nil || c == nil {
		t.Fatalf("list subcommand not found: %v", err)
	}
	if c.Flags().Lookup("page-size") == nil {
		t.Error("expected --page-size flag")
	}
	if c.Flags().Lookup("start-cursor") == nil {
		t.Error("expected --start-cursor flag")
	}
}

func TestCustomEmojiList_AllFlagRegistered(t *testing.T) {
	root := &cobra.Command{Use: "notion-cli"}
	RegisterCustomEmojiCommands(root)
	c, _, err := root.Find([]string{"custom-emoji", "list"})
	if err != nil || c == nil {
		t.Fatalf("list subcommand not found: %v", err)
	}
	if c.Flags().Lookup("all") == nil {
		t.Error("expected --all flag")
	}
}

func TestCustomEmojiList_Success(t *testing.T) {
	_, cleanup := testCustomEmojiServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/custom_emojis" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"results":[{"id":"emoji-1","name":"party"}],"has_more":false}`))
	})
	defer cleanup()

	_, _, err := runCustomEmojiRoot(t, "custom-emoji", "list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCustomEmojiRetrieve_Success(t *testing.T) {
	_, cleanup := testCustomEmojiServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":"00000000000000000000000000000001","name":"party"}`))
	})
	defer cleanup()

	_, _, err := runCustomEmojiRoot(t, "custom-emoji", "retrieve", "00000000000000000000000000000001")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
