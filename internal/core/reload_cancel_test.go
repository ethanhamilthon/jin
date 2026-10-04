package core

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCancelledReloadWhileRunningDoesNotApply(t *testing.T) {
	agent, systems, release := controlledReloadAgent(t)
	t.Cleanup(release)
	_, prompts, _, _ := startReloadRun(t, agent)
	prompts <- Request{Prompt: "first", Model: "m"}
	<-systems
	ctx, cancel := context.WithCancel(context.Background())
	reloaded := make(chan error, 1)
	go func() { reloaded <- agent.ReloadPrompts(ctx, "cancelled", "", "", nil) }()
	waitReloadQueued(t, agent)
	cancel()
	if err := <-reloaded; !errors.Is(err, context.Canceled) {
		t.Fatalf("reload error = %v, want cancellation", err)
	}
	release()
	prompts <- Request{Prompt: "second", Model: "m"}
	select {
	case got := <-systems:
		if got != "old" {
			t.Fatalf("cancelled reload changed system prompt to %q", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("second request did not reach the provider")
	}
	if agent.SystemPrompt() != "old" {
		t.Fatalf("system prompt = %q after cancellation", agent.SystemPrompt())
	}
}

func TestReloadAfterRunShutdownReturns(t *testing.T) {
	agent, _ := newFakeAgent(t, "unused")
	cancel, _, _, done := startReloadRun(t, agent)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("agent did not shut down")
	}
	ctx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if err := agent.ReloadPrompts(ctx, "new", "", "", nil); !errors.Is(err, errAgentStopped) {
		t.Fatalf("reload after shutdown = %v", err)
	}
}
