package store

import (
	"path/filepath"
	"strings"
	"time"
)

// RenameProject sets the display name of a registered project. An empty name
// falls back to the folder name.
func (db *DB) RenameProject(path, name string) error {
	path, err := normalizedProjectPath(path)
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		name = filepath.Base(path)
	}
	return retryBusy(func() error {
		_, err := db.sql.Exec(`UPDATE projects SET name = ? WHERE path = ?`, name, path)
		return err
	})
}

// RememberProject marks a project as used and validates any supplied session.
func (db *DB) RememberProject(path, session string) error {
	path, err := normalizedProjectPath(path)
	if err != nil {
		return err
	}
	return retryBusy(func() error {
		tx, err := db.sql.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		now := time.Now().Unix()
		project, err := ensureProjectTx(tx, path, now)
		if err != nil {
			return err
		}
		if session != "" {
			if err := bindSessionTx(tx, session, project); err != nil {
				return err
			}
		}
		last := session
		if session == "" {
			last = project.LastSession
		}
		if _, err := tx.Exec(`UPDATE projects SET last_session_id = ?, last_opened_at = ? WHERE id = ?`, last, now, project.ID); err != nil {
			return err
		}
		return tx.Commit()
	})
}
