package ui

import (
	"jin/internal/core"
	"jin/internal/session"
	"jin/internal/tasks"
)

// taskEntry marks a chat entry that shows a background task result.
const taskEntry = "task"

// receiveTask hands the result of a finished task to the session that
// started it. A session that cannot take it now keeps it until it can.
func (a *app) receiveTask(event tasks.Event) {
	s, err := a.sessionFor(event.Owner)
	if err != nil {
		a.report(err)
		return
	}
	if s.providerMissing || s.readOnlyPID != 0 || s.projectError() != nil || (s.render != nil && s.render.reload) || s.sendRefusal() != nil {
		a.heldTasks = append(a.heldTasks, event)
		return
	}
	s.sendTaskResult(event.Text)
}

// retryTasks delivers the held results of a session again.
func (a *app) retryTasks(sessionID string) {
	held := a.heldTasks
	a.heldTasks = nil
	for _, event := range held {
		if event.Owner == sessionID {
			a.receiveTask(event)
		} else {
			a.heldTasks = append(a.heldTasks, event)
		}
	}
}

// sessionFor finds a session, opening it in the background when it is not
// open; the focus stays where it is.
func (a *app) sessionFor(id string) (*chatSession, error) {
	if s, ok := a.sessions[id]; ok {
		return s, nil
	}
	rec, found, err := a.store.GetSession(id)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errTaskSession(id)
	}
	return a.openSession(rec)
}

type errTaskSession string

func (e errTaskSession) Error() string {
	return "the result of a background task was dropped: session " + string(e) + " does not exist"
}

// sendTaskResult queues a task result for the agent. It is not typed by the
// user: no #prompt is expanded, no todo note is added, and the chat shows a
// short summary instead of a user bubble.
func (s *chatSession) sendTaskResult(text string) {
	s.closeOpenEntry()
	s.pending = append(s.pending, core.Request{Prompt: text, Model: s.model, Effort: s.effort, Window: s.window(), NoVision: s.noVision()})
	s.appendEntry(taskChatEntry(text))
	if s.persisted {
		_ = s.store.TouchProvider(s.id, s.path, s.model, s.effort, s.title, s.provider)
		_ = s.store.SetRunning(s.id, true)
	}
}

func taskChatEntry(text string) chatEntry { return chatEntryOf(session.TaskEntry(text)) }

// backgroundWaiting reports a session that has background tasks running and
// no request of its own. A running request has priority.
func (a *app) backgroundWaiting(s *chatSession) bool {
	return !s.working && a.tasksRunning[s.id] > 0
}
