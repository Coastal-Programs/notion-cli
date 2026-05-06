BINARY_NAME=notion-cli
BUILD_DIR=build

# Auto-load maintainer secrets from .env.local if present.
# CI sets the same vars via repository secrets, so this is a no-op there.
# .env.local is gitignored — see CONTRIBUTING.md for the format.
ifneq (,$(wildcard .env.local))
    include .env.local
    export
endif

VERSION=$(shell grep '"version"' package.json | head -1 | sed 's/.*"\([0-9]*\.[0-9]*\.[0-9]*\)".*/\1/')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
OAUTH_CLIENT_ID?=$(NOTION_OAUTH_CLIENT_ID)
OAUTH_CLIENT_SECRET?=$(NOTION_OAUTH_SECRET)
LDFLAGS=-ldflags "-s -w \
	-X github.com/Coastal-Programs/notion-cli/internal/config.Version=$(VERSION) \
	-X github.com/Coastal-Programs/notion-cli/internal/config.Commit=$(COMMIT) \
	-X github.com/Coastal-Programs/notion-cli/internal/config.Date=$(DATE) \
	-X github.com/Coastal-Programs/notion-cli/internal/config.OAuthClientID=$(OAUTH_CLIENT_ID) \
	-X github.com/Coastal-Programs/notion-cli/internal/config.OAuthClientSecret=$(OAUTH_CLIENT_SECRET)"

.PHONY: build test lint clean release release-check install fmt tidy

build:
	@mkdir -p $(BUILD_DIR)
	@echo "Building $(BINARY_NAME) $(VERSION) ($(COMMIT))..."
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/notion-cli
	@echo "Built $(BUILD_DIR)/$(BINARY_NAME)"

test:
	go test ./... -v -count=1

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

release-check:
	@if [ -z "$(OAUTH_CLIENT_ID)" ] || [ -z "$(OAUTH_CLIENT_SECRET)" ]; then \
		echo "ERROR: NOTION_OAUTH_CLIENT_ID and NOTION_OAUTH_SECRET must be set for release builds."; \
		echo "       Export them in your shell or CI environment before running 'make release'."; \
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

fmt:
	gofmt -s -w .

tidy:
	go mod tidy
