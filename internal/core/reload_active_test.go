package core

import (
	"context"
	"testing"
	"time"
)

func TestReloadWhileRunningPrecedesQueuedRequests(t *testing.T) {
	agent, systems, release := controlledReloadAgent(t)
	t.Cleanup(release)
	_, prompts, _, _ := startReloadRun(t, agent)
	prompts <- Request{Prompt: "first", Model: "m"}
	if got := <-systems; got != "old" {
		t.Fatalf("first request system = %q", got)
	}
	reloaded := make(chan error, 1)
	go func() { reloaded <- agent.ReloadPrompts(context.Background(), "new", "compact", "handoff", nil) }()
	waitReloadQueued(t, agent)
	prompts <- Request{Prompt: "queued during render", Model: "m"}
	release()
	select {
	case got := <-systems:
		if got != "new" {
			t.Fatalf("queued request system = %q, want new", got)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("queued request did not reach the provider")
	}
	select {
	case err := <-reloaded:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("reload did not complete")
	}
}
