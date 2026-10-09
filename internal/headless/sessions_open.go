package headless

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"jin/internal/pricing"
	"jin/internal/session"
	"jin/internal/store"
	"jin/internal/tasks"
)

// managed is one saved session, opened with the same manager that jin web
// uses, so each action does what the web page does.
type managed struct {
	m       *session.Manager
	id      string
	handoff chan string
}

// openManaged finds the session by id or unique prefix and waits until it is
// ready. Close ends the manager.
func openManaged(ctx context.Context, db *store.DB, arg string) (*managed, func(), error) {
	rec, found, err := db.GetSession(arg)
	if err == nil && !found {
		rec, found, err = db.SessionByPrefix(arg)
	}
	if err != nil {
		return nil, nil, err
	}
	if !found {
		return nil, nil, fmt.Errorf("no session with id %q", arg)
	}
	s := &managed{id: rec.ID, handoff: make(chan string, 1)}
	var once sync.Once
	s.m = session.NewManager(ctx, db, "", func(ev session.Event) {
		if ev.Type == "handoff" {
			once.Do(func() { s.handoff <- ev.Text })
		}
	})
	prices := make(chan pricing.Table, 1)
	go func() { prices <- loadPricing(ctx) }()
	s.m.Start(prices, tasks.Shared().Events())
	closeAll := s.m.Shutdown
	snap, err := s.m.Open(rec.ID)
	if err == nil && snap.State.ReadOnly != 0 {
		err = fmt.Errorf("session %s is running in another jin process (pid %d)", rec.ID, snap.State.ReadOnly)
	}
	if err == nil {
		err = s.wait(ctx, func(st session.State) bool { return st.Ready })
	}
	if err != nil {
		closeAll()
		return nil, nil, err
	}
	return s, closeAll, nil
}

// wait polls the state of the session until done says so.
func (s *managed) wait(ctx context.Context, done func(session.State) bool) error {
	for {
		for _, st := range s.m.Live() {
			if st.ID == s.id && done(st) {
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return errors.New("interrupted")
		case <-time.After(20 * time.Millisecond):
		}
	}
}

func (s *managed) idle(ctx context.Context) error {
	return s.wait(ctx, func(st session.State) bool { return !st.Busy })
}
