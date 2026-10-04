package ui

import (
	"errors"

	"jin/internal/core"
	"jin/internal/store"
)

func (a *app) claimShell(s *chatSession) bool {
	if s.readOnlyPID != 0 || s.bash != nil && s.bash.running {
		return false
	}
	if err := s.projectError(); err != nil {
		s.appendEntry(chatEntry{kind: core.UpdateError, text: err.Error()})
		return false
	}
	if !s.persisted || s.store == nil {
		return true
	}
	if err := s.store.SetRunning(s.id, true); err != nil {
		var busy store.ErrSessionBusy
		if errors.As(err, &busy) {
			s.makeReadOnly(busy.PID)
		} else {
			s.persistenceError("Shell could not claim the session", err)
		}
		return false
	}
	return true
}
