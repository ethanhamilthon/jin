package store

import (
	"database/sql"
	"encoding/json"
	"errors"
)

func (db *DB) mutateProviders(change func([]ProviderEntry, string) ([]ProviderEntry, string, error)) error {
	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	values := map[string]string{}
	for _, key := range []string{keyProviders, keyActiveProvider} {
		var value string
		if err = tx.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&value); err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		values[key] = value
	}
	active, list := activeProviderFrom(values)
	list, active, err = change(list, active)
	if err != nil {
		return err
	}
	if list == nil {
		list = []ProviderEntry{}
	}
	if _, ok := findProvider(list, active); !ok {
		active = ""
		if len(list) > 0 {
			active = list[0].ID
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
	for key, value := range updates {
		if _, err = tx.Exec(`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value); err != nil {
			return err
		}
	}
	return tx.Commit()
}
