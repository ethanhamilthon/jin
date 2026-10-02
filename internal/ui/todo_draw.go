package ui

import (
	"jin/internal/todo"
)

// drawTodos paints the pinned list. The view follows the first item in
// progress, or the first pending one.
func (a *app) drawTodos(items []todo.Item, top, height, w int) {
	s := a.active
	focus := 0
	for i, item := range items {
		if item.Status == todo.InProgress {
			focus = i
			break
		}
		if item.Status == todo.Pending && items[focus].Status == todo.Done {
			focus = i
		}
	}
	s.todoTop = fitScroll(s.todoTop, focus, len(items), height)
	for i := 0; i < height && s.todoTop+i < len(items); i++ {
		item := items[s.todoTop+i]
		mark, style := "[ ]", muted
		switch item.Status {
		case todo.InProgress:
			mark, style = "[~]", accent.Bold(true)
		case todo.Done:
			mark, style = "[x]", dim
		}
		put(a.screen, 2, top+i, mark+" "+truncate(item.Text, w-8), style)
	}
}
