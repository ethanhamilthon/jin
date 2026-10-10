package session

import (
	"errors"

	"jin/internal/core"
	"jin/internal/tools"
)

// Interrupt stops a prompt reload, a shell command or the running request.
func (m *Manager) Interrupt(id string) error {
	return m.Do(id, func(s *Session) error {
		s.paused = true
		switch {
		case s.render != nil:
			s.render.cancel()
		case s.shell != nil:
			s.shell()
		default:
			s.agent.Interrupt()
		}
		s.emitState()
		return nil
	})
}

func (m *Manager) Resume(id string) error {
	return m.Do(id, func(s *Session) error {
		s.paused = false
		m.flush()
		s.emitState()
		return nil
	})
}

// Answer delivers the answers to a pending ask_user call.
func (m *Manager) Answer(id string, answers []string) error {
	return m.AnswerQuestion(id, 0, answers)
}

func (m *Manager) AnswerQuestion(id string, question int64, answers []string) error {
	return m.Do(id, func(s *Session) error {
		if question != 0 && question != s.questionRevision {
			return errors.New("This question has changed or was already answered")
		}
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
