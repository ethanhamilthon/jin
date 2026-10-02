package ui

import (
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func TestEscapeTogglesPanelPreservingDraft(t *testing.T) {
	a, _ := layoutApp(t)
	a.active.input, a.active.cursor = clusters("draft"), 2
	esc := tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone)
	a.handleEvent(esc)
	if a.sel == nil || !a.sel.tabbed {
		t.Fatal("Esc should open the tabbed panel")
	}
	a.openField("Nested", "value", false, nil)
	a.handleEvent(esc)
	if a.sel != nil || !a.inputBox().focused {
		t.Fatal("Esc should return directly to input from a nested panel")
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

func TestInputTabFocusFirstAndClear(t *testing.T) {
	a, screen := layoutApp(t)
	a.active.input, a.active.cursor = clusters("draft"), 2
	a.openTab(len(tabNames) - 1)
	if a.sel.title != "Input" || a.sel.options[0].label != "Focus input" {
		t.Fatal("Input tab should start with Focus input")
	}
	a.draw()
	if !strings.Contains(rowText(screen, 12, 60), "Input") {
		t.Fatal("Active Input tab should remain visible on a narrow screen")
	}
	a.submitSelector()
	if a.sel != nil || a.active.cursor != 2 || strings.Join(a.active.input, "") != "draft" {
		t.Fatal("Focus input should close the panel without changing the draft")
	}
	a.openInputFlow().index = 1
	a.active.inputTop = 3
	a.submitSelector()
	if a.sel != nil || len(a.active.input) != 0 || a.active.cursor != 0 || a.active.inputTop != 0 {
		t.Fatal("Clear should reset the draft and return to input")
	}
}
