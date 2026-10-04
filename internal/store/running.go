package store

import (
	"os"
	"syscall"
)

// SetRunning claims a session for an in-flight request of this process, so
// a killed process can leave an unread indicator on the next launch. It
// fails with ErrSessionBusy while another live process owns the session.
// Only the owner releases it.
func (db *DB) SetRunning(id string, running bool) error {
	if running {
		return db.claimSession(id)
	}
	_, err := db.sql.Exec(`DELETE FROM running_sessions WHERE session_id=? AND pid=?`, id, os.Getpid())
	return err
}

// RecoverInterrupted turns the requests of dead processes into unread
// sessions. Requests of other jin processes that are still alive are left
// alone.
func (db *DB) RecoverInterrupted() error {
	rows, err := db.sql.Query(`SELECT session_id, pid FROM running_sessions`)
	if err != nil {
		return err
	}
	type owned struct {
		id  string
		pid int
	}
	var dead []owned
	for rows.Next() {
		var r owned
		if err := rows.Scan(&r.id, &r.pid); err != nil {
			rows.Close()
			return err
		}
		if !processAlive(r.pid) {
			dead = append(dead, r)
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, r := range dead {
		if _, err := db.sql.Exec(`INSERT OR IGNORE INTO unread_sessions(session_id) VALUES (?)`, r.id); err != nil {
			return err
		}
		if _, err := db.sql.Exec(`DELETE FROM running_sessions WHERE session_id=? AND pid=?`, r.id, r.pid); err != nil {
			return err
		}
	}
	return rows.Err()
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	return err == nil || err == syscall.EPERM
}
