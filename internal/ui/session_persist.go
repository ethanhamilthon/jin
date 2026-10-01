package ui

import (
	"jin/internal/core"
	"jin/internal/provider"
)

const maxTitle = 60

func (s *chatSession) persistMessage(msg provider.Message) {
	if !s.persisted {
		return
	}
	if err := s.store.AppendMessage(s.id, msg); err != nil {
		s.persistenceError("History was not saved", err)
	}
}

// touch creates the session row on the first prompt and refreshes model,
// effort, and recency on every prompt after that.
func (s *chatSession) touch(prompt string) {
	title := s.title
	if title == "" {
		title = truncate(firstLine(prompt), maxTitle)
	}
	if err := s.store.Touch(s.id, s.path, s.model, s.effort, title); err != nil {
		s.persistenceError("Session was not saved", err)
		return
	}
	s.persisted, s.title = true, title
	_ = s.store.SetRunning(s.id, true)
}

func (s *chatSession) persistUsage() {
	if !s.persisted {
		return
	}
	if err := s.store.SaveUsage(s.id, s.usage); err != nil {
		s.persistenceError("Usage was not saved", err)
	}
}

func (s *chatSession) persistenceError(label string, err error) {
	s.closeOpenEntry()
	s.appendEntry(chatEntry{kind: core.UpdateError, text: label + ": " + err.Error()})
}
