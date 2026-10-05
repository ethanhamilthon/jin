package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/tasks"
)

func persistedApp(t *testing.T) (*app, *chatSession) {
	a := startingApp(t)
	if err := a.store.TouchProvider("s1", a.dir, "m", "", "title", ""); err != nil {
		t.Fatal(err)
	}
	rec, _, _ := a.store.GetSession("s1")
	s, err := a.openSession(rec)
	if err != nil {
		t.Fatal(err)
	}
	return a, s
}

func TestTaskResultIsNotExpandedAndShowsAsSummary(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".jin-dev", "prompts")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "style.md"), []byte("Be brief."), 0o600); err != nil {
		t.Fatal(err)
	}
	s := &chatSession{width: 80, model: "m"}
	result := tasks.ResultText("a1", tasks.Done, 0, "use #plan and #style now")
	s.sendTaskResult(result)
	if len(s.pending) != 1 || s.pending[0].Prompt != result {
		t.Fatalf("the result must reach the agent unchanged: %+v", s.pending)
	}
	last := s.history[len(s.history)-1]
	if last.tool != taskEntry || last.kind == core.UpdateUser || !strings.HasPrefix(last.text, "background task a1 done (exit 0)") {
		t.Errorf("entry = %+v, want a task summary and not a user bubble", last)
	}
}

func TestHistoryShowsTaskResultsAsSummaries(t *testing.T) {
	text := tasks.ResultText("b2", tasks.Failed, 3, "boom")
	entries := historyToEntries([]provider.Message{{Role: "user", Content: text}}, nil)
	if len(entries) != 1 || entries[0].tool != taskEntry || !strings.Contains(entries[0].text, "b2 failed") {
		t.Errorf("entries = %+v", entries)
	}
}

func TestBackgroundLoaderYieldsToARunningAgent(t *testing.T) {
	a := &app{tasksRunning: map[string]int{"s1": 2}}
	s := &chatSession{id: "s1"}
	if !a.backgroundWaiting(s) {
		t.Error("an idle session with background tasks must show the loader")
	}
	s.working = true
	if a.backgroundWaiting(s) {
		t.Error("a running agent has priority over the background loader")
	}
}

func TestTaskResultWaitsForAReadOnlySession(t *testing.T) {
	a, s := persistedApp(t)
	s.readOnlyPID = 42
	event := tasks.Event{Owner: "s1", Text: tasks.ResultText("c3", tasks.Done, 0, "ok")}
	a.receiveTask(event)
	if len(s.pending) != 0 || len(a.heldTasks) != 1 {
		t.Fatalf("pending %d held %d", len(s.pending), len(a.heldTasks))
	}
	s.readOnlyPID = 0
	a.retryTasks("s1")
	if len(a.heldTasks) != 0 || len(s.pending) != 1 {
		t.Fatalf("after retry: pending %d held %d", len(s.pending), len(a.heldTasks))
	}
}
