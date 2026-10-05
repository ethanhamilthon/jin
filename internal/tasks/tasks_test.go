package tasks

import (
	"strings"
	"testing"
	"time"
)

func waitEvent(t *testing.T, m *Manager) Event {
	t.Helper()
	select {
	case ev := <-m.Events():
		return ev
	case <-time.After(5 * time.Second):
		t.Fatal("no event")
		return Event{}
	}
}

func TestTaskEndsWithEvent(t *testing.T) {
	m := New()
	info, err := m.Start("s1", "echo hello; exit 3", t.TempDir(), false)
	if err != nil {
		t.Fatal(err)
	}
	ev := waitEvent(t, m)
	if ev.Owner != "s1" || ev.Task.ID != info.ID || ev.Task.Status != Failed || ev.Task.Exit != 3 {
		t.Fatalf("event %+v", ev)
	}
	if !strings.Contains(ev.Text, `status="failed" exit="3"`) || !strings.Contains(ev.Text, "hello") {
		t.Fatalf("text %q", ev.Text)
	}
	if len(m.Running("s1")) != 0 || len(m.List("s1")) != 1 {
		t.Fatal("task still listed as running")
	}
}

func TestInputAndCheck(t *testing.T) {
	m := New()
	info, err := m.Start("s1", "read line; echo got:$line", t.TempDir(), true)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Input("s2", info.ID, "x\n"); err == nil {
		t.Fatal("another owner wrote to the task")
	}
	if err := m.Input("s1", info.ID, "ping\n"); err != nil {
		t.Fatal(err)
	}
	waitEvent(t, m)
	_, out, err := m.Check("s1", info.ID, 0)
	if err != nil || !strings.Contains(out, "got:ping") {
		t.Fatalf("out %q err %v", out, err)
	}
}

func TestStopByAgentSendsNoEvent(t *testing.T) {
	m := New()
	info, _ := m.Start("s1", "sleep 30", t.TempDir(), false)
	if len(m.Running("s1")) != 1 || m.Counts()["s1"] != 1 {
		t.Fatal("task not running")
	}
	if err := m.Stop("s1", info.ID, ByAgent); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for len(m.Running("s1")) > 0 && time.Now().Before(deadline) {
		time.Sleep(20 * time.Millisecond)
	}
	if got := m.List("s1")[0].Status; got != Stopped {
		t.Fatalf("status %s", got)
	}
	select {
	case ev := <-m.Events():
		t.Fatalf("unexpected event %+v", ev)
	case <-time.After(200 * time.Millisecond):
	}
}

func TestStopByUserSendsEvent(t *testing.T) {
	m := New()
	info, _ := m.Start("s1", "sleep 30", t.TempDir(), false)
	if err := m.StopAny(info.ID); err != nil {
		t.Fatal(err)
	}
	if ev := waitEvent(t, m); ev.Task.Status != Stopped || !strings.Contains(ev.Text, "stopped manually by the user") {
		t.Fatalf("event %+v", ev)
	}
}

func TestSummary(t *testing.T) {
	got, ok := Summary(ResultText("ab12", Done, 0, "a\nb"))
	if !ok || got != "background task ab12 done (exit 0)\na\nb" {
		t.Fatalf("got %q", got)
	}
}
