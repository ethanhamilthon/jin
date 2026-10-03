package ui

import (
	"jin/internal/async"
	"jin/internal/core"
	"jin/internal/store"
)

// sessionForEvent finds the session of an event, opening it in the
// background when it is not open yet. The focus stays where it is.
func (a *app) sessionForEvent(event store.AsyncEvent) (*chatSession, error) {
	if s, ok := a.sessions[event.SessionID]; ok {
		return s, nil
	}
	rec, found, err := a.store.GetSession(event.SessionID)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errAsyncSession(event.SessionID)
	}
	return a.openSession(rec)
}

type errAsyncSession string

func (e errAsyncSession) Error() string {
	return "async: the result of a task was dropped, session " + string(e) + " does not exist"
}

// sendAsync queues a task result for the agent. It is not typed by the user:
// no #prompt is expanded, no todo note is added, and the chat shows a short
// summary instead of a user bubble.
func (s *chatSession) sendAsync(text string) {
	s.closeOpenEntry()
	s.pending = append(s.pending, core.Request{Prompt: text, Model: s.model, Effort: s.effort, Window: s.window(), NoVision: s.noVision()})
	s.appendEntry(asyncChatEntry(text))
	if s.persisted {
		_ = s.store.TouchProvider(s.id, s.path, s.model, s.effort, s.title, s.provider)
		_ = s.store.SetRunning(s.id, true)
	}
}

func asyncChatEntry(text string) chatEntry {
	summary, ok := async.Summary(text)
	if !ok {
		summary = text
	}
	return chatEntry{kind: core.UpdateInfo, tool: asyncEntry, text: summary}
}
