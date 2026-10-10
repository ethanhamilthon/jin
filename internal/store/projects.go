package store

import (
	"database/sql"
	"errors"
	"sort"
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
	// Archived projects are hidden from the session picker of jin web; nothing is deleted.
	Archived bool `json:"archived"`
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
	archived, err := db.archivedPaths()
	if err != nil {
		return nil, err
	}
	for i := range projects {
		projects[i].Archived = archived[projects[i].Path]
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

// GetProject finds a project by ID; found is false when it does not exist.
func (db *DB) GetProject(id string) (Project, bool, error) {
	project, err := scanProject(db.sql.QueryRow(`SELECT id, path, name, last_session_id, created_at, last_opened_at FROM projects WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Project{}, false, nil
	}
	return project, err == nil, err
}
