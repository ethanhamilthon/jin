package ui

import (
	"strings"

	"github.com/gdamore/tcell/v3"
)

// mentionKey lets the open panel take navigation and completion keys. Enter
// completes only while the typed name is not already complete, so a finished
// #name still sends.
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
	case tcell.KeyTab:
		a.acceptMention()
	case tcell.KeyEnter:
		typed := strings.Join(m.sel.query, "")
		if a.pasting || ev.Modifiers()&(tcell.ModShift|tcell.ModAlt) != 0 || typed == m.sel.current() {
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
	completion := clusters(m.sel.current())
	if s.cursor == len(s.input) || strings.TrimSpace(s.input[s.cursor]) != "" {
		completion = append(completion, " ")
	}
	tail := append(completion, s.input[s.cursor:]...)
	s.input = append(s.input[:m.start+1], tail...)
	s.cursor = m.start + 1 + len(completion)
}
