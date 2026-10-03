package store

import (
	"time"
)

// AsyncEvent is a message for the agent of a session: a finished task or a
// note sent with `jin async run "echo ..."`.
type AsyncEvent struct {
	ID        int64
	SessionID string
	Path      string
	Text      string
}

// AddAsyncEvent queues a message for a session.
func (db *DB) AddAsyncEvent(sessionID, path, text string) error {
	_, err := db.sql.Exec(`INSERT INTO async_events (session_id, path, text, created_at) VALUES (?, ?, ?, ?)`,
		sessionID, path, text, time.Now().Unix())
	return err
}

// ClaimAsyncEvents takes the unclaimed events of a directory for this
// process. They stay in the table until AckAsyncEvents, so a process that
// dies before delivery does not lose them (see ReleaseDeadAsyncClaims).
func (db *DB) ClaimAsyncEvents(path string, pid int) ([]AsyncEvent, error) {
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
		events = append(events, e)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}
	if _, err := tx.Exec(`UPDATE async_events SET claimed_by = ? WHERE path = ? AND claimed_by = 0 AND id <= ?`,
		pid, path, events[len(events)-1].ID); err != nil {
		return nil, err
	}
	return events, tx.Commit()
}

// AckAsyncEvents removes events that were handed to an agent.
func (db *DB) AckAsyncEvents(ids ...int64) error {
	for _, id := range ids {
		if _, err := db.sql.Exec(`DELETE FROM async_events WHERE id = ?`, id); err != nil {
			return err
		}
	}
	return nil
}

// ReleaseDeadAsyncClaims frees events claimed by processes that are gone.
func (db *DB) ReleaseDeadAsyncClaims() error {
	rows, err := db.sql.Query(`SELECT DISTINCT claimed_by FROM async_events WHERE claimed_by != 0`)
	if err != nil {
		return err
	}
	var dead []int
	for rows.Next() {
		var pid int
		if err := rows.Scan(&pid); err != nil {
			rows.Close()
			return err
		}
		if !processAlive(pid) {
			dead = append(dead, pid)
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, pid := range dead {
		if _, err := db.sql.Exec(`UPDATE async_events SET claimed_by = 0 WHERE claimed_by = ?`, pid); err != nil {
			return err
		}
	}
	return nil
}
