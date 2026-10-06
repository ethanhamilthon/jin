package session

import (
	"jin/internal/core"
	"jin/internal/tasks"
)

// receiveTask hands the result of a finished task to the session that
// started it. A session that cannot take it now keeps it until it can.
func (m *Manager) receiveTask(ev tasks.Event) {
	s, err := m.openLocked(ev.Owner)
	if err != nil {
		m.publish(Event{Type: "notice", Text: "The result of a background task was dropped: " + err.Error()})
		return
	}
	if s.providerMissing || s.readOnlyPID != 0 || s.projectError() != nil || (s.render != nil && s.render.reload) || s.sendRefusal() != nil {
		m.held = append(m.held, ev)
		return
	}
	s.pending = append(s.pending, core.Request{Prompt: ev.Text, Model: s.model, Effort: s.effort, Window: s.window(), NoVision: s.noVision()})
	s.add(TaskEntry(ev.Text))
	if s.persisted {
		_ = m.db.TouchProvider(s.id, s.path, s.model, s.effort, s.title, s.provider)
		_ = m.db.SetRunning(s.id, true)
	}
	m.flush()
	s.emitState()
}

// retryTasks delivers the held results of a session again.
func (m *Manager) retryTasks(id string) {
	held := m.held
	m.held = nil
	for _, ev := range held {
		if ev.Owner == id {
			m.receiveTask(ev)
		} else {
			m.held = append(m.held, ev)
		}
	}
}
