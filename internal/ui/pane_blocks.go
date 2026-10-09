package ui

import "github.com/gdamore/tcell/v3"

// drawPaneBlocks puts a session's question and todo list at the bottom of
// its pane, the question lowest, and returns the rows left for the timeline.
func drawPaneBlocks(screen tcell.Screen, s *chatSession, w, h int, focused bool) int {
	s.todoRule = 0
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
	rows := s.todoBlockRows(lines, h)
	h -= rows
	if rows > 0 {
		drawTodos(screen, s, lines, todoFocus(pinned), h, rows, w)
	}
	h--
	s.drawTodoRule(screen, h, w, pinned)
	return h
}
