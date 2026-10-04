package store

import (
	"encoding/json"

	"jin/internal/provider"
)

// HasUserMessage reports whether a session already holds a user message with
// exactly this text. The TUI uses it to see that an async event reached the
// history, and to skip an event that was delivered before a crash.
func (db *DB) HasUserMessage(sessionID, text string) (bool, error) {
	data, err := json.Marshal(provider.Message{Role: "user", Content: text})
	if err != nil {
		return false, err
	}
	var found bool
	err = db.sql.QueryRow(`SELECT EXISTS(SELECT 1 FROM messages WHERE session_id = ? AND data = ?)`, sessionID, data).Scan(&found)
	return found, err
}

// ReleaseAsyncEventsFor frees the events of a session that this process
// claimed but could not deliver, so the next poll hands them out again.
func (db *DB) ReleaseAsyncEventsFor(sessionID string, pid int) error {
	_, err := db.sql.Exec(`UPDATE async_events SET claimed_by = 0 WHERE session_id = ? AND claimed_by = ?`, sessionID, pid)
	return err
}
