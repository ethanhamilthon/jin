package ui

import (
	"jin/internal/core"
	"jin/internal/store"
)

func (a *app) resumeSession(rec store.Session) error {
	a.retryAsync(rec.ID)
	s, err := a.openSession(rec)
	if err != nil {
		return err
	}
	a.focus(s)
	return nil
}

// openSession starts the backend of a saved session without moving the
// focus, or returns it when it is already open.
func (a *app) openSession(rec store.Session) (*chatSession, error) {
	if s, ok := a.sessions[rec.ID]; ok {
		return s, nil
	}
	messages, err := a.store.LoadMessages(rec.ID)
	if err != nil {
		return nil, err
	}
	owner := a.busyOwner(rec.ID)
	if owner == 0 {
		for _, msg := range core.InterruptedToolMessages(messages) {
			if err := a.store.AppendMessage(rec.ID, msg); err != nil {
				return nil, err
			}
			messages = append(messages, msg)
		}
	}
	s := a.startSessionAt(rec.Path, rec.ID, rec.Provider, rec.Model, rec.Effort, core.SinceLastSummary(messages), historyToEntries(messages, a.registry))
	s.persisted, s.title, s.usage = true, rec.Title, rec.Usage
	if owner != 0 {
		s.makeReadOnly(owner)
	}
	s.agent.SetContextSize(rec.Usage.Context)
	if items, err := a.store.LoadTodos(rec.ID); err == nil {
		s.todos = items
	}
	return s, nil
}
