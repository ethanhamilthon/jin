package ui

import (
	"errors"

	"jin/internal/core"
)

func (s *chatSession) window() int {
	entry, _ := s.pricing.Lookup(s.model)
	return entry.MaxInputTokens
}

// noVision is true only when the catalogues say the model takes no images.
func (s *chatSession) noVision() bool {
	entry, ok := s.pricing.Lookup(s.model)
	return ok && entry.VisionKnown && !entry.Vision
}

// queueSide asks the agent to compact the conversation or write a handoff
// brief. It refuses, with an error row, when the request cannot run now.
func (a *app) queueSide(kind core.RequestKind) {
	s := a.active
	if err := a.sideRefusal(s); err != nil {
		s.closeOpenEntry()
		s.appendEntry(chatEntry{kind: core.UpdateError, text: err.Error()})
		s.scroll = 0
		return
	}
	s.pending = append(s.pending, core.Request{Kind: kind, Model: s.model, Effort: s.effort, Window: s.window()})
	_ = s.store.SetRunning(s.id, true)
}

func (a *app) sideRefusal(s *chatSession) error {
	switch {
	case !a.cfg.Provider.Ready():
		return errors.New("Provider is not ready: type /provider")
	case !s.persisted:
		return errors.New("Nothing to work with yet: this session has no messages")
	case s.working || len(s.pending) > 0:
		return errors.New("The session is working: wait for it or interrupt it first")
	}
	return nil
}

func (a *app) compactSession() { a.queueSide(core.RequestCompact) }

func (a *app) handoffSession() { a.queueSide(core.RequestHandoff) }

// startHandoff opens a new session whose input holds the brief, ready to edit.
func (a *app) startHandoff(brief string) {
	a.newSession()
	s := a.active
	s.input = clusters(brief)
	s.cursor = len(s.input)
	a.sel = nil
}
