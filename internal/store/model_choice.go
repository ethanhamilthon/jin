package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func EffortKey(provider, model string) string {
	return fmt.Sprintf("%d:%s%s", len(provider), provider, model)
}

func (db *DB) RememberEffort(provider, model, effort string) error {
	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var raw string
	_ = tx.QueryRow(`SELECT value FROM settings WHERE key=?`, keyEfforts).Scan(&raw)
	efforts := parseEfforts(raw)
	efforts[EffortKey(provider, model)] = effort
	data, err := json.Marshal(efforts)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, keyEfforts, string(data)); err != nil {
		return err
	}
	return tx.Commit()
}

func (db *DB) ChooseModel(sessionID, id, model, effort string, makeDefault bool) error {
	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var raw string
	if err = tx.QueryRow(`SELECT value FROM settings WHERE key=?`, keyProviders).Scan(&raw); err != nil {
		return err
	}
	entry, ok := findProvider(parseProviders(raw), id)
	if !ok {
		return errors.New("provider not found: " + id)
	}
	if entry.Disabled {
		return errors.New("provider is disabled: " + entry.Name)
	}
	if entry.Source == "cliproxy" && !strings.HasPrefix(model, entry.Profile+"/") {
		return errors.New("model does not belong to this subscription provider")
	}
	var effortRaw string
	_ = tx.QueryRow(`SELECT value FROM settings WHERE key=?`, keyEfforts).Scan(&effortRaw)
	efforts := parseEfforts(effortRaw)
	efforts[EffortKey(id, model)] = effort
	data, err := json.Marshal(efforts)
	if err != nil {
		return err
	}
	updates := map[string]string{keyEfforts: string(data)}
	if makeDefault {
		updates[keyActiveProvider], updates[keyModel], updates[keyEffort] = id, model, effort
		updates[keyBaseURL], updates[keyAPIKey] = entry.BaseURL, entry.APIKey
	}
	for key, value := range updates {
		if _, err = tx.Exec(`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, value); err != nil {
			return err
		}
	}
	if sessionID != "" {
		if _, err = tx.Exec(`UPDATE sessions SET provider=?,model=?,effort=? WHERE id=?`, id, model, effort, sessionID); err != nil {
			return err
		}
	}
	return tx.Commit()
}
