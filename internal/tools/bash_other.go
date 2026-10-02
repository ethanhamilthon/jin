//go:build !unix

package tools

import "os/exec"

func killGroup(*exec.Cmd) {}
