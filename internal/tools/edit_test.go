package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "file.txt")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func TestEditSummaryShowsSingleLine(t *testing.T) {
	path := writeTempFile(t, "one\ntwo\nthree\n")
	args, _ := json.Marshal(map[string]any{"path": path, "old_string": "two", "new_string": "TWO"})
	got, ok := (Edit{}).Summary(string(args))
	if !ok || got != path+":2" {
		t.Fatalf("expected %s:2, got %q ok=%v", path, got, ok)
	}
}

func TestEditSummaryShowsLineRangeForMultilineMatch(t *testing.T) {
	path := writeTempFile(t, "one\ntwo\nthree\nfour\n")
	args, _ := json.Marshal(map[string]any{"path": path, "old_string": "two\nthree", "new_string": "TWO\nTHREE"})
	got, ok := (Edit{}).Summary(string(args))
	if !ok || got != path+":2-3" {
		t.Fatalf("expected %s:2-3, got %q ok=%v", path, got, ok)
	}
}

func TestEditSummaryFallsBackToPathWhenTextNotFound(t *testing.T) {
	path := writeTempFile(t, "one\ntwo\n")
	args, _ := json.Marshal(map[string]any{"path": path, "old_string": "missing", "new_string": "x"})
	got, ok := (Edit{}).Summary(string(args))
	if !ok || got != path {
		t.Fatalf("expected bare path, got %q ok=%v", got, ok)
	}
}

func TestEditRunStillWorksAfterSummaryChange(t *testing.T) {
	path := writeTempFile(t, "hello world\n")
	args, _ := json.Marshal(map[string]any{"path": path, "old_string": "world", "new_string": "burn"})
	if _, err := (Edit{}).Run(context.Background(), string(args)); err != nil {
		t.Fatalf("Run: %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "hello burn\n" {
		t.Fatalf("expected file updated, got %q", string(data))
	}
}
