package core

import (
	"context"
	"strings"
	"testing"
)

// runSteps runs the requests one after another and waits for each to end.
func runSteps(t *testing.T, agent *Agent, requests ...Request) {
	t.Helper()
	prompts, updates := make(chan Request), make(chan Update, 64)
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go agent.Run(ctx, nil, prompts, updates)
	for _, request := range requests {
		prompts <- request
		for update := range updates {
			if update.Kind == UpdateDone {
				break
			}
		}
	}
}

func lastUser(f *fakeProvider) (system, user string) {
	last := f.sent(f.count() - 1)
	return last[0].Content, last[len(last)-1].Content
}

func newRefreshRun(t *testing.T, replies ...string) (*Agent, *fakeProvider) {
	agent, fake := newFakeAgent(t, replies...)
	clock := &fakeClock{now: noon}
	agent.clock = clock.read
	agent.SetSystemPrompt("sys")
	agent.SetRefresher(func(context.Context) string { return "sys v2" })
	return agent, fake
}

func TestRefreshAfterExplicitCompactionThenStaysQuiet(t *testing.T) {
	agent, fake := newRefreshRun(t, "answer one", "summary", "answer two", "answer three")
	runSteps(t, agent, Request{Prompt: "one", Model: "m"}, Request{Kind: RequestCompact, Model: "m"}, Request{Prompt: "two", Model: "m"})
	if system, user := lastUser(fake); system != "sys v2" || !strings.HasPrefix(user, refreshedNote) {
		t.Fatalf("after compaction: system %q, last %q", system, user)
	}
	runSteps(t, agent, Request{Prompt: "three", Model: "m"})
	if _, user := lastUser(fake); strings.Contains(user, "system-refreshed") {
		t.Errorf("a hot cache must not refresh: %q", user)
	}
}

func TestRefreshAfterAutoCompactionInTheSameTurn(t *testing.T) {
	agent, fake := newRefreshRun(t, "answer one", "summary", "answer two")
	runSteps(t, agent, Request{Prompt: "one", Model: "m"}, Request{Prompt: "two", Model: "m", Window: 15})
	if system, user := lastUser(fake); system != "sys v2" || !strings.HasPrefix(user, refreshedNote+"two") {
		t.Fatalf("after auto-compaction: system %q, last %q", system, user)
	}
}

func TestRefreshAfterOverflowRecovery(t *testing.T) {
	agent, fake := newRefreshRun(t, "answer one", "overflow", "summary", "answer two", "answer three")
	runSteps(t, agent, Request{Prompt: "one", Model: "m"}, Request{Prompt: "two", Model: "m"})
	if retry := fake.sent(3); retry[0].Content != "sys v2" {
		t.Fatalf("the retry after overflow kept the old system: %q", retry[0].Content)
	}
	runSteps(t, agent, Request{Prompt: "three", Model: "m"})
	if _, user := lastUser(fake); !strings.HasPrefix(user, refreshedNote+"three") {
		t.Errorf("the note must reach the next user message: %q", user)
	}
}

func TestRefreshedNoteIsStrippedFromHistory(t *testing.T) {
	got := StripNotes(refreshedNote + "hello")
	if got != "hello" {
		t.Errorf("got %q", got)
	}
	if got := StripNotes(todoEditedOpen + "x" + todoEditedClose + "\n\n" + refreshedNote + "hi"); got != "hi" {
		t.Errorf("got %q", got)
	}
}
