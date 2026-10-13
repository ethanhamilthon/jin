package store

import (
	"database/sql"
	"encoding/json"
	"strings"
)

const keyEpochs = "provider.epochs"

func identityChanged(old, now ProviderEntry) bool {
	return old.Kind != now.Kind || old.Protocol != now.Protocol || normalizeProviderURL(old.BaseURL) != normalizeProviderURL(now.BaseURL)
}

func readJSONMap[T any](tx *sql.Tx, key string) (map[string]T, error) {
	var raw string
	if err := tx.QueryRow(`SELECT value FROM settings WHERE key=?`, key).Scan(&raw); err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	out := map[string]T{}
	if raw != "" {
		if err := json.Unmarshal([]byte(raw), &out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func writeJSONMap[T any](tx *sql.Tx, key string, value map[string]T) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = tx.Exec(`INSERT INTO settings(key,value) VALUES(?,?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, key, string(data))
	return err
}

func invalidateProtocols(tx *sql.Tx, before, after []ProviderEntry) error {
	var stale []string
	for _, old := range before {
		now, ok := findProvider(after, old.ID)
		if !ok || identityChanged(old, now) {
			stale = append(stale, old.ID)
		}
	}
	if len(stale) == 0 {
		return nil
	}
	epochs, err := readJSONMap[int](tx, keyEpochs)
	if err != nil {
		return err
	}
	choices, err := readJSONMap[string](tx, keyProtocols)
	if err != nil {
		return err
	}
	for _, id := range stale {
		epochs[id]++
		prefix := protocolKey(id, "", "")
		prefix = prefix[:strings.Index(prefix, ":")+1+len(id)]
		for key := range choices {
			if strings.HasPrefix(key, prefix) {
				delete(choices, key)
			}
		}
	}
	if err = writeJSONMap(tx, keyEpochs, epochs); err != nil {
		return err
	}
	return writeJSONMap(tx, keyProtocols, choices)
}

func (db *DB) ProviderEpoch(id string) (int, error) {
	tx, err := db.sql.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	epochs, err := readJSONMap[int](tx, keyEpochs)
	return epochs[id], err
}
