//go:build unix

package core

import (
	"strings"
	"testing"

	"jin/internal/provider"
	"jin/internal/tasks"
)

func TestFinalAnswerWithRunningTasksGetsOneNote(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	agent, fake := newFakeAgent(t, "first answer", "keeping it", "never asked")
	m := tasks.New()
	t.Cleanup(m.StopAll)
	if _, err := m.Start("s1", "sleep 30", t.TempDir(), false); err != nil {
		t.Fatal(err)
	}
	agent.SetBackground(m, "s1", false)
	history := []provider.Message{{Role: "system", Content: "sys"}}
	agent.turn(t.Context(), Request{Prompt: "go", Model: "m"}, &history, nil, make(chan Update, 256))
	if fake.count() != 2 {
		t.Fatalf("requests = %d, want 2", fake.count())
	}
	second := fake.sent(1)
	note := second[len(second)-1]
	if note.Role != "user" || !IsTasksNote(note.Content) || !strings.Contains(note.Content, "sleep 30") || !strings.Contains(note.Content, "keep running") {
		t.Fatalf("note = %+v", note)
	}
	if last := history[len(history)-1]; last.Content != "keeping it" {
		t.Fatalf("last message = %+v", last)
	}
}

func TestNoNoteWithoutRunningTasks(t *testing.T) {
	agent, fake := newFakeAgent(t, "done")
	agent.SetBackground(tasks.New(), "s1", false)
	history := []provider.Message{{Role: "system", Content: "sys"}}
	agent.turn(t.Context(), Request{Prompt: "go", Model: "m"}, &history, nil, make(chan Update, 256))
	if fake.count() != 1 {
		t.Fatalf("requests = %d, want 1", fake.count())
	}
}

func TestHeadlessNoteSaysTasksStop(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	m := tasks.New()
	t.Cleanup(m.StopAll)
	_, _ = m.Start("run", "sleep 30", t.TempDir(), false)
	agent, _ := newFakeAgent(t, "x")
	agent.SetBackground(m, "run", true)
	note, ok := agent.tasksNote()
	if !ok || !strings.Contains(note, "stopped when you finish") {
		t.Fatalf("note = %q", note)
	}
}
