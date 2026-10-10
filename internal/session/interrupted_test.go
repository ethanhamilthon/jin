package session

import (
	"strings"
	"testing"

	"jin/internal/core"
)

// A daemon that dies while it answered leaves the session marked: the next
// open says that the request and its queued messages were lost. Reopening the
// session again does not repeat the notice.
func TestInterruptedTurnIsReportedOnOpen(t *testing.T) {
	manager, events, dir := newTestManager(t)
	snap, err := manager.Create(dir)
	if err != nil {
		t.Fatal(err)
	}
	id := snap.State.ID
	waitFor(t, events, func(event Event) bool { return event.State != nil && event.State.Ready })
	if err := manager.Send(id, "say hi", nil, nil); err != nil {
		t.Fatal(err)
	}
	waitFor(t, events, func(event Event) bool { return event.Type == "ring" && event.Kind == core.UpdateDone })
	db := manager.DB()
	if err := db.SetRunning(id, true); err != nil {
		t.Fatal(err)
	}
	if rows, err := db.SetDeadForTest(); err != nil || rows != 1 {
		t.Fatalf("could not mark the request dead: %d %v", rows, err)
	}
	manager.Shutdown()
	if err := db.RecoverInterrupted(); err != nil {
		t.Fatal(err)
	}
	second := NewManager(t.Context(), db, "v0", nil)
	defer second.Shutdown()
	fresh, err := second.Open(id)
	if err != nil {
		t.Fatal(err)
	}
	if !hasInterruptedNotice(fresh.Entries) {
		t.Fatalf("no interrupted notice in %+v", fresh.Entries)
	}
	if lost, err := db.InterruptedSessions(); err != nil || len(lost) != 0 {
		t.Fatalf("marker not cleared: %v %v", lost, err)
	}
	third := NewManager(t.Context(), db, "v0", nil)
	defer third.Shutdown()
	reopened, err := third.Open(id)
	if err != nil {
		t.Fatal(err)
	}
	if hasInterruptedNotice(reopened.Entries) {
		t.Fatal("the notice repeated on a later open")
	}
}

func hasInterruptedNotice(entries []Entry) bool {
	for _, entry := range entries {
		if entry.Kind == core.UpdateError && strings.Contains(entry.Text, "interrupted") {
			return true
		}
	}
	return false
}
