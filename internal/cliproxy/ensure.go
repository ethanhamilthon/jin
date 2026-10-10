package cliproxy

import (
	"context"
	"errors"
	"io"
	"net"
	"os"
	"os/exec"
	"time"
)

func ensureBroker(ctx context.Context, root string) error {
	dial := func() bool {
		conn, err := net.DialTimeout("unix", socketPath(root), 100*time.Millisecond)
		if err != nil {
			return false
		}
		conn.Close()
		return true
	}
	if dial() {
		return nil
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "--internal-cliproxy", root)
	cmd.Stdout, cmd.Stderr = io.Discard, io.Discard
	isolate(cmd)
	if err = cmd.Start(); err != nil {
		return errors.New("cannot launch managed proxy supervisor")
	}
	go cmd.Wait()
	timeout := time.NewTimer(25 * time.Second)
	defer timeout.Stop()
	for {
		if dial() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout.C:
			return errors.New("managed proxy supervisor did not start")
		case <-time.After(100 * time.Millisecond):
		}
	}
}
