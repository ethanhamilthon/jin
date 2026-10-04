package store

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
)

// ErrSessionBusy says that another live jin process owns a session.
type ErrSessionBusy struct{ PID int }

func (e ErrSessionBusy) Error() string {
	return fmt.Sprintf("session is busy: jin process %d is using it", e.PID)
}

// claimSession makes this process the owner of a session, unless another
// live process owns it. An immediate transaction makes the check and the
// claim one step for all jin processes.
func (db *DB) claimSession(id string) error {
	return retryBusy(func() error {
		tx, err := db.sql.Begin()
		if err != nil {
			return err
		}
		defer tx.Rollback()
		var pid int
		err = tx.QueryRow(`SELECT pid FROM running_sessions WHERE session_id = ?`, id).Scan(&pid)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		if err == nil && pid != os.Getpid() && processAlive(pid) {
			return ErrSessionBusy{PID: pid}
		}
		if _, err := tx.Exec(`INSERT OR REPLACE INTO running_sessions(session_id, pid) VALUES (?, ?)`, id, os.Getpid()); err != nil {
			return err
		}
		return tx.Commit()
	})
}

// SessionOwner tells which process owns a session and whether it is alive.
// pid is 0 when nobody owns it.
func (db *DB) SessionOwner(id string) (pid int, alive bool, err error) {
	err = db.sql.QueryRow(`SELECT pid FROM running_sessions WHERE session_id = ?`, id).Scan(&pid)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return pid, processAlive(pid), nil
}

// busySessions lists the sessions that other live processes own.
func (db *DB) busySessions(self int) (map[string]bool, error) {
	rows, err := db.sql.Query(`SELECT session_id, pid FROM running_sessions WHERE pid != ?`, self)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	busy := map[string]bool{}
	for rows.Next() {
		var id string
		var pid int
		if err := rows.Scan(&id, &pid); err != nil {
			return nil, err
		}
		if processAlive(pid) {
			busy[id] = true
		}
	}
	return busy, rows.Err()
}
