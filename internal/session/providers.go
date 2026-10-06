package session

import (
	"slices"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/store"
)

// ProvidersChanged follows a change of the saved providers: sessions whose
// provider is gone are cut off, those cut off move to the default provider,
// and fresh sessions without a working provider take the default one.
func (m *Manager) ProvidersChanged() error {
	cfg, err := m.db.LoadConfig()
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		known := slices.ContainsFunc(cfg.Providers, func(p store.ProviderEntry) bool { return p.ID == s.provider })
		switch {
		case !known && !s.providerMissing && s.provider != "" && s.persisted:
			s.providerMissing = true
			s.client.Configure(provider.Config{})
			s.add(Entry{Kind: core.UpdateError, Text: missingProviderText(s.provider)})
		case s.providerMissing && cfg.Provider.Ready():
			m.rebind(cfg, s)
		case !s.persisted && !s.busy() && (!known || !s.client.Config().Ready()) && cfg.Provider.Ready():
			s.provider, s.providerMissing = cfg.ActiveProvider, false
			s.client.Configure(cfg.Provider)
			s.model, s.effort = cfg.Model, cfg.Effort
		}
		s.emitState()
	}
	m.publish(Event{Type: "config"})
	return nil
}

// rebind moves a session whose provider was deleted to the default one.
func (m *Manager) rebind(cfg store.Config, s *Session) {
	s.provider, s.providerMissing = cfg.ActiveProvider, false
	s.client.Configure(cfg.Provider)
	m.retryTasks(s.id)
	if s.persisted {
		_ = m.db.TouchProvider(s.id, s.path, s.model, s.effort, s.title, s.provider)
	}
	s.add(Entry{Kind: core.UpdateInfo, Text: "Moved to provider " + cfg.ActiveProvider})
}
