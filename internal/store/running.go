package store

import (
	"os"
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
// sessions and records them as interrupted, so the daemon can say in the
// session that the work was lost. Requests of other jin processes that are
// still alive are left alone.
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
		if _, err := db.sql.Exec(`INSERT OR IGNORE INTO interrupted_sessions(session_id) VALUES (?)`, r.id); err != nil {
			return err
		}
	}
	return rows.Err()
}

// InterruptedSessions lists the sessions whose work a dead jin process lost
// and that no client has shown as recovered yet.
func (db *DB) InterruptedSessions() ([]string, error) {
	rows, err := db.sql.Query(`SELECT session_id FROM interrupted_sessions`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// MarkRecovered clears the interrupted mark of one session after a client
// showed it.
func (db *DB) MarkRecovered(id string) error {
	_, err := db.sql.Exec(`DELETE FROM interrupted_sessions WHERE session_id=?`, id)
	return err
}

// SetDeadForTest copies the requests of this process into rows of a process
// that is gone, so a test can exercise the recovery path.
func (db *DB) SetDeadForTest() (int, error) {
	result, err := db.sql.Exec(`INSERT OR REPLACE INTO running_sessions(session_id, pid)
		SELECT session_id, ? FROM running_sessions WHERE pid = ?`, 999999, os.Getpid())
	if err != nil {
		return 0, err
	}
	rows, err := result.RowsAffected()
	return int(rows), err
}
