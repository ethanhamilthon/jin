package ui

import "github.com/gdamore/tcell/v3"

// drawPaneBlocks puts a session's question at the bottom of its pane and
// returns the rows left for the timeline.
func drawPaneBlocks(screen tcell.Screen, s *chatSession, w, h int, focused bool) int {
	if s.ask != nil {
		height := min(s.ask.height(w), max(2, h/2))
		h -= height
		drawAsk(screen, s.ask, h, height, w, focused)
	}
	return h
}
