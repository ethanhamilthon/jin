package ui

import (
	"context"

	"jin/internal/agentkit"
	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/session"
	"jin/internal/tools"
)

type taggedUpdate struct {
	id     string
	update core.Update
}

func newSessionID() string { return session.NewID() }

func (a *app) startSession(id, providerID, model, effort string, messages []provider.Message, entries []chatEntry) *chatSession {
	return a.startSessionAt(a.dir, id, providerID, model, effort, messages, entries)
}

func (a *app) startSessionAt(dir, id, providerID, model, effort string, messages []provider.Message, entries []chatEntry) *chatSession {
	ctx, stop := context.WithCancel(a.ctx)
	names := tools.Without(a.cfg.ToolsDisabled)
	client, providerID, missing := a.clientFor(providerID)
	client.SetStallTimeout(a.cfg.StallTimeout)
	// The system prompt is filled in by the background render; the agent
	// starts when it is done.
	agent := agentkit.New(agentkit.Spec{
		Mode: agentkit.Interactive, Client: client, Names: names,
		Dir: dir, Owner: id,
	})
	requests := make(chan core.Request, 8)
	updates := make(chan core.Update, 64)
	s := &chatSession{
		id: id, path: dir, store: a.store, provider: providerID, client: client, agent: agent, prompts: requests, requests: requests, stop: stop,
		model: model, effort: effort, width: a.width, pricing: a.pricing, fold: a.fold, toolNames: names,
		runCtx: ctx, updatesOut: updates,
	}
	s.providerMissing = missing
	for _, entry := range entries {
		s.appendEntry(entry)
	}
	if missing {
		s.noteMissingProvider()
	}
	go func() {
		for update := range updates {
			select {
			case a.updates <- taggedUpdate{id: id, update: update}:
			case <-ctx.Done():
				return
			}
		}
	}()
	a.sessions[id] = s
	a.beginRender(s, ctx, names, messages)
	return s
}

func (a *app) newSession() {
	a.focus(a.startSession(newSessionID(), a.cfg.ActiveProvider, a.cfg.Model, a.cfg.Effort, nil, a.introEntries()))
	a.checkUpdate()
}
