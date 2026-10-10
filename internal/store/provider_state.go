package store

import (
	"errors"
	"jin/internal/provider"
)

func (p ProviderEntry) Config() provider.Config {
	kind := p.Kind
	if kind == "" {
		kind = provider.KindOpenAI
	}
	return provider.Config{Kind: kind, BaseURL: p.BaseURL, APIKey: p.APIKey, Managed: p.Source == "cliproxy", Disabled: p.Disabled}
}

func (db *DB) Provider(id string) (ProviderEntry, error) {
	list, _, err := db.LoadProviders()
	if err != nil {
		return ProviderEntry{}, err
	}
	if entry, ok := findProvider(list, id); ok {
		return entry, nil
	}
	return ProviderEntry{}, errors.New("provider not found: " + id)
}

func (db *DB) SetProviderEnabled(id string, enabled bool) error {
	return db.mutateProviders(func(list []ProviderEntry, active string) ([]ProviderEntry, string, error) {
		for i := range list {
			if list[i].ID == id {
				list[i].Disabled = !enabled
				return list, active, nil
			}
		}
		return nil, active, errors.New("provider not found: " + id)
	})
}

func (db *DB) SelectModel(id, model, effort string, makeDefault bool) error {
	entry, err := db.Provider(id)
	if err != nil {
		return err
	}
	if entry.Disabled {
		return errors.New("provider is disabled: " + entry.Name)
	}
	if !makeDefault {
		return nil
	}
	return db.setSettings(map[string]string{keyActiveProvider: id, keyModel: model, keyEffort: effort, keyBaseURL: entry.BaseURL, keyAPIKey: entry.APIKey})
}
