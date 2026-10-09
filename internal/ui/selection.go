package ui

import (
	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

type textSelection struct {
	startRow, startCol, endRow, endCol int
	dragging, active                   bool
}

func (s *chatSession) handleMouse(ev *tcell.EventMouse, screen tcell.Screen) {
	switch {
	case ev.Buttons()&tcell.WheelUp != 0:
		s.scrollBy(3)
	case ev.Buttons()&tcell.WheelDown != 0:
		s.scrollBy(-3)
	case ev.Buttons()&(tcell.WheelLeft|tcell.WheelRight) == 0:
		s.handleSelection(ev, screen)
	}
}

// handleSelection tracks a drag over the timeline and copies the selected
// text when the button is released.
func (s *chatSession) handleSelection(ev *tcell.EventMouse, screen tcell.Screen) {
	w, _ := screen.Size()
	first, height := s.view.first, s.view.height
	last := min(len(s.rows), first+height)
	if first >= last {
		return
	}
	pressed := ev.Buttons()&tcell.ButtonPrimary != 0
	x, y := ev.Position()
	if !s.selection.dragging && (!pressed || y >= height) {
		return
	}
	if y < 0 {
		x = 0
	} else if y >= last-first {
		x = w
	}
	row := first + min(max(y, 0), last-first-1)
	col := min(max(x, 2), 2+displaywidth.String(s.rows[row].text))
	if !s.selection.dragging {
		s.selection = textSelection{startRow: row, startCol: col, endRow: row, endCol: col, dragging: true, active: true}
		return
	}
	s.selection.endRow, s.selection.endCol = row, col
	if !pressed {
		s.selection.dragging = false
		if text := selectedText(s.rows, s.selection); text != "" {
			copySelection(screen, text)
		} else if link := linkAt(s.rows[row], x); link != "" && y >= 0 && y < last-first {
			s.selection = textSelection{}
			linkOpener(link)
		}
	}
}
