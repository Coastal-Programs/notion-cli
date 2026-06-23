BINARY_NAME=notion-cli
BUILD_DIR=build

# Auto-load maintainer secrets from $HOME/.config/notion-cli-dev/.env if present.
# The dotfile lives OUTSIDE the repo tree so a stray `git add .` can never stage it,
# and a typo in .gitignore can never expose it. CI sets the same vars via
# repository secrets, so this include is a no-op there.
# See CONTRIBUTING.md for the file format and migration notes.
MAINTAINER_ENV := $(HOME)/.config/notion-cli-dev/.env
ifneq (,$(wildcard $(MAINTAINER_ENV)))
    include $(MAINTAINER_ENV)
    export
endif

VERSION=$(shell grep '"version"' package.json | head -1 | sed 's/.*"\([0-9]*\.[0-9]*\.[0-9]*\)".*/\1/')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
OAUTH_CLIENT_ID?=$(NOTION_OAUTH_CLIENT_ID)
OAUTH_CLIENT_SECRET?=$(NOTION_OAUTH_SECRET)
LDFLAGS=-ldflags "-s -w \
	-X github.com/Coastal-Programs/notion-cli/v6/internal/config.Version=$(VERSION) \
	-X github.com/Coastal-Programs/notion-cli/v6/internal/config.Commit=$(COMMIT) \
	-X github.com/Coastal-Programs/notion-cli/v6/internal/config.Date=$(DATE) \
	-X github.com/Coastal-Programs/notion-cli/v6/internal/config.OAuthClientID=$(OAUTH_CLIENT_ID) \
	-X github.com/Coastal-Programs/notion-cli/v6/internal/config.OAuthClientSecret=$(OAUTH_CLIENT_SECRET)"

.PHONY: build test test-race test-cover lint clean release release-check cross-compile install fmt fmt-check tidy

build:
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) $(VERSION) ($(COMMIT))..."
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/notion-cli
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

test:
	go test ./... -v -count=1

test-race:
	go test ./... -race -count=1

test-cover:
	@mkdir -p $(BUILD_DIR)
	go test ./... -count=1 -coverprofile=$(BUILD_DIR)/coverage.out
	@go tool cover -func=$(BUILD_DIR)/coverage.out | tail -1
	@echo "HTML report: go tool cover -html=$(BUILD_DIR)/coverage.out"

lint:
	go vet ./...
	@if command -v golangci-lint >/dev/null 2>&1; then golangci-lint run ./...; fi

clean:
	rm -rf $(BUILD_DIR)
	go clean

install:
	@go install $(LDFLAGS) ./cmd/notion-cli

# Cross-compilation targets
PLATFORMS=darwin/arm64 darwin/amd64 linux/amd64 linux/arm64 windows/amd64

# SHA-256 of the historically-leaked OAuth client secret. We compare hashes
# instead of the literal value so the tripwire itself doesn't trip GitHub's
# secret-scanning push protection. Any release build that still embeds the
# leaked value means rotation never happened in the Notion dev portal — fail
# loudly so we never re-ship a known-compromised credential.
# To rotate, see SECURITY.md → "Credential rotation".
LEAKED_OAUTH_SECRET_SHA256 := 4221a9d34ed19a38fb0e904941b2b19347009de04669f825852c3fd1f50d0a39

release-check:
	@if [ -z "$(OAUTH_CLIENT_ID)" ] || [ -z "$(OAUTH_CLIENT_SECRET)" ]; then \
		echo "ERROR: NOTION_OAUTH_CLIENT_ID and NOTION_OAUTH_SECRET must be set for release builds."; \
		echo "       Export them in your shell or CI environment before running 'make release'."; \
		exit 1; \
	fi
	@actual_sha=$$(printf '%s' "$(OAUTH_CLIENT_SECRET)" | shasum -a 256 | cut -d' ' -f1); \
	if [ "$$actual_sha" = "$(LEAKED_OAUTH_SECRET_SHA256)" ]; then \
		echo "ERROR: refusing to build a release with the known-leaked OAuth client secret."; \
		echo "       Rotate it in the Notion dev portal and update your secrets."; \
		echo "       See SECURITY.md \"Credential rotation\" for the procedure."; \
		exit 1; \
	fi

release: release-check clean
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "Building $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch$$ext ./cmd/notion-cli >/dev/null; \
	done
	@echo "Release binaries built in $(BUILD_DIR)/"

# cross-compile verifies the code builds for every target platform WITHOUT
# embedding the OAuth credentials. It is the safe target for pull-request CI:
# GitHub does not expose repository secrets to fork/Dependabot PRs, so a build
# that requires them produces a false-negative failure. The real, credential-
# embedding artifacts are produced by 'make release' on tag/release in
# publish.yml. OAUTH_CLIENT_ID/SECRET are forced empty here so no secret is ever
# baked into a downloadable CI artifact.
cross-compile: OAUTH_CLIENT_ID :=
cross-compile: OAUTH_CLIENT_SECRET :=
cross-compile: clean
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d/ -f1); \
		arch=$$(echo $$platform | cut -d/ -f2); \
		ext=""; \
		if [ "$$os" = "windows" ]; then ext=".exe"; fi; \
		echo "Compiling $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch$$ext ./cmd/notion-cli >/dev/null; \
	done
	@echo "Cross-compile verification passed for: $(PLATFORMS)"

fmt:
	gofmt -s -w .

fmt-check:
	@unformatted=$$(gofmt -s -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "The following files are not gofmt'd:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

tidy:
	go mod tidy
