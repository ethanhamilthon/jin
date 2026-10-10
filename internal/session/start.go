package session

import (
	"context"

	"jin/internal/agentkit"
	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/sources"
	"jin/internal/store"
	"jin/internal/tools"
)

// start makes a session with its agent; the agent runs once the prompt
// commands are rendered.
func (m *Manager) start(cfg store.Config, dir, id, providerID, model, effort string, messages []provider.Message, entries []Entry) *Session {
	ctx, stop := context.WithCancel(m.ctx)
	names := tools.Without(cfg.ToolsDisabled)
	client, providerID, missing := sources.ClientFor(m.db, cfg, providerID)
	client.SetStallTimeout(cfg.StallTimeout)
	agent := agentkit.New(agentkit.Spec{
		Mode: agentkit.Interactive, Client: client, Names: names,
		Dir: dir, Owner: id,
	})
	requests := make(chan core.Request, 8)
	s := &Session{
		m: m, id: id, path: dir, provider: providerID, client: client, agent: agent, names: names,
		prompts: requests, requests: requests, updates: make(chan core.Update, 64), runCtx: ctx, stop: stop,
		model: model, effort: effort, entries: entries, providerMissing: missing,
	}
	if missing {
		s.add(Entry{Kind: core.UpdateError, Text: missingProviderText(providerID)})
	}
	m.sessions[id] = s
	go m.pump(s)
	m.beginRender(cfg, s, messages)
	return s
}

// pump applies the updates of a session's agent in order.
func (m *Manager) pump(s *Session) {
	for {
		select {
		case <-s.runCtx.Done():
			return
		case update, ok := <-s.updates:
			if !ok {
				return
			}
			m.mu.Lock()
			m.apply(s, update)
			m.flush()
			m.mu.Unlock()
		}
	}
}

// clientFor builds the client of a provider; missing is true when the
// session's provider was deleted.
func clientFor(cfg store.Config, id string) (*provider.Client, string, bool) {
	return sources.ClientFor(nil, cfg, id)
}

func missingProviderText(id string) string {
	return "Provider " + id + " of this session was deleted. Pick another one in Providers before you send."
}
