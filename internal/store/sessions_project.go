package store

import (
	"database/sql"
	"errors"
	"time"
)

// TouchProvider creates or refreshes a session and its project relationship atomically.
func (db *DB) TouchProvider(id, path, model, effort, title, provider string) error {
	_, err := db.TouchFirst(id, path, model, effort, title, provider)
	return err
}

// TouchFirst is TouchProvider that also reports whether it created the session.
// A new session gets its one title claim in the same transaction.
func (db *DB) TouchFirst(id, path, model, effort, title, provider string) (created bool, err error) {
	now := time.Now().Unix()
	err = retryBusy(func() error {
		created = false
		tx, err := db.sql.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		var known int
		switch err := tx.QueryRow(`SELECT 1 FROM sessions WHERE id = ?`, id).Scan(&known); {
		case errors.Is(err, sql.ErrNoRows):
			created = true
		case err != nil:
			return err
		}
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
		if created {
			if _, err := tx.Exec(`INSERT INTO title_claims(session_id) VALUES (?)`, id); err != nil {
				return err
			}
		}
		if project.ID != "" {
			if _, err := tx.Exec(`UPDATE projects SET last_session_id = ?, last_opened_at = ? WHERE id = ?`, id, now, project.ID); err != nil {
				return err
			}
		}
		return tx.Commit()
	})
	return created && err == nil, err
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
