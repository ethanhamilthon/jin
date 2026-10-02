package store

import (
	"database/sql"
	"errors"
	"time"
)

type Session struct {
	ID        string
	Path      string
	Model     string
	Effort    string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Usage     Usage
	Provider  string
}

type Usage struct {
	Input   int
	Output  int
	Context int
	Cost    float64
}

// Touch creates the session on first use and refreshes model, effort, and
// updated_at afterwards. Title and path are fixed at creation.
func (db *DB) Touch(id, path, model, effort, title string) error {
	return db.TouchProvider(id, path, model, effort, title, "")
}

func (db *DB) TouchProvider(id, path, model, effort, title, provider string) error {
	now := time.Now().Unix()
	_, err := db.sql.Exec(`
		INSERT INTO sessions (id, path, model, effort, title, created_at, updated_at, provider)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET model = excluded.model, effort = excluded.effort, updated_at = excluded.updated_at,
			provider = CASE WHEN excluded.provider != '' THEN excluded.provider ELSE sessions.provider END
	`, id, path, model, effort, title, now, now, provider)
	return err
}

func (db *DB) TouchWithProvider(id, path, model, effort, title, provider string) error {
	return db.TouchProvider(id, path, model, effort, title, provider)
}

func (db *DB) SetSessionProvider(id, provider string) error {
	_, err := db.sql.Exec(`UPDATE sessions SET provider = ? WHERE id = ?`, provider, id)
	return err
}

func (db *DB) SaveUsage(id string, u Usage) error {
	_, err := db.sql.Exec(`UPDATE sessions SET input_tokens = ?, output_tokens = ?, context_tokens = ?, cost = ? WHERE id = ?`,
		u.Input, u.Output, u.Context, u.Cost, id)
	return err
}

func (db *DB) ListByPath(path string) ([]Session, error) {
	rows, err := db.sql.Query(`SELECT id, path, model, effort, title, created_at, updated_at,
		input_tokens, output_tokens, context_tokens, cost, provider
		FROM sessions WHERE path = ? ORDER BY updated_at DESC`, path)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Session
	for rows.Next() {
		var s Session
		var created, updated int64
		if err := rows.Scan(&s.ID, &s.Path, &s.Model, &s.Effort, &s.Title, &created, &updated,
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
	err := db.sql.QueryRow(`SELECT id, path, model, effort, title, created_at, updated_at,
		input_tokens, output_tokens, context_tokens, cost, provider FROM sessions WHERE id = ?`, id).
		Scan(&s.ID, &s.Path, &s.Model, &s.Effort, &s.Title, &created, &updated,
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
