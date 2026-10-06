package session

import (
	"context"
	"time"
)

// Shutdown stops every session, marks the ones still answering as unread
// and releases the ones that stopped.
func (m *Manager) Shutdown() {
	m.mu.Lock()
	var stopping []*Session
	for _, s := range m.sessions {
		if (s.working || s.inflight > 0 || len(s.pending) > 0) && s.persisted && s.readOnlyPID == 0 {
			_ = m.db.SetUnread(s.id, true)
		}
		if s.render != nil {
			s.render.cancel()
		}
		if s.shell != nil {
			s.shell()
		}
		s.stop()
		stopping = append(stopping, s)
	}
	m.mu.Unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	for _, s := range stopping {
		stopped := true
		if s.done != nil {
			select {
			case <-s.done:
			case <-ctx.Done():
				stopped = false
			}
		}
		if stopped && s.persisted && s.readOnlyPID == 0 {
			_ = m.db.SetRunning(s.id, false)
		}
	}
}

// Working reports whether any session or shell command is running.
func (m *Manager) Working() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, s := range m.sessions {
		if s.busy() {
			return true
		}
	}
	return false
}
