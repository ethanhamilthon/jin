package core

import (
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
