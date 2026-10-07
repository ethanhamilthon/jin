package core

import (
	"context"
	"testing"
	"time"
)

// TestTakenForRequestsWithoutTurn checks that every request the agent takes
// ends with UpdateDone or UpdateTaken: a prompt folded into a running turn
// and a blank prompt get no turn of their own.
func TestTakenForRequestsWithoutTurn(t *testing.T) {
	agent, _ := newFakeAgent(t, "tool", "answer")
	prompts := make(chan Request, 3)
	updates := make(chan Update, 64)
	prompts <- Request{Prompt: "first", Model: "m"}
	prompts <- Request{Prompt: "second", Model: "m"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go agent.Run(ctx, nil, prompts, updates)
	done, taken := 0, 0
	timeout := time.After(10 * time.Second)
	for done < 1 || taken < 2 {
		select {
		case u := <-updates:
			switch u.Kind {
			case UpdateDone:
				done++
				if taken == 1 {
					prompts <- Request{Prompt: "  "}
				}
			case UpdateTaken:
				taken++
			}
		case <-timeout:
			t.Fatalf("done = %d, taken = %d", done, taken)
		}
	}
	if done != 1 {
		t.Fatalf("done = %d, want 1", done)
	}
}
