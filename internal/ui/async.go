package ui

import (
	"time"

	"jin/internal/store"
)

// asyncEntry marks a chat entry that shows an async task result.
const asyncEntry = "async"

const asyncPoll = time.Second

// asyncBatch is what one poll of the database found for this directory.
type asyncBatch struct {
	events  []store.AsyncEvent
	running map[string]int
}

// pollAsync reads the async tables once a second. Events are claimed here,
// so the other TUIs of the same directory do not hand them out twice; the UI
// thread never waits for the database.
func (a *app) pollAsync(paths []string) {
	ticker := time.NewTicker(asyncPoll)
	defer ticker.Stop()
	for {
		select {
		case <-a.ctx.Done():
			return
		case next := <-a.asyncPaths:
			paths = next
			continue
		case <-ticker.C:
		}
		_ = a.store.ReleaseDeadAsyncClaims()
		batch := a.pollAsyncPaths(paths)
		events, running := batch.events, batch.running
		select {
		case a.asyncs <- asyncBatch{events: events, running: running}:
		case <-a.ctx.Done():
			return
		}
	}
}

// receiveAsync hands the events of a poll to their sessions. An event is
// removed from the database when its message is saved in the session history.
func (a *app) receiveAsync(batch asyncBatch) {
	a.asyncRunning = batch.running
	for _, event := range batch.events {
		a.deliverEvent(event)
	}
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

// backgroundWaiting reports a session that has background tasks running and
// no request of its own. A running request has priority.
func (a *app) backgroundWaiting(s *chatSession) bool {
	return !s.working && a.asyncRunning[s.id] > 0
}
