package ui

import (
	"context"
	"crypto/rand"
	"encoding/hex"

	"jin/internal/async"
	"jin/internal/core"
	"jin/internal/pricing"
	"jin/internal/provider"
	"jin/internal/store"
	"jin/internal/tools"
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

func (a *app) startSession(id, providerID, model, effort string, messages []provider.Message, entries []chatEntry) *chatSession {
	ctx, stop := context.WithCancel(a.ctx)
	names := tools.Without(a.cfg.ToolsDisabled)
	registry := tools.Build(names, store.SessionTodos{DB: a.store, ID: id})
	client, providerID := a.clientFor(providerID)
	client.SetStallTimeout(a.cfg.StallTimeout)
	// The system prompt is filled in by the background render; the agent
	// starts when it is done.
	agent := core.NewAgent(client, "", registry)
	agent.SetBackground(func(adoption tools.Adoption) (string, error) {
		return async.Adopt(a.version, id, adoption.Command, adoption.PID, adoption.PGID, adoption.Log, adoption.Exit)
	})
	requests := make(chan core.Request, 8)
	updates := make(chan core.Update, 64)
	s := &chatSession{
		id: id, path: a.dir, store: a.store, provider: providerID, client: client, agent: agent, prompts: requests, requests: requests, stop: stop,
		model: model, effort: effort, width: a.width, pricing: a.pricing, fold: a.fold,
		runCtx: ctx, updatesOut: updates,
	}
	for _, entry := range entries {
		s.appendEntry(entry)
	}
	go func() {
		for update := range updates {
			a.updates <- taggedUpdate{id: id, update: update}
		}
	}()
	a.sessions[id] = s
	a.beginRender(s, ctx, names, messages)
	return s
}

func (a *app) setPricing(table pricing.Table) {
	a.pricing = table
	for _, s := range a.sessions {
		s.pricing = table
	}
}

func (a *app) newSession() {
	a.focus(a.startSession(newSessionID(), a.cfg.ActiveProvider, a.cfg.Model, a.cfg.Effort, nil, a.introEntries()))
}

func (a *app) resumeSession(rec store.Session) error {
	s, err := a.openSession(rec)
	if err != nil {
		return err
	}
	a.focus(s)
	return nil
}

// openSession starts the backend of a saved session without moving the
// focus, or returns it when it is already open.
func (a *app) openSession(rec store.Session) (*chatSession, error) {
	if s, ok := a.sessions[rec.ID]; ok {
		return s, nil
	}
	messages, err := a.store.LoadMessages(rec.ID)
	if err != nil {
		return nil, err
	}
	for _, msg := range core.InterruptedToolMessages(messages) {
		if err := a.store.AppendMessage(rec.ID, msg); err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}
	s := a.startSession(rec.ID, rec.Provider, rec.Model, rec.Effort, core.SinceLastSummary(messages), historyToEntries(messages, a.registry))
	s.persisted, s.title, s.usage = true, rec.Title, rec.Usage
	if items, err := a.store.LoadTodos(rec.ID); err == nil {
		s.todos = items
	}
	return s, nil
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
