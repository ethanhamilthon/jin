package ui

import (
	"errors"
	"jin/internal/provider"
	"jin/internal/store"
)

// activateProvider makes a saved provider the active one and asks for its model.
func (a *app) configuredProvider(id string) (store.ProviderEntry, bool) {
	for _, entry := range a.cfg.Providers {
		if entry.ID == id {
			return entry, true
		}
	}
	return store.ProviderEntry{}, false
}

func (a *app) activateProvider(id string) error {
	entry, found := a.configuredProvider(id)
	if !found {
		return errors.New("provider not found")
	}
	if id == a.cfg.ActiveProvider {
		if s := a.active; s != nil && s.providerMissing {
			return a.rebindProvider(s)
		}
		return nil
	}
	cfg := provider.Config{Kind: entry.Kind, BaseURL: entry.BaseURL, APIKey: entry.APIKey}
	scope, err := a.store.LoadScopeFor(id)
	if err != nil {
		return err
	}
	a.openModelPicker(provider.NewClient(cfg), scope, func(model, effort string) error {
		if _, ok := a.configuredProvider(id); !ok {
			return errors.New("provider not found")
		}
		if err := a.store.SetActiveProvider(id); err != nil {
			return err
		}
		if err := a.store.SaveModel(model, effort); err != nil {
			return err
		}
		return a.defaultProviderChanged(model, effort)
	})
	return nil
}

func (a *app) confirmDeleteProvider(id string) {
	name := id
	for _, p := range a.cfg.Providers {
		if p.ID == id {
			name = p.Name
		}
	}
	a.confirmDelete(name, func(string) error {
		if err := a.store.DeleteProvider(id); err != nil {
			return err
		}
		return a.reloadProviders()
	}, func() { a.showProviders("") })
}

// reloadProviders re-reads the provider list and the active provider from the
// store, so the model list, scope and picker follow it.
func (a *app) reloadProviders() error {
	cfg, err := a.store.LoadConfig()
	if err != nil {
		return err
	}
	a.cfg.Providers, a.cfg.ActiveProvider, a.cfg.Provider, a.cfg.Scope = cfg.Providers, cfg.ActiveProvider, cfg.Provider, cfg.Scope
	if a.client != nil {
		a.client.Configure(cfg.Provider)
	}
	a.markMissingProviders()
	return nil
}

// defaultProviderChanged updates defaults for future sessions without rebinding work.
func (a *app) defaultProviderChanged(model, effort string) error {
	if err := a.reloadProviders(); err != nil {
		return err
	}
	a.cfg.Model, a.cfg.Effort = model, effort
	a.rebindMissingSessions()
	return nil
}

// rebindMissingSessions moves the sessions whose provider was deleted to the
// provider the user just picked.
func (a *app) rebindMissingSessions() {
	for _, s := range a.sessions {
		if s.providerMissing {
			_ = a.rebindProvider(s)
		}
	}
}

// providerChanged completes provider setup and updates only an unconfigured onboarding session.
func (a *app) providerChanged(model, effort string) error {
	onboarding := a.onboarding()
	if err := a.reloadProviders(); err != nil {
		return err
	}
	a.cfg.Model, a.cfg.Effort = model, effort
	if s := a.active; onboarding && s != nil && s.provider == "" && !s.persisted && !s.working && len(s.pending) == 0 {
		s.provider = a.cfg.ActiveProvider
		s.client.Configure(a.cfg.Provider)
		s.model, s.effort = model, effort
		a.refreshIntro()
	}
	return nil
}
