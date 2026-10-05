package ui

import (
	"fmt"

	"github.com/gdamore/tcell/v3"

	"jin/internal/todo"
)

// drawPaneBlocks puts a session's question and todo list at the bottom of
// its pane, the question lowest, and returns the rows left for the timeline.
func drawPaneBlocks(screen tcell.Screen, s *chatSession, w, h int, focused bool) int {
	if s.ask != nil {
		height := min(s.ask.height(w), max(2, h/2))
		h -= height
		drawAsk(screen, s.ask, h, height, w, focused)
	}
	pinned := s.pinnedTodos()
	if pinned == nil || h < 4 {
		return h
	}
	lines := todoRows(pinned, w)
	rows := min(maxBlockRows, len(lines), max(1, h/4))
	h -= rows
	drawTodos(screen, s, lines, todoFocus(pinned), h, rows, w)
	h--
	done, total := todo.Counts(pinned)
	rule(screen, h, w, fmt.Sprintf("todo %d/%d", done, total))
	return h
}
