package ui

// shift moves the focused row to its previous or next value and reports the
// change. The ends do not wrap, so volume stops at 10% and 100%.
func (sel *selector) shift(delta int) {
	if sel.current() == "" {
		return
	}
	opt := &sel.options[sel.index]
	next := min(max(opt.chosen+delta, 0), len(opt.choices)-1)
	if len(opt.choices) == 0 || next == opt.chosen {
		return
	}
	opt.chosen = next
	sel.err = ""
	if err := sel.onChoice(opt.value, next); err != nil {
		sel.err = err.Error()
	}
}

// drawChoices writes the values of a row side by side with the current one
// in brackets.
func (a *app) drawChoices(opt option, x, y, w int, focused bool) {
	for i, choice := range opt.choices {
		style, text := dim, choice
		if i == opt.chosen {
			style, text = base, "["+choice+"]"
			if focused {
				style = accent.Bold(true)
			}
		}
		if x+len(text) >= w {
			return
		}
		put(a.screen, x, y, text, style)
		x += len(text) + 2
	}
}
