package store

import (
	"encoding/json"

	"jin/internal/provider"
)

// DeliverAsyncEvent saves the message that carries an event into its session
// and removes the event in one transaction. An event that is already gone
// was delivered before: nothing is saved and delivered is false, so a
// replayed event never shows up twice.
func (db *DB) DeliverAsyncEvent(eventID int64, sessionID string, msg provider.Message) (delivered bool, err error) {
	data, err := json.Marshal(msg)
	if err != nil {
		return false, err
	}
	err = retryBusy(func() error {
		tx, err := db.sql.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		res, err := tx.Exec(`DELETE FROM async_events WHERE id = ?`, eventID)
		if err != nil {
			return err
		}
		n, err := res.RowsAffected()
		if err != nil || n == 0 {
			delivered = false
			return err
		}
		if _, err := tx.Exec(`INSERT INTO messages (session_id, data) VALUES (?, ?)`, sessionID, data); err != nil {
			return err
		}
		delivered = true
		return tx.Commit()
	})
	return delivered && err == nil, err
}
