package store

import (
	"encoding/json"

	"jin/internal/provider"
)

// AppendMessage records one conversation message. Messages are stored as
// opaque JSON so the schema never needs to migrate when provider.Message
// grows a field.
func (db *DB) AppendMessage(sessionID string, msg provider.Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = db.sql.Exec(`INSERT INTO messages (session_id, data) VALUES (?, ?)`, sessionID, data)
	return err
}

// LoadMessages returns every message for sessionID in the order it was sent.
func (db *DB) LoadMessages(sessionID string) ([]provider.Message, error) {
	rows, err := db.sql.Query(`SELECT data FROM messages WHERE session_id = ? ORDER BY id`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []provider.Message
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, err
		}
		var msg provider.Message
		if err := json.Unmarshal([]byte(data), &msg); err != nil {
			return nil, err
		}
		out = append(out, msg)
	}
	return out, rows.Err()
}
