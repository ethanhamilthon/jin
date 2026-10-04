package ui

import (
	"os"
	"testing"
	"time"

	"jin/internal/core"
)

func ownedTestSession(t *testing.T) (*app, *chatSession) {
	t.Helper()
	a, _ := layoutApp(t)
	db, _ := openFoldDB(t)
	s := &chatSession{id: "owned", path: t.TempDir(), store: db, persisted: true, ready: true, model: "m", width: 60}
	a.store, a.active, a.sessions[s.id] = db, s, s
	if err := db.Touch(s.id, s.path, s.model, "", "owned"); err != nil {
		t.Fatal(err)
	}
	if err := db.SetRunning(s.id, true); err != nil {
		t.Fatal(err)
	}
	return a, s
}

func assertOwned(t *testing.T, a *app, s *chatSession, want bool) {
	t.Helper()
	pid, alive, err := a.store.SessionOwner(s.id)
	if err != nil || alive != want || want && pid != os.Getpid() {
		t.Fatalf("owner pid=%d alive=%v err=%v, want owned=%v", pid, alive, err, want)
	}
}

func TestShellCompletionCannotReleaseAHandedOffRequest(t *testing.T) {
	a, s := ownedTestSession(t)
	s.prompts = make(chan core.Request, 1)
	s.pending = []core.Request{{Prompt: "queued request"}}
	s.bash = &bashState{running: true}
	a.flushPending()
	if len(s.pending) != 0 || s.inflight != 1 || s.working {
		t.Fatal("test did not reproduce the pre-UpdateWorking handoff window")
	}
	a.receiveBash(bashResult{session: s.id, output: "finished"})
	assertOwned(t, a, s, true)
	a.applyUpdate(s.id, core.Update{Kind: core.UpdateDone})
	assertOwned(t, a, s, false)
}

func TestShutdownCancelsBeforeReleasingBackendOwnership(t *testing.T) {
	a, s := ownedTestSession(t)
	cancelled, backendDone, finished := make(chan struct{}), make(chan struct{}), make(chan struct{})
	s.backendDone, s.inflight = backendDone, 1
	s.stop = func() { close(cancelled) }
	go func() { a.shutdownSessions(); close(finished) }()
	select {
	case <-cancelled:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not cancel the backend")
	}
	assertOwned(t, a, s, true)
	close(backendDone)
	select {
	case <-finished:
	case <-time.After(time.Second):
		t.Fatal("shutdown did not finish after backend termination")
	}
	assertOwned(t, a, s, false)
	unread, err := a.store.UnreadSessions(s.path)
	if err != nil || !unread[s.id] {
		t.Fatalf("unfinished request not marked unread: %v, %v", unread, err)
	}
}
