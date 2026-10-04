package ui

import (
	"errors"
	"slices"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/store"
)

func (s *chatSession) noteMissingProvider() {
	s.closeOpenEntry()
	s.appendEntry(chatEntry{kind: core.UpdateError, text: "Provider " + s.provider + " of this session was deleted. Type /provider to pick another one before you send."})
}

// sessionReady reports whether the session's own provider can take requests.
func (s *chatSession) sessionReady() bool {
	return !s.providerMissing && s.client.Config().Ready()
}

// sessionScope is the model scope of the provider the session belongs to.
func (a *app) sessionScope(s *chatSession) []string {
	if s.provider == "" || s.provider == a.cfg.ActiveProvider {
		return a.cfg.Scope
	}
	scope, _ := a.store.LoadScopeFor(s.provider)
	return scope
}

// markMissingProviders cuts off the sessions whose provider is gone.
func (a *app) markMissingProviders() {
	for _, s := range a.sessions {
		s.models, s.modelsFor = nil, ""
		if s.provider == "" || s.providerMissing || a.hasProvider(s.provider) {
			continue
		}
		s.providerMissing = true
		s.client.Configure(provider.Config{})
		s.noteMissingProvider()
	}
}

func (a *app) hasProvider(id string) bool {
	return slices.ContainsFunc(a.cfg.Providers, func(p store.ProviderEntry) bool { return p.ID == id })
}

// rebindProvider moves a session whose provider was deleted to the active one.
func (a *app) rebindProvider(s *chatSession) error {
	if !a.cfg.Provider.Ready() {
		return errors.New("provider is not ready")
	}
	s.provider, s.providerMissing = a.cfg.ActiveProvider, false
	a.retryAsync(s.id)
	s.client.Configure(a.cfg.Provider)
	if s.persisted {
		return s.store.TouchProvider(s.id, s.path, s.model, s.effort, s.title, s.provider)
	}
	return nil
}
