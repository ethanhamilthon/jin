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
	out, err := shellCommand(context.Background(), t.TempDir(), "echo $GIT_TERMINAL_PROMPT $GIT_EDITOR $PAGER $DEBIAN_FRONTEND; cat /dev/tty").CombinedOutput()
	if err == nil || !strings.Contains(string(out), "0 true cat noninteractive") {
		t.Fatalf("out = %q, err = %v", out, err)
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
