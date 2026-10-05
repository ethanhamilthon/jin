package ui

import (
	"strings"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

func drawAsk(screen tcell.Screen, q *askState, top, height, w int, focused bool) {
	if focused {
		screen.HideCursor()
	}
	fillBlock(screen, top, height, w, askPanel)
	rows, selected := q.lines(w)
	rows = rows[:len(rows)-1] // the free-text row is drawn separately
	questionRows := height - 1
	q.top = fitScroll(q.top, min(selected, len(rows)-1), len(rows), questionRows)
	for i := 0; i < questionRows && q.top+i < len(rows); i++ {
		row, style := rows[q.top+i], askPanel
		if row.current {
			style = askPanel.Foreground(colorBlueFG).Bold(true)
		}
		put(screen, 2, top+i, truncate(row.text, w-3), style)
	}
	y := top + height - 1
	prefix, style := "  ", askPanel.Foreground(colorDim)
	if q.onFreeRow() {
		prefix, style = "› ", askPanel.Foreground(colorBlueFG).Bold(true)
	}
	put(screen, 2, y, prefix, style)
	if len(q.text) == 0 {
		put(screen, 4, y, "Type your own answer...", askPanel.Foreground(colorDim))
		if focused && q.onFreeRow() {
			screen.ShowCursor(4, y)
		}
		return
	}
	line := strings.Join(q.text, "")
	put(screen, 4, y, truncate(strings.ReplaceAll(line, "\n", " "), w-5), askPanel)
	if focused && q.onFreeRow() {
		screen.ShowCursor(min(w-1, 4+displaywidth.String(strings.Join(q.text[:q.cursor], ""))), y)
	}
}
