package store

import (
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
