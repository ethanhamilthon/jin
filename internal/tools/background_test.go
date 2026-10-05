//go:build unix

package tools

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"jin/internal/tasks"
)

func backgroundCtx(t *testing.T, detach <-chan struct{}) (context.Context, *tasks.Manager) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	m := tasks.New()
	t.Cleanup(m.StopAll)
	return WithBackground(context.Background(), Background{Tasks: m, Owner: "s1", Detach: detach}), m
}

func adoptedTask(t *testing.T, m *tasks.Manager) tasks.Info {
	t.Helper()
	running := m.Running("s1")
	if len(running) != 1 {
		t.Fatalf("running tasks = %+v", running)
	}
	return running[0]
}

func TestBashTimeoutMovesToTheBackground(t *testing.T) {
	ctx, m := backgroundCtx(t, nil)
	start := time.Now()
	out := runBash(ctx, "echo started; sleep 30", time.Second, "", true)
	if time.Since(start) > 5*time.Second {
		t.Fatalf("took %v", time.Since(start))
	}
	info := adoptedTask(t, m)
	for _, want := range []string{"started", "timed out", "background task " + info.ID, "task tool"} {
		if !strings.Contains(out, want) {
			t.Errorf("result misses %q:\n%s", want, out)
		}
	}
	if info.Command != "echo started; sleep 30" {
		t.Fatalf("task = %+v", info)
	}
	if _, err := os.Stat(info.Log); err != nil {
		t.Errorf("the log must stay for the task: %v", err)
	}
}

func TestBashDetachMovesToTheBackgroundAtOnce(t *testing.T) {
	detach := make(chan struct{})
	ctx, m := backgroundCtx(t, detach)
	go func() { time.Sleep(300 * time.Millisecond); close(detach) }()
	start := time.Now()
	out := runBash(ctx, "sleep 30", 60*time.Second, "", true)
	if time.Since(start) > 5*time.Second {
		t.Fatalf("detach did not end the wait: %v", time.Since(start))
	}
	adoptedTask(t, m)
	if !strings.Contains(out, "user's request") {
		t.Errorf("result = %s", out)
	}
}

func TestAdoptedTaskReportsItsExit(t *testing.T) {
	ctx, m := backgroundCtx(t, nil)
	runBash(ctx, "sleep 1.2; exit 7", 500*time.Millisecond, "", true)
	select {
	case ev := <-m.Events():
		if ev.Task.Exit != 7 || ev.Task.Status != tasks.Failed {
			t.Fatalf("event = %+v", ev)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("no event for the adopted task")
	}
}

func TestBashTimeoutKillsWhenItMayNotAdopt(t *testing.T) {
	ctx, m := backgroundCtx(t, nil)
	out := runBash(ctx, "echo before; sleep 30", time.Second, "", false)
	if !strings.Contains(out, "before") || !strings.Contains(out, "[command timed out]") {
		t.Errorf("result = %s", out)
	}
	if len(m.List("")) != 0 {
		t.Error("a headless bash must not adopt")
	}
}

func TestBashLogIsRemovedAfterANormalFinish(t *testing.T) {
	ctx, _ := backgroundCtx(t, nil)
	if out := runBash(ctx, "echo hi", 5*time.Second, "", true); !strings.Contains(out, "hi") {
		t.Fatalf("result = %s", out)
	}
	entries, _ := os.ReadDir(os.Getenv("HOME") + "/.jin-dev/tasks")
	if len(entries) != 0 {
		t.Errorf("files left behind: %v", entries)
	}
}
