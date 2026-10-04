package store

import (
	"testing"
	"time"

	"jin/internal/provider"
)

func newTestDB(t *testing.T) *DB {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	db, err := Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestListSessions(t *testing.T) {
	db := newTestDB(t)
	now := time.Now().Unix()

	_ = db.Touch("s1", "/dir/a", "m", "high", "First")
	_ = db.Touch("s2", "/dir/b", "m", "high", "Second")
	_ = db.Touch("s3", "/dir/a", "m", "high", "Third")

	// Update s1 to be newest
	_, _ = db.sql.Exec(`UPDATE sessions SET updated_at = ? WHERE id = ?`, now+10, "s1")
	_, _ = db.sql.Exec(`UPDATE sessions SET updated_at = ? WHERE id = ?`, now+5, "s3")

	// Scoped to /dir/a
	list, err := db.ListSessions("/dir/a", false)
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(list) != 2 || list[0].ID != "s1" || list[1].ID != "s3" {
		t.Fatalf("unexpected scoped list: %+v", list)
	}

	// All directories
	all, err := db.ListSessions("/dir/a", true)
	if err != nil {
		t.Fatalf("ListSessions all: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3 sessions, got %d", len(all))
	}
}

func TestSearchSessions(t *testing.T) {
	db := newTestDB(t)
	_ = db.Touch("s1", "/work", "m", "high", "Fix login bug")
	_ = db.AppendMessage("s1", provider.Message{Role: "user", Content: "Password verification fails on special chars"})
	_ = db.AppendMessage("s1", provider.Message{Role: "assistant", Content: "Here is the auth solution"})

	_ = db.Touch("s2", "/work", "m", "high", "Database migration")
	_ = db.AppendMessage("s2", provider.Message{Role: "user", Content: "Add indexes to users table"})

	_ = db.Touch("s3", "/other", "m", "high", "Other login work")
	_ = db.AppendMessage("s3", provider.Message{Role: "user", Content: "OAuth verification setup"})

	// Single word in title (case insensitive)
	res, err := db.SearchSessions("/work", false, []string{"LOGIN"})
	if err != nil || len(res) != 1 || res[0].ID != "s1" {
		t.Fatalf("search LOGIN: %+v, %v", res, err)
	}

	// Word in user message
	res, err = db.SearchSessions("/work", false, []string{"verification"})
	if err != nil || len(res) != 1 || res[0].ID != "s1" || res[0].Snippet == "" {
		t.Fatalf("search verification: %+v, %v", res, err)
	}

	// Assistant-only word should NOT match
	if res, err = db.SearchSessions("/work", false, []string{"auth"}); err != nil || len(res) != 0 {
		t.Fatalf("assistant match: %+v", res)
	}

	// Multi-word search (one in title, one in user message)
	if res, err = db.SearchSessions("/work", false, []string{"fix", "password"}); err != nil || len(res) != 1 {
		t.Fatalf("multi-word match: %+v", res)
	}

	// Multi-word search missing one word
	if res, err = db.SearchSessions("/work", false, []string{"fix", "missing"}); err != nil || len(res) != 0 {
		t.Fatalf("missing word match: %+v", res)
	}

	// Scope filtering
	if res, err = db.SearchSessions("/work", false, []string{"login"}); err != nil || len(res) != 1 {
		t.Fatalf("scoped count: %d", len(res))
	}
	if res, err = db.SearchSessions("/work", true, []string{"login"}); err != nil || len(res) != 2 {
		t.Fatalf("all count: %d", len(res))
	}
}
