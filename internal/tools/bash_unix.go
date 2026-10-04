//go:build unix

package tools

import (
	"os/exec"
	"sync"
	"syscall"
	"time"
)

var (
	groupsMu sync.Mutex
	groups   = map[int]struct{}{}
)

// setSession runs the command in its own session and process group (pgid
// equals pid), detached from the terminal of jin. Ending the group ends
// everything the command started, such as a sub-agent, and the async daemon
// can take the group over.
func setSession(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
}

// killProcessGroup sends SIGTERM to the whole group, and SIGKILL a second
// later when something is still alive.
func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process == nil {
		return
	}
	pgid := cmd.Process.Pid
	_ = syscall.Kill(-pgid, syscall.SIGTERM)
	go func() {
		time.Sleep(time.Second)
		if syscall.Kill(-pgid, 0) == nil {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		}
	}()
}

// signalCode is the shell convention for a process ended by a signal.
func signalCode(cmd *exec.Cmd) int {
	if ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return 1
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
