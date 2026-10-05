package ui

import (
	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

func (sel *selector) labelWidth(limit int) int {
	width := 0
	for _, opt := range sel.options {
		width = max(width, displaywidth.String(opt.label))
	}
	return min(width, limit)
}

var dotGreen, dotBlue, dotPurple tcell.Style

// selectorMark is the dot of a list row: the list's own, else a session's.
func (a *app) selectorMark(sel *selector, value string) (string, tcell.Style) {
	switch {
	case sel.dot != nil:
		return sel.dot(value)
	case sel.mark != nil:
		return sel.mark(value), dotGreen
	}
	return a.optionMark(value)
}

// optionMark flags sessions. The one on screen is always green. For the
// others, in order of priority: a blinking blue dot while the agent answers,
// a blinking purple dot while a background task runs, a steady blue dot for
// an unread answer. Tasks count for closed sessions too.
func (a *app) optionMark(id string) (string, tcell.Style) {
	s, live := a.sessions[id]
	blink := a.frame%8 < 5
	switch {
	case live && s == a.active:
		return "●", dotGreen
	case live && s.working:
		if blink {
			return "●", dotBlue
		}
		return " ", dotBlue
	case a.tasksRunning[id] > 0:
		if blink {
			return "●", dotPurple
		}
		return " ", dotPurple
	case live && s.unread, !live && a.unread[id]:
		return "●", dotBlue
	}
	return " ", dotBlue
}
