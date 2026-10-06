package session

import "errors"

// SetModel picks the model and effort of a session. A session of the
// default provider also makes them the default for new sessions.
func (m *Manager) SetModel(id, model, effort string) error {
	cfg, err := m.db.LoadConfig()
	if err != nil {
		return err
	}
	return m.Do(id, func(s *Session) error {
		if s.providerMissing || !s.client.Config().Ready() {
			return errors.New("Provider is not ready: pick one in Providers")
		}
		if s.provider == cfg.ActiveProvider {
			if err := m.db.SaveModel(model, effort); err != nil {
				return err
			}
		}
		efforts := cfg.ModelEfforts
		if efforts == nil {
			efforts = map[string]string{}
		}
		efforts[model] = effort
		if err := m.db.SaveEfforts(efforts); err != nil {
			return err
		}
		s.model, s.effort = model, effort
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
