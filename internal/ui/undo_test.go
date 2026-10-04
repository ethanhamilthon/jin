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

func TestUndoPartialKeepsFailedFilesUndoable(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("permissions are not enforced for root")
	}
	db, _ := openFoldDB(t)
	dir := t.TempDir()
	good := filepath.Join(dir, "good.txt")
	lockedDir := filepath.Join(dir, "locked")
	_ = os.Mkdir(lockedDir, 0o755)
	bad := filepath.Join(lockedDir, "bad.txt")
	_ = os.WriteFile(good, []byte("new"), 0o644)
	_ = os.WriteFile(bad, []byte("new"), 0o644)
	_ = os.Chmod(lockedDir, 0o500)
	t.Cleanup(func() { _ = os.Chmod(lockedDir, 0o755) })
	s := &chatSession{id: "s1", store: db, width: 400, persisted: true}
	if err := db.Touch("s1", "/", "m", "", "t"); err != nil {
		t.Fatal(err)
	}
	s.showUpdate(core.Update{Kind: core.UpdateToolResult, Tool: "edit", Changes: []tools.Change{
		{Path: good, Existed: true, Before: "old", After: "new"},
		{Path: bad, Existed: true, Before: "old", After: "new"},
	}})
	s.showUpdate(core.Update{Kind: core.UpdateDone})
	a := &app{active: s}
	a.undoLastTurn()
	screen := plain(s.rows)
	for _, want := range []string{"Undone: " + good, "Undo failed", "Not restored, run /undo again to retry: " + bad, "changes made through bash are not covered"} {
		if !strings.Contains(screen, want) {
			t.Errorf("screen lacks %q:\n%s", want, screen)
		}
	}
	if !strings.Contains(s.undoNote, good) || strings.Contains(s.undoNote, bad) {
		t.Errorf("note %q", s.undoNote)
	}
	_, left, err := db.LastChanges("s1")
	if err != nil || len(left) != 1 || left[0].Path != bad {
		t.Fatalf("kept records = %+v %v", left, err)
	}
}
