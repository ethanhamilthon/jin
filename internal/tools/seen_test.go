package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestEditRefusesFileChangedSinceRead(t *testing.T) {
	path := writeTempFile(t, "one\n")
	seen := NewSeen()
	read, _ := json.Marshal(map[string]any{"path": path})
	if _, err := NewReadSeen(seen).Run(context.Background(), string(read)); err != nil {
		t.Fatal(err)
	}
	later := time.Now().Add(time.Second)
	if err := os.WriteFile(path, []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Chtimes(path, later, later)
	edit, _ := json.Marshal(map[string]any{"path": path, "old_string": "one", "new_string": "ONE"})
	if _, err := NewEditSeen(seen).Run(context.Background(), string(edit)); err == nil || !strings.Contains(err.Error(), "read it again") || !strings.Contains(err.Error(), "by a command") {
		t.Fatalf("expected a stale-file error, got %v", err)
	}
	write, _ := json.Marshal(map[string]any{"path": path, "content": "x"})
	if _, err := NewWriteSeen(seen).Run(context.Background(), string(write)); err == nil {
		t.Fatal("write must refuse a changed file")
	}
	if _, err := NewReadSeen(seen).Run(context.Background(), string(read)); err != nil {
		t.Fatal(err)
	}
	if _, err := NewEditSeen(seen).Run(context.Background(), string(edit)); err != nil {
		t.Fatalf("edit after a new read: %v", err)
	}
	if _, err := NewEditSeen(seen).Run(context.Background(), string(edit)); err == nil || strings.Contains(err.Error(), "read it again") {
		t.Fatalf("the agent's own edit must not count as a change: %v", err)
	}
}

func TestSeenIgnoresUnreadFiles(t *testing.T) {
	path := writeTempFile(t, "a")
	if err := NewSeen().Check(path); err != nil {
		t.Fatal(err)
	}
}

func readThrough(t *testing.T, seen *Seen, path string) {
	t.Helper()
	args, _ := json.Marshal(map[string]any{"path": path})
	if _, err := NewReadSeen(seen).Run(context.Background(), string(args)); err != nil {
		t.Fatal(err)
	}
}

func aliasedDir(t *testing.T) (alias, real string) {
	t.Helper()
	base, _ := filepath.EvalSymlinks(t.TempDir())
	real = filepath.Join(base, "real")
	if err := os.MkdirAll(filepath.Join(real, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	alias = filepath.Join(base, "alias")
	if err := os.Symlink(filepath.Join(real, "sub"), alias); err != nil {
		t.Fatal(err)
	}
	return alias, real
}

func changeFile(t *testing.T, path string) {
	t.Helper()
	later := time.Now().Add(time.Hour)
	if err := os.WriteFile(path, []byte("changed by someone"), 0o644); err != nil {
		t.Fatal(err)
	}
	_ = os.Chtimes(path, later, later)
}

func TestSeenKnowsFileThroughAlias(t *testing.T) {
	alias, real := aliasedDir(t)
	direct := filepath.Join(real, "sub", "f.txt")
	_ = os.WriteFile(direct, []byte("one"), 0o644)
	seen := NewSeen()
	readThrough(t, seen, filepath.Join(alias, "f.txt"))
	changeFile(t, direct)
	if err := seen.Check(direct); err == nil {
		t.Fatal("direct path must see the stamp recorded through the alias")
	}
}

func TestSeenKeepsDotDotMeaning(t *testing.T) {
	alias, real := aliasedDir(t)
	lexical := filepath.Join(filepath.Dir(real), "x")
	_ = os.WriteFile(lexical, []byte("lexical"), 0o644)
	_ = os.WriteFile(filepath.Join(real, "x"), []byte("real"), 0o644)
	seen := NewSeen()
	readThrough(t, seen, lexical)
	changeFile(t, lexical)
	if err := seen.Check(alias + "/../x"); err != nil {
		t.Fatalf("alias/../x is real/x, never read, but got: %v", err)
	}
	if err := seen.Check(lexical); err == nil {
		t.Fatal("the file that was read must show as changed")
	}
}

func TestSeenReportsDeletionThroughSameSpelling(t *testing.T) {
	path := writeTempFile(t, "one")
	seen := NewSeen()
	readThrough(t, seen, path)
	_ = os.Remove(path)
	if err := seen.Check(path); err == nil || !strings.Contains(err.Error(), "deleted") {
		t.Fatalf("err = %v", err)
	}
}
