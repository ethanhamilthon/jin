package store

import "time"

type Session struct {
	ID        string
	Path      string
	ProjectID string
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

// Touch creates a session on first use; its path and title stay fixed afterwards.
func (db *DB) Touch(id, path, model, effort, title string) error {
	return db.TouchProvider(id, path, model, effort, title, "")
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
