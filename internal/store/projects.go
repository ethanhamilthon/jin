package store

import (
	"database/sql"
	"errors"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Project is a registered working directory and its most recently used session.
type Project struct {
	ID           string    `json:"id"`
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	LastSession  string    `json:"last_session,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	LastOpenedAt time.Time `json:"last_opened_at"`
}

// Projects returns the registered projects in path order.
func (db *DB) Projects() ([]Project, error) {
	rows, err := db.sql.Query(`SELECT id, path, name, last_session_id, created_at, last_opened_at FROM projects`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projects := make([]Project, 0)
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(projects, func(i, j int) bool { return projects[i].Path < projects[j].Path })
	return projects, nil
}

// EnsureProject registers an absolute path without creating its directory.
func (db *DB) EnsureProject(path string) (Project, error) {
	path, err := normalizedProjectPath(path)
	if err != nil {
		return Project{}, err
	}
	var project Project
	err = retryBusy(func() error {
		tx, err := db.sql.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		project, err = ensureProjectTx(tx, path, time.Now().Unix())
		if err != nil {
			return err
		}
		return tx.Commit()
	})
	return project, err
}

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

// GetProject finds a project by ID; found is false when it does not exist.
func (db *DB) GetProject(id string) (Project, bool, error) {
	project, err := scanProject(db.sql.QueryRow(`SELECT id, path, name, last_session_id, created_at, last_opened_at FROM projects WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, false, nil
	}
	return project, err == nil, err
}
