package commands

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type apiInlineField struct {
	Path  string
	Value any
}

type parsedAPIInline struct {
	BodyFields []apiInlineField
	Query      url.Values
	Headers    http.Header
}

type apiPathPart struct {
	Key    string
	Index  *int
	Append bool
}

func parseAPIInlineArgs(args []string) (*parsedAPIInline, error) {
	parsed := &parsedAPIInline{
		Query:   url.Values{},
		Headers: http.Header{},
	}
	for _, arg := range args {
		switch {
		case strings.Contains(arg, "=="):
			key, value, _ := strings.Cut(arg, "==")
			if strings.TrimSpace(key) == "" {
				return nil, fmt.Errorf("query parameter %q is missing a name", arg)
			}
			parsed.Query.Add(key, value)
		case strings.Contains(arg, ":="):
			path, raw, _ := strings.Cut(arg, ":=")
			if strings.TrimSpace(path) == "" {
				return nil, fmt.Errorf("body field %q is missing a path", arg)
			}
			var value any
			if err := json.Unmarshal([]byte(raw), &value); err != nil {
				return nil, fmt.Errorf("body field %q has invalid JSON value: %w", path, err)
			}
			parsed.BodyFields = append(parsed.BodyFields, apiInlineField{Path: path, Value: value})
		case looksLikeHeaderArg(arg):
			name, value, _ := strings.Cut(arg, ":")
			if strings.TrimSpace(name) == "" {
				return nil, fmt.Errorf("header %q is missing a name", arg)
			}
			parsed.Headers.Add(name, value)
		case strings.Contains(arg, "="):
			path, value, _ := strings.Cut(arg, "=")
			if strings.TrimSpace(path) == "" {
				return nil, fmt.Errorf("body field %q is missing a path", arg)
			}
			parsed.BodyFields = append(parsed.BodyFields, apiInlineField{Path: path, Value: value})
		default:
			return nil, fmt.Errorf("cannot parse inline argument %q", arg)
		}
	}
	return parsed, nil
}

func looksLikeHeaderArg(arg string) bool {
	colon := strings.Index(arg, ":")
	if colon <= 0 {
		return false
	}
	if strings.Contains(arg, ":=") {
		return false
	}
	eq := strings.Index(arg, "=")
	return eq == -1 || colon < eq
}

func buildAPIBody(fields []apiInlineField) (map[string]any, error) {
	body := map[string]any{}
	for _, field := range fields {
		parts, err := parseAPIPath(field.Path)
		if err != nil {
			return nil, err
		}
		if len(parts) == 0 || parts[0].Key == "" {
			return nil, fmt.Errorf("body field path %q must start with an object key", field.Path)
		}
		next, err := setAPIPathValue(body, parts, field.Value)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", field.Path, err)
		}
		var ok bool
		body, ok = next.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("body field path %q cannot replace the root object", field.Path)
		}
	}
	return body, nil
}

func apiFieldsAsFormFields(fields []apiInlineField) (map[string]string, error) {
	form := map[string]string{}
	for _, field := range fields {
		switch v := field.Value.(type) {
		case string:
			form[field.Path] = v
		default:
			b, err := json.Marshal(v)
			if err != nil {
				return nil, err
			}
			form[field.Path] = string(b)
		}
	}
	return form, nil
}

func parseAPIPath(path string) ([]apiPathPart, error) {
	var parts []apiPathPart
	var token strings.Builder
	for i := 0; i < len(path); i++ {
		switch path[i] {
		case '.':
			if token.Len() > 0 {
				parts = append(parts, apiPathPart{Key: token.String()})
				token.Reset()
			}
		case '[':
			if token.Len() > 0 {
				parts = append(parts, apiPathPart{Key: token.String()})
				token.Reset()
			}
			end := strings.IndexByte(path[i+1:], ']')
			if end < 0 {
				return nil, fmt.Errorf("unclosed bracket in path %q", path)
			}
			raw := path[i+1 : i+1+end]
			switch {
			case raw == "":
				parts = append(parts, apiPathPart{Append: true})
			case isDigits(raw):
				idx, _ := strconv.Atoi(raw)
				parts = append(parts, apiPathPart{Index: &idx})
			default:
				parts = append(parts, apiPathPart{Key: raw})
			}
			i += end + 1
		default:
			token.WriteByte(path[i])
		}
	}
	if token.Len() > 0 {
		parts = append(parts, apiPathPart{Key: token.String()})
	}
	return parts, nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func setAPIPathValue(node any, parts []apiPathPart, value any) (any, error) {
	if len(parts) == 0 {
		return value, nil
	}
	part := parts[0]
	rest := parts[1:]
	switch {
	case part.Key != "":
		m, ok := node.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("expected object before %q", part.Key)
		}
		child, exists := m[part.Key]
		if !exists || child == nil {
			child = newAPIContainer(rest)
		}
		next, err := setAPIPathValue(child, rest, value)
		if err != nil {
			return nil, err
		}
		m[part.Key] = next
		return m, nil
	case part.Index != nil:
		s, ok := node.([]any)
		if !ok {
			return nil, fmt.Errorf("expected array before index %d", *part.Index)
		}
		idx := *part.Index
		for len(s) <= idx {
			s = append(s, nil)
		}
		child := s[idx]
		if child == nil {
			child = newAPIContainer(rest)
		}
		next, err := setAPIPathValue(child, rest, value)
		if err != nil {
			return nil, err
		}
		s[idx] = next
		return s, nil
	case part.Append:
		s, ok := node.([]any)
		if !ok {
			return nil, fmt.Errorf("expected array before append")
		}
		if len(rest) == 0 {
			return append(s, value), nil
		}
		child := newAPIContainer(rest)
		next, err := setAPIPathValue(child, rest, value)
		if err != nil {
			return nil, err
		}
		return append(s, next), nil
	default:
		return nil, fmt.Errorf("empty path segment")
	}
}

func newAPIContainer(parts []apiPathPart) any {
	if len(parts) == 0 {
		return nil
	}
	if parts[0].Key != "" {
		return map[string]any{}
	}
	return []any{}
}
