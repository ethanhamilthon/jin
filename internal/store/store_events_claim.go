package store

// ClaimAsyncEvents takes the unclaimed events of a directory for this
// process. Events of a session that another live process owns stay for that
// owner. They stay in the table until AckAsyncEvents, so a process that dies
// before delivery does not lose them (see ReleaseDeadAsyncClaims).
func (db *DB) ClaimAsyncEvents(path string, pid int) ([]AsyncEvent, error) {
	busy, err := db.busySessions(pid)
	if err != nil {
		return nil, err
	}
	tx, err := db.sql.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(`SELECT id, session_id, path, text FROM async_events WHERE path = ? AND claimed_by = 0 ORDER BY id`, path)
	if err != nil {
		return nil, err
	}
	var events []AsyncEvent
	for rows.Next() {
		var e AsyncEvent
		if err := rows.Scan(&e.ID, &e.SessionID, &e.Path, &e.Text); err != nil {
			rows.Close()
			return nil, err
		}
		if !busy[e.SessionID] {
			events = append(events, e)
		}
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}
	for _, e := range events {
		if _, err := tx.Exec(`UPDATE async_events SET claimed_by = ? WHERE id = ?`, pid, e.ID); err != nil {
			return nil, err
		}
	}
	return events, tx.Commit()
}
