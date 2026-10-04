package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"
)

// twoPaneApp is the layout app with a second pane next to the first.
func twoPaneApp(t *testing.T) *app {
	t.Helper()
	a := startingApp(t)
	a.initPanes()
	if !a.splitPane(true) {
		t.Fatal("could not split the pane")
	}
	leaves := paneLeaves(a.panes)
	if len(leaves) != 2 {
		t.Fatalf("panes = %d, want 2", len(leaves))
	}
	for _, leaf := range leaves {
		leaf.session.ready = true
	}
	return a
}

func tab(a *app) { a.handleEvent(tcell.NewEventKey(tcell.KeyTab, "", tcell.ModNone)) }

// TestTabCyclesPaneFocus checks that Tab moves the focus to the next pane and
// wraps around at the end.
func TestTabCyclesPaneFocus(t *testing.T) {
	a := twoPaneApp(t)
	leaves := paneLeaves(a.panes)
	a.focus(leaves[0].session)
	if a.focusedLeaf() != leaves[0] {
		t.Fatalf("focus = %v, want the first pane", a.focusedLeaf())
	}
	tab(a)
	if a.focusedLeaf() != leaves[1] {
		t.Fatalf("after Tab focus = %v, want the second pane", a.focusedLeaf())
	}
	if a.active != leaves[1].session {
		t.Fatal("the focused pane must be the active session")
	}
	tab(a)
	if a.focusedLeaf() != leaves[0] {
		t.Fatalf("Tab must wrap to the first pane, focus = %v", a.focusedLeaf())
	}
}

// TestTabLeavesOnePaneAlone checks that Tab stays free of side effects with a
// single pane.
func TestTabLeavesOnePaneAlone(t *testing.T) {
	a := startingApp(t)
	a.initPanes()
	before := a.focusedLeaf()
	tab(a)
	if a.focusedLeaf() != before {
		t.Fatal("Tab moved the focus with one pane")
	}
}

// TestTabCompletesWhenAListIsOpen checks that an open autocomplete list keeps
// Tab for its own completion.
func TestTabCompletesWhenAListIsOpen(t *testing.T) {
	a := twoPaneApp(t)
	before := a.focusedLeaf()
	a.active.input, a.active.cursor = clusters("/se"), 3
	a.refreshSlash()
	if a.slash == nil {
		t.Fatal("the command list did not open")
	}
	tab(a)
	if a.focusedLeaf() != before {
		t.Fatal("Tab switched panes while the command list was open")
	}
	if a.slash != nil {
		t.Fatal("Tab must complete the command and close the list")
	}
	tok, ok := tokenOf(a.active.input[0])
	if !ok || tok.payload != "sessions" {
		t.Fatalf("completed command = %+v %v, want the sessions token", tok, ok)
	}
}
