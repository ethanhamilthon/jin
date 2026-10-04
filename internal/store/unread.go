package store

func (db *DB) SetUnread(id string, unread bool) error {
	if !unread {
		_, err := db.sql.Exec(`DELETE FROM unread_sessions WHERE session_id=?`, id)
		return err
	}
	_, err := db.sql.Exec(`INSERT OR IGNORE INTO unread_sessions(session_id) VALUES (?)`, id)
	return err
}

func (db *DB) UnreadSessions(path string) (map[string]bool, error) {
	canonical, err := canonicalStoredPath(path)
	if err != nil {
		return nil, err
	}
	rows, err := db.sql.Query(`SELECT u.session_id FROM unread_sessions u JOIN sessions s ON s.id=u.session_id
		WHERE s.path = ? OR s.project_id IN (SELECT id FROM projects WHERE path = ?)`, path, canonical)
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
