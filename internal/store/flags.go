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

// Setting reads one setting; a missing key is "".
func (db *DB) Setting(key string) (string, error) {
	var value string
	err := db.sql.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// SetSetting stores one setting.
func (db *DB) SetSetting(key, value string) error {
	return db.setSettings(map[string]string{key: value})
}

// Empty reports whether no session was ever saved: a fresh install.
func (db *DB) Empty() (bool, error) {
	var n int
	err := db.sql.QueryRow(`SELECT COUNT(*) FROM sessions`).Scan(&n)
	return n == 0, err
}
