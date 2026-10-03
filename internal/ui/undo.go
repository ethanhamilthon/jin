package ui

import (
	"strings"

	"jin/internal/core"
	"jin/internal/tools"
)

// recordChanges saves the files an edit or write call changed, grouped by
// the turn they belong to, so /undo can revert the whole turn.
func (s *chatSession) recordChanges(changes []tools.Change) {
	if !s.persisted || len(changes) == 0 {
		return
	}
	if s.changeTurn == 0 {
		turn, err := s.store.NextTurn(s.id)
		if err != nil {
			s.persistenceError("File changes were not saved", err)
			return
		}
		s.changeTurn = turn
	}
	for _, change := range changes {
		if err := s.store.SaveChange(s.id, s.changeTurn, change); err != nil {
			s.persistenceError("File changes were not saved", err)
			return
		}
	}
}

// undoLastTurn restores the files the agent changed in its last turn that
// changed files, and tells the model about it with the next message.
func (a *app) undoLastTurn() {
	s := a.active
	switch {
	case s.working:
		s.appendEntry(chatEntry{kind: core.UpdateError, text: "The agent is working; /stop it first, then /undo"})
		return
	case !s.persisted:
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Nothing to undo"})
		return
	}
	turn, changes, err := s.store.LastChanges(s.id)
	if err != nil || turn == 0 {
		if err == nil {
			s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Nothing to undo"})
		} else {
			s.persistenceError("Undo failed", err)
		}
		return
	}
	result, err := tools.Revert(changes)
	if err != nil {
		s.appendEntry(chatEntry{kind: core.UpdateError, text: "Undo failed: " + err.Error()})
		return
	}
	if err := s.store.DropTurn(s.id, turn); err != nil {
		s.persistenceError("Undo was not recorded", err)
	}
	s.appendEntry(chatEntry{kind: core.UpdateInfo, text: undoReport(result)})
	if len(result.Restored) > 0 {
		s.undoNote += core.UndoBlock(result.Restored)
	}
}

func undoReport(result tools.RevertResult) string {
	var lines []string
	if len(result.Restored) > 0 {
		lines = append(lines, "Undone: "+strings.Join(result.Restored, ", "))
	}
	if len(result.Skipped) > 0 {
		lines = append(lines, "Left as is, changed after the agent wrote them: "+strings.Join(result.Skipped, ", "))
	}
	return strings.Join(lines, "\n")
}
