//go:build unix

package async

import (
	"fmt"
	"jin/internal/store"
	"os/exec"
	"syscall"
	"time"
)

// wait ends a task when its process does. A task that was stopped has
// already been closed and announced by stop, so nothing more is reported.
func (d *daemon) wait(id string, cmd *exec.Cmd, session, path, logPath string) {
	err := cmd.Wait()
	code := exitCode(cmd, err)
	d.mu.Lock()
	delete(d.tasks, id)
	d.active = time.Now()
	d.mu.Unlock()
	status := store.AsyncDone
	if code != 0 {
		status = store.AsyncFailed
	}
	won, dbErr := d.db.FinishAsyncTask(id, status, code)
	if dbErr != nil || !won {
		return
	}
	output, truncated, _ := Tail(logPath, eventTail)
	if truncated {
		output = fmt.Sprintf("[output cut: last %d characters; run `jin async check --id %s` for more]\n%s", eventTail, id, output)
	}
	_ = d.db.AddAsyncEvent(session, path, ResultText(id, status, code, output))
}

func exitCode(cmd *exec.Cmd, err error) int {
	if cmd.ProcessState == nil {
		if err != nil {
			return 1
		}
		return 0
	}
	if code := cmd.ProcessState.ExitCode(); code >= 0 {
		return code
	}
	if ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return 1
}
