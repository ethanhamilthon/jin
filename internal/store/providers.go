package store

import "encoding/json"

type ProviderEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Protocol string `json:"protocol,omitempty"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
	Disabled bool   `json:"disabled,omitempty"`
	Source   string `json:"source,omitempty"`
	Profile  string `json:"profile,omitempty"`
}

const (
	keyProviders      = "providers"
	keyActiveProvider = "provider.active"
)

func (p ProviderEntry) Removed() bool { return p.Source == "cliproxy" }

func parseProviders(value string) (list []ProviderEntry) {
	if value != "" {
		_ = json.Unmarshal([]byte(value), &list)
	}
	return list
}

func activeProviderFrom(values map[string]string) (string, []ProviderEntry) {
	providers := parseProviders(values[keyProviders])
	active := values[keyActiveProvider]
	for _, p := range providers {
		if p.ID == active {
			return active, providers
		}
	}
	if len(providers) > 0 {
		return providers[0].ID, providers
	}
	return active, providers
}

func findProvider(providers []ProviderEntry, id string) (ProviderEntry, bool) {
	for _, p := range providers {
		if p.ID == id {
			return p, true
		}
	}
	return ProviderEntry{}, false
}

func (db *DB) ActiveProviderID() (string, error) {
	values, err := db.settings()
	if err != nil {
		return "", err
	}
	active, _ := activeProviderFrom(values)
	return active, nil
}

func (db *DB) LoadProviders() ([]ProviderEntry, string, error) {
	if err := migrateProviders(db.sql); err != nil {
		return nil, "", err
	}
	if err := migrateProviderKinds(db.sql); err != nil {
		return nil, "", err
	}
	values, err := db.settings()
	if err != nil {
		return nil, "", err
	}
	active, providers := activeProviderFrom(values)
	for i := range providers {
		if providers[i].Removed() {
			providers[i].Disabled = true
		}
	}
	return providers, active, nil
}

func (db *DB) SaveProviders(list []ProviderEntry, active string) error {
	return db.mutateProviders(func([]ProviderEntry, string) ([]ProviderEntry, string, error) {
		return list, active, nil
	})
}
