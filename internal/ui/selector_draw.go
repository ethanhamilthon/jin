package ui

import "github.com/clipperhouse/displaywidth"

const maxSelectorRows = 6

func (a *app) selectorHeight(screenHeight int) int {
	sel := a.sel
	if sel == nil || screenHeight < 16 {
		return 0
	}
	if sel.field || sel.loading || sel.err != "" {
		return 1
	}
	per := 1
	if sel.twoLines {
		per = 2
	}
	return min(maxSelectorRows, max(1, len(sel.visible())*per))
}

func (a *app) drawSelector(top, w, height int) {
	sel := a.sel
	switch {
	case sel.err != "":
		put(a.screen, 2, top, truncate(sel.err, w-4), errorStyle)
	case sel.loading:
		put(a.screen, 2, top, spinnerFrames[a.frame%len(spinnerFrames)]+" Loading...", muted)
	case sel.field:
		put(a.screen, 2, top, "Enter to confirm · Esc to cancel", dim)
	default:
		a.drawOptions(top, w, height)
	}
}

func (a *app) drawOptions(top, w, height int) {
	sel := a.sel
	visible := sel.visible()
	if len(visible) == 0 {
		put(a.screen, 2, top, "Nothing matches", dim)
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
		put(a.screen, 3, y, a.optionMark(opt.value), base.Foreground(colorGreen).Bold(true))
		label := truncate(opt.label, w-8)
		put(a.screen, 5, y, label, style)
		if sel.twoLines {
			put(a.screen, 5, y+1, truncate(opt.detail, w-7), dim)
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

// optionMark flags live sessions: a blinking dot while answering, a steady
// dot when an answer has not been read yet.
func (a *app) optionMark(id string) string {
	s, live := a.sessions[id]
	switch {
	case live && s.working && a.frame%8 < 5:
		return "●"
	case live && s.unread:
		return "●"
	case !live && a.unread[id]:
		return "●"
	}
	return " "
}
