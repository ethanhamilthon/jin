package ui

import (
	"testing"

	"jin/internal/provider"
	"jin/internal/store"
)

func TestResponsesProviderKind(t *testing.T) {
	if got := kindLabel(provider.KindResponses); got != "OpenAI Responses" {
		t.Fatalf("label = %q", got)
	}
	existing := []store.ProviderEntry{{ID: "responses", Name: "responses"}}
	if got := defaultProviderName(provider.KindResponses, existing); got != "responses-2" {
		t.Fatalf("name = %q", got)
	}
}
