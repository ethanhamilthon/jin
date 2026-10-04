package store

import "database/sql"

func migrateProjects(db *sql.DB) error {
	return retryBusy(func() error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		found, err := hasProjectIDColumn(tx)
		if err != nil {
			return err
		}
		if !found {
			if _, err := tx.Exec(`ALTER TABLE sessions ADD COLUMN project_id TEXT REFERENCES projects(id)`); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(`CREATE INDEX IF NOT EXISTS sessions_project ON sessions(project_id, updated_at)`); err != nil {
			return err
		}
		var done string
		err = tx.QueryRow(`SELECT name FROM store_migrations WHERE name = 'projects_v1'`).Scan(&done)
		if err == nil {
			return tx.Commit()
		}
		if err != sql.ErrNoRows {
			return err
		}
		if err := migrateProjectRows(tx); err != nil {
			return err
		}
		if _, err := tx.Exec(`INSERT INTO store_migrations(name) VALUES ('projects_v1')`); err != nil {
			return err
		}
		return tx.Commit()
	})
}

func hasProjectIDColumn(tx *sql.Tx) (bool, error) {
	rows, err := tx.Query(`PRAGMA table_info(sessions)`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var cid, notnull, pk int
		var name, kind string
		var defaultValue sql.NullString
		if err := rows.Scan(&cid, &name, &kind, &notnull, &defaultValue, &pk); err != nil {
			return false, err
		}
		if name == "project_id" {
			return true, nil
		}
	}
	return false, rows.Err()
}
