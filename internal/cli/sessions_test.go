package cli

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"jin/internal/provider"
	"jin/internal/store"
)

func openTestDB(t *testing.T) *store.DB {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestSessionsMainList(t *testing.T) {
	db := openTestDB(t)
	_ = db.Touch("s1", "/test/dir", "m", "high", "Session One")
	_ = db.Touch("s2", "/other/dir", "m", "high", "Session Two")

	var out, errOut bytes.Buffer
	code := SessionsMain([]string{"list"}, db, "/test/dir", &out, &errOut)
	if code != 0 {
		t.Fatalf("code = %d, err = %s", code, errOut.String())
	}
	if !strings.Contains(out.String(), "Session One") || strings.Contains(out.String(), "Session Two") {
		t.Fatalf("unexpected scoped list output: %q", out.String())
	}

	// JSON format with --all
	out.Reset()
	code = SessionsMain([]string{"list", "--all", "--format", "json"}, db, "/test/dir", &out, &errOut)
	if code != 0 {
		t.Fatalf("code = %d, err = %s", code, errOut.String())
	}
	var items []store.SessionResult
	if err := json.Unmarshal(out.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal json: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestSessionsMainSearch(t *testing.T) {
	db := openTestDB(t)
	_ = db.Touch("s1", "/test/dir", "m", "high", "Fix login bug")
	_ = db.AppendMessage("s1", provider.Message{Role: "user", Content: "Password verification fails on special chars"})

	var out, errOut bytes.Buffer
	code := SessionsMain([]string{"search", "verification"}, db, "/test/dir", &out, &errOut)
	if code != 0 {
		t.Fatalf("code = %d, err = %s", code, errOut.String())
	}
	parts := strings.Split(strings.TrimSpace(out.String()), "\t")
	if len(parts) != 4 || parts[0] != "s1" || parts[2] != "Fix login bug" {
		t.Fatalf("search tab format mismatch: %q (parts: %d)", out.String(), len(parts))
	}

	// Empty search words
	out.Reset()
	errOut.Reset()
	code = SessionsMain([]string{"search"}, db, "/test/dir", &out, &errOut)
	if code != 2 {
		t.Fatalf("expected exit 2 on empty search, got %d", code)
	}
}

func TestSessionsMainErrors(t *testing.T) {
	db := openTestDB(t)
	var out, errOut bytes.Buffer
	if code := SessionsMain(nil, db, "/test/dir", &out, &errOut); code != 2 {
		t.Fatalf("expected code 2 for nil args, got %d", code)
	}
	if code := SessionsMain([]string{"list", "--unknown"}, db, "/test/dir", &out, &errOut); code != 2 {
		t.Fatalf("expected code 2 for bad flag, got %d", code)
	}
	if code := SessionsMain([]string{"unknown"}, db, "/test/dir", &out, &errOut); code != 2 {
		t.Fatalf("expected code 2 for bad subcmd, got %d", code)
	}
}
