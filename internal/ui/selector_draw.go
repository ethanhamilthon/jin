package ui

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
		put(a.screen, 2, top, truncate(sel.fieldHint(), w-4), dim)
		a.drawCandidates(sel, top+1, w, max(0, height-1))
	default:
		a.drawOptions(sel, top, w, height)
	}
}

// fieldHint is the help line of a field, which grows a completion key when
// the field completes its text.
func (sel *selector) fieldHint() string {
	if sel.complete == nil {
		return "Enter to confirm · Esc to cancel"
	}
	return "Tab complete · ↑/↓ choose · Enter to confirm · Esc to cancel"
}

// drawCandidates draws the completion list of a field.
func (a *app) drawCandidates(sel *selector, top, w, height int) {
	if len(sel.cands) == 0 || height <= 0 {
		return
	}
	sel.candIdx = min(max(0, sel.candIdx), len(sel.cands)-1)
	start := max(0, min(sel.candIdx-height/2, len(sel.cands)-height))
	for i := start; i < min(len(sel.cands), start+height); i++ {
		style, marker := muted, "  "
		if i == sel.candIdx {
			style, marker = accent.Bold(true), "› "
		}
		put(a.screen, 1, top+i-start, marker, style)
		put(a.screen, 3, top+i-start, truncate(sel.cands[i].label, w-6), style)
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
