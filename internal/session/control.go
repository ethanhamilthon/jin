package session

import (
	"errors"

	"jin/internal/core"
	"jin/internal/tools"
)

// Interrupt stops a prompt reload, a shell command or the running request.
func (m *Manager) Interrupt(id string) error {
	return m.Do(id, func(s *Session) error {
		switch {
		case s.render != nil:
			s.render.cancel()
		case s.shell != nil:
			s.shell()
		default:
			s.agent.Interrupt()
		}
		return nil
	})
}

// Answer delivers the answers to a pending ask_user call.
func (m *Manager) Answer(id string, answers []string) error {
	return m.Do(id, func(s *Session) error {
		if s.ask == nil {
			return errors.New("No question is waiting")
		}
		s.add(Entry{Kind: core.UpdateAsk, Text: "Questions\n" + tools.FormatAnswers(s.ask, answers)})
		s.ask = nil
		s.agent.Answer(answers)
		s.emitState()
		return nil
	})
}
