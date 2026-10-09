package ui

const statusLines = 1

// draw lays the screen out bottom-up: the status line never moves, a rule
// sits above it, the input grows upward from that rule, and a panel (menu or autocomplete) opens above
// the input between two rules. The timeline fills the rest.
func (a *app) draw() {
	screen := a.screen
	w, h := screen.Size()
	screen.Fill(' ', base)
	if w < 10 || h < 8 {
		screen.Show()
		return
	}
	s := a.active
	if a.onboarding() {
		a.drawOnboarding(w, h)
		screen.Show()
		return
	}
	if a.panes == nil {
		s.resize(w)
	}
	statusY := h - statusLines
	box := a.inputBox()
	inHeight := inputHeight(box.visible(), box.cursor, w, h)
	asking := a.sel == nil && s.ask != nil && a.panes == nil
	if asking {
		inHeight = min(s.ask.height(w), max(2, h/2))
	}
	inputEnd := statusY - 1
	inTop := inputEnd - inHeight
	ruleY := inTop - 1
	timelineEnd := ruleY
	a.drawInputRule(ruleY, w)
	a.drawInputRule(inputEnd, w)
	if panel := a.panel(); a.selectorHeight(panel, h) > 0 {
		selHeight := a.selectorHeight(panel, h)
		a.drawSelector(panel, ruleY-selHeight, w, selHeight)
		tintBlock(screen, ruleY-selHeight, selHeight, w, a.panelColor(panel))
		timelineEnd = ruleY - selHeight - 1
		a.drawRuleTitle(timelineEnd, w, panel)
	}
	if asking {
		drawAsk(screen, s.ask, inTop, inHeight, w, true)
	} else {
		drawInput(screen, box, inTop, inHeight, w)
	}
	if a.panes == nil {
		a.drawTimeline(max(0, timelineEnd), w)
	} else {
		a.drawPanes(max(0, timelineEnd), w)
	}
	a.drawStatus(statusY, w)
	screen.Show()
}

func (a *app) drawTimeline(height, w int) {
	s := a.active
	rows := append(s.rows[:len(s.rows):len(s.rows)], a.tailRows(s)...)
	s.scroll = min(s.scroll, max(0, len(rows)-height))
	end := len(rows) - s.scroll
	start := max(0, end-height)
	s.view = viewport{first: start, height: height}
	for y, row := range rows[start:end] {
		drawRow(a.screen, y, w, row, a.frame, a.glowFrame())
	}
	paintSelection(a.screen, s.selection, start, height)
}
