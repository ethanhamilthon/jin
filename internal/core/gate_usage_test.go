package core

import (
	"testing"

	"jin/internal/provider"
)

func TestGateUsageIncludesAllBilledTokens(t *testing.T) {
	var agent Agent
	failed := provider.Usage{Known: true, Input: 10, Output: 3, CachedInput: 2, CacheWriteInput: 1,
		Reasoning: 2, CacheKnown: true, CacheWriteKnown: true, ReasoningKnown: true}
	agent.addGateUsage(failed)
	agent.addGateUsage(provider.Usage{Input: 999})
	agent.addGateUsage(provider.Usage{Known: true, Input: 20, Output: 5, CachedInput: 4, CacheWriteInput: 2, Reasoning: 3})
	want := provider.Usage{Known: true, Input: 30, Output: 8, CachedInput: 6, CacheWriteInput: 3,
		Reasoning: 5, CacheKnown: true, CacheWriteKnown: true, ReasoningKnown: true}
	if agent.gateUsage != want {
		t.Errorf("gate usage = %+v, want %+v", agent.gateUsage, want)
	}
}
