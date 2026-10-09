package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"

	"jin/internal/todo"
)

func foldApp(t *testing.T) *app {
	a, _ := goldenApp(t)
	a.active.todos = []todo.Item{{Text: "one", Status: todo.Done}, {Text: "two", Status: todo.Pending}}
	a.draw()
	return a
}

func click(a *app, y int) {
	for _, b := range []tcell.ButtonMask{tcell.ButtonPrimary, tcell.ButtonNone} {
		a.handleEvent(tcell.NewEventMouse(5, y, b, tcell.ModNone))
	}
}

func TestTodoBlockRowsFolded(t *testing.T) {
	s := &chatSession{}
	lines := make([]todoRow, 5)
	if got := s.todoBlockRows(lines, 28); got != 5 {
		t.Errorf("open rows = %d, want 5", got)
	}
	s.toggleTodos()
	if got := s.todoBlockRows(lines, 28); got != 0 {
		t.Errorf("folded rows = %d, want 0", got)
	}
}

func TestCtrlTTogglesTodos(t *testing.T) {
	a := foldApp(t)
	key := tcell.NewEventKey(tcell.KeyCtrlT, "", tcell.ModNone)
	a.handleEvent(key)
	if !a.active.todoFolded {
		t.Fatal("Ctrl+T did not fold the list")
	}
	a.handleEvent(key)
	if a.active.todoFolded {
		t.Fatal("Ctrl+T did not open the list")
	}
}

func TestClickOnTodoRuleToggles(t *testing.T) {
	a := foldApp(t)
	rule := a.active.todoRule - 1
	click(a, rule-1)
	if a.active.todoFolded {
		t.Fatal("click above the rule folded the list")
	}
	click(a, rule)
	if !a.active.todoFolded {
		t.Fatal("click on the rule did not fold the list")
	}
	a.draw()
	click(a, a.active.todoRule-1)
	if a.active.todoFolded {
		t.Fatal("second click did not open the list")
	}
}

func TestFoldedTodoKeepsOnlyTheRule(t *testing.T) {
	a := foldApp(t)
	open := a.active.todoRule
	a.active.toggleTodos()
	a.draw()
	if a.active.todoRule <= open {
		t.Errorf("rule row %d did not move down from %d", a.active.todoRule, open)
	}
}
