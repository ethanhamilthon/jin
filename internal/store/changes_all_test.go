package store

import (
	"testing"

	"jin/internal/tools"
)

func TestAllChangesOrdersByTurnAndSkipsOtherSessions(t *testing.T) {
	db, _ := openTwo(t)
	save := func(session string, turn int, path string) {
		t.Helper()
		if err := db.SaveChange(session, turn, tools.Change{Path: path, Existed: true, Before: "b", After: "a"}); err != nil {
			t.Fatal(err)
		}
	}
	save("s1", 2, "c")
	save("s1", 1, "a")
	save("s2", 1, "other")
	save("s1", 1, "b")
	save("s1", 2, "d")
	got, err := db.AllChanges("s1")
	if err != nil {
		t.Fatal(err)
	}
	var paths string
	for _, c := range got {
		paths += c.Path
	}
	if paths != "abcd" {
		t.Fatalf("paths = %q", paths)
	}
	if !got[0].Existed || got[0].Before != "b" || got[0].After != "a" {
		t.Fatalf("decoded = %+v", got[0])
	}
	if none, err := db.AllChanges("missing"); err != nil || len(none) != 0 {
		t.Fatalf("missing session = %v, %v", none, err)
	}
}
