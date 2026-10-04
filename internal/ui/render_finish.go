package ui

import (
	"jin/internal/core"
	"jin/internal/startup"
	"slices"
	"strings"
)

// finishRender takes the texts into the session and starts its agent.
func (a *app) finishRender(s *chatSession, out startup.Output, cancelled bool) {
	s.render.cancel()
	s.render = nil
	s.promptBodies = out.Prompts
	s.customSystem = out.Custom
	s.agent.SetSystemPrompt(out.System)
	s.agent.SetSidePrompts(out.Compact, out.Handoff)
	s.ready = true
	s.syncLoading()
	switch {
	case cancelled:
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Prompt commands cancelled"})
	case len(out.Warnings) > 0:
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Prompt commands: " + strings.Join(out.Warnings, "; ")})
	}
	done := make(chan struct{})
	s.backendDone = done
	initial := s.initial
	agent, ctx, requests, updates := s.agent, s.runCtx, s.requests, s.updatesOut
	go func() {
		defer close(done)
		agent.Run(ctx, initial, requests, updates)
	}()
	s.initial = nil
}

// cancelRender is Ctrl+C while a session starts: the commands that still run
// stop, and the session becomes ready. It reports whether there was a render.
func (a *app) cancelRender(s *chatSession) bool {
	if s.render == nil || s.render.reload {
		return false
	}
	s.render.cancel()
	return true
}

// isLoading reports whether a #prompt of the session still runs commands.
func (s *chatSession) isLoading(name string) bool {
	return s.render != nil && !s.render.reload && slices.Contains(s.render.loading, name)
}

// loadingPrompts lists the prompts that still run commands, for the intro.
func (s *chatSession) loadingPrompts() []string {
	if s.render == nil || s.render.reload {
		return nil
	}
	return slices.Clone(s.render.loading)
}
