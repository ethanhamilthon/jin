package store

import "jin/internal/tools"

const changesSchema = `
CREATE TABLE IF NOT EXISTS file_changes (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id TEXT NOT NULL REFERENCES sessions(id),
	turn INTEGER NOT NULL,
	path TEXT NOT NULL,
	existed INTEGER NOT NULL,
	before TEXT NOT NULL,
	after TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS file_changes_session ON file_changes(session_id, turn);
`

// NextTurn is the number for the next turn of a session that changes files.
func (db *DB) NextTurn(sessionID string) (int, error) {
	var turn int
	err := db.sql.QueryRow(`SELECT COALESCE(MAX(turn), 0) + 1 FROM file_changes WHERE session_id = ?`, sessionID).Scan(&turn)
	return turn, err
}

// SaveChange records one file write of a turn, so /undo can revert it.
func (db *DB) SaveChange(sessionID string, turn int, change tools.Change) error {
	existed := 0
	if change.Existed {
		existed = 1
	}
	_, err := db.sql.Exec(`INSERT INTO file_changes (session_id, turn, path, existed, before, after) VALUES (?, ?, ?, ?, ?, ?)`,
		sessionID, turn, change.Path, existed, change.Before, change.After)
	return err
}

// LastChanges returns the newest turn that changed files, with its changes
// in the order they were made. turn is 0 when there is none.
func (db *DB) LastChanges(sessionID string) (int, []tools.Change, error) {
	rows, err := db.sql.Query(`SELECT turn, path, existed, before, after FROM file_changes
		WHERE session_id = ? AND turn = (SELECT MAX(turn) FROM file_changes WHERE session_id = ?) ORDER BY id`, sessionID, sessionID)
	if err != nil {
		return 0, nil, err
	}
	defer rows.Close()
	var turn int
	var changes []tools.Change
	for rows.Next() {
		var c tools.Change
		var existed int
		if err := rows.Scan(&turn, &c.Path, &existed, &c.Before, &c.After); err != nil {
			return 0, nil, err
		}
		c.Existed = existed == 1
		changes = append(changes, c)
	}
	return turn, changes, rows.Err()
}

// DropTurn forgets the changes of a turn once they were undone.
func (db *DB) DropTurn(sessionID string, turn int) error {
	_, err := db.sql.Exec(`DELETE FROM file_changes WHERE session_id = ? AND turn = ?`, sessionID, turn)
	return err
}
