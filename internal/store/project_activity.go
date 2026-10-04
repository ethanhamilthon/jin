package store

// SessionProjects maps every session to the path of its project, or to its
// own path when it has none.
func (db *DB) SessionProjects() (map[string]string, error) {
	rows, err := db.sql.Query(`SELECT s.id, COALESCE(p.path, s.path) FROM sessions s LEFT JOIN projects p ON p.id = s.project_id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var id, path string
		if err := rows.Scan(&id, &path); err != nil {
			return nil, err
		}
		out[id] = path
	}
	return out, rows.Err()
}

// AllUnread returns every session with an answer the user has not seen.
func (db *DB) AllUnread() (map[string]bool, error) {
	rows, err := db.sql.Query(`SELECT session_id FROM unread_sessions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}
