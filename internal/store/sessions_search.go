package store

import (
	"strings"
	"time"
)

type SessionResult struct {
	ID        string    `json:"id"`
	Date      string    `json:"date"`
	Title     string    `json:"title"`
	Snippet   string    `json:"snippet,omitempty"`
	Path      string    `json:"path"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (db *DB) ListSessions(path string, all bool) ([]SessionResult, error) {
	query := `SELECT id, path, title, updated_at FROM sessions`
	var args []any
	if !all {
		canonical, err := canonicalStoredPath(path)
		if err != nil {
			return nil, err
		}
		query += ` WHERE path = ? OR project_id IN (SELECT id FROM projects WHERE path = ?)`
		args = append(args, path, canonical)
	}
	query += ` ORDER BY updated_at DESC`
	rows, err := db.sql.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SessionResult
	for rows.Next() {
		var id, p, title string
		var updated int64
		if err := rows.Scan(&id, &p, &title, &updated); err != nil {
			return nil, err
		}
		t := time.Unix(updated, 0)
		results = append(results, SessionResult{
			ID:        id,
			Date:      t.Format("2006-01-02 15:04"),
			Title:     title,
			Path:      p,
			UpdatedAt: t,
		})
	}
	return results, rows.Err()
}

func (db *DB) SearchSessions(path string, all bool, words []string) ([]SessionResult, error) {
	if len(words) == 0 {
		return nil, nil
	}
	var conds []string
	var args []any
	if !all {
		canonical, err := canonicalStoredPath(path)
		if err != nil {
			return nil, err
		}
		conds = append(conds, "(s.path = ? OR s.project_id IN (SELECT id FROM projects WHERE path = ?))")
		args = append(args, path, canonical)
	}
	for _, w := range words {
		pat := "%" + escapeLike(strings.ToLower(w)) + "%"
		conds = append(conds, `(LOWER(coalesce(s.title, '')) LIKE ? ESCAPE '\' OR EXISTS (SELECT 1 FROM messages WHERE session_id = s.id AND json_extract(data, '$.role') = 'user' AND LOWER(coalesce(json_extract(data, '$.content'), '')) LIKE ? ESCAPE '\'))`)
		args = append(args, pat, pat)
	}
	query := `SELECT s.id, s.path, s.title, s.updated_at FROM sessions s WHERE ` +
		strings.Join(conds, " AND ") + ` ORDER BY s.updated_at DESC LIMIT 20`

	rows, err := db.sql.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SessionResult
	for rows.Next() {
		var id, p, title string
		var updated int64
		if err := rows.Scan(&id, &p, &title, &updated); err != nil {
			return nil, err
		}
		t := time.Unix(updated, 0)
		results = append(results, SessionResult{
			ID:        id,
			Date:      t.Format("2006-01-02 15:04"),
			Title:     title,
			Path:      p,
			UpdatedAt: t,
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range results {
		results[i].Snippet = db.findSnippet(results[i].ID, results[i].Title, words)
	}
	return results, nil
}
