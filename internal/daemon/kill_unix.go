//go:build unix

package daemon

import (
	"errors"
	"os"
	"strconv"
	"syscall"
	"time"
)

// Kill ends a daemon of another version, which cannot be asked to stop over
// its control socket: a daemon refuses the stop of a client with another
// version. The signal is the same as an interrupt, so the daemon shuts its
// sessions down cleanly. A process that is already gone is no error.
func Kill(pid int) error {
	if pid <= 0 {
		return errors.New("no daemon process to stop")
	}
	process, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := process.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if err := process.Signal(syscall.Signal(0)); err != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return errors.New("the daemon did not stop; stop it yourself with kill " + strconv.Itoa(pid))
}
