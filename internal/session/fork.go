package session

import "errors"

// Points lists the messages the user typed in a saved session.
func (m *Manager) Points(id string) ([]RewindPoint, error) {
	messages, err := m.db.LoadMessages(id)
	if err != nil {
		return nil, err
	}
	return RewindPoints(messages), nil
}

// Fork saves a new session with the history before point n of session id
// and opens it with that message as the draft. The original stays as it is.
func (m *Manager) Fork(id string, n int) (Snapshot, error) {
	rec, found, err := m.db.GetSession(id)
	if err != nil {
		return Snapshot{}, err
	}
	if !found {
		return Snapshot{}, errors.New("no session with id " + id)
	}
	messages, err := m.db.LoadMessages(id)
	if err != nil {
		return Snapshot{}, err
	}
	points := RewindPoints(messages)
	if n < 0 || n >= len(points) {
		return Snapshot{}, errors.New("no such message")
	}
	next := NewID()
	if err := m.db.TouchProvider(next, rec.Path, rec.Model, rec.Effort, rec.Title, rec.Provider); err != nil {
		return Snapshot{}, err
	}
	for _, msg := range messages[:points[n].Index] {
		if err := m.db.AppendMessage(next, msg); err != nil {
			return Snapshot{}, err
		}
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	s, err := m.openLocked(next)
	if err != nil {
		return Snapshot{}, err
	}
	s.setDraft(points[n].Text)
	m.publish(Event{Type: "sessions"})
	return s.snapshot(), nil
}
