package ui

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

// slashKey handles the command list. Enter runs the selected command;
// Tab completes it. Commands that need arguments are only completed.
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
		cmd, ok := slashByName(p.sel.current())
		if ok && !cmd.args {
			s := a.active
			cutRange(s, p.start, s.cursor)
			a.slash = nil
			cmd.run(a, "")
		} else {
			a.acceptSlash()
		}
	case tcell.KeyEscape:
		s := a.active
		a.closed = closedToken{session: s.id, kind: '/', start: p.start, query: strings.ToLower(strings.Join(s.input[p.start+1:s.cursor], ""))}
		a.slash = nil
	default:
		return false
	}
	return true
}

// acceptSlash turns the typed /name into a token of the highlighted command.
// It runs when the draft is sent.
func (a *app) acceptSlash() {
	s, p := a.active, a.slash
	if cmd, ok := slashByName(p.sel.current()); ok {
		replaceRange(&s.input, &s.cursor, p.start, s.cursor, commandToken(cmd.name), " ")
		a.slash = nil
	}
}
