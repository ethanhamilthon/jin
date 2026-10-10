package ui

import (
	"jin/internal/core"
	"jin/internal/daemon"
	"jin/internal/session"
)

func (a *app) receiveBackend(event session.Event) {
	if event.Type == "resync" {
		for id := range a.sessions {
			snap, err := a.backend.Snapshot(a.ctx, id)
			if err == nil {
				a.backendSession(snap)
			}
		}
		return
	}
	if event.Type == "tasks" {
		a.tasksRunning = map[string]int{}
		for _, task := range a.taskList() {
			if task.Status == "running" {
				a.tasksRunning[task.Owner]++
			}
		}
		return
	}
	if event.Type == "config" {
		if cfg, err := a.store.LoadConfig(); err == nil {
			a.cfg = cfg
		}
		return
	}
	s := a.sessions[event.Session]
	if s == nil || event.Seq <= s.remoteSeq {
		return
	}
	s.remoteSeq = event.Seq
	switch event.Type {
	case "state":
		if event.State != nil {
			a.backendState(s, *event.State)
		}
	case "entry":
		if event.Entry != nil {
			s.remoteEntries = append(s.remoteEntries, *event.Entry)
			if !s.persisted && event.Entry.Kind == core.UpdateUser {
				s.history, s.rows = nil, nil
			}
			s.appendEntry(backendEntry(*event.Entry))
		}
	case "delta":
		if event.Index >= 0 && event.Index < len(s.remoteEntries) {
			s.remoteEntries[event.Index].Text += event.Text
			if len(s.history) > 0 {
				s.history[len(s.history)-1].text += event.Text
				s.rebuildRows(s.width)
			}
		}
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

func (a *app) backendAction(action string) {
	if a.active == nil {
		return
	}
	a.backendError(a.backend.Command(a.ctx, daemon.Command{Action: action, Session: a.active.id}, nil))
}
