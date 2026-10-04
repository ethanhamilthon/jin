package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func pumpReload(t *testing.T, a *app, s *chatSession) {
	t.Helper()
	deadline := time.After(10 * time.Second)
	for s.render != nil {
		select {
		case ev := <-a.rendered:
			a.receiveRender(ev)
		case <-deadline:
			t.Fatal("reload did not finish")
		}
	}
}

func TestReloadRefreshesInstructionsAndFreshDisabledPrompts(t *testing.T) {
	a := startingApp(t)
	instructions := filepath.Join(a.dir, "AGENTS.md")
	writeFileT(t, instructions, "BEFORE_RELOAD")
	writeFileT(t, filepath.Join(os.Getenv("HOME"), ".jin-dev", "prompts", "fresh.md"), "fresh body")
	s := newStarting(t, a)
	pump(t, a, s)
	if s.promptBodies["fresh"] != "fresh body" {
		t.Fatal("initial prompt was not loaded")
	}
	writeFileT(t, instructions, "AFTER_RELOAD")
	if err := a.store.SavePromptsDisabled([]string{"fresh"}); err != nil {
		t.Fatal(err)
	}
	a.reloadSession()
	pumpReload(t, a, s)
	text := s.agent.SystemPrompt()
	if !strings.Contains(text, "AFTER_RELOAD") || strings.Contains(text, "BEFORE_RELOAD") {
		t.Fatal("reload did not replace project instructions")
	}
	if _, found := s.promptBodies["fresh"]; found {
		t.Fatal("reload ignored fresh disabled settings")
	}
	if !s.ready || !strings.Contains(lastText(s), "Reloaded system") {
		t.Fatalf("reload did not preserve ready state: %q", lastText(s))
	}
}

func TestReloadCompletionStaysWithItsOriginAfterFocusChanges(t *testing.T) {
	a := startingApp(t)
	s := newStarting(t, a)
	pump(t, a, s)
	a.reloadSession()
	other := &chatSession{id: "other", ready: true, input: clusters("untouched")}
	a.active = other
	pumpReload(t, a, s)
	if a.active != other || draftPayload(other.input) != "untouched" || len(other.history) != 0 {
		t.Fatal("reload completion changed another session")
	}
	if !strings.Contains(lastText(s), "Reloaded system") {
		t.Fatal("origin did not receive reload completion")
	}
}

func TestReloadRefusesBusyAndReadOnlySessions(t *testing.T) {
	a := startingApp(t)
	s := newStarting(t, a)
	pump(t, a, s)
	for _, busy := range []bool{true, false} {
		s.working, s.readOnlyPID = busy, 0
		if !busy {
			s.readOnlyPID = 123
		}
		a.reloadSession()
		if s.render != nil {
			t.Fatal("busy/read-only session started a reload")
		}
	}
}
