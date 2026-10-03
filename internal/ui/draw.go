package ui

import (
	"fmt"
	"strings"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"

	"jin/internal/todo"
)

const statusLines = 2

// draw lays the screen out bottom-up: the two status lines never move, a rule
// sits above them, the input grows upward from that rule, and a panel (menu or autocomplete) opens above
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
	asking := a.sel == nil && s.ask != nil
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
		timelineEnd = ruleY - selHeight - 1
		a.drawRuleTitle(timelineEnd, w, panel)
	} else if pinned := s.pinnedTodos(); pinned != nil && a.sel == nil {
		rows := min(maxBlockRows, len(pinned), max(1, h/4))
		a.drawTodos(pinned, ruleY-rows, rows, w)
		timelineEnd = ruleY - rows - 1
		done, total := todo.Counts(pinned)
		rule(screen, timelineEnd, w, fmt.Sprintf("todo %d/%d · Ctrl+T edit", done, total))
	}
	a.drawTimeline(max(0, timelineEnd), w)
	if asking {
		a.drawAsk(s.ask, inTop, inHeight, w)
	} else {
		drawInput(screen, box, inTop, inHeight, w)
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

func drawRow(screen tcell.Screen, y, w int, row chatRow, frame, glow int) {
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
		text := span.text
		if span.spin {
			text = spinnerFrames[frame%len(spinnerFrames)]
		}
		if span.shimmer {
			x = drawShimmer(screen, x, y, text, span.style, glow)
			continue
		}
		put(screen, x, y, text, span.style)
		x += displaywidth.String(text)
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
