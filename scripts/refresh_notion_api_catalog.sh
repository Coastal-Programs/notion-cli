#!/usr/bin/env bash
set -euo pipefail

# Refresh helper for the embedded `notion-cli api` endpoint catalog.
#
# The checked-in catalog in internal/cli/commands/api_catalog.go is intentionally
# deterministic and offline. This script gives maintainers a repeatable starting
# point for comparing it with the current official docs index before updating
# the Go source by hand.

curl -fsSL https://developers.notion.com/llms.txt
