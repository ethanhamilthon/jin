package ui

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/store"
)

type taggedUpdate struct {
	id     string
	update core.Update
}

func newSessionID() string {
	var buf [16]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

func (a *app) startSession(id, model, effort string, messages []provider.Message, entries []chatEntry) *chatSession {
	ctx, stop := context.WithCancel(a.ctx)
	agent := core.NewAgent(a.client, core.SystemPrompt(a.dir), a.registry)
	prompts := make(chan core.Request, 8)
	updates := make(chan core.Update, 64)
	s := &chatSession{
		id: id, path: a.dir, store: a.store, agent: agent, prompts: prompts, stop: stop,
		model: model, effort: effort, width: a.width,
	}
	for _, entry := range entries {
		s.appendEntry(entry)
	}
	go agent.Run(ctx, messages, prompts, updates)
	go func() {
		for update := range updates {
			a.updates <- taggedUpdate{id: id, update: update}
		}
	}()
	a.sessions[id] = s
	return s
}

func (a *app) newSession() {
	a.focus(a.startSession(newSessionID(), a.cfg.Model, a.cfg.Effort, nil, a.introEntries()))
}

func (a *app) resumeSession(rec store.Session) error {
	if s, ok := a.sessions[rec.ID]; ok {
		a.focus(s)
		return nil
	}
	messages, err := a.store.LoadMessages(rec.ID)
	if err != nil {
		return err
	}
	for _, msg := range core.InterruptedToolMessages(messages) {
		if err := a.store.AppendMessage(rec.ID, msg); err != nil {
			return err
		}
		messages = append(messages, msg)
	}
	s := a.startSession(rec.ID, rec.Model, rec.Effort, messages, historyToEntries(messages, a.registry))
	s.persisted, s.title, s.usage = true, rec.Title, rec.Usage
	a.focus(s)
	return nil
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
		a.dropBlank()
	}
	a.active = s
	s.unread = false
	if s.persisted {
		_ = a.store.SetUnread(s.id, false)
	}
}
