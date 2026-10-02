package ui

import (
	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

// The two status lines sit on a blue band, so they read as one bar.
var (
	statusBar   = base.Background(colorBlue).Foreground(colorFG)
	statusTitle = statusBar.Foreground(colorWhite).Bold(true)
	statusSoft  = statusBar.Foreground(colorOnBlue)
	statusWarn  = statusBar.Foreground(colorAmber).Bold(true)
)

func (a *app) drawStatus(y, w int) {
	s := a.active
	model, modelStyle := s.model, statusTitle
	if model == "" {
		model, modelStyle = "no model", statusWarn
	} else if s.effort != "" {
		model += " / " + s.effort
	}
	title := s.title
	if title == "" {
		title = "new session"
	}
	for x := range w {
		put(a.screen, x, y, " ", statusBar)
		put(a.screen, x, y+1, " ", statusBar)
	}
	statusRow(a.screen, y, 1, w, title, model, statusTitle, modelStyle)
	statusRow(a.screen, y+1, 1, w, shortPath(a.dir), s.statusUsage(), statusSoft, statusSoft)
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
	focused := a.sel == nil
	if sel := a.sel; sel != nil {
		prefix, placeholder := "/ ", "Search · ↑/↓ move · Enter select · Esc close"
		switch {
		case sel.field:
			prefix, placeholder = "› ", sel.title+"..."
		case !sel.searching():
			placeholder = sel.hint
		case sel.tabbed:
			placeholder = "Search · ←/→ tab · ↑/↓ move · Enter select · Esc close"
		}
		style := accent.Bold(true)
		if !sel.searching() {
			style = dim
		}
		return inputBox{text: sel.query, cursor: sel.cursor, prefix: prefix, prefixStyle: style,
			placeholder: placeholder, focused: sel.searching(), secret: sel.secret}
	}
	s := a.active
	if b := s.bash; b != nil {
		return inputBox{text: b.input, cursor: b.cursor, prefix: "$ ", prefixStyle: base.Foreground(colorGreen).Bold(true),
			placeholder: "Shell command · Enter run · Ctrl+C stop · Esc close", focused: true, scroll: &b.top}
	}
	box := inputBox{text: s.input, cursor: s.cursor, prefix: "❯ ", prefixStyle: accent.Bold(true),
		placeholder: "Message...", focused: focused, scroll: &s.inputTop}
	if !s.ready {
		box.placeholder = "Loading prompts... · Ctrl+C skips the commands"
		box.focused = false
	}
	return box
}
