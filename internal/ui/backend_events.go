package ui

import (
	"jin/internal/core"
	"jin/internal/daemon"
	"jin/internal/session"
)

// receiveBackend applies one event from the daemon. Events for a session this
// client does not show are ignored.
func (a *app) receiveBackend(event session.Event) {
	switch event.Type {
	case "resync":
		a.resyncBackend()
		return
	case "tasks":
		a.receiveBackendTasks()
		return
	case "config":
		a.receiveBackendConfig()
		return
	case "handoff":
		a.receiveHandoff(event)
		return
	}
	s := a.sessions[event.Session]
	if s == nil || event.Seq <= s.remoteSeq {
		return
	}
	s.remoteSeq = event.Seq
	a.applyBackendEvent(s, event)
}

func (a *app) applyBackendEvent(s *chatSession, event session.Event) {
	switch event.Type {
	case "state":
		if event.State != nil {
			a.backendState(s, *event.State)
		}
	case "entry":
		if event.Entry == nil {
			return
		}
		s.remoteEntries = append(s.remoteEntries, *event.Entry)
		s.remoteMap = append(s.remoteMap, len(s.history))
		s.appendEntry(backendEntry(*event.Entry))
	case "delta":
		a.applyBackendDelta(s, event)
	case "update", "entries":
		snap, err := a.backend.Snapshot(a.ctx, s.id)
		if err == nil {
			a.backendSession(snap)
		}
	case "ring":
		if event.Kind == core.UpdateAsk || event.Text == "final" {
			if shouldRing(a.cfg.Sound, !a.blurred) {
				a.ring()
			}
		}
	}
}

// applyBackendDelta streams text into the entry the daemon names. A delta
// never touches the last history row on its own: this client adds rows of its
// own while the agent streams, so only the index mapping is correct.
func (a *app) applyBackendDelta(s *chatSession, event session.Event) {
	if event.Index < 0 || event.Index >= len(s.remoteEntries) {
		return
	}
	index, ok := s.remoteRow(event.Index)
	if !ok || index >= len(s.history) {
		return
	}
	if s.history[index].kind != s.remoteEntries[event.Index].Kind {
		return
	}
	s.remoteEntries[event.Index].Text += event.Text
	s.history[index].text += event.Text
	s.rebuildRows(s.width)
}

func (a *app) resyncBackend() {
	for id := range a.sessions {
		snap, err := a.backend.Snapshot(a.ctx, id)
		if err == nil {
			a.backendSession(snap)
		}
	}
}
// receiveHandoff follows a handoff only in the client that asked for it and
// only while that client shows the session that made the brief.
func (a *app) receiveHandoff(event session.Event) {
	if a.backend == nil || event.Origin == "" || event.Origin != a.backend.ID || event.Text == "" {
		return
	}
	from := a.sessions[event.Session]
	if from == nil || a.active != from {
		return
	}
	snap, err := a.backend.Open(a.ctx, event.Text)
	if err != nil {
		a.backendError(err)
		return
	}
	a.focus(a.backendSession(snap))
	a.sel = nil
}

func (a *app) backendAction(action string) {
	if a.active == nil {
		return
	}
	a.backendError(a.backend.Command(a.ctx, daemon.Command{Action: action, Session: a.active.id}, nil))
}
