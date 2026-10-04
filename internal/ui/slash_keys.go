package ui

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

// slashKey lets the open list take navigation, completion and Esc. Enter
// completes only while the list is open, so a typed /name never sends.
func (a *app) slashKey(ev *tcell.EventKey) bool {
	p := a.slash
	if p == nil {
		return false
	}
	switch ev.Key() {
	case tcell.KeyUp:
		p.sel.move(-1)
	case tcell.KeyDown:
		p.sel.move(1)
	case tcell.KeyTab:
		a.acceptSlash()
	case tcell.KeyEnter:
		if a.pasting || ev.Modifiers()&(tcell.ModShift|tcell.ModAlt) != 0 {
			return false
		}
		a.acceptSlash()
	case tcell.KeyEscape:
		s := a.active
		a.closed = closedToken{session: s.id, kind: '/', start: p.start, query: strings.ToLower(strings.Join(s.input[p.start+1:s.cursor], ""))}
		a.slash = nil
	default:
		return false
	}
	return true
}

// acceptSlash runs the highlighted command and cuts its token out of the
// draft. A command that takes arguments becomes a token for its arguments.
func (a *app) acceptSlash() {
	s, p := a.active, a.slash
	cmd, ok := slashByName(p.sel.current())
	if !ok {
		return
	}
	if cmd.args {
		replaceRange(&s.input, &s.cursor, p.start, s.cursor, commandToken(cmd.name), " ")
		return
	}
	cutRange(s, p.start, s.cursor)
	a.slash = nil
	cmd.run(a, "")
}
