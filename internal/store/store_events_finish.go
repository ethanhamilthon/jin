package store

import (
	"strings"
	"time"
)

const busyRetries = 5

// FinishAsyncTaskWithEvent closes a running task and queues its wake-up
// event in one transaction, so a crash can never keep one without the other.
// A task that is no longer running keeps its state and gets no event.
func (db *DB) FinishAsyncTaskWithEvent(id, status string, exitCode int, sessionID, path, text string) (bool, error) {
	var won bool
	err := retryBusy(func() error {
		tx, err := db.sql.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		now := time.Now().Unix()
		res, err := tx.Exec(`UPDATE async_tasks SET status = ?, exit_code = ?, finished_at = ? WHERE id = ? AND status = ?`,
			status, exitCode, now, id, AsyncRunning)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil || n == 0 {
			won = false
			return err
		}
		if _, err := tx.Exec(`INSERT INTO async_events (session_id, path, text, created_at) VALUES (?, ?, ?, ?)`,
			sessionID, path, text, now); err != nil {
			return err
		}
		won = true
		return tx.Commit()
	})
	return won && err == nil, err
}

// retryBusy runs fn again while the database is busy or locked by another
// jin process.
func retryBusy(fn func() error) error {
	var err error
	for attempt := 1; attempt <= busyRetries; attempt++ {
		if err = fn(); err == nil || !isBusy(err) {
			return err
		}
		time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
	}
	return err
}

func isBusy(err error) bool {
	text := err.Error()
	return strings.Contains(text, "SQLITE_BUSY") || strings.Contains(text, "SQLITE_LOCKED") ||
		strings.Contains(text, "database is locked") || strings.Contains(text, "database table is locked")
}
