package ui

import (
	"errors"
	"strconv"
	"strings"

	"jin/internal/core"
)

// window is the model's context size in tokens: the catalogues first, then
// the setting models.window.<model id>, else 0 for unknown.
func (s *chatSession) window() int {
	if entry, _ := s.pricing.Lookup(s.model); entry.MaxInputTokens > 0 {
		return entry.MaxInputTokens
	}
	if s.store == nil {
		return 0
	}
	value, _ := s.store.Setting("models.window." + s.model)
	tokens, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return 0
	}
	return max(0, tokens)
}

// noVision is true only when the catalogues say the model takes no images.
func (s *chatSession) noVision() bool {
	entry, ok := s.pricing.Lookup(s.model)
	return ok && entry.VisionKnown && !entry.Vision
}

// queueSide asks the agent to compact the conversation or write a handoff
// brief. It refuses, with an error row, when the request cannot run now.
func (a *app) queueSide(kind core.RequestKind) {
	if a.backend != nil {
		action := "compact"
		if kind == core.RequestHandoff {
			action = "handoff"
		}
		a.backendAction(action)
		return
	}
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
	case !s.persisted:
		return errors.New("Nothing to work with yet: this session has no messages")
	case s.working || s.inflight > 0 || len(s.pending) > 0:
		return errors.New("The session is working: wait for it or interrupt it first")
	}
	return s.sendRefusal()
}

func (a *app) compactSession() { a.queueSide(core.RequestCompact) }

func (a *app) handoffSession() { a.queueSide(core.RequestHandoff) }

// startHandoff opens a new session whose input holds the brief, ready to edit.
func (a *app) startHandoff(brief string) { a.startHandoffFrom(a.active, brief) }

func (a *app) startHandoffFrom(from *chatSession, brief string) {
	dir := from.path
	if dir == "" {
		dir = a.dir
	}
	s := a.startSessionAt(dir, newSessionID(), from.provider, from.model, from.effort, nil, a.introEntriesAt(dir))
	s.input = clusters(brief)
	s.cursor = len(s.input)
	s.draftRevision++
	if a.active == from {
		a.focus(s)
		a.sel = nil
	} else {
		from.appendEntry(chatEntry{kind: core.UpdateInfo, text: "Handoff ready in session " + shortID(s.id) + " · open it with /sessions"})
	}
}
