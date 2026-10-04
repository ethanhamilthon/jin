package store

import (
	"errors"
	"os"
	"testing"
	"time"

	"jin/internal/provider"
)

func TestFinishAsyncTaskWithEventIsAtomic(t *testing.T) {
	db := openTest(t)
	if err := db.AddAsyncTask(AsyncTask{ID: "t1", SessionID: "s1", Path: "/p", Command: "x", StartedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	won, err := db.FinishAsyncTaskWithEvent("t1", AsyncDone, 0, "s1", "/p", "result")
	if err != nil || !won {
		t.Fatalf("finish = %v, %v", won, err)
	}
	if won, _ := db.FinishAsyncTaskWithEvent("t1", AsyncStopped, 143, "s1", "/p", "again"); won {
		t.Error("a finished task must not finish twice")
	}
	events, _ := db.ClaimAsyncEvents("/p", os.Getpid())
	if len(events) != 1 || events[0].Text != "result" {
		t.Errorf("events = %v", events)
	}
	if task, _, _ := db.AsyncTask("t1"); task.Status != AsyncDone {
		t.Errorf("status = %q", task.Status)
	}
}

func TestDeliverAsyncEventSavesOnce(t *testing.T) {
	db := openTest(t)
	if err := db.AddAsyncEvent("s1", "/p", "result"); err != nil {
		t.Fatal(err)
	}
	events, _ := db.ClaimAsyncEvents("/p", os.Getpid())
	msg := provider.Message{Role: "user", Content: "result"}
	for i, want := range []bool{true, false} {
		got, err := db.DeliverAsyncEvent(events[0].ID, "s1", msg)
		if err != nil || got != want {
			t.Fatalf("delivery %d = %v, %v", i, got, err)
		}
	}
	if msgs, _ := db.LoadMessages("s1"); len(msgs) != 1 {
		t.Errorf("messages = %v", msgs)
	}
}

func TestRetryBusy(t *testing.T) {
	cases := []struct {
		name  string
		errs  []error
		calls int
		fails bool
	}{
		{"ok", []error{nil}, 1, false},
		{"busy then ok", []error{errors.New("database is locked (5) (SQLITE_BUSY)"), nil}, 2, false},
		{"other error", []error{errors.New("no such table")}, 1, true},
	}
	for _, c := range cases {
		calls := 0
		err := retryBusy(func() error { calls++; return c.errs[calls-1] })
		if calls != c.calls || (err != nil) != c.fails {
			t.Errorf("%s: calls %d err %v", c.name, calls, err)
		}
	}
}
