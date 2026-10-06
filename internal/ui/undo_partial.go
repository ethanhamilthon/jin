package ui

import (
	"jin/internal/session"
	"slices"

	"jin/internal/tools"
)

// forgetUndone drops the records of an undone turn. After a failed undo the
// records of the files that were not reached stay, so /undo can try again.
func (s *chatSession) forgetUndone(turn int, changes []tools.Change, result tools.RevertResult, failure error) []string {
	if err := s.store.DropTurn(s.id, turn); err != nil {
		s.persistenceError("Undo was not recorded", err)
		return nil
	}
	if failure == nil {
		return nil
	}
	handled := map[string]bool{}
	for _, path := range append(slices.Clone(result.Restored), result.Skipped...) {
		handled[path] = true
	}
	var left []string
	for _, c := range changes {
		if handled[c.Path] {
			continue
		}
		if !slices.Contains(left, c.Path) {
			left = append(left, c.Path)
		}
		if err := s.store.SaveChange(s.id, turn, c); err != nil {
			s.persistenceError("Failed files were not kept for /undo", err)
		}
	}
	return left
}

func undoReport(result tools.RevertResult, left []string, failure error) string {
	return session.UndoReport(result, left, failure)
}
