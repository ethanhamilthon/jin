package ui

import (
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
// changed files, and tells the model about it with the next message. When
// some files changed since, it shows a preview and waits for confirmation.
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
	plan, err := tools.PlanRevert(changes)
	if err != nil {
		s.appendEntry(chatEntry{kind: core.UpdateError, text: "Undo failed: " + err.Error()})
		return
	}
	if len(plan.Skipped) > 0 {
		a.previewUndo(turn, changes, plan)
		return
	}
	a.applyUndo(turn, changes)
}

// applyUndo reverts the changes of a turn and reports what happened.
func (a *app) applyUndo(turn int, changes []tools.Change) {
	s := a.active
	result, err := tools.Revert(changes)
	if err != nil && len(result.Restored) == 0 {
		s.appendEntry(chatEntry{kind: core.UpdateError, text: "Undo failed: " + err.Error()})
		return
	}
	left := s.forgetUndone(turn, changes, result, err)
	kind := core.UpdateInfo
	if err != nil {
		kind = core.UpdateError
	}
	s.appendEntry(chatEntry{kind: kind, text: undoReport(result, left, err)})
	if len(result.Restored) > 0 {
		s.undoNote += core.UndoBlock(result.Restored)
	}
}
