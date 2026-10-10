package store

import (
	"errors"

	"jin/internal/provider"
)

func (db *DB) SaveProvider(cfg provider.Config, model, effort string) error {
	providers, active, err := db.LoadProviders()
	if err != nil {
		return err
	}
	kind := cfg.Kind
	if kind == "" {
		kind = "openai"
	}
	found := false
	for i, p := range providers {
		if p.ID == active {
			providers[i].BaseURL = cfg.BaseURL
			providers[i].APIKey = cfg.APIKey
			if cfg.Kind != "" {
				providers[i].Kind = cfg.Kind
			}
			found = true
			break
		}
	}
	if !found {
		entry := ProviderEntry{ID: "default", Name: "default", Kind: kind, BaseURL: cfg.BaseURL, APIKey: cfg.APIKey}
		providers, active = append(providers, entry), entry.ID
	}
	if err := db.SaveProviders(providers, active); err != nil {
		return err
	}
	return db.SaveModel(model, effort)
}

func (db *DB) AddProvider(entry ProviderEntry) error {
	if entry.Kind == "" {
		entry.Kind = "openai"
	}
	return db.mutateProviders(func(providers []ProviderEntry, active string) ([]ProviderEntry, string, error) {
		for i, p := range providers {
			if p.ID == entry.ID {
				providers[i] = entry
				return providers, active, nil
			}
		}
		providers = append(providers, entry)
		if active == "" {
			active = entry.ID
		}
		return providers, active, nil
	})
}

func (db *DB) DeleteProvider(id string) error {
	return db.mutateProviders(func(providers []ProviderEntry, active string) ([]ProviderEntry, string, error) {
		var remaining []ProviderEntry
		for _, p := range providers {
			if p.ID != id {
				remaining = append(remaining, p)
			}
		}
		if active == id {
			active = ""
		}
		return remaining, active, nil
	})
}

func (db *DB) SetActiveProvider(id string) error {
	return db.mutateProviders(func(providers []ProviderEntry, _ string) ([]ProviderEntry, string, error) {
		entry, ok := findProvider(providers, id)
		if !ok && id != "" {
			return nil, "", errors.New("provider not found: " + id)
		}
		if entry.Disabled {
			return nil, "", errors.New("provider is disabled: " + entry.Name)
		}
		return providers, id, nil
	})
}
