package store

import (
	"database/sql"
	"encoding/json"
)

func migrateProviderKinds(db *sql.DB) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var raw string
	if err = tx.QueryRow(`SELECT value FROM settings WHERE key=?`, keyProviders).Scan(&raw); err != nil {
		return err
	}
	list := parseProviders(raw)
	changed := false
	for i := range list {
		if list[i].Removed() {
			if !list[i].Disabled {
				list[i].Disabled, changed = true, true
			}
			continue
		}
		if list[i].Kind == "openai" || list[i].Kind == "responses" || list[i].Kind == "" {
			if list[i].Protocol == "" {
				list[i].Protocol = list[i].Kind
				if list[i].Protocol == "" {
					list[i].Protocol = "openai"
				}
			}
			list[i].Kind, changed = "auto", true
		}
	}
	if !changed {
		return tx.Commit()
	}
	data, err := json.Marshal(list)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(`UPDATE settings SET value=? WHERE key=?`, string(data), keyProviders); err != nil {
		return err
	}
	return tx.Commit()
}
