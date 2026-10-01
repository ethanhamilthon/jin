package ui

import (
	"strings"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

const statusLines = 2

// draw lays the screen out bottom-up: the two status lines never move, the
// input grows upward from them, and a panel (menu or autocomplete) opens above
// the input between two rules. The timeline fills the rest.
func (a *app) draw() {
	screen, s := a.screen, a.active
	w, h := screen.Size()
	screen.Fill(' ', base)
	if w < 10 || h < 8 {
		screen.Show()
		return
	}
	s.resize(w)
	statusY := h - statusLines
	box := a.inputBox()
	inHeight := inputHeight(box.visible(), box.cursor, w, h)
	inTop := statusY - inHeight
	ruleY := inTop - 1
	timelineEnd := ruleY
	rule(screen, ruleY, w, "")
	if panel := a.panel(); a.selectorHeight(panel, h) > 0 {
		selHeight := a.selectorHeight(panel, h)
		a.drawSelector(panel, ruleY-selHeight, w, selHeight)
		timelineEnd = ruleY - selHeight - 1
		a.drawRuleTitle(timelineEnd, w, panel)
	}
	a.drawTimeline(max(0, timelineEnd), w)
	drawInput(screen, box, inTop, inHeight, w)
	a.drawStatus(statusY, w)
	screen.Show()
}

func (a *app) drawTimeline(height, w int) {
	s := a.active
	s.scroll = min(s.scroll, max(0, len(s.rows)-height))
	end := len(s.rows) - s.scroll
	start := max(0, end-height)
	s.view = viewport{first: start, height: height}
	for y, row := range s.rows[start:end] {
		drawRow(a.screen, y, w, row)
	}
	paintSelection(a.screen, s.selection, start, height)
}

func drawRow(screen tcell.Screen, y, w int, row chatRow) {
	if row.hasFill {
		left, right := 1, 3+displaywidth.String(row.text)
		if row.fillWide {
			left, right = 0, w
		}
		for x := max(0, left); x < min(w, right); x++ {
			put(screen, x, y, " ", row.fill)
		}
	}
	if len(row.spans) == 0 {
		style := rowStyle(row.kind)
		if row.hasFill {
			style = row.fill
		}
		put(screen, 2, y, row.text, style)
		return
	}
	x := 2
	for _, span := range row.spans {
		put(screen, x, y, span.text, span.style)
		x += displaywidth.String(span.text)
	}
}

// rule draws a horizontal line, optionally carrying a title: "── Title ────".
func rule(screen tcell.Screen, y, w int, title string) {
	line := strings.Repeat("─", w)
	put(screen, 0, y, line, border)
	if title != "" {
		put(screen, 2, y, " "+truncate(title, w-6)+" ", accent.Bold(true))
	}
}
