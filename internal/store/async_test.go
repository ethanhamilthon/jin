package store

import (
	"os"
	"testing"
	"time"
)

func openTest(t *testing.T) *DB {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	db, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func TestAsyncTaskLifecycle(t *testing.T) {
	db := openTest(t)
	task := AsyncTask{ID: "t1", SessionID: "s1", Path: "/p", Command: "sleep 1", PID: 10, PGID: 10, StartedAt: time.Now()}
	if err := db.AddAsyncTask(task); err != nil {
		t.Fatal(err)
	}
	running, err := db.RunningAsyncTasks("/p")
	if err != nil || len(running) != 1 || running[0].ID != "t1" {
		t.Fatalf("running = %v, %v", running, err)
	}
	if other, _ := db.RunningAsyncTasks("/other"); len(other) != 0 {
		t.Errorf("other path sees %v", other)
	}
	if ok, err := db.FinishAsyncTask("t1", AsyncStopped, 143); err != nil || !ok {
		t.Fatalf("finish = %v, %v", ok, err)
	}
	if ok, _ := db.FinishAsyncTask("t1", AsyncDone, 0); ok {
		t.Error("a finished task must keep its state")
	}
	got, found, err := db.AsyncTask("t1")
	if err != nil || !found || got.Status != AsyncStopped || got.ExitCode != 143 || got.FinishedAt.IsZero() {
		t.Errorf("task = %+v, %v, %v", got, found, err)
	}
	if running, _ := db.RunningAsyncTasks(""); len(running) != 0 {
		t.Errorf("still running: %v", running)
	}
}

func TestAsyncEventsClaimAckAndRelease(t *testing.T) {
	db := openTest(t)
	for _, text := range []string{"one", "two"} {
		if err := db.AddAsyncEvent("s1", "/p", text); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.AddAsyncEvent("s2", "/other", "elsewhere"); err != nil {
		t.Fatal(err)
	}
	events, err := db.ClaimAsyncEvents("/p", os.Getpid())
	if err != nil || len(events) != 2 || events[0].Text != "one" || events[1].Text != "two" {
		t.Fatalf("claim = %v, %v", events, err)
	}
	if again, _ := db.ClaimAsyncEvents("/p", os.Getpid()); len(again) != 0 {
		t.Errorf("claimed twice: %v", again)
	}
	// A claim by a dead process is released and can be claimed again.
	if _, err := db.sql.Exec(`UPDATE async_events SET claimed_by = 999999999 WHERE text = 'two'`); err != nil {
		t.Fatal(err)
	}
	if err := db.ReleaseDeadAsyncClaims(); err != nil {
		t.Fatal(err)
	}
	again, _ := db.ClaimAsyncEvents("/p", os.Getpid())
	if len(again) != 1 || again[0].Text != "two" {
		t.Fatalf("after release = %v", again)
	}
	if err := db.AckAsyncEvents(events[0].ID, again[0].ID); err != nil {
		t.Fatal(err)
	}
	if left, _ := db.ClaimAsyncEvents("/other", os.Getpid()); len(left) != 1 {
		t.Errorf("other path events = %v", left)
	}
}

func TestAdoptedTaskKeepsItsExitPath(t *testing.T) {
	db := openTest(t)
	task := AsyncTask{ID: "b1", SessionID: "s", Path: "/p", Command: "make", PID: 5, PGID: 5, LogPath: "/l", ExitPath: "/e", StartedAt: time.Now()}
	if err := db.AddAsyncTask(task); err != nil {
		t.Fatal(err)
	}
	got, found, err := db.AsyncTask("b1")
	if err != nil || !found || got.ExitPath != "/e" {
		t.Fatalf("task = %+v, %v, %v", got, found, err)
	}
	if ok, err := db.FinishAsyncTask("b1", AsyncEnded, 0); err != nil || !ok {
		t.Fatalf("finish = %v, %v", ok, err)
	}
	if got, _, _ := db.AsyncTask("b1"); got.Status != AsyncEnded {
		t.Errorf("status = %q", got.Status)
	}
}
