package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestEscapeWithNothingOpenDoesNothing(t *testing.T) {
	a, _ := layoutApp(t)
	a.active.input, a.active.cursor = clusters("draft"), 2
	a.handleEvent(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone))
	if a.sel != nil || a.slash != nil {
		t.Fatal("Esc with nothing open must not open anything")
	}
	if strings.Join(a.active.input, "") != "draft" || a.active.cursor != 2 {
		t.Fatal("Esc changed the draft or cursor")
	}
}

func TestEscapeClosesAPanelAndKeepsDraft(t *testing.T) {
	a, _ := layoutApp(t)
	a.active.input, a.active.cursor = clusters("draft"), 2
	a.openField("Nested", "value", false, nil)
	a.handleEvent(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone))
	if a.sel != nil || !a.inputBox().focused {
		t.Fatal("Esc should close the panel and return to the input")
	}
	if strings.Join(a.active.input, "") != "draft" || a.active.cursor != 2 {
		t.Fatal("Esc changed the draft or cursor")
	}
}

func TestInputAcceptsFormerModeKeysAndSpace(t *testing.T) {
	a, _ := layoutApp(t)
	for _, text := range []string{"i", "a", " "} {
		a.handleEvent(tcell.NewEventKey(tcell.KeyRune, text, tcell.ModNone))
	}
	if got := strings.Join(a.active.input, ""); got != "ia " || a.sel != nil {
		t.Fatalf("input=%q panel=%v", got, a.sel)
	}
}
