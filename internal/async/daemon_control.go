//go:build unix

package async

import (
	"fmt"
	"jin/internal/store"
	"syscall"
	"time"
)

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
	note := "stopped by the agent"
	if req.By == StoppedByUser {
		note = "stopped manually by the user"
	}
	won, err := d.db.FinishAsyncTaskWithEvent(rec.ID, store.AsyncStopped, 143,
		rec.SessionID, rec.Path, ResultText(rec.ID, store.AsyncStopped, -1, note))
	if err != nil {
		return err
	}
	if !won {
		return fmt.Errorf("task %s is not running", req.ID)
	}
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
