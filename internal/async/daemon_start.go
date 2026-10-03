package async

import (
	"errors"
	"os"
	"time"
)

const (
	dialTimeout  = 500 * time.Millisecond
	startTimeout = 2 * time.Second
	callTimeout  = 15 * time.Second
)

// ensureDaemon makes sure a daemon of this version answers. A daemon of
// another version is replaced when it has no running task; otherwise it keeps
// serving until it is idle.
func ensureDaemon(version string) error {
	resp, err := call(request{Op: opHello})
	if err == nil && resp.Version == version {
		return nil
	}
	if err == nil {
		if !resp.Idle {
			return nil
		}
		if _, err := call(request{Op: opShutdown}); err != nil {
			return nil
		}
		waitGone()
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if err := startDetached(exe); err != nil {
		return err
	}
	deadline := time.Now().Add(startTimeout)
	for time.Now().Before(deadline) {
		if _, err := call(request{Op: opHello}); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return errors.New("the async daemon did not start")
}

func waitGone() {
	deadline := time.Now().Add(startTimeout)
	for time.Now().Before(deadline) {
		if _, err := call(request{Op: opHello}); err != nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}
