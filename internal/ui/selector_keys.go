package ui

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

func (a *app) selectorKey(ev *tcell.EventKey) {
	sel := a.sel
	before := sel.current()
	defer func() {
		if a.sel == sel && sel.onMove != nil && sel.current() != before && sel.current() != "" {
			sel.onMove(sel.current())
		}
	}()
	switch {
	case ev.Key() == tcell.KeyEscape:
		a.sel = nil
		if sel.onCancel != nil {
			sel.onCancel()
		}
	case ev.Key() == tcell.KeyEnter && !a.pasting:
		a.submitSelector()
	case sel.onChoice != nil && ev.Key() == tcell.KeyLeft:
		sel.shift(-1)
	case sel.onChoice != nil && ev.Key() == tcell.KeyRight:
		sel.shift(1)
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

// searchKey routes everything else to the list. A list with actions reads
// plain letters as actions until "/" opens its search; any other list types
// into its search right away.
func (a *app) searchKey(ev *tcell.EventKey) {
	sel := a.sel
	if sel.hasActions() && !sel.search {
		sel.actionKey(ev)
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

func (sel *selector) actionKey(ev *tcell.EventKey) {
	if ev.Key() != tcell.KeyRune || ev.Modifiers()&(tcell.ModCtrl|tcell.ModAlt) != 0 {
		return
	}
	if ev.Str() == "/" {
		sel.search = true
	} else if action := sel.actions[keyLetter(ev)]; action != nil {
		action(sel.current())
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
	if a.sel == sel && !sel.keepOpen {
		a.sel = nil
	}
}
