//go:build unix

package dyn

import (
	"os/exec"
	"syscall"
)

// setGroup gives the command its own process group, so that ending it ends
// everything it started.
func setGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

func killGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
}
