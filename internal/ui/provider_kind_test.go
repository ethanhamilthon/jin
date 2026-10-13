package ui

import (
	"testing"

	"jin/internal/provider"
	"jin/internal/store"
)

func TestProviderKindLabels(t *testing.T) {
	if got := kindLabel("auto"); got != "OpenAI-compatible" {
		t.Fatalf("auto label = %q", got)
	}
	if got := kindLabel(provider.KindAnthropic); got != "Anthropic" {
		t.Fatalf("Anthropic label = %q", got)
	}
	existing := []store.ProviderEntry{{ID: "openai", Name: "openai"}}
	if got := defaultProviderName("auto", existing); got != "openai-2" {
		t.Fatalf("name = %q", got)
	}
}
