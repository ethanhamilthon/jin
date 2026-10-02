package store

import "jin/internal/provider"

type Config struct {
	Provider       provider.Config
	Providers      []ProviderEntry
	ActiveProvider string
	Model          string
	Effort         string
	Editor         string
	Sound          Sound
	Scope          []string
	Fold           int

	ModelEfforts map[string]string

	HooksDisabled []string
	ToolsDisabled []string

	PromptsDisabled []string
}

const (
	keyBaseURL = "provider.base_url"
	keyAPIKey  = "provider.api_key"
	keyModel   = "model"
	keyEffort  = "effort"
	keyEditor  = "editor"
)

func (db *DB) LoadConfig() (Config, error) {
	if err := migrateProviders(db.sql); err != nil {
		return Config{}, err
	}
	values, err := db.settings()
	if err != nil {
		return Config{}, err
	}
	activeID, providers := activeProviderFrom(values)
	var activeCfg provider.Config
	if entry, ok := findProvider(providers, activeID); ok {
		kind := entry.Kind
		if kind == "" {
			kind = "openai"
		}
		activeCfg = provider.Config{Kind: kind, BaseURL: entry.BaseURL, APIKey: entry.APIKey}
	}
	scopeRaw := values[scopeKey(activeID)]
	if scopeRaw == "" && (activeID == "default" || activeID == "") {
		scopeRaw = values[keyScope]
	}

	return Config{
		Provider:        activeCfg,
		Providers:       providers,
		ActiveProvider:  activeID,
		Model:           values[keyModel],
		Effort:          values[keyEffort],
		Editor:          values[keyEditor],
		Sound:           parseSound(values),
		Scope:           parseScope(scopeRaw),
		Fold:            parseFold(values[keyFold]),
		ModelEfforts:    parseEfforts(values[keyEfforts]),
		HooksDisabled:   parseHooksDisabled(values[keyHooksDisabled]),
		ToolsDisabled:   parseToolsDisabled(values[keyToolsDisabled]),
		PromptsDisabled: parsePromptsDisabled(values[keyPromptsDisabled]),
	}, nil
}

func (db *DB) SaveModel(model, effort string) error {
	return db.setSettings(map[string]string{keyModel: model, keyEffort: effort})
}

func (db *DB) settings() (map[string]string, error) {
	rows, err := db.sql.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	values := map[string]string{}
	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, err
		}
		values[key] = value
	}
	return values, rows.Err()
}

func (db *DB) setSettings(values map[string]string) error {
	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for key, value := range values {
		if _, err := tx.Exec(`INSERT INTO settings(key, value) VALUES (?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value`, key, value); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) SaveEditor(name string) error {
	return db.setSettings(map[string]string{keyEditor: name})
}
