package ui

import (
	"os"
	"os/exec"
	"testing"
)

func TestCrashBetweenClaimAndSaveDoesNotLoseTheEvent(t *testing.T) {
	a, s := asyncApp(t)
	cmd := exec.Command("true")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	a.receiveAsync(claimedBatch(t, a, cmd.Process.Pid))
	if len(s.pending) != 1 {
		t.Fatalf("pending = %d", len(s.pending))
	}
	// the process dies here: nothing was saved, nothing was acknowledged
	if err := a.store.ReleaseDeadAsyncClaims(); err != nil {
		t.Fatal(err)
	}
	if n := eventCount(t, a, os.Getpid()); n != 1 {
		t.Errorf("events after the restart = %d, want 1", n)
	}
}

func TestEventOfASessionThatCannotTakeItStaysForARetry(t *testing.T) {
	a, s := asyncApp(t)
	s.providerMissing = true
	a.receiveAsync(claimedBatch(t, a, os.Getpid()))
	if len(s.pending) != 0 {
		t.Fatal("delivered to a session without a provider")
	}
	s.providerMissing = false
	a.retryAsync("s1")
	if n := eventCount(t, a, os.Getpid()); n != 1 {
		t.Errorf("events after the retry = %d, want 1", n)
	}
}

func TestEventOfAMissingSessionIsDropped(t *testing.T) {
	a := startingApp(t)
	_ = a.store.AddAsyncEvent("nope", a.dir, "x")
	events, _ := a.store.ClaimAsyncEvents(a.dir, os.Getpid())
	a.receiveAsync(asyncBatch{events: events})
	a.store.ReleaseAsyncEventsFor("nope", os.Getpid())
	if n := eventCount(t, a, os.Getpid()); n != 0 {
		t.Errorf("events = %d", n)
	}
}
