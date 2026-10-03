package ui

import (
	"time"

	jinupdate "jin/internal/update"
)

// checkUpdate looks for a newer release in the background when a session
// starts. The answer comes back through a.newVersion; the network is asked
// at most every few hours, sessions in between reuse the saved answer.
func (a *app) checkUpdate() {
	if a.store == nil || a.newVersion == nil || a.checkingUpdate {
		return
	}
	a.checkingUpdate = true
	go func() {
		tag := jinupdate.Available(a.ctx, a.store, a.version, time.Now())
		select {
		case a.newVersion <- tag:
		case <-a.ctx.Done():
		}
	}()
}

// receiveUpdate records a found release. The intro of every new session
// shows it from then on; a fresh session already on screen gets it at once.
func (a *app) receiveUpdate(tag string) {
	a.checkingUpdate = false
	if tag == "" || tag == a.latest {
		return
	}
	a.latest = tag
	if s := a.active; s != nil && !s.persisted && len(s.pending) == 0 {
		s.appendEntry(updateSection(a.version, tag))
	}
}

func updateSection(current, tag string) chatEntry {
	return section("Update", "jin "+tag+" is available (you have "+current+") · run jin update in a shell")
}
