package ui

import (
	"os"
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

func TestEventLeavesTheDatabaseWithItsSavedMessage(t *testing.T) {
	a, s := asyncApp(t)
	a.receiveAsync(claimedBatch(t, a, os.Getpid()))
	if len(s.pending) != 1 || len(s.asyncAcks) != 1 {
		t.Fatalf("pending %d, acks %d", len(s.pending), len(s.asyncAcks))
	}
	a.store.ReleaseAsyncEventsFor("s1", os.Getpid())
	if n := eventCount(t, a, os.Getpid()); n != 1 {
		t.Fatalf("event removed before its message was saved: %d", n)
	}
	s.persistMessage(provider.Message{Role: "user", Content: taskResultText})
	if msgs, _ := a.store.LoadMessages("s1"); len(msgs) != 1 || msgs[0].Content != taskResultText {
		t.Errorf("messages = %+v", msgs)
	}
	a.store.ReleaseAsyncEventsFor("s1", os.Getpid())
	if n := eventCount(t, a, os.Getpid()); n != 0 || len(s.asyncAcks) != 0 {
		t.Errorf("saved event still there: %d, acks %d", n, len(s.asyncAcks))
	}
}

func TestEventsKeepTheirOwnIdentity(t *testing.T) {
	a, s := asyncApp(t)
	for _, text := range []string{"first", "second"} {
		if err := a.store.AddAsyncEvent("s1", a.dir, text); err != nil {
			t.Fatal(err)
		}
	}
	events, _ := a.store.ClaimAsyncEvents(a.dir, os.Getpid())
	a.receiveAsync(asyncBatch{events: events})
	s.persistMessage(provider.Message{Role: "user", Content: "second"})
	if len(s.asyncAcks) != 1 || s.asyncAcks[0].text != "first" {
		t.Fatalf("acks = %+v", s.asyncAcks)
	}
	a.store.ReleaseAsyncEventsFor("s1", os.Getpid())
	left, _ := a.store.ClaimAsyncEvents(a.dir, os.Getpid())
	if len(left) != 1 || left[0].Text != "first" {
		t.Errorf("left = %+v", left)
	}
}
