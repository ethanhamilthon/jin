package ui

import (
	"os"
	"time"

	"jin/internal/async"
	"jin/internal/core"
	"jin/internal/store"
)

// asyncEntry marks a chat entry that shows an async task result.
const asyncEntry = "async"

// asyncFrames is the loader shown while background tasks run and the agent
// itself is idle.
var asyncFrames = []string{"◐", "◓", "◑", "◒"}

const asyncPoll = time.Second

// asyncBatch is what one poll of the database found for this directory.
type asyncBatch struct {
	events  []store.AsyncEvent
	running map[string]int
}

// pollAsync reads the async tables once a second. Events are claimed here,
// so the other TUIs of the same directory do not hand them out twice; the UI
// thread never waits for the database.
func (a *app) pollAsync() {
	ticker := time.NewTicker(asyncPoll)
	defer ticker.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
		}
		_ = a.store.ReleaseDeadAsyncClaims()
		events, _ := a.store.ClaimAsyncEvents(a.dir, os.Getpid())
		tasks, _ := a.store.RunningAsyncTasks(a.dir)
		running := map[string]int{}
		for _, task := range tasks {
			running[task.SessionID]++
		}
		select {
		case a.asyncs <- asyncBatch{events: events, running: running}:
		case <-a.ctx.Done():
			return
		}
	}
}

// receiveAsync hands the events of a poll to their sessions.
func (a *app) receiveAsync(batch asyncBatch) {
	a.asyncRunning = batch.running
	var delivered []int64
	for _, event := range batch.events {
		delivered = append(delivered, event.ID)
		s, err := a.sessionForEvent(event)
		if err != nil {
			a.report(err)
			continue
		}
		s.sendAsync(event.Text)
	}
	if len(delivered) > 0 {
		_ = a.store.AckAsyncEvents(delivered...)
	}
}

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

// drain handles whatever the background sources have ready, without waiting.
// The /tui loop uses it to keep the agents and async tasks moving while a
// full-screen program owns the terminal.
func (a *app) drain() {
	for {
		select {
		case tagged := <-a.updates:
			a.applyUpdate(tagged.id, tagged.update)
		case result := <-a.loads:
			a.receiveLoad(result)
		case result := <-a.bashDone:
			a.receiveBash(result)
		case result := <-a.modelsLoaded:
			a.receiveModels(result)
		case batch := <-a.asyncs:
			a.receiveAsync(batch)
		case ev := <-a.rendered:
			a.receiveRender(ev)
		default:
			return
		}
	}
}

// backgroundLoader is the loader of a session that has background tasks and
// no running request: frames and a flag. A running request has priority.
func (a *app) backgroundLoader(s *chatSession) (string, bool) {
	if s.working || a.asyncRunning[s.id] == 0 {
		return "", false
	}
	return asyncFrames[a.frame%len(asyncFrames)], true
}
