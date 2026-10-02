package store

import "jin/internal/todo"

// SaveTodos replaces the todo list of a session.
func (db *DB) SaveTodos(sessionID string, items []todo.Item) error {
	tx, err := db.sql.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.Exec(`DELETE FROM todos WHERE session_id=?`, sessionID); err != nil {
		return err
	}
	for i, item := range items {
		if _, err := tx.Exec(`INSERT INTO todos(session_id, position, text, status) VALUES (?, ?, ?, ?)`,
			sessionID, i, item.Text, string(item.Status)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (db *DB) LoadTodos(sessionID string) ([]todo.Item, error) {
	rows, err := db.sql.Query(`SELECT text, status FROM todos WHERE session_id=? ORDER BY position`, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []todo.Item
	for rows.Next() {
		var item todo.Item
		var status string
		if err := rows.Scan(&item.Text, &status); err != nil {
			return nil, err
		}
		item.Status = todo.Status(status)
		items = append(items, item)
	}
	return items, rows.Err()
}

// MarkTodosEdited records that the user changed the list, so the model must
// be told before its next update.
func (db *DB) MarkTodosEdited(sessionID string) error {
	_, err := db.sql.Exec(`INSERT INTO todo_state(session_id, edited) VALUES (?, 1)
		ON CONFLICT(session_id) DO UPDATE SET edited = 1`, sessionID)
	return err
}

// TakeTodosEdited reports whether the list was edited by the user and clears
// the flag.
func (db *DB) TakeTodosEdited(sessionID string) (bool, error) {
	res, err := db.sql.Exec(`UPDATE todo_state SET edited = 0 WHERE session_id=? AND edited = 1`, sessionID)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// TodosEdited reports the flag without clearing it.
func (db *DB) TodosEdited(sessionID string) (bool, error) {
	var edited int
	err := db.sql.QueryRow(`SELECT edited FROM todo_state WHERE session_id=?`, sessionID).Scan(&edited)
	if err != nil {
		return false, nil
	}
	return edited == 1, nil
}
