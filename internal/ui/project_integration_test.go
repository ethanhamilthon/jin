package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/core"
)

func TestSessionsRenderAndSendInTheirOwnProjects(t *testing.T) {
	a := startingApp(t)
	other := t.TempDir()
	writeFileT(t, filepath.Join(a.dir, "AGENTS.md"), "PROJECT_A_ONLY")
	writeFileT(t, filepath.Join(other, "AGENTS.md"), "PROJECT_B_ONLY")
	writeFileT(t, filepath.Join(other, "note.txt"), "attached")
	one := a.startSessionAt(a.dir, "one", "", "m", "", nil, nil)
	two := a.startSessionAt(other, "two", "", "m", "", nil, nil)
	a.active = one
	pump(t, a, one)
	pump(t, a, two)
	for _, check := range []struct {
		s            *chatSession
		want, absent string
	}{{one, "PROJECT_A_ONLY", "PROJECT_B_ONLY"}, {two, "PROJECT_B_ONLY", "PROJECT_A_ONLY"}} {
		text := check.s.agent.SystemPrompt()
		if !strings.Contains(text, check.want) || strings.Contains(text, check.absent) {
			t.Fatalf("wrong project context for %s", check.s.id)
		}
	}
	if !a.sendSessionDraft(two, "read @note.txt") {
		t.Fatal("targeted send was refused")
	}
	if a.active != one || len(one.pending) != 0 || len(two.pending) != 1 {
		t.Fatal("targeted send changed focus or queued to the wrong session")
	}
	if !strings.Contains(two.pending[0].Prompt, filepath.Join(other, "note.txt")) {
		t.Fatalf("wrong attachment path: %q", two.pending[0].Prompt)
	}
	record, found, err := a.store.GetSession(two.id)
	if err != nil || !found || record.Path != other {
		t.Fatalf("wrong saved project: %+v, %v", record, err)
	}
}

func TestOpeningSavedSessionDoesNotUseFocusedProject(t *testing.T) {
	a := startingApp(t)
	other := t.TempDir()
	writeFileT(t, filepath.Join(other, "AGENTS.md"), "RESUMED_PROJECT_ONLY")
	if err := a.store.Touch("saved", other, "m", "", "saved"); err != nil {
		t.Fatal(err)
	}
	record, _, _ := a.store.GetSession("saved")
	s, err := a.openSession(record)
	if err != nil {
		t.Fatal(err)
	}
	pump(t, a, s)
	if s.path != other || a.dir == other || !strings.Contains(s.agent.SystemPrompt(), "RESUMED_PROJECT_ONLY") {
		t.Fatal("opening a saved session leaked the focused project's directory")
	}
}

func TestUnavailableProjectKeepsHistoryAndRefusesRequests(t *testing.T) {
	a := startingApp(t)
	s := newStarting(t, a)
	pump(t, a, s)
	if err := os.Remove(a.dir); err != nil {
		t.Fatal(err)
	}
	if a.sendSessionDraft(s, "do not run elsewhere") || len(s.pending) != 0 {
		t.Fatal("a request escaped an unavailable project")
	}
	if !strings.Contains(lastText(s), "Project unavailable") {
		t.Fatalf("missing availability error: %q", lastText(s))
	}
	s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "history remains available"})
	if len(s.history) == 0 {
		t.Fatal("unavailable project lost its visible history")
	}
}
