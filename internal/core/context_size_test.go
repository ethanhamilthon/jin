package core

import (
	"strings"
	"testing"

	"jin/internal/provider"
	"jin/internal/tools"
)

func TestSetContextSizeEnablesResumeCompaction(t *testing.T) {
	agent := NewAgent(provider.NewClient(provider.Config{}), "sys", tools.NewRegistry())
	agent.SetContextSize(90_000)
	if !needsCompaction(agent.size, 100_000) {
		t.Fatal("restored size must trigger compaction")
	}
	agent.SetContextSize(-5)
	if agent.size != 0 {
		t.Fatalf("size = %d", agent.size)
	}
}

func TestContextSizeCountsMessagesSinceReport(t *testing.T) {
	agent := NewAgent(provider.NewClient(provider.Config{}), "sys", tools.NewRegistry())
	history := []provider.Message{{Role: "system", Content: "sys"}, {Role: "user", Content: "hi"}}
	agent.reported(provider.Usage{Known: true, Input: 1000, Output: 50}, history)
	if got := agent.contextSize(history); got != 1050 {
		t.Fatalf("size = %d, want 1050", got)
	}
	history = append(history, provider.Message{Role: "tool", ToolCallID: "c1", Content: strings.Repeat("x", 4000)})
	if got := agent.contextSize(history); got != 2050 {
		t.Errorf("size with a new tool result = %d, want 2050", got)
	}
	agent.reported(provider.Usage{}, history)
	if agent.mark != 2 {
		t.Errorf("unknown usage moved the mark to %d", agent.mark)
	}
	agent.reseed(history)
	if agent.size < 1000 || agent.mark != 3 {
		t.Errorf("reseed: size = %d, mark = %d", agent.size, agent.mark)
	}
}

func TestStartSizeKeepsRestoredSize(t *testing.T) {
	agent := NewAgent(provider.NewClient(provider.Config{}), "sys", tools.NewRegistry())
	history := []provider.Message{{Role: "system", Content: strings.Repeat("s", 400)}, {Role: "user", Content: "hi"}}
	agent.SetContextSize(5000)
	agent.startSize(history)
	if agent.contextSize(history) != 5000 {
		t.Errorf("restored size = %d, want 5000", agent.contextSize(history))
	}
	fresh := NewAgent(provider.NewClient(provider.Config{}), "sys", tools.NewRegistry())
	fresh.startSize(history)
	if fresh.contextSize(history) < 100 {
		t.Errorf("fresh estimate = %d, want the system prompt counted", fresh.contextSize(history))
	}
}

func TestResumeOfPrunedSessionCountsFullHistory(t *testing.T) {
	full := toolTurns(6, 4000)
	before, _ := newFakeAgent(t, "unused")
	pruned := full
	if !before.prune(&pruned, pruneKeepTurns) {
		t.Fatal("nothing pruned")
	}
	before.reseed(pruned)
	before.reported(provider.Usage{Known: true, Input: before.size}, pruned)

	resumed, _ := newFakeAgent(t, "unused")
	resumed.SetContextSize(before.size)
	history := full
	resumed.startSize(history)
	if got, want := resumed.contextSize(history), 6000; got < want {
		t.Fatalf("resumed estimate = %d, want at least %d", got, want)
	}
	window := before.size * 10 / 6
	resumed.compactIfNeeded(t.Context(), t.Context(), Request{Model: "m", Window: window}, &history, make(chan Update, 8))
	if omitted(history) != 2 {
		t.Errorf("resume did not prune again: %d results omitted", omitted(history))
	}
}
