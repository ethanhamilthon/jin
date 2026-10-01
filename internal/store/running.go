package store

// SetRunning records an in-flight request so a killed process can leave an
// unread indicator on the next launch.
func (db *DB) SetRunning(id string, running bool) error {
	if running {
		_, err := db.sql.Exec(`INSERT OR IGNORE INTO running_sessions(session_id) VALUES (?)`, id)
		return err
	}
	_, err := db.sql.Exec(`DELETE FROM running_sessions WHERE session_id=?`, id)
	return err
}

func (db *DB) RecoverInterrupted() error {
	_, err := db.sql.Exec(`INSERT OR IGNORE INTO unread_sessions(session_id) SELECT session_id FROM running_sessions`)
	if err != nil {
		return err
	}
	_, err = db.sql.Exec(`DELETE FROM running_sessions`)
	return err
}
