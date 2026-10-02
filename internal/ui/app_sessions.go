package ui

import (
	"context"
	"crypto/rand"
	"encoding/hex"

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
	agent := core.NewAgent(client, core.SystemPrompt(a.dir, a.cfg.HooksDisabled, a.cfg.JinDocs, names), registry)
	prompts := make(chan core.Request, 8)
	updates := make(chan core.Update, 64)
	s := &chatSession{
		id: id, path: a.dir, store: a.store, provider: providerID, client: client, agent: agent, prompts: prompts, stop: stop,
		model: model, effort: effort, width: a.width, pricing: a.pricing, fold: a.fold,
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
	s := a.startSession(rec.ID, rec.Provider, rec.Model, rec.Effort, core.SinceLastSummary(messages), historyToEntries(messages, a.registry))
	s.persisted, s.title, s.usage = true, rec.Title, rec.Usage
	if items, err := a.store.LoadTodos(rec.ID); err == nil {
		s.todos = items
	}
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
