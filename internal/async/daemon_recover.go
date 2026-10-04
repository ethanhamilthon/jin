//go:build unix

package async

import (
	"hash/fnv"
	"jin/internal/store"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// recover settles tasks a previous daemon left behind: their supervisor is
// gone, so nobody would announce them. A dead process is marked failed; a
// live group that still belongs to the task is ended.
func (d *daemon) recover() {
	tasks, err := d.db.RunningAsyncTasks("")
	if err != nil {
		return
	}
	for _, t := range tasks {
		if t.ExitPath != "" {
			d.watch(t)
			continue
		}
		alive := t.PID > 0 && syscall.Kill(t.PID, 0) == nil && sameProcess(t)
		note := "the async daemon restarted and lost this task"
		if alive {
			killGroup(t.PGID)
			note += "; its process was stopped"
		}
		_, _ = d.db.FinishAsyncTaskWithEvent(t.ID, store.AsyncFailed, 1, t.SessionID, t.Path, ResultText(t.ID, store.AsyncFailed, 1, note))
	}
}

// sameProcess guards against a reused pid: the start time recorded with the
// task must still match. An unknown stamp (0) counts as a match.
func sameProcess(t store.AsyncTask) bool {
	if t.ProcStartedAt == 0 {
		return true
	}
	now := procStamp(t.PID)
	return now == 0 || now == t.ProcStartedAt
}

// procStamp identifies a process by its start time. It is 0 when unknown.
func procStamp(pid int) int64 {
	out, err := exec.Command("ps", "-o", "lstart=", "-p", strconv.Itoa(pid)).Output()
	if err != nil || len(strings.TrimSpace(string(out))) == 0 {
		return 0
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.TrimSpace(string(out))))
	return int64(h.Sum64() >> 1)
}

// startDetached starts `jin daemon` in its own session with no terminal, so
// it survives the shell, the TUI and the terminal that asked for it.
func startDetached(exe string) error {
	cmd := exec.Command(exe, "daemon")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
