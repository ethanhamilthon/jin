//go:build unix

package tools

import (
	"os/exec"
	"sync"
	"syscall"
)

var (
	groupsMu sync.Mutex
	groups   = map[int]struct{}{}
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

// rememberGroup records the group of a finished command when background
// processes (for example async sub-agents) still run in it.
func rememberGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	pgid := cmd.Process.Pid
	if syscall.Kill(-pgid, 0) != nil {
		return
	}
	groupsMu.Lock()
	groups[pgid] = struct{}{}
	groupsMu.Unlock()
}

// KillBackground stops every background process left by finished bash calls.
// Call it when jin exits.
func KillBackground() {
	groupsMu.Lock()
	defer groupsMu.Unlock()
	for pgid := range groups {
		_ = syscall.Kill(-pgid, syscall.SIGTERM)
		delete(groups, pgid)
	}
}
