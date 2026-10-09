package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v3"

	"jin/internal/todo"
)

// todoBlockRows is the height of the pinned list in a pane of height h.
// A folded list has no rows: only its rule shows.
func (s *chatSession) todoBlockRows(lines []todoRow, h int) int {
	if s.todoFolded {
		return 0
	}
	return min(maxBlockRows, len(lines), max(1, h/4))
}

func (s *chatSession) toggleTodos() { s.todoFolded = !s.todoFolded }

// drawTodoRule paints the rule above the list and remembers its row for clicks.
func (s *chatSession) drawTodoRule(screen tcell.Screen, y, w int, items []todo.Item) {
	mark := "▾"
	if s.todoFolded {
		mark = "▸"
	}
	done, total := todo.Counts(items)
	rule(screen, y, w, fmt.Sprintf("todo %d/%d %s", done, total, mark))
	s.todoRule = y + 1
}

// todoClick toggles the list when the rule is clicked: the button goes down
// and comes up on the rule row.
func (s *chatSession) todoClick(ev *tcell.EventMouse) bool {
	_, y := ev.Position()
	if s.todoRule == 0 || y != s.todoRule-1 {
		s.todoPressed = false
		return false
	}
	if ev.Buttons()&tcell.ButtonPrimary != 0 {
		s.todoPressed = true
		return true
	}
	if !s.todoPressed {
		return false
	}
	s.todoPressed = false
	s.toggleTodos()
	return true
}
