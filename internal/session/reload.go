package session

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"jin/internal/agentkit"
	"jin/internal/core"
	"jin/internal/hooks"
	"jin/internal/startup"
	"jin/internal/store"
)

// Reload runs the session's prompts, hooks and instructions again without
// starting another agent loop.
func (m *Manager) Reload(id string) error {
	cfg, err := m.db.LoadConfig()
	if err != nil {
		return err
	}
	return m.Do(id, func(s *Session) error {
		switch {
		case !s.ready:
			return errors.New("The session is still starting")
		case s.render != nil && s.render.reload:
			return errors.New("A prompt reload is already running")
		case s.busy():
			return errors.New("Cannot reload while the session is working")
		case s.readOnlyPID != 0:
			return errors.New(readOnlyText(s.readOnlyPID))
		}
		in := m.renderInput(cfg, s, true)
		trust, _ := m.db.HooksTrust(s.path)
		hookCount := len(hooks.ActiveIn(s.path, cfg.HooksDisabled, trust == store.Trusted))
		ctx, cancel := context.WithCancel(s.runCtx)
		render := &rendering{cancel: cancel, reload: true}
		s.render = render
		s.add(Entry{Kind: core.UpdateInfo, Text: "Reloading session prompts"})
		s.emitState()
		go func() {
			out := startup.Render(ctx, in, nil)
			var failure error
			if ctx.Err() == nil {
				failure = s.agent.ReloadPrompts(ctx, out.System, out.Compact, out.Handoff, agentkit.Refresher(in))
			}
			m.mu.Lock()
			defer m.mu.Unlock()
			if s.render == render {
				m.finishReload(s, out, hookCount, ctx.Err() != nil, failure)
			}
		}()
		return nil
	})
}

func (m *Manager) finishReload(s *Session, out startup.Output, hookCount int, cancelled bool, failure error) {
	s.render.cancel()
	s.render = nil
	m.retryTasks(s.id)
	switch {
	case cancelled:
		s.add(Entry{Kind: core.UpdateInfo, Text: "Prompt reload cancelled"})
	case failure != nil:
		s.add(Entry{Kind: core.UpdateError, Text: "Prompt reload failed: " + failure.Error()})
	default:
		s.bodies = out.Prompts
		s.add(Entry{Kind: core.UpdateInfo, Text: "Reloaded system, compact, and handoff prompts: " + countLabel(hookCount, "hook") + ", " + countLabel(len(out.Prompts), "#prompt")})
		if len(out.Warnings) > 0 {
			s.add(Entry{Kind: core.UpdateInfo, Text: "Prompt reload warnings: " + strings.Join(out.Warnings, "; ")})
		}
	}
	s.emitState()
}

func countLabel(count int, name string) string {
	if count == 1 {
		return "1 " + name
	}
	return strconv.Itoa(count) + " " + name + "s"
}
