package store

import "jin/internal/tools"

// AllChanges returns every recorded file change of a session, ordered by
// turn and then by the order they were made.
func (db *DB) AllChanges(sessionID string) ([]tools.Change, error) {
	rows, err := db.sql.Query(`SELECT path, existed, before, after FROM file_changes
		WHERE session_id = ? ORDER BY turn, id`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var changes []tools.Change
	for rows.Next() {
		var c tools.Change
		var existed int
		if err := rows.Scan(&c.Path, &existed, &c.Before, &c.After); err != nil {
			return nil, err
		}
		c.Existed = existed == 1
		changes = append(changes, c)
	}
	return changes, rows.Err()
}
