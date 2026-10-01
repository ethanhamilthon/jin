package ui

import (
	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
)

// foldMode is how much of the agent's work the chat shows. Ctrl+O steps
// through the modes in a loop.
type foldMode int

const (
	foldAll foldMode = iota
	foldNoTools
	foldMessages
)

func (m foldMode) shows(kind core.UpdateKind) bool {
	switch kind {
	case core.UpdateToolCall:
		return m == foldAll
	case core.UpdateReasoning:
		return m != foldMessages
	}
	return true
}

func (m foldMode) next() foldMode { return (m + 1) % 3 }

// hint names what the next Ctrl+O does.
func (m foldMode) hint() string {
	switch m {
	case foldAll:
		return "Ctrl+O to hide tool calls"
	case foldNoTools:
		return "Ctrl+O to hide reasoning too"
	}
	return "Ctrl+O to show everything"
}

func isFoldKey(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyCtrlO {
		return true
	}
	return ev.Key() == tcell.KeyRune && ev.Modifiers()&tcell.ModCtrl != 0 && (ev.Str() == "o" || ev.Str() == "O")
}

// cycleFold applies the next mode to every session and saves it.
func (a *app) cycleFold() {
	a.fold = a.fold.next()
	for _, s := range a.sessions {
		s.setFold(a.fold)
	}
	if err := a.store.SaveFold(int(a.fold)); err != nil {
		a.active.persistenceError("Folding was not saved", err)
	}
}

func (s *chatSession) setFold(mode foldMode) {
	s.fold, s.scroll, s.selection = mode, 0, textSelection{}
	s.rebuildRows(s.width)
}

func (s *chatSession) entryRows(entry chatEntry) []chatRow {
	if !s.fold.shows(entry.kind) {
		return nil
	}
	return entryRows(entry, s.width)
}

// lastShown is the newest entry that is on screen; blank-row spacing follows it.
func (s *chatSession) lastShown() *chatEntry {
	for i := len(s.history) - 1; i >= 0; i-- {
		if s.fold.shows(s.history[i].kind) {
			return &s.history[i]
		}
	}
	return nil
}
