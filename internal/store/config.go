package store

import "jin/internal/provider"

type Config struct {
	Provider provider.Config
	Model    string
	Effort   string
	Editor   string
	Scope    []string
}

const (
	keyBaseURL = "provider.base_url"
	keyAPIKey  = "provider.api_key"
	keyModel   = "model"
	keyEffort  = "effort"
	keyEditor  = "editor"
)

func (db *DB) LoadConfig() (Config, error) {
	values, err := db.settings()
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		Provider: provider.Config{BaseURL: values[keyBaseURL], APIKey: values[keyAPIKey]},
		Model:    values[keyModel],
		Effort:   values[keyEffort],
		Editor:   values[keyEditor],
		Scope:    parseScope(values[keyScope]),
	}
	return cfg, nil
}

func (db *DB) SaveProvider(cfg provider.Config, model, effort string) error {
	return db.setSettings(map[string]string{keyBaseURL: cfg.BaseURL, keyAPIKey: cfg.APIKey, keyModel: model, keyEffort: effort})
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
