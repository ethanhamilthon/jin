package store

import (
	"database/sql"
	"path/filepath"
	"time"
)

func persistProjectGroups(tx *sql.Tx, groups map[string]*migrationProject, sessions []migrationSession, sessionKeys map[string]string) error {
	if _, err := tx.Exec(`UPDATE sessions SET project_id = NULL`); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM projects`); err != nil {
		return err
	}
	now := time.Now().Unix()
	for key, group := range groups {
		if group.id == "" {
			id, err := projectID()
			if err != nil {
				return err
			}
			group.id = id
		}
		if group.name == "" {
			group.name = filepath.Base(group.path)
			if group.name == "." || group.name == string(filepath.Separator) {
				group.name = group.path
			}
		}
		if group.created == 0 {
			group.created = now
		}
		if group.opened < group.latestUpdate {
			group.opened = group.latestUpdate
		}
		if group.opened == 0 {
			group.opened = group.created
		}
		lastSession := migratedLastSession(key, group, sessionKeys)
		if _, err := tx.Exec(`INSERT INTO projects(id, path, name, last_session_id, created_at, last_opened_at) VALUES (?, ?, ?, ?, ?, ?)`,
			group.id, group.path, group.name, lastSession, group.created, group.opened); err != nil {
			return err
		}
	}
	for _, session := range sessions {
		key := sessionKeys[session.id]
		if key == "" {
			continue
		}
		if group := groups[key]; group != nil {
			if _, err := tx.Exec(`UPDATE sessions SET project_id = ? WHERE id = ?`, group.id, session.id); err != nil {
				return err
			}
		}
	}
	return nil
}

func migratedLastSession(key string, group *migrationProject, sessionKeys map[string]string) string {
	for _, id := range group.lastSessions {
		if sessionKeys[id] == key {
			return id
		}
	}
	if sessionKeys[group.latestSession] == key {
		return group.latestSession
	}
	return ""
}
