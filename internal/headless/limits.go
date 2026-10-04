package headless

import (
	"context"
	"errors"
	"os"
	"time"
)

// withLimits makes the context of a whole run: it ends at the timeout (total
// wall time) or on the first signal.
func withLimits(ctx context.Context, timeout time.Duration, signals <-chan os.Signal) (context.Context, func()) {
	ctx, cancelCause := context.WithCancelCause(ctx)
	stopTimer := func() {}
	if timeout > 0 {
		ctx, stopTimer = context.WithTimeoutCause(ctx, timeout, context.DeadlineExceeded)
	}
	go func() {
		select {
		case <-signals:
			cancelCause(errors.New("interrupted"))
		case <-ctx.Done():
		}
	}()
	return ctx, func() { stopTimer(); cancelCause(nil) }
}

func timedOut(ctx context.Context) bool {
	return errors.Is(context.Cause(ctx), context.DeadlineExceeded)
}

// endedBy explains why a run context ended.
func endedBy(ctx context.Context, timeout time.Duration) error {
	if timedOut(ctx) {
		return errors.New("timed out after " + timeout.String())
	}
	return errors.New("interrupted")
}

func exitCode(ctx context.Context) int {
	if ctx.Err() != nil && !timedOut(ctx) {
		return exitInterrupted
	}
	return exitError
}
