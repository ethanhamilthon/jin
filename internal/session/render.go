package session

import (
	"context"
	"slices"
	"strings"

	"jin/internal/agentkit"
	"jin/internal/core"
	"jin/internal/prompts"
	"jin/internal/provider"
	"jin/internal/startup"
	"jin/internal/store"
)

func (m *Manager) renderInput(cfg store.Config, s *Session, withPrompts bool) startup.Input {
	trust, _ := m.db.HooksTrust(s.path)
	return startup.Input{
		Dir: s.path, SessionID: s.id, ToolNames: slices.Clone(s.names),
		HooksDisabled: slices.Clone(cfg.HooksDisabled), PromptsDisabled: slices.Clone(cfg.PromptsDisabled),
		ProjectHooks: trust == store.Trusted, WithPrompts: withPrompts,
	}
}

// beginRender runs the commands of the session's prompts in the background;
// the agent starts when they are done.
func (m *Manager) beginRender(cfg store.Config, s *Session, messages []provider.Message) {
	in := m.renderInput(cfg, s, true)
	s.agent.SetRefresher(agentkit.Refresher(in))
	ctx, cancel := context.WithCancel(s.runCtx)
	s.render = &rendering{cancel: cancel}
	s.initial = messages
	infos, _ := prompts.ListInfo()
	for _, info := range infos {
		if !slices.Contains(cfg.PromptsDisabled, info.Name) {
			s.render.loading = append(s.render.loading, info.Name)
		}
	}
	render := s.render
	go func() {
		out := startup.Render(ctx, in, func(name string) {
			m.mu.Lock()
			defer m.mu.Unlock()
			if s.render == render {
				render.loading = slices.DeleteFunc(render.loading, func(n string) bool { return n == name })
				s.emitState()
			}
		})
		m.mu.Lock()
		defer m.mu.Unlock()
		if s.render == render {
			m.finishRender(s, out, ctx.Err() != nil && s.runCtx.Err() == nil)
		}
	}()
}

// finishRender takes the texts into the session and starts its agent,
// unless the session was stopped meanwhile.
func (m *Manager) finishRender(s *Session, out startup.Output, cancelled bool) {
	s.render.cancel()
	s.render = nil
	if s.runCtx.Err() != nil {
		return
	}
	s.bodies = out.Prompts
	agentkit.Apply(s.agent, out)
	s.ready = true
	switch {
	case cancelled:
		s.add(Entry{Kind: core.UpdateInfo, Text: "Prompt commands cancelled"})
	case len(out.Warnings) > 0:
		s.add(Entry{Kind: core.UpdateInfo, Text: "Prompt commands: " + strings.Join(out.Warnings, "; ")})
	}
	s.done = make(chan struct{})
	done, initial := s.done, s.initial
	agent, ctx, requests, updates := s.agent, s.runCtx, s.requests, s.updates
	go func() {
		defer close(done)
		agent.Run(ctx, initial, requests, updates)
	}()
	s.initial = nil
	m.flush()
	s.emitState()
}

func (s *Session) loadingNames() []string {
	if s.render == nil || s.render.reload {
		return nil
	}
	return slices.Clone(s.render.loading)
}
