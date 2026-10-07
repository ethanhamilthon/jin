package session

import (
	"jin/internal/core"
	"jin/internal/todo"
)

// apply shows one update of the session's agent.
func (m *Manager) apply(s *Session, u core.Update) {
	switch u.Kind {
	case core.UpdateReset:
		s.dropAttempt()
		return
	case core.UpdateHandoff:
		m.handoff(s, u.Text)
		return
	case core.UpdateAssistantDelta:
		s.appendDelta(core.UpdateAssistant, u.Text)
		return
	case core.UpdateReasoningDelta:
		s.appendDelta(core.UpdateReasoning, u.Text)
		return
	case core.UpdateTaken:
		m.release(s)
		s.emitState()
		return
	}
	s.closeOpen()
	switch u.Kind {
	case core.UpdateWorking:
		s.working = true
	case core.UpdateDone:
		m.done(s, u.Final)
	case core.UpdateUsage:
		s.applyUsage(u)
	case core.UpdateCompacted:
		s.applyUsage(u)
		if u.Usage.Known {
			s.usage.Context = u.Usage.Output
			s.persistUsage()
		}
		label := u.Text
		if label == "" {
			label = CompactedLabel
		}
		s.add(Entry{Kind: core.UpdateCompacted, Text: label})
	case core.UpdateHistory:
		if u.Message.Role == "assistant" {
			s.attempt = nil
		}
		s.joinNext = u.Message.Role == "user" && core.IsTasksNote(u.Message.Content)
		s.persistMessage(u.Message)
		return
	case core.UpdateToolCall:
		if u.Tool != "tell_user" {
			s.add(Entry{Kind: u.Kind, Tool: u.Tool, Text: u.Text})
		}
	case core.UpdateSuggest:
		s.suggestion = u.Text
	case core.UpdateAsk:
		s.ask = u.Questions
		m.publish(Event{Type: "ring", Session: s.id, Kind: core.UpdateAsk})
	case core.UpdateTodo:
		s.todos = u.Todos
		if todo.AllDone(u.Todos) {
			s.add(Entry{Kind: core.UpdateTodo, Text: TodoText(u.Todos)})
		}
	case core.UpdateToolResult:
		s.recordChanges(u.Changes)
		s.add(Entry{Kind: u.Kind, Tool: u.Tool, Lines: toLines(capLines(ResultLines(u.Tool, u.Text, u.Changes)))})
	default:
		s.add(Entry{Kind: u.Kind, Tool: u.Tool, Text: u.Text})
	}
	s.emitState()
}

// done ends a request: the session may be released and marked unread.
func (m *Manager) done(s *Session, final bool) {
	s.attempt = nil
	s.working, s.changeTurn, s.ask = false, 0, nil
	m.release(s)
	m.publish(Event{Type: "ring", Session: s.id, Kind: core.UpdateDone, Text: boolText(final)})
	if !s.persisted {
		return
	}
	s.unread = true
	_ = m.db.SetUnread(s.id, true)
	m.publish(Event{Type: "sessions"})
}

// release counts one request as finished and gives the session back to
// other processes once nothing of it runs.
func (m *Manager) release(s *Session) {
	if s.inflight > 0 {
		s.inflight--
	}
	if s.persisted && !s.busy() {
		_ = m.db.SetRunning(s.id, false)
	}
}

func boolText(b bool) string {
	if b {
		return "final"
	}
	return ""
}
