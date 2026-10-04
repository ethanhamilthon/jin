//go:build !unix

package tools

import "os/exec"

func setSession(*exec.Cmd) {}

func killProcessGroup(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}

func signalCode(*exec.Cmd) int { return 1 }

func rememberGroup(*exec.Cmd) {}

// KillBackground is a no-op on this platform.
func KillBackground() {}
