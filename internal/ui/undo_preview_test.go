package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/tools"
)

func TestUndoPreviewListsRestoreAndSkipThenConfirms(t *testing.T) {
	db, _ := openFoldDB(t)
	dir := t.TempDir()
	keep := filepath.Join(dir, "keep.txt")
	moved := filepath.Join(dir, "moved.txt")
	_ = os.WriteFile(keep, []byte("new"), 0o644)
	_ = os.WriteFile(moved, []byte("mine"), 0o644)
	s := &chatSession{id: "s1", store: db, width: 400, persisted: true}
	if err := db.Touch("s1", "/", "m", "", "t"); err != nil {
		t.Fatal(err)
	}
	s.showUpdate(core.Update{Kind: core.UpdateToolResult, Tool: "edit", Changes: []tools.Change{
		{Path: keep, Existed: true, Before: "old", After: "new"},
		{Path: moved, Existed: true, Before: "old", After: "new"},
	}})
	s.showUpdate(core.Update{Kind: core.UpdateDone})
	a := &app{active: s}
	a.undoLastTurn()
	if a.sel == nil || len(a.sel.options) != 2 {
		t.Fatalf("preview = %+v", a.sel)
	}
	got := map[string]string{}
	for _, o := range a.sel.options {
		got[o.label] = o.detail
	}
	if got[keep] != "restore" || got[moved] != "skip (changed since)" {
		t.Fatalf("options = %v", got)
	}
	if data, _ := os.ReadFile(keep); string(data) != "new" {
		t.Fatalf("preview changed a file: %q", data)
	}
	if err := a.sel.submit(keep); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(keep); string(data) != "old" {
		t.Fatalf("keep = %q", data)
	}
	if data, _ := os.ReadFile(moved); string(data) != "mine" {
		t.Fatalf("moved = %q", data)
	}
	if !strings.Contains(plain(s.rows), "Left as is, changed after the agent wrote them: "+moved) {
		t.Fatalf("rows:\n%s", plain(s.rows))
	}
}

func TestDiffTextForChangeRecords(t *testing.T) {
	text := diffText([]tools.Change{
		{Path: "a.txt", Existed: true, Before: "one\ntwo\n", After: "one\n2\n"},
		{Path: "a.txt", Existed: true, Before: "one\n2\n", After: "one\n2\nthree\n"},
		{Path: "b.txt", Before: "", After: "x\n"},
	})
	for _, want := range []string{"--- a.txt\n+++ a.txt\n-two\n+2\n+three\n", "--- b.txt (new file)\n+++ b.txt\n+x\n", diffBashNote} {
		if !strings.Contains(text, want) {
			t.Errorf("diff lacks %q:\n%s", want, text)
		}
	}
}
