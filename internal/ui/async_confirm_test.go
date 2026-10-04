package ui

import (
	"os"
	"os/exec"
	"testing"

	"jin/internal/provider"
)

const taskResultText = "<async-task-result id=\"a1\" status=\"done\" exit=\"0\">\nok\n</async-task-result>"

func eventCount(t *testing.T, a *app, pid int) int {
	t.Helper()
	events, err := a.store.ClaimAsyncEvents(a.dir, pid)
	if err != nil {
		t.Fatal(err)
	}
	return len(events)
}

func asyncApp(t *testing.T) (*app, *chatSession) {
	a := startingApp(t)
	if err := a.store.TouchProvider("s1", a.dir, "m", "", "title", ""); err != nil {
		t.Fatal(err)
	}
	rec, _, _ := a.store.GetSession("s1")
	s, err := a.openSession(rec)
	if err != nil {
		t.Fatal(err)
	}
	return a, s
}

func claimedBatch(t *testing.T, a *app, pid int) asyncBatch {
	t.Helper()
	if err := a.store.AddAsyncEvent("s1", a.dir, taskResultText); err != nil {
		t.Fatal(err)
	}
	events, err := a.store.ClaimAsyncEvents(a.dir, pid)
	if err != nil || len(events) != 1 {
		t.Fatalf("claim = %v, %v", events, err)
	}
	return asyncBatch{events: events}
}

func TestEventIsAcknowledgedOnlyAfterItsMessageIsSaved(t *testing.T) {
	a, s := asyncApp(t)
	a.receiveAsync(claimedBatch(t, a, os.Getpid()))
	if len(s.pending) != 1 || len(s.asyncAcks) != 1 {
		t.Fatalf("pending %d, acks %d", len(s.pending), len(s.asyncAcks))
	}
	a.receiveAsync(asyncBatch{})
	if len(s.asyncAcks) != 1 {
		t.Fatal("acknowledged before the message was saved")
	}
	if err := a.store.AppendMessage("s1", provider.Message{Role: "user", Content: taskResultText}); err != nil {
		t.Fatal(err)
	}
	a.receiveAsync(asyncBatch{})
	if len(s.asyncAcks) != 0 {
		t.Error("still waiting after the message was saved")
	}
	a.store.ReleaseAsyncEventsFor("s1", os.Getpid())
	if n := eventCount(t, a, os.Getpid()); n != 0 {
		t.Errorf("the saved event is still in the database: %d", n)
	}
}

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

func TestEventAlreadyInTheHistoryIsAcknowledgedWithoutAnotherDelivery(t *testing.T) {
	a, s := asyncApp(t)
	if err := a.store.AppendMessage("s1", provider.Message{Role: "user", Content: taskResultText}); err != nil {
		t.Fatal(err)
	}
	a.receiveAsync(claimedBatch(t, a, os.Getpid()))
	if len(s.pending) != 0 || len(s.asyncAcks) != 0 {
		t.Errorf("replayed event delivered again: pending %d", len(s.pending))
	}
	a.store.ReleaseAsyncEventsFor("s1", os.Getpid())
	if n := eventCount(t, a, os.Getpid()); n != 0 {
		t.Errorf("the replayed event stays in the database: %d", n)
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
