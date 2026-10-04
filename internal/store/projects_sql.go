package store

import (
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"time"
)

type rowScanner interface{ Scan(...any) error }

func scanProject(row rowScanner) (Project, error) {
	var project Project
	var created, opened int64
	err := row.Scan(&project.ID, &project.Path, &project.Name, &project.LastSession, &created, &opened)
	if err != nil {
		return Project{}, err
	}
	project.CreatedAt, project.LastOpenedAt = projectTime(created), projectTime(opened)
	return project, nil
}

func projectTime(value int64) time.Time {
	if value == 0 {
		return time.Time{}
	}
	return time.Unix(value, 0)
}

func ensureProjectTx(tx *sql.Tx, path string, now int64) (Project, error) {
	project, err := scanProject(tx.QueryRow(`SELECT id, path, name, last_session_id, created_at, last_opened_at FROM projects WHERE path = ?`, path))
	if err == nil {
		return project, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return Project{}, err
	}
	id, err := projectID()
	if err != nil {
		return Project{}, err
	}
	name := filepath.Base(path)
	if name == "." || name == string(filepath.Separator) {
		name = path
	}
	if _, err := tx.Exec(`INSERT INTO projects(id, path, name, created_at, last_opened_at) VALUES (?, ?, ?, ?, ?)`, id, path, name, now, now); err != nil {
		return Project{}, err
	}
	return scanProject(tx.QueryRow(`SELECT id, path, name, last_session_id, created_at, last_opened_at FROM projects WHERE id = ?`, id))
}

func bindSessionTx(tx *sql.Tx, session string, project Project) error {
	var linked sql.NullString
	var path string
	err := tx.QueryRow(`SELECT project_id, path FROM sessions WHERE id = ?`, session).Scan(&linked, &path)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("last session %q does not exist", session)
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
	if path == "" {
		return errors.New("session has no project path")
	}
	canonical, err := canonicalStoredPath(path)
	if err != nil {
		return err
	}
	if canonical != project.Path {
		return errors.New("session belongs to another project")
	}
	_, err = tx.Exec(`UPDATE sessions SET project_id = ? WHERE id = ? AND project_id IS NULL`, project.ID, session)
	return err
}
