package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
)

func TestAsyncResultIsNotExpandedAndShowsAsSummary(t *testing.T) {
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
	result := "<async-task-result id=\"a1\" status=\"done\" exit=\"0\">\nuse #plan and #style now\n</async-task-result>"
	s.sendAsync(result)

	if len(s.pending) != 1 {
		t.Fatalf("pending = %d, want 1", len(s.pending))
	}
	if got := s.pending[0].Prompt; got != result {
		t.Errorf("the result must reach the agent unchanged, got:\n%s", got)
	}
	if strings.Contains(s.pending[0].Prompt, "<pasted-prompts>") {
		t.Error("#prompts inside a task result must not be expanded")
	}
	last := s.history[len(s.history)-1]
	if last.tool != asyncEntry || last.kind == core.UpdateUser {
		t.Errorf("entry = %+v, want an async summary and not a user bubble", last)
	}
	if !strings.HasPrefix(last.text, "async task a1 done (exit 0)") {
		t.Errorf("summary = %q", last.text)
	}
}

func TestHistoryShowsAsyncResultsAsSummaries(t *testing.T) {
	text := "<async-task-result id=\"b2\" status=\"failed\" exit=\"3\">\nboom\n</async-task-result>"
	entries := historyToEntries([]provider.Message{{Role: "user", Content: text}}, nil)
	if len(entries) != 1 || entries[0].tool != asyncEntry || !strings.Contains(entries[0].text, "b2 failed") {
		t.Errorf("entries = %+v", entries)
	}
}

func TestBackgroundLoaderYieldsToARunningAgent(t *testing.T) {
	a := &app{asyncRunning: map[string]int{"s1": 2}}
	s := &chatSession{id: "s1"}
	if _, ok := a.backgroundLoader(s); !ok {
		t.Error("an idle session with background tasks must show the loader")
	}
	s.working = true
	if _, ok := a.backgroundLoader(s); ok {
		t.Error("a running agent has priority over the background loader")
	}
	if _, ok := a.backgroundLoader(&chatSession{id: "other"}); ok {
		t.Error("a session without tasks must not show the loader")
	}
}
