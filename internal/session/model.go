package session

import (
	"errors"
	"jin/internal/sources"
)

// SetModel picks the model and effort of a session. A session of the
// default provider also makes them the default for new sessions.
func (m *Manager) SetModel(id, model, effort string) error {
	return m.SetModelProvider(id, "", model, effort)
}

func (m *Manager) SetModelProvider(id, providerID, model, effort string) error {
	cfg, err := m.db.LoadConfig()
	if err != nil {
		return err
	}
	return m.Do(id, func(s *Session) error {
		if providerID == "" {
			providerID = s.provider
		}
		if s.busy() && providerID != s.provider {
			return errors.New("wait for this turn before changing providers")
		}
		entry, err := m.db.Provider(providerID)
		if err != nil {
			return err
		}
		if err = m.db.ChooseModel(s.id, providerID, model, effort, s.provider == cfg.ActiveProvider); err != nil {
			return err
		}
		s.client.Bind(sources.Client(m.db, entry))
		s.provider, s.providerMissing, s.model, s.effort = providerID, false, model, effort
		s.emitState()
		return nil
	})
}

// Seen marks a session as read.
func (m *Manager) Seen(id string) error {
	return m.Do(id, func(s *Session) error {
		if !s.unread {
			return nil
		}
		s.unread = false
		if s.persisted {
			_ = m.db.SetUnread(s.id, false)
		}
		s.emitState()
		m.publish(Event{Type: "sessions"})
		return nil
	})
}
