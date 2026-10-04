package ui

import (
	"errors"
	"os"

	"jin/internal/provider"
	"jin/internal/store"
)

// asyncAck is an event handed to a session. Its row leaves the database in
// the same transaction that saves its message.
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

// saveMessage appends a message to the history. The message of an async
// event is saved together with the removal of its event.
func (s *chatSession) saveMessage(msg provider.Message) error {
	if msg.Role == "user" {
		for i, ack := range s.asyncAcks {
			if ack.text != msg.Content {
				continue
			}
			if _, err := s.store.DeliverAsyncEvent(ack.id, s.id, msg); err != nil {
				return err
			}
			s.asyncAcks = append(s.asyncAcks[:i], s.asyncAcks[i+1:]...)
			return nil
		}
	}
	return s.store.AppendMessage(s.id, msg)
}

// retryAsync frees the events this process could not deliver to a session,
// so the next poll tries again.
func (a *app) retryAsync(sessionID string) {
	_ = a.store.ReleaseAsyncEventsFor(sessionID, os.Getpid())
}
