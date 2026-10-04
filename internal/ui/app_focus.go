package ui

import (
	"jin/internal/pricing"
)

func (a *app) setPricing(table pricing.Table) {
	a.pricing = table
	for _, s := range a.sessions {
		s.pricing = table
	}
}

// dropBlank stops the focused session when nothing was ever sent to it, so
// switching away leaves no idle backend behind.
func (a *app) dropBlank() {
	if s := a.active; s != nil && !s.persisted && len(s.pending) == 0 && !s.working {
		s.stop()
		delete(a.sessions, s.id)
	}
}

func (a *app) focus(s *chatSession) {
	if a.active != s {
		a.cancelVoice()
		a.dropBlank()
	}
	a.active = s
	s.unread = false
	if s.persisted {
		_ = a.store.SetUnread(s.id, false)
	}
}
