package session

import (
	"context"
	"errors"
	"strings"

	"jin/internal/core"
	"jin/internal/sources"
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

// autoTitle names a session that has just reached the configured number of
// messages. It runs once in the background; a failure is reported, not retried.
func (m *Manager) autoTitle(s *Session) {
	cfg, err := m.db.LoadConfig()
	if err != nil || cfg.Title.After == 0 || !s.persisted {
		return
	}
	if n, err := m.db.CountMessages(s.id); err != nil || n != cfg.Title.After {
		return
	}
	id, providerID, model := s.id, s.provider, s.model
	go func() {
		if _, err := m.nameSession(m.ctx, id, providerID, model); err != nil {
			m.publish(Event{Type: "notice", Text: "Title was not generated: " + err.Error()})
		}
	}()
}

// nameSession asks the title model, or the session's own model when none is
// set, and saves what it answers. It must be called without the lock.
func (m *Manager) nameSession(ctx context.Context, id, providerID, model string) (string, error) {
	cfg, err := m.db.LoadConfig()
	if err != nil {
		return "", err
	}
	if cfg.Title.Model != "" {
		providerID, model = cfg.Title.Provider, cfg.Title.Model
	}
	client, _, missing := sources.ClientFor(m.db, cfg, providerID)
	if missing {
		return "", errors.New("no provider " + providerID)
	}
	messages, err := m.db.LoadMessages(id)
	if err != nil {
		return "", err
	}
	answer, err := core.Title(ctx, client, model, cfg.Title.Effort, core.SinceLastSummary(messages), m.registry.SchemaJSON(), cfg.Title.Prompt)
	if err != nil {
		return "", err
	}
	title := cleanTitle(answer)
	if title == "" {
		return "", errors.New("the model returned no title")
	}
	if err := m.db.SetTitle(id, title); err != nil {
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

// cleanTitle keeps the first line of an answer, without the quotes and
// full stop a model may put around a title, and cuts it to maxTitle.
func cleanTitle(answer string) string {
	return Cut(strings.Trim(FirstLine(answer), " \t\"'`*.“”‘’"), maxTitle)
}
