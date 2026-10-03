package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/tools"
)

func TestUndoRestoresTheLastTurn(t *testing.T) {
	db, _ := openFoldDB(t)
	path := filepath.Join(t.TempDir(), "f.txt")
	_ = os.WriteFile(path, []byte("new"), 0o644)
	s := &chatSession{id: "s1", store: db, width: 400, persisted: true}
	if err := db.Touch("s1", "/", "m", "", "t"); err != nil {
		t.Fatal(err)
	}
	s.showUpdate(core.Update{Kind: core.UpdateToolResult, Tool: "edit", Changes: []tools.Change{{Path: path, Existed: true, Before: "old", After: "new"}}})
	s.showUpdate(core.Update{Kind: core.UpdateDone})
	a := &app{active: s}
	a.undoLastTurn()
	if data, _ := os.ReadFile(path); string(data) != "old" {
		t.Fatalf("file = %q", data)
	}
	if !strings.Contains(plain(s.rows), "Undone: "+path) || !strings.Contains(s.undoNote, path) {
		t.Fatalf("rows:\n%s\nnote %q", plain(s.rows), s.undoNote)
	}
	a.undoLastTurn()
	if !strings.Contains(plain(s.rows), "Nothing to undo") {
		t.Fatalf("second undo:\n%s", plain(s.rows))
	}
	if got := core.StripUndo(s.undoNote + "hi"); got != "hi" {
		t.Fatalf("strip = %q", got)
	}
}
