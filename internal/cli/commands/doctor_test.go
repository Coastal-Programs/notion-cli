package commands

import (
	"testing"

	"github.com/Coastal-Programs/notion-cli/internal/config"
)

// TestOAuthCredentialsEmbedded_DefaultBuild ensures the helper returns false
// when ldflags are not set (the case for `go test` builds). This guards
// against accidental hard-coded defaults in `internal/config`.
func TestOAuthCredentialsEmbedded_DefaultBuild(t *testing.T) {
	origID, origSecret := config.OAuthClientID, config.OAuthClientSecret
	t.Cleanup(func() {
		config.OAuthClientID = origID
		config.OAuthClientSecret = origSecret
	})

	config.OAuthClientID = ""
	config.OAuthClientSecret = ""
	if oauthCredentialsEmbedded() {
		t.Error("expected false when both vars empty")
	}

	config.OAuthClientID = "id"
	if oauthCredentialsEmbedded() {
		t.Error("expected false when only client id set")
	}

	config.OAuthClientID = ""
	config.OAuthClientSecret = "secret"
	if oauthCredentialsEmbedded() {
		t.Error("expected false when only client secret set")
	}

	config.OAuthClientID = "id"
	config.OAuthClientSecret = "secret"
	if !oauthCredentialsEmbedded() {
		t.Error("expected true when both set")
	}
}
