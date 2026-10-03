package ui

import (
	"strconv"
	"strings"

	"github.com/clipperhouse/displaywidth"
)

// askRow is one drawn row of the ask block; current marks the rows of the
// option under the cursor.
type askRow struct {
	text    string
	current bool
}

// lines lays the current question out as word-wrapped rows and returns the
// row that holds the selection. A wrapped option keeps its indent.
func (q *askState) lines(width int) (rows []askRow, selected int) {
	title := q.question().Question
	if len(q.questions) > 1 {
		title = "(" + strconv.Itoa(q.current+1) + "/" + strconv.Itoa(len(q.questions)) + ") " + title
	}
	for _, line := range wrapChat(title, width-3) {
		rows = append(rows, askRow{text: line})
	}
	for i, option := range q.question().Options {
		current := i == q.row
		if current {
			selected = len(rows)
		}
		marker := "  "
		if current {
			marker = "› "
		}
		if q.question().Multiple {
			checked := q.current < len(q.selected) && i < len(q.selected[q.current]) && q.selected[q.current][i]
			box := "[ ] "
			if checked {
				box = "[x] "
			}
			marker += box
		}
		marker += strconv.Itoa(i+1) + ". "
		indent := strings.Repeat(" ", displaywidth.String(marker))
		for j, line := range wrapChat(option, max(1, width-3-len(indent))) {
			prefix := indent
			if j == 0 {
				prefix = marker
			}
			rows = append(rows, askRow{text: prefix + line, current: current})
		}
	}
	if q.onFreeRow() {
		selected = len(rows)
	}
	rows = append(rows, askRow{})
	return rows, selected
}

// height is how many rows the block takes: the question and options, then
// the free-text row.
func (q *askState) height(width int) int {
	rows, _ := q.lines(width)
	return min(maxBlockRows, len(rows))
}

func (a *app) drawAsk(q *askState, top, height, w int) {
	a.screen.HideCursor()
	fillBlock(a.screen, top, height, w, askPanel)
	rows, selected := q.lines(w)
	rows = rows[:len(rows)-1] // the free-text row is drawn separately
	questionRows := height - 1
	q.top = fitScroll(q.top, min(selected, len(rows)-1), len(rows), questionRows)
	for i := 0; i < questionRows && q.top+i < len(rows); i++ {
		row, style := rows[q.top+i], askPanel
		if row.current {
			style = askPanel.Foreground(colorBlueFG).Bold(true)
		}
		put(a.screen, 2, top+i, truncate(row.text, w-3), style)
	}
	y := top + height - 1
	prefix, style := "  ", askPanel.Foreground(colorDim)
	if q.onFreeRow() {
		prefix, style = "› ", askPanel.Foreground(colorBlueFG).Bold(true)
	}
	put(a.screen, 2, y, prefix, style)
	if len(q.text) == 0 {
		put(a.screen, 4, y, "Type your own answer...", askPanel.Foreground(colorDim))
		if q.onFreeRow() {
			a.screen.ShowCursor(4, y)
		}
		return
	}
	line := strings.Join(q.text, "")
	put(a.screen, 4, y, truncate(strings.ReplaceAll(line, "\n", " "), w-5), askPanel)
	if q.onFreeRow() {
		a.screen.ShowCursor(min(w-1, 4+displaywidth.String(strings.Join(q.text[:q.cursor], ""))), y)
	}
}
