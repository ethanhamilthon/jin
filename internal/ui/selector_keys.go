package ui

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

func (a *app) selectorKey(ev *tcell.EventKey) {
	sel := a.sel
	switch {
	case ev.Key() == tcell.KeyEscape && (sel.field || a.mode == modeNormal):
		a.sel, a.mode = nil, modeNormal
	case ev.Key() == tcell.KeyEscape:
		a.mode = modeNormal
	case ev.Key() == tcell.KeyEnter && !a.pasting:
		a.submitSelector()
	case sel.field:
		if ev.Key() != tcell.KeyEnter {
			handleInput(ev, &sel.query, &sel.cursor)
		}
	case ev.Key() == tcell.KeyUp, a.mode == modeNormal && ev.Str() == "k":
		sel.move(-1)
	case ev.Key() == tcell.KeyDown, a.mode == modeNormal && ev.Str() == "j":
		sel.move(1)
	case a.mode == modeNormal && (ev.Str() == "i" || ev.Str() == "/"):
		a.mode = modeInsert
	case a.mode == modeInsert && ev.Key() != tcell.KeyEnter:
		handleInput(ev, &sel.query, &sel.cursor)
		if visible := sel.visible(); len(visible) > 0 && !sel.matches(sel.index) {
			sel.index = visible[0]
		}
	}
}

func (sel *selector) move(direction int) {
	for i := sel.index + direction; i >= 0 && i < len(sel.options); i += direction {
		if sel.matches(i) {
			sel.index = i
			return
		}
	}
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
	if a.sel == sel {
		a.sel, a.mode = nil, modeInsert
	}
}
