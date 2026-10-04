package ui

import (
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/tools"
)

func TestDiffRejectsUnknownArgument(t *testing.T) {
	s := &chatSession{width: 400}
	(&app{active: s}).showDiff("everything")
	if !strings.Contains(plain(s.rows), "usage: /diff or /diff session") {
		t.Fatalf("rows:\n%s", plain(s.rows))
	}
}

func TestDiffSessionWithoutChangesSaysNothingToDiff(t *testing.T) {
	db, _ := openFoldDB(t)
	if err := db.Touch("s1", "/", "m", "", "t"); err != nil {
		t.Fatal(err)
	}
	s := &chatSession{id: "s1", store: db, width: 400, persisted: true}
	(&app{active: s}).showDiff("  Session ")
	if !strings.Contains(plain(s.rows), "Nothing to diff") {
		t.Fatalf("rows:\n%s", plain(s.rows))
	}
}

func TestDiffSessionTextCoversAllTurns(t *testing.T) {
	db, _ := openFoldDB(t)
	if err := db.Touch("s1", "/", "m", "", "t"); err != nil {
		t.Fatal(err)
	}
	s := &chatSession{id: "s1", store: db, width: 400, persisted: true}
	s.showUpdate(core.Update{Kind: core.UpdateToolResult, Tool: "edit", Changes: []tools.Change{{Path: "a.txt", Existed: true, Before: "one\n", After: "two\n"}}})
	s.showUpdate(core.Update{Kind: core.UpdateDone})
	s.showUpdate(core.Update{Kind: core.UpdateToolResult, Tool: "write", Changes: []tools.Change{{Path: "b.txt", Before: "", After: "new\n"}}})
	s.showUpdate(core.Update{Kind: core.UpdateDone})
	changes, err := db.AllChanges("s1")
	if err != nil {
		t.Fatal(err)
	}
	text := diffText(changes)
	for _, want := range []string{"--- a.txt\n", "-one\n", "+two\n", "--- b.txt (new file)\n", "+new\n"} {
		if !strings.Contains(text, want) {
			t.Errorf("text lacks %q:\n%s", want, text)
		}
	}
}
