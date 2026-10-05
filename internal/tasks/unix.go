//go:build unix

package tasks

import (
	"os/exec"
	"syscall"
	"time"
)

func setGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func signalCode(exit *exec.ExitError) int {
	if ws, ok := exit.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return 1
}

// terminate sends SIGTERM to the group of cmd and SIGKILL after 3 s when the
// task has not ended by then.
func terminate(cmd *exec.Cmd, ended <-chan struct{}) {
	if cmd.Process == nil {
		return
	}
	pgid := cmd.Process.Pid
	_ = syscall.Kill(-pgid, syscall.SIGTERM)
	go func() {
		select {
		case <-ended:
		case <-time.After(3 * time.Second):
		}
		_ = syscall.Kill(-pgid, syscall.SIGKILL)
	}()
}

func killNow(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
