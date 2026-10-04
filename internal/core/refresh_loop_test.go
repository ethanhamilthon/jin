package core

import (
	"strings"
	"testing"
)

func TestRefreshAfterCompactionInsideTheToolLoop(t *testing.T) {
	agent, fake := newRefreshRun(t, "tool", "summary", "answer", "answer two")
	runSteps(t, agent, Request{Prompt: "one", Model: "m", Window: 15})
	if system, _ := lastUser(fake); system != "sys v2" {
		t.Fatalf("the request after the in-loop compaction kept the old system: %q", system)
	}
	runSteps(t, agent, Request{Prompt: "two", Model: "m"})
	if _, user := lastUser(fake); !strings.HasPrefix(user, refreshedNote+"two") {
		t.Errorf("the note must reach the next user message: %q", user)
	}
}
