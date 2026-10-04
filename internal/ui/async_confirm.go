package ui

import (
	"errors"
	"os"

	"jin/internal/store"
)

// asyncAck is an event handed to a session that waits for its message to be
// saved before the event is removed from the database.
type asyncAck struct {
	id   int64
	text string
}

// deliverEvent puts the event into its session. Every failure except a
// session that no longer exists leaves the event in the database, claimed by
// this process, until the session is opened or jin starts again.
func (a *app) deliverEvent(event store.AsyncEvent) {
	s, err := a.sessionForEvent(event)
	var gone errAsyncSession
	switch {
	case errors.As(err, &gone):
		a.report(err)
		_ = a.store.AckAsyncEvents(event.ID)
		return
	case err != nil:
		a.report(err)
		return
	case s.providerMissing || s.readOnlyPID != 0 || s.hasAck(event.ID):
		return
	}
	if saved, err := a.store.HasUserMessage(s.id, event.Text); err == nil && saved {
		_ = a.store.AckAsyncEvents(event.ID)
		return
	}
	s.asyncAcks = append(s.asyncAcks, asyncAck{id: event.ID, text: event.Text})
	s.sendAsync(event.Text)
}

func (s *chatSession) hasAck(id int64) bool {
	for _, ack := range s.asyncAcks {
		if ack.id == id {
			return true
		}
	}
	return false
}

// confirmAsync acknowledges the events whose message reached the history.
func (a *app) confirmAsync() {
	for _, s := range a.sessions {
		kept := s.asyncAcks[:0]
		for _, ack := range s.asyncAcks {
			saved, err := a.store.HasUserMessage(s.id, ack.text)
			if err == nil && saved && a.store.AckAsyncEvents(ack.id) == nil {
				continue
			}
			kept = append(kept, ack)
		}
		s.asyncAcks = kept
	}
}

// retryAsync frees the events this process could not deliver to a session,
// so the next poll tries again.
func (a *app) retryAsync(sessionID string) {
	_ = a.store.ReleaseAsyncEventsFor(sessionID, os.Getpid())
}
