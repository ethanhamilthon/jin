package store

import "database/sql"

// Flag reports whether a one-time step was recorded as done.
func (db *DB) Flag(name string) (bool, error) {
	var value string
	err := db.sql.QueryRow(`SELECT value FROM settings WHERE key = ?`, name).Scan(&value)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return value == "1", err
}

// SetFlag records a one-time step as done.
func (db *DB) SetFlag(name string) error {
	return db.setSettings(map[string]string{name: "1"})
}
