//go:build !unix

package tools

import "os/exec"

func killGroup(*exec.Cmd) {}

func rememberGroup(*exec.Cmd) {}

// KillBackground is a no-op on this platform.
func KillBackground() {}
