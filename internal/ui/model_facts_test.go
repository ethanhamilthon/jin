package ui

import (
	"testing"

	"jin/internal/pricing"
)

func TestModelFacts(t *testing.T) {
	table := pricing.Table{"gpt-x": {MaxInputTokens: 200_000, InputCostPerToken: 1.25e-6, OutputCostPerToken: 10e-6, ReasoningKnown: true, Reasoning: true, VisionKnown: true, Vision: true}}
	options := modelOptions([]string{"gpt-x", "unknown"}, table)
	if got := options[0].detail; got != "200K ctx · $1.25 / $10 · reasoning · vision" {
		t.Errorf("detail = %q", got)
	}
	if options[1].detail != "" {
		t.Errorf("unknown model detail = %q", options[1].detail)
	}
}
