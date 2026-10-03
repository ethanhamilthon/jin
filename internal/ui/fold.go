package ui

import (
	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
)

// foldMode is how much of the agent's work the chat shows. Ctrl+O steps
// through the modes in a loop: output, all, no tools, messages. The values
// are saved, so foldOutput, added last, keeps the old numbers valid.
type foldMode int

const (
	foldAll foldMode = iota
	foldNoTools
	foldMessages
	foldOutput
)

func (m foldMode) shows(kind core.UpdateKind) bool {
	switch kind {
	case core.UpdateToolResult:
		return m == foldOutput
	case core.UpdateToolCall:
		return m == foldAll || m == foldOutput
	case core.UpdateReasoning:
		return m != foldMessages
	}
	return true
}

func (m foldMode) next() foldMode {
	switch m {
	case foldOutput:
		return foldAll
	case foldAll:
		return foldNoTools
	case foldNoTools:
		return foldMessages
	}
	return foldOutput
}

// hint names what the next Ctrl+O does.
func (m foldMode) hint() string {
	switch m {
	case foldOutput:
		return "Ctrl+O to hide tool output"
	case foldAll:
		return "Ctrl+O to hide tool calls"
	case foldNoTools:
		return "Ctrl+O to hide reasoning too"
	}
	return "Ctrl+O to show tool output"
}

func isFoldKey(ev *tcell.EventKey) bool { return isCtrl(ev, 'o', false) }

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
