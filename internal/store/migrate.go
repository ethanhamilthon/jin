package store

import (
	"database/sql"
	"strings"
)

// addedColumns are applied to databases created before the column existed.
// CREATE TABLE already includes them for fresh installs, so "duplicate
// column" just means there is nothing to do.
var addedColumns = []string{
	`cost REAL NOT NULL DEFAULT 0`,
}

func migrate(db *sql.DB) error {
	for _, column := range addedColumns {
		_, err := db.Exec(`ALTER TABLE sessions ADD COLUMN ` + column)
		if err != nil && !strings.Contains(err.Error(), "duplicate column name") {
			return err
		}
	}
	return nil
}
