package ui

import (
	"context"
	"slices"
	"strconv"
	"strings"

	"jin/internal/core"
	"jin/internal/hooks"
	"jin/internal/startup"
)

func (a *app) sessionRenderInput(s *chatSession, names []string, withPrompts bool) startup.Input {
	return startup.Input{
		Dir: s.path, SessionID: s.id, ToolNames: slices.Clone(names),
		HooksDisabled: slices.Clone(a.cfg.HooksDisabled), PromptsDisabled: slices.Clone(a.cfg.PromptsDisabled),
		ProjectHooks: a.projectHooksTrustedAt(s.path), WithPrompts: withPrompts,
	}
}

func systemRefresher(in startup.Input) func(context.Context) string {
	in.WithPrompts = false
	return func(ctx context.Context) string { return startup.Render(ctx, in, nil).System }
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
	hookCount := len(hooks.ActiveIn(in.Dir, in.HooksDisabled, in.ProjectHooks))
	ctx, cancel := context.WithCancel(s.runCtx)
	s.render = &rendering{cancel: cancel, reload: true}
	s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Reloading session prompts"})
	id, agent := s.id, s.agent
	go func() {
		out := startup.Render(ctx, in, nil)
		ev := renderEvent{session: id, out: &out, reload: true, hookCount: hookCount, cancelled: ctx.Err() != nil}
		if !ev.cancelled {
			if err := agent.ReloadPrompts(ctx, out.System, out.Compact, out.Handoff, systemRefresher(in)); err != nil {
				ev.err = err.Error()
			}
		}
		select {
		case a.rendered <- ev:
		case <-a.ctx.Done():
		}
	}()
}

func (a *app) finishReload(s *chatSession, ev renderEvent) {
	if s.render == nil || !s.render.reload {
		return
	}
	s.render.cancel()
	s.render = nil
	if a.store != nil {
		a.retryAsync(s.id)
	}
	switch {
	case ev.cancelled:
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Prompt reload cancelled"})
	case ev.err != "":
		s.appendEntry(chatEntry{kind: core.UpdateError, text: "Prompt reload failed: " + ev.err})
	case ev.out == nil:
		s.appendEntry(chatEntry{kind: core.UpdateError, text: "Prompt reload failed: no render output"})
	default:
		s.promptBodies = ev.out.Prompts
		s.customSystem = ev.out.Custom
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Reloaded system, compact, and handoff prompts: " + reloadCounts(ev.hookCount, len(ev.out.Prompts))})
		if len(ev.out.Warnings) > 0 {
			s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Prompt reload warnings: " + strings.Join(ev.out.Warnings, "; ")})
		}
	}
}

func reloadCounts(hooks, prompts int) string {
	return strings.Join([]string{countLabel(hooks, "hook"), countLabel(prompts, "#prompt")}, ", ")
}

func countLabel(count int, name string) string {
	if count == 1 {
		return "1 " + name
	}
	return strconv.Itoa(count) + " " + name + "s"
}
