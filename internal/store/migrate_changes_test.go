package store

import (
	"testing"
)

func TestMigrateDropsFileChangesTable(t *testing.T) {
	db := openTest(t)
	if _, err := db.sql.Exec(`CREATE TABLE IF NOT EXISTS file_changes (id INTEGER PRIMARY KEY, path TEXT)`); err != nil {
		t.Fatal(err)
	}
	if err := migrate(db.sql); err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.sql.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'file_changes'`).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("file_changes still exists after migrate")
	}
}
