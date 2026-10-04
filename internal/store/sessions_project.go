package store

import (
	"database/sql"
	"errors"
	"time"
)

// TouchProvider creates or refreshes a session and its project relationship atomically.
func (db *DB) TouchProvider(id, path, model, effort, title, provider string) error {
	now := time.Now().Unix()
	return retryBusy(func() error {
		tx, err := db.sql.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		var project Project
		if path != "" {
			canonical, err := canonicalStoredPath(path)
			if err != nil {
				return err
			}
			project, err = ensureProjectTx(tx, canonical, now)
			if err != nil {
				return err
			}
			if err := validateSessionProjectTx(tx, id, project); err != nil {
				return err
			}
		}
		var projectID any
		if project.ID != "" {
			projectID = project.ID
		}
		_, err = tx.Exec(`INSERT INTO sessions (id, path, project_id, model, effort, title, created_at, updated_at, provider)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(id) DO UPDATE SET model = excluded.model, effort = excluded.effort, updated_at = excluded.updated_at,
				project_id = CASE WHEN excluded.project_id IS NOT NULL THEN excluded.project_id ELSE sessions.project_id END,
				provider = CASE WHEN excluded.provider != '' THEN excluded.provider ELSE sessions.provider END`,
			id, path, projectID, model, effort, title, now, now, provider)
		if err != nil {
			return err
		}
		if project.ID != "" {
			if _, err := tx.Exec(`UPDATE projects SET last_session_id = ?, last_opened_at = ? WHERE id = ?`, id, now, project.ID); err != nil {
				return err
			}
		}
		return tx.Commit()
	})
}

func validateSessionProjectTx(tx *sql.Tx, id string, project Project) error {
	var linked sql.NullString
	var storedPath string
	err := tx.QueryRow(`SELECT project_id, path FROM sessions WHERE id = ?`, id).Scan(&linked, &storedPath)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if linked.Valid {
		if linked.String != project.ID {
			return errors.New("session belongs to another project")
		}
		return nil
	}
	if storedPath == "" {
		return errors.New("session has no project path")
	}
	canonical, err := canonicalStoredPath(storedPath)
	if err != nil {
		return err
	}
	if canonical != project.Path {
		return errors.New("session belongs to another project")
	}
	return nil
}
