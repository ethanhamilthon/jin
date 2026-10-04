package headless

import (
	"errors"
	"fmt"

	"jin/internal/core"
	"jin/internal/store"
)

// claim makes this process the owner of the session before anything is
// written to it; a session another live jin process uses is refused.
func (r *runState) claim() error {
	err := r.db.SetRunning(r.id, true)
	var busy store.ErrSessionBusy
	if errors.As(err, &busy) {
		return fmt.Errorf("session %s is in use by process %d", r.id, busy.PID)
	}
	return nil
}

// persist creates or touches the session and closes tool calls an earlier
// run left without an answer.
func (r *runState) persist(dir, prompt string) error {
	r.id = r.record.ID
	if !r.save {
		r.history = append(r.history, core.InterruptedToolMessages(r.history)...)
		return nil
	}
	if r.id == "" {
		r.id = newID()
	}
	if err := r.claim(); err != nil {
		return err
	}
	title := r.record.Title
	if title == "" {
		title = truncate(firstLine(prompt), maxTitle)
	}
	if err := r.db.TouchProvider(r.id, dir, r.request.Model, r.request.Effort, title, r.provider); err != nil {
		return err
	}
	for _, msg := range core.InterruptedToolMessages(r.history) {
		if err := r.db.AppendMessage(r.id, msg); err != nil {
			return err
		}
		r.history = append(r.history, msg)
	}
	r.close = func() { _ = r.db.SetRunning(r.id, false) }
	return nil
}
