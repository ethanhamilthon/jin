package ui

import (
	"context"
	"time"
)

func (a *app) shutdownSessions() {
	if a.backend != nil {
		if a.cancel != nil {
			a.cancel()
		}
		return
	}
	a.markInterruptedUnread()
	if a.cancel != nil {
		a.cancel()
	}
	for _, s := range a.sessions {
		if s.render != nil {
			s.render.cancel()
		}
		if s.bash != nil && s.bash.cancel != nil {
			s.bash.cancel()
		}
		if s.stop != nil {
			s.stop()
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, s := range a.sessions {
		stopped := waitSessionStop(ctx, s.backendDone)
		if s.bash != nil {
			stopped = waitSessionStop(ctx, s.bash.done) && stopped
		}
		if stopped && s.persisted && s.store != nil && s.readOnlyPID == 0 {
			_ = s.store.SetRunning(s.id, false)
		}
	}
}

func waitSessionStop(ctx context.Context, done <-chan struct{}) bool {
	if done == nil {
		return true
	}
	select {
	case <-done:
		return true
	case <-ctx.Done():
		return false
	}
}
