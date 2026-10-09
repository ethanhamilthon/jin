package ui

import (
	"context"
	"slices"

	"jin/internal/agentkit"
	"jin/internal/core"
	"jin/internal/hooks"
	"jin/internal/startup"
)

func (a *app) sessionRenderInput(s *chatSession, names []string, withPrompts bool) startup.Input {
	return startup.Input{Dir: s.path, SessionID: s.id, PromptsDisabled: slices.Clone(a.cfg.PromptsDisabled), WithPrompts: withPrompts}
}

// reloadSession refreshes one session without starting another agent loop.
func (a *app) reloadSession() {
	s := a.active
	if s == nil {
		return
	}
	switch {
	case !s.ready:
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "The session is still starting"})
		return
	case s.render != nil && s.render.reload:
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "A prompt reload is already running"})
		return
	case s.working || s.inflight > 0 || len(s.pending) > 0:
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Cannot reload while the session is working"})
		return
	case s.readOnlyPID != 0:
		s.appendEntry(chatEntry{kind: core.UpdateError, text: readOnlyText(s.readOnlyPID)})
		return
	}

	if a.store != nil {
		cfg, err := a.store.LoadConfig()
		if err != nil {
			a.report(err)
			return
		}
		a.cfg.HooksDisabled, a.cfg.PromptsDisabled = cfg.HooksDisabled, cfg.PromptsDisabled
	}
	in := a.sessionRenderInput(s, s.toolNames, true)
	hookCount := len(hooks.ActiveIn(s.path, a.cfg.HooksDisabled, a.projectHooksTrustedAt(s.path)))
	ctx, cancel := context.WithCancel(s.runCtx)
	s.render = &rendering{cancel: cancel, reload: true}
	s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Reloading session prompts"})
	id, agent := s.id, s.agent
	go func() {
		out := startup.Render(ctx, in, nil)
		ev := renderEvent{session: id, out: &out, reload: true, hookCount: hookCount, cancelled: ctx.Err() != nil}
		if !ev.cancelled {
			if err := agent.ReloadPrompts(ctx, out.System, out.Compact, out.Handoff, agentkit.Refresher(in)); err != nil {
				ev.err = err.Error()
			}
		}
		select {
		case a.rendered <- ev:
		case <-a.ctx.Done():
		}
	}()
}
