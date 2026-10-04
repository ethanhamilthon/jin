package core

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestReloadPromptsWhileIdleDoesNotCallModel(t *testing.T) {
	agent, fake := newFakeAgent(t, "answer")
	clock := &fakeClock{now: noon}
	agent.clock = clock.read
	_, prompts, updates, _ := startReloadRun(t, agent)
	agent.refreshDue = true
	stopReaders, readersDone := make(chan struct{}), make(chan struct{})
	go func() {
		defer close(readersDone)
		for {
			select {
			case <-stopReaders:
				return
			default:
				agent.SystemPrompt()
			}
		}
	}()
	var stopOnce sync.Once
	stop := func() { stopOnce.Do(func() { close(stopReaders) }); <-readersDone }
	t.Cleanup(stop)
	refreshStarted, refreshRelease := make(chan struct{}), make(chan struct{})
	refresh := func(ctx context.Context) string {
		close(refreshStarted)
		select {
		case <-refreshRelease:
			return "refreshed"
		case <-ctx.Done():
			return ""
		}
	}
	for i := range 50 {
		if err := agent.ReloadPrompts(context.Background(), "new system", "compact v2", "handoff v2", refresh); err != nil {
			t.Fatal(err)
		}
		if i == 0 && fake.count() != 0 {
			t.Fatal("prompt reload called the model")
		}
	}
	if agent.SystemPrompt() != "new system" || agent.compactPrompt() != "compact v2" || agent.handoffPrompt() != "handoff v2" {
		t.Fatalf("reload prompts mismatch: system=%q", agent.SystemPrompt())
	}
	if agent.renderedAt != noon || agent.refreshDue {
		t.Fatalf("reload timestamp/state = %v, due=%v", agent.renderedAt, agent.refreshDue)
	}
	agent.refreshDue = true
	prompts <- Request{Prompt: "next", Model: "m"}
	select {
	case <-refreshStarted:
	case <-time.After(3 * time.Second):
		t.Fatal("system refresh did not start")
	}
	close(refreshRelease)
	deadline := time.After(3 * time.Second)
	for {
		select {
		case update := <-updates:
			if update.Kind == UpdateDone {
				stop()
				if fake.count() != 1 || fake.sent(0)[0].Content != "refreshed" || agent.SystemPrompt() != "refreshed" {
					t.Fatal("concurrent system getter missed the refreshed prompt")
				}
				return
			}
		case <-deadline:
			t.Fatal("agent did not finish the request")
		}
	}
}
