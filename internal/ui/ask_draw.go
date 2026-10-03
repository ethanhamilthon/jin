package ui

import (
	"strconv"
	"strings"

	"github.com/clipperhouse/displaywidth"
)

// askLines lays the current question out as rows and returns the row that
// holds the selection.
func (q *askState) lines(width int) (rows []string, selected int) {
	title := q.question().Question
	if len(q.questions) > 1 {
		title = "(" + strconv.Itoa(q.current+1) + "/" + strconv.Itoa(len(q.questions)) + ") " + title
	}
	rows = append(rows, wrapChat(title, width-2)...)
	for i, option := range q.question().Options {
		if i == q.row {
			selected = len(rows)
		}
		marker := "  "
		if i == q.row {
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
		rows = append(rows, marker+strconv.Itoa(i+1)+". "+option)
	}
	if q.onFreeRow() {
		selected = len(rows)
	}
	rows = append(rows, "")
	return rows, selected
}

// askHeight is how many rows the block takes: the question and options, then
// the free-text row.
func (q *askState) height(width int) int {
	rows, _ := q.lines(width)
	return min(maxBlockRows, len(rows))
}

func (a *app) drawAsk(q *askState, top, height, w int) {
	a.screen.HideCursor()
	rows, selected := q.lines(w)
	rows = rows[:len(rows)-1] // the free-text row is drawn separately
	questionRows := height - 1
	q.top = fitScroll(q.top, min(selected, len(rows)-1), len(rows), questionRows)
	for i := 0; i < questionRows && q.top+i < len(rows); i++ {
		style := base
		if strings.HasPrefix(rows[q.top+i], "› ") {
			style = accent.Bold(true)
		}
		put(a.screen, 2, top+i, truncate(rows[q.top+i], w-3), style)
	}
	y := top + height - 1
	prefix, style := "  ", dim
	if q.onFreeRow() {
		prefix, style = "› ", accent.Bold(true)
	}
	put(a.screen, 2, y, prefix, style)
	if len(q.text) == 0 {
		put(a.screen, 4, y, "Type your own answer...", dim)
		if q.onFreeRow() {
			a.screen.ShowCursor(4, y)
		}
		return
	}
	line := strings.Join(q.text, "")
	put(a.screen, 4, y, truncate(strings.ReplaceAll(line, "\n", " "), w-5), base)
	if q.onFreeRow() {
		a.screen.ShowCursor(min(w-1, 4+displaywidth.String(strings.Join(q.text[:q.cursor], ""))), y)
	}
}
