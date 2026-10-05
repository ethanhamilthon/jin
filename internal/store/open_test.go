package store

import "testing"

func openTest(t *testing.T) *DB {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	db, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
