package session

func (m *Manager) startTitle(s *Session, prompt string) {
	if s.providerMissing {
		return
	}
	id, providerID, model, effort := s.id, s.provider, s.model, s.effort
	go func() {
		cfg, err := m.db.LoadConfig()
		var title string
		if err == nil {
			title, err = NamePrompt(m.ctx, m.db, cfg, id, providerID, model, effort, prompt)
		}
		if err != nil {
			m.publish(Event{Type: "notice", Session: id, Text: "Title was not generated: " + err.Error()})
			return
		}
		if title == "" {
			return
		}
		m.mu.Lock()
		defer m.mu.Unlock()
		if live, ok := m.sessions[id]; ok {
			live.title = title
			live.emitState()
		}
		m.publish(Event{Type: "sessions"})
	}()
}
