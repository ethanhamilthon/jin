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
