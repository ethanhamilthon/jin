package ui

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

// mentionKey lets the open panel take navigation and completion keys. Enter
// and Tab turn the highlighted name into a prompt token; only a token expands
// on send.
func (a *app) mentionKey(ev *tcell.EventKey) bool {
	m := a.mention
	if m == nil {
		return false
	}
	switch ev.Key() {
	case tcell.KeyUp:
		m.sel.move(-1)
	case tcell.KeyDown:
		m.sel.move(1)
	case tcell.KeyEscape:
		s := a.active
		a.closed = closedToken{session: s.id, kind: '#', start: m.start, query: strings.Join(m.sel.query, "")}
		a.mention = nil
	case tcell.KeyTab:
		a.acceptMention()
	case tcell.KeyEnter:
		if a.pasting || ev.Modifiers()&(tcell.ModShift|tcell.ModAlt) != 0 {
			return false
		}
		a.acceptMention()
	default:
		return false
	}
	return true
}

func (a *app) acceptMention() {
	s, m := a.active, a.mention
	completion := []string{promptToken(m.sel.current())}
	if s.cursor == len(s.input) || strings.TrimSpace(s.input[s.cursor]) != "" {
		completion = append(completion, " ")
	}
	replaceRange(&s.input, &s.cursor, m.start, s.cursor, completion...)
}
