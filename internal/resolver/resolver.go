// Package resolver handles extraction, validation, and formatting of Notion
// resource IDs from various input formats (raw hex, hyphenated UUIDs, Notion
// URLs).
package resolver

import (
	"fmt"
	"regexp"
	"strings"
)

// hexPattern matches exactly 32 hexadecimal characters.
var hexPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

// uuidPattern matches a hyphenated UUID in 8-4-4-4-12 format.
var uuidPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// urlIDPattern extracts the last 32 hex characters from a Notion URL.
// Notion URLs encode the ID as the last 32 hex chars of the path segment,
// optionally preceded by a hyphen.
var urlIDPattern = regexp.MustCompile(`([0-9a-f]{32})(?:\?|#|$)`)

// ExtractID extracts a Notion resource ID from various input formats:
//   - Raw 32-char hex: "abc123def456..." -> formatted with hyphens
//   - Hyphenated UUID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" -> pass through
//   - Notion URL: "https://notion.so/page-title-abc123def456..." -> extract ID
//
// Returns the ID in standard hyphenated UUID format, or an error if no valid
// ID can be extracted.
func ExtractID(input string) (string, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return "", fmt.Errorf("empty input")
	}

	lower := strings.ToLower(input)

	// Case 1: already a hyphenated UUID.
	if uuidPattern.MatchString(lower) {
		return lower, nil
	}

	// Case 2: raw 32-char hex.
	if hexPattern.MatchString(lower) {
		return FormatID(lower), nil
	}

	// Case 3: Notion URL — extract the last 32 hex chars.
	if strings.Contains(lower, "notion.so") || strings.Contains(lower, "notion.site") {
		if m := urlIDPattern.FindStringSubmatch(lower); len(m) > 1 {
			return FormatID(m[1]), nil
		}
	}

	// Case 4: the input might contain a hyphenated UUID embedded in a URL
	// or other text that isn't a notion.so URL. Try stripping hyphens.
	stripped := StripHyphens(lower)
	if hexPattern.MatchString(stripped) && len(stripped) == 32 {
		return FormatID(stripped), nil
	}

	return "", fmt.Errorf("could not extract a valid Notion ID from: %s", input)
}

// FormatID inserts hyphens into a 32-character hex string to produce the
// standard 8-4-4-4-12 UUID format. If the input is not exactly 32 hex
// characters, it is returned unchanged.
func FormatID(id string) string {
	id = strings.ToLower(strings.TrimSpace(id))
	clean := StripHyphens(id)
	if len(clean) != 32 || !hexPattern.MatchString(clean) {
		return id
	}
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		clean[0:8], clean[8:12], clean[12:16], clean[16:20], clean[20:32])
}

// IsValidID reports whether s is a valid Notion ID (32 hex characters with
// or without hyphens).
func IsValidID(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	if uuidPattern.MatchString(s) {
		return true
	}
	if hexPattern.MatchString(s) {
		return true
	}
	return false
}

// StripHyphens removes all hyphens from s.
func StripHyphens(s string) string {
	return strings.ReplaceAll(s, "-", "")
}
