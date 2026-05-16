package commands

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type apiCatalogEndpoint struct {
	Method  string         `json:"method"`
	Path    string         `json:"path"`
	Summary string         `json:"summary"`
	Spec    map[string]any `json:"spec"`
	Docs    string         `json:"docs"`
}

var apiCatalog = []apiCatalogEndpoint{
	catalogEndpoint("GET", "v1/users", "List users"),
	catalogEndpoint("GET", "v1/users/me", "Retrieve bot user"),
	catalogEndpoint("GET", "v1/users/{user_id}", "Retrieve a user"),
	catalogEndpoint("POST", "v1/search", "Search pages and data sources"),
	catalogEndpoint("POST", "v1/pages", "Create a page"),
	catalogEndpoint("GET", "v1/pages/{page_id}", "Retrieve a page"),
	catalogEndpoint("PATCH", "v1/pages/{page_id}", "Update page properties, icon, cover, or trash state"),
	catalogEndpoint("GET", "v1/pages/{page_id}/properties/{property_id}", "Retrieve a page property item"),
	catalogEndpoint("GET", "v1/pages/{page_id}/markdown", "Retrieve page content as enhanced markdown"),
	catalogEndpoint("PATCH", "v1/pages/{page_id}/markdown", "Update page content from enhanced markdown"),
	catalogEndpoint("POST", "v1/pages/{page_id}/move", "Move a page to a new parent"),
	catalogEndpoint("GET", "v1/blocks/{block_id}", "Retrieve a block"),
	catalogEndpoint("PATCH", "v1/blocks/{block_id}", "Update a block"),
	catalogEndpoint("DELETE", "v1/blocks/{block_id}", "Delete a block"),
	catalogEndpoint("GET", "v1/blocks/{block_id}/children", "List block children"),
	catalogEndpoint("PATCH", "v1/blocks/{block_id}/children", "Append block children"),
	catalogEndpoint("GET", "v1/databases/{database_id}", "Retrieve database metadata"),
	catalogEndpoint("PATCH", "v1/databases/{database_id}", "Update a database"),
	catalogEndpoint("POST", "v1/databases", "Create a database"),
	catalogEndpoint("POST", "v1/databases/{database_id}/query", "Query a database"),
	catalogEndpoint("GET", "v1/data_sources/{data_source_id}", "Retrieve a data source"),
	catalogEndpoint("PATCH", "v1/data_sources/{data_source_id}", "Update a data source"),
	catalogEndpoint("POST", "v1/data_sources", "Create a data source"),
	catalogEndpoint("POST", "v1/data_sources/{data_source_id}/query", "Query a data source"),
	catalogEndpoint("GET", "v1/data_sources/{data_source_id}/templates", "List data source templates"),
	catalogEndpoint("PATCH", "v1/data_sources/{data_source_id}/properties", "Update data source properties"),
	catalogEndpoint("POST", "v1/comments", "Create a comment"),
	catalogEndpoint("GET", "v1/comments", "List comments"),
	catalogEndpoint("GET", "v1/comments/{comment_id}", "Retrieve a comment"),
	catalogEndpoint("PATCH", "v1/comments/{comment_id}", "Update a comment"),
	catalogEndpoint("DELETE", "v1/comments/{comment_id}", "Delete a comment"),
	catalogEndpoint("GET", "v1/file_uploads", "List file uploads"),
	catalogEndpoint("POST", "v1/file_uploads", "Create a file upload"),
	catalogEndpoint("GET", "v1/file_uploads/{file_upload_id}", "Retrieve a file upload"),
	catalogEndpoint("POST", "v1/file_uploads/{file_upload_id}/send", "Send file upload bytes"),
	catalogEndpoint("POST", "v1/file_uploads/{file_upload_id}/complete", "Complete a multipart file upload"),
	catalogEndpoint("GET", "v1/views", "List views"),
	catalogEndpoint("POST", "v1/views", "Create a view"),
	catalogEndpoint("GET", "v1/views/{view_id}", "Retrieve a view"),
	catalogEndpoint("PATCH", "v1/views/{view_id}", "Update a view"),
	catalogEndpoint("DELETE", "v1/views/{view_id}", "Delete a view"),
	catalogEndpoint("POST", "v1/views/{view_id}/queries", "Create a cached view query"),
	catalogEndpoint("GET", "v1/views/{view_id}/queries/{query_id}", "Retrieve cached view query results"),
	catalogEndpoint("DELETE", "v1/views/{view_id}/queries/{query_id}", "Delete a cached view query"),
	catalogEndpoint("GET", "v1/custom_emojis", "List custom emojis"),
	catalogEndpoint("GET", "v1/custom_emojis/{custom_emoji_id}", "Retrieve a custom emoji"),
}

func catalogEndpoint(method, path, summary string) apiCatalogEndpoint {
	return apiCatalogEndpoint{
		Method:  method,
		Path:    path,
		Summary: summary,
		Spec: map[string]any{
			"method":      method,
			"path":        "/" + path,
			"summary":     summary,
			"description": "Reduced embedded endpoint fragment for offline notion-cli api introspection.",
		},
		Docs: fmt.Sprintf("# %s %s\n\n%s\n\nThis offline help is generated from notion-cli's embedded public API catalog. For the complete current reference, use https://developers.notion.com/reference.\n", method, "/"+path, summary),
	}
}

func apiCatalogRows() []map[string]any {
	rows := make([]map[string]any, len(apiCatalog))
	for i, endpoint := range apiCatalog {
		rows[i] = map[string]any{
			"method":  endpoint.Method,
			"path":    endpoint.Path,
			"summary": endpoint.Summary,
		}
	}
	sort.Slice(rows, func(i, j int) bool {
		pi, _ := rows[i]["path"].(string)
		pj, _ := rows[j]["path"].(string)
		if pi == pj {
			mi, _ := rows[i]["method"].(string)
			mj, _ := rows[j]["method"].(string)
			return mi < mj
		}
		return pi < pj
	})
	return rows
}

func findCatalogEndpoint(path, method string, explicitMethod bool) (*apiCatalogEndpoint, error) {
	path = normalizeAPIPathForCatalog(path)
	method = strings.ToUpper(strings.TrimSpace(method))
	var matches []apiCatalogEndpoint
	for _, endpoint := range apiCatalog {
		if endpoint.Path == path {
			matches = append(matches, endpoint)
		}
	}
	if len(matches) == 0 {
		return nil, fmt.Errorf("no embedded API catalog entry for %s", path)
	}
	if method == "" || !explicitMethod {
		if len(matches) > 1 {
			var methods []string
			for _, match := range matches {
				methods = append(methods, match.Method)
			}
			sort.Strings(methods)
			return nil, fmt.Errorf("method is ambiguous for %s; pass -X with one of: %s", path, strings.Join(methods, ", "))
		}
		return &matches[0], nil
	}
	for _, match := range matches {
		if match.Method == method {
			return &match, nil
		}
	}
	return nil, fmt.Errorf("no embedded API catalog entry for %s %s", method, path)
}

func normalizeAPIPathForCatalog(path string) string {
	path = strings.TrimSpace(strings.TrimLeft(path, "/"))
	if !strings.HasPrefix(path, "v1/") {
		path = "v1/" + path
	}
	if idx := strings.Index(path, "?"); idx >= 0 {
		path = path[:idx]
	}
	return path
}

func endpointSpecJSON(endpoint *apiCatalogEndpoint) ([]byte, error) {
	return json.MarshalIndent(endpoint.Spec, "", "  ")
}
