package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"jin/internal/daemon"
	"jin/internal/update"
)

// daemonControl lets jin update stop the running daemon before it replaces the
// binary the daemon runs and start a fresh one afterwards. Client and daemon
// share one executable in this release, so a daemon left running would make
// every later command fail with a version mismatch.
func daemonControl(root, version string) update.Control {
	return update.Control{
		Stop: func(ctx context.Context, force bool, out io.Writer) error {
			return stopDaemon(ctx, root, version, force, out)
		},
		Start: func(ctx context.Context, out io.Writer) error {
			_, err := daemon.Ensure(ctx, root, version)
			if err != nil && daemonAnswers(ctx, root) {
				// The new binary reports the tag of the release, which an older
				// client does not know: the daemon of the update is running.
				err = nil
			}
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(out, "started the daemon again")
			return err
		},
	}
}

// daemonAnswers reports whether any daemon answers on the socket.
func daemonAnswers(ctx context.Context, root string) bool {
	_, err := daemon.NewClient(root).Status(ctx)
	return err == nil
}

// stopDaemon ends the daemon that runs at root. A daemon of the same version
// stops over its own socket; a daemon of another version refuses such a stop,
// so only --force signals it directly.
func stopDaemon(ctx context.Context, root, version string, force bool, out io.Writer) error {
	client := daemon.NewClient(root)
	status, err := client.Status(ctx)
	if err != nil {
		return nil // no daemon is running
	}
	switch {
	case status.Version == version:
		if err := client.Stop(ctx, version, force); err != nil {
			return busyAdvice(err)
		}
	case !force:
		return fmt.Errorf("the running daemon is %s and this jin is %s; run jin update --force to stop it", status.Version, version)
	default:
		if err := daemon.Kill(status.PID); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(out, "stopped the running daemon"); err != nil {
		return err
	}
	return waitStopped(ctx, root)
}

// busyAdvice adds the update flag to the refusal of a busy daemon.
func busyAdvice(err error) error {
	if !strings.Contains(err.Error(), "busy") {
		return err
	}
	return errors.New(err.Error() + "; or run jin update --force")
}

// waitStopped waits until the socket of the old daemon stops answering, so the
// new daemon does not race with it for the data folder lock.
func waitStopped(ctx context.Context, root string) error {
	client := daemon.NewClient(root)
	for range 100 {
		if _, err := client.Status(ctx); err != nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
	return errors.New("the old daemon still answers; wait a moment and run jin")
}
