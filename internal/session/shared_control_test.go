package session

import (
	"testing"

	"jin/internal/core"
	"jin/internal/tools"
)

func TestStopPausesQueueUntilResume(t *testing.T) {
	manager, events, dir := newTestManager(t)
	snap, err := manager.Create(dir)
	if err != nil {
		t.Fatal(err)
	}
	id := snap.State.ID
	waitFor(t, events, func(event Event) bool { return event.State != nil && event.State.Ready })
	if err := manager.Interrupt(id); err != nil {
		t.Fatal(err)
	}
	if err := manager.Send(id, "queued while stopped", nil, nil); err != nil {
		t.Fatal(err)
	}
	snap, _ = manager.Snapshot(id)
	if !snap.State.Paused || snap.State.Queued != 1 || snap.State.Working {
		t.Fatalf("paused state: %+v", snap.State)
	}
	if err := manager.Resume(id); err != nil {
		t.Fatal(err)
	}
	waitFor(t, events, func(event Event) bool { return event.Type == "ring" && event.Kind == core.UpdateDone })
	snap, _ = manager.Snapshot(id)
	if snap.State.Paused || snap.State.Queued != 0 {
		t.Fatalf("resumed state: %+v", snap.State)
	}
}

func TestStaleQuestionCannotAnswerNewQuestion(t *testing.T) {
	manager, events, dir := newTestManager(t)
	snap, _ := manager.Create(dir)
	waitFor(t, events, func(event Event) bool { return event.State != nil && event.State.Ready })
	_ = manager.Do(snap.State.ID, func(s *Session) error {
		s.ask = []tools.Question{{Question: "new question"}}
		s.questionRevision = 2
		return nil
	})
	if err := manager.AnswerQuestion(snap.State.ID, 1, []string{"stale"}); err == nil {
		t.Fatal("accepted stale answer")
	}
	snap, _ = manager.Snapshot(snap.State.ID)
	if len(snap.State.Ask) != 1 {
		t.Fatal("stale answer removed current question")
	}
}
