package ui

import (
	"testing"

	"jin/internal/store"
)

func openFoldDB(t *testing.T) (*store.DB, error) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db, err
}
