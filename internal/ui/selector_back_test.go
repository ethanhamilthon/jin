package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestEscapeGoesBackOnePanel(t *testing.T) {
	a, _ := layoutApp(t)
	esc := tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone)
	root := a.openList("Settings", []option{{label: "Motion", value: "motion"}}, "", func(string) error {
		a.openList("Motion", []option{{label: "Off", value: "off"}}, "", func(string) error { return nil })
		return nil
	})
	a.submitSelector()
	if a.sel == root || a.sel.title != "Motion" {
		t.Fatalf("child did not open: %+v", a.sel)
	}
	a.selectorKey(esc)
	if a.sel != root {
		t.Fatalf("Esc in a child went to %+v, want the parent", a.sel)
	}
	a.selectorKey(esc)
	if a.sel != nil {
		t.Fatal("Esc in the root panel did not close it")
	}
}

func TestReopenedPanelReplacesItsEarlierCopy(t *testing.T) {
	a, _ := layoutApp(t)
	a.openList("Projects", nil, "", nil)
	a.openList("Remove project x?", nil, "", nil)
	again := a.openList("Projects", nil, "", nil)
	if again.back != nil {
		t.Fatalf("reopened list stacks on %q", again.back.title)
	}
}
