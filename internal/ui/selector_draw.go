package ui

import (
	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

const maxSelectorRows = 6

// selectorHeight is the same for every panel, whatever it holds, so the
// layout never jumps when one panel replaces another.
func (a *app) selectorHeight(sel *selector, screenHeight int) int {
	if sel == nil || screenHeight < 16 {
		return 0
	}
	return maxSelectorRows
}

func (a *app) drawSelector(sel *selector, top, w, height int) {
	switch {
	case sel.err != "":
		put(a.screen, 2, top, truncate(sel.err, w-4), errorStyle)
	case sel.loading:
		put(a.screen, 2, top, spinnerFrames[a.frame%len(spinnerFrames)]+" Loading...", muted)
	case sel.field:
		put(a.screen, 2, top, "Enter to confirm · Esc to cancel", dim)
	default:
		a.drawOptions(sel, top, w, height)
	}
}

func (a *app) drawOptions(sel *selector, top, w, height int) {
	visible := sel.visible()
	if len(visible) == 0 {
		message := "Nothing matches"
		if len(sel.options) == 0 && sel.empty != "" {
			message = sel.empty
		}
		put(a.screen, 2, top, message, dim)
		return
	}
	per := 1
	if sel.twoLines {
		per = 2
	}
	slots := max(1, height/per)
	position := 0
	for i, idx := range visible {
		if idx == sel.index {
			position = i
		}
	}
	start := max(0, min(position-slots/2, len(visible)-slots))
	for i := start; i < min(len(visible), start+slots); i++ {
		opt := sel.options[visible[i]]
		y := top + (i-start)*per
		style, marker := muted, "  "
		if visible[i] == sel.index {
			style, marker = accent.Bold(true), "› "
		}
		put(a.screen, 1, y, marker, style)
		mark, markStyle := a.optionMark(opt.value)
		if sel.mark != nil {
			mark, markStyle = sel.mark(opt.value), dotGreen
		}
		put(a.screen, 3, y, mark, markStyle)
		label := truncate(opt.label, w-8)
		put(a.screen, 5, y, label, style)
		if sel.twoLines {
			put(a.screen, 5, y+1, truncate(opt.detail, w-7), dim)
		} else if len(opt.choices) > 0 {
			a.drawChoices(opt, 7+sel.labelWidth(w/2), y, w, visible[i] == sel.index)
		} else if opt.detail != "" {
			offset := 7 + sel.labelWidth(w/2)
			put(a.screen, offset, y, truncate(opt.detail, w-offset-2), dim)
		}
	}
}

func (sel *selector) labelWidth(limit int) int {
	width := 0
	for _, opt := range sel.options {
		width = max(width, displaywidth.String(opt.label))
	}
	return min(width, limit)
}

var (
	dotGreen  = base.Foreground(colorGreen).Bold(true)
	dotBlue   = base.Foreground(colorBlueFG).Bold(true)
	dotPurple = base.Foreground(colorPurple).Bold(true)
)

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
	case a.asyncRunning[id] > 0:
		if blink {
			return "●", dotPurple
		}
		return " ", dotPurple
	case live && s.unread, !live && a.unread[id]:
		return "●", dotBlue
	}
	return " ", dotBlue
}
