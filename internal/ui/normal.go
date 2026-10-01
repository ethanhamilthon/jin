package ui

import "github.com/gdamore/tcell/v3"

func (a *app) normalKey(ev *tcell.EventKey) {
	s := a.active
	page := max(1, s.view.height/2)
	switch ev.Key() {
	case tcell.KeyCtrlD, tcell.KeyPgDn:
		s.scrollBy(-page)
		return
	case tcell.KeyCtrlU, tcell.KeyPgUp:
		s.scrollBy(page)
		return
	case tcell.KeyDown:
		s.scrollBy(-1)
		return
	case tcell.KeyUp:
		s.scrollBy(1)
		return
	case tcell.KeySpace:
		a.openCommandsFlow()
		return
	case tcell.KeyRune:
	default:
		return
	}
	switch ev.Str() {
	case " ":
		a.openCommandsFlow()
	case "i":
		a.mode = modeInsert
	case "a":
		a.mode, s.cursor = modeInsert, len(s.input)
	case "j":
		s.scrollBy(-1)
	case "k":
		s.scrollBy(1)
	case "g":
		s.scroll = len(s.rows)
	case "G":
		s.scroll = 0
	}
}

func (s *chatSession) scrollBy(delta int) {
	s.scroll = max(0, s.scroll+delta)
}
