//go:build unix

package tools

import (
	"os/exec"
	"syscall"
)

// killGroup runs the command in its own process group and, on timeout or
// interrupt, sends SIGTERM to the whole group. Without it only bash would die
// and anything it started, such as a sub-agent, would keep running.
func killGroup(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
	}
}
