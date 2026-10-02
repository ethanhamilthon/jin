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
	providers, active, err := db.LoadProviders()
	if err != nil {
		return err
	}
	found := false
	for i, p := range providers {
		if p.ID == entry.ID {
			providers[i], found = entry, true
			break
		}
	}
	if !found {
		providers = append(providers, entry)
	}
	if active == "" {
		active = entry.ID
	}
	return db.SaveProviders(providers, active)
}

func (db *DB) DeleteProvider(id string) error {
	providers, active, err := db.LoadProviders()
	if err != nil {
		return err
	}
	var remaining []ProviderEntry
	for _, p := range providers {
		if p.ID != id {
			remaining = append(remaining, p)
		}
	}
	if active == id {
		active = ""
	}
	return db.SaveProviders(remaining, active)
}

func (db *DB) SetActiveProvider(id string) error {
	providers, _, err := db.LoadProviders()
	if err != nil {
		return err
	}
	if _, ok := findProvider(providers, id); !ok && id != "" {
		return errors.New("provider not found: " + id)
	}
	return db.SaveProviders(providers, id)
}
