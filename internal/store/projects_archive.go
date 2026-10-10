package store

// SetProjectArchived hides a project from the session picker of jin web, or shows it
// again. The project, its sessions and its settings stay as they are.
func (db *DB) SetProjectArchived(path string, archived bool) error {
	path, err := normalizedProjectPath(path)
	if err != nil {
		return err
	}
	value := 0
	if archived {
		value = 1
	}
	return retryBusy(func() error {
		_, err := db.sql.Exec(`UPDATE projects SET archived = ? WHERE path = ?`, value, path)
		return err
	})
}

func (db *DB) archivedPaths() (map[string]bool, error) {
	rows, err := db.sql.Query(`SELECT path FROM projects WHERE archived = 1`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return nil, err
		}
		out[path] = true
	}
	return out, rows.Err()
}
