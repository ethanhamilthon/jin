package store

import (
	"os"
	"path/filepath"
	"testing"

	"jin/internal/provider"
)

func TestProjectQueriesIncludeAliasSessionPaths(t *testing.T) {
	db := openTest(t)
	dir := t.TempDir()
	alias := filepath.Join(t.TempDir(), "alias")
	if err := os.Symlink(dir, alias); err != nil {
		t.Fatal(err)
	}
	if err := db.Touch("alias-session", alias, "m", "", "alias title"); err != nil {
		t.Fatal(err)
	}
	if err := db.AppendMessage("alias-session", provider.Message{Role: "user", Content: "searchable alias"}); err != nil {
		t.Fatal(err)
	}
	if err := db.SetUnread("alias-session", true); err != nil {
		t.Fatal(err)
	}
	canonical, _ := canonicalStoredPath(dir)
	for _, path := range []string{dir, canonical, alias} {
		sessions, err := db.ListSessions(path, false)
		if err != nil || len(sessions) != 1 {
			t.Fatalf("list %q = %+v, %v", path, sessions, err)
		}
		results, err := db.SearchSessions(path, false, []string{"searchable"})
		if err != nil || len(results) != 1 {
			t.Fatalf("search %q = %+v, %v", path, results, err)
		}
		unread, err := db.UnreadSessions(path)
		if err != nil || !unread["alias-session"] {
			t.Fatalf("unread %q = %v, %v", path, unread, err)
		}
	}
}
