//go:build !unix

package tasks

import "os/exec"

func setGroup(*exec.Cmd) {}

func signalCode(*exec.ExitError) int { return 1 }

func terminate(cmd *exec.Cmd, _ <-chan struct{}) { killNow(cmd) }

func killNow(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
