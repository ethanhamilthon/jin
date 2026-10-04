package ui

import (
	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

// The status line sits on a colored band, so it reads as one bar.
var statusBar, statusTitle, statusSoft, statusWarn tcell.Style

// drawStatus shows the model and effort on the left and the usage figures on
// the right. The session title and project path live in the pane title.
func (a *app) drawStatus(y, w int) {
	s := a.active
	model, modelStyle := s.model, statusTitle
	if model == "" {
		model, modelStyle = "no model", statusWarn
	} else if s.effort != "" {
		model += " / " + s.effort
	}
	for x := range w {
		put(a.screen, x, y, " ", statusBar)
	}
	usageStyle := statusSoft
	if s.contextFilling() {
		usageStyle = statusWarn
	}
	statusRow(a.screen, y, 1, w, model, s.statusUsage(), modelStyle, usageStyle)
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
		case sel.complete != nil:
			prefix, placeholder = "› ", "Path · Tab complete · Enter confirm"
		case sel.field:
			prefix, placeholder = "› ", sel.title+"..."
		case sel.hint != "":
			placeholder = sel.hint
		}
		style := accent.Bold(true)
		if !sel.searching() {
			style = dim
		}
		return inputBox{text: sel.query, cursor: sel.cursor, prefix: prefix, prefixStyle: style,
			placeholder: placeholder, focused: sel.searching(), secret: sel.secret}
	}
	s := a.active
	if s.bashInput() {
		return inputBox{text: s.input, cursor: s.cursor, prefix: "❯ ", prefixStyle: base.Foreground(colorGreen).Bold(true),
			placeholder: "Shell command · Enter run · Ctrl+C stop", focused: focused, scroll: &s.inputTop}
	}
	box := inputBox{text: s.input, cursor: s.cursor, prefix: "❯ ", prefixStyle: accent.Bold(true),
		placeholder: "Message...", focused: focused, scroll: &s.inputTop}
	if !s.ready {
		box.placeholder = "Loading prompts... · Ctrl+C skips the commands"
		box.focused = false
	}
	return box
}
