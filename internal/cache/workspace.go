package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
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

// WorkspaceData holds the full workspace cache data.
type WorkspaceData struct {
	Databases []DatabaseEntry `json:"databases"`
	LastSync  time.Time       `json:"last_sync"`
}

// WorkspaceCache manages persistent workspace database caching.
type WorkspaceCache struct {
	filePath string
	data     *WorkspaceData
}

// NewWorkspaceCache creates a new WorkspaceCache using ~/.notion-cli/databases.json.
func NewWorkspaceCache() *WorkspaceCache {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	return &WorkspaceCache{
		filePath: filepath.Join(home, ".notion-cli", "databases.json"),
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
