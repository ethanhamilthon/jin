package ui

import (
	"errors"
	"jin/internal/provider"
	"jin/internal/store"
)

// activateProvider makes a saved provider the active one and asks for its model.
func (a *app) activateProvider(id string) error {
	var entry store.ProviderEntry
	found := false
	for _, p := range a.cfg.Providers {
		if p.ID == id {
			entry, found = p, true
		}
	}
	if !found {
		return errors.New("provider not found")
	}
	if id == a.cfg.ActiveProvider {
		if a.active.providerMissing {
			return a.rebindProvider(a.active)
		}
		return nil
	}
	cfg := provider.Config{Kind: entry.Kind, BaseURL: entry.BaseURL, APIKey: entry.APIKey}
	scope, err := a.store.LoadScopeFor(id)
	if err != nil {
		return err
	}
	a.openModelPicker(provider.NewClient(cfg), scope, func(model, effort string) error {
		if err := a.store.SetActiveProvider(id); err != nil {
			return err
		}
		if err := a.store.SaveModel(model, effort); err != nil {
			return err
		}
		return a.providerChanged(model, effort)
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
	a.client.Configure(cfg.Provider)
	a.markMissingProviders()
	return nil
}

// providerChanged applies a new active provider with its model. A session
// that has history keeps the provider it started with, so the focused one is
// replaced by a new session; an empty one just follows. A session whose
// provider was deleted moves to the new one, because the user picked it.
func (a *app) providerChanged(model, effort string) error {
	if err := a.reloadProviders(); err != nil {
		return err
	}
	if s := a.active; s.providerMissing {
		if err := a.rebindProvider(s); err != nil {
			return err
		}
	} else if s.persisted || s.working || len(s.pending) > 0 {
		a.newSession()
	} else {
		s.provider = a.cfg.ActiveProvider
		s.client.Configure(a.cfg.Provider)
	}
	a.useModel(model, effort)
	a.refreshIntro()
	return nil
}
