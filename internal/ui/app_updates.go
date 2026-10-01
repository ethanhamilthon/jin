package ui

import "jin/internal/core"

func (a *app) applyUpdate(id string, update core.Update) {
	s, ok := a.sessions[id]
	if !ok {
		return
	}
	if update.Kind == core.UpdateHandoff {
		a.startHandoff(update.Text)
		return
	}
	s.showUpdate(update)
	if update.Kind == core.UpdateDone && update.Final {
		_ = a.screen.Beep()
	}
	if update.Kind != core.UpdateDone || !s.persisted {
		return
	}
	s.unread = s != a.active
	_ = a.store.SetUnread(id, s.unread)
	if len(s.pending) == 0 {
		_ = a.store.SetRunning(id, false)
	}
}

// flushPending hands queued prompts to each backend without blocking the UI.
func (a *app) flushPending() {
	for _, s := range a.sessions {
	deliver:
		for len(s.pending) > 0 {
			select {
			case s.prompts <- s.pending[0]:
				s.pending = s.pending[1:]
			default:
				break deliver
			}
		}
	}
}

func (a *app) anyWorking() bool {
	for _, s := range a.sessions {
		if s.working {
			return true
		}
	}
	return false
}

// markInterruptedUnread flags sessions still answering at exit so the
// sessions list shows them as unread next time.
func (a *app) markInterruptedUnread() {
	for id, s := range a.sessions {
		if s.working && s.persisted {
			_ = a.store.SetUnread(id, true)
			_ = a.store.SetRunning(id, false)
		}
	}
}
