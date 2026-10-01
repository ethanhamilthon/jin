package ui

import (
	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

func (a *app) drawStatus(y, w int) {
	s := a.active
	badge, badgeStyle := " NORMAL ", normalMode
	if a.mode == modeInsert {
		badge, badgeStyle = " INSERT ", insertMode
	}
	put(a.screen, 1, y, badge, badgeStyle)
	model, modelStyle := s.model, accent
	if model == "" {
		model, modelStyle = "no model", errorStyle
	} else if s.effort != "" {
		model += " / " + s.effort
	}
	title := s.title
	if title == "" {
		title = "new session"
	}
	left := 2 + displaywidth.String(badge)
	statusRow(a.screen, y, left, w, title, model, muted, modelStyle)
	right := usageLine(s.usage) + "  ⌕ " + string(a.cfg.Search.Backend)
	statusRow(a.screen, y+1, 1, w, shortPath(a.dir), right, dim, dim)
}

func statusRow(screen tcell.Screen, y, x, w int, left, right string, leftStyle, rightStyle tcell.Style) {
	rightX := w - displaywidth.String(right) - 1
	put(screen, x, y, truncate(left, rightX-x-2), leftStyle)
	if rightX > x {
		put(screen, rightX, y, right, rightStyle)
	}
}

// inputBox shows what the input row is editing: a selector field, a list
// filter, or the chat draft.
func (a *app) inputBox() inputBox {
	focused := a.mode == modeInsert
	if sel := a.sel; sel != nil {
		prefix, placeholder := "/ ", "Filter..."
		switch {
		case sel.field:
			prefix, placeholder = "› ", sel.title+"..."
		case !focused && sel.tabbed:
			placeholder = "←/→ tab · j/k move · Enter select · / filter · Esc close"
		case !focused:
			placeholder = "j/k move · Enter select · / filter · Esc close"
		}
		return inputBox{text: sel.query, cursor: sel.cursor, prefix: prefix, prefixStyle: accent.Bold(true),
			placeholder: placeholder, focused: focused, secret: sel.secret}
	}
	s := a.active
	box := inputBox{text: s.input, cursor: s.cursor, prefix: "❯ ", prefixStyle: accent.Bold(true),
		placeholder: "Message...", focused: focused}
	if !focused {
		box.placeholder = "Press i to type · Space for settings, commands, sessions"
		box.prefixStyle = dim
	}
	if s.working {
		box.prefix = spinnerFrames[a.frame%len(spinnerFrames)] + " "
		box.prefixStyle = base.Foreground(colorAmber)
	}
	return box
}
