//go:build !unix

package cliproxy

import (
	"os/exec"
	"strconv"
)

func itoa(port int) string { return strconv.Itoa(port) }
func isolate(*exec.Cmd)    {}
func terminate(cmd *exec.Cmd) {
	if cmd.Process != nil {
		_ = cmd.Process.Kill()
	}
}
func kill(cmd *exec.Cmd) { terminate(cmd) }
