package store

import (
	"os"
	"os/exec"
	"sync"
	"testing"
)

func openTwo(t *testing.T) (*DB, *DB) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	first, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	second, err := Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { first.Close(); second.Close() })
	return first, second
}

func TestConcurrentWritersDoNotFail(t *testing.T) {
	first, second := openTwo(t)
	var wg sync.WaitGroup
	errs := make(chan error, 400)
	for _, db := range []*DB{first, second} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 100; i++ {
				if err := db.SetRunning("x", i%2 == 0); err != nil {
					errs <- err
				}
				if err := db.SaveModel("m", "high"); err != nil {
					errs <- err
				}
			}
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("a parallel write failed: %v", err)
	}
}

func TestRecoverInterruptedSkipsLiveProcesses(t *testing.T) {
	first, second := openTwo(t)
	for _, id := range []string{"mine", "dead"} {
		if err := first.Touch(id, "/p", "m", "", id); err != nil {
			t.Fatal(err)
		}
	}
	if err := first.SetRunning("mine", true); err != nil {
		t.Fatal(err)
	}
	done := exec.Command("true")
	if err := done.Run(); err != nil {
		t.Fatal(err)
	}
	if _, err := first.sql.Exec(`INSERT INTO running_sessions(session_id, pid) VALUES ('dead', ?)`, done.Process.Pid); err != nil {
		t.Fatal(err)
	}
	if err := second.RecoverInterrupted(); err != nil {
		t.Fatal(err)
	}
	unread, err := second.UnreadSessions("/p")
	if err != nil {
		t.Fatal(err)
	}
	if !unread["dead"] || unread["mine"] {
		t.Errorf("unread = %v, want only the session of the dead process", unread)
	}
	var left int
	if err := second.sql.QueryRow(`SELECT COUNT(*) FROM running_sessions WHERE pid = ?`, os.Getpid()).Scan(&left); err != nil || left != 1 {
		t.Errorf("a live process lost its running marker: %d %v", left, err)
	}
}
