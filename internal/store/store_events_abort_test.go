package store

import (
	"os"
	"testing"
	"time"

	"jin/internal/provider"
)

func abortInserts(t *testing.T, db *DB, table string) func() {
	t.Helper()
	trigger := "abort_" + table
	if _, err := db.sql.Exec(`CREATE TRIGGER ` + trigger + ` BEFORE INSERT ON ` + table +
		` BEGIN SELECT RAISE(ABORT, 'forced failure'); END`); err != nil {
		t.Fatal(err)
	}
	return func() {
		if _, err := db.sql.Exec(`DROP TRIGGER ` + trigger); err != nil {
			t.Fatal(err)
		}
	}
}

func TestFinishAsyncTaskWithEventRollsBackWhenTheEventFails(t *testing.T) {
	db := openTest(t)
	if err := db.AddAsyncTask(AsyncTask{ID: "t1", SessionID: "s1", Path: "/p", Command: "x", StartedAt: time.Now()}); err != nil {
		t.Fatal(err)
	}
	allow := abortInserts(t, db, "async_events")
	if won, err := db.FinishAsyncTaskWithEvent("t1", AsyncDone, 0, "s1", "/p", "result"); err == nil || won {
		t.Fatalf("finish with a failing event = %v, %v", won, err)
	}
	if task, _, _ := db.AsyncTask("t1"); task.Status != AsyncRunning {
		t.Errorf("status = %q, want running", task.Status)
	}
	if events, _ := db.ClaimAsyncEvents("/p", os.Getpid()); len(events) != 0 {
		t.Errorf("events = %v", events)
	}
	allow()
	if won, err := db.FinishAsyncTaskWithEvent("t1", AsyncDone, 0, "s1", "/p", "result"); err != nil || !won {
		t.Fatalf("finish after the fix = %v, %v", won, err)
	}
}

func TestDeliverAsyncEventKeepsTheEventWhenSavingFails(t *testing.T) {
	db := openTest(t)
	if err := db.AddAsyncEvent("s1", "/p", "result"); err != nil {
		t.Fatal(err)
	}
	events, _ := db.ClaimAsyncEvents("/p", os.Getpid())
	msg := provider.Message{Role: "user", Content: "result"}
	allow := abortInserts(t, db, "messages")
	if got, err := db.DeliverAsyncEvent(events[0].ID, "s1", msg); err == nil || got {
		t.Fatalf("delivery with a failing save = %v, %v", got, err)
	}
	var left int
	_ = db.sql.QueryRow(`SELECT COUNT(*) FROM async_events WHERE id = ?`, events[0].ID).Scan(&left)
	if left != 1 {
		t.Errorf("event rows = %d, want 1", left)
	}
	allow()
	if got, err := db.DeliverAsyncEvent(events[0].ID, "s1", msg); err != nil || !got {
		t.Fatalf("delivery after the fix = %v, %v", got, err)
	}
	if msgs, _ := db.LoadMessages("s1"); len(msgs) != 1 {
		t.Errorf("messages = %v", msgs)
	}
}
