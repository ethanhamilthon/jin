package ui

import (
	"strings"

	"github.com/gdamore/tcell/v3"

	"jin/internal/todo"
)

type todoRow struct {
	text  string
	style tcell.Style
	item  int
}

// todoRows word-wraps the items; a wrapped item keeps the indent of its text.
func todoRows(items []todo.Item, width int) []todoRow {
	var rows []todoRow
	for i, item := range items {
		mark, style := "[ ] ", todoPanel.Foreground(colorMuted)
		switch item.Status {
		case todo.InProgress:
			mark, style = "[~] ", todoPanel.Foreground(colorBlueFG).Bold(true)
		case todo.Done:
			mark, style = "[x] ", todoPanel.Foreground(colorDim)
		}
		for j, line := range wrapChat(item.Text, max(1, width-4-len(mark))) {
			prefix := strings.Repeat(" ", len(mark))
			if j == 0 {
				prefix = mark
			}
			rows = append(rows, todoRow{text: prefix + line, style: style, item: i})
		}
	}
	return rows
}

// todoFocus is the first item in progress, or the first pending one.
func todoFocus(items []todo.Item) int {
	focus := 0
	for i, item := range items {
		if item.Status == todo.InProgress {
			return i
		}
		if item.Status == todo.Pending && items[focus].Status == todo.Done {
			focus = i
		}
	}
	return focus
}

// drawTodos paints the pinned list. The view follows the focused item.
func (a *app) drawTodos(rows []todoRow, focus, top, height, w int) {
	s := a.active
	fillBlock(a.screen, top, height, w, todoPanel)
	focusRow := 0
	for i, row := range rows {
		if row.item == focus {
			focusRow = i
			break
		}
	}
	s.todoTop = fitScroll(s.todoTop, focusRow, len(rows), height)
	for i := 0; i < height && s.todoTop+i < len(rows); i++ {
		row := rows[s.todoTop+i]
		put(a.screen, 2, top+i, truncate(row.text, w-4), row.style)
	}
}
