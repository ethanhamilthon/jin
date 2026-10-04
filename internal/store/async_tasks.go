package store

import (
	"database/sql"
	"errors"
	"time"
)

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
		canonical, err := canonicalStoredPath(path)
		if err != nil {
			return nil, err
		}
		query += ` AND (path = ? OR session_id IN (SELECT id FROM sessions WHERE project_id IN (SELECT id FROM projects WHERE path = ?)))`
		args = append(args, path, canonical)
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
