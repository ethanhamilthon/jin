package store

import (
	"database/sql"
	"encoding/json"
)

type migrationProject struct {
	id, path, name  string
	lastSessions    []string
	created, opened int64
	latestUpdate    int64
	latestSession   string
}

type migrationSession struct {
	id, path, projectID string
	created, updated    int64
}

func migrateProjectRows(tx *sql.Tx) error {
	projects, err := readProjectRows(tx)
	if err != nil {
		return err
	}
	var saved []Project
	var raw string
	err = tx.QueryRow(`SELECT value FROM settings WHERE key = 'projects'`).Scan(&raw)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == nil && raw != "" {
		if err := json.Unmarshal([]byte(raw), &saved); err != nil {
			return err
		}
	}
	sessions, err := readMigrationSessions(tx)
	if err != nil {
		return err
	}
	return rebuildProjectRows(tx, projects, saved, sessions)
}

func readProjectRows(tx *sql.Tx) ([]Project, error) {
	rows, err := tx.Query(`SELECT id, path, name, last_session_id, created_at, last_opened_at FROM projects ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var projects []Project
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

func readMigrationSessions(tx *sql.Tx) ([]migrationSession, error) {
	rows, err := tx.Query(`SELECT id, path, project_id, created_at, updated_at FROM sessions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sessions []migrationSession
	for rows.Next() {
		var session migrationSession
		var projectID sql.NullString
		if err := rows.Scan(&session.id, &session.path, &projectID, &session.created, &session.updated); err != nil {
			return nil, err
		}
		if projectID.Valid {
			session.projectID = projectID.String
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}
