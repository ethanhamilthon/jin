package store

import "time"

type Session struct {
	ID        string
	Path      string
	Model     string
	Effort    string
	Title     string
	CreatedAt time.Time
	UpdatedAt time.Time
	Usage     Usage
}

type Usage struct {
	Input   int
	Output  int
	Context int
}

// Touch creates the session on first use and refreshes model, effort, and
// updated_at afterwards. Title and path are fixed at creation.
func (db *DB) Touch(id, path, model, effort, title string) error {
	now := time.Now().Unix()
	_, err := db.sql.Exec(`
		INSERT INTO sessions (id, path, model, effort, title, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET model = excluded.model, effort = excluded.effort, updated_at = excluded.updated_at
	`, id, path, model, effort, title, now, now)
	return err
}

func (db *DB) SaveUsage(id string, u Usage) error {
	_, err := db.sql.Exec(`UPDATE sessions SET input_tokens = ?, output_tokens = ?, context_tokens = ? WHERE id = ?`,
		u.Input, u.Output, u.Context, id)
	return err
}

func (db *DB) ListByPath(path string) ([]Session, error) {
	rows, err := db.sql.Query(`SELECT id, path, model, effort, title, created_at, updated_at,
		input_tokens, output_tokens, context_tokens
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
			&s.Usage.Input, &s.Usage.Output, &s.Usage.Context); err != nil {
			return nil, err
		}
		s.CreatedAt, s.UpdatedAt = time.Unix(created, 0), time.Unix(updated, 0)
		out = append(out, s)
	}
	return out, rows.Err()
}
