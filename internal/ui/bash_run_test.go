package ui

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestShellCommandIsNonInteractive(t *testing.T) {
	dir := t.TempDir()
	cmd := shellCommand(context.Background(), dir, "echo $GIT_TERMINAL_PROMPT $GIT_EDITOR $PAGER $DEBIAN_FRONTEND; cat /dev/tty")
	if cmd.Dir != dir {
		t.Fatalf("command directory = %q, want %q", cmd.Dir, dir)
	}
	out, err := cmd.CombinedOutput()
	if err == nil || !strings.Contains(string(out), "0 true cat noninteractive") {
		t.Fatalf("out = %q, err = %v", out, err)
	}
}

func TestBashResultReturnsToOriginatingSession(t *testing.T) {
	a, _ := layoutApp(t)
	origin, other := &chatSession{id: "origin"}, &chatSession{id: "other"}
	a.sessions = map[string]*chatSession{"origin": origin, "other": other}
	a.receiveBash(bashResult{session: "origin", output: "result"})
	if len(origin.history) != 1 || origin.history[0].text != "result" || len(other.history) != 0 {
		t.Fatalf("origin=%+v other=%+v", origin.history, other.history)
	}
}

func TestShellCancelKillsChildProcess(t *testing.T) {
	dir := t.TempDir()
	pidFile := filepath.Join(dir, "pid")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		_, _ = shellCommand(ctx, dir, "sleep 60 & echo $! > pid; wait").CombinedOutput()
		close(done)
	}()
	var pid int
	for deadline := time.Now().Add(5 * time.Second); pid == 0; time.Sleep(20 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("child never started")
		}
		data, _ := os.ReadFile(pidFile)
		pid, _ = strconv.Atoi(strings.TrimSpace(string(data)))
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("shell did not end after cancel")
	}
	for deadline := time.Now().Add(3 * time.Second); syscall.Kill(pid, 0) == nil; time.Sleep(50 * time.Millisecond) {
		if time.Now().After(deadline) {
			t.Fatal("child process still alive")
		}
	}
}
