package core

import (
	"context"
	"testing"
	"time"

	"jin/internal/tools"
)

func TestDetachToolsClosesTheWatchedChannel(t *testing.T) {
	a := NewAgent(nil, "sys", tools.NewRegistry())
	ch := a.toolDetach()
	select {
	case <-ch:
		t.Fatal("nothing waits, the channel must stay open")
	default:
	}
	a.DetachTools()
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("DetachTools did not close the channel")
	}
	a.DetachTools() // no tool running: must not panic
}

func TestAMessageThatWaitedDetachesANewTool(t *testing.T) {
	a := NewAgent(nil, "sys", tools.NewRegistry())
	a.Expect()
	ch := a.toolDetach()
	select {
	case <-ch:
		t.Fatal("a quick tool must be given a moment to finish")
	default:
	}
	select {
	case <-ch:
	case <-time.After(3 * time.Second):
		t.Fatal("a tool that runs while a message waits must move to the background")
	}
}

func TestConsumedRequestBalancesExpect(t *testing.T) {
	a := NewAgent(nil, "sys", tools.NewRegistry())
	a.Expect()
	a.Expect()
	a.consumedRequest(Request{Interactive: true})
	if a.waiting.Load() != 1 {
		t.Errorf("waiting = %d, want 1", a.waiting.Load())
	}
	a.consumedRequest(Request{}) // a task result was never expected
	if a.waiting.Load() != 1 {
		t.Errorf("a request that was not counted changed the count: %d", a.waiting.Load())
	}
	a.consumedRequest(Request{Interactive: true})
	a.consumedRequest(Request{Interactive: true})
	if a.waiting.Load() != 0 {
		t.Errorf("the count must not go below zero: %d", a.waiting.Load())
	}
}

func TestBackgroundContextNeedsAnAdopter(t *testing.T) {
	a := NewAgent(nil, "sys", tools.NewRegistry())
	if a.backgroundContext(context.Background()) != context.Background() {
		t.Error("without SetBackground the context must stay as it is")
	}
	a.SetBackground(func(tools.Adoption) (string, error) { return "id", nil })
	if a.backgroundContext(context.Background()) == context.Background() {
		t.Error("with SetBackground the context must carry the background")
	}
}

func TestAgentCallsAreSafeWithoutAnAgent(t *testing.T) {
	var a *Agent
	a.Expect()
	a.DetachTools()
}
