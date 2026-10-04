package tools

import (
	"context"
	"encoding/json"
	"os"
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
