package store

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"jin/internal/paths"
)

func TestProjectMigrationImportsSettingsAndDeduplicatesAliases(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	database, err := paths.Global("jin.db")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(database), 0o700); err != nil {
		t.Fatal(err)
	}
	real := t.TempDir()
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(real, alias); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(t.TempDir(), "gone")
	standalone := filepath.Join(t.TempDir(), "saved-only")
	seedLegacyDatabase(t, database, []Project{
		{Path: alias, LastSession: "s1"},
		{Path: missing, LastSession: "not-a-session"},
		{Path: standalone},
	}, real, alias, missing)

	db, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	expectedReal, _ := canonicalStoredPath(real)
	expectedMissing, _ := canonicalStoredPath(missing)
	expectedStandalone, _ := canonicalStoredPath(standalone)
	projects, err := db.Projects()
	if err != nil || len(projects) != 3 {
		t.Fatalf("migrated projects = %+v, %v", projects, err)
	}
	byPath := make(map[string]Project)
	for _, project := range projects {
		byPath[project.Path] = project
	}
	if byPath[expectedReal].LastSession != "s1" || byPath[expectedReal].ID == "" {
		t.Errorf("alias project = %+v", byPath[expectedReal])
	}
	if byPath[expectedMissing].Path != expectedMissing || byPath[expectedMissing].LastSession != "s3" {
		t.Errorf("missing project not preserved safely: %+v", byPath[expectedMissing])
	}
	if byPath[expectedStandalone].Path != expectedStandalone {
		t.Errorf("settings-only project missing: %+v", byPath[expectedStandalone])
	}
	first, foundFirst, err := db.GetSession("s1")
	if err != nil || !foundFirst || first.ProjectID != byPath[expectedReal].ID {
		t.Errorf("first migrated session = %+v, %v, %v", first, foundFirst, err)
	}
	second, foundSecond, err := db.GetSession("s2")
	if err != nil || !foundSecond || second.ProjectID != byPath[expectedReal].ID {
		t.Errorf("alias session = %+v, %v, %v", second, foundSecond, err)
	}
	listed, err := db.ListByPath(expectedReal)
	if err != nil || len(listed) != 2 {
		t.Fatalf("canonical project omitted migrated alias sessions: %+v, %v", listed, err)
	}
	legacySetting, err := db.Setting("projects")
	if err != nil || legacySetting == "" {
		t.Errorf("legacy setting was removed: %v", err)
	}
	if _, err := os.Stat(missing); !os.IsNotExist(err) {
		t.Errorf("migration created missing project directory: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatal(err)
	}
	db, err = Open()
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	again, err := db.Projects()
	if err != nil || len(again) != 3 || projectAt(again, expectedReal).ID != byPath[expectedReal].ID {
		t.Fatalf("repeat migration changed registry: %+v, %v", again, err)
	}
}

func seedLegacyDatabase(t *testing.T, path string, projects []Project, paths ...string) {
	t.Helper()
	database, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	_, err = database.Exec(`CREATE TABLE sessions (
		id TEXT PRIMARY KEY, path TEXT NOT NULL, model TEXT NOT NULL, effort TEXT NOT NULL,
		title TEXT NOT NULL, created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL,
		input_tokens INTEGER NOT NULL DEFAULT 0, output_tokens INTEGER NOT NULL DEFAULT 0,
		context_tokens INTEGER NOT NULL DEFAULT 0)`)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`CREATE TABLE settings (key TEXT PRIMARY KEY, value TEXT NOT NULL)`); err != nil {
		t.Fatal(err)
	}
	for i, path := range paths {
		id := []string{"s1", "s2", "s3"}[i]
		if _, err := database.Exec(`INSERT INTO sessions(id, path, model, effort, title, created_at, updated_at) VALUES (?, ?, 'm', '', ?, ?, ?)`, id, path, id, 100+i, 200+i); err != nil {
			t.Fatal(err)
		}
	}
	data, err := json.Marshal(projects)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO settings(key, value) VALUES ('projects', ?)`, string(data)); err != nil {
		t.Fatal(err)
	}
}

func projectAt(projects []Project, path string) Project {
	for _, project := range projects {
		if project.Path == path {
			return project
		}
	}
	return Project{}
}
