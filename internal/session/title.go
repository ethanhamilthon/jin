package session

import (
	"context"
	"errors"
)

// GenerateTitle names an open session with the title model and saves the name.
func (m *Manager) GenerateTitle(ctx context.Context, id string) (string, error) {
	var providerID, model string
	err := m.Do(id, func(s *Session) error {
		if !s.persisted {
			return errors.New("Nothing to title yet: this session has no messages")
		}
		providerID, model = s.provider, s.model
		return nil
	})
	if err != nil {
		return "", err
	}
	return m.nameSession(ctx, id, providerID, model)
}

// autoTitle names a session whose user messages reached the configured count.
// It runs in the background; a failure is reported, not retried.
func (m *Manager) autoTitle(s *Session) {
	cfg, err := m.db.LoadConfig()
	if err != nil || cfg.Title.After == 0 || !s.persisted {
		return
	}
	messages, err := m.db.LoadMessages(s.id)
	if err != nil {
		return
	}
	turns := UserTurns(messages)
	if !TitleDue(cfg.Title, turns, s.titledAt) {
		return
	}
	s.titledAt = turns
	id, providerID, model := s.id, s.provider, s.model
	go func() {
		if _, err := m.nameSession(m.ctx, id, providerID, model); err != nil {
			m.publish(Event{Type: "notice", Text: "Title was not generated: " + err.Error()})
		}
	}()
}

// nameSession names the session and shows the new title. Call it without the lock.
func (m *Manager) nameSession(ctx context.Context, id, providerID, model string) (string, error) {
	cfg, err := m.db.LoadConfig()
	if err != nil {
		return "", err
	}
	title, err := NameSession(ctx, m.db, cfg, m.registry.SchemaJSON(), id, providerID, model)
	if err != nil {
		return "", err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, ok := m.sessions[id]; ok {
		s.title = title
		s.emitState()
	}
	m.publish(Event{Type: "sessions"})
	return title, nil
}
