package store

import (
	"errors"
	"os"
	"os/exec"
	"testing"
)

func deadPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command("true")
	if err := cmd.Run(); err != nil {
		t.Fatal(err)
	}
	return cmd.Process.Pid
}

func TestSetRunningClaimsTheSession(t *testing.T) {
	cases := []struct {
		name  string
		owner int
		busy  bool
	}{
		{"free", 0, false},
		{"same process", os.Getpid(), false},
		{"dead owner", -1, false},
		{"live owner", os.Getppid(), true},
	}
	for _, c := range cases {
		db := openTest(t)
		owner := c.owner
		if owner == -1 {
			owner = deadPID(t)
		}
		if owner != 0 {
			db.sql.Exec(`INSERT INTO running_sessions(session_id, pid) VALUES ('s1', ?)`, owner)
		}
		err := db.SetRunning("s1", true)
		var busy ErrSessionBusy
		if errors.As(err, &busy) != c.busy || (c.busy && busy.PID != owner) || (!c.busy && err != nil) {
			t.Errorf("%s: err = %v", c.name, err)
		}
		pid, alive, err := db.SessionOwner("s1")
		want := os.Getpid()
		if c.busy {
			want = owner
		}
		if err != nil || pid != want || !alive {
			t.Errorf("%s: owner = %d %v %v", c.name, pid, alive, err)
		}
		if err := db.SetRunning("s1", false); err != nil {
			t.Fatal(err)
		}
		if pid, _, _ := db.SessionOwner("s1"); c.busy && pid != owner || !c.busy && pid != 0 {
			t.Errorf("%s: release by %d left owner %d", c.name, os.Getpid(), pid)
		}
	}
}

func TestClaimAsyncEventsSkipsSessionsOfOtherLiveProcesses(t *testing.T) {
	db := openTest(t)
	db.sql.Exec(`INSERT INTO running_sessions(session_id, pid) VALUES ('busy', ?), ('gone', ?)`, os.Getppid(), deadPID(t))
	for _, session := range []string{"busy", "gone", "free"} {
		if err := db.AddAsyncEvent(session, "/p", session); err != nil {
			t.Fatal(err)
		}
	}
	events, err := db.ClaimAsyncEvents("/p", os.Getpid())
	if err != nil || len(events) != 2 || events[0].Text != "gone" || events[1].Text != "free" {
		t.Fatalf("claim = %v, %v", events, err)
	}
	db.sql.Exec(`DELETE FROM running_sessions WHERE session_id = 'busy'`)
	if events, _ := db.ClaimAsyncEvents("/p", os.Getpid()); len(events) != 1 || events[0].Text != "busy" {
		t.Errorf("after release = %v", events)
	}
}
