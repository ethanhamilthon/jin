//go:build unix

package tools

import (
	"context"
	"errors"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"
)

type adopted struct {
	got Adoption
	n   int
}

func backgroundCtx(t *testing.T, detach <-chan struct{}, fail error) (context.Context, *adopted) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	seen := &adopted{}
	ctx := WithBackground(context.Background(), Background{
		Detach: detach,
		Adopt: func(a Adoption) (string, error) {
			seen.got, seen.n = a, seen.n+1
			if fail != nil {
				return "", fail
			}
			return "task42", nil
		},
	})
	return ctx, seen
}

func alive(pid int) bool { return syscall.Kill(pid, 0) == nil }

// reap ends the process group a test left running and waits until it is gone,
// so that nothing writes into the temporary home while it is removed.
func reap(t *testing.T, pgid int) {
	t.Helper()
	_ = syscall.Kill(-pgid, syscall.SIGKILL)
	for i := 0; i < 100 && syscall.Kill(-pgid, 0) == nil; i++ {
		time.Sleep(20 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)
}

func TestBashTimeoutMovesToTheBackground(t *testing.T) {
	ctx, seen := backgroundCtx(t, nil, nil)
	start := time.Now()
	out := runBash(ctx, "echo started; sleep 30", time.Second)
	if time.Since(start) > 5*time.Second {
		t.Fatalf("took %v", time.Since(start))
	}
	for _, want := range []string{"started", "task42", "timed out", "jin async check --id task42", "jin async stop --id task42"} {
		if !strings.Contains(out, want) {
			t.Errorf("result misses %q:\n%s", want, out)
		}
	}
	if seen.n != 1 || seen.got.PID <= 1 || seen.got.Command != "echo started; sleep 30" {
		t.Fatalf("adoption = %+v", seen)
	}
	defer func() { reap(t, seen.got.PGID) }()
	if !alive(seen.got.PID) {
		t.Error("the command must keep running after the timeout")
	}
	if _, err := os.Stat(seen.got.Log); err != nil {
		t.Errorf("the log must stay for the daemon: %v", err)
	}
}

func TestBashDetachMovesToTheBackgroundAtOnce(t *testing.T) {
	detach := make(chan struct{})
	ctx, seen := backgroundCtx(t, detach, nil)
	go func() { time.Sleep(300 * time.Millisecond); close(detach) }()
	start := time.Now()
	out := runBash(ctx, "sleep 30", 60*time.Second)
	if time.Since(start) > 5*time.Second {
		t.Fatalf("detach did not end the wait: %v", time.Since(start))
	}
	defer func() { reap(t, seen.got.PGID) }()
	if !strings.Contains(out, "task42") || !strings.Contains(out, "user's request") {
		t.Errorf("result = %s", out)
	}
	if !alive(seen.got.PID) {
		t.Error("the command must keep running")
	}
}

func TestBashExitCodeIsWrittenAfterTheProcessEnds(t *testing.T) {
	ctx, seen := backgroundCtx(t, nil, nil)
	runBash(ctx, "sleep 1.2; exit 7", 500*time.Millisecond)
	defer func() { reap(t, seen.got.PGID) }()
	deadline := time.Now().Add(5 * time.Second)
	for {
		if data, err := os.ReadFile(seen.got.Exit); err == nil {
			if strings.TrimSpace(string(data)) != "7" {
				t.Fatalf("exit file = %q", data)
			}
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("the exit file never appeared")
		}
		time.Sleep(50 * time.Millisecond)
	}
}

func TestBashTimeoutKillsWhenNothingCanAdopt(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	out := runBash(context.Background(), "echo before; sleep 30", time.Second)
	if !strings.Contains(out, "before") || !strings.Contains(out, "[command timed out]") {
		t.Errorf("result = %s", out)
	}
}

func TestBashKillsWhenAdoptionFails(t *testing.T) {
	ctx, seen := backgroundCtx(t, nil, errors.New("no daemon"))
	out := runBash(ctx, "sleep 30", time.Second)
	if !strings.Contains(out, "no daemon") || !strings.Contains(out, "was stopped") {
		t.Errorf("result = %s", out)
	}
	time.Sleep(1500 * time.Millisecond)
	if alive(seen.got.PID) {
		t.Error("a command that could not be adopted must not be left running")
	}
	if _, err := os.Stat(seen.got.Log); err == nil {
		t.Error("the log of a stopped command must be removed")
	}
}

func TestBashLogIsRemovedAfterANormalFinish(t *testing.T) {
	ctx, _ := backgroundCtx(t, nil, nil)
	if out := runBash(ctx, "echo hi", 5*time.Second); !strings.Contains(out, "hi") {
		t.Fatalf("result = %s", out)
	}
	dir := os.Getenv("HOME") + "/.jin-dev/async"
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Errorf("files left behind: %v", entries)
	}
}
