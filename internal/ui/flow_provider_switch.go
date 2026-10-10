package ui

import (
	"errors"
	"jin/internal/sources"
	"jin/internal/store"
)

func (a *app) configuredProvider(id string) (store.ProviderEntry, bool) {
	for _, entry := range a.cfg.Providers {
		if entry.ID == id {
			return entry, true
		}
	}
	return store.ProviderEntry{}, false
}

// activateProvider makes a saved provider the active one and asks for its model.
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
	if entry.Disabled {
		return errors.New("provider is disabled")
	}
	scope, err := a.store.LoadScopeFor(id)
	if err != nil {
		return err
	}
	a.openModelPicker(sources.Client(a.store, entry), scope, func(model, effort string) error {
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
