package session

import (
	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/tools"
)

const maxTitle = 60

func (s *Session) persistMessage(msg provider.Message) {
	if s.persisted {
		if err := s.m.db.AppendMessage(s.id, msg); err != nil {
			s.persistenceError("History was not saved", err)
		}
	}
}

// touch creates the session row on the first prompt and refreshes model,
// effort, and recency on every prompt after that.
func (s *Session) touch(prompt string) {
	title := s.title
	if title == "" {
		title = Cut(FirstLine(prompt), maxTitle)
	}
	if err := s.m.db.TouchProvider(s.id, s.path, s.model, s.effort, title, s.provider); err != nil {
		s.persistenceError("Session was not saved", err)
		return
	}
	first := !s.persisted
	s.persisted, s.title = true, title
	_ = s.m.db.SetRunning(s.id, true)
	_ = s.m.db.RememberProject(s.path, s.id)
	if first {
		s.m.publish(Event{Type: "sessions"})
	}
}

func (s *Session) persistUsage() {
	if s.persisted {
		if err := s.m.db.SaveUsage(s.id, s.usage); err != nil {
			s.persistenceError("Usage was not saved", err)
		}
	}
}

func (s *Session) persistenceError(label string, err error) {
	s.add(Entry{Kind: core.UpdateError, Text: label + ": " + err.Error()})
}

// recordChanges saves the files an edit or write call changed, grouped by
// the turn they belong to, so undo can revert the whole turn.
func (s *Session) recordChanges(changes []tools.Change) {
	if !s.persisted || len(changes) == 0 {
		return
	}
	if s.changeTurn == 0 {
		turn, err := s.m.db.NextTurn(s.id)
		if err != nil {
			s.persistenceError("File changes were not saved", err)
			return
		}
		s.changeTurn = turn
	}
	for _, change := range changes {
		if err := s.m.db.SaveChange(s.id, s.changeTurn, change); err != nil {
			s.persistenceError("File changes were not saved", err)
			return
		}
	}
}
