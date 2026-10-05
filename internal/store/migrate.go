package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
)

var addedColumns = []struct{ table, column string }{
	{"sessions", `cost REAL NOT NULL DEFAULT 0`},
	{"running_sessions", `pid INTEGER NOT NULL DEFAULT 0`},
	{"sessions", `provider TEXT NOT NULL DEFAULT ''`},
}

func migrate(db *sql.DB) error {
	for _, added := range addedColumns {
		_, err := db.Exec(`ALTER TABLE ` + added.table + ` ADD COLUMN ` + added.column)
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
	}
	if err := migrateProjects(db); err != nil {
		return err
	}
	return migrateProviders(db)
}

func migrateProviders(db *sql.DB) error {
	var providersVal string
	err := db.QueryRow(`SELECT value FROM settings WHERE key = ?`, keyProviders).Scan(&providersVal)
	if err == nil && strings.TrimSpace(providersVal) != "" {
		return nil
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	var baseURL, apiKey string
	var hasBaseURL, hasAPIKey bool

	rows, err := db.Query(`SELECT key, value FROM settings WHERE key IN (?, ?)`, keyBaseURL, keyAPIKey)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return err
		}
		if k == keyBaseURL {
			baseURL = v
			hasBaseURL = true
		} else if k == keyAPIKey {
			apiKey = v
			hasAPIKey = true
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if !hasBaseURL && !hasAPIKey {
		return nil
	}

	entry := ProviderEntry{
		ID:      "default",
		Name:    "default",
		Kind:    "openai",
		BaseURL: baseURL,
		APIKey:  apiKey,
	}
	data, err := json.Marshal([]ProviderEntry{entry})
	if err != nil {
		return err
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	upsert := `INSERT INTO settings(key, value) VALUES (?, ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value`
	if _, err := tx.Exec(upsert, keyProviders, string(data)); err != nil {
		return err
	}
	if _, err := tx.Exec(upsert, keyActiveProvider, "default"); err != nil {
		return err
	}
	return tx.Commit()
}
