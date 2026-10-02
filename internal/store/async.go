package store

import (
	"database/sql"
	"errors"
	"time"
)

// Async task states.
const (
	AsyncRunning = "running"
	AsyncDone    = "done"
	AsyncFailed  = "failed"
	AsyncStopped = "stopped"
	// AsyncEnded is a task whose process ended while nobody could record its
	// exit code, for example because jin was closed.
	AsyncEnded = "ended"
)

const asyncSchema = `
CREATE TABLE IF NOT EXISTS async_tasks (
	id TEXT PRIMARY KEY,
	session_id TEXT NOT NULL DEFAULT '',
	path TEXT NOT NULL,
	command TEXT NOT NULL,
	pid INTEGER NOT NULL DEFAULT 0,
	pgid INTEGER NOT NULL DEFAULT 0,
	proc_started_at INTEGER NOT NULL DEFAULT 0,
	status TEXT NOT NULL,
	exit_code INTEGER NOT NULL DEFAULT 0,
	log_path TEXT NOT NULL DEFAULT '',
	started_at INTEGER NOT NULL,
	finished_at INTEGER NOT NULL DEFAULT 0,
	exit_path TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS async_tasks_status ON async_tasks(status);
CREATE TABLE IF NOT EXISTS async_events (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	session_id TEXT NOT NULL,
	path TEXT NOT NULL,
	text TEXT NOT NULL,
	claimed_by INTEGER NOT NULL DEFAULT 0,
	created_at INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS async_events_path ON async_events(path, claimed_by);
`

// AsyncTask is one background command run by the async daemon.
type AsyncTask struct {
	ID            string
	SessionID     string
	Path          string
	Command       string
	PID           int
	PGID          int
	ProcStartedAt int64
	Status        string
	ExitCode      int
	LogPath       string
	// ExitPath is set for a task that moved to the background from the bash
	// tool: its process is not a child of the daemon, and jin writes the exit
	// code to this file when the process ends.
	ExitPath   string
	StartedAt  time.Time
	FinishedAt time.Time
}

// AsyncEvent is a message for the agent of a session: a finished task or a
// note sent with `jin async run "echo ..."`.
type AsyncEvent struct {
	ID        int64
	SessionID string
	Path      string
	Text      string
}

const asyncTaskColumns = `id, session_id, path, command, pid, pgid, proc_started_at, status, exit_code, log_path, started_at, finished_at, exit_path`

func scanAsyncTask(row interface{ Scan(...any) error }) (AsyncTask, error) {
	var t AsyncTask
	var started, finished int64
	err := row.Scan(&t.ID, &t.SessionID, &t.Path, &t.Command, &t.PID, &t.PGID, &t.ProcStartedAt,
		&t.Status, &t.ExitCode, &t.LogPath, &started, &finished, &t.ExitPath)
	t.StartedAt = time.Unix(started, 0)
	if finished > 0 {
		t.FinishedAt = time.Unix(finished, 0)
	}
	return t, err
}

// AddAsyncTask records a task that has just started.
func (db *DB) AddAsyncTask(t AsyncTask) error {
	_, err := db.sql.Exec(`INSERT INTO async_tasks (`+asyncTaskColumns+`) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, ?, 0, ?)`,
		t.ID, t.SessionID, t.Path, t.Command, t.PID, t.PGID, t.ProcStartedAt, AsyncRunning, t.LogPath, t.StartedAt.Unix(), t.ExitPath)
	return err
}

// FinishAsyncTask closes a running task. A task that is no longer running
// (for example one already marked stopped) keeps its state.
func (db *DB) FinishAsyncTask(id, status string, exitCode int) (bool, error) {
	res, err := db.sql.Exec(`UPDATE async_tasks SET status = ?, exit_code = ?, finished_at = ? WHERE id = ? AND status = ?`,
		status, exitCode, time.Now().Unix(), id, AsyncRunning)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

func (db *DB) AsyncTask(id string) (AsyncTask, bool, error) {
	t, err := scanAsyncTask(db.sql.QueryRow(`SELECT `+asyncTaskColumns+` FROM async_tasks WHERE id = ?`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return AsyncTask{}, false, nil
	}
	return t, err == nil, err
}

// RunningAsyncTasks lists running tasks, oldest first. An empty path means
// every directory.
func (db *DB) RunningAsyncTasks(path string) ([]AsyncTask, error) {
	query := `SELECT ` + asyncTaskColumns + ` FROM async_tasks WHERE status = ?`
	args := []any{AsyncRunning}
	if path != "" {
		query += ` AND path = ?`
		args = append(args, path)
	}
	rows, err := db.sql.Query(query+` ORDER BY started_at, id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []AsyncTask
	for rows.Next() {
		t, err := scanAsyncTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

// AddAsyncEvent queues a message for a session.
func (db *DB) AddAsyncEvent(sessionID, path, text string) error {
	_, err := db.sql.Exec(`INSERT INTO async_events (session_id, path, text, created_at) VALUES (?, ?, ?, ?)`,
		sessionID, path, text, time.Now().Unix())
	return err
}

// ClaimAsyncEvents takes the unclaimed events of a directory for this
// process. They stay in the table until AckAsyncEvents, so a process that
// dies before delivery does not lose them (see ReleaseDeadAsyncClaims).
func (db *DB) ClaimAsyncEvents(path string, pid int) ([]AsyncEvent, error) {
	tx, err := db.sql.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	rows, err := tx.Query(`SELECT id, session_id, path, text FROM async_events WHERE path = ? AND claimed_by = 0 ORDER BY id`, path)
	if err != nil {
		return nil, err
	}
	var events []AsyncEvent
	for rows.Next() {
		var e AsyncEvent
		if err := rows.Scan(&e.ID, &e.SessionID, &e.Path, &e.Text); err != nil {
			rows.Close()
			return nil, err
		}
		events = append(events, e)
	}
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}
	if _, err := tx.Exec(`UPDATE async_events SET claimed_by = ? WHERE path = ? AND claimed_by = 0 AND id <= ?`,
		pid, path, events[len(events)-1].ID); err != nil {
		return nil, err
	}
	return events, tx.Commit()
}

// AckAsyncEvents removes events that were handed to an agent.
func (db *DB) AckAsyncEvents(ids ...int64) error {
	for _, id := range ids {
		if _, err := db.sql.Exec(`DELETE FROM async_events WHERE id = ?`, id); err != nil {
			return err
		}
	}
	return nil
}

// ReleaseDeadAsyncClaims frees events claimed by processes that are gone.
func (db *DB) ReleaseDeadAsyncClaims() error {
	rows, err := db.sql.Query(`SELECT DISTINCT claimed_by FROM async_events WHERE claimed_by != 0`)
	if err != nil {
		return err
	}
	var dead []int
	for rows.Next() {
		var pid int
		if err := rows.Scan(&pid); err != nil {
			rows.Close()
			return err
		}
		if !processAlive(pid) {
			dead = append(dead, pid)
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	for _, pid := range dead {
		if _, err := db.sql.Exec(`UPDATE async_events SET claimed_by = 0 WHERE claimed_by = ?`, pid); err != nil {
			return err
		}
	}
	return nil
}
