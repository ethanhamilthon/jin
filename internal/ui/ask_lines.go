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
