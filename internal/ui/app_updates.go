package ui

import "jin/internal/core"

func (a *app) applyUpdate(id string, update core.Update) {
	s, ok := a.sessions[id]
	if !ok {
		return
	}
	if update.Kind == core.UpdateHandoff {
		a.startHandoffFrom(s, update.Text)
		return
	}
	if update.Kind == core.UpdateTaken {
		if s.inflight > 0 {
			s.inflight--
		}
		a.releaseIdle(s)
		return
	}
	s.showUpdate(update)
	if update.Kind == core.UpdateDone && s.inflight > 0 {
		s.inflight--
	}
	if update.Kind == core.UpdateWorking && s == a.active {
		a.rememberFocus(s)
	}
	if update.Kind == core.UpdateAsk && shouldRing(a.cfg.Sound, !a.blurred && s == a.active) {
		a.ring()
	}
	if update.Kind == core.UpdateDone && update.Final && shouldRing(a.cfg.Sound, !a.blurred) {
		a.ring()
	}
	if update.Kind != core.UpdateDone || !s.persisted {
		return
	}
	if update.Final {
		a.autoTitle(s)
	}
	s.unread = s != a.active
	_ = a.store.SetUnread(id, s.unread)
	a.releaseIdle(s)
}

// releaseIdle gives a session back to other processes once nothing of it
// runs: no request in flight or queued, no shell command.
func (a *app) releaseIdle(s *chatSession) {
	if s.persisted && s.inflight == 0 && len(s.pending) == 0 && (s.bash == nil || !s.bash.running) {
		_ = a.store.SetRunning(s.id, false)
	}
}

// flushPending hands queued prompts to each backend without blocking the UI.
func (a *app) flushPending() {
	for _, s := range a.sessions {
		if !s.ready {
			// The agent is not running yet; task results wait here.
			continue
		}
	deliver:
		for len(s.pending) > 0 {
			select {
			case s.prompts <- s.pending[0]:
				s.inflight++
				if s.pending[0].Interactive {
					s.agent.DetachTools()
				}
				s.pending = s.pending[1:]
			default:
				break deliver
			}
		}
	}
}

func (a *app) anyWorking() bool {
	for _, s := range a.sessions {
		if s.working || s.inflight > 0 || len(s.pending) > 0 || s.render != nil || (s.bash != nil && s.bash.running) {
			return true
		}
	}
	return false
}

// markInterruptedUnread flags sessions still answering at exit so the
// sessions list shows them as unread next time.
func (a *app) markInterruptedUnread() {
	for id, s := range a.sessions {
		if (s.working || s.inflight > 0 || len(s.pending) > 0) && s.persisted && s.readOnlyPID == 0 {
			_ = a.store.SetUnread(id, true)
		}
	}
}
