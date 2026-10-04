package store

import (
	"database/sql"
	"errors"
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

// RemoveProject unregisters a path. Its sessions keep their path and stay in
// the database; only the project link is cleared.
func (db *DB) RemoveProject(path string) error {
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
		var id string
		err = tx.QueryRow(`SELECT id FROM projects WHERE path = ?`, path).Scan(&id)
		if errors.Is(err, sql.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if _, err := tx.Exec(`UPDATE sessions SET project_id = NULL WHERE project_id = ?`, id); err != nil {
			return err
		}
		if _, err := tx.Exec(`DELETE FROM projects WHERE id = ?`, id); err != nil {
			return err
		}
		return tx.Commit()
	})
}
