package store

import "encoding/json"

type ProviderEntry struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Kind     string `json:"kind"`
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
	values, err := db.settings()
	if err != nil {
		return nil, "", err
	}
	active, providers := activeProviderFrom(values)
	return providers, active, nil
}

func (db *DB) SaveProviders(list []ProviderEntry, active string) error {
	if list == nil {
		list = []ProviderEntry{}
	}
	if _, ok := findProvider(list, active); !ok {
		if len(list) > 0 {
			active = list[0].ID
		} else {
			active = ""
		}
	}
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}
	updates := map[string]string{keyProviders: string(data), keyActiveProvider: active}
	if entry, ok := findProvider(list, active); ok {
		updates[keyBaseURL], updates[keyAPIKey] = entry.BaseURL, entry.APIKey
	}
	return db.setSettings(updates)
}
