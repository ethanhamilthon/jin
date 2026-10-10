package session

import (
	"errors"

	"jin/internal/core"
)

// Compact asks the agent to summarize the conversation.
func (m *Manager) Compact(id string) error { return m.side(id, core.RequestCompact) }

// Handoff asks the agent for a brief that a fresh session continues from.
func (m *Manager) Handoff(id string) error { return m.side(id, core.RequestHandoff) }

func (m *Manager) side(id string, kind core.RequestKind) error {
	return m.Do(id, func(s *Session) error {
		switch {
		case !s.persisted:
			return errors.New("Nothing to work with yet: this session has no messages")
		case s.working || s.inflight > 0 || (!s.paused && len(s.pending) > 0):
			return errors.New("The session is working: wait for it or interrupt it first")
		}
		if err := s.sendRefusal(); err != nil {
			return err
		}
		request := core.Request{Kind: kind, Model: s.model, Effort: s.effort, Window: s.window()}
		if s.paused {
			select {
			case s.prompts <- request:
				s.inflight++
			default:
				return errors.New("The session is still stopping; wait before compacting")
			}
		} else {
			s.pending = append(s.pending, request)
			m.flush()
		}
		s.emitState()
		return nil
	})
}

// handoff opens a new session whose draft holds the brief.
func (m *Manager) handoff(from *Session, brief string) {
	cfg, err := m.db.LoadConfig()
	if err != nil {
		from.persistenceError("Handoff failed", err)
		return
	}
	s := m.start(cfg, from.path, NewID(), from.provider, from.model, from.effort, nil, nil)
	s.intro = m.intro(cfg, s.path, nil)
	s.setDraft(brief)
	from.add(Entry{Kind: core.UpdateInfo, Text: "Handoff ready in session " + ShortID(s.id)})
	from.emit(Event{Type: "handoff", Text: s.id})
}

func (s *Session) setDraft(text string) {
	s.draft = text
	s.draftRev++
}
