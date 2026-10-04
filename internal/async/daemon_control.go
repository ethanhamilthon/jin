//go:build unix

package async

import (
	"fmt"
	"io"
	"jin/internal/store"
	"syscall"
	"time"
)

func (d *daemon) input(req request) error {
	d.mu.Lock()
	t := d.tasks[req.ID]
	d.mu.Unlock()
	if t != nil && t.stdin == nil {
		return fmt.Errorf("task %s has no stdin you can write to; only tasks started with `jin async run --stdin` have one", req.ID)
	}
	if t == nil {
		return fmt.Errorf("task %s is not running", req.ID)
	}
	text := req.Text
	if !req.NoNewline {
		text += "\n"
	}
	_, err := io.WriteString(t.stdin, text)
	return err
}

// stop closes the task first, so the wait goroutine stays quiet, then
// announces it and ends the process group: SIGTERM, and SIGKILL after a
// grace period.
func (d *daemon) stop(req request) error {
	rec, found, err := d.db.AsyncTask(req.ID)
	if err != nil {
		return err
	}
	if !found || rec.Status != store.AsyncRunning {
		return fmt.Errorf("task %s is not running", req.ID)
	}
	won, err := d.db.FinishAsyncTask(rec.ID, store.AsyncStopped, 143)
	if err != nil {
		return err
	}
	if !won {
		return fmt.Errorf("task %s is not running", req.ID)
	}
	note := "stopped by the agent"
	if req.By == StoppedByUser {
		note = "stopped manually by the user"
	}
	_ = d.db.AddAsyncEvent(rec.SessionID, rec.Path, ResultText(rec.ID, store.AsyncStopped, -1, note))
	killGroup(rec.PGID)
	return nil
}

// killGroup sends SIGTERM to a process group and SIGKILL when it is still
// there after the grace period.
func killGroup(pgid int) {
	if pgid <= 1 {
		return
	}
	_ = syscall.Kill(-pgid, syscall.SIGTERM)
	go func() {
		time.Sleep(killGrace)
		if syscall.Kill(-pgid, 0) == nil {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		}
	}()
}
