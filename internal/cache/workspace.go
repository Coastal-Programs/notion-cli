package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Coastal-Programs/notion-cli/v6/internal/config"
)

// DatabaseEntry represents a cached Notion database.
type DatabaseEntry struct {
	ID           string    `json:"id"`
	Title        string    `json:"title"`
	DataSourceID string    `json:"data_source_id"`
	URL          string    `json:"url"`
	Aliases      []string  `json:"aliases"`
	LastEdited   time.Time `json:"last_edited"`
}

// DataSourceEntry represents a cached Notion data source.
type DataSourceEntry struct {
	ID         string    `json:"id"`
	DatabaseID string    `json:"database_id"`
	Title      string    `json:"title"`
	URL        string    `json:"url"`
	LastEdited time.Time `json:"last_edited"`
}

// WorkspaceData holds the full workspace cache data.
type WorkspaceData struct {
	Databases   []DatabaseEntry   `json:"databases"`
	DataSources []DataSourceEntry `json:"data_sources,omitempty"`
	LastSync    time.Time         `json:"last_sync"`
}

// WorkspaceCache manages persistent workspace database caching.
type WorkspaceCache struct {
	filePath string
	data     *WorkspaceData
}

// NewWorkspaceCache creates a new WorkspaceCache using ~/.notion-cli/databases.json.
func NewWorkspaceCache() *WorkspaceCache {
	return &WorkspaceCache{
		filePath: config.GetWorkspaceCachePath(""),
		data:     &WorkspaceData{},
	}
}

// NewWorkspaceCacheForWorkspace creates a WorkspaceCache scoped to a named
// auth workspace. An empty slug uses the legacy cache path.
func NewWorkspaceCacheForWorkspace(slug string) *WorkspaceCache {
	path := config.GetWorkspaceCachePath(slug)
	if path == "" {
		path = filepath.Join(".", ".notion-cli", "databases.json")
	}
	return &WorkspaceCache{
		filePath: path,
		data:     &WorkspaceData{},
	}
}

// NewWorkspaceCacheWithPath creates a WorkspaceCache with a custom file path (for testing).
func NewWorkspaceCacheWithPath(path string) *WorkspaceCache {
	return &WorkspaceCache{
		filePath: path,
		data:     &WorkspaceData{},
	}
}

// Load reads the workspace cache from disk.
func (w *WorkspaceCache) Load() error {
	data, err := os.ReadFile(w.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			w.data = &WorkspaceData{}
			return nil
		}
		return err
	}

	var wd WorkspaceData
	if err := json.Unmarshal(data, &wd); err != nil {
		return err
	}

	w.data = &wd
	return nil
}

// Save writes the workspace cache to disk atomically.
func (w *WorkspaceCache) Save() error {
	dir := filepath.Dir(w.filePath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	data, err := json.MarshalIndent(w.data, "", "  ")
	if err != nil {
		return err
	}

	// Write to temp file then rename for atomic write
	tmpFile := w.filePath + ".tmp"
	if err := os.WriteFile(tmpFile, data, 0o600); err != nil {
		return err
	}

	return os.Rename(tmpFile, w.filePath)
}

// GetDatabases returns all cached database entries.
func (w *WorkspaceCache) GetDatabases() []DatabaseEntry {
	if w.data == nil {
		return nil
	}
	return w.data.Databases
}

// SetDatabases replaces the cached databases and updates the LastSync time.
func (w *WorkspaceCache) SetDatabases(entries []DatabaseEntry) {
	w.data.Databases = entries
	w.data.LastSync = time.Now()
}

// FindByName returns the first database whose title contains the given
// name (case-insensitive substring match), or nil if not found.
func (w *WorkspaceCache) FindByName(name string) *DatabaseEntry {
	lower := strings.ToLower(name)
	for i := range w.data.Databases {
		if strings.Contains(strings.ToLower(w.data.Databases[i].Title), lower) {
			return &w.data.Databases[i]
		}
		// Also check aliases
		for _, alias := range w.data.Databases[i].Aliases {
			if strings.EqualFold(alias, name) {
				return &w.data.Databases[i]
			}
		}
	}
	return nil
}

// FindByID returns the database with the given ID, or nil if not found.
func (w *WorkspaceCache) FindByID(id string) *DatabaseEntry {
	for i := range w.data.Databases {
		if w.data.Databases[i].ID == id {
			return &w.data.Databases[i]
		}
	}
	return nil
}

// IsStale returns true if the last sync was more than 24 hours ago.
func (w *WorkspaceCache) IsStale() bool {
	return time.Since(w.data.LastSync) > 24*time.Hour
}

// LastSyncTime returns the time of the last workspace sync.
func (w *WorkspaceCache) LastSyncTime() time.Time {
	return w.data.LastSync
}

// Count returns the number of cached databases.
func (w *WorkspaceCache) Count() int {
	if w.data == nil {
		return 0
	}
	return len(w.data.Databases)
}

// GetDataSources returns all cached data source entries.
func (w *WorkspaceCache) GetDataSources() []DataSourceEntry {
	if w.data == nil {
		return nil
	}
	return w.data.DataSources
}

// SetDataSources replaces the cached data sources.
func (w *WorkspaceCache) SetDataSources(entries []DataSourceEntry) {
	w.data.DataSources = entries
}

// FindDataSourceByID returns the data source with the given ID, or nil if not found.
func (w *WorkspaceCache) FindDataSourceByID(id string) *DataSourceEntry {
	for i := range w.data.DataSources {
		if w.data.DataSources[i].ID == id {
			return &w.data.DataSources[i]
		}
	}
	return nil
}

// FindDataSourceByName returns the first data source whose title contains name
// (case-insensitive substring match), or nil if not found.
func (w *WorkspaceCache) FindDataSourceByName(name string) *DataSourceEntry {
	lower := strings.ToLower(name)
	for i := range w.data.DataSources {
		if strings.Contains(strings.ToLower(w.data.DataSources[i].Title), lower) {
			return &w.data.DataSources[i]
		}
	}
	return nil
}

// DataSourceCount returns the number of cached data sources.
func (w *WorkspaceCache) DataSourceCount() int {
	if w.data == nil {
		return 0
	}
	return len(w.data.DataSources)
}
