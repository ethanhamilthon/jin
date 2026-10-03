//go:build unix

package async

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"jin/internal/store"
	"jin/internal/tasklog"
)

const (
	watchEvery = time.Second
	// exitGrace is how long the daemon waits for the exit file after the
	// process is gone; jin writes it a moment after the process ends.
	exitGrace = 2 * time.Second
	// A log above maxLog keeps its last keepLog bytes.
	maxLog  = 64 << 20
	keepLog = 16 << 20
)

// adopt takes over a process that the bash tool started. The daemon is not
// its parent, so it cannot wait for it: watch polls the process and reads the
// exit code from the file jin writes.
func (d *daemon) adopt(req request) (string, error) {
	if req.PID <= 1 || req.PGID <= 1 {
		return "", fmt.Errorf("invalid process")
	}
	if !tasklog.Within(req.Log) || !tasklog.Within(req.Exit) {
		return "", fmt.Errorf("task files must be inside the jin async folder")
	}
	if req.Session == "" {
		return "", fmt.Errorf("session id is required")
	}
	if syscall.Kill(req.PID, 0) != nil {
		return "", fmt.Errorf("the process is gone")
	}
	path := req.Cwd
	if rec, ok, err := d.db.GetSession(req.Session); err == nil && ok {
		path = rec.Path
	}
	id := strings.TrimSuffix(baseName(req.Log), ".log")
	task := store.AsyncTask{
		ID: id, SessionID: req.Session, Path: path, Command: req.Command,
		PID: req.PID, PGID: req.PGID, ProcStartedAt: procStamp(req.PID),
		LogPath: req.Log, ExitPath: req.Exit, StartedAt: time.Now(),
	}
	if err := d.db.AddAsyncTask(task); err != nil {
		return "", err
	}
	d.watch(task)
	return id, nil
}

func baseName(path string) string {
	if i := strings.LastIndexByte(path, os.PathSeparator); i >= 0 {
		return path[i+1:]
	}
	return path
}

// watch follows an adopted task until its process ends.
func (d *daemon) watch(t store.AsyncTask) {
	d.mu.Lock()
	d.tasks[t.ID] = &task{}
	d.mu.Unlock()
	go func() {
		ticker := time.NewTicker(watchEvery)
		defer ticker.Stop()
		for range ticker.C {
			_ = tasklog.Trim(t.LogPath, maxLog, keepLog)
			if rec, found, err := d.db.AsyncTask(t.ID); err == nil && (!found || rec.Status != store.AsyncRunning) {
				// Stopped by hand: stop already closed and announced it.
				d.forget(t.ID)
				return
			}
			if syscall.Kill(t.PID, 0) == nil && sameProcess(t) {
				continue
			}
			d.endAdopted(t)
			return
		}
	}()
}

func (d *daemon) forget(id string) {
	d.mu.Lock()
	delete(d.tasks, id)
	d.active = time.Now()
	d.mu.Unlock()
}
