package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func setupTempWorkspace(t *testing.T) (*WorkspaceCache, string) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "databases.json")
	wc := NewWorkspaceCacheWithPath(path)
	return wc, path
}

func TestNewWorkspaceCache(t *testing.T) {
	wc := NewWorkspaceCache()
	if wc.filePath == "" {
		t.Error("expected non-empty filePath")
	}
	if wc.data == nil {
		t.Error("expected non-nil data")
	}
}

func TestWorkspaceCacheLoadNoFile(t *testing.T) {
	wc, _ := setupTempWorkspace(t)

	err := wc.Load()
	if err != nil {
		t.Fatalf("Load() with no file should not error, got: %v", err)
	}
	if wc.Count() != 0 {
		t.Errorf("expected 0 databases, got %d", wc.Count())
	}
}

func TestWorkspaceCacheSaveAndLoad(t *testing.T) {
	wc, _ := setupTempWorkspace(t)

	entries := []DatabaseEntry{
		{
			ID:           "db-1",
			Title:        "Tasks",
			DataSourceID: "ds-1",
			URL:          "https://notion.so/tasks",
			Aliases:      []string{"todos", "t"},
			LastEdited:   time.Now().Truncate(time.Second),
		},
		{
			ID:           "db-2",
			Title:        "Projects",
			DataSourceID: "ds-2",
			URL:          "https://notion.so/projects",
			Aliases:      []string{"proj", "p"},
			LastEdited:   time.Now().Truncate(time.Second),
		},
	}

	wc.SetDatabases(entries)

	if err := wc.Save(); err != nil {
		t.Fatalf("Save() error: %v", err)
	}

	// Load into new instance
	wc2, _ := setupTempWorkspace(t)
	wc2.filePath = wc.filePath

	if err := wc2.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if wc2.Count() != 2 {
		t.Errorf("expected 2 databases, got %d", wc2.Count())
	}

	dbs := wc2.GetDatabases()
	if dbs[0].Title != "Tasks" {
		t.Errorf("expected 'Tasks', got %q", dbs[0].Title)
	}
	if dbs[1].Title != "Projects" {
		t.Errorf("expected 'Projects', got %q", dbs[1].Title)
	}
}

func TestWorkspaceCacheLoadInvalidJSON(t *testing.T) {
	wc, path := setupTempWorkspace(t)

	if err := os.WriteFile(path, []byte("invalid json{{{"), 0o644); err != nil {
		t.Fatal(err)
	}

	err := wc.Load()
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestWorkspaceCacheSaveCreatesDir(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "nested", "databases.json")
	wc := NewWorkspaceCacheWithPath(path)

	wc.SetDatabases([]DatabaseEntry{{ID: "db-1", Title: "Test"}})

	if err := wc.Save(); err != nil {
		t.Fatalf("Save() should create directories, got: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected file to exist after save")
	}
}

func TestWorkspaceCacheSaveAtomicity(t *testing.T) {
	wc, path := setupTempWorkspace(t)

	wc.SetDatabases([]DatabaseEntry{{ID: "db-1", Title: "Test"}})
	if err := wc.Save(); err != nil {
		t.Fatal(err)
	}

	// Temp file should not exist after successful save
	tmpPath := path + ".tmp"
	if _, err := os.Stat(tmpPath); !os.IsNotExist(err) {
		t.Error("expected temp file to be cleaned up after save")
	}

	// Main file should be valid JSON
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var wd WorkspaceData
	if err := json.Unmarshal(data, &wd); err != nil {
		t.Errorf("expected valid JSON, got: %v", err)
	}
}

func TestGetDatabases(t *testing.T) {
	wc, _ := setupTempWorkspace(t)

	// Empty case
	dbs := wc.GetDatabases()
	if len(dbs) != 0 {
		t.Errorf("expected 0 databases, got %d", len(dbs))
	}

	// After setting
	wc.SetDatabases([]DatabaseEntry{
		{ID: "db-1", Title: "Test"},
	})

	dbs = wc.GetDatabases()
	if len(dbs) != 1 {
		t.Errorf("expected 1 database, got %d", len(dbs))
	}
}

func TestSetDatabasesUpdatesLastSync(t *testing.T) {
	wc, _ := setupTempWorkspace(t)

	before := time.Now()
	wc.SetDatabases([]DatabaseEntry{{ID: "db-1"}})
	after := time.Now()

	syncTime := wc.LastSyncTime()
	if syncTime.Before(before) || syncTime.After(after) {
		t.Errorf("LastSyncTime %v not between %v and %v", syncTime, before, after)
	}
}

func TestFindByName(t *testing.T) {
	wc, _ := setupTempWorkspace(t)

	wc.SetDatabases([]DatabaseEntry{
		{ID: "db-1", Title: "Project Tasks"},
		{ID: "db-2", Title: "Meeting Notes"},
		{ID: "db-3", Title: "Bug Tracker", Aliases: []string{"bugs", "issues"}},
	})

	tests := []struct {
		name     string
		query    string
		expected string
	}{
		{"exact match", "Project Tasks", "db-1"},
		{"case insensitive", "project tasks", "db-1"},
		{"substring match", "Tasks", "db-1"},
		{"partial match", "meet", "db-2"},
		{"alias match", "bugs", "db-3"},
		{"alias case insensitive", "ISSUES", "db-3"},
		{"no match", "Nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := wc.FindByName(tt.query)
			if tt.expected == "" {
				if result != nil {
					t.Errorf("expected nil, got %+v", result)
				}
				return
			}
			if result == nil {
				t.Fatalf("expected result for %q, got nil", tt.query)
			}
			if result.ID != tt.expected {
				t.Errorf("expected ID %q, got %q", tt.expected, result.ID)
			}
		})
	}
}

func TestFindByID(t *testing.T) {
	wc, _ := setupTempWorkspace(t)

	wc.SetDatabases([]DatabaseEntry{
		{ID: "db-1", Title: "Tasks"},
		{ID: "db-2", Title: "Notes"},
	})

	// Found
	result := wc.FindByID("db-2")
	if result == nil {
		t.Fatal("expected to find db-2")
	}
	if result.Title != "Notes" {
		t.Errorf("expected 'Notes', got %q", result.Title)
	}

	// Not found
	result = wc.FindByID("nonexistent")
	if result != nil {
		t.Errorf("expected nil, got %+v", result)
	}
}

func TestIsStale(t *testing.T) {
	wc, _ := setupTempWorkspace(t)

	// Never synced - should be stale
	if !wc.IsStale() {
		t.Error("expected fresh cache with zero time to be stale")
	}

	// Just synced - should not be stale
	wc.SetDatabases([]DatabaseEntry{})
	if wc.IsStale() {
		t.Error("expected recently synced cache to not be stale")
	}

	// Manually set old sync time
	wc.data.LastSync = time.Now().Add(-25 * time.Hour)
	if !wc.IsStale() {
		t.Error("expected 25-hour-old cache to be stale")
	}

	// Exactly 24h boundary - should be stale (> 24h)
	wc.data.LastSync = time.Now().Add(-24*time.Hour - 1*time.Second)
	if !wc.IsStale() {
		t.Error("expected cache just over 24h to be stale")
	}
}

func TestCount(t *testing.T) {
	wc, _ := setupTempWorkspace(t)

	if wc.Count() != 0 {
		t.Errorf("expected 0, got %d", wc.Count())
	}

	wc.SetDatabases([]DatabaseEntry{
		{ID: "db-1"},
		{ID: "db-2"},
		{ID: "db-3"},
	})

	if wc.Count() != 3 {
		t.Errorf("expected 3, got %d", wc.Count())
	}
}

func TestCountNilData(t *testing.T) {
	wc := &WorkspaceCache{data: nil}
	if wc.Count() != 0 {
		t.Errorf("expected 0 for nil data, got %d", wc.Count())
	}
}

func TestGetDatabasesNilData(t *testing.T) {
	wc := &WorkspaceCache{data: nil}
	dbs := wc.GetDatabases()
	if dbs != nil {
		t.Errorf("expected nil for nil data, got %v", dbs)
	}
}

func TestWorkspaceCacheRoundTrip(t *testing.T) {
	wc, _ := setupTempWorkspace(t)

	now := time.Now().Truncate(time.Second)
	entries := []DatabaseEntry{
		{
			ID:           "db-abc-123",
			Title:        "My Database",
			DataSourceID: "ds-xyz-789",
			URL:          "https://notion.so/my-db",
			Aliases:      []string{"mydb", "db"},
			LastEdited:   now,
		},
	}

	wc.SetDatabases(entries)
	if err := wc.Save(); err != nil {
		t.Fatal(err)
	}

	wc2 := NewWorkspaceCacheWithPath(wc.filePath)
	if err := wc2.Load(); err != nil {
		t.Fatal(err)
	}

	db := wc2.FindByID("db-abc-123")
	if db == nil {
		t.Fatal("expected to find database after round-trip")
	}
	if db.Title != "My Database" {
		t.Errorf("expected 'My Database', got %q", db.Title)
	}
	if db.DataSourceID != "ds-xyz-789" {
		t.Errorf("expected 'ds-xyz-789', got %q", db.DataSourceID)
	}
	if len(db.Aliases) != 2 || db.Aliases[0] != "mydb" {
		t.Errorf("expected aliases [mydb, db], got %v", db.Aliases)
	}
}
