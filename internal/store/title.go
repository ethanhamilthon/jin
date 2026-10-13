package store

// SetTitle names a session that still holds its one title claim and
// reports whether the title was applied. Applying spends the claim.
func (db *DB) SetTitle(id, title string) (bool, error) {
	applied := false
	err := retryBusy(func() error {
		tx, err := db.sql.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		spent, err := tx.Exec(`DELETE FROM title_claims WHERE session_id = ?`, id)
		if err != nil {
			return err
		}
		if n, _ := spent.RowsAffected(); n != 1 {
			applied = false
			return tx.Commit()
		}
		result, err := tx.Exec(`UPDATE sessions SET title = ? WHERE id = ?`, title, id)
		if err != nil {
			return err
		}
		n, _ := result.RowsAffected()
		applied = n == 1
		return tx.Commit()
	})
	return applied, err
}
