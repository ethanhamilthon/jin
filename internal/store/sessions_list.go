package store

import (
	"database/sql"
	"errors"
	"time"
)

func (db *DB) ListByPath(path string) ([]Session, error) {
	canonical, err := canonicalStoredPath(path)
	if err != nil {
		return nil, err
	}
	rows, err := db.sql.Query(`SELECT id, path, COALESCE(project_id, ''), model, effort, title, created_at, updated_at,
		input_tokens, output_tokens, context_tokens, cost, provider
		FROM sessions WHERE path = ? OR project_id IN (SELECT id FROM projects WHERE path = ?)
		ORDER BY updated_at DESC`, path, canonical)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Session
	for rows.Next() {
		var s Session
		var created, updated int64
		if err := rows.Scan(&s.ID, &s.Path, &s.ProjectID, &s.Model, &s.Effort, &s.Title, &created, &updated,
			&s.Usage.Input, &s.Usage.Output, &s.Usage.Context, &s.Usage.Cost, &s.Provider); err != nil {
			return nil, err
		}
		s.CreatedAt, s.UpdatedAt = time.Unix(created, 0), time.Unix(updated, 0)
		out = append(out, s)
	}
	return out, rows.Err()
}

// GetSession returns one session by id; ok is false when there is none.
func (db *DB) GetSession(id string) (Session, bool, error) {
	var s Session
	var created, updated int64
	err := db.sql.QueryRow(`SELECT id, path, COALESCE(project_id, ''), model, effort, title, created_at, updated_at,
		input_tokens, output_tokens, context_tokens, cost, provider FROM sessions WHERE id = ?`, id).
		Scan(&s.ID, &s.Path, &s.ProjectID, &s.Model, &s.Effort, &s.Title, &created, &updated,
			&s.Usage.Input, &s.Usage.Output, &s.Usage.Context, &s.Usage.Cost, &s.Provider)
	if errors.Is(err, sql.ErrNoRows) {
		return Session{}, false, nil
	}
	if err != nil {
		return Session{}, false, err
	}
	s.CreatedAt, s.UpdatedAt = time.Unix(created, 0), time.Unix(updated, 0)
	return s, true, nil
}

// SessionByPrefix finds the one session whose id starts with prefix.
func (db *DB) SessionByPrefix(prefix string) (Session, bool, error) {
	if prefix == "" {
		return Session{}, false, nil
	}
	rows, err := db.sql.Query(`SELECT id FROM sessions WHERE substr(id, 1, ?) = ? LIMIT 2`, len(prefix), prefix)
	if err != nil {
		return Session{}, false, err
	}
	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return Session{}, false, err
		}
		ids = append(ids, id)
	}
	rows.Close()
	switch len(ids) {
	case 0:
		return Session{}, false, nil
	case 1:
		return db.GetSession(ids[0])
	}
	return Session{}, false, errors.New("session id prefix " + prefix + " matches several sessions")
}
