package ui

import (
	"strings"
	"unicode"

	"github.com/gdamore/tcell/v3"
)

func (a *app) selectorKey(ev *tcell.EventKey) {
	sel := a.sel
	switch {
	case ev.Key() == tcell.KeyEscape:
		a.sel, a.mode = nil, modeNormal
	case ev.Key() == tcell.KeyEnter && !a.pasting:
		a.submitSelector()
	case sel.tabbed && ev.Key() == tcell.KeyLeft:
		a.openTab(sel.tab - 1)
	case sel.tabbed && ev.Key() == tcell.KeyRight:
		a.openTab(sel.tab + 1)
	case sel.field:
		if ev.Key() != tcell.KeyEnter {
			handleInput(ev, &sel.query, &sel.cursor)
		}
	case ev.Key() == tcell.KeyUp:
		sel.move(-1)
	case ev.Key() == tcell.KeyDown:
		sel.move(1)
	default:
		a.searchKey(ev)
	}
}

// searchKey routes everything else to the search: a bound action, or typing.
func (a *app) searchKey(ev *tcell.EventKey) {
	sel := a.sel
	if letter, ok := ctrlLetter(ev); ok {
		if action := sel.actions[letter]; action != nil {
			action(sel.current())
		}
		return
	}
	if ev.Key() == tcell.KeyEnter {
		return
	}
	handleInput(ev, &sel.query, &sel.cursor)
	if visible := sel.visible(); len(visible) > 0 && !sel.matches(sel.index) {
		sel.index = visible[0]
	}
}

// ctrlLetter reads Ctrl+A..Z both as legacy control codes and as the kitty
// protocol's rune-with-modifier.
func ctrlLetter(ev *tcell.EventKey) (rune, bool) {
	if key := ev.Key(); key >= tcell.KeyCtrlA && key <= tcell.KeyCtrlZ {
		return 'a' + rune(key-tcell.KeyCtrlA), true
	}
	if ev.Key() == tcell.KeyRune && ev.Modifiers()&tcell.ModCtrl != 0 && len(ev.Str()) == 1 {
		return unicode.ToLower(rune(ev.Str()[0])), true
	}
	return 0, false
}

// current is the value under the cursor, or "" when nothing is selectable.
func (sel *selector) current() string {
	if sel.loading || len(sel.options) == 0 || !sel.matches(sel.index) {
		return ""
	}
	return sel.options[sel.index].value
}

// move steps through the visible options and wraps around at both ends, so
// Up from the first option lands on the last.
func (sel *selector) move(direction int) {
	visible := sel.visible()
	if len(visible) == 0 {
		return
	}
	position := -1
	for i, idx := range visible {
		if idx == sel.index {
			position = i
		}
	}
	switch {
	case position < 0 && direction < 0:
		position = len(visible) - 1
	case position < 0:
		position = 0
	default:
		position = (position + direction + len(visible)) % len(visible)
	}
	sel.index = visible[position]
}

func (a *app) submitSelector() {
	sel := a.sel
	value := strings.TrimSpace(strings.Join(sel.query, ""))
	if !sel.field {
		if sel.loading || len(sel.options) == 0 || !sel.matches(sel.index) {
			return
		}
		value = sel.options[sel.index].value
	}
	sel.err = ""
	if err := sel.submit(value); err != nil {
		sel.err = err.Error()
		return
	}
	if a.sel == sel && !sel.keepOpen {
		a.sel, a.mode = nil, modeInsert
	}
}
