package web

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
)

// keepAwake stops macOS from sleeping while this process lives. Other
// systems are left alone.
func keepAwake(out io.Writer) {
	if runtime.GOOS != "darwin" {
		fmt.Fprintln(out, "Keep this computer awake yourself; jin only does it on macOS.")
		return
	}
	if err := exec.Command("caffeinate", "-i", "-w", strconv.Itoa(os.Getpid())).Start(); err != nil {
		fmt.Fprintln(out, "Could not run caffeinate; the computer may go to sleep.")
	}
}
