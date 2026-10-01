// Package editor describes the external text editors jin can hand a file to.
package editor

import (
	"os"
	"os/exec"
	"slices"
)

var Choices = []string{"nano", "vim", "hx"}

func Valid(name string) bool {
	return slices.Contains(Choices, name)
}

func Installed(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Command wires the editor to the real terminal; the caller must suspend the
// TUI around Run.
func Command(name, path string) *exec.Cmd {
	cmd := exec.Command(name, path)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	return cmd
}
