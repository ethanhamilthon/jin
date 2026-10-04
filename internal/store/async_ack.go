package store

// ReleaseAsyncEventsFor frees the events of a session that this process
// claimed but could not deliver, so the next poll hands them out again.
func (db *DB) ReleaseAsyncEventsFor(sessionID string, pid int) error {
	_, err := db.sql.Exec(`UPDATE async_events SET claimed_by = 0 WHERE session_id = ? AND claimed_by = ?`, sessionID, pid)
	return err
}
