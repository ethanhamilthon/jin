package session

import (
	"errors"
	"slices"

	"jin/internal/core"
	"jin/internal/tools"
)

// UndoPreview lists the files of the last turn that changed since the agent
// wrote them; undo leaves them as they are.
func (m *Manager) UndoPreview(id string) (skipped []string, err error) {
	err = m.Do(id, func(s *Session) error {
		_, changes, err := s.lastChanges()
		if err != nil {
			return err
		}
		plan, err := tools.PlanRevert(changes)
		skipped = plan.Skipped
		return err
	})
	return skipped, err
}

// Undo restores the files the agent changed in its last turn that changed
// files and tells the model with the next message.
func (m *Manager) Undo(id string) error {
	return m.Do(id, func(s *Session) error {
		turn, changes, err := s.lastChanges()
		if err != nil {
			return err
		}
		result, failure := tools.Revert(changes)
		if failure != nil && len(result.Restored) == 0 {
			return errors.New("Undo failed: " + failure.Error())
		}
		left := s.forgetUndone(turn, changes, result, failure)
		kind := core.UpdateInfo
		if failure != nil {
			kind = core.UpdateError
		}
		s.add(Entry{Kind: kind, Text: UndoReport(result, left, failure)})
		if len(result.Restored) > 0 {
			s.undoNote += core.UndoBlock(result.Restored)
		}
		return nil
	})
}

func (s *Session) lastChanges() (int, []tools.Change, error) {
	if s.working {
		return 0, nil, errors.New("The agent is working; stop it first, then undo")
	}
	if !s.persisted {
		return 0, nil, errors.New("Nothing to undo")
	}
	turn, changes, err := s.m.db.LastChanges(s.id)
	if err == nil && turn == 0 {
		err = errors.New("Nothing to undo")
	}
	return turn, changes, err
}

// forgetUndone drops the records of an undone turn; after a failed undo the
// files that were not reached stay, so undo can try again.
func (s *Session) forgetUndone(turn int, changes []tools.Change, result tools.RevertResult, failure error) []string {
	if err := s.m.db.DropTurn(s.id, turn); err != nil || failure == nil {
		return nil
	}
	handled := append(slices.Clone(result.Restored), result.Skipped...)
	var left []string
	for _, c := range changes {
		if slices.Contains(handled, c.Path) {
			continue
		}
		if !slices.Contains(left, c.Path) {
			left = append(left, c.Path)
		}
		_ = s.m.db.SaveChange(s.id, turn, c)
	}
	return left
}
